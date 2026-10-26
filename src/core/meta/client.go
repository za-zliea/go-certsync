package meta

import (
	"errors"
	"fmt"
	"net/url"
)

type ClientConfig struct {
	Server        string      `yaml:"server"`
	Token         string      `yaml:"token"`
	CertAlias     string      `yaml:"cert_alias"`
	CertAuth      string      `yaml:"cert_auth"`
	DomainCheck   string      `yaml:"domain_check"`
	LocalCheck    bool        `yaml:"local_check,omitempty"`
	CertUpdateDir string      `yaml:"cert_update_dir"`
	CertUpdateCmd string      `yaml:"cert_update_cmd"`
	Interval      int         `yaml:"interval"`
	Proxy         ProxyConfig `yaml:"proxy"`
}

type ProxyConfig struct {
	Enable     bool   `yaml:"enable"`
	URL        string `yaml:"url"`
	Username   string `yaml:"username,omitempty"`
	Password   string `yaml:"password,omitempty"`
	ServerAPI  *bool  `yaml:"server_api,omitempty"`
	LocalCheck *bool  `yaml:"local_check,omitempty"`
}

// UseServerAPI 判断访问 server 的 API 请求是否走代理（未显式配置时默认走代理）。
func (p *ProxyConfig) UseServerAPI() bool {
	return p.Enable && (p.ServerAPI == nil || *p.ServerAPI)
}

// UseLocalCheck 判断 local_check 探测是否走代理（未显式配置时默认走代理）。
func (p *ProxyConfig) UseLocalCheck() bool {
	return p.Enable && (p.LocalCheck == nil || *p.LocalCheck)
}

// Parse 解析代理配置：未启用时返回 (nil, nil)；启用时校验并返回内嵌认证信息的代理 URL。
func (p *ProxyConfig) Parse() (*url.URL, error) {
	if !p.Enable {
		return nil, nil
	}
	if p.URL == "" {
		return nil, errors.New("proxy is enabled but url is empty")
	}

	u, err := url.Parse(p.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy url: %v", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported proxy scheme: %s", u.Scheme)
	}
	if p.Username != "" {
		u.User = url.UserPassword(p.Username, p.Password)
	}
	return u, nil
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
