package server

import (
	"certsync/src/core/cert"
	"certsync/src/core/meta"
	"errors"
	"fmt"
	"log/slog"
	"mime/multipart"
	"time"

	"github.com/savsgio/atreugo/v11"
)

var MetaData *meta.ServerConfig
var CertStorage *cert.CertStorage
var GlobalScheduler *Scheduler

type CheckResponse struct {
	NeedUpdate   bool   `json:"need_update"`
	LocalExpiry  string `json:"local_expiry"`
	RemoteExpiry string `json:"remote_expiry"`
	Reason       string `json:"reason"`
}

type ExpireResponse struct {
	Expiry string `json:"expiry"`
}

func IndexHandler(ctx *atreugo.RequestCtx) error {
	return ctx.JSONResponse(Success())
}

func CheckHandler(ctx *atreugo.RequestCtx) error {
	if !authGlobal(ctx) {
		return ctx.JSONResponse(FailedWithS("auth failed", 401), 401)
	}

	certAlias, certAuth, domainCheck, err := getCertAuthAndDomainCheck(ctx)
	if err != nil {
		return ctx.JSONResponse(FailedWithS(err.Error(), 401), 401)
	}

	certConfig, ok := MetaData.CertMap[certAlias]
	if !ok {
		return ctx.JSONResponse(FailedWithS("cert not found", 404), 404)
	}

	if certConfig.Auth != certAuth {
		return ctx.JSONResponse(FailedWithS("cert auth failed", 401), 401)
	}

	response := CheckResponse{}

	// 第1步: 检查远端domain_check证书，如果无法访问直接返回报错
	var remoteExpiry time.Time
	remoteExpiry, remoteErr := cert.GetRemoteCertExpiry(domainCheck)
	if remoteErr != nil {
		slog.Error("failed to check remote certificate", "alias", certAlias, "domain_check", domainCheck, "error", remoteErr)
		return ctx.JSONResponse(FailedWithS(fmt.Sprintf("failed to check remote certificate: %v", remoteErr), 500), 500)
	}
	response.RemoteExpiry = remoteExpiry.Format(time.RFC3339)

	// 第2步: 检查本地是否有证书
	localExpiry, localErr := cert.GetLocalCertExpiry(CertStorage.GetFullchainPath(certAlias))

	if localErr == nil {
		response.LocalExpiry = localExpiry.Format(time.RFC3339)
	}

	// 第3步: 如果还是没有证书，则返回need_update=false
	if localErr != nil {
		response.NeedUpdate = false
		response.Reason = "no local certificate available"
		return ctx.JSONResponse(SuccessWithD(response))
	}

	// 第4步: 如果有证书，判断远端是否早于本地证书
	if !remoteExpiry.IsZero() && remoteExpiry.Before(localExpiry) {
		response.NeedUpdate = true
		response.Reason = "remote certificate expires before local"
	} else {
		response.NeedUpdate = false
		response.Reason = "local certificate is up to date"
	}

	return ctx.JSONResponse(SuccessWithD(response))
}

func DownloadHandler(ctx *atreugo.RequestCtx) error {
	if !authGlobal(ctx) {
		return ctx.JSONResponse(FailedWithS("auth failed", 401), 401)
	}

	certAlias, certAuth, err := getCertAuth(ctx)
	if err != nil {
		return ctx.JSONResponse(FailedWithS(err.Error(), 401), 401)
	}

	certConfig, ok := MetaData.CertMap[certAlias]
	if !ok {
		return ctx.JSONResponse(FailedWithS("cert not found", 404), 404)
	}

	if certConfig.Auth != certAuth {
		return ctx.JSONResponse(FailedWithS("cert auth failed", 401), 401)
	}

	if !CertStorage.CertExists(certAlias) {
		return ctx.JSONResponse(FailedWithS("certificate files not found", 404), 404)
	}

	// Read all certificate files
	var files cert.CertFiles

	files.Fullchain, err = CertStorage.ReadFullchain(certAlias)
	if err != nil {
		slog.Error("failed to read fullchain", "alias", certAlias, "error", err)
		return ctx.JSONResponse(Failed("failed to read certificate"))
	}

	files.Privkey, err = CertStorage.ReadPrivkey(certAlias)
	if err != nil {
		slog.Error("failed to read privkey", "alias", certAlias, "error", err)
		return ctx.JSONResponse(Failed("failed to read certificate"))
	}

	files.Cert, err = CertStorage.ReadCert(certAlias)
	if err != nil {
		slog.Error("failed to read cert", "alias", certAlias, "error", err)
		return ctx.JSONResponse(Failed("failed to read certificate"))
	}

	files.Chain, err = CertStorage.ReadChain(certAlias)
	if err != nil {
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

func ExpireHandler(ctx *atreugo.RequestCtx) error {
	if !authGlobal(ctx) {
		return ctx.JSONResponse(FailedWithS("auth failed", 401), 401)
	}

	certAlias, certAuth, err := getCertAuth(ctx)
	if err != nil {
		return ctx.JSONResponse(FailedWithS(err.Error(), 401), 401)
	}

	certConfig, ok := MetaData.CertMap[certAlias]
	if !ok {
		return ctx.JSONResponse(FailedWithS("cert not found", 404), 404)
	}

	if certConfig.Auth != certAuth {
		return ctx.JSONResponse(FailedWithS("cert auth failed", 401), 401)
	}

	localExpiry, err := cert.GetLocalCertExpiry(CertStorage.GetFullchainPath(certAlias))
	if err != nil {
		var remoteExpiry time.Time
		_remoteExpire := string(ctx.QueryArgs().Peek("remoteExpire"))
		if _remoteExpire != "" {
			remoteExpiry, err = time.Parse(time.RFC3339, _remoteExpire)

			if err != nil {
				slog.Error("failed to parse remote expiry", "error", err)
				return ctx.JSONResponse(Failed("failed to parse remote expiry"))
			}

		}

		if cert.NeedsRenewal(remoteExpiry, 10) {
			go func(cfg *meta.CertConfig) {
				slog.Info("async renewal triggered by expire check", "alias", cfg.Alias)
				status := GlobalScheduler.CheckAndRenewCertStatus(cfg)
				slog.Info("async renewal completed", "alias", cfg.Alias, "status", status)
			}(certConfig)
		}

		slog.Error("failed to get local certificate expiry", "alias", certAlias, "error", err)
		return ctx.JSONResponse(FailedWithS("failed to get certificate expiry: "+err.Error(), 500), 500)
	}

	response := ExpireResponse{
		Expiry: localExpiry.Format(time.RFC3339),
	}

	return ctx.JSONResponse(SuccessWithD(response))
}

func authGlobal(ctx *atreugo.RequestCtx) bool {
	token := string(ctx.Request.Header.Peek("Authorization"))
	return token != "" && token == MetaData.Token
}

func UploadHandler(ctx *atreugo.RequestCtx) error {
	certAlias := ctx.UserValue("alias").(string)
	if certAlias == "" {
		return ctx.JSONResponse(FailedWithS("cert alias is required", 400), 400)
	}

	certConfig, ok := MetaData.CertMap[certAlias]
	if !ok {
		return ctx.JSONResponse(FailedWithS("cert not found", 404), 404)
	}

	// Check if manual upload is allowed (auto_renew must be false)
	if certConfig.AutoRenew {
		return ctx.JSONResponse(FailedWithS("upload is not allowed for auto-renewal certificate", 403), 403)
	}

	// Validate upload token
	uploadToken := string(ctx.QueryArgs().Peek("upload_token"))
	if uploadToken == "" {
		return ctx.JSONResponse(FailedWithS("upload_token is required", 401), 401)
	}

	if certConfig.UploadToken != uploadToken {
		return ctx.JSONResponse(FailedWithS("invalid upload token", 401), 401)
	}

	// Parse multipart form
	form, err := ctx.MultipartForm()
	if err != nil {
		return ctx.JSONResponse(FailedWithS("invalid multipart form: "+err.Error(), 400), 400)
	}

	// Get fullchain file
	fullchainFiles, ok := form.File["fullchain"]
	if !ok || len(fullchainFiles) == 0 {
		return ctx.JSONResponse(FailedWithS("fullchain file is required", 400), 400)
	}
	fullchainFile := fullchainFiles[0]
	fullchainData, err := readFileFromMultipartFile(fullchainFile)
	if err != nil {
		return ctx.JSONResponse(Failed("failed to read fullchain file"))
	}

	// Get privkey file
	privkeyFiles, ok := form.File["privkey"]
	if !ok || len(privkeyFiles) == 0 {
		return ctx.JSONResponse(FailedWithS("privkey file is required", 400), 400)
	}
	privkeyFile := privkeyFiles[0]
	privkeyData, err := readFileFromMultipartFile(privkeyFile)
	if err != nil {
		return ctx.JSONResponse(Failed("failed to read privkey file"))
	}

	// Save certificate
	if err := CertStorage.SaveCert(certAlias, fullchainData, privkeyData); err != nil {
		slog.Error("failed to save certificate", "alias", certAlias, "error", err)
		return ctx.JSONResponse(Failed("failed to save certificate: " + err.Error()))
	}

	slog.Info("certificate uploaded", "alias", certAlias)
	return ctx.JSONResponse(SuccessWithM("certificate uploaded successfully"))
}

func readFileFromMultipartFile(fh *multipart.FileHeader) ([]byte, error) {
	file, err := fh.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data := make([]byte, fh.Size)
	_, err = file.Read(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func getCertAuth(ctx *atreugo.RequestCtx) (string, string, error) {
	certAlias := ctx.UserValue("alias").(string)
	if certAlias == "" {
		return "", "", errors.New("cert alias is required")
	}

	certAuth := string(ctx.QueryArgs().Peek("auth"))
	if certAuth == "" {
		return "", "", errors.New("cert auth is required")
	}

	return certAlias, certAuth, nil
}

func getCertAuthAndDomainCheck(ctx *atreugo.RequestCtx) (string, string, string, error) {
	certAlias := ctx.UserValue("alias").(string)
	if certAlias == "" {
		return "", "", "", errors.New("cert alias is required")
	}

	certAuth := string(ctx.QueryArgs().Peek("auth"))
	if certAuth == "" {
		return "", "", "", errors.New("cert auth is required")
	}

	domainCheck := string(ctx.QueryArgs().Peek("domain_check"))
	if domainCheck == "" {
		return "", "", "", errors.New("domain_check is required")
	}

	return certAlias, certAuth, domainCheck, nil
}
