package encoding

import (
	"fmt"
	"io"
	"os"

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

	// Read first 1024 bytes for detection
	buffer := make([]byte, 1024)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	result, err := d.detector.DetectBest(buffer[:n])
	if err != nil {
		return "", fmt.Errorf("failed to detect encoding: %w", err)
	}

	if result == nil {
		return "UTF-8", nil // Default to UTF-8 if detection fails
	}

	return result.Charset, nil
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