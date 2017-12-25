package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"io"
)

// Encrypt a message given a secret key.
func Encrypt(key, message []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, aes.BlockSize+len(message))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], message)
	return ciphertext, nil
}

// Decrypt a message given a secret key.
func Decrypt(key, message []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(message) < aes.BlockSize {
		return nil, errors.New("Invalid ciphertext")
	}
	iv := message[:aes.BlockSize]
	message = message[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(message, message)
	return message, nil
}

// Hmac of the message based on the given key.
func Hmac(key, message []byte) []byte {
	mac := hmac.New(sha512.New, key)
	mac.Write(message)
	return mac.Sum(nil)
}

// VerifyHmac verifies the two given HMAC's are equal.
// Returns true if equal, false otherwise.
func VerifyHmac(mac1, mac2 []byte) bool {
	return hmac.Equal(mac1, mac2)
}
