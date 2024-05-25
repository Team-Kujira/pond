package feeder

import (
	"fmt"
	"os"
	"strconv"

	"pond/utils"

	"github.com/rs/zerolog"
)

type Feeder struct {
	logger  zerolog.Logger
	Name    string
	Command string
	Home    string
	Port    string
	IpAddr  string
}

func NewFeeder(
	logger zerolog.Logger,
	command, address string,
	chainNum, nodeNum uint,
) (Feeder, error) {
	name := fmt.Sprintf("feeder%d-%d", chainNum, nodeNum)
	port := strconv.Itoa(int(100+chainNum*10+nodeNum)) + "71"

	logger = logger.With().Str("node", name).Logger()

	home, err := os.UserHomeDir()
	if err != nil {
		logger.Error().Msg("could not get home directory")
		return Feeder{}, err
	}

	feeder := Feeder{
		logger:  logger,
		Command: command,
		Name:    name,
		Home:    home + "/.pond/" + name,
		Port:    port,
		IpAddr:  address,
	}

	return feeder, nil
}

func (f *Feeder) Init(namespace string) error {
	version, err := utils.GetVersion(f.logger, "feeder")
	if err != nil {
		return f.error(err)
	}

	os.MkdirAll(f.Home, 0o755)

	image := fmt.Sprintf("docker.io/%s/feeder:%s", namespace, version)

	err = f.CreateContainer(image)
	if err != nil {
		return f.error(err)
	}

	src := "config/kujira/feeder.toml"
	dst := fmt.Sprintf("%s/config.toml", f.Home)

	err = utils.Template(src, dst, f)
	if err != nil {
		return f.error(err)
	}

	return nil
}

func (f *Feeder) CreateContainer(image string) error {
	f.logger.Debug().Msg("create container")

	command := []string{
		f.Command, "container", "create", "--name", f.Name,
		"--network-alias", f.Name, "-v", f.Home + ":/home/feeder",
		"-p", fmt.Sprintf("%s:%s:%s", f.IpAddr, f.Port, f.Port),
		"--log-opt", "max-size=10m",
	}

	if f.Command == "docker" {
		command = append(command, []string{"--network", "pond"}...)
	}

	command = append(command, []string{image, "price-feeder", "/home/feeder/config.toml"}...)

	return utils.Run(f.logger, command)
}

func (f *Feeder) Start() error {
	f.logger.Info().Msg("start node")

	command := []string{f.Command, "start", f.Name}

	return utils.Run(f.logger, command)
}

func (f *Feeder) Stop() error {
	f.logger.Info().Msg("stop node")

	command := []string{f.Command, "stop", f.Name}

	return utils.Run(f.logger, command)
}

func (f *Feeder) error(err error) error {
	f.logger.Err(err).Msg("")
	return err
}
