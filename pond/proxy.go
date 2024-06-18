package pond

import (
	"fmt"
	"os"

	"pond/utils"

	"github.com/rs/zerolog"
)

type Proxy struct {
	logger  zerolog.Logger
	Command string
	Home    string
	Address string
}

func NewProxy(logger zerolog.Logger, command, address string) (Proxy, error) {
	logger.Debug().Msg("create proxy")

	home, err := os.UserHomeDir()
	if err != nil {
		logger.Error().Msg("could not get home directory")
		return Proxy{}, err
	}

	return Proxy{
		logger:  logger.With().Str("node", "proxy").Logger(),
		Command: command,
		Home:    home + "/.pond/proxy",
		Address: address,
	}, nil
}

func (p *Proxy) Init(namespace string, local bool) error {
	version, err := utils.GetVersion(p.logger, "proxy")
	if err != nil {
		p.logger.Err(err).Msg("")
		return err
	}

	os.MkdirAll(p.Home, 0o755)

	config := struct{ Host string }{}

	if p.Command == "docker" {
		config.Host = "kujira1-1"
	}

	if local {
		config.Host = "host.docker.internal"
	}

	src := "config/proxy.conf"
	dst := p.Home + "/proxy-https.conf"

	utils.Template(src, dst, config)

	image := fmt.Sprintf("docker.io/%s/proxy:%s", namespace, version)

	return p.CreateContainer(image)
}

func (p *Proxy) CreateContainer(image string) error {
	p.logger.Debug().Msg("create container")

	command := []string{
		p.Command, "container", "create", "--name", "proxy",
		"--network-alias", "proxy", "-v", p.Home + ":/etc/nginx/conf.d",
		"-p", "127.0.0.1:10443:443", "-p", "127.0.0.1:10157:80",
		"--log-opt", "max-size=10m",
	}

	if p.Command == "docker" {
		command = append(command, []string{"--network", "pond"}...)
	}

	command = append(command, image)

	return utils.Run(p.logger, command)
}

func (p *Proxy) Start() error {
	p.logger.Info().Msg("start node")

	command := []string{p.Command, "start", "proxy"}

	return utils.Run(p.logger, command)
}

func (p *Proxy) Stop() error {
	p.logger.Info().Msg("stop node")

	command := []string{p.Command, "stop", "proxy"}

	return utils.Run(p.logger, command)
}
