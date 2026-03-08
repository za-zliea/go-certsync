package meta

type ServerConfig struct {
	Address   string                 `yaml:"address"`
	Port      int                    `yaml:"port"`
	Token     string                 `yaml:"token"`
	Storage   string                 `yaml:"storage"`
	CheckTime string                 `yaml:"cert_check_time"`
	DNS       string                 `yaml:"dns"`
	Certs     []*CertConfig          `yaml:"certs"`
	CertMap   map[string]*CertConfig `yaml:"-"`
}

func (s *ServerConfig) Generate() interface{} {
	s.CertMap = make(map[string]*CertConfig)
	for _, cert := range s.Certs {
		s.CertMap[cert.Alias] = cert
	}
	return s
}

func (s *ServerConfig) Empty(meta interface{}) bool {
	return len(s.Certs) == 0
}
