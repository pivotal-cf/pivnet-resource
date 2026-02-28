package versions

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

const (
	fingerprintDelimiter = "#"
)

// FileMetadata represents a product file's identity and release time for fingerprinting.
// Used to compute a version fingerprint that only changes when product files change.
type FileMetadata struct {
	ID         int
	ReleasedAt string
}

// FingerprintFromFileMetadata returns a deterministic fingerprint from product file metadata.
// Only changes when the set of files or their ReleasedAt values change (not on metadata-only updates).
func FingerprintFromFileMetadata(files []FileMetadata) string {
	if len(files) == 0 {
		return ""
	}
	// Copy and sort by ID for deterministic output
	sorted := make([]FileMetadata, len(files))
	copy(sorted, files)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ID < sorted[j].ID })
	var b strings.Builder
	for _, f := range sorted {
		b.WriteString(fmt.Sprintf("%d:%s\n", f.ID, f.ReleasedAt))
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

func Since(versions []string, since string) ([]string, error) {
	for i, v := range versions {
		if v == since {
			return versions[:i+1], nil
		}
	}

	return versions[:1], nil
}

func Reverse(versions []string) ([]string, error) {
	var reversed []string
	for i := len(versions) - 1; i >= 0; i-- {
		reversed = append(reversed, versions[i])
	}

	return reversed, nil
}

func SplitIntoVersionAndFingerprint(versionWithFingerprint string) (string, string, error) {
	split := strings.Split(versionWithFingerprint, fingerprintDelimiter)
	if len(split) != 2 {
		return "", "", fmt.Errorf("Invalid version and Fingerprint: %s", versionWithFingerprint)
	}
	return split[0], split[1], nil
}

func CombineVersionAndFingerprint(version string, fingerprint string) (string, error) {
	if fingerprint == "" {
		return version, nil
	}
	return combineVersionAndFingerprint(version, fingerprint), nil
}

func combineVersionAndFingerprint(version string, fingerprint string) string {
	return fmt.Sprintf("%s%s%s", version, fingerprintDelimiter, fingerprint)
}
