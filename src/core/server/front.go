package server

import (
	"certsync/src/core/cert"
	"fmt"
	"log/slog"
	"time"

	"github.com/savsgio/atreugo/v11"
)

// CertListItem is the public projection of a CertConfig exposed via /front/list.
// Sensitive fields (AccessKey, AccessKeySecret, Auth, UploadToken) are intentionally omitted.
//
// NotBefore / NotAfter / DaysRemaining / ExpectedRenewAt are populated only
// when a local cert exists for the alias; otherwise they stay nil and, due
// to `omitempty`, are omitted from the JSON — the frontend renders them as
// `-`. Keep them as pointers (not zero values) so "missing" is distinct
// from "expired at epoch".
type CertListItem struct {
	Alias           string     `json:"alias"`
	Email           string     `json:"email"`
	Domain          string     `json:"domain"`
	DomainCN        []string   `json:"domain_cn"`
	DomainCheck     string     `json:"domain_check"`
	Provider        string     `json:"provider"`
	AutoRenew       bool       `json:"auto_renew"`
	DisableCname    *bool      `json:"disable_cname,omitempty"`
	NotBefore       *time.Time `json:"not_before,omitempty"`
	NotAfter        *time.Time `json:"not_after,omitempty"`
	DaysRemaining   *int       `json:"days_remaining,omitempty"`
	ExpectedRenewAt *time.Time `json:"expected_renew_at,omitempty"`
}

// CertInfo is the parsed-certificate metadata returned by /front/api/{alias}/info.
// Remote* 字段在 CertConfig.DomainCheck 配置且远端可达时填充；本地缺失时由远端兜底。
// Remote* 使用指针 + omitempty，以区分 "远端不可达/未配置" 与 "epoch 0 值"。
type CertInfo struct {
	Alias               string     `json:"alias"`
	CN                  string     `json:"cn"`
	DNSNames            []string   `json:"dns_names"`
	NotBefore           time.Time  `json:"not_before"`
	NotAfter            time.Time  `json:"not_after"`
	DaysRemaining       int        `json:"days_remaining"`
	ExpectedRenewAt     *time.Time `json:"expected_renew_at,omitempty"`
	RemoteNotBefore     *time.Time `json:"remote_not_before,omitempty"`
	RemoteNotAfter      *time.Time `json:"remote_not_after,omitempty"`
	RemoteDaysRemaining *int       `json:"remote_days_remaining,omitempty"`
}

// FrontAliasListHandler returns the list of configured certificates (without secrets).
func FrontAliasListHandler(ctx *atreugo.RequestCtx) error {
	if _, ok := authJWT(ctx); !ok {
		return ctx.JSONResponse(FailedWithS("auth failed", 401), 401)
	}

	items := make([]CertListItem, 0, len(MetaData.Certs))
	for _, c := range MetaData.Certs {
		item := CertListItem{
			Alias:        c.Alias,
			Email:        c.Email,
			Domain:       c.Domain,
			DomainCN:     c.DomainCN,
			DomainCheck:  c.DomainCheck,
			Provider:     c.Provider,
			AutoRenew:    c.AutoRenew,
			DisableCname: c.DisableCname,
		}

		// When a local cert exists, enrich the row with parsed validity
		// metadata. Errors here (e.g. corrupted file) are logged but do
		// not fail the whole list — the row simply renders with `-` for
		// the dynamic columns.
		if CertStorage.CertExists(c.Alias) {
			if data, err := CertStorage.ReadFullchain(c.Alias); err == nil {
				if x509Cert, err := cert.ParseCertExpiry(data); err == nil {
					notBefore := x509Cert.NotBefore
					notAfter := x509Cert.NotAfter
					days := int(time.Until(notAfter).Hours() / 24)
					item.NotBefore = &notBefore
					item.NotAfter = &notAfter
					item.DaysRemaining = &days
					item.ExpectedRenewAt = GlobalScheduler.ExpectedRenewAt(notAfter, c.AutoRenew)
				} else {
					slog.Warn("failed to parse certificate for list", "alias", c.Alias, "error", err)
				}
			} else {
				slog.Warn("failed to read fullchain for list", "alias", c.Alias, "error", err)
			}
		}

		items = append(items, item)
	}

	return ctx.JSONResponse(SuccessWithD(items))
}

// FrontDownloadHandler streams a ZIP archive containing all certificate files for the given alias.
func FrontDownloadHandler(ctx *atreugo.RequestCtx) error {
	if _, ok := authJWT(ctx); !ok {
		return ctx.JSONResponse(FailedWithS("auth failed", 401), 401)
	}

	certAlias, ok := ctx.UserValue("alias").(string)
	if !ok || certAlias == "" {
		return ctx.JSONResponse(FailedWithS("cert alias is required", 400), 400)
	}

	if _, ok := MetaData.CertMap[certAlias]; !ok {
		return ctx.JSONResponse(FailedWithS("cert not found", 404), 404)
	}

	if !CertStorage.CertExists(certAlias) {
		return ctx.JSONResponse(FailedWithS("certificate files not found", 404), 404)
	}

	var files cert.CertFiles
	var err error

	if files.Fullchain, err = CertStorage.ReadFullchain(certAlias); err != nil {
		slog.Error("failed to read fullchain", "alias", certAlias, "error", err)
		return ctx.JSONResponse(Failed("failed to read certificate"))
	}

	if files.Privkey, err = CertStorage.ReadPrivkey(certAlias); err != nil {
		slog.Error("failed to read privkey", "alias", certAlias, "error", err)
		return ctx.JSONResponse(Failed("failed to read certificate"))
	}

	if files.Cert, err = CertStorage.ReadCert(certAlias); err != nil {
		slog.Error("failed to read cert", "alias", certAlias, "error", err)
		return ctx.JSONResponse(Failed("failed to read certificate"))
	}

	if files.Chain, err = CertStorage.ReadChain(certAlias); err != nil {
		slog.Error("failed to read chain", "alias", certAlias, "error", err)
		return ctx.JSONResponse(Failed("failed to read certificate"))
	}

	zipData, err := cert.CreateCertZip(files)
	if err != nil {
		slog.Error("failed to create zip", "alias", certAlias, "error", err)
		return ctx.JSONResponse(Failed("failed to create certificate package"))
	}

	ctx.Response.Header.Set("Content-Type", "application/zip")
	ctx.Response.Header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", certAlias))
	ctx.Response.AppendBody(zipData)
	return nil
}

// FrontUploadHandler accepts a manually uploaded certificate for a non-auto-renew alias.
// Authentication is performed via JWT (replacing the legacy upload_token flow).
func FrontUploadHandler(ctx *atreugo.RequestCtx) error {
	if _, ok := authJWT(ctx); !ok {
		return ctx.JSONResponse(FailedWithS("auth failed", 401), 401)
	}

	certAlias, ok := ctx.UserValue("alias").(string)
	if !ok || certAlias == "" {
		return ctx.JSONResponse(FailedWithS("cert alias is required", 400), 400)
	}

	certConfig, ok := MetaData.CertMap[certAlias]
	if !ok {
		return ctx.JSONResponse(FailedWithS("cert not found", 404), 404)
	}

	if certConfig.AutoRenew {
		return ctx.JSONResponse(FailedWithS("upload is not allowed for auto-renewal certificate", 403), 403)
	}

	form, err := ctx.MultipartForm()
	if err != nil {
		return ctx.JSONResponse(FailedWithS("invalid multipart form: "+err.Error(), 400), 400)
	}

	fullchainFiles, ok := form.File["fullchain"]
	if !ok || len(fullchainFiles) == 0 {
		return ctx.JSONResponse(FailedWithS("fullchain file is required", 400), 400)
	}
	fullchainData, err := readFileFromMultipartFile(fullchainFiles[0])
	if err != nil {
		return ctx.JSONResponse(Failed("failed to read fullchain file"))
	}

	privkeyFiles, ok := form.File["privkey"]
	if !ok || len(privkeyFiles) == 0 {
		return ctx.JSONResponse(FailedWithS("privkey file is required", 400), 400)
	}
	privkeyData, err := readFileFromMultipartFile(privkeyFiles[0])
	if err != nil {
		return ctx.JSONResponse(Failed("failed to read privkey file"))
	}

	if err := CertStorage.SaveCert(certAlias, fullchainData, privkeyData); err != nil {
		slog.Error("failed to save certificate", "alias", certAlias, "error", err)
		return ctx.JSONResponse(Failed("failed to save certificate: " + err.Error()))
	}

	slog.Info("certificate uploaded via front endpoint", "alias", certAlias)
	return ctx.JSONResponse(SuccessWithM("certificate uploaded successfully"))
}

// FrontCertInfoHandler returns parsed metadata (CN, DNS names, validity window) for a certificate.
//
// 本地与远端证书信息相互独立：本地存在则解析填充 not_*/days_remaining；
// CertConfig.DomainCheck 配置时拨号填充 remote_*。本地缺失时 CN/DNSNames 由远端兜底，
// 仅当两者都拿不到才返回 404。远端不可达记 warn 后跳过，不影响响应。
func FrontCertInfoHandler(ctx *atreugo.RequestCtx) error {
	if _, ok := authJWT(ctx); !ok {
		return ctx.JSONResponse(FailedWithS("auth failed", 401), 401)
	}

	certAlias, ok := ctx.UserValue("alias").(string)
	if !ok || certAlias == "" {
		return ctx.JSONResponse(FailedWithS("cert alias is required", 400), 400)
	}

	certConfig, ok := MetaData.CertMap[certAlias]
	if !ok {
		return ctx.JSONResponse(FailedWithS("cert not found", 404), 404)
	}

	info := CertInfo{Alias: certAlias, DNSNames: []string{}}

	// 本地证书
	var localNotAfter time.Time
	var hasLocal bool
	if CertStorage.CertExists(certAlias) {
		data, err := CertStorage.ReadFullchain(certAlias)
		if err == nil {
			if x, err := cert.ParseCertExpiry(data); err == nil {
				info.CN = x.Subject.CommonName
				info.DNSNames = x.DNSNames
				info.NotBefore = x.NotBefore
				info.NotAfter = x.NotAfter
				info.DaysRemaining = int(time.Until(x.NotAfter).Hours() / 24)
				localNotAfter = x.NotAfter
				hasLocal = true
			} else {
				slog.Warn("failed to parse local certificate", "alias", certAlias, "error", err)
			}
		} else {
			slog.Warn("failed to read fullchain", "alias", certAlias, "error", err)
		}
	}

	// 远端证书
	var remoteNotAfter time.Time
	var hasRemote bool
	if certConfig.DomainCheck != "" {
		if x, err := cert.GetRemoteCert(certConfig.DomainCheck); err == nil {
			rBefore, rAfter := x.NotBefore, x.NotAfter
			rDays := int(time.Until(rAfter).Hours() / 24)
			info.RemoteNotBefore = &rBefore
			info.RemoteNotAfter = &rAfter
			info.RemoteDaysRemaining = &rDays
			remoteNotAfter = rAfter
			hasRemote = true
			if !hasLocal {
				info.CN = x.Subject.CommonName
				info.DNSNames = x.DNSNames
			}
		} else {
			slog.Warn("failed to fetch remote certificate", "alias", certAlias,
				"domain_check", certConfig.DomainCheck, "error", err)
		}
	}

	if !hasLocal && !hasRemote {
		return ctx.JSONResponse(FailedWithS("certificate files not found", 404), 404)
	}

	renewAfter := localNotAfter
	if !hasLocal {
		renewAfter = remoteNotAfter
	}
	info.ExpectedRenewAt = GlobalScheduler.ExpectedRenewAt(renewAfter, certConfig.AutoRenew)

	return ctx.JSONResponse(SuccessWithD(info))
}
