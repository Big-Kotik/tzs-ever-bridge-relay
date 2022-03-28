package tezos

type Config struct {
	Server struct {
		TezosURL string `yaml:"URL"`
	} `yaml:"server"`
	Wallet struct {
		PrivateKey string `yaml:"privateKey"`
	} `yaml:"wallet"`
	Contracts struct {
		QuorumAddress  string `yaml:"quorumAddress"`
		DepositAddress string `yaml:"depositAddress"`
	} `yaml:"contracts"`
}
