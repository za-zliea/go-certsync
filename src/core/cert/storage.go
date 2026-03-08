package cert

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type CertStorage struct {
	StorageDir string
}

func NewCertStorage(storageDir string) *CertStorage {
	return &CertStorage{StorageDir: storageDir}
}

func (s *CertStorage) GetCertDir(alias string) string {
	return filepath.Join(s.StorageDir, alias)
}

func (s *CertStorage) GetCertPath(alias string) string {
	return filepath.Join(s.GetCertDir(alias), "cert.pem")
}

func (s *CertStorage) GetFullchainPath(alias string) string {
	return filepath.Join(s.GetCertDir(alias), "fullchain.pem")
}

func (s *CertStorage) GetChainPath(alias string) string {
	return filepath.Join(s.GetCertDir(alias), "chain.pem")
}

func (s *CertStorage) GetPrivkeyPath(alias string) string {
	return filepath.Join(s.GetCertDir(alias), "privkey.pem")
}

func (s *CertStorage) EnsureCertDir(alias string) error {
	certDir := s.GetCertDir(alias)
	return os.MkdirAll(certDir, 0755)
}

func (s *CertStorage) SaveCert(alias string, fullchain, privkey []byte) error {
	if err := s.EnsureCertDir(alias); err != nil {
		return fmt.Errorf("failed to create cert directory: %v", err)
	}

	if err := os.WriteFile(s.GetFullchainPath(alias), fullchain, 0644); err != nil {
		return fmt.Errorf("failed to write fullchain.pem: %v", err)
	}

	if err := os.WriteFile(s.GetPrivkeyPath(alias), privkey, 0600); err != nil {
		return fmt.Errorf("failed to write privkey.pem: %v", err)
	}

	// Extract cert (leaf certificate) - first certificate in chain
	cert := extractCert(fullchain)
	if len(cert) > 0 {
		if err := os.WriteFile(s.GetCertPath(alias), cert, 0644); err != nil {
			return fmt.Errorf("failed to write cert.pem: %v", err)
		}
	}

	// Extract chain (intermediate certificates) - remaining certificates
	chain := extractChain(fullchain)
	if len(chain) > 0 {
		if err := os.WriteFile(s.GetChainPath(alias), chain, 0644); err != nil {
			return fmt.Errorf("failed to write chain.pem: %v", err)
		}
	}

	return nil
}

// extractCert extracts the leaf certificate (first certificate) from fullchain
func extractCert(fullchain []byte) []byte {
	block, _ := pem.Decode(fullchain)
	if block == nil {
		return nil
	}
	return pem.EncodeToMemory(block)
}

// extractChain extracts intermediate certificates from fullchain
// (all certificates except the first one, which is the leaf certificate)
func extractChain(fullchain []byte) []byte {
	var chain []byte
	var blocks [][]byte
	remaining := fullchain

	for {
		block, rest := pem.Decode(remaining)
		if block == nil {
			break
		}
		blocks = append(blocks, pem.EncodeToMemory(block))
		remaining = rest
	}

	// Skip the first block (leaf certificate), keep the rest (intermediates)
	for i := 1; i < len(blocks); i++ {
		chain = append(chain, blocks[i]...)
	}

	return chain
}

func (s *CertStorage) ReadCert(alias string) ([]byte, error) {
	return os.ReadFile(s.GetCertPath(alias))
}

func (s *CertStorage) ReadFullchain(alias string) ([]byte, error) {
	return os.ReadFile(s.GetFullchainPath(alias))
}

func (s *CertStorage) ReadChain(alias string) ([]byte, error) {
	return os.ReadFile(s.GetChainPath(alias))
}

func (s *CertStorage) ReadPrivkey(alias string) ([]byte, error) {
	return os.ReadFile(s.GetPrivkeyPath(alias))
}

func (s *CertStorage) CertExists(alias string) bool {
	_, err := os.Stat(s.GetFullchainPath(alias))
	return err == nil
}

func ParseCertExpiry(data []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}
	return x509.ParseCertificate(block.Bytes)
}
