package tool

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
	
	"github.com/kingparks/cursor-vip/tui/params"
	"github.com/tidwall/gjson"
)

const (
	ConfigFileName = ".cursor-viprc"
	ConfigVersion  = "1.1"
)

// Enhanced configuration structure
type EnhancedConfig struct {
	Version     string            `json:"version"`
	Lang        string            `json:"lang"`
	Mode        int64             `json:"mode"`
	Promotion   string            `json:"promotion"`
	HTTPTimeout int               `json:"http_timeout_seconds"`
	MaxRetries  int               `json:"max_retries"`
	LogLevel    string            `json:"log_level"`
	LastUpdated time.Time         `json:"last_updated"`
	Features    map[string]bool   `json:"features"`
	Advanced    map[string]string `json:"advanced"`
}

// Default configuration values
func getDefaultConfig() *EnhancedConfig {
	return &EnhancedConfig{
		Version:     ConfigVersion,
		Lang:        "en",
		Mode:        2,
		Promotion:   "",
		HTTPTimeout: 30,
		MaxRetries:  3,
		LogLevel:    "info",
		LastUpdated: time.Now(),
		Features: map[string]bool{
			"auto_update":     true,
			"error_reporting": true,
			"analytics":       false,
		},
		Advanced: map[string]string{
			"proxy_mode": "auto",
			"theme":      "default",
		},
	}
}

// Validate configuration values
func (c *EnhancedConfig) Validate() error {
	validLangs := map[string]bool{
		"en": true, "zh": true, "nl": true, "ru": true, 
		"hu": true, "tr": true, "es": true,
	}
	
	if !validLangs[c.Lang] {
		return &params.AppError{
			Code:    400,
			Message: fmt.Sprintf("invalid language: %s", c.Lang),
		}
	}
	
	if c.Mode < 1 || c.Mode > 4 {
		return &params.AppError{
			Code:    400,
			Message: fmt.Sprintf("invalid mode: %d (must be 1-4)", c.Mode),
		}
	}
	
	if c.HTTPTimeout < 5 || c.HTTPTimeout > 300 {
		return &params.AppError{
			Code:    400,
			Message: fmt.Sprintf("invalid HTTP timeout: %d (must be 5-300 seconds)", c.HTTPTimeout),
		}
	}
	
	if c.MaxRetries < 1 || c.MaxRetries > 10 {
		return &params.AppError{
			Code:    400,
			Message: fmt.Sprintf("invalid max retries: %d (must be 1-10)", c.MaxRetries),
		}
	}
	
	return nil
}

// Get configuration file path
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ConfigFileName), nil
}

// Generate encryption key from device ID
func getEncryptionKey() []byte {
	deviceID := GetMachineID()
	hash := sha256.Sum256([]byte(deviceID + "cursor-vip-salt"))
	return hash[:]
}

// Encrypt configuration data
func encryptConfig(data []byte) (string, error) {
	key := getEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	
	encrypted := gcm.Seal(nonce, nonce, data, nil)
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// Decrypt configuration data
func decryptConfig(encryptedData string) ([]byte, error) {
	key := getEncryptionKey()
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}
	
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("encrypted data too short")
	}
	
	nonce, encrypted := data[:nonceSize], data[nonceSize:]
	decrypted, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return nil, err
	}
	
	return decrypted, nil
}

// Enhanced GetConfig with better error handling and validation
func GetEnhancedConfig() (*EnhancedConfig, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}
	
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		defaultConfig := getDefaultConfig()
		if err := SaveEnhancedConfig(defaultConfig); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return defaultConfig, nil
	}
	
	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	var config *EnhancedConfig
	
	// Try to parse as encrypted config first
	if decrypted, err := decryptConfig(string(data)); err == nil {
		config = &EnhancedConfig{}
		if err := json.Unmarshal(decrypted, config); err == nil {
			goto validate
		}
	}
	
	// Fall back to legacy plain text config
	config = &EnhancedConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		// Try legacy format
		config = migrateFromLegacy(string(data))
	}
	
validate:
	// Set defaults for missing fields
	if config.Version == "" {
		config.Version = ConfigVersion
	}
	if config.HTTPTimeout == 0 {
		config.HTTPTimeout = 30
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.LogLevel == "" {
		config.LogLevel = "info"
	}
	if config.Features == nil {
		config.Features = getDefaultConfig().Features
	}
	if config.Advanced == nil {
		config.Advanced = getDefaultConfig().Advanced
	}
	
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	return config, nil
}

// Save enhanced configuration with encryption
func SaveEnhancedConfig(config *EnhancedConfig) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	
	config.LastUpdated = time.Now()
	
	configPath, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}
	
	// Marshal config to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	// Encrypt config data
	encrypted, err := encryptConfig(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt config: %w", err)
	}
	
	// Write to file with proper permissions
	if err := os.WriteFile(configPath, []byte(encrypted), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// Migrate from legacy configuration format
func migrateFromLegacy(data string) *EnhancedConfig {
	config := getDefaultConfig()
	
	// Parse legacy format
	if lang := gjson.Get(data, "lang").String(); lang != "" {
		config.Lang = lang
	}
	if mode := gjson.Get(data, "mode").Int(); mode != 0 {
		config.Mode = mode
	}
	if promotion := gjson.Get(data, "promotion").String(); promotion != "" {
		config.Promotion = promotion
	}
	
	return config
}

// Backup configuration
func BackupConfig() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}
	
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil // No config to backup
	}
	
	backupPath := configPath + ".backup." + time.Now().Format("20060102150405")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	
	return os.WriteFile(backupPath, data, 0600)
}

// Update legacy GetConfig to use enhanced version
func GetConfig() (lang, promotion string, mode int64) {
	config, err := GetEnhancedConfig()
	if err != nil {
		// Fall back to old behavior
		lang, _ = GetLocale()
		if lang == "" {
			lang = "en"
		}
		mode = 2
		return
	}
	
	return config.Lang, config.Promotion, config.Mode
}

// Update legacy SetConfig to use enhanced version
func SetConfig(lang string, mode int64) {
	config, err := GetEnhancedConfig()
	if err != nil {
		config = getDefaultConfig()
	}
	
	config.Lang = lang
	config.Mode = mode
	config.Promotion = params.Promotion
	
	_ = SaveEnhancedConfig(config)
}