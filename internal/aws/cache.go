package aws

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
)

// CacheEntry represents the structure of AWS SSO cache files
type CacheEntry struct {
	StartURL              string    `json:"startUrl,omitempty"`
	Region                string    `json:"region,omitempty"`
	AccessToken           string    `json:"accessToken,omitempty"`
	ExpiresAt             time.Time `json:"expiresAt"`
	ClientID              string    `json:"clientId,omitempty"`
	ClientSecret          string    `json:"clientSecret,omitempty"`
	RegistrationExpiresAt time.Time `json:"registrationExpiresAt,omitempty"`
}

// HasValidCredentials checks if there are valid (non-expired) credentials in the SSO cache
func HasValidCredentials(startURL string) bool {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	cachePath := filepath.Join(dirname, ".aws", "sso", "cache")

	// Check if cache directory exists
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return false
	}

	// Read all files in the cache directory
	files, err := filepath.Glob(filepath.Join(cachePath, "*.json"))
	if err != nil || len(files) == 0 {
		return false
	}

	now := time.Now()

	// Check each cache file for valid credentials
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var cacheEntry CacheEntry
		if err := json.Unmarshal(data, &cacheEntry); err != nil {
			continue
		}

		// If startURL is specified, only check matching cache entries
		if startURL != "" && cacheEntry.StartURL != startURL {
			continue
		}

		// Check if this entry has an access token and hasn't expired
		if cacheEntry.AccessToken != "" && cacheEntry.ExpiresAt.After(now) {
			return true
		}
	}

	return false
}

// ClearCache removes cached SSO credentials for the specified startURL
func ClearCache(startURL string) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Warning: Could not get home directory: %v\n", err)
		return
	}

	cachePath := filepath.Join(dirname, ".aws", "sso", "cache")

	// Check if cache directory exists
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return // Nothing to clear
	}

	// Read all files in the cache directory
	files, err := filepath.Glob(filepath.Join(cachePath, "*.json"))
	if err != nil {
		log.Printf("Warning: Could not read cache directory: %v\n", err)
		return
	}

	clearedCount := 0
	// Remove cache files that match the startURL (or all if no startURL specified)
	for _, file := range files {
		shouldRemove := false

		if startURL == "" {
			// If no startURL specified, remove all cache files
			shouldRemove = true
		} else {
			// Check if this cache file matches the startURL
			data, err := os.ReadFile(file)
			if err != nil {
				continue
			}

			var cacheEntry CacheEntry
			if err := json.Unmarshal(data, &cacheEntry); err != nil {
				continue
			}

			if cacheEntry.StartURL == startURL {
				shouldRemove = true
			}
		}

		if shouldRemove {
			if err := os.Remove(file); err != nil {
				log.Printf("Warning: Could not remove cache file %s: %v\n", file, err)
			} else {
				clearedCount++
			}
		}
	}

	// Also remove the browser cookie file
	cookieFile := filepath.Join(dirname, ".aws-mfa-sso")
	if _, err := os.Stat(cookieFile); err == nil {
		if err := os.Remove(cookieFile); err != nil {
			log.Printf("Warning: Could not remove browser cookie file: %v\n", err)
		} else {
			log.Println("Cleared browser cookies")
		}
	}

	if clearedCount > 0 {
		color.Yellow("Cleared existing SSO credentials")
	}
}
