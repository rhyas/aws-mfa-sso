package aws

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
)

// GetProfileStartURL reads the AWS config to get the SSO start URL for a profile
func GetProfileStartURL(profile string) string {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	configPath := filepath.Join(dirname, ".aws", "config")
	file, err := os.Open(configPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	var currentProfile string
	var startURL string
	scanner := bufio.NewScanner(file)

	// Determine the profile section name to look for
	profileSection := "[default]"
	if profile != "" {
		profileSection = "[profile " + profile + "]"
	}

	for scanner.Scan() {
		line := scanner.Text()
		line = regexp.MustCompile(`^\s+|\s+$`).ReplaceAllString(line, "") // Trim whitespace

		// Check if this is a profile section
		if regexp.MustCompile(`^\[.*\]$`).MatchString(line) {
			currentProfile = line
			continue
		}

		// If we're in the target profile, look for sso_start_url
		if currentProfile == profileSection {
			if regexp.MustCompile(`^sso_start_url\s*=`).MatchString(line) {
				parts := regexp.MustCompile(`\s*=\s*`).Split(line, 2)
				if len(parts) == 2 {
					startURL = parts[1]
					break
				}
			}
		}
	}

	return startURL
}

// GetSourceProfile reads the AWS config to get the source_profile for a profile
func GetSourceProfile(profile string) string {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	configPath := filepath.Join(dirname, ".aws", "config")
	file, err := os.Open(configPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	var currentProfile string
	var sourceProfile string
	scanner := bufio.NewScanner(file)

	// Determine the profile section name to look for
	profileSection := "[default]"
	if profile != "" {
		profileSection = "[profile " + profile + "]"
	}

	for scanner.Scan() {
		line := scanner.Text()
		line = regexp.MustCompile(`^\s+|\s+$`).ReplaceAllString(line, "") // Trim whitespace

		// Check if this is a profile section
		if regexp.MustCompile(`^\[.*\]$`).MatchString(line) {
			currentProfile = line
			continue
		}

		// If we're in the target profile, look for source_profile
		if currentProfile == profileSection {
			if regexp.MustCompile(`^source_profile\s*=`).MatchString(line) {
				parts := regexp.MustCompile(`\s*=\s*`).Split(line, 2)
				if len(parts) == 2 {
					sourceProfile = parts[1]
					break
				}
			}
		}
	}

	return sourceProfile
}
