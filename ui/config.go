package ui

type Config struct {
	RpcUrl       string `yaml:"rpc_url,omitempty"`
	EtherscanKey string `yaml:"etherscan_key,omitempty"`
	DisableENS   bool   `yaml:"disable_ens,omitempty"`
}
