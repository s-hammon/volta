package hl7crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

func Encrypt(plainText, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", fmt.Errorf("error creating encrypt block: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("cipher.NewGCM: %v", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("error reading block: %v", err)
	}
	cipherText := base64.StdEncoding.EncodeToString(gcm.Seal(nonce, nonce, []byte(plainText), nil))
	return cipherText, nil
}

func Decrypt(cipherText, key string) (plainText string, err error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", fmt.Errorf("error creating encrypt block: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("cipher.NewGCM: %v", err)
	}
	nonceSize := gcm.NonceSize()
	cipherBytes, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", fmt.Errorf("error decoding cipherText: %v", err)
	}
	nonce, cleanedBytes := cipherBytes[:nonceSize], cipherBytes[nonceSize:]
	plainBytes, err := gcm.Open(nil, nonce, cleanedBytes, nil)
	if err != nil {
		return "", fmt.Errorf("gcm.Open: %v", err)
	}
	return string(plainBytes), nil
}
