package extractor

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strings"
	"time"

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

// pdftotextTimeout caps a single extraction. Generous: a 50-page CV
// takes ~200ms; we allow 30s as the worst plausible case before the
// process is forcibly killed.
const pdftotextTimeout = 30 * time.Second

// extractPDFText prefers the external `pdftotext` binary (poppler-utils)
// because the in-process ledongthuc/pdf cannot recover word boundaries on
// PDFs that lack positional metadata — common in Cyrillic résumés where
// the result is gibberish like "ЭдуардКурочкинGo-разработчик".
//
// We fall back to ledongthuc/pdf if pdftotext is missing or errors out,
// so dev environments without poppler installed still produce something.
// The fallback is best-effort and may return joined-words; the primary
// path is the source of truth.
func extractPDFText(data []byte) (string, error) {
	if text, err := extractViaPdftotext(data); err == nil {
		return text, nil
	} else {
		slog.Warn("pdftotext failed, falling back to in-process pdf parser",
			"err", err.Error())
	}
	return extractViaLedongthuc(data)
}

// extractViaPdftotext shells out to `pdftotext - -` (stdin → stdout).
// The `-layout` flag preserves visual columns for tabular résumés;
// `-enc UTF-8` forces UTF-8 output regardless of the PDF's encoding.
func extractViaPdftotext(data []byte) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), pdftotextTimeout)
	defer cancel()

	// `-` for input + `-` for output → stream both via stdin/stdout, no
	// temp files. -nopgbrk drops the form-feed page separators that would
	// otherwise pollute the heuristic year/skill extractors downstream.
	cmd := exec.CommandContext(ctx, "pdftotext",
		"-layout", "-nopgbrk", "-enc", "UTF-8", "-", "-")
	cmd.Stdin = bytes.NewReader(data)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pdftotext: %w (stderr=%q)", err, stderr.String())
	}

	if stdout.Len() > MaxExtractedTextBytes {
		return "", errExtractedTooLarge
	}
	out := strings.TrimSpace(stdout.String())
	if out == "" {
		return "", errors.New("pdftotext: empty output")
	}
	return out, nil
}

// extractViaLedongthuc is the fallback path used only when pdftotext is
// unavailable. Returns whatever GetPlainText produces — often missing
// spaces, but better than nothing while ops install poppler-utils.
func extractViaLedongthuc(data []byte) (string, error) {
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
