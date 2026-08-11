package installer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// Verify checks optional size and SHA-256 constraints for a local file.
func Verify(path, expectedSHA256 string, expectedSize int64) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("verify stat: %w", err)
	}
	if info.Size() == 0 {
		return fmt.Errorf("verify: file kosong (%s)", path)
	}
	if expectedSize > 0 && info.Size() != expectedSize {
		return fmt.Errorf("verify size: got %d, want %d", info.Size(), expectedSize)
	}
	if expectedSHA256 == "" {
		return nil
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("verify hash: %w", err)
	}
	sum := hex.EncodeToString(h.Sum(nil))
	if !equalFoldHex(sum, expectedSHA256) {
		return fmt.Errorf("verify sha256 mismatch")
	}
	return nil
}

func equalFoldHex(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'F' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'F' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}
