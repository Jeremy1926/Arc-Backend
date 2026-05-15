package anticheat

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
)

var key = []byte{0x47, 0x97, 0xba, 0x3e, 0x7c, 0x8a, 0xc3, 0xf3, 0x2b, 0xc8, 0x79, 0x41, 0x66, 0x13, 0x8b, 0x49, 0x17, 0xe9, 0x1a, 0x67, 0xff, 0xb5, 0xcb, 0x2e, 0x1b, 0xac, 0x08, 0xe4, 0x31, 0x44, 0x08, 0x21}

func Sign(str string, _key []byte) []byte {
	h := hmac.New(sha256.New, _key)
	h.Write([]byte(str))
	return h.Sum(nil)
}

func Decrypt(buf []byte, iv []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	stream := cipher.NewCTR(block, iv)
	decrypted := make([]byte, len(buf))
	stream.XORKeyStream(decrypted, buf)

	return string(decrypted), nil
}

func Encrypt(str string, _key []byte) []byte {
	iv := make([]byte, 16)
	rand.Read(iv)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil
	}

	stream := cipher.NewCTR(block, iv)
	encrypted := make([]byte, len(str))
	stream.XORKeyStream(encrypted, []byte(str))

	signature := Sign(str, _key)

	result := make([]byte, len(encrypted)+len(iv)+len(signature))
	copy(result, encrypted)
	copy(result[len(encrypted):], iv)
	copy(result[len(encrypted)+len(iv):], signature)

	return result
}
