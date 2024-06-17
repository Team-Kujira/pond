package chain

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"pond/pond/chain/node"
	"pond/pond/globals"
	"pond/utils"
)

func (c *Chain) Init(namespace string, options map[string]string) error {
	c.logger.Info().Msg("init chain")

	// Prepare gentx dir for the first node, in case it doesn't finish its
	// init process before other nodes
	os.MkdirAll(c.Nodes[0].Home+"/config/gentx", 0o755)

	version, found := globals.Versions[c.Type]
	if !found {
		err := fmt.Errorf("version not found")
		c.logger.Err(err).
			Str("type", c.Type).
			Msg("")
		return err
	}

	image := fmt.Sprintf("docker.io/%s/%s:%s", namespace, c.Type, version)

	amount := 10_000_000_000_000

	var wg sync.WaitGroup

	for i := range c.Feeders {
		wg.Add(1)
		go func(i int) {
			err := c.Feeders[i].Init(namespace)
			if err != nil {
				wg.Done()
				return
			}
			wg.Done()
		}(i)
	}

	for i := range c.Nodes {
		wg.Add(1)

		go func(i int) {
			if !c.Nodes[i].Local {
				err := c.Nodes[i].CreateInitContainer(image)
				if err != nil {
					wg.Done()
					return
				}

				err = c.Nodes[i].Start()
				if err != nil {
					wg.Done()
					return
				}

				c.WaitForNode(c.Nodes[i].Moniker)
			}

			err := c.Nodes[i].Init(namespace, amount)
			if err != nil {
				wg.Done()
				return
			}

			if i == 0 {
				wg.Done()
				return
			}

			src := c.Nodes[i].Home + "/config/gentx/gentx-" + c.Nodes[i].NodeId + ".json"
			dest := c.Nodes[0].Home + "/config/gentx/gentx-" + c.Nodes[i].NodeId + ".json"

			err = utils.CopyFile(c.logger, src, dest)
			if err != nil {
				wg.Done()
				return
			}

			wg.Done()
		}(i)
	}

	wg.Wait()

	mnemonics := map[string]string{}
	for wallet, mnemonic := range globals.Mnemonics {
		if strings.HasPrefix(wallet, "validator") {
			continue
		}

		mnemonics[wallet] = mnemonic
	}

	err := c.Nodes[0].AddKeys(mnemonics)
	if err != nil {
		return err
	}

	wallets, err := c.Nodes[0].GetAddresses()
	if err != nil {
		return err
	}

	c.Addresses = wallets

	accounts := []node.Account{}
	for wallet, address := range wallets {
		// this has already been added in node.Init()
		if wallet == "validator" {
			continue
		}

		accounts = append(accounts, node.Account{
			Address: address,
			Amount:  amount,
		})
	}

	// add remaining validator accounts
	for i := 1; i < len(c.Nodes); i++ {
		accounts = append(accounts, node.Account{
			Address: c.Nodes[i].Address,
			Amount:  amount,
		})
	}

	err = c.Nodes[0].AddGenesisAccounts(accounts)
	if err != nil {
		return err
	}

	err = c.Nodes[0].CollectGentxs()
	if err != nil {
		return err
	}

	err = c.UpdateGenesis(options)
	if err != nil {
		return err
	}

	for i := 1; i < len(c.Nodes); i++ {
		src := c.Nodes[0].Home + "/config/genesis.json"
		dst := c.Nodes[i].Home + "/config/genesis.json"

		err := utils.CopyFile(c.logger, src, dst)
		if err != nil {
			return err
		}
	}

	// Deploy configs and create run containers

	for i := range c.Nodes {
		if !c.Nodes[i].Local {
			wg.Add(1)
			go func(i int) {
				err := c.Nodes[i].CreateRunContainer(image)
				if err != nil {
					wg.Done()
					c.error(err)
				}
				wg.Done()
			}(i)
		}

		wg.Add(1)
		go func(i int) {
			for _, name := range []string{"app", "config", "client"} {
				c.logger.Debug().
					Str("node", c.Nodes[i].Moniker).
					Str("file", name).
					Msg("deploy config")

				// set peers
				peers := []string{}
				for _, node := range c.Nodes {
					if c.Nodes[i].Moniker == node.Moniker {
						continue
					}

					peers = append(peers, fmt.Sprintf(
						"%s@%s:%s", node.NodeId, node.Host, node.Ports.App,
					))
				}

				c.Nodes[i].Peers = strings.Join(peers, ",")

				src := fmt.Sprintf("config/%s/%s.toml", c.Nodes[i].Type, name)
				dst := fmt.Sprintf("%s/config/%s.toml", c.Nodes[i].Home, name)

				err = utils.Template(src, dst, c.Nodes[i])
				if err != nil {
					wg.Done()
					c.error(err)
				}
			}

			wg.Done()
		}(i)
	}

	wg.Wait()

	return nil
}
