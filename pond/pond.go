package pond

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"pond/pond/chain"
	"pond/pond/chain/node"
	"pond/pond/deployer"
	"pond/pond/relayer"
	"pond/pond/templates"
	"pond/utils"

	"github.com/rs/zerolog"
)

type Pond struct {
	logger   zerolog.Logger
	home     string
	config   Config
	info     Info
	chains   []chain.Chain
	relayer  relayer.Relayer
	deployer deployer.Deployer
	proxy    Proxy
	registry Registry
}

func NewPond(logLevel string) (Pond, error) {
	level := zerolog.InfoLevel
	switch logLevel {
	case "debug":
		level = zerolog.DebugLevel
	case "trace":
		level = zerolog.TraceLevel
	case "warn":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	}

	logger := zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.Stamp,
	}).With().Timestamp().Logger()

	zerolog.TimeFieldFormat = time.RFC3339Nano

	zerolog.SetGlobalLevel(level)

	home, err := os.UserHomeDir()
	if err != nil {
		logger.Error().Msg("could not get home directory")
		return Pond{}, err
	}

	pond := Pond{
		logger: logger,
		home:   home + "/.pond",
		info:   Info{},
		config: Config{},
	}

	pond.LoadConfig()
	pond.LoadInfo()

	pond.init()

	return pond, nil
}

func (p *Pond) init() error {
	for i, config := range p.config.Chains {
		chain, err := chain.NewChain(
			p.logger,
			p.config.Command,
			p.config.Namespace,
			p.config.Address,
			config.Type,
			config.TypeNum,
			config.Nodes,
			uint(i+1),
		)
		if err != nil {
			panic(err)
		}

		p.chains = append(p.chains, chain)
	}

	nodes := make([]node.Node, len(p.chains))
	for i, chain := range p.chains {
		nodes[i] = chain.Nodes[0]
	}

	var err error

	if len(nodes) == 0 {
		return nil
	}

	accounts := []string{}
	for name, account := range p.info.Accounts {
		if strings.HasPrefix(name, "test") {
			accounts = append(accounts, account.Addresses["kujira"])
		}
	}

	p.deployer, err = deployer.NewDeployer(
		p.logger, p.home, nodes[0], p.config.ApiUrl, accounts,
	)
	if err != nil {
		return p.error(err)
	}

	p.proxy, err = NewProxy(p.logger, p.config.Command, p.config.Address)
	if err != nil {
		return err
	}

	p.registry, err = NewRegistry(p.logger, p.home+"/registry.json")
	if err != nil {
		return err
	}

	if len(p.config.Chains) == 1 {
		return nil
	}

	p.relayer, err = relayer.NewRelayer(
		p.logger, p.config.Command, p.config.Address, nodes,
	)
	if err != nil {
		p.logger.Err(err).Msg("")
		return err
	}

	return nil
}

func (p *Pond) Start() error {
	var wg sync.WaitGroup
	for i := range p.chains {
		wg.Add(1)
		go func(i int) {
			p.chains[i].Start()
			wg.Done()
		}(i)
	}
	wg.Wait()

	p.proxy.Start()
	p.relayer.Start()

	// wait for kujira-1
	p.logger.Info().Msg("wait for pond to start")

	address := net.JoinHostPort(p.config.Address, p.chains[0].Nodes[0].Ports.Rpc)
	conn, err := net.DialTimeout("tcp", address, time.Second*5)
	if err != nil {
		return p.error(err)
	}

	if conn == nil {
		err = fmt.Errorf("unable to connect to kujira-1")
		return p.error(err)
	}

	p.chains[0].WaitBlocks(1)

	for _, plan := range p.config.Plans {
		data, err := templates.Templates.ReadFile("plan/" + plan + ".json")
		if err != nil {
			return p.error(err)
		}

		err = p.deployer.LoadPlan(data, plan)
		if err != nil {
			return err
		}
	}

	err = p.deployer.DeployPlan()
	if err != nil {
		return err
	}

	err = p.UpdateCodes()
	if err != nil {
		return err
	}

	err = p.UpdateContracts()
	if err != nil {
		return err
	}

	p.config.Plans = []string{}
	return p.SaveConfig()
}

func (p *Pond) Stop() error {
	var wg sync.WaitGroup
	for i := range p.chains {
		wg.Add(1)
		go func(i int) {
			p.chains[i].Stop()
			wg.Done()
		}(i)
	}

	wg.Add(1)
	go func() {
		p.relayer.Stop()
		p.proxy.Stop()
		wg.Done()
	}()

	wg.Wait()

	return nil
}

func (p *Pond) Clear() error {
	p.logger.Debug().Msg("clear pond")

	var command []string

	// remove all containers

	switch p.config.Command {
	case "docker":
		command = []string{
			"docker", "ps", "-af", "network=pond", "-q",
		}
	}

	output, err := utils.RunO(p.logger, command)
	if err != nil {
		return err
	}

	containers := strings.Split(string(output), "\n")
	command = []string{p.config.Command, "rm", "-f"}

	for _, container := range containers {
		if container == "" {
			continue
		}

		command = append(command, container)
	}

	if len(command) > 3 {
		err = utils.Run(p.logger, command)
		if err != nil {
			return err
		}
	}

	p.RemoveNetwork()

	time.Sleep(time.Millisecond * 500)

	return nil
}

func (p *Pond) Get() error {
	return nil
}

func (p *Pond) Set() error {
	return nil
}

func (p *Pond) Query(chainId string, args []string) ([]byte, error) {
	for _, chain := range p.chains {
		if chain.ChainId != chainId {
			continue
		}

		return chain.Nodes[0].Query(args)
	}

	err := fmt.Errorf("chain not found")
	p.logger.Err(err).Str("chain", chainId).Msg("")
	return nil, err
}

func (p *Pond) Tx(chainId string, args []string) ([]byte, error) {
	for _, chain := range p.chains {
		if chain.ChainId != chainId {
			continue
		}

		return chain.Nodes[0].Tx(args)
	}

	err := fmt.Errorf("chain not found")
	p.logger.Err(err).Str("chain", chainId).Msg("")
	return nil, err
}

func (p *Pond) CreateNetwork() error {
	p.RemoveNetwork()

	command := []string{
		p.config.Command, "network", "create", "pond",
	}

	err := utils.Run(p.logger, command)
	if err != nil {
		return err
	}

	return nil
}

func (p *Pond) RemoveNetwork() error {
	var command []string

	switch p.config.Command {
	case "docker":
		command = []string{
			p.config.Command, "network", "rm", "-f", "pond",
		}
	}

	err := utils.Run(p.logger, command)
	if err != nil {
		return err
	}

	return nil
}

func (p *Pond) Deploy(filenames []string) error {
	err := p.deployer.Deploy(filenames)
	if err != nil {
		return err
	}

	err = p.UpdateCodes()
	if err != nil {
		return err
	}

	err = p.UpdateContracts()
	if err != nil {
		return err
	}

	return nil
}

func (p *Pond) UpdateCodes() error {
	var err error
	p.info.Codes, err = p.deployer.GetDeployedCodes()
	if err != nil {
		return err
	}

	return p.SaveInfo()
}

func (p *Pond) UpdateContracts() error {
	p.logger.Debug().Msg("update contracts")

	contracts, err := p.deployer.GetDeployedContracts()
	if err != nil {
		return err
	}

	known := map[string]struct{}{}
	for _, contract := range p.info.Contracts {
		known[contract.Address] = struct{}{}
	}

	info := []Contract{}

	for _, contract := range contracts {
		_, found := known[contract.Address]
		if found {
			continue
		}

		info = append(info, Contract{
			Address: contract.Address,
			CodeId:  contract.Code,
			Label:   contract.Label,
		})
	}

	p.info.Contracts = append(p.info.Contracts, info...)
	return p.SaveInfo()
}

func (p *Pond) SubmitProposal(filename, option string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		p.logger.Err(err).Msg("failed reading proposal json")
		return err
	}

	return p.chains[0].SubmitProposal(data, option)
}

func (p *Pond) error(err error) error {
	p.logger.Err(err).Msg("")
	return err
}
