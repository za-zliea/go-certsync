package server

import (
	"certsync/src/core/cert"
	"certsync/src/core/meta"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/savsgio/atreugo/v11"
)

var MetaData *meta.ServerConfig
var CertStorage *cert.CertStorage

type CheckResponse struct {
	NeedUpdate   bool   `json:"need_update"`
	LocalExpiry  string `json:"local_expiry"`
	RemoteExpiry string `json:"remote_expiry"`
	Reason       string `json:"reason"`
}

func IndexHandler(ctx *atreugo.RequestCtx) error {
	return ctx.JSONResponse(Success())
}

func CheckHandler(ctx *atreugo.RequestCtx) error {
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

	response := CheckResponse{}

	localExpiry, localErr := cert.GetLocalCertExpiry(CertStorage.GetFullchainPath(certAlias))
	if localErr == nil {
		response.LocalExpiry = localExpiry.Format(time.RFC3339)
	}

	remoteExpiry, remoteErr := cert.GetRemoteCertExpiry(certConfig.DomainCheck)
	if remoteErr == nil {
		response.RemoteExpiry = remoteExpiry.Format(time.RFC3339)
	}

	if localErr != nil {
		response.NeedUpdate = true
		response.Reason = "local certificate not found"
		return ctx.JSONResponse(SuccessWithD(response))
	}

	if remoteErr != nil {
		response.NeedUpdate = true
		response.Reason = fmt.Sprintf("failed to check remote certificate: %v", remoteErr)
		return ctx.JSONResponse(SuccessWithD(response))
	}

	if remoteExpiry.Before(localExpiry) {
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

func authGlobal(ctx *atreugo.RequestCtx) bool {
	token := string(ctx.Request.Header.Peek("Authorization"))
	return token != "" && token == MetaData.Token
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
