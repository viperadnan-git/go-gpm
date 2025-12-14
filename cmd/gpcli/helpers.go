package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	gpm "github.com/viperadnan-git/go-gpm"
)

var configPath string
var authOverride string
var cfgManager *ConfigManager

func loadConfig() error {
	var err error
	cfgManager, err = NewConfigManager(configPath)
	return err
}

// createAPIClient creates a new Google Photos API client with token caching
func createAPIClient() (*gpm.GooglePhotosAPI, error) {
	authData := getAuthData()
	if authData == "" {
		return nil, fmt.Errorf("no authentication configured. Use 'gpcli auth add' to add credentials")
	}

	email := getSelectedEmail()
	account := cfgManager.GetSelectedAccount()

	var proxy string
	if account != nil {
		proxy = account.Proxy
	}

	// Create token cache for persistent token storage
	var tokenCache gpm.TokenCache
	if email != "" && authOverride == "" {
		tokenCache = NewConfigTokenCache(cfgManager, email)
	}

	return gpm.NewGooglePhotosAPI(gpm.ApiConfig{
		AuthData:   authData,
		Proxy:      proxy,
		TokenCache: tokenCache,
	})
}

// getAuthData returns the auth data string based on authOverride or selected config
func getAuthData() string {
	if authOverride != "" {
		return authOverride
	}
	account := cfgManager.GetSelectedAccount()
	if account != nil {
		return account.Auth
	}
	return ""
}

// getSelectedEmail returns the email of the currently selected account
func getSelectedEmail() string {
	if authOverride != "" {
		params, err := ParseAuthString(authOverride)
		if err == nil {
			return params.Get("Email")
		}
		return ""
	}
	return cfgManager.GetConfig().Selected
}

// resolveEmailFromArg resolves an email from either an index number (1-based) or email string
func resolveEmailFromArg(arg string, emails []string) (string, error) {
	// Try to parse as number first
	if num, err := fmt.Sscanf(arg, "%d", new(int)); err == nil && num == 1 {
		var idx int
		fmt.Sscanf(arg, "%d", &idx)
		if idx < 1 || idx > len(emails) {
			return "", fmt.Errorf("invalid index %d: must be between 1 and %d", idx, len(emails))
		}
		return emails[idx-1], nil
	}

	// Otherwise treat as email - try exact match first
	for _, email := range emails {
		if email == arg {
			return email, nil
		}
	}

	// Try fuzzy matching
	var candidates []string
	for _, email := range emails {
		if containsSubstring(email, arg) {
			candidates = append(candidates, email)
		}
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("no authentication found matching '%s'", arg)
	} else if len(candidates) == 1 {
		return candidates[0], nil
	}
	return "", fmt.Errorf("multiple accounts match '%s': %v - please be more specific", arg, candidates)
}

func containsSubstring(str, substr string) bool {
	strLower := strings.ToLower(str)
	substrLower := strings.ToLower(substr)
	return strings.Contains(strLower, substrLower)
}

// readLinesFromFile reads lines from a file, one per line, skipping empty lines and comments
func readLinesFromFile(filepath string) ([]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return lines, nil
}
