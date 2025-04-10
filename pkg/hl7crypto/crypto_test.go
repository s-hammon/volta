package hl7crypto

import (
	"crypto/rand"
	"testing"
)

func TestCrypto(t *testing.T) {
	wantText := "Graham"
	key := genKey(t, 32)

	cipherText, err := Encrypt(wantText, key)
	if err != nil {
		t.Fatal(err)
	}
	plainText, err := Decrypt(cipherText, key)
	if err != nil {
		t.Fatal(err)
	}
	if plainText != wantText {
		t.Errorf("decrypt returned '%s', expected '%s'", plainText, plainText)
	}
}

func TestCrypto_Error(t *testing.T) {
	tests := []struct {
		name      string
		plainText string
		key       string
		valid     bool
	}{
		{"invalid key length", "Graham", genKey(t, 33), false},
		{"incorrect key", "Graham", genKey(t, 32), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.valid {
				cipherText, err := Encrypt(tt.plainText, tt.key)
				if err == nil {
					t.Fatal("encrypt(): expected error for key length = 33, but got nil")
				}
				if _, err := Decrypt(cipherText, tt.key); err == nil {
					t.Fatal("decrypt(): expected error for key length = 33, but got nil")
				}
			} else {
				cipherText, err := Encrypt(tt.plainText, tt.key)
				if err != nil {
					t.Fatalf("got unexpected error: %v", err)
				}
				if _, err := Decrypt(cipherText, genKey(t, 32)); err == nil {
					t.Fatalf("decrypt(): expected error for wrong key, but got nil")
				}
			}
		})
	}
}

func genKey(t *testing.T, size int) string {
	key := make([]byte, size)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("error generating key: %v", err)
	}
	return string(key)
}
