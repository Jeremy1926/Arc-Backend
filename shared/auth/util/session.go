package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/Arc-Services/Arc/shared/database"
)

var SECRET_KEY = []byte{0x85, 0x3a, 0x1f, 0x7c, 0x2d, 0x9e, 0xab, 0x4b, 0x6f, 0x12, 0x34, 0x56, 0x78, 0x90, 0xaa, 0xbb,
	0xcc, 0xdd, 0xee, 0xff, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88,
	0x99, 0x00, 0xab, 0xcd, 0xef}

var b32 = base32.StdEncoding.WithPadding(base32.NoPadding)

func deriveKey(salt []byte) []byte {
	h := sha256.New()
	h.Write(SECRET_KEY)
	h.Write(salt)
	return h.Sum(nil)
}

func HashSession(session *database.Session) (string, error) {
	payload, err := json.Marshal(session)
	if err != nil {
		return "", err
	}

	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}

	key := deriveKey(salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nil, nonce, payload, nil)

	mac := hmac.New(sha512.New, SECRET_KEY)
	mac.Write(ciphertext)
	signature := mac.Sum(nil)

	combined := make([]byte, 0, len(salt)+len(nonce)+len(ciphertext)+len(signature))
	combined = append(combined, salt...)
	combined = append(combined, nonce...)
	combined = append(combined, ciphertext...)
	combined = append(combined, signature...)

	layer1 := base64.RawURLEncoding.EncodeToString(combined)
	layer2 := hex.EncodeToString([]byte(layer1))
	layer3 := strings.ToLower(b32.EncodeToString([]byte(layer2)))

	return obfuscate(layer3), nil
}

func UnhashSession(token string) (*database.Session, error) {
	if token == "" {
		return nil, errors.New("empty token")
	}

	layer3, err := deobfuscate(token)
	if err != nil {
		return nil, err
	}

	return decodeAndDecrypt(layer3)
}

func decodeAndDecrypt(layer3 string) (*database.Session, error) {
	layer2Bytes, err := b32.DecodeString(strings.ToUpper(layer3))
	if err != nil {
		return nil, err
	}
	layer2 := string(layer2Bytes)

	if len(layer2)%2 != 0 {
		return nil, errors.New("invalid hex length")
	}
	layer1Bytes, err := hex.DecodeString(layer2)
	if err != nil {
		return nil, err
	}
	layer1 := string(layer1Bytes)

	combined, err := base64.RawURLEncoding.DecodeString(layer1)
	if err != nil {
		return nil, err
	}

	if len(combined) < 16+12+64 {
		return nil, errors.New("token too short")
	}

	salt := combined[:16]

	key := deriveKey(salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	minLen := 16 + nonceSize + 64
	if len(combined) < minLen {
		return nil, errors.New("token too short for nonce/signature")
	}

	nonce := combined[16 : 16+nonceSize]
	body := combined[16+nonceSize:]
	if len(body) < 64 {
		return nil, errors.New("missing signature")
	}

	ciphertext := body[:len(body)-64]
	signature := body[len(body)-64:]

	mac := hmac.New(sha512.New, SECRET_KEY)
	mac.Write(ciphertext)
	expected := mac.Sum(nil)
	if !hmac.Equal(expected, signature) {
		return nil, errors.New("invalid signature")
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	var session database.Session
	if err := json.Unmarshal(plaintext, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

const obfInsertCount = 8

func obfuscate(input string) string {
	if len(input) == 0 {
		return input
	}

	positions, letters := obfPlan(len(input))

	var b strings.Builder
	b.Grow(len(input) + obfInsertCount)

	for i := 0; i < len(input); i++ {
		if positions[i] {
			b.WriteByte(letters[i%len(letters)])
		}
		b.WriteByte(input[i])
	}

	return b.String()
}

func deobfuscate(obf string) (string, error) {
	if len(obf) < obfInsertCount {
		return "", errors.New("token too short")
	}

	origLen := len(obf) - obfInsertCount
	if origLen <= 0 {
		return "", errors.New("invalid token length")
	}

	positions, _ := obfPlan(origLen)

	var b strings.Builder
	b.Grow(origLen)

	// if positions[i] => skip 1 obf char, then take 1 real char cuz like yeah
	j := 0
	for i := 0; i < origLen; i++ {
		if positions[i] {
			if j >= len(obf) {
				return "", errors.New("malformed token")
			}
			j++
		}
		if j >= len(obf) {
			return "", errors.New("malformed token")
		}
		b.WriteByte(obf[j])
		j++
	}

	if j != len(obf) {
		return "", errors.New("malformed token")
	}

	return b.String(), nil
}

func obfPlan(origLen int) (map[int]bool, []byte) {
	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], uint32(origLen))

	h := sha256.New()
	h.Write(SECRET_KEY)
	h.Write(lenBuf[:])
	sum := h.Sum(nil)

	positions := make(map[int]bool, obfInsertCount)

	for i := 0; i < obfInsertCount; i++ {
		if origLen == 1 {
			positions[0] = true
			continue
		}
		p := int(sum[i]) % origLen
		for positions[p] {
			p = (p + 1) % origLen
		}
		positions[p] = true
	}

	letters := make([]byte, 32)
	for i := 0; i < 32; i++ {
		letters[i] = byte('a' + (sum[i] % 26))
	}

	return positions, letters
}
