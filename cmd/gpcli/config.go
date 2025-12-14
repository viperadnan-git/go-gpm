package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"github.com/pelletier/go-toml/v2"
)

// CachedToken holds the cached access token and expiry
type CachedToken struct {
	Token  string `toml:"token"`
	Expiry int64  `toml:"expiry"`
}

// AccountConfig holds per-account settings
type AccountConfig struct {
	Auth          string       `toml:"auth"`           // Auth string (androidId, Token, Email, etc.)
	AuthToken     *CachedToken `toml:"auth_token"`     // Cached access token
	Quality       string       `toml:"quality"`        // "original" or "storage-saver"
	UseQuota      bool         `toml:"use_quota"`      // If true, uploads count against storage quota
	UploadThreads int          `toml:"upload_threads"` // Number of upload threads
	Proxy         string       `toml:"proxy"`          // Proxy URL
}

// Config represents the TOML configuration
type Config struct {
	Selected string                    `toml:"selected"` // Currently selected account email
	Accounts map[string]*AccountConfig `toml:"accounts"` // Map of email -> account config
}

// DefaultAccountConfig returns the default account configuration
func DefaultAccountConfig() *AccountConfig {
	return &AccountConfig{
		Quality:       "original",
		UploadThreads: 3,
	}
}

// ConfigManager manages configuration loading and saving
type ConfigManager struct {
	config     Config
	configPath string
	mu         sync.RWMutex
}

// NewConfigManager creates a new ConfigManager and loads the configuration
func NewConfigManager(configPath string) (*ConfigManager, error) {
	if configPath != "" {
		absPath, err := filepath.Abs(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve path: %w", err)
		}
		configPath = absPath
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}

		xdgPath := filepath.Join(homeDir, ".config", "gpcli", "gpcli.toml")
		localPath, err := filepath.Abs("gpcli.toml")
		if err != nil {
			return nil, fmt.Errorf("failed to resolve path: %w", err)
		}

		// Check local path first
		if _, err := os.Stat(localPath); err == nil {
			configPath = localPath
		} else if _, err := os.Stat(xdgPath); err == nil {
			configPath = xdgPath
		} else {
			// Neither exists, use XDG path for new config
			configPath = xdgPath
		}
	}

	m := &ConfigManager{
		configPath: configPath,
		config: Config{
			Accounts: make(map[string]*AccountConfig),
		},
	}

	// Load config from file if it exists
	if data, err := os.ReadFile(configPath); err == nil && len(data) > 0 {
		if err := toml.Unmarshal(data, &m.config); err != nil {
			return nil, fmt.Errorf("failed to parse config: %w", err)
		}
		// Initialize map if nil
		if m.config.Accounts == nil {
			m.config.Accounts = make(map[string]*AccountConfig)
		}
	}

	return m, nil
}

// GetConfig returns the current configuration
func (m *ConfigManager) GetConfig() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// GetConfigPath returns the path to the config file
func (m *ConfigManager) GetConfigPath() string {
	return m.configPath
}

// GetSelectedAccount returns the currently selected account config
func (m *ConfigManager) GetSelectedAccount() *AccountConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.config.Selected == "" {
		return nil
	}
	return m.config.Accounts[m.config.Selected]
}

// Save persists the current configuration to disk
func (m *ConfigManager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := toml.Marshal(m.config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Create directory if it doesn't exist
	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// SetSelected updates the selected email
func (m *ConfigManager) SetSelected(email string) error {
	m.mu.Lock()
	if _, exists := m.config.Accounts[email]; !exists {
		m.mu.Unlock()
		return fmt.Errorf("account %s does not exist", email)
	}
	m.config.Selected = email
	m.mu.Unlock()
	return m.Save()
}

// AddCredentials adds a new account with the given auth string
func (m *ConfigManager) AddCredentials(authString string) (email string, err error) {
	requiredFields := []string{"androidId", "app", "client_sig", "Email", "Token", "lang", "service"}

	params, err := url.ParseQuery(authString)
	if err != nil {
		return "", fmt.Errorf("invalid auth string format: %w", err)
	}

	// Validate required fields
	var missingFields []string
	for _, field := range requiredFields {
		if params.Get(field) == "" {
			missingFields = append(missingFields, field)
		}
	}
	if len(missingFields) > 0 {
		return "", fmt.Errorf("auth string missing required fields: %v", missingFields)
	}

	email = params.Get("Email")
	if email == "" {
		return "", fmt.Errorf("email cannot be empty")
	}

	m.mu.Lock()
	if _, exists := m.config.Accounts[email]; exists {
		m.mu.Unlock()
		return "", fmt.Errorf("account %s already exists", email)
	}

	// Create new account with defaults
	account := DefaultAccountConfig()
	account.Auth = authString
	m.config.Accounts[email] = account
	m.config.Selected = email
	m.mu.Unlock()

	if err := m.Save(); err != nil {
		return "", err
	}

	return email, nil
}

// RemoveCredentials removes an account by email
func (m *ConfigManager) RemoveCredentials(email string) error {
	m.mu.Lock()
	if _, exists := m.config.Accounts[email]; !exists {
		m.mu.Unlock()
		return fmt.Errorf("account %s does not exist", email)
	}

	delete(m.config.Accounts, email)

	// Clear selection if we're removing the selected account
	if m.config.Selected == email {
		m.config.Selected = ""
		// Select first available account if any
		for e := range m.config.Accounts {
			m.config.Selected = e
			break
		}
	}
	m.mu.Unlock()

	return m.Save()
}

// UpdateAccountToken updates the cached auth token for an account
func (m *ConfigManager) UpdateAccountToken(email, token string, expiry int64) error {
	m.mu.Lock()
	account, exists := m.config.Accounts[email]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("account %s does not exist", email)
	}

	account.AuthToken = &CachedToken{
		Token:  token,
		Expiry: expiry,
	}
	m.mu.Unlock()

	return m.Save()
}

// GetAccountEmails returns a list of all account emails
func (m *ConfigManager) GetAccountEmails() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	emails := make([]string, 0, len(m.config.Accounts))
	for email := range m.config.Accounts {
		emails = append(emails, email)
	}
	return emails
}

// ParseAuthString parses an auth string and returns url.Values
func ParseAuthString(authString string) (url.Values, error) {
	return url.ParseQuery(authString)
}

// ConfigTokenCache implements gpm.TokenCache for file-based persistence
type ConfigTokenCache struct {
	manager *ConfigManager
	email   string
}

// NewConfigTokenCache creates a new ConfigTokenCache for the given account
func NewConfigTokenCache(manager *ConfigManager, email string) *ConfigTokenCache {
	return &ConfigTokenCache{
		manager: manager,
		email:   email,
	}
}

// Get retrieves the cached token and expiry
func (c *ConfigTokenCache) Get() (string, int64) {
	c.manager.mu.RLock()
	defer c.manager.mu.RUnlock()

	account, exists := c.manager.config.Accounts[c.email]
	if !exists || account.AuthToken == nil {
		return "", 0
	}
	return account.AuthToken.Token, account.AuthToken.Expiry
}

// Set stores the token with its expiry timestamp
func (c *ConfigTokenCache) Set(token string, expiry int64) {
	c.manager.UpdateAccountToken(c.email, token, expiry)
}
