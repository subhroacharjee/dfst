package store

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func CreateChecksum(data []byte) string {
	hash := sha512.Sum512(data)
	return fmt.Sprintf("%x", hash)
}

func CreateFileChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hasher := sha512.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
