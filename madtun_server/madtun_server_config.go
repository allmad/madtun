package madtun_server

type Config struct {
	Meta string
}

func (c *Config) init() {
	if c.Meta == "" {
		c.Meta = ":14234"
	}
}

func (c *Config) FlaglyDesc() string { return "server mode" }

func (c *Config) FlaglyHandle() error {
	c.init()
	return Run(c)
}
