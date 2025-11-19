package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// SignPayload signs a payload using HMAC-SHA256
func SignPayload(payload interface{}, secret string) (string, error) {
	// Convert payload to JSON
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HMAC
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(jsonBytes)
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return signature, nil
}

// VerifySignature verifies a payload signature
func VerifySignature(payload interface{}, signature, secret string) bool {
	expectedSignature, err := SignPayload(payload, secret)
	if err != nil {
		return false
	}
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// SignString signs a string directly
func SignString(data, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// VerifyStringSignature verifies a string signature
func VerifyStringSignature(data, signature, secret string) bool {
	expectedSignature := SignString(data, secret)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
