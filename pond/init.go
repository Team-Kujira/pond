package pond

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"pond/pond/chain"
	"pond/pond/chain/node"
	"pond/pond/globals"
	"pond/pond/templates"
)

func (p *Pond) Init(
	config Config,
	chains []string,
	options map[string]string,
) error {
	p.logger.Info().Msg("init pond")

	local := false
	if config.Binary != "" {
		local = true
	}

	_, err := os.Stat(p.home)
	if err == nil {
		var input string
		for {
			fmt.Printf("Delete existing chain data and init new chains? [y/N] ")
			fmt.Scanln(&input)
			input = strings.ToLower(input)
			if input == "" || input == "n" || input == "no" {
				return nil
			}

			if input == "y" || input == "yes" {
				os.RemoveAll(p.home)
				break
			}
		}
	}

	p.info = Info{
		Validators: map[string][]node.Node{},
		Accounts:   map[string]Account{},
	}

	p.config = config

	types := map[string]int{
		"kujira": 1,
	}
	for _, name := range chains {
		number, found := types[name]
		if !found {
			number = 0
		}
		number += 1

		p.config.Chains = append(p.config.Chains, chain.Config{
			Type: name, TypeNum: uint(number), Nodes: 1,
		})

		types[name] = number
	}

	p.Clear()

	err = os.MkdirAll(p.home+"/planfiles", 0o755)
	if err != nil {
		return p.error(err)
	}

	// deploy registry.json

	data, err := json.Marshal(globals.Registry)
	if err != nil {
		p.error(err)
	}

	err = os.WriteFile(p.home+"/registry.json", data, 0o644)
	if err != nil {
		p.error(err)
	}

	plans, err := templates.GetPlans()
	if err != nil {
		p.error(err)
	}
	for _, plan := range plans {
		src := fmt.Sprintf("plan/%s.json", plan)
		dst := fmt.Sprintf("%s/planfiles/%s.json", p.home, plan)

		content, err := templates.Templates.ReadFile(src)
		if err != nil {
			return p.error(err)
		}

		err = os.WriteFile(dst, content, 0o644)
		if err != nil {
			return p.error(err)
		}
	}

	// remove all current chains
	p.chains = nil

	err = p.CreateNetwork()
	if err != nil {
		return err
	}

	err = p.init()
	if err != nil {
		return err
	}

	var mtx sync.Mutex
	var wg sync.WaitGroup

	for i := range p.chains {
		wg.Add(1)
		go func(i int) {
			p.chains[i].Init(p.config.Namespace, options)

			mtx.Lock()
			p.info.Validators[p.chains[i].ChainId] = p.chains[i].Nodes
			mtx.Unlock()
			wg.Done()
		}(i)
	}

	wg.Add(1)
	go func() {
		p.proxy.Init(p.config.Namespace, local)
		wg.Done()
	}()

	wg.Wait()

	if len(p.chains) > 1 {
		p.relayer.Init(p.config.Namespace)
	}

	p.info.Accounts = map[string]Account{}

	for _, chain := range p.chains {
		for name, address := range chain.Addresses {
			if name == "validator" {
				continue
			}
			mnemonic, found := globals.Mnemonics[name]
			if !found {
				err = fmt.Errorf("mnemonic not found")
				p.logger.Err(err).Str("wallet", name).Msg("")
				return err
			}
			_, found = p.info.Accounts[name]
			if !found {
				p.info.Accounts[name] = Account{
					Addresses: map[string]string{},
					Mnemonic:  mnemonic,
				}
			}

			p.info.Accounts[name].Addresses[chain.Type] = address
		}
	}

	p.SaveConfig()
	p.SaveInfo()

	return nil
}
