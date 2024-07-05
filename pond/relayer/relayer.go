package relayer

import (
	"fmt"
	"os"

	"pond/pond/chain/node"
	"pond/pond/globals"
	"pond/utils"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v2"
)

type Relayer struct {
	logger  zerolog.Logger
	nodes   []node.Node
	Command string
	Name    string
	Home    string
	Port    string
	Paths   []string
	Address string
}

func NewRelayer(
	logger zerolog.Logger,
	command, address string,
	nodes []node.Node,
) (Relayer, error) {
	logger.Debug().Msg("create relayer")

	logger = logger.With().Str("node", "relayer").Logger()

	home, err := os.UserHomeDir()
	if err != nil {
		logger.Error().Msg("could not get home directory")
		return Relayer{}, err
	}

	relayer := Relayer{
		logger:  logger,
		nodes:   nodes,
		Command: command,
		Home:    home + "/.pond/relayer",
		Port:    "11183",
		Name:    "relayer",
		Address: address,
	}

	return relayer, nil
}

func (r *Relayer) Init(namespace string) error {
	version, err := utils.GetVersion(r.logger, "relayer")
	if err != nil {
		return r.error(err)
	}

	os.MkdirAll(r.Home+"/config", 0o755)
	os.MkdirAll(r.Home+"/keys", 0o755)

	image := fmt.Sprintf("docker.io/%s/relayer:%s", namespace, version)

	config := NewConfig(r.Port)

	for i, node := range r.nodes {
		src := node.Home + "/keyring-test"
		dst := r.Home + "/keys/" + node.ChainId + "/keyring-test"
		utils.CopyDir(r.logger, src, dst)

		info, found := globals.Chains[node.Type]
		if !found {
			err := fmt.Errorf("no chain info found")
			r.logger.Err(err).Str("type", node.Type)
			return err
		}

		host := node.Host
		if node.Local {
			host = "host.docker.internal"
		}

		chain := NewChainConfig()
		chain.Value.KeyDirectory = fmt.Sprintf("/relayer/keys/%s", node.ChainId)
		chain.Value.RpcAddr = fmt.Sprintf("http://%s:%s", host, node.Ports.Rpc)
		chain.Value.AccountPrefix = info.Prefix
		if info.Prefix == "dydx" {
			chain.Value.GasPrices = "10000000000" + info.Denom
		} else {
			chain.Value.GasPrices = "0.01" + info.Denom
		}
		chain.Value.ChainId = node.ChainId

		config.Chains[node.ChainId] = chain

		if i == 0 {
			continue
		}

		src = r.nodes[0].ChainId
		dst = node.ChainId

		path := Path{}
		path.Src.ChainId = src
		path.Dst.ChainId = dst

		name := src + "-" + dst

		config.Paths[name] = path
		r.Paths = append(r.Paths, name)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		r.error(err)
	}
	os.WriteFile(r.Home+"/config/config.yaml", data, 0o666)

	err = r.CreateContainer(image)
	if err != nil {
		return r.error(err)
	}

	return nil
}

func (r *Relayer) CreateContainer(image string) error {
	r.logger.Debug().Msg("create container")

	command := []string{
		r.Command, "container", "create", "--name", "relayer",
		"--network-alias", "relayer", "-v", r.Home + ":/home/relayer",
		"-p", fmt.Sprintf("%s:%s:%s", r.Address, r.Port, r.Port),
		"--log-opt", "max-size=10m",
	}

	if r.Command == "docker" {
		command = append(command, []string{"--network", "pond"}...)
	}

	// command = append(command, []string{image, "tail", "-f", "/dev/null"}...)

	command = append(command, []string{image, "link-and-start.sh"}...)
	command = append(command, r.Paths...)

	return utils.Run(r.logger, command)
}

func (r *Relayer) Start() error {
	r.logger.Info().Msg("start node")

	command := []string{r.Command, "start", r.Name}

	return utils.Run(r.logger, command)
}

func (r *Relayer) Stop() error {
	r.logger.Info().Msg("stop node")

	command := []string{r.Command, "stop", r.Name}

	return utils.Run(r.logger, command)
}

func (r *Relayer) error(err error) error {
	r.logger.Err(err).Msg("")
	return err
}
