package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

// Encryption type stores secret key based on which data is being encrypted
type Encryption struct {
	Key []byte // Secret Key
}

// Encrypt encrypts text based on a secret key.
func (e *Encryption) Encrypt(text string) (string, error) {
	plainText := []byte(text)

	block, err := aes.NewCipher(e.Key)
	if err != nil {
		return "", err
	}

	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	// Encode cipherText (i.e. slice of bytes) to text, so it is safe to use on web pages
	return base64.URLEncoding.EncodeToString(cipherText), nil
}

// Decrypt decrypts previously encrypted text based on a specific secret key.
func (e *Encryption) Decrypt(cryptoText string) (string, error) {
	cipherText, err := base64.URLEncoding.DecodeString(cryptoText)
	if err != nil {
		return "", nil
	}

	block, err := aes.NewCipher(e.Key)
	if err != nil {
		return "", nil
	}

	if len(cipherText) < aes.BlockSize {
		return "", err
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return string(cipherText), nil
}
