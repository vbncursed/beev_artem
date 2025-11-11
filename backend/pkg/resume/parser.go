package resume

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	pdf "github.com/ledongthuc/pdf"
)

// ParseResumeText extracts plain text from supported resume formats.
// Supports: .pdf and .docx
func ParseResumeText(filename string, data []byte) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".pdf":
		return extractTextFromPDF(data)
	case ".docx":
		return extractTextFromDocx(data)
	default:
		return "", errors.New("unsupported file format: only pdf and docx are allowed")
	}
}

func extractTextFromPDF(data []byte) (string, error) {
	reader := bytes.NewReader(data)
	r, err := pdf.NewReader(reader, int64(len(data)))
	if err != nil {
		return "", err
	}
	rs, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if _, err = io.Copy(&buf, rs); err != nil {
		return "", err
	}
	return normalizeWhitespace(buf.String()), nil
}

func extractTextFromDocx(data []byte) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	var docXML []byte
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				return "", err
			}
			defer rc.Close()
			docXML, err = io.ReadAll(rc)
			if err != nil {
				return "", err
			}
			break
		}
	}
	if len(docXML) == 0 {
		return "", errors.New("no document.xml found in docx")
	}
	xml := string(docXML)
	// Convert paragraph boundaries to newlines (very naive but effective).
	xml = strings.ReplaceAll(xml, "</w:p>", "\n")
	xml = strings.ReplaceAll(xml, "<w:tab/>", "\t")
	// Remove all XML tags.
	reTags := regexp.MustCompile(`<[^>]+>`)
	txt := reTags.ReplaceAllString(xml, " ")
	return normalizeWhitespace(txt), nil
}

func normalizeWhitespace(s string) string {
	// Collapse excessive whitespace and trim
	re := regexp.MustCompile(`[ \t\r\f\v]+`)
	s = re.ReplaceAllString(s, " ")
	s = strings.ReplaceAll(s, "\u00A0", " ")
	// Preserve newlines but collapse runs
	reN := regexp.MustCompile(`\n+`)
	s = reN.ReplaceAllString(s, "\n")
	return strings.TrimSpace(s)
}
