package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

const RevisionPrefix = "sha256:"

func Revision(data []byte) string {
	sum := sha256.Sum256(data)
	return RevisionPrefix + hex.EncodeToString(sum[:])
}

func IsRevision(value string) bool {
	if !strings.HasPrefix(value, RevisionPrefix) || len(value) != len(RevisionPrefix)+64 {
		return false
	}
	_, err := hex.DecodeString(value[len(RevisionPrefix):])
	return err == nil && value == strings.ToLower(value)
}
