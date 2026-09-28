package z01

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultBanner = "standard"
	MaxInputSize  = 10 * 1024
	MaxTextLength = 250
	MaxLines      = 250
)

// Allowed normalized (LF-only) SHA-256 checksums
var bannerChecksums = map[string]string{
	"standard":   "c3ec7584fb7ecfbd739e6b3f6f63fd1fe557d2ae3e24f870730d9cf8b2559e94",
	"shadow":     "78ccd616680eb9068fe1465db1c852ceaffd8c0f318e3aa0414e1635508e85bf",
	"thinkertoy": "e3c7a11f41a473d9b0d3bf2132a8f6dabb754bd16efa3897fa835a432d3b9caa",
}

// CheckBannerChanges validates banner files against expected checksums regardless of OS line endings.
func CheckBannerChanges() error {
	for name, expectedHash := range bannerChecksums {
		path := filepath.Join("banners", name+".txt")
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("failed to read banner %s: %w", path, err)
		}

		// Normalize line endings (\r\n -> \n) to ensure cross-platform consistency
		normalizedData := bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))

		sum := sha256.Sum256(normalizedData)
		actualHash := hex.EncodeToString(sum[:])

		if actualHash != expectedHash {
			return fmt.Errorf("banner file has been modified: %s (actual hash: %s, expected: %s)", path, actualHash, expectedHash)
		}
	}
	return nil
}