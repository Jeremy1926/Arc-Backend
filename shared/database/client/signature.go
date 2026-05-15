package client

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

const (
	HeaderSignature = "X-Arc-Signature"
	HeaderTimestamp = "X-Arc-Timestamp"
	HeaderNonce     = "X-Arc-Nonce"
	MaxClockSkew    = 30 * time.Second
)

var signingKey []byte

func InitSigning(key []byte) {
	signingKey = key
}

func computeSignature(clientID, timestamp, nonce string, body []byte) string {
	mac := hmac.New(sha256.New, signingKey)
	mac.Write([]byte(clientID))
	mac.Write([]byte(":"))
	mac.Write([]byte(timestamp))
	mac.Write([]byte(":"))
	mac.Write([]byte(nonce))
	mac.Write([]byte(":"))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func SignRequest(clientID string, body []byte) (signature, timestamp, nonce string) {
	ts := fmt.Sprintf("%d", time.Now().Unix())
	n := generateNonce()
	sig := computeSignature(clientID, ts, n, body)
	return sig, ts, n
}

func VerifySignature(clientID, signature, timestamp, nonce string, body []byte) error {
	if len(signingKey) == 0 {
		return fmt.Errorf("signing key not initialized")
	}

	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp")
	}

	requestTime := time.Unix(ts, 0)
	diff := time.Since(requestTime)
	if diff < 0 {
		diff = -diff
	}
	if diff > MaxClockSkew {
		return fmt.Errorf("request expired")
	}

	expected := computeSignature(clientID, timestamp, nonce, body)
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

func generateNonce() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
