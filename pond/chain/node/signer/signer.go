package signer

import (
	"fmt"

	"github.com/rs/zerolog"
)

type Config struct {
	Type     string
	Command  string
	ChainNum uint
	NodeNum  uint
	NodeUrl  string
}

type Signer interface {
	Start() error
	Stop() error
	Init(namespace, keyfile string) error
}

func NewSigner(
	logger zerolog.Logger,
	config Config,
) (Signer, error) {
	switch config.Type {
	case "horcrux":
		return NewHorcrux(logger, config)
	}

	return nil, fmt.Errorf("type not supported")
}
