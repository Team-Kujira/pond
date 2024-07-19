package pond

import (
	"encoding/json"
	"fmt"
	"time"
)

type Block struct{}

func (p *Pond) Upgrade(version, binary string) error {
	if version == "" {
		return fmt.Errorf("no version provided")
	}

	if binary == "" {
		return fmt.Errorf("upgrade only supported for local binaries")
	}

	chain := p.chains[0]
	node := chain.Nodes[0]

	// Get voting period
	var params struct {
		Params struct {
			VotingPeriod string `json:"voting_period"`
		} `json:"params"`
	}

	output, err := node.Query([]string{"gov", "params", "--output", "json"})
	if err != nil {
		return p.error(err)
	}

	err = json.Unmarshal(output, &params)
	if err != nil {
		return p.error(err)
	}

	period, err := time.ParseDuration(params.Params.VotingPeriod)
	if err != nil {
		return p.error(err)
	}

	height, err := chain.GetHeight()
	if err != nil {
		return err
	}

	if height < 2 {
		amount := 2 - height
		p.logger.Info().Int64("blocks", amount).Msg("wait for blocks")
		chain.WaitBlocks(2 - height)
	}

	// Get block time

	blockTime, err := chain.GetBlockTime(20)
	if err != nil {
		return err
	}

	blocks := (period.Milliseconds() / blockTime.Milliseconds())

	upgradeHeight := height + blocks + 10

	prop := []byte(fmt.Sprintf(`
	{
		"messages": [
			{
				"@type": "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
				"authority": "kujira10d07y265gmmuvt4z0w9aw880jnsr700jt23ame",
				"plan": {
					"name": "%s",
					"time": "0001-01-01T00:00:00Z",
					"height": "%d",
					"info": "",
					"upgraded_client_state": null
				}
			}
		],
		"metadata": "ipfs://CID",
		"deposit": "100000000ukuji",
		"title": "%s",
		"summary": "%s"
	}`, version, upgradeHeight, version, version))

	p.logger.Info().
		Str("version", version).
		Int64("height", upgradeHeight).
		Msg("submit upgrade proposal")

	chain.SubmitProposal(prop, "yes")

	var remain int64 = -1
	for height < upgradeHeight {
		if upgradeHeight-height != remain {
			remain = upgradeHeight - height

			p.logger.Info().
				Int64("blocks", remain).
				Msg("waiting for upgrade height")
		}

		time.Sleep(time.Millisecond * 300)
		height, err = chain.GetHeight()
		if err != nil {
			return err
		}
	}

	p.Stop()

	p.config.Binary = binary

	for i := range p.chains[0].Nodes {
		p.chains[0].Nodes[i].Binary = binary
	}

	err = p.SaveConfig()
	if err != nil {
		return err
	}

	p.Start()

	return nil
}
