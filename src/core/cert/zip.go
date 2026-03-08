package cert

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
)

// CertFiles contains all certificate files for packaging
type CertFiles struct {
	Cert      []byte
	Chain     []byte
	Fullchain []byte
	Privkey   []byte
}

// CreateCertZip creates a ZIP archive containing all certificate files
func CreateCertZip(files CertFiles) ([]byte, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// Add cert.pem (leaf certificate)
	if len(files.Cert) > 0 {
		writer, err := zipWriter.Create("cert.pem")
		if err != nil {
			return nil, err
		}
		if _, err := writer.Write(files.Cert); err != nil {
			return nil, err
		}
	}

	// Add chain.pem (intermediate certificates)
	if len(files.Chain) > 0 {
		writer, err := zipWriter.Create("chain.pem")
		if err != nil {
			return nil, err
		}
		if _, err := writer.Write(files.Chain); err != nil {
			return nil, err
		}
	}

	// Add fullchain.pem (leaf + intermediates)
	if len(files.Fullchain) > 0 {
		writer, err := zipWriter.Create("fullchain.pem")
		if err != nil {
			return nil, err
		}
		if _, err := writer.Write(files.Fullchain); err != nil {
			return nil, err
		}
	}

	// Add privkey.pem (private key)
	if len(files.Privkey) > 0 {
		writer, err := zipWriter.Create("privkey.pem")
		if err != nil {
			return nil, err
		}
		if _, err := writer.Write(files.Privkey); err != nil {
			return nil, err
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ExtractCertZip extracts all files from a ZIP archive to the destination directory
func ExtractCertZip(zipData []byte, destDir string) error {
	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return err
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	for _, file := range reader.File {
		destPath := filepath.Join(destDir, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(destPath, 0755)
			continue
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		srcFile, err := file.Open()
		if err != nil {
			destFile.Close()
			return err
		}

		_, err = io.Copy(destFile, srcFile)
		srcFile.Close()
		destFile.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
