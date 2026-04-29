package extractor

import (
	"archive/zip"
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

func DetectFileType(fileName string, data []byte) (string, error) {
	if len(data) == 0 {
		return "", errors.New("empty file")
	}

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(strings.TrimSpace(fileName)), "."))
	if ext == "pdf" && isPDF(data) {
		return "pdf", nil
	}
	if ext == "docx" && isDOCX(data) {
		return "docx", nil
	}
	if ext == "txt" && isLikelyText(data) {
		return "txt", nil
	}

	if isPDF(data) {
		return "pdf", nil
	}
	if isDOCX(data) {
		return "docx", nil
	}
	if isLikelyText(data) {
		return "txt", nil
	}

	return "", errors.New("unsupported file format")
}

func BuildFileName(fileName, fileType string) string {
	base := strings.TrimSpace(fileName)
	if base == "" {
		return "resume." + fileType
	}
	if strings.TrimPrefix(strings.ToLower(filepath.Ext(base)), ".") != fileType {
		return strings.TrimSuffix(base, filepath.Ext(base)) + "." + fileType
	}
	return base
}

func isPDF(data []byte) bool {
	return len(data) >= 5 && bytes.Equal(data[:5], []byte("%PDF-"))
}

func isDOCX(data []byte) bool {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return false
	}
	for _, f := range zr.File {
		if f.Name == docxDocumentXMLPath {
			return true
		}
	}
	return false
}

func isLikelyText(data []byte) bool {
	if !utf8.Valid(data) {
		return false
	}
	for _, b := range data {
		if b == 0 {
			return false
		}
	}
	return true
}
