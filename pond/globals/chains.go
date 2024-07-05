package globals

type Chain struct {
	Denom   string
	Command string
	Prefix  string
	Home    string
}

var Chains = map[string]Chain{
	"kujira": {
		Denom:   "ukuji",
		Command: "kujirad",
		Prefix:  "kujira",
		Home:    ".kujira",
	},
	"cosmoshub": {
		Denom:   "uatom",
		Command: "gaiad",
		Prefix:  "cosmos",
		Home:    ".gaia",
	},
	"terra2": {
		Denom:   "uluna",
		Command: "terrad",
		Prefix:  "terra",
		Home:    ".terra",
	},
	"dydx": {
		Denom:   "adydx",
		Command: "dydxprotocold",
		Prefix:  "dydx",
		Home:    ".dydxprotocol",
	},
}
