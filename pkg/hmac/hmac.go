package hmac

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// Helper manages secure token signing and verification using HMAC-SHA256.
type Helper struct {
	secretKey []byte
}

// NewHelper creates a new HMAC signing Helper.
func NewHelper(secret string) *Helper {
	return &Helper{
		secretKey: []byte(secret),
	}
}

// GenerateQRToken generates an HMAC-SHA256 signature and returns the token in format "uuid.timestamp.signature".
func (h *Helper) GenerateQRToken(uuidStr string, timestamp int64) (string, error) {
	if uuidStr == "" {
		return "", errors.New("uuid string cannot be empty")
	}

	payload := fmt.Sprintf("%s.%d", uuidStr, timestamp)

	mac := hmac.New(sha256.New, h.secretKey)
	mac.Write([]byte(payload))
	signature := hex.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("%s.%s", payload, signature), nil
}

// ValidateQRToken validates the signature of a token and returns the parsed participant UUID.
func (h *Helper) ValidateQRToken(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", errors.New("invalid token structure")
	}

	uuidStr := parts[0]
	timestampStr := parts[1]
	signatureHex := parts[2]

	payload := fmt.Sprintf("%s.%s", uuidStr, timestampStr)

	mac := hmac.New(sha256.New, h.secretKey)
	mac.Write([]byte(payload))
	expectedSignature := mac.Sum(nil)

	actualSignature, err := hex.DecodeString(signatureHex)
	if err != nil {
		return "", errors.New("invalid signature encoding format")
	}

	if !hmac.Equal(actualSignature, expectedSignature) {
		return "", errors.New("token signature mismatch or tampered")
	}

	return uuidStr, nil
}
