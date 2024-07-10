package signer

import (
	"fmt"
	"os"
	"strconv"
	"time"

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
	NodeUrl string
}

func NewHorcrux(
	logger zerolog.Logger,
	config Config,
) (*Horcrux, error) {
	name := fmt.Sprintf("horcrux%d-%d", config.ChainNum, config.NodeNum)

	logger = logger.With().Str("node", name).Logger()

	home, err := os.UserHomeDir()
	if err != nil {
		logger.Error().Msg("could not get home directory")
		return nil, err
	}

	base := strconv.Itoa(int(100 + config.ChainNum*10 + config.NodeNum))

	Horcrux := Horcrux{
		logger:  logger,
		Command: config.Command,
		Name:    name,
		Home:    home + "/.pond/" + name,
		Port:    base + "22",
		NodeUrl: config.NodeUrl,
	}

	return &Horcrux, nil
}

func (h *Horcrux) Init(namespace, keyfile string) error {
	h.logger.Debug().Msg("init")

	version, err := utils.GetVersion(h.logger, "horcrux")
	if err != nil {
		return h.error(err)
	}

	os.MkdirAll(h.Home+"/state", 0o755)

	image := fmt.Sprintf("docker.io/%s/horcrux:%s", namespace, version)

	err = h.CreateContainer(image, true)
	if err != nil {
		return h.error(err)
	}

	h.Start()

	time.Sleep(time.Second)

	src := "config/horcrux.yaml"
	dst := h.Home + "/config.yaml"

	err = utils.Template(src, dst, h)
	if err != nil {
		return h.error(err)
	}

	src = keyfile
	dst = h.Home + "/priv_validator_key.json"

	err = utils.CopyFile(h.logger, src, dst)
	if err != nil {
		return err
	}

	time.Sleep(time.Second)

	command := h.NewCommand([]string{
		"create-ecies-shards", "--shards", "1",
	})

	utils.Run(h.logger, command)

	command = h.NewCommand([]string{
		"create-ed25519-shards", "--chain-id", "kujira-1",
		"--key-file", "priv_validator_key.json",
		"--threshold", "1", "--shards", "1",
	})

	utils.Run(h.logger, command)

	for _, filename := range []string{"ecies_keys", "kujira-1_shard"} {
		src := fmt.Sprintf("%s/cosigner_1/%s.json", h.Home, filename)
		dst := fmt.Sprintf("%s/%s.json", h.Home, filename)
		utils.CopyFile(h.logger, src, dst)
	}

	h.RemoveContainer()

	h.CreateContainer(image, false)

	// command = []string{
	// 	h.Command, "exec", "--user", "horcrux",
	// 	"create-ecies-shards", "--shards", "1",
	// }

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
	if !h.init {
		h.logger.Info().Msg("start node")
	}

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

func (h *Horcrux) NewCommand(command []string) []string {
	return append([]string{
		h.Command, "exec", "--user", "horcrux",
		"-w", "/home/horcrux/.horcrux", h.Name, "horcrux",
	}, command...)
}
