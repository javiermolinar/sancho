package llm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// LoadGitHubToken loads the GitHub OAuth token from standard locations.
// It checks in order:
// 1. GITHUB_TOKEN environment variable
// 2. ~/.config/github-copilot/hosts.json
// 3. ~/.config/github-copilot/apps.json
func LoadGitHubToken() (string, error) {
	// First check environment variable
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token, nil
	}

	// Get config directory based on OS
	configDir, err := getConfigDir()
	if err != nil {
		return "", fmt.Errorf("getting config directory: %w", err)
	}

	// Try both hosts.json and apps.json files
	filePaths := []string{
		filepath.Join(configDir, "github-copilot", "hosts.json"),
		filepath.Join(configDir, "github-copilot", "apps.json"),
	}

	for _, filePath := range filePaths {
		token, err := loadTokenFromFile(filePath)
		if err == nil && token != "" {
			return token, nil
		}
	}

	return "", fmt.Errorf("GitHub token not found: set GITHUB_TOKEN or authenticate with GitHub Copilot in your IDE")
}

// getConfigDir returns the user's config directory based on OS.
func getConfigDir() (string, error) {
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return xdgConfig, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if runtime.GOOS == "windows" {
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			return localAppData, nil
		}
		return filepath.Join(home, "AppData", "Local"), nil
	}

	return filepath.Join(home, ".config"), nil
}

// loadTokenFromFile reads a GitHub Copilot config file and extracts the oauth_token.
func loadTokenFromFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	var config map[string]map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		return "", err
	}

	for key, value := range config {
		if strings.Contains(key, "github.com") {
			if oauthToken, ok := value["oauth_token"].(string); ok {
				return oauthToken, nil
			}
		}
	}

	return "", fmt.Errorf("oauth_token not found in %s", filePath)
}
