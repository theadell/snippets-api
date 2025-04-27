package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

type Service struct {
	systemKey []byte
}

func NewService(systemKey string) (*Service, error) {
	key, err := base64.StdEncoding.DecodeString(systemKey)
	if err != nil {
		return nil, fmt.Errorf("invalid system key: %w", err)
	}

	switch len(key) {
	case 16, 24, 32:
	default:
		return nil, errors.New("system key must be 16, 24, or 32 bytes when decoded")
	}

	return &Service{
		systemKey: key,
	}, nil
}

// Encrypt encrypts data using AES-GCM
func (s *Service) Encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.systemKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	// prepend nonce to the encrypted text
	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	return ciphertext, nil
}

// Decrypt decrypts data using AES-GCM
func (s *Service) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.systemKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()

	if len(ciphertext) < nonceSize {
		return nil, errors.New("cipher text is too short")
	}
	nonce, encryptedData := ciphertext[:nonceSize], ciphertext[nonceSize:]

	return gcm.Open(nil, nonce, encryptedData, nil)
}
