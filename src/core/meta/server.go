package meta

import (
	"crypto/rand"

	"golang.org/x/crypto/bcrypt"
)

const (
	adminDefaultUserName  = "admin"
	adminPasswordLen      = 16
	adminPasswordAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	adminFallbackPassword = "CHANGE_ME_IMMEDIATELY"
)

type ServerConfig struct {
	Address   string                 `yaml:"address"`
	Port      int                    `yaml:"port"`
	Token     string                 `yaml:"token"`
	Storage   string                 `yaml:"storage"`
	CheckTime string                 `yaml:"cert_check_time"`
	DNS       string                 `yaml:"dns"`
	Certs     []*CertConfig          `yaml:"certs"`
	Admin     *Admin                 `yaml:"admin"`
	CertMap   map[string]*CertConfig `yaml:"-"`
}

type Admin struct {
	UserName string `yaml:"username"`
	Password string `yaml:"password"`
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

// generateRandomPassword generates a cryptographically secure alphanumeric password.
func generateRandomPassword(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	out := make([]byte, n)
	for i, b := range buf {
		out[i] = adminPasswordAlphabet[int(b)%len(adminPasswordAlphabet)]
	}
	return string(out), nil
}

// EnsureAdmin populates s.Admin with a freshly generated bcrypt-hashed credential if no
// usable admin account is present. It returns the generated plaintext password and true
// when a new account was created, or "" and false when the existing admin is retained.
func (s *ServerConfig) EnsureAdmin() (string, bool) {
	if s.Admin != nil && s.Admin.UserName != "" && s.Admin.Password != "" {
		return "", false
	}

	plaintext, err := generateRandomPassword(adminPasswordLen)
	if err != nil {
		// Extremely unlikely; surface a static placeholder so operators notice and replace it.
		plaintext = adminFallbackPassword
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		// Should never happen with bcrypt; fall back to storing the plaintext verbatim.
		hashed = []byte(plaintext)
	}

	s.Admin = &Admin{
		UserName: adminDefaultUserName,
		Password: string(hashed),
	}
	return plaintext, true
}
