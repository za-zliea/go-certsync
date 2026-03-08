package client

import (
	"bytes"
	"certsync/src/core/cert"
	"certsync/src/core/meta"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var MetaData *meta.ClientConfig

type CheckResponse struct {
	NeedUpdate   bool   `json:"need_update"`
	LocalExpiry  string `json:"local_expiry"`
	RemoteExpiry string `json:"remote_expiry"`
	Reason       string `json:"reason"`
}

type httpClient struct {
	client *http.Client
}

func newHTTPClient() *httpClient {
	return &httpClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (h *httpClient) doGet(url string, token string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", token)

	rsp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	return io.ReadAll(rsp.Body)
}

func (h *httpClient) doGetWithResponse(url string, token string) ([]byte, *http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", token)

	rsp, err := h.client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		rsp.Body.Close()
		return nil, nil, err
	}

	return body, rsp, nil
}

func Check() (*CheckResponse, error) {
	client := newHTTPClient()

	serverURL := MetaData.Server
	if !strings.HasSuffix(serverURL, "/") {
		serverURL += "/"
	}

	checkURL := fmt.Sprintf("%sapi/%s/check?auth=%s", serverURL, MetaData.CertAlias, MetaData.CertAuth)

	body, err := client.doGet(checkURL, MetaData.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to call check API: %v", err)
	}

	var response ResponseDTO
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	if !response.IsSuccess() {
		return nil, errors.New(response.Message)
	}

	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %v", err)
	}

	var checkResponse CheckResponse
	if err := json.Unmarshal(dataBytes, &checkResponse); err != nil {
		return nil, fmt.Errorf("failed to parse check response: %v", err)
	}

	return &checkResponse, nil
}

func Download() ([]byte, error) {
	client := newHTTPClient()

	serverURL := MetaData.Server
	if !strings.HasSuffix(serverURL, "/") {
		serverURL += "/"
	}

	downloadURL := fmt.Sprintf("%sapi/%s/download?auth=%s", serverURL, MetaData.CertAlias, MetaData.CertAuth)

	body, rsp, err := client.doGetWithResponse(downloadURL, MetaData.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to call download API: %v", err)
	}
	defer rsp.Body.Close()

	contentType := rsp.Header.Get("Content-Type")
	if contentType != "application/zip" {
		var response ResponseDTO
		if err := json.Unmarshal(body, &response); err == nil && !response.IsSuccess() {
			return nil, errors.New(response.Message)
		}
		return nil, fmt.Errorf("unexpected content type: %s", contentType)
	}

	return body, nil
}

func Sync() error {
	slog.Info("Checking certificate", "alias", MetaData.CertAlias)

	checkResponse, err := Check()
	if err != nil {
		return fmt.Errorf("check failed: %v", err)
	}

	slog.Info("Check result", "need_update", checkResponse.NeedUpdate, "reason", checkResponse.Reason)

	if !checkResponse.NeedUpdate {
		slog.Info("Certificate is up to date, skipping sync")
		return nil
	}

	slog.Info("Downloading certificate")
	zipData, err := Download()
	if err != nil {
		return fmt.Errorf("download failed: %v", err)
	}

	slog.Info("Extracting certificate")
	tempDir, err := os.MkdirTemp("", "certsync-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := cert.ExtractCertZip(zipData, tempDir); err != nil {
		return fmt.Errorf("failed to extract certificate: %v", err)
	}

	slog.Info("Copying certificate", "dest", MetaData.CertUpdateDir)
	if err := copyCertificates(tempDir, MetaData.CertUpdateDir); err != nil {
		return fmt.Errorf("failed to copy certificate: %v", err)
	}

	if MetaData.CertUpdateCmd != "" {
		slog.Info("Executing update command", "cmd", MetaData.CertUpdateCmd)
		if err := executeCommand(MetaData.CertUpdateCmd); err != nil {
			return fmt.Errorf("failed to execute update command: %v", err)
		}
	}

	slog.Info("Certificate sync completed successfully")
	return nil
}

func copyCertificates(srcDir, destDir string) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	// Copy all certificate files from source to destination
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		srcPath := filepath.Join(srcDir, entry.Name())
		destPath := filepath.Join(destDir, entry.Name())

		if err := copyFile(srcPath, destPath); err != nil {
			return fmt.Errorf("failed to copy %s: %v", entry.Name(), err)
		}
	}

	return nil
}

func copyFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}

func executeCommand(cmdStr string) error {
	cmd := exec.Command("sh", "-c", cmdStr)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("command failed: %v, stderr: %s", err, stderr.String())
	}

	if stdout.Len() > 0 {
		slog.Info("Command output", "output", stdout.String())
	}

	return nil
}

func buildURL(base, path string, params url.Values) string {
	u, _ := url.Parse(base)
	u.Path = path
	u.RawQuery = params.Encode()
	return u.String()
}
