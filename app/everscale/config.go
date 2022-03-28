package everscale

type Config struct {
	Servers   []string `yaml:"servers"`
	Contracts struct {
		DepositContract struct {
			ABI     string `yaml:"abi"`
			Address string `yaml:"address"`
		} `yaml:"depositContract"`
	} `yaml:"contracts"`
}
