package relayer

type Config struct {
	Global Global           `yaml:"global"`
	Chains map[string]Chain `yaml:"chains"`
	Paths  map[string]Path  `yaml:"paths"`
}

type Global struct {
	ApiListenAddr   string `yaml:"api-listen-addr"`
	Timeout         string `yaml:"timeout"`
	Memo            string `yaml:"memo"`
	LightCacheSize  int    `yaml:"light-cache-size"`
	LogLevel        string `yaml:"log-level"`
	Ics20MemoLimit  int    `yaml:"ics20-memo-limit"`
	MaxReceiverSize int    `yaml:"max-receiver-size"`
}

func NewConfig(port string) Config {
	return Config{
		Global: Global{
			ApiListenAddr:   ":" + port,
			Timeout:         "10s",
			LightCacheSize:  20,
			LogLevel:        "info",
			MaxReceiverSize: 150,
		},
		Chains: map[string]Chain{},
		Paths:  map[string]Path{},
	}
}

type Chain struct {
	Type  string `yaml:"type"`
	Value struct {
		KeyDirectory     string      `yaml:"key-directory"`
		Key              string      `yaml:"key"`
		ChainId          string      `yaml:"chain-id"`
		RpcAddr          string      `yaml:"rpc-addr"`
		AccountPrefix    string      `yaml:"account-prefix"`
		KeyringBackend   string      `yaml:"keyring-backend"`
		GasAdjustment    float64     `yaml:"gas-adjustment"`
		GasPrices        string      `yaml:"gas-prices"`
		MinGasAmount     int         `yaml:"min-gas-amount"`
		MaxGasAmount     int         `yaml:"max-gas-amount"`
		Debug            bool        `yaml:"debug"`
		Timeout          string      `yaml:"timeout"`
		BlockTimeout     string      `yaml:"block-timeout"`
		OutputFormat     string      `yaml:"output-format"`
		SignMode         string      `yaml:"sign-mode"`
		ExtraCodecs      []string    `yaml:"extra-codecs"`
		CoinType         int         `yaml:"coin-type"`
		SigningAlgorithm string      `yaml:"signing-algorithm"`
		BroadcastMode    string      `yaml:"broadcast-mode"`
		MinLoopDuration  string      `yaml:"min-loop-duration"`
		ExtensionOptions []string    `yaml:"extension-options"`
		Feegrants        interface{} `yaml:"feegrants"`
	} `yaml:"value"`
}

func NewChainConfig() Chain {
	chain := Chain{}
	chain.Type = "cosmos"
	chain.Value.Key = "relayer"
	chain.Value.KeyringBackend = "test"
	chain.Value.GasAdjustment = 1.5
	chain.Value.Timeout = "20s"
	chain.Value.OutputFormat = "json"
	chain.Value.SignMode = "direct"
	chain.Value.CoinType = 118
	chain.Value.BroadcastMode = "batch"
	chain.Value.MinLoopDuration = "0s"
	chain.Value.Feegrants = nil

	return chain
}

type PathItem struct {
	ChainId string `yaml:"chain-id"`
}

type SrcChannelFilter struct {
	Rule        string   `yaml:"rule"`
	ChannelList []string `yaml:"channel-list"`
}

type Path struct {
	Src              PathItem         `yaml:"src"`
	Dst              PathItem         `yaml:"dst"`
	SrcChannelFilter SrcChannelFilter `yaml:"src-channel-filter"`
}
