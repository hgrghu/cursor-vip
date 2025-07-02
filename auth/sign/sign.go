package sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

// applicationKey is the secret key for signing
const applicationKey = "cursor-vip-2024-secret-key"

// Sign generates a signature for the given device ID
func Sign(deviceID string) string {
	// Create a timestamp-based signature
	timestamp := time.Now().Unix()
	
	// Create the signature payload
	payload := fmt.Sprintf("%s:%d", deviceID, timestamp)
	
	// Generate HMAC-SHA256 signature
	signature := generateHMAC(payload, applicationKey)
	
	// Return timestamp:signature format
	return fmt.Sprintf("%d:%s", timestamp, signature)
}

// generateHMAC generates HMAC-SHA256 hash
func generateHMAC(data, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// ValidateSign validates a signature against a device ID
func ValidateSign(deviceID, signature string) bool {
	// Parse the signature
	parts := parseSignature(signature)
	if len(parts) != 2 {
		return false
	}
	
	timestampStr, signatureHash := parts[0], parts[1]
	
	// Parse timestamp
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return false
	}
	
	// Check if signature is not too old (5 minutes)
	now := time.Now().Unix()
	if now-timestamp > 300 {
		return false
	}
	
	// Recreate the expected signature
	payload := fmt.Sprintf("%s:%d", deviceID, timestamp)
	expectedSignature := generateHMAC(payload, applicationKey)
	
	// Compare signatures
	return hmac.Equal([]byte(signatureHash), []byte(expectedSignature))
}

// parseSignature splits signature into timestamp and hash parts
func parseSignature(signature string) []string {
	for i := 0; i < len(signature); i++ {
		if signature[i] == ':' {
			return []string{signature[:i], signature[i+1:]}
		}
	}
	return []string{}
}

// SignWithCustomKey generates a signature with a custom key
func SignWithCustomKey(deviceID, customKey string) string {
	timestamp := time.Now().Unix()
	payload := fmt.Sprintf("%s:%d", deviceID, timestamp)
	signature := generateHMAC(payload, customKey)
	return fmt.Sprintf("%d:%s", timestamp, signature)
}

// GetSignatureInfo extracts information from a signature
func GetSignatureInfo(signature string) map[string]interface{} {
	parts := parseSignature(signature)
	if len(parts) != 2 {
		return map[string]interface{}{
			"valid": false,
			"error": "invalid signature format",
		}
	}
	
	timestampStr, hash := parts[0], parts[1]
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return map[string]interface{}{
			"valid":     false,
			"error":     "invalid timestamp",
			"timestamp": timestampStr,
		}
	}
	
	return map[string]interface{}{
		"valid":     true,
		"timestamp": timestamp,
		"hash":      hash,
		"age":       time.Now().Unix() - timestamp,
	}
}