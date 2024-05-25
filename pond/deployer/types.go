package deployer

import "encoding/json"

type (
	Code struct {
		Id       string `json:"id"`
		Name     string `json:"name"`
		Code     []byte `json:"-"`
		Source   string `json:"source,omitempty"`
		Checksum string `json:"checksum"`
	}

	Denom struct {
		Name  string `json:"name"`
		Path  string `json:"path"`
		Nonce string `json:"nonce"`
		Mint  string `json:"mint"`
	}

	Contract struct {
		Address string
		Name    string                     `json:"name"`
		Code    string                     `json:"code"`
		Label   string                     `json:"label"`
		Funds   string                     `json:"funds"`
		Msg     map[string]json.RawMessage `json:"msg"`
		Creates []Denom                    `json:"creates"`
		Actions []Action                   `json:"actions"`
	}

	Funds struct {
		Denom  string `json:"denom"`
		Amount string `json:"amount"`
	}

	Action struct {
		Contract string          `json:"contract"`
		Msg      json.RawMessage `json:"msg"`
		Funds    string          `json:"funds"`
	}
)
