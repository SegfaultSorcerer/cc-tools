package encoding

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// Detector handles character encoding detection and conversion
type Detector struct {
	detector *chardet.Detector
}

// NewDetector creates a new encoding detector
func NewDetector() *Detector {
	return &Detector{
		detector: chardet.NewTextDetector(),
	}
}

// DetectFileEncoding detects the encoding of a file
func (d *Detector) DetectFileEncoding(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read larger sample for better detection (4096 bytes)
	buffer := make([]byte, 4096)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	result, err := d.detector.DetectBest(buffer[:n])
	if err != nil {
		return "", fmt.Errorf("failed to detect encoding: %w", err)
	}

	if result == nil {
		// Enhanced detection for Pascal/Delphi files
		if d.looksLikePascalFile(buffer[:n]) {
			return "ISO-8859-1", nil // Common encoding for legacy Delphi files
		}
		return "UTF-8", nil // Default to UTF-8 if detection fails
	}

	// Additional check for Pascal files that might be misdetected
	if d.looksLikePascalFile(buffer[:n]) && (result.Charset == "UTF-8" || result.Charset == "ascii") {
		// For Pascal files with ASCII/UTF-8 detection, check if it might actually be ISO-8859-1
		if d.containsExtendedLatin1(buffer[:n]) {
			return "ISO-8859-1", nil
		}
	}

	return result.Charset, nil
}

// looksLikePascalFile checks if the content appears to be Pascal/Delphi code
func (d *Detector) looksLikePascalFile(data []byte) bool {
	content := string(data)
	keywords := []string{"unit", "interface", "implementation", "program", "procedure", "function", "type", "var", "const", "begin", "end."}

	for _, keyword := range keywords {
		if strings.Contains(strings.ToLower(content), keyword) {
			return true
		}
	}
	return false
}

// containsExtendedLatin1 checks for characters that suggest ISO-8859-1 encoding
func (d *Detector) containsExtendedLatin1(data []byte) bool {
	for _, b := range data {
		if b >= 0x80 && b <= 0xFF { // Extended Latin-1 range
			return true
		}
	}
	return false
}

// GetEncoder returns the appropriate encoder for the given charset
func GetEncoder(charset string) (encoding.Encoding, error) {
	switch charset {
	case "UTF-8", "utf-8":
		return unicode.UTF8, nil
	case "UTF-16LE", "utf-16le":
		return unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM), nil
	case "UTF-16BE", "utf-16be":
		return unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM), nil
	case "ISO-8859-1", "iso-8859-1":
		return charmap.ISO8859_1, nil
	case "ISO-8859-15", "iso-8859-15":
		return charmap.ISO8859_15, nil
	case "Windows-1252", "windows-1252":
		return charmap.Windows1252, nil
	case "Windows-1251", "windows-1251":
		return charmap.Windows1251, nil
	case "GB18030", "gb18030":
		return simplifiedchinese.GB18030, nil
	case "GBK", "gbk":
		return simplifiedchinese.GBK, nil
	case "Big5", "big5":
		return traditionalchinese.Big5, nil
	case "Shift_JIS", "shift_jis":
		return japanese.ShiftJIS, nil
	case "EUC-JP", "euc-jp":
		return japanese.EUCJP, nil
	case "EUC-KR", "euc-kr":
		return korean.EUCKR, nil
	default:
		// Try to default to UTF-8 for unknown encodings
		return unicode.UTF8, nil
	}
}

// DecodeBytes converts bytes from the source encoding to UTF-8
func DecodeBytes(data []byte, sourceEncoding string) (string, error) {
	enc, err := GetEncoder(sourceEncoding)
	if err != nil {
		return "", err
	}

	if sourceEncoding == "UTF-8" || sourceEncoding == "utf-8" {
		return string(data), nil
	}

	decoder := enc.NewDecoder()
	result, _, err := transform.Bytes(decoder, data)
	if err != nil {
		return "", fmt.Errorf("failed to decode bytes: %w", err)
	}

	return string(result), nil
}

// EncodeString converts a UTF-8 string to the target encoding
func EncodeString(text string, targetEncoding string) ([]byte, error) {
	enc, err := GetEncoder(targetEncoding)
	if err != nil {
		return nil, err
	}

	if targetEncoding == "UTF-8" || targetEncoding == "utf-8" {
		return []byte(text), nil
	}

	encoder := enc.NewEncoder()
	result, _, err := transform.Bytes(encoder, []byte(text))
	if err != nil {
		return nil, fmt.Errorf("failed to encode string: %w", err)
	}

	return result, nil
}