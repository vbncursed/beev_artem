package domain

import "strings"

// FileType is the closed set of resume formats this service understands.
// Strings match what the database column expects so we can use the type
// directly in storage rows without a converter; behaviour-rich entry point
// for parsing and validation lives here, not in the extractor adapter.
type FileType string

const (
	FileTypeUnknown FileType = ""
	FileTypePDF     FileType = "pdf"
	FileTypeDOCX    FileType = "docx"
	FileTypeTXT     FileType = "txt"
)

// ParseFileType normalises caller input (case, whitespace) and rejects values
// outside the allowed set. Returns FileTypeUnknown + false instead of
// panicking — callers wrap that into ErrInvalidArgument at the use case
// boundary.
func ParseFileType(s string) (FileType, bool) {
	switch FileType(strings.ToLower(strings.TrimSpace(s))) {
	case FileTypePDF:
		return FileTypePDF, true
	case FileTypeDOCX:
		return FileTypeDOCX, true
	case FileTypeTXT:
		return FileTypeTXT, true
	default:
		return FileTypeUnknown, false
	}
}

func (f FileType) String() string { return string(f) }

func (f FileType) IsValid() bool {
	_, ok := ParseFileType(string(f))
	return ok
}
