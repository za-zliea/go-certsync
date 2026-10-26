package cert

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const dialTimeout = 10 * time.Second

// GetRemoteCert 通过 TLS 拨号到 domain_check（host[:port]，默认 443）并返回对端证书。
func GetRemoteCert(domainCheck string) (*x509.Certificate, error) {
	return getRemoteCert(domainCheck, nil)
}

// GetRemoteCertWithProxy 同 GetRemoteCert，支持经 HTTP 代理（CONNECT 隧道）拨号。
func GetRemoteCertWithProxy(domainCheck string, proxy *url.URL) (*x509.Certificate, error) {
	return getRemoteCert(domainCheck, proxy)
}

func getRemoteCert(domainCheck string, proxy *url.URL) (*x509.Certificate, error) {
	parts := strings.Split(domainCheck, ":")
	host := parts[0]
	port := "443"
	if len(parts) > 1 {
		port = parts[1]
	}

	var conn net.Conn
	var err error
	if proxy != nil {
		conn, err = dialViaProxy(proxy, host+":"+port)
	} else {
		conn, err = net.DialTimeout("tcp", host+":"+port, dialTimeout)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %v", domainCheck, err)
	}
	defer conn.Close()

	tlsConn := tls.Client(conn, &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	})
	if err := tlsConn.Handshake(); err != nil {
		return nil, fmt.Errorf("failed to TLS handshake with %s: %v", domainCheck, err)
	}

	certs := tlsConn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return nil, errors.New("no certificates found")
	}
	return certs[0], nil
}

// dialViaProxy 建立到 HTTP 代理的连接并发送 CONNECT 建立隧道，返回隧道裸连接（未做 TLS）。
func dialViaProxy(proxy *url.URL, target string) (net.Conn, error) {
	var conn net.Conn
	var err error
	if proxy.Scheme == "https" {
		conn, err = tls.DialWithDialer(&net.Dialer{Timeout: dialTimeout}, "tcp", proxy.Host, &tls.Config{})
	} else {
		conn, err = net.DialTimeout("tcp", proxy.Host, dialTimeout)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to proxy %s: %v", proxy.Host, err)
	}

	req := &http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Opaque: target},
		Host:   target,
		Header: make(http.Header),
	}
	if u := proxy.User; u != nil {
		if password, ok := u.Password(); ok {
			req.Header.Set("Proxy-Authorization", "Basic "+
				base64.StdEncoding.EncodeToString([]byte(u.Username()+":"+password)))
		}
	}
	if err := req.Write(conn); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to send CONNECT to proxy: %v", err)
	}

	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to read proxy response: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		conn.Close()
		return nil, fmt.Errorf("proxy CONNECT failed: %s", resp.Status)
	}

	return conn, nil
}

// GetRemoteCertExpiry 保留原签名以兼容现有调用方（scheduler / CheckHandler）。
func GetRemoteCertExpiry(domainCheck string) (time.Time, error) {
	c, err := GetRemoteCert(domainCheck)
	if err != nil {
		return time.Time{}, err
	}
	return c.NotAfter, nil
}

// GetRemoteCertExpiryWithProxy 同 GetRemoteCertExpiry，支持经 HTTP 代理（CONNECT 隧道）拨号。
func GetRemoteCertExpiryWithProxy(domainCheck string, proxy *url.URL) (time.Time, error) {
	c, err := GetRemoteCertWithProxy(domainCheck, proxy)
	if err != nil {
		return time.Time{}, err
	}
	return c.NotAfter, nil
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
