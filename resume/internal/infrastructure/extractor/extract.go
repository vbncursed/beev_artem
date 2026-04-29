package extractor

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"

	pdf "github.com/ledongthuc/pdf"
)

const docxDocumentXMLPath = "word/document.xml"

// MaxExtractedTextBytes caps the size of text extracted from a single resume.
// Defends against zip-bombed DOCX and PDFs with inflated text layers.
const MaxExtractedTextBytes = 2 * 1024 * 1024

var errExtractedTooLarge = errors.New("extracted text exceeds size limit")

// Extractor is a stateless adapter that satisfies usecase.TextExtractor. The
// methods just delegate to the package-level functions kept for backwards
// compatibility — once everything is on the struct we can collapse them.
type Extractor struct{}

func New() *Extractor { return &Extractor{} }

func (Extractor) DetectFileType(fileName string, data []byte) (string, error) {
	return DetectFileType(fileName, data)
}

func (Extractor) ExtractText(fileType string, data []byte) (string, error) {
	return ExtractText(fileType, data)
}

func (Extractor) BuildFileName(fileName, fileType string) string {
	return BuildFileName(fileName, fileType)
}

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
	// In-memory only: pdf.NewReader(io.ReaderAt, size) avoids the os.CreateTemp
	// dance the older code path used. No disk I/O, no leftover temp files if
	// the process panics mid-extraction.
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}

	plainReader, err := reader.GetPlainText()
	if err != nil {
		return "", err
	}

	textBytes, err := io.ReadAll(io.LimitReader(plainReader, MaxExtractedTextBytes+1))
	if err != nil {
		return "", err
	}
	if len(textBytes) > MaxExtractedTextBytes {
		return "", errExtractedTooLarge
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

	// Reject zip-bombs up front: refuse if the declared uncompressed size
	// already exceeds our budget. Defense in depth follows via LimitReader below
	// in case the header lies.
	if documentFile.UncompressedSize64 > MaxExtractedTextBytes {
		return "", errExtractedTooLarge
	}

	rc, err := documentFile.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	decoder := xml.NewDecoder(io.LimitReader(rc, MaxExtractedTextBytes+1))
	var b strings.Builder
	needsSpace := false

	for {
		tok, err := decoder.Token()
		if errors.Is(err, io.EOF) {
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
		if b.Len() > MaxExtractedTextBytes {
			return "", errExtractedTooLarge
		}
	}

	return strings.TrimSpace(b.String()), nil
}
