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
	remoteExpiry, remoteErr := cert.GetRemoteCertExpiry(domainCheck)
	if remoteErr != nil {
		slog.Error("failed to check remote certificate", "alias", certAlias, "domain_check", domainCheck, "error", remoteErr)
		return ctx.JSONResponse(FailedWithS(fmt.Sprintf("failed to check remote certificate: %v", remoteErr), 500), 500)
	}
	response.RemoteExpiry = remoteExpiry.Format(time.RFC3339)

	// 第2步: 检查本地是否有证书
	localExpiry, localErr := cert.GetLocalCertExpiry(CertStorage.GetFullchainPath(certAlias))
	if localErr != nil {
		// 没有本地证书，执行一次更新证书的逻辑
		slog.Info("no local cert found, attempting to renew", "alias", certAlias)

		status := GlobalScheduler.CheckAndRenewCertStatus(certConfig)

		// 如果操作状态为CERT_RENEW_FAILED直接返回报错
		if status == CERT_RENEW_FAILED {
			return ctx.JSONResponse(FailedWithS("certificate renewal failed", 500), 500)
		}

		// 如果操作状态为CERT_RENEW_SUCCESS，重新检查本地是否有证书
		if status == CERT_RENEW_SUCCESS {
			localExpiry, localErr = cert.GetLocalCertExpiry(CertStorage.GetFullchainPath(certAlias))
		}
	}

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

func VerifyUploadTokenHandler(ctx *atreugo.RequestCtx) error {
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

	return ctx.JSONResponse(SuccessWithM("upload token is valid"))
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

// uploadPageHTML contains the static HTML for certificate upload page
const uploadPageHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Certificate Upload</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: #f5f5f5;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            padding: 30px;
            width: 100%;
            max-width: 400px;
        }
        h2 { color: #333; margin-bottom: 20px; text-align: center; }
        .form-group { margin-bottom: 15px; }
        label { display: block; margin-bottom: 5px; color: #555; font-weight: 500; }
        input[type="text"], input[type="file"] {
            width: 100%;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 14px;
        }
        input[type="file"] { padding: 8px; }
        button {
            width: 100%;
            padding: 12px;
            background: #007bff;
            color: white;
            border: none;
            border-radius: 4px;
            font-size: 16px;
            cursor: pointer;
            transition: background 0.2s;
        }
        button:hover { background: #0056b3; }
        button:disabled { background: #ccc; cursor: not-allowed; }
        .hidden { display: none; }
        .error { color: #dc3545; font-size: 14px; margin-top: 10px; text-align: center; }
        .success { color: #28a745; font-size: 14px; margin-top: 10px; text-align: center; }
        .info { color: #666; font-size: 13px; margin-top: 10px; text-align: center; }
    </style>
</head>
<body>
    <div class="container">
        <h2>Certificate Upload</h2>

        <div id="verify-section">
            <div class="form-group">
                <label for="alias">Certificate Alias</label>
                <input type="text" id="alias" placeholder="Enter certificate alias">
            </div>
            <div class="form-group">
                <label for="token">Upload Token</label>
                <input type="text" id="token" placeholder="Enter upload token">
            </div>
            <button id="verify-btn" onclick="verifyToken()">Verify Token</button>
            <div id="verify-msg"></div>
        </div>

        <div id="upload-section" class="hidden">
            <p class="info">Token verified. Session valid until browser closes.</p>
            <div class="form-group">
                <label for="fullchain">Fullchain Certificate</label>
                <input type="file" id="fullchain" accept=".pem,.crt,.cer">
            </div>
            <div class="form-group">
                <label for="privkey">Private Key</label>
                <input type="file" id="privkey" accept=".pem,.key">
            </div>
            <button id="upload-btn" onclick="uploadCert()">Upload Certificate</button>
            <div id="upload-msg"></div>
        </div>
    </div>

    <script>
        const sessionKey = 'certsync_upload_token';
        const aliasKey = 'certsync_upload_alias';

        function getStoredToken() {
            return sessionStorage.getItem(sessionKey);
        }

        function getStoredAlias() {
            return sessionStorage.getItem(aliasKey);
        }

        function showMessage(elementId, message, isError) {
            const el = document.getElementById(elementId);
            el.textContent = message;
            el.className = isError ? 'error' : 'success';
        }

        async function verifyToken() {
            const alias = document.getElementById('alias').value.trim();
            const token = document.getElementById('token').value.trim();

            if (!alias || !token) {
                showMessage('verify-msg', 'Please fill in all fields', true);
                return;
            }

            const btn = document.getElementById('verify-btn');
            btn.disabled = true;
            btn.textContent = 'Verifying...';

            try {
                const resp = await fetch('/api/' + encodeURIComponent(alias) + '/upload_verify?upload_token=' + encodeURIComponent(token));
                const data = await resp.json();

                if (data.code === 0) {
                    sessionStorage.setItem(sessionKey, token);
                    sessionStorage.setItem(aliasKey, alias);
                    document.getElementById('verify-section').classList.add('hidden');
                    document.getElementById('upload-section').classList.remove('hidden');
                } else {
                    showMessage('verify-msg', data.msg || 'Verification failed', true);
                }
            } catch (e) {
                showMessage('verify-msg', 'Network error: ' + e.message, true);
            } finally {
                btn.disabled = false;
                btn.textContent = 'Verify Token';
            }
        }

        async function uploadCert() {
            const token = getStoredToken();
            const alias = getStoredAlias();

            if (!token || !alias) {
                showMessage('upload-msg', 'Session expired, please verify token again', true);
                document.getElementById('upload-section').classList.add('hidden');
                document.getElementById('verify-section').classList.remove('hidden');
                return;
            }

            const fullchainFile = document.getElementById('fullchain').files[0];
            const privkeyFile = document.getElementById('privkey').files[0];

            if (!fullchainFile || !privkeyFile) {
                showMessage('upload-msg', 'Please select both files', true);
                return;
            }

            const btn = document.getElementById('upload-btn');
            btn.disabled = true;
            btn.textContent = 'Uploading...';

            const formData = new FormData();
            formData.append('fullchain', fullchainFile);
            formData.append('privkey', privkeyFile);

            try {
                const resp = await fetch('/api/' + encodeURIComponent(alias) + '/upload?upload_token=' + encodeURIComponent(token), {
                    method: 'POST',
                    body: formData
                });
                const data = await resp.json();

                if (data.code === 0) {
                    showMessage('upload-msg', data.msg || 'Upload successful', false);
                    document.getElementById('fullchain').value = '';
                    document.getElementById('privkey').value = '';
                } else {
                    showMessage('upload-msg', data.msg || 'Upload failed', true);
                }
            } catch (e) {
                showMessage('upload-msg', 'Network error: ' + e.message, true);
            } finally {
                btn.disabled = false;
                btn.textContent = 'Upload Certificate';
            }
        }

        // Check if already verified in this session
        window.onload = function() {
            const token = getStoredToken();
            const alias = getStoredAlias();
            if (token && alias) {
                document.getElementById('alias').value = alias;
                document.getElementById('token').value = token;
                document.getElementById('verify-section').classList.add('hidden');
                document.getElementById('upload-section').classList.remove('hidden');
            }
        };
    </script>
</body>
</html>`

// UploadPageHandler serves the certificate upload page
func UploadPageHandler(ctx *atreugo.RequestCtx) error {
	ctx.SetContentType("text/html; charset=utf-8")
	return ctx.HTTPResponse(uploadPageHTML)
}
