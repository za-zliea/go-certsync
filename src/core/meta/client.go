package meta

type ClientConfig struct {
	Server        string `yaml:"server"`
	Token         string `yaml:"token"`
	CertAlias     string `yaml:"cert_alias"`
	CertAuth      string `yaml:"cert_auth"`
	DomainCheck   string `yaml:"domain_check"`
	CertUpdateDir string `yaml:"cert_update_dir"`
	CertUpdateCmd string `yaml:"cert_update_cmd"`
	Interval      int    `yaml:"interval"`
}

func (c *ClientConfig) Generate() interface{} {
	if c.Interval == 0 {
		c.Interval = 86400
	}
	return c
}

func (c *ClientConfig) Empty(meta interface{}) bool {
	return c.Server == ""
}
