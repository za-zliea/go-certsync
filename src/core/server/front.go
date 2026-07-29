package server

import (
	"certsync/src/core/cert"
	"fmt"
	"log/slog"

	"github.com/savsgio/atreugo/v11"
)

// CertListItem is the public projection of a CertConfig exposed via /front/list.
// Sensitive fields (AccessKey, AccessKeySecret, Auth, UploadToken) are intentionally omitted.
type CertListItem struct {
	Alias        string   `json:"alias"`
	Email        string   `json:"email"`
	Domain       string   `json:"domain"`
	DomainCN     []string `json:"domain_cn"`
	DomainCheck  string   `json:"domain_check"`
	Provider     string   `json:"provider"`
	AutoRenew    bool     `json:"auto_renew"`
	DisableCname *bool    `json:"disable_cname,omitempty"`
}

// FrontAliasListHandler returns the list of configured certificates (without secrets).
func FrontAliasListHandler(ctx *atreugo.RequestCtx) error {
	if _, ok := authJWT(ctx); !ok {
		return ctx.JSONResponse(FailedWithS("auth failed", 401), 401)
	}

	items := make([]CertListItem, 0, len(MetaData.Certs))
	for _, c := range MetaData.Certs {
		items = append(items, CertListItem{
			Alias:        c.Alias,
			Email:        c.Email,
			Domain:       c.Domain,
			DomainCN:     c.DomainCN,
			DomainCheck:  c.DomainCheck,
			Provider:     c.Provider,
			AutoRenew:    c.AutoRenew,
			DisableCname: c.DisableCname,
		})
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
