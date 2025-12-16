package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
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
		if params, err := ParseAuthString(authOverride); err == nil {
			return params.Get("Email")
		}
		return ""
	}
	return cfgManager.GetConfig().Selected
}

// resolveEmailFromArg resolves an email from either an index number (1-based) or email string
func resolveEmailFromArg(arg string, emails []string) (string, error) {
	if idx, err := strconv.Atoi(arg); err == nil {
		if idx < 1 || idx > len(emails) {
			return "", fmt.Errorf("invalid index %d: must be between 1 and %d", idx, len(emails))
		}
		return emails[idx-1], nil
	}

	argLower := strings.ToLower(arg)
	var candidates []string
	for _, email := range emails {
		if email == arg {
			return email, nil
		}
		if strings.Contains(strings.ToLower(email), argLower) {
			candidates = append(candidates, email)
		}
	}

	if len(candidates) == 1 {
		return candidates[0], nil
	} else if len(candidates) > 1 {
		return "", fmt.Errorf("multiple accounts match '%s': %v - please be more specific", arg, candidates)
	}
	return "", fmt.Errorf("no authentication found matching '%s'", arg)
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
