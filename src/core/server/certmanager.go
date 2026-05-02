package server

import (
	"certsync/src/core/cert"
	"certsync/src/core/meta"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net"
	"os"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/alidns"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/providers/dns/gcloud"
	"github.com/go-acme/lego/v4/providers/dns/godaddy"
	"github.com/go-acme/lego/v4/providers/dns/tencentcloud"
	"github.com/go-acme/lego/v4/registration"
)

type CertManager struct {
	Storage      *cert.CertStorage
	DNS          string
	preCheckOnce sync.Once
}

type loggingProvider struct {
	provider challenge.Provider
}

func (lp *loggingProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	slog.Info("DNS API: Present", "domain", domain, "fqdn", fqdn, "value", value)
	return lp.provider.Present(domain, token, keyAuth)
}

func (lp *loggingProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	slog.Info("DNS API: CleanUp", "domain", domain, "fqdn", fqdn, "value", value)
	return lp.provider.CleanUp(domain, token, keyAuth)
}

type AcmeUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *AcmeUser) GetEmail() string {
	return u.Email
}

func (u *AcmeUser) GetRegistration() *registration.Resource {
	return u.Registration
}

func (u *AcmeUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func NewCertManager(storage *cert.CertStorage, dns string) *CertManager {
	return &CertManager{Storage: storage, DNS: dns}
}

func (cm *CertManager) RenewCert(certConfig *meta.CertConfig) error {
	if len(certConfig.DomainCN) == 0 {
		return errors.New("domain_cn is required")
	}

	if certConfig.Email == "" {
		return errors.New("email is required")
	}

	// Disable CNAME support for DNS-01 challenge
	if certConfig.DisableCname == nil || *certConfig.DisableCname {
		os.Setenv("LEGO_DISABLE_CNAME_SUPPORT", "true")
		defer os.Unsetenv("LEGO_DISABLE_CNAME_SUPPORT")
	}

	// Reset sync.Once for this certificate request
	cm.preCheckOnce = sync.Once{}

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %v", err)
	}

	user := &AcmeUser{
		Email: certConfig.Email,
		key:   privateKey,
	}

	config := lego.NewConfig(user)
	config.CADirURL = lego.LEDirectoryProduction

	client, err := lego.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to create lego client: %v", err)
	}

	provider, err := cm.newChallengeProvider(certConfig)
	if err != nil {
		return fmt.Errorf("failed to create DNS provider: %v", err)
	}

	loggingWrapper := &loggingProvider{provider: provider}

	var dnsOpts []dns01.ChallengeOption
	if cm.DNS != "" {
		dnsOpts = append(dnsOpts, dns01.AddRecursiveNameservers([]string{cm.DNS + ":53"}))
	}
	dnsOpts = append(dnsOpts, dns01.WrapPreCheck(cm.dnsPreCheck))

	err = client.Challenge.SetDNS01Provider(loggingWrapper, dnsOpts...)
	if err != nil {
		return fmt.Errorf("failed to set DNS challenge provider: %v", err)
	}

	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return fmt.Errorf("failed to register with ACME server: %v", err)
	}
	user.Registration = reg

	certRequest := certificate.ObtainRequest{
		Domains: certConfig.DomainCN,
		Bundle:  true,
	}

	certificates, err := client.Certificate.Obtain(certRequest)
	if err != nil {
		return fmt.Errorf("failed to obtain certificate: %v", err)
	}

	if err := cm.Storage.SaveCert(certConfig.Alias, certificates.Certificate, certificates.PrivateKey); err != nil {
		return fmt.Errorf("failed to save certificate: %v", err)
	}

	return nil
}

func (cm *CertManager) newChallengeProvider(certConfig *meta.CertConfig) (challenge.Provider, error) {
	switch certConfig.Provider {
	case "TENCENT":
		config := tencentcloud.NewDefaultConfig()
		config.SecretID = certConfig.AccessKey
		config.SecretKey = certConfig.AccessKeySecret
		return tencentcloud.NewDNSProviderConfig(config)
	case "ALIYUN":
		config := alidns.NewDefaultConfig()
		config.APIKey = certConfig.AccessKey
		config.SecretKey = certConfig.AccessKeySecret
		return alidns.NewDNSProviderConfig(config)
	case "GODADDY":
		config := godaddy.NewDefaultConfig()
		config.APIKey = certConfig.AccessKey
		config.APISecret = certConfig.AccessKeySecret
		return godaddy.NewDNSProviderConfig(config)
	case "GOOGLE":
		os.Setenv("GCE_PROJECT", certConfig.AccessKey)
		return gcloud.NewDNSProvider()
	case "CLOUDFLARE":
		config := cloudflare.NewDefaultConfig()
		config.AuthToken = certConfig.AccessKey
		return cloudflare.NewDNSProviderConfig(config)
	default:
		return nil, errors.New("unsupported DNS provider: " + certConfig.Provider)
	}
}

func (cm *CertManager) dnsPreCheck(domain, fqdn, value string, check dns01.PreCheckFunc) (bool, error) {
	cm.preCheckOnce.Do(func() {
		slog.Info("DNS PreCheck: waiting 30s for DNS propagation after API call")
		time.Sleep(30 * time.Second)
	})

	slog.Info("DNS PreCheck: starting poll", "domain", domain, "fqdn", fqdn, "expected_value", value)
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var resolver *net.Resolver
	if cm.DNS != "" {
		resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: 5 * time.Second}
				return d.DialContext(ctx, network, cm.DNS+":53")
			},
		}
	}

	for {
		select {
		case <-ticker.C:
			var txtRecords []string
			var err error
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			if resolver != nil {
				txtRecords, err = resolver.LookupTXT(ctx, fqdn)
			} else {
				txtRecords, err = net.DefaultResolver.LookupTXT(ctx, fqdn)
			}
			cancel()

			if err != nil {
				slog.Warn("DNS PreCheck: lookup failed", "domain", domain, "fqdn", fqdn, "error", err)
				continue
			}

			for _, txt := range txtRecords {
				if txt == value {
					slog.Info("DNS PreCheck: passed", "domain", domain, "fqdn", fqdn, "value", value)
					return true, nil
				}
			}
			slog.Info("DNS PreCheck: not ready, retrying in 5s", "domain", domain, "fqdn", fqdn, "records", txtRecords)

		case <-timeout:
			slog.Warn("DNS PreCheck: timeout, proceeding anyway", "domain", domain, "fqdn", fqdn)
			return true, nil
		}
	}
}

func (cm *CertManager) GenerateSelfSignedCert(certConfig *meta.CertConfig) error {
	if len(certConfig.DomainCN) == 0 {
		return errors.New("domain_cn is required")
	}

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %v", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"CertSync"},
			CommonName:   certConfig.Domain,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              certConfig.DomainCN,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %v", err)
	}

	privBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %v", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})

	if err := cm.Storage.SaveCert(certConfig.Alias, certPEM, keyPEM); err != nil {
		return fmt.Errorf("failed to save certificate: %v", err)
	}

	return nil
}
