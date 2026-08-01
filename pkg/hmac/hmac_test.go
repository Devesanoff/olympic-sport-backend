package hmac

import (
	"testing"
	"time"
)

func TestHMACQRToken(t *testing.T) {
	secret := "testsecretsigningkey"
	helper := NewHelper(secret)

	uuidStr := "123e4567-e89b-12d3-a456-426614174000"
	timestamp := time.Now().Unix()

	// 1. Generate Token
	token, err := helper.GenerateQRToken(uuidStr, timestamp)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Fatal("Token is empty")
	}

	// 2. Validate Token (Should succeed)
	parsedUUID, err := helper.ValidateQRToken(token)
	if err != nil {
		t.Fatalf("Failed to validate genuine token: %v", err)
	}

	if parsedUUID != uuidStr {
		t.Errorf("Expected UUID %s, got %s", uuidStr, parsedUUID)
	}

	// 3. Validate Tampered Token (Should fail)
	tamperedToken := token + "a"
	_, err = helper.ValidateQRToken(tamperedToken)
	if err == nil {
		t.Fatal("Expected error when validating tampered token, but got nil")
	}

	// 4. Validate Bad Token Structure (Should fail)
	_, err = helper.ValidateQRToken("invalidtoken")
	if err == nil {
		t.Fatal("Expected error for invalid token structure, but got nil")
	}
}
