package util

import (
	"crypto/rand"
	"os"
)

func GenerateKey() ([]byte, error) {
	key := make([]byte, 32) // 32 bytes is a good default for AES keys
	_, err := rand.Read(key)
	return key, err
}

func WriteKeyToFile(key string, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(key)
	return err
}
