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
	Email         string       `toml:"email"`          // Account email
	Auth          string       `toml:"auth"`           // Auth string (androidId, Token, Email, etc.)
	AuthToken     *CachedToken `toml:"auth_token"`     // Cached access token
	Quality       string       `toml:"quality"`        // "original" or "storage-saver"
	UseQuota      bool         `toml:"use_quota"`      // If true, uploads count against storage quota
	UploadThreads int          `toml:"upload_threads"` // Number of upload threads
	Proxy         string       `toml:"proxy"`          // Proxy URL
}

// Config represents the TOML configuration
type Config struct {
	Selected string           `toml:"selected"` // Selected account email
	Accounts []*AccountConfig `toml:"accounts"` // List of account configs (order preserved)
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
	}

	// Load config from file if it exists
	if data, err := os.ReadFile(configPath); err == nil && len(data) > 0 {
		if err := toml.Unmarshal(data, &m.config); err != nil {
			return nil, fmt.Errorf("failed to parse config: %w", err)
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

// findAccountIndex returns the index of account with given email, or -1 if not found (caller must hold lock)
func (m *ConfigManager) findAccountIndex(email string) int {
	for i, acc := range m.config.Accounts {
		if acc.Email == email {
			return i
		}
	}
	return -1
}

// GetSelectedAccount returns the currently selected account config
func (m *ConfigManager) GetSelectedAccount() *AccountConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if idx := m.findAccountIndex(m.config.Selected); idx >= 0 {
		return m.config.Accounts[idx]
	}
	return nil
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

// SetSelected updates the selected account by email
func (m *ConfigManager) SetSelected(email string) error {
	m.mu.Lock()
	if m.findAccountIndex(email) < 0 {
		m.mu.Unlock()
		return fmt.Errorf("account %s does not exist", email)
	}
	m.config.Selected = email
	m.mu.Unlock()
	return m.Save()
}

// AddCredentials adds a new account with the given auth string
func (m *ConfigManager) AddCredentials(authString string) (string, error) {
	requiredFields := []string{"androidId", "app", "client_sig", "Email", "Token", "lang", "service"}

	params, err := url.ParseQuery(authString)
	if err != nil {
		return "", fmt.Errorf("invalid auth string format: %w", err)
	}

	var missingFields []string
	for _, field := range requiredFields {
		if params.Get(field) == "" {
			missingFields = append(missingFields, field)
		}
	}
	if len(missingFields) > 0 {
		return "", fmt.Errorf("auth string missing required fields: %v", missingFields)
	}

	email := params.Get("Email")
	m.mu.Lock()
	if m.findAccountIndex(email) >= 0 {
		m.mu.Unlock()
		return "", fmt.Errorf("account %s already exists", email)
	}

	account := DefaultAccountConfig()
	account.Email = email
	account.Auth = authString
	m.config.Accounts = append(m.config.Accounts, account)
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
	idx := m.findAccountIndex(email)
	if idx < 0 {
		m.mu.Unlock()
		return fmt.Errorf("account %s does not exist", email)
	}

	m.config.Accounts = append(m.config.Accounts[:idx], m.config.Accounts[idx+1:]...)

	if m.config.Selected == email {
		m.config.Selected = ""
		if len(m.config.Accounts) > 0 {
			m.config.Selected = m.config.Accounts[0].Email
		}
	}
	m.mu.Unlock()
	return m.Save()
}

// UpdateAccountToken updates the cached auth token for an account
func (m *ConfigManager) UpdateAccountToken(email, token string, expiry int64) error {
	m.mu.Lock()
	idx := m.findAccountIndex(email)
	if idx < 0 {
		m.mu.Unlock()
		return fmt.Errorf("account %s does not exist", email)
	}
	m.config.Accounts[idx].AuthToken = &CachedToken{Token: token, Expiry: expiry}
	m.mu.Unlock()
	return m.Save()
}

// GetAccountEmails returns a list of all account emails (in config order)
func (m *ConfigManager) GetAccountEmails() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	emails := make([]string, len(m.config.Accounts))
	for i, acc := range m.config.Accounts {
		emails[i] = acc.Email
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
	if idx := c.manager.findAccountIndex(c.email); idx >= 0 {
		if t := c.manager.config.Accounts[idx].AuthToken; t != nil {
			return t.Token, t.Expiry
		}
	}
	return "", 0
}

// Set stores the token with its expiry timestamp
func (c *ConfigTokenCache) Set(token string, expiry int64) {
	c.manager.UpdateAccountToken(c.email, token, expiry)
}
