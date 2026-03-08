package cert

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

func GetRemoteCertExpiry(domainCheck string) (time.Time, error) {
	parts := strings.Split(domainCheck, ":")
	host := parts[0]
	port := "443"
	if len(parts) > 1 {
		port = parts[1]
	}

	conn, err := tls.Dial("tcp", host+":"+port, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to connect to %s: %v", domainCheck, err)
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return time.Time{}, errors.New("no certificates found")
	}

	return certs[0].NotAfter, nil
}

func GetLocalCertExpiry(fullchainPath string) (time.Time, error) {
	data, err := os.ReadFile(fullchainPath)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to read certificate file: %v", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return time.Time{}, errors.New("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse certificate: %v", err)
	}

	return cert.NotAfter, nil
}

func NeedsRenewal(expiry time.Time, daysThreshold int) bool {
	return time.Until(expiry) < time.Duration(daysThreshold)*24*time.Hour
}
