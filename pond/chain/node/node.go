package node

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"pond/pond/globals"
	"pond/utils"
)

type Account struct {
	Address string
	Amount  int
}

type Ports struct {
	// 10 + chain + node + xx
	Abci   string // ex.: 11158
	Api    string // ex.: 11117
	App    string // ex.: 11156
	Feeder string // ex.: 11171
	Grpc   string // ex.: 11190
	Pprof  string // ex.: 11160
	Rpc    string // ex.: 11157
}

type Node struct {
	logger    zerolog.Logger
	initState bool
	Local     bool
	Image     string `json:"-"`        // ex.: docker.io/teamkujira/kujira:v0.8.4
	Command   string `json:"-"`        // ex.: docker
	Binary    string `json:"-"`        // ex.: kujirad or /usr/bin/kujirad
	Type      string `json:"-"`        // ex.: kujira
	ChainId   string `json:"-"`        // ex.: kujira-1
	Home      string `json:"-"`        // ex.: ~/.pond/kujira1-2
	Denom     string `json:"-"`        // ex.: ukuji
	Moniker   string `json:"moniker"`  // ex.: kujira1-2
	Mnemonic  string `json:"mnemonic"` // ex.: symbol rebuild hotel chief ensure hand coach ...
	NodeId    string `json:"node_id"`  // ex.: bf26617b40af84e1004c5e345bbbf7da12f121b3
	Address   string `json:"address"`  // ex.: kujira1r8u3eyf0axnsq9myrgtemtc9xpapxcezr6ek46
	Valoper   string `json:"valoper"`  // ex.: kujiravaloper1r8u3eyf0axnsq9myrgtemtc9xpapxcezy029f4
	Peers     string `json:"-"`        // ex.: bf26617b40af84e1004c5e345bbbf7da12f121b3@kujira1-2:11256,...
	Host      string `json:"-"`        // ex.: kujira1-1 or 127.0.0.1
	Ports     Ports  `json:"-"`
	ApiUrl    string `json:"api_url"`
	AppUrl    string `json:"app_url"`
	RpcUrl    string `json:"rpc_url"`
	GrpcUrl   string `json:"grpc_url"`
	FeederUrl string `json:"feeder_url"`
	OracleUrl string `json:"-"`
	IpAddr    string `json:"-"`
}

func NewNode(
	logger zerolog.Logger,
	command, binary, address, chainType string,
	typeNum, nodeNum, chainNum uint,
) (Node, error) {
	moniker := fmt.Sprintf("%s%d-%d", chainType, typeNum, nodeNum)

	logger = logger.With().
		Str("node", moniker).
		Logger()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Error().Msg("could not get home directory")
		return Node{}, err
	}

	mnemonic := globals.Mnemonics[fmt.Sprintf("validator%d", nodeNum)]

	base := strconv.Itoa(int(100 + chainNum*10 + nodeNum))
	ports := Ports{
		Abci:   base + "58",
		Api:    base + "17",
		App:    base + "56",
		Feeder: base + "71",
		Grpc:   base + "90",
		Pprof:  base + "60",
		Rpc:    base + "57",
	}

	node := Node{
		logger:   logger,
		Local:    false,
		Type:     chainType,
		Moniker:  moniker,
		Home:     homeDir + "/.pond/" + moniker,
		ChainId:  fmt.Sprintf("%s-%d", chainType, typeNum),
		Ports:    ports,
		Mnemonic: mnemonic,
		Command:  command,
		Binary:   globals.Chains[chainType].Command,
		Denom:    globals.Chains[chainType].Denom,
		AppUrl:   "tcp://" + address + ":" + ports.App,
		ApiUrl:   "http://" + address + ":" + ports.Api,
		RpcUrl:   "http://" + address + ":" + ports.Rpc,
		GrpcUrl:  "http://" + address + ":" + ports.Grpc,
		IpAddr:   address,
	}

	if binary != "" {
		node.Local = true
		node.Binary = binary
	}

	var feeder string
	if node.Command == "docker" && !node.Local {
		node.Host = node.Moniker
		feeder = fmt.Sprintf("feeder%d-%d", chainNum, nodeNum)
	} else {
		node.Host = "127.0.0.1"
	}

	if chainNum == 1 {
		node.OracleUrl = fmt.Sprintf(
			"http://%s:%s/api/v1/prices", feeder, ports.Feeder,
		)
		node.FeederUrl = fmt.Sprintf(
			"http://127.0.0.1:%s/api/v1/prices", ports.Feeder,
		)
	}

	return node, nil
}

func (n *Node) NewCommand(command []string) []string {
	if n.Local {
		return append(command, []string{"--home", n.Home}...)
	}

	return append([]string{
		n.Command, "exec", "--user", n.Type, n.Moniker,
	}, command...)
}

func (n *Node) Init(namespace string, amount int) error {
	command := n.NewCommand([]string{
		n.Binary, "init", n.Moniker, "--chain-id", n.ChainId,
	})

	if n.Type != "terra2" {
		command = append(command, []string{"--default-denom", n.Denom}...)
	}

	err := utils.Run(n.logger, command)
	if err != nil {
		n.logger.Err(err)
		return err
	}

	for i := 0; i < 10; i++ {
		_, err = os.Stat(fmt.Sprintf("%s/config/genesis.json", n.Home))
		if err == nil {
			break
		}
		n.logger.Debug().Msg("wait genesis.json")
		time.Sleep(time.Millisecond * 200)
	}
	if err != nil {
		return err
	}

	err = n.AddKey("validator", n.Mnemonic)
	if err != nil {
		n.logger.Err(err)
		return err
	}

	address, err := n.GetAddress("validator")
	if err != nil {
		return err
	}

	n.Address = address

	err = n.AddGenesisAccount(n.Address, amount)
	if err != nil {
		return err
	}

	// updating the genesis file takes a little bit, so sleep 100ms
	// to ensure the address is found and prevent sleeping once for 500ms
	time.Sleep(time.Millisecond * 100)

	for i := 0; i < 5; i++ {
		data, err := os.ReadFile(fmt.Sprintf("%s/config/genesis.json", n.Home))
		if err != nil {
			return err
		}

		if strings.Contains(string(data), address) {
			break
		}

		n.logger.Debug().Msg("wait for genesis account")
		time.Sleep(time.Millisecond * 500)
	}

	err = n.CreateGentx(amount / 2)
	if err != nil {
		return err
	}

	return nil
}

func (n *Node) AddGenesisAccounts(accounts []Account) error {
	if n.Local {
		return n.AddGenesisAccountsLocal(accounts)
	}

	return n.AddGenesisAccountsContainer(accounts)
}

func (n *Node) AddGenesisAccountsContainer(accounts []Account) error {
	n.logger.Debug().Msg("add genesis accounts")

	addresses := make([]string, len(accounts))
	command := []string{
		n.Command, "exec", "--user", n.Type,
	}

	for i, account := range accounts {
		command = append(command, []string{
			"-e", fmt.Sprintf(
				"%s=%d%s", account.Address, account.Amount, n.Denom),
		}...)
		addresses[i] = account.Address
	}

	command = append(command, []string{n.Moniker, "bash", "-c"}...)
	command = append(command, fmt.Sprintf(
		`for address in %s; do \
			%s genesis add-genesis-account $address ${!address}
		done`, strings.Join(addresses, " "), n.Binary,
	))

	return utils.Run(n.logger, command)
}

func (n *Node) AddGenesisAccountsLocal(accounts []Account) error {
	n.logger.Debug().Msg("add genesis accounts")

	for _, account := range accounts {
		command := []string{
			n.Binary, "--home", n.Home, "genesis", "add-genesis-account",
			account.Address, fmt.Sprintf("%d%s", account.Amount, n.Denom),
		}

		err := utils.Run(n.logger, command)
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *Node) AddGenesisAccount(address string, amount int) error {
	n.logger.Debug().
		Str("address", address).
		Msg("add genesis account")

	// TODO: if init is too slow, add '-d' for containers
	// n.Command, "exec", "--user", n.Type, "-d", n.Moniker,
	command := n.NewCommand([]string{
		n.Binary, "genesis", "add-genesis-account", address,
		strconv.Itoa(amount) + n.Denom,
	})

	return utils.Run(n.logger, command)
}

func (n *Node) CreateGentx(amount int) error {
	n.logger.Debug().Msg("create gentx")

	command := n.NewCommand([]string{
		n.Binary, "genesis", "gentx", "validator", "--keyring-backend", "test",
		strconv.Itoa(amount) + n.Denom, "--chain-id", n.ChainId,
		"--output", "json",
	})

	err := utils.Run(n.logger, command)
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(n.Home + "/config/gentx")
	if err != nil {
		n.logger.Err(err).Msg("")
		return err
	}

	var filename string

	regex, err := regexp.Compile("^gentx-[0-9a-f]{40}.json$")
	if err != nil {
		return n.error(err)
	}

	for _, entry := range entries {
		name := entry.Name()
		if regex.MatchString(name) {
			filename = fmt.Sprintf("%s/config/gentx/%s", n.Home, name)
			break
		}
	}

	if filename == "" {
		err = fmt.Errorf("no gentx found")
		return n.error(err)
	}

	n.NodeId = (filename[len(filename)-45 : len(filename)-5])

	data, err := os.ReadFile(filename)
	if err != nil {
		return n.error(err)
	}

	var gentx struct {
		Body struct {
			Messages []struct {
				Valoper string `json:"validator_address"`
			} `json:"messages"`
		} `json:"body"`
	}

	err = json.Unmarshal(data, &gentx)
	if err != nil {
		return n.error(err)
	}

	n.Valoper = gentx.Body.Messages[0].Valoper

	return nil
}

func (n *Node) CollectGentxs() error {
	n.logger.Debug().Msg("collect gentxs")
	command := n.NewCommand([]string{
		n.Binary, "genesis", "collect-gentxs",
	})

	return utils.Run(n.logger, command)
}

func (n *Node) AddKey(wallet, mnemonic string) error {
	n.logger.Debug().Str("wallet", wallet).Msg("add key")

	command := []string{
		n.Binary, "--keyring-backend", "test",
		"keys", "add", wallet, "--recover",
	}

	if n.Local {
		// just append home
		command = n.NewCommand(command)
	} else {
		command = append([]string{
			n.Command, "exec", "--user", n.Type, "-i", n.Moniker,
		}, command...)
	}

	return utils.RunI(n.logger, command, mnemonic)
}

func (n *Node) AddKeys(mnemonics map[string]string) error {
	n.logger.Debug().Msg("add keys")

	if n.Local {
		n.logger.Debug().Msg("add keys")

		for wallet, mnemnemonic := range mnemonics {
			err := n.AddKey(wallet, mnemnemonic)
			if err != nil {
				return err
			}
		}

		return nil
	}

	wallets := []string{}
	command := []string{
		n.Command, "exec", "--user", n.Type,
	}

	for wallet, mnemonic := range mnemonics {
		command = append(command, []string{
			"-e", fmt.Sprintf("%s=%s", wallet, mnemonic),
		}...)
		wallets = append(wallets, wallet)
	}

	command = append(command, []string{n.Moniker, "bash", "-c"}...)
	command = append(command, fmt.Sprintf(
		`for wallet in %s; do \
			echo -n ${!wallet} | %s \
			--keyring-backend test keys add $wallet --recover;\
		done`, strings.Join(wallets, " "), n.Binary,
	))

	return utils.Run(n.logger, command)
}

func (n *Node) GetAddress(wallet string) (string, error) {
	command := n.NewCommand([]string{
		n.Binary, "--keyring-backend", "test",
		"keys", "show", "-a", wallet,
	})

	output, err := utils.RunO(n.logger, command)
	if err != nil {
		return "", err
	}

	address := strings.TrimSuffix(string(output), "\n")

	return address, nil
}

func (n *Node) GetAddresses() (map[string]string, error) {
	command := n.NewCommand([]string{
		n.Binary, "--keyring-backend", "test",
		"keys", "list", "--output", "json",
	})

	output, err := utils.RunO(n.logger, command)
	if err != nil {
		return nil, err
	}

	var keys []struct {
		Name    string `json:"name"`
		Address string `json:"address"`
	}

	err = json.Unmarshal(output, &keys)

	addresses := map[string]string{}
	for _, key := range keys {
		addresses[key.Name] = key.Address
	}

	return addresses, err
}

func (n *Node) CreateInitContainer(image string) error {
	n.initState = true
	return n.createContainer(image, true)
}

func (n *Node) CreateRunContainer(image string) error {
	n.initState = false
	return n.createContainer(image, false)
}

func (n *Node) createContainer(image string, init bool) error {
	err := n.RemoveContainer()
	if err != nil {
		return err
	}

	n.logger.Debug().Msg("create container")

	config, found := globals.Chains[n.Type]
	if !found {
		err = fmt.Errorf("home not set")
		n.error(err)
	}

	command := []string{
		n.Command, "container", "create", "--name", n.Moniker,
		"--network-alias", n.Moniker, "--log-opt", "max-size=10m",
		"-v", fmt.Sprintf("%s:/home/%s/%s", n.Home, n.Type, config.Home),
	}

	if n.Command == "docker" {
		command = append(command, []string{"--network", "pond"}...)
	}

	if init {
		command = append(command, []string{
			"--stop-signal", "SIGKILL", image, "tail", "-f", "/dev/null",
		}...)
		return utils.Run(n.logger, command)
	}

	ports := []string{n.Ports.Api, n.Ports.App, n.Ports.Rpc, n.Ports.Grpc}
	for _, port := range ports {
		command = append(command, "-p")
		command = append(command, fmt.Sprintf("%s:%s:%s", n.IpAddr, port, port))
	}

	command = append(command, []string{
		image, n.Binary, "start",
	}...)

	return utils.Run(n.logger, command)
}

func (n *Node) RemoveContainer() error {
	n.logger.Debug().Msg("remove container")

	command := []string{n.Command, "rm", "-f", n.Moniker}

	return utils.Run(n.logger, command)
}

func (n *Node) error(err error) error {
	n.logger.Err(err).Msg("")
	return err
}

func (n *Node) Start() error {
	if n.Local {
		pid, err := n.GetPid()
		if err != nil {
			return err
		}

		if pid != "" {
			n.logger.Debug().Msg("node already running")
			err := n.Stop()
			if err != nil {
				return err
			}
		}

		command := []string{
			"nohup", n.Binary, "--home", n.Home, "start",
		}

		logfile := filepath.Join(n.Home, "kujirad.log")

		return utils.RunB(n.logger, command, logfile)
	}

	if n.initState {
		n.logger.Debug().Str("state", "init").Msg("start node")
	} else {
		n.logger.Info().Msg("start node")
	}

	command := []string{n.Command, "start", n.Moniker}

	return utils.Run(n.logger, command)
}

func (n *Node) Stop() error {
	n.logger.Info().Msg("stop node")

	n.RemoveTemp()

	if !n.Local {
		command := []string{n.Command, "stop", n.Moniker}
		return utils.Run(n.logger, command)
	}

	pid, err := n.GetPid()
	if err != nil {
		return err
	}

	if pid == "" {
		err := fmt.Errorf("pid not found")
		return n.error(err)
	}

	return utils.Run(n.logger, []string{"kill", pid})
}

func (n *Node) Query(args []string) ([]byte, error) {
	command := n.NewCommand([]string{
		n.Binary, "query",
	})

	command = append(command, args...)

	output, err := utils.RunO(zerolog.Nop(), command)

	// some output rewrite to avoid confision
	lines := []string{}

	for _, line := range strings.Split(string(output), "\n") {
		line = strings.Replace(line, n.Binary+" query", "pond query", -1)

		// skip docker error msg: 'ERR failure when running app err=...'
		if strings.Contains(line, "failure when running app") {
			continue
		}

		lines = append(lines, line)
	}

	return []byte(strings.Join(lines, "\n")), err
}

func (n *Node) Tx(args []string) ([]byte, error) {
	command := n.NewCommand([]string{
		n.Binary, "tx",
	})

	command = append(command, args...)
	command = append(command, []string{
		"--keyring-backend", "test", "--chain-id", n.ChainId, "--yes",
	}...)

	output, err := utils.RunO(zerolog.Nop(), command)

	// some output rewrite to avoid confusion
	lines := []string{}

	for _, line := range strings.Split(string(output), "\n") {
		line = strings.Replace(line, n.Binary+" tx", "pond tx", -1)

		// skip docker error msg: 'ERR failure when running app err=...'
		if strings.Contains(line, "failure when running app") {
			continue
		}

		lines = append(lines, line)
	}

	if err != nil {
		fmt.Println(string(output))
	}

	return []byte(strings.Join(lines, "\n")), err
}

func (n *Node) Generate(args []string) ([]byte, error) {
	args = append(args, []string{"--generate-only", "--gas", "100000000"}...)
	return n.Tx(args)
}

func (n *Node) Status() ([]byte, error) {
	n.logger.Debug().Msg("get status")

	command := n.NewCommand([]string{
		n.Binary, "status",
	})

	return utils.RunO(n.logger, command)
}

func (n *Node) WaitForTx(hash string) error {
	n.logger.Debug().Str("hash", hash).Msg("wait for tx")

	cycles := 10
	interval := time.Second * 1

	args := []string{"tx", hash}

	var (
		err  error
		data []byte
	)

	for i := 0; i < cycles; i++ {
		n.logger.Debug().
			Int("cycle", i).
			Msg("query tx")
		time.Sleep(interval)

		data, err = n.Query(args)
		if err == nil {
			_, err := utils.CheckTxResponse(data)
			if err != nil {
				return n.error(err)
			}

			return nil
		}
	}

	return err
}

func (n *Node) GetPid() (string, error) {
	output, err := utils.RunO(n.logger, []string{"ps", "-a"})
	if err != nil {
		return "", err
	}

	var pid string

	substr := fmt.Sprintf("%s --home %s start", n.Binary, n.Home)

	for _, line := range strings.Split(string(output), "\n") {
		if strings.Contains(line, substr) {
			pid = strings.Split(line, " ")[0]
			break
		}
	}

	return pid, nil
}

func (n *Node) CreateTemp(data []byte, pattern string) (string, error) {
	tmpdir := n.Home + "/tmp"
	err := os.MkdirAll(tmpdir, 0o755)
	if err != nil {
		return "", n.error(err)
	}

	tmp, err := os.CreateTemp(tmpdir, pattern)
	if err != nil {
		return "", n.error(err)
	}

	err = os.WriteFile(tmp.Name(), data, 0o644)
	if err != nil {
		return "", n.error(err)
	}

	if n.Local {
		return tmp.Name(), nil
	}

	return "/home/kujira/.kujira/tmp/" + filepath.Base(tmp.Name()), nil
}

func (n *Node) RemoveTemp() error {
	return os.RemoveAll(n.Home + "/tmp")
}
