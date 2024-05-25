package pond

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"pond/pond/chain/node"
	"pond/pond/deployer"
)

type Account struct {
	Addresses map[string]string `json:"addresses"`
	Mnemonic  string            `json:"mnemonic"`
}

type Contract struct {
	Address string `json:"address"`
	CodeId  string `json:"code_id"`
	Label   string `json:"label"`
}

type Info struct {
	Validators map[string][]node.Node `json:"validators"`
	Accounts   map[string]Account     `json:"accounts"`
	Codes      []deployer.Code        `json:"codes"`
	Contracts  []Contract             `json:"contracts"`
}

func (p *Pond) LoadInfo() error {
	filename := p.home + "/info.json"

	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		p.logger.Err(err).Msg("")
		return err
	}

	var info Info
	err = json.Unmarshal(data, &info)
	if err != nil {
		return err
	}

	p.info = info

	return nil
}

func (p *Pond) SaveInfo() error {
	p.logger.Debug().Msg("save info")

	filename := p.home + "/info.json"

	data, err := json.Marshal(p.info)
	if err != nil {
		p.logger.Err(err).Msg("")
		return err
	}

	os.WriteFile(filename, data, 0o666)

	return nil
}

func LoadInfo() (info Info, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return info, err
	}

	filename := home + "/.pond/info.json"

	_, err = os.Stat(filename)
	if os.IsNotExist(err) {
		return info, err
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return info, err
	}

	err = json.Unmarshal(data, &info)
	if err != nil {
		return info, err
	}

	return info, nil
}

func (i *Info) ListAccounts() error {
	padChain := 0
	padName := 0

	for name, account := range i.Accounts {
		if len(name) > padName {
			padName = len(name)
		}

		if padChain > 0 {
			continue
		}

		for chain := range account.Addresses {
			if len(chain) > padChain {
				padChain = len(chain)
			}
		}
	}

	lines := []string{}

	for name, account := range i.Accounts {
		for chain, address := range account.Addresses {
			lines = append(lines, fmt.Sprintf(
				"%-*s %-*s %s", padChain, chain, padName, name, address,
			))
		}
	}

	sort.Strings(lines)

	fmt.Printf("%-*s %-*s %s\n", padChain, "chain", padName, "name", "address")
	for _, line := range lines {
		fmt.Println(line)
	}

	return nil
}

func (i *Info) ListSeed(name string) error {
	account, found := i.Accounts[name]
	if !found {
		return fmt.Errorf("account not found")
	}

	fmt.Println(account.Mnemonic)

	return nil
}

func (i *Info) ListCodes() error {
	lines := []string{}

	if len(i.Codes) == 0 {
		return nil
	}

	padding := len(fmt.Sprintf("%d", len(i.Codes)))
	if padding == 1 {
		padding = 2
	}

	fmt.Printf("%*s checksum          name\n", padding, "id")

	for _, code := range i.Codes {
		checksum := code.Checksum[:8] + "…" + code.Checksum[56:]

		lines = append(lines, fmt.Sprintf(
			"%*s %s %s", padding, code.Id, checksum, code.Name,
		))
	}

	sort.Strings(lines)

	for _, line := range lines {
		fmt.Println(line)
	}

	return nil
}

func (i *Info) ListContracts() error {
	lines := []string{}

	padding := len(fmt.Sprintf("%d", len(i.Codes)))
	if padding == 1 {
		padding = 2
	}

	contracts := i.Contracts

	if len(contracts) == 0 {
		return nil
	}

	fmt.Printf("%*s %-*s label\n", padding, "id", 65, "address")

	sort.Slice(contracts, func(i, j int) bool {
		if contracts[i].CodeId != contracts[j].CodeId {
			if len(contracts[i].CodeId) == len(contracts[j].CodeId) {
				return contracts[i].CodeId < contracts[j].CodeId
			} else {
				return len(contracts[i].CodeId) < len(contracts[j].CodeId)
			}
		}
		return contracts[i].Label < contracts[j].Label
	})

	for _, contract := range contracts {
		lines = append(lines, fmt.Sprintf(
			"%*s %s %s", padding, contract.CodeId, contract.Address, contract.Label,
		))
	}

	for _, line := range lines {
		fmt.Println(line)
	}

	return nil
}

func (i *Info) ListUrls() error {
	for chain, nodes := range i.Validators {
		for _, node := range nodes {
			fmt.Printf("%s\n", node.Moniker)
			fmt.Printf(" %-6s %s\n", "api", node.ApiUrl)
			fmt.Printf(" %-6s %s\n", "rpc", node.RpcUrl)
			fmt.Printf(" %-6s %s\n", "grpc", node.GrpcUrl)
			if chain == "kujira-1" {
				fmt.Printf(" %-6s %s\n", "feeder", node.FeederUrl)
			}
		}
	}

	return nil
}
