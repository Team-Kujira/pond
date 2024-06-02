package deployer

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"pond/pond/chain/node"
	"pond/utils"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v2"
)

type Deployer struct {
	logger    zerolog.Logger
	node      node.Node
	Denoms    map[string]Denom
	Contracts map[string]Contract
	CodeIds   map[string]string
	addresses map[string]struct{}
	codes     map[string]string
	plan      Plan
	address   string
	registry  map[string]Code
	apiUrl    string
	home      string
	accounts  []string // test accounts for minting assets
}

type Plan struct {
	Denoms    []Denom      `yaml:"denoms"`
	Contracts [][]Contract `yaml:"contracts"`
	Names     []string     // holds plan file names, used only for logging
}

type CodeMsg struct {
	Type       string           `json:"@type"`
	Sender     string           `json:"sender"`
	Code       string           `json:"wasm_byte_code"`
	Permission *json.RawMessage `json:"instantiate_permission"`
}

type DenomMsg struct {
	Type   string `json:"@type"`
	Sender string `json:"sender"`
	Nonce  string `json:"nonce"`
}

type ContractMsg struct {
	Type   string          `json:"@type"`
	Sender string          `json:"sender"`
	Admin  string          `json:"admin"`
	CodeId string          `json:"code_id"`
	Label  string          `json:"label"`
	Msg    json.RawMessage `json:"msg"`
	Funds  []Funds         `json:"funds"`
	Salt   string          `json:"salt"`
	FixMsg bool            `json:"fix_msg"`
}

type ActionMsg struct {
	Type     string          `json:"@type"`
	Sender   string          `json:"sender"`
	Contract string          `json:"contract"`
	Msg      json.RawMessage `json:"msg"`
	Funds    []Funds         `json:"funds"`
}

type MintMsg struct {
	Type      string `json:"@type"`
	Sender    string `json:"sender"`
	Amount    Funds  `json:"amount"`
	Recipient string `json:"recipient"`
}

func NewDeployer(
	logger zerolog.Logger,
	home string,
	node node.Node,
	apiUrl string,
	accounts []string,
) (Deployer, error) {
	logger.Debug().Msg("create deployer")

	deployer := Deployer{
		logger: logger,
		node:   node,
		plan: Plan{
			Denoms:    []Denom{},
			Contracts: [][]Contract{},
		},
		codes:     map[string]string{},
		Denoms:    map[string]Denom{},
		Contracts: map[string]Contract{},
		CodeIds:   map[string]string{},
		address:   "kujira1k3g54c2sc7g9mgzuzaukm9pvuzcjqy92nk9wse",
		addresses: map[string]struct{}{},
		apiUrl:    apiUrl,
		home:      home,
		accounts:  accounts,
	}

	err := deployer.LoadRegistry()
	if err != nil {
		return deployer, err
	}

	return deployer, nil
}

func (d *Deployer) LoadRegistry() error {
	data, err := os.ReadFile(d.home + "/registry.json")
	if err != nil {
		return err
	}

	var registry map[string]Code
	err = json.Unmarshal(data, &registry)
	if err != nil {
		return d.error(err)
	}

	d.registry = map[string]Code{}

	for name, code := range registry {
		d.registry[name] = Code{
			Source:   code.Source,
			Checksum: strings.ToUpper(code.Checksum),
		}
	}

	return nil
}

func (d *Deployer) Deploy(filenames []string) error {
	err := d.UpdateDeployedCodes()
	if err != nil {
		return d.error(err)
	}

	for _, filename := range filenames {
		info, err := os.Stat(filename)
		if err != nil {
			return d.error(err)
		}

		if info.IsDir() {
			return fmt.Errorf("not yet implemented")
		}

		file, err := os.Open(filename)
		if err != nil {
			return d.error(err)
		}

		defer file.Close()

		buf := make([]byte, 512)

		_, err = file.Read(buf)
		if err != nil {
			return d.error(err)
		}

		contentType := http.DetectContentType(buf)

		switch contentType {
		case "text/plain; charset=utf-8", "application/octet-stream":
			// return d.DeployPlanfiles([]string{filename})
			err := d.LoadPlanFile(filename)
			if err != nil {
				return err
			}
		case "application/wasm":
			err = d.DeployWasmFile(filename)
			if err != nil {
				return err
			}
		default:
			d.logger.Warn().
				Str("type", contentType).
				Msg("type not supported")
		}
	}

	if len(d.plan.Denoms)+len(d.plan.Contracts) == 0 {
		d.logger.Debug().Msg("no plan tasks found")
		return nil
	}

	return d.DeployPlan()
}

func (d *Deployer) DeployWasmFile(filename string) error {
	d.logger.Debug().Str("file", filename).Msg("deploy wasm")

	data, err := os.ReadFile(filename)
	if err != nil {
		return d.error(err)
	}

	id, deployed := d.getCodeId(data)

	if deployed {
		// already deployed -> done
		d.logger.Info().Str("code_id", id).Msg("code already deployed")
		return nil
	}

	err = d.DeployCode(data)
	if err != nil {
		return err
	}

	// err = d.UpdateDeployedCodes()
	// if err != nil {
	// 	return err
	// }

	name := filepath.Base(filename)

	return d.UpdateRegistry(name, "file://"+filename, data)
}

func (d *Deployer) DeployCode(data []byte) error {
	d.logger.Debug().Msg("deploy code")

	file, err := os.CreateTemp(d.node.Home, "wasm")
	if err != nil {
		return d.error(err)
	}
	defer os.Remove(file.Name())

	os.WriteFile(file.Name(), data, 0o644)

	args := []string{
		"wasm", "store", "/home/kujira/.kujira/" + filepath.Base(file.Name()),
		"--from", "deployer", "--gas", "auto", "--gas-adjustment", "1.5",
	}

	output, err := d.node.Tx(args)
	if err != nil {
		return err
	}

	hash, err := utils.CheckTxResponse(output)
	if err != nil {
		return d.error(err)
	}

	err = d.node.WaitForTx(hash)
	if err != nil {
		return d.error(err)
	}

	// update deployed codes
	err = d.UpdateDeployedCodes()
	if err != nil {
		return err
	}

	id, deployed := d.getCodeId(data)
	if !deployed {
		return fmt.Errorf("code not found")
	}
	d.logger.Info().Str("code_id", id).Msg("code deployed")

	return nil
}

func (d *Deployer) DeployPlan() error {
	d.logger.Debug().Msg("deploy plan")

	err := d.UpdateDeployedCodes()
	if err != nil {
		return d.error(err)
	}

	codes, err := d.GetMissingCodes()
	if err != nil {
		return d.error(err)
	}

	err = d.UpdateDeployedContracts()
	if err != nil {
		return err
	}

	// remove duplicate denoms

	denoms := map[string]Denom{}

	for _, denom := range d.plan.Denoms {
		denoms[denom.Name] = denom
	}

	available, err := d.GetDenomsFromCreator(d.address)
	if err != nil {
		return d.error(err)
	}

	nonces := map[string]struct{}{}
	for _, denom := range available {
		nonces[filepath.Base(denom)] = struct{}{}
	}

	d.plan.Denoms = []Denom{}

	for symbol, denom := range denoms {
		_, found := d.Denoms[symbol]
		if found {
			d.logger.Info().
				Str("symbol", symbol).
				Msg("denom already exists")
			continue
		}

		_, found = nonces[denom.Nonce]
		if found {
			denom.Path = fmt.Sprintf("factory/%s/%s", d.address, denom.Nonce)
			d.Denoms[denom.Name] = denom

			d.logger.Debug().
				Str("symbol", symbol).
				Msg("denom already exists")

			continue
		}

		d.plan.Denoms = append(d.plan.Denoms, denom)
	}

	denomMsgs, err := d.CreateDenomMsgs(d.plan.Denoms)
	if err != nil {
		return err
	}

	codeMsgs, err := d.CreateCodeMsgs(codes)
	if err != nil {
		return err
	}

	if len(codeMsgs) > 0 {
		d.logger.Info().Msg("deploy codes")
	}

	if len(denomMsgs) > 0 {
		d.logger.Info().Msg("create denoms")
	}

	combinedMsgs := append(codeMsgs, denomMsgs...)

	if len(combinedMsgs) > 0 {
		err = d.SignAndSend(combinedMsgs)
		if err != nil {
			return err
		}
	}

	if len(codeMsgs) > 0 {
		err := d.UpdateDeployedCodes()
		if err != nil {
			return err
		}
	}

	// update internal denom map
	for _, denom := range d.plan.Denoms {
		d.Denoms[denom.Name] = Denom{
			Name: denom.Name,
			Path: fmt.Sprintf("factory/%s/%s", d.address, denom.Nonce),
		}
	}

	for _, contracts := range d.plan.Contracts {
		for _, contract := range contracts {
			if len(contract.Creates) == 0 {
				continue
			}

			contract, found := d.Contracts[contract.Name]
			if !found {
				continue
			}

			for _, denom := range contract.Creates {
				path := fmt.Sprintf("factory/%s/%s", contract.Address, denom.Nonce)
				denom.Path = path
				d.Denoms[denom.Name] = denom
			}
		}
	}

	total := len(d.plan.Contracts)
	for i, contracts := range d.plan.Contracts {
		msgs, err := d.CreateContractMsgs(contracts)
		if err != nil {
			return err
		}

		step := fmt.Sprintf("%d/%d", i+1, total)

		if len(msgs) == 0 {
			d.logger.Info().
				Str("plan", d.plan.Names[i]).
				Str("step", step).
				Msg("contracts already deployed")
			continue
		}

		d.logger.Info().
			Str("plan", d.plan.Names[i]).
			Str("step", step).
			Msg("deploy contracts")

		err = d.SignAndSend(msgs)
		if err != nil {
			return err
		}

		for _, contract := range contracts {
			if len(contract.Creates) == 0 {
				continue
			}

			contract, found := d.Contracts[contract.Name]
			if !found {
				continue
			}

			for _, denom := range contract.Creates {
				path := fmt.Sprintf("factory/%s/%s", contract.Address, denom.Nonce)
				denom.Path = path
				d.Denoms[denom.Name] = denom
			}
		}
	}

	return nil
}

func (d *Deployer) getCodeId(data []byte) (string, bool) {
	hash := sha256.New()
	hash.Write(data)

	id, found := d.codes[fmt.Sprintf("%X", hash.Sum(nil))]

	return id, found
}

func (d *Deployer) LoadPlan(data []byte, name string) error {
	var plan Plan
	err := json.Unmarshal(data, &plan)
	if err != nil {
		return d.error(err)
	}

	// load denom tasks

	for _, denom := range plan.Denoms {
		if denom.Path != "" {
			d.Denoms[denom.Name] = denom
			continue
		}

		d.plan.Denoms = append(d.plan.Denoms, denom)
	}

	d.plan.Contracts = append(d.plan.Contracts, plan.Contracts...)

	names := make([]string, len(d.plan.Contracts))
	for i := range names {
		names[i] = name
	}
	d.plan.Names = append(d.plan.Names, names...)

	return nil
}

func (d *Deployer) LoadPlanFile(filename string) error {
	d.logger.Debug().Str("file", filename).Msg("load plan file")

	content, err := os.ReadFile(filename)
	if err != nil {
		return d.error(err)
	}

	name := strings.Replace(filepath.Base(filename), ".json", "", -1)
	return d.LoadPlan(content, name)
}

func (d *Deployer) BuildAddress(hash, salt string) (string, error) {
	args := []string{"wasm", "build-address", hash, d.address, salt}

	output, err := d.node.Query(args)
	if err != nil {
		return "", err
	}

	address := strings.TrimSpace(string(output))

	return address, nil
}

func (d *Deployer) CreateCodeMsgs(codes []Code) ([]json.RawMessage, error) {
	d.logger.Debug().Msg("create code msgs")

	msgs := make([]json.RawMessage, len(codes))

	for i, code := range codes {
		msg := CodeMsg{
			Type:       "/cosmwasm.wasm.v1.MsgStoreCode",
			Sender:     d.address,
			Code:       base64.StdEncoding.EncodeToString(code.Code),
			Permission: nil,
		}

		data, err := json.Marshal(msg)
		if err != nil {
			return nil, d.error(err)
		}

		msgs[i] = data
	}

	return msgs, nil
}

func (d *Deployer) CreateDenomMsgs(denoms []Denom) ([]json.RawMessage, error) {
	d.logger.Debug().Msg("create denom msgs")

	msgs := []json.RawMessage{}

	for _, denom := range denoms {
		data, err := json.Marshal(DenomMsg{
			Type:   "/kujira.denom.MsgCreateDenom",
			Sender: d.address,
			Nonce:  denom.Nonce,
		})
		if err != nil {
			return nil, d.error(err)
		}

		msgs = append(msgs, data)

		if denom.Mint == "" {
			continue
		}

		for _, address := range d.accounts {
			data, err = json.Marshal(MintMsg{
				Type:   "/kujira.denom.MsgMint",
				Sender: d.address,
				Amount: Funds{
					Amount: strings.Replace(denom.Mint, "_", "", -1),
					Denom:  fmt.Sprintf("factory/%s/%s", d.address, denom.Nonce),
				},
				Recipient: address,
			})
			if err != nil {
				return nil, d.error(err)
			}

			msgs = append(msgs, data)
		}
	}

	return msgs, nil
}

func (d *Deployer) CreateActionMsg(action Action) (json.RawMessage, error) {
	d.logger.Debug().Msg("create action msg")

	data, err := json.Marshal(action)
	if err != nil {
		return nil, d.error(err)
	}

	tmpl, err := template.New("").Parse(string(data))
	if err != nil {
		return nil, d.error(err)
	}

	var buffer bytes.Buffer

	err = tmpl.Execute(&buffer, d)
	if err != nil {
		return nil, d.error(err)
	}

	err = json.Unmarshal(buffer.Bytes(), &action)
	if err != nil {
		return nil, err
	}

	funds, err := d.StringToFunds(action.Funds)
	if err != nil {
		return nil, err
	}

	msg, err := json.Marshal(ActionMsg{
		Type:     "/cosmwasm.wasm.v1.MsgExecuteContract",
		Sender:   d.address,
		Contract: action.Contract,
		Msg:      action.Msg,
		Funds:    funds,
	})
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (d *Deployer) CreateContractMsgs(
	contracts []Contract,
) ([]json.RawMessage, error) {
	d.logger.Debug().Msg("deploy contracts")

	msgs := []json.RawMessage{}

	for _, contract := range contracts {
		code, found := d.registry[contract.Code]
		if !found {
			err := fmt.Errorf("code not found in registry")
			d.logger.Err(err).Str("name", contract.Code).Msg("")
			return nil, err
		}

		id, found := d.codes[code.Checksum]
		if !found {
			err := fmt.Errorf("code not yet deployed")
			d.logger.Err(err).Str("checksum", code.Checksum).Msg("")
			return nil, err
		}

		_, found = contract.Msg["owner"]
		if !found {
			contract.Msg["owner"] = []byte(`"` + d.address + `"`)
		}

		hash := sha256.New()
		hash.Write([]byte(contract.Label))
		saltBase64 := base64.StdEncoding.EncodeToString(hash.Sum(nil))
		saltSha256 := utils.Sha256([]byte(contract.Label))

		address, err := d.BuildAddress(code.Checksum, saltSha256)
		if err != nil {
			return nil, err
		}

		contract.Address = address

		// append already, it can't be used if something breaks later
		d.Contracts[contract.Name] = contract

		_, found = d.addresses[address]
		if found {
			d.logger.Debug().
				Str("address", address).
				Msg("contract already deployed")
			continue
		}

		data, err := json.Marshal(contract.Msg)
		if err != nil {
			return nil, d.error(err)
		}

		tmpl, err := template.New(address).Parse(string(data))
		if err != nil {
			return nil, d.error(err)
		}

		var buffer bytes.Buffer

		err = tmpl.Execute(&buffer, d)
		if err != nil {
			return nil, d.error(err)
		}

		msg, err := d.Convert(buffer.Bytes())
		if err != nil {
			return nil, err
		}

		funds, err := d.StringToFunds(contract.Funds)
		if err != nil {
			return nil, err
		}

		data, err = json.Marshal(ContractMsg{
			Type:   "/cosmwasm.wasm.v1.MsgInstantiateContract2",
			Sender: d.address,
			Admin:  d.address,
			CodeId: id,
			Label:  contract.Label,
			Msg:    msg,
			Salt:   saltBase64,
			Funds:  funds,
			FixMsg: false,
		})
		if err != nil {
			return nil, d.error(err)
		}

		msgs = append(msgs, data)

		if len(contract.Actions) == 0 {
			continue
		}

		for _, action := range contract.Actions {
			msg, err := d.CreateActionMsg(action)
			if err != nil {
				return nil, err
			}

			msgs = append(msgs, msg)
		}
	}

	return msgs, nil
}

func (d *Deployer) CreateDenom(nonce string) error {
	d.logger.Info().Str("nonce", nonce).Msg("create denom")
	args := []string{"denom", "create-denom", nonce, "--from", "deployer"}

	output, err := d.node.Tx(args)
	if err != nil {
		return d.error(err)
	}

	hash, err := utils.CheckTxResponse(output)
	if err != nil {
		return d.error(err)
	}

	err = d.node.WaitForTx(hash)
	if err != nil {
		return err
	}

	return nil
}

func (d *Deployer) error(err error) error {
	d.logger.Err(err).Msg("")
	return err
}

func (d *Deployer) UpdateDeployedCodes() error {
	d.logger.Debug().Msg("get deployed codes")

	args := []string{"wasm", "list-code", "--output", "json"}

	var info struct {
		CodeInfos []struct {
			CodeId   string `json:"code_id"`
			DataHash string `json:"data_hash"`
		} `json:"code_infos"`
		Pagination struct {
			NextKey string `json:"next_key"`
		} `json:"pagination"`
	}

	key := "dummy"
	for key != "" {
		args := args
		if key != "dummy" {
			args = append(args, []string{"--page-key", key}...)
		}

		output, err := d.node.Query(args)
		if err != nil {
			return err
		}

		err = json.Unmarshal(output, &info)
		if err != nil {
			return err
		}

		for _, code := range info.CodeInfos {
			d.codes[code.DataHash] = code.CodeId
		}

		for name, code := range d.registry {
			codeId, found := d.codes[code.Checksum]
			if !found {
				continue
			}
			d.CodeIds[name] = codeId
		}

		key = info.Pagination.NextKey
	}

	return nil
}

func (d *Deployer) UpdateDeployedContracts() error {
	d.logger.Debug().Msg("update deployed contracts")

	args := []string{
		"wasm", "list-contracts-by-creator", d.address, "--output", "json",
	}

	var info struct {
		Addresses  []string `json:"contract_addresses"`
		Pagination struct {
			NextKey string `json:"next_key"`
		} `json:"pagination"`
	}

	key := "dummy"
	for key != "" {
		args := args
		if key != "dummy" {
			args = append(args, []string{"--page-key", key}...)
		}

		output, err := d.node.Query(args)
		if err != nil {
			return d.error(err)
		}

		err = json.Unmarshal(output, &info)
		if err != nil {
			return d.error(err)
		}

		for _, address := range info.Addresses {
			d.addresses[address] = struct{}{}
		}

		key = info.Pagination.NextKey
	}

	return nil
}

func (d *Deployer) UpdateRegistry(name, source string, code []byte) error {
	_, found := d.registry[name]
	if found {
		d.logger.Debug().Str("name", name).Msg("code already registered")
		return nil
	}

	d.registry[name] = Code{
		Checksum: utils.Sha256(code),
		Source:   source,
		Code:     code,
	}

	data, err := json.Marshal(d.registry)
	if err != nil {
		return d.error(err)
	}

	err = os.WriteFile(d.home+"/registry.json", data, 0o644)
	if err != nil {
		return d.error(err)
	}

	return nil
}

func (d *Deployer) SignAndSend(msgs []json.RawMessage) error {
	data, err := json.Marshal(msgs)
	if err != nil {
		return d.error(err)
	}

	msg, err := utils.NewTxMsg(data)
	if err != nil {
		return err
	}

	d.logger.Trace().Msg(string(data))

	unsigned, err := os.CreateTemp(d.node.Home, "tx")
	if err != nil {
		return d.error(err)
	}
	defer os.Remove(unsigned.Name())

	err = os.WriteFile(unsigned.Name(), msg, 0o644)
	if err != nil {
		return d.error(err)
	}

	os.WriteFile("/tmp/unsigned.json", msg, 0o644)

	// sign
	d.logger.Debug().Msg("sign tx")
	output, err := d.node.Tx([]string{
		"sign", "/home/kujira/.kujira/" + filepath.Base(unsigned.Name()),
		"--from", "deployer", "--gas", "1000000000",
	})
	if err != nil {
		return err
	}

	signed, err := os.CreateTemp(d.node.Home, "tx")
	if err != nil {
		return d.error(err)
	}
	defer os.Remove(signed.Name())

	err = os.WriteFile(signed.Name(), output, 0o644)
	if err != nil {
		return d.error(err)
	}

	// broadcast
	d.logger.Debug().Msg("broadcast tx")

	output, err = d.node.Tx([]string{
		"broadcast", "/home/kujira/.kujira/" + filepath.Base(signed.Name()),
		"--gas", "auto", "--gas-adjustment", "1.5",
	})
	if err != nil {
		return err
	}

	hash, err := utils.CheckTxResponse(output)
	if err != nil {
		return d.error(err)
	}

	err = d.node.WaitForTx(hash)
	if err != nil {
		return d.error(err)
	}

	return nil
}

func (d *Deployer) GetCode(code Code) ([]byte, error) {
	parts, err := url.Parse(code.Source)
	if err != nil {
		return nil, d.error(err)
	}

	var data []byte

	switch parts.Scheme {
	case "kaiyo-1":
		d.logger.Info().
			Str("code_id", parts.Host).
			Msg("download code from mainnet")

		data, err = d.GetCodeFromApi(parts.Host)
		if err != nil {
			return nil, d.error(err)
		}

	case "file":
		data, err = os.ReadFile(parts.Path)
		if err != nil {
			return nil, d.error(err)
		}
	default:
		err = fmt.Errorf("scheme not supported")
		d.logger.Err(err).Str("scheme", parts.Scheme).Msg("")
		return nil, err
	}

	if len(data) == 0 {
		err = fmt.Errorf("failed loading code")
		return nil, d.error(err)
	}

	// if no checksum is defined, don't try to check it
	if code.Checksum == "" {
		return data, nil
	}

	checksum := utils.Sha256(data)
	if !strings.EqualFold(checksum, code.Checksum) {
		err = fmt.Errorf("checksum mismatch")
		d.logger.Err(err).
			Str("source", code.Source).
			Str("checksum", code.Checksum).
			Msg("")
		return nil, err
	}

	return data, nil
}

func (d *Deployer) GetMissingCodes() ([]Code, error) {
	d.logger.Debug().Msg("get missing codes")
	missing := map[string]Code{}

	for _, contracts := range d.plan.Contracts {
		for _, contract := range contracts {
			code, found := d.registry[contract.Code]
			if !found {
				err := fmt.Errorf("code not found in registry")
				d.logger.Err(err).Str("code", contract.Code)
				return nil, err
			}

			_, found = d.codes[code.Checksum]
			if found {
				// already deployed
				continue
			}

			code.Name = contract.Code
			missing[contract.Code] = code
		}
	}

	var mtx sync.Mutex
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	codes := []Code{}

	for _, code := range missing {
		wg.Add(1)
		go func(code Code) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			default:
			}

			data, err := d.GetCode(code)
			if err != nil {
				cancel()
				return
			}

			code.Code = data

			mtx.Lock()
			codes = append(codes, code)
			mtx.Unlock()
		}(code)
	}

	wg.Wait()

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	return codes, nil
}

func (d *Deployer) GetCodeFromApi(codeId string) ([]byte, error) {
	url := d.apiUrl + "/cosmwasm/wasm/v1/code/" + codeId

	data, err := utils.HttpGet(d.logger, url)
	if err != nil {
		return nil, err
	}

	var response struct {
		Info struct {
			Hash string `json:"data_hash"`
		} `json:"code_info"`
		Data string `json:"data"`
	}

	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, d.error(err)
	}

	code, err := base64.StdEncoding.DecodeString(response.Data)
	if err != nil {
		return nil, d.error(err)
	}

	return code, nil
}

func (d *Deployer) GetDenomsFromCreator(address string) ([]string, error) {
	d.logger.Debug().Msg("get available denoms")

	args := []string{"denom", "denoms-from-creator", address}

	output, err := d.node.Query(args)
	if err != nil {
		return nil, d.error(err)
	}

	var response struct {
		Denoms []string `yaml:"denoms"`
	}

	err = yaml.Unmarshal(output, &response)
	if err != nil {
		return nil, d.error(err)
	}

	return response.Denoms, nil
}

func (d *Deployer) GetDeployedCodes() ([]Code, error) {
	codes := []Code{}
	names := map[string]string{}

	err := d.LoadRegistry()
	if err != nil {
		return nil, err
	}

	for name, code := range d.registry {
		fmt.Println(name)
		names[code.Checksum] = name
	}

	for checksum, id := range d.codes {
		name, found := names[checksum]
		if !found {
			d.logger.Warn().
				Str("checksum", checksum).
				Str("code_id", id).
				Msg("code not registered")
		}

		codes = append(codes, Code{
			Checksum: checksum,
			Name:     name,
			Id:       id,
		})
	}

	return codes, nil
}

func (d *Deployer) GetDeployedContracts() ([]Contract, error) {
	contracts := []Contract{}

	for _, contract := range d.Contracts {
		code, found := d.registry[contract.Code]
		if !found {
			err := fmt.Errorf("contract not registered")
			d.logger.Err(err).Str("name", contract.Code).Msg("")
			return nil, err
		}

		id, found := d.codes[code.Checksum]
		if !found {
			err := fmt.Errorf("code not deployed")
			d.logger.Err(err).Str("name", contract.Code).Msg("")
			return nil, err
		}

		contract.Code = id
		contracts = append(contracts, contract)
	}

	return contracts, nil
}

func (d *Deployer) StringToFunds(str string) ([]Funds, error) {
	funds := []Funds{}

	if str == "" {
		return funds, nil
	}

	tmpl, err := template.New("").Parse(str)
	if err != nil {
		return nil, d.error(err)
	}

	var buffer bytes.Buffer

	err = tmpl.Execute(&buffer, d)
	if err != nil {
		return nil, d.error(err)
	}

	str = buffer.String()

	regex := regexp.MustCompile(`^(\d+)([/A-Za-z0-9]+)$`)

	for _, part := range strings.Split(str, ",") {
		matches := regex.FindStringSubmatch(part)
		if len(matches) != 3 {
			return nil, d.error(fmt.Errorf("funds malformed"))
		}
		funds = append(funds, Funds{
			Amount: matches[1],
			Denom:  matches[2],
		})
	}

	sort.Slice(funds, func(i, j int) bool {
		return funds[i].Denom < funds[j].Denom
	})

	return funds, nil
}

func (d *Deployer) Convert(data []byte) ([]byte, error) {
	raw := json.RawMessage(data)

	data, err := json.Marshal(raw)
	if err != nil {
		return nil, d.error(err)
	}

	buffer := new(bytes.Buffer)
	err = json.Compact(buffer, data)
	if err != nil {
		return nil, d.error(err)
	}

	regex := regexp.MustCompile(`"((\d+)\s*\|\s*int\s*)"`)

	return []byte(regex.ReplaceAllString(string(data), "$2")), nil
}
