package meta

type CertConfig struct {
	Alias           string   `yaml:"alias"`
	Auth            string   `yaml:"auth"`
	Email           string   `yaml:"email"`
	Domain          string   `yaml:"domain"`
	DomainCN        []string `yaml:"domain_cn"`
	DomainCheck     string   `yaml:"domain_check"`
	Provider        string   `yaml:"provider"`
	AccessKey       string   `yaml:"ak"`
	AccessKeySecret string   `yaml:"sk"`
	AutoRenew       bool     `yaml:"auto_renew"`
	UploadToken     string   `yaml:"upload_token"`
	DisableCname    *bool    `yaml:"disable_cname"`
}
