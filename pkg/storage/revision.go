package storage

import (
	"crypto/sha256"
	"encoding/hex"
)

const RevisionPrefix = "sha256:"

func Revision(data []byte) string {
	sum := sha256.Sum256(data)
	return RevisionPrefix + hex.EncodeToString(sum[:])
}
