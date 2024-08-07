package chain

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"pond/pond/chain/feeder"
	"pond/pond/chain/node"
	"pond/pond/globals"
	"pond/pond/templates"
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

type Block struct {
	Header struct {
		Time time.Time `json:"time"`
	} `json:"header"`
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

func (c *Chain) UpdateGenesis(overrides []byte) error {
	c.logger.Debug().Msg("update genesis")

	node := c.Nodes[0]

	filename := node.Home + "/config/genesis.json"

	genesis, err := os.ReadFile(filename)
	if err != nil {
		return c.error(err)
	}

	version, found := globals.Versions[node.Type]
	if !found {
		return c.error(fmt.Errorf("version not found"))
	}

	keys := []string{"default", node.Type, node.Type + "-" + version}
	for _, key := range keys {
		src := fmt.Sprintf("genesis/%s.json", key)
		content, err := templates.Templates.ReadFile(src)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return c.error(err)
		}

		genesis, err = utils.JsonMerge(genesis, content)
		if err != nil {
			return c.error(err)
		}
	}

	genesis, err = utils.JsonMerge(genesis, overrides)
	if err != nil {
		return c.error(err)
	}

	os.WriteFile(filename, genesis, 0o666)

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
		// TODO: < sdk-50
		"gov", "proposals", "--status", "voting_period", "--reverse", "--output", "json",
		// "gov", "proposals", "--output", "json", "--page-reverse",
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

func (c *Chain) GetBlock(height int64) (Block, error) {
	var block Block
	// < sdk-50
	var response struct {
		Block Block `json:"block"`
	}

	node := c.Nodes[0]

	args := []string{
		// --output param only available > sdk-50
		// "block", "--output", "json", "--type", "height", fmt.Sprint(height),
		"block", fmt.Sprint(height),
	}

	output, err := node.Query(args)
	if err != nil {
		fmt.Println(string(output))
		return block, err
	}

	err = json.Unmarshal(output, &response)
	if err != nil {
		return block, err
	}

	// < sdk-50
	block = response.Block

	return block, nil
}

func (c *Chain) GetBlockTime(interval int64) (time.Duration, error) {
	c.logger.Info().Msg("calculate block time")

	height, err := c.GetHeight()
	if err != nil {
		return -1, c.error(err)
	}

	if height < 2 {
		return -1, c.error(fmt.Errorf("height < 2"))
	}

	block, err := c.GetBlock(height)
	if err != nil {
		return -1, c.error(err)
	}

	timestamp1 := block.Header.Time

	if height > interval {
		height = height - interval
	} else {
		interval = height - 1
		height = 1
	}

	block, err = c.GetBlock(height)
	if err != nil {
		return -1, c.error(err)
	}

	timestamp0 := block.Header.Time

	blockTime := timestamp1.Sub(timestamp0) / time.Duration(interval)

	return blockTime, nil
}
