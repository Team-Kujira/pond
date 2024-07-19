package pond

import (
	"encoding/json"
	"fmt"
	"os"

	"pond/pond/chain"
)

type Config struct {
	Command   string            `json:"command"`
	Namespace string            `json:"namespace"`
	Versions  map[string]string `json:"versions"`
	Chains    []chain.Config    `json:"chains"`
	Plans     []string          `json:"plans"`
	ApiUrl    string            `json:"api_url"`
	RpcUrl    string            `json:"rpc_url"`
	Address   string            `json:"address"`
	Binary    string            `json:"binary"`
}

func (p *Pond) LoadConfig() error {
	filename := p.home + "/config.json"

	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		p.logger.Err(err).Msg("")
		return err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		p.logger.Err(err).Msg("")
		return err
	}

	p.config = config

	return nil
}

func (p *Pond) SaveConfig() error {
	p.logger.Debug().Msg("save config")

	filename := p.home + "/config.json"

	data, err := json.Marshal(p.config)
	if err != nil {
		return p.error(err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return p.error(err)
	}

	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return p.error(err)
	}

	return nil
}

func (p *Pond) GetVersion(name string) (string, error) {
	version, found := p.config.Versions[name]
	if found {
		return version, nil
	}

	err := fmt.Errorf("version not found")
	p.logger.Err(err).Str("name", name)
	return "", err
}
