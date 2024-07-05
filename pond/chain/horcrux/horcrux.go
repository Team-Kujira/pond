package horcrux

import (
	"fmt"
	"os"
	"strconv"

	"pond/utils"

	"github.com/rs/zerolog"
)

type Horcrux struct {
	logger  zerolog.Logger
	init    bool
	Name    string
	Command string
	Home    string
	Port    string
	IpAddr  string
}

func NewHorcrux(
	logger zerolog.Logger,
	command, address string,
	chainNum, nodeNum uint,
) (Horcrux, error) {
	name := fmt.Sprintf("horcrux%d-%d", chainNum, nodeNum)
	port := strconv.Itoa(int(100+chainNum*10+nodeNum)) + "71"

	logger = logger.With().Str("node", name).Logger()

	home, err := os.UserHomeDir()
	if err != nil {
		logger.Error().Msg("could not get home directory")
		return Horcrux{}, err
	}

	Horcrux := Horcrux{
		logger:  logger,
		Command: command,
		Name:    name,
		Home:    home + "/.pond/" + name,
		Port:    port,
		IpAddr:  address,
	}

	return Horcrux, nil
}

func (h *Horcrux) Init(namespace string) error {
	version, err := utils.GetVersion(h.logger, "horcrux")
	if err != nil {
		return h.error(err)
	}

	os.MkdirAll(h.Home+"/.horcrux", 0o755)

	image := fmt.Sprintf("docker.io/%s/horcrux:%s", namespace, version)

	err = h.CreateContainer(image, true)
	if err != nil {
		return h.error(err)
	}

	// src := "config/kujira/Horcrux.toml"
	// dst := fmt.Sprintf("%s/config.toml", h.Home)

	// err = utils.Template(src, dst, h)
	// if err != nil {
	// 	return h.error(err)
	// }

	return nil
}

func (h *Horcrux) RemoveContainer() error {
	h.logger.Debug().Msg("remove container")

	command := []string{h.Command, "rm", "-f", h.Name}

	return utils.Run(h.logger, command)
}

func (h *Horcrux) CreateContainer(image string, init bool) error {
	h.init = init

	err := h.RemoveContainer()
	if err != nil {
		return err
	}

	h.logger.Debug().Msg("create container")

	command := []string{
		h.Command, "container", "create", "--name", h.Name,
		"--network-alias", h.Name, "--log-opt", "max-size=10m",
		"-v", fmt.Sprintf("%s:/home/horcrux/.horcrux", h.Home),
	}

	if h.Command == "docker" {
		command = append(command, []string{"--network", "pond"}...)
	}

	if h.init {
		command = append(command, []string{
			"--stop-signal", "SIGKILL", image, "tail", "-f", "/dev/null",
		}...)
		return utils.Run(h.logger, command)
	}

	command = append(command, []string{
		image, "horcrux", "start",
	}...)

	return utils.Run(h.logger, command)
}

func (h *Horcrux) Start() error {
	h.logger.Info().Msg("start node")

	command := []string{h.Command, "start", h.Name}

	return utils.Run(h.logger, command)
}

func (h *Horcrux) Stop() error {
	h.logger.Info().Msg("stop node")

	command := []string{h.Command, "stop", h.Name}

	return utils.Run(h.logger, command)
}

func (h *Horcrux) error(err error) error {
	h.logger.Err(err).Msg("")
	return err
}
