package extractor

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	pdf "github.com/ledongthuc/pdf"
)

const docxDocumentXMLPath = "word/document.xml"

func ExtractText(fileType string, data []byte) (string, error) {
	switch strings.ToLower(strings.TrimSpace(fileType)) {
	case "txt":
		return strings.TrimSpace(string(data)), nil
	case "docx":
		return extractDOCXText(data)
	case "pdf":
		return extractPDFText(data)
	default:
		return "", fmt.Errorf("unsupported file type: %s", fileType)
	}
}

func extractPDFText(data []byte) (string, error) {
	tmpFile, err := os.CreateTemp("", "resume-*.pdf")
	if err != nil {
		return "", err
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName)

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return "", err
	}
	if err := tmpFile.Close(); err != nil {
		return "", err
	}

	f, reader, err := pdf.Open(tmpName)
	if err != nil {
		return "", err
	}
	defer f.Close()

	plainReader, err := reader.GetPlainText()
	if err != nil {
		return "", err
	}

	textBytes, err := io.ReadAll(plainReader)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(textBytes)), nil
}

func extractDOCXText(data []byte) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}

	var documentFile *zip.File
	for _, f := range zr.File {
		if f.Name == docxDocumentXMLPath {
			documentFile = f
			break
		}
	}
	if documentFile == nil {
		return "", errors.New("invalid docx: missing word/document.xml")
	}

	rc, err := documentFile.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	decoder := xml.NewDecoder(rc)
	var b strings.Builder
	needsSpace := false

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "p", "br":
				if b.Len() > 0 {
					b.WriteString("\n")
				}
				needsSpace = false
			case "tab":
				b.WriteString("\t")
				needsSpace = false
			}
		case xml.CharData:
			text := strings.TrimSpace(string(t))
			if text == "" {
				continue
			}
			if needsSpace {
				b.WriteString(" ")
			}
			b.WriteString(text)
			needsSpace = true
		}
	}

	return strings.TrimSpace(b.String()), nil
}
