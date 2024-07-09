package chain

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"pond/pond/chain/feeder"
	"pond/pond/chain/node"
	"pond/pond/globals"
	"pond/utils"

	"github.com/rs/zerolog"
)

type Chain struct {
	logger    zerolog.Logger
	Nodes     []node.Node
	Feeders   []feeder.Feeder
	Command   string
	Type      string
	ChainId   string
	Addresses map[string]string
	Signers   []string
}

type Config struct {
	Type    string   `json:"type"`     // ex.: kujira
	TypeNum uint     `json:"type_num"` // ex.: 1
	Nodes   uint     `json:"nodes"`    // ex.: 2
	Signers []string `json:"signers"`  // ex.: ["local", "horcrux"]
}

func NewChain(
	logger zerolog.Logger,
	command, binary, namespace, address string,
	// typeNum, numNodes, chainNum uint,
	config Config,
	chainNum uint,
) (Chain, error) {
	chainId := fmt.Sprintf("%s-%d", config.Type, config.TypeNum)

	logger = logger.With().
		Str("chain", chainId).
		Logger()

	chain := Chain{
		logger:  logger,
		Type:    config.Type,
		Nodes:   make([]node.Node, config.Nodes),
		Feeders: []feeder.Feeder{},
		ChainId: chainId,
		Command: command,
		Signers: config.Signers,
	}

	for i := 0; i < len(chain.Nodes); i++ {
		var signer string
		if len(config.Signers) > i {
			switch config.Signers[i] {
			case "horcrux":
				signer = config.Signers[i]
			}
		}

		node, err := node.NewNode(
			logger, command, binary, address,
			config.Type, config.TypeNum, uint(i+1), chainNum, node.Config{
				Signer: signer,
			},
		)
		if err != nil {
			logger.Err(err).Msg("")
			return Chain{}, err
		}

		chain.Nodes[i] = node

		if chainId == "kujira-1" {
			feeder, err := feeder.NewFeeder(logger, command, address, chainNum, uint(i+1))
			if err != nil {
				logger.Err(err).Msg("")
				return Chain{}, err
			}

			chain.Feeders = append(chain.Feeders, feeder)
		}
	}

	return chain, nil
}

func (c *Chain) Start() error {
	var wg sync.WaitGroup
	for i := range c.Nodes {
		wg.Add(1)
		go func(i int) {
			c.Nodes[i].Start()
			wg.Done()
		}(i)
	}

	for i := range c.Feeders {
		wg.Add(1)
		go func(i int) {
			c.Feeders[i].Start()
			wg.Done()
		}(i)
	}

	wg.Wait()

	return nil
}

func (c *Chain) Stop() error {
	var wg sync.WaitGroup
	for i := range c.Nodes {
		wg.Add(1)
		go func(i int) {
			c.Nodes[i].Stop()
			wg.Done()
		}(i)
	}

	for i := range c.Feeders {
		wg.Add(1)
		go func(i int) {
			c.Feeders[i].Stop()
			wg.Done()
		}(i)
	}

	wg.Wait()

	return nil
}

// func (c *Chain) Clear() error {
// 	var wg sync.WaitGroup
// 	for i := range c.Nodes {
// 		wg.Add(1)
// 		go func(i int) {
// 			c.Nodes[i].RemoveContainer()
// 			wg.Done()
// 		}(i)
// 	}

// 	for i := range c.Feeders {
// 		wg.Add(1)
// 		go func(i int) {
// 			c.Feeders[i].RemoveContainer()
// 			wg.Done()
// 		}(i)
// 	}

// 	wg.Wait()

// 	return nil
// }

func (c *Chain) error(err error) error {
	c.logger.Err(err).Msg("")
	return err
}

func (c *Chain) UpdateGenesis(overrides map[string]string) error {
	denoms := []string{
		"ADA",
		"ARB",
		"AVAX",
		"BNB",
		"BOME",
		"BOND",
		"BONK",
		"BTC",
		"CRV",
		"DOGE",
		"DOT",
		"DYM",
		"ENA",
		"ETH",
		"ETHFI",
		"FET",
		"FIL",
		"FLOKI",
		"INJ",
		"JUP",
		"KUJI",
		"LDO",
		"LINK",
		"LISTA",
		"LTC",
		"MATIC",
		"MKR",
		"NEAR",
		"NOT",
		"ORDI",
		"PENDLE",
		"PEOPLE",
		"PEPE",
		"RNDR",
		"RUNE",
		"SAGA",
		"SEI",
		"SHIB",
		"SOL",
		"STETH",
		"STRK",
		"SUI",
		"TAO",
		"TIA",
		"TRX",
		"USDC",
		"USDT",
		"USDT",
		"USK",
		"WIF",
		"WLD",
		"XAI",
		"XRP",
		"ZEN",
		"FTM",
		"ICP",
		"EGLD",
		"JTO",
		"REZ",
		"APT",
		"UNI",
		"YGG",
		"PIXEL",
		"ETC",
		"AEVO",
		"MANTA",
		"BETA",
		"ATOM",
		"PYTH",
		"ARKM",
		"EOS",
		"CHR",
		"DYDX",
		"BAKE",
		"CFX",
		"SSV",
		"VET",
		"ELF",
		"ASTR",
		"ALT",
		"TNSR",
		"MDX",
		"EDU",
		"STX",
		"AXL",
		"HBAR",
	}
	config := map[string]map[string]interface{}{
		"_default": {
			"app_state/gov/params/max_deposit_period": "120s",
			"app_state/gov/params/voting_period":      "120s",
		},
		"kujira": {
			"app_state/mint/minter/inflation": "0.0",
		},
		"kujira-99f7924-2": {
			"app_state/oracle/params/required_denoms":             denoms,
			"consensus/params/abci/vote_extensions_enable_height": "1",
		},
	}

	node := c.Nodes[0]

	filename := node.Home + "/config/genesis.json"

	data, err := os.ReadFile(filename)
	if err != nil {
		return c.error(err)
	}

	var root interface{}
	err = json.Unmarshal(data, &root)
	if err != nil {
		return c.error(err)
	}

	version, found := globals.Versions[node.Type]
	if !found {
		return fmt.Errorf("version not found")
	}
	keys := []string{"_default", node.Type, node.Type + "-" + version}
	for _, key := range keys {
		values, found := config[key]
		if !found {
			continue
		}

		for path, value := range values {
			err = utils.JsonReplace(root, path, value)
			if err != nil {
				return c.error(err)
			}
		}
	}

	for path, value := range overrides {
		err = utils.JsonReplace(root, path, value)
		if err != nil {
			return c.error(err)
		}
	}

	out, err := json.Marshal(root)
	if err != nil {
		return c.error(err)
	}

	os.WriteFile(filename, out, 0o666)

	return nil
}

func (c *Chain) GetHeight() (int64, error) {
	c.logger.Debug().Msg("get height")

	type Status struct {
		SyncInfo struct {
			LatestBlockHeight string `json:"latest_block_height"`
		} `json:"sync_info"`
		SyncInfoOld struct {
			LatestBlockHeight string `json:"latest_block_height"`
		} `json:"SyncInfo"`
	}

	var status Status

	output, err := c.Nodes[0].Status()
	if err != nil {
		return -1, c.error(err)
	}

	err = json.Unmarshal(output, &status)
	if err != nil {
		return -1, c.error(err)
	}

	strHeight := status.SyncInfoOld.LatestBlockHeight
	if strHeight == "" {
		strHeight = status.SyncInfo.LatestBlockHeight
	}

	height, err := strconv.ParseInt(strHeight, 10, 64)
	if err != nil {
		return -1, c.error(err)
	}

	return height, nil
}

func (c *Chain) WaitBlocks(amount int64) error {
	c.logger.Debug().Int64("blocks", amount).Msg("wait")

	current, err := c.GetHeight()
	if err != nil {
		return nil
	}

	target := current + amount

	for current < target {
		time.Sleep(time.Millisecond * 500)

		current, err = c.GetHeight()
		if err != nil {
			return nil
		}
	}

	return nil
}

func (c *Chain) SubmitProposal(data []byte, option string) error {
	node := c.Nodes[0]

	filename, err := node.CreateTemp(data, "json")
	if err != nil {
		return err
	}

	args := []string{
		"gov", "submit-proposal", filename,
		"--from", "validator", "--gas", "auto", "--gas-adjustment", "1.5",
	}

	output, err := node.Tx(args)
	if err != nil {
		return err
	}

	hash, err := utils.CheckTxResponse(output)
	if err != nil {
		return c.error(err)
	}

	err = node.WaitForTx(hash)
	if err != nil {
		return err
	}

	if option == "" {
		return nil
	}

	// get the latest proposal

	args = []string{
		// "gov", "proposals", "--status", "voting_period", "--reverse",
		"gov", "proposals", "--output", "json", "--page-reverse",
	}

	output, err = node.Query(args)
	if err != nil {
		fmt.Println(string(output))
		return err
	}

	type ProposalResponse struct {
		Proposals []struct {
			Id string `json:"id"`
		} `json:"proposals"`
	}

	var response ProposalResponse

	err = json.Unmarshal(output, &response)
	if err != nil {
		return c.error(err)
	}

	if len(response.Proposals) == 0 {
		err := fmt.Errorf("no proposal found")
		return c.error(err)
	}

	proposal := response.Proposals[0].Id

	var wg sync.WaitGroup

	for i := range c.Nodes {
		wg.Add(1)

		go func(i int) {
			args := []string{
				"gov", "vote", proposal, option, "--from", "validator",
				"--gas", "auto", "--gas-adjustment", "1.5",
			}

			output, err := c.Nodes[i].Tx(args)
			if err != nil {
				wg.Done()
				return
			}

			hash, err := utils.CheckTxResponse(output)
			if err != nil {
				wg.Done()
				return
			}

			c.Nodes[i].WaitForTx(hash)

			wg.Done()
		}(i)
	}

	wg.Wait()

	return nil
}

func (c *Chain) WaitForNode(name string) error {
	c.logger.Debug().Str("node", name).Msg("wait for node")

	command := []string{
		c.Command, "inspect", "--format", "'{{ json .State.Running }}'", name,
	}
	c.logger.Debug().Msg(strings.Join(command, " "))

	retries := 10
	running := false
	for i := 0; i < retries; i++ {
		output, err := utils.RunO(c.logger, command)
		if err != nil {
			return err
		}

		if strings.Contains(string(output), "true") {
			running = true
			break
		}

		time.Sleep(time.Millisecond * 200)
	}

	if !running {
		msg := "node not running"
		c.logger.Error().Str("node", name).Msg(msg)
		return fmt.Errorf(msg)
	}

	return nil
}
