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
type CertInfo struct {
	Alias           string     `json:"alias"`
	CN              string     `json:"cn"`
	DNSNames        []string   `json:"dns_names"`
	NotBefore       time.Time  `json:"not_before"`
	NotAfter        time.Time  `json:"not_after"`
	DaysRemaining   int        `json:"days_remaining"`
	ExpectedRenewAt *time.Time `json:"expected_renew_at,omitempty"`
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

	if !CertStorage.CertExists(certAlias) {
		return ctx.JSONResponse(FailedWithS("certificate files not found", 404), 404)
	}

	data, err := CertStorage.ReadFullchain(certAlias)
	if err != nil {
		slog.Error("failed to read fullchain", "alias", certAlias, "error", err)
		return ctx.JSONResponse(Failed("failed to read certificate"))
	}

	x509Cert, err := cert.ParseCertExpiry(data)
	if err != nil {
		slog.Error("failed to parse certificate", "alias", certAlias, "error", err)
		return ctx.JSONResponse(Failed("failed to parse certificate"))
	}

	daysRemaining := int(time.Until(x509Cert.NotAfter).Hours() / 24)
	expectedRenew := GlobalScheduler.ExpectedRenewAt(x509Cert.NotAfter, certConfig.AutoRenew)

	return ctx.JSONResponse(SuccessWithD(CertInfo{
		Alias:           certAlias,
		CN:              x509Cert.Subject.CommonName,
		DNSNames:        x509Cert.DNSNames,
		NotBefore:       x509Cert.NotBefore,
		NotAfter:        x509Cert.NotAfter,
		DaysRemaining:   daysRemaining,
		ExpectedRenewAt: expectedRenew,
	}))
}
