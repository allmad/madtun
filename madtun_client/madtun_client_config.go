package madtun_client

type Config struct {
}

func (c *Config) FlaglyDesc() string { return "client mode" }

func (c *Config) FlaglyHandle() error {
	return Run(c)
}
