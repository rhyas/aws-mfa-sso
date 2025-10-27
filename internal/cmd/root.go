package cmd

import (
	"bufio"
	"io"
	"log"
	"os/exec"
	"regexp"
	"time"

	"github.com/fatih/color"
	"github.com/rhyas/aws-mfa-sso/internal/auth"
	"github.com/rhyas/aws-mfa-sso/internal/aws"
	"github.com/spf13/cobra"
)

var (
	forceLogin bool
	profile    string
)

var rootCmd = &cobra.Command{
	Use:   "aws-mfa-sso",
	Short: "Headless browser AWS SSO login with with MFA",
	Long: `This tool automates AWS SSO login using a headless browser.
It handles MFA authentication without requiring a browser window.`,
	Run: runLogin,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	rootCmd.Flags().BoolVar(&forceLogin, "force", false, "Force login even if valid credentials exist")
	rootCmd.Flags().StringVarP(&profile, "profile", "p", "", "AWS profile to use for SSO login")
}

func runLogin(cmd *cobra.Command, args []string) {
	// Get the SSO start URL for the profile
	startURL := aws.GetProfileStartURL(profile)

	// Check if already logged in with valid credentials (unless --force is specified)
	if !forceLogin && aws.HasValidCredentials(startURL) {
		color.Green("Already logged in with valid AWS SSO credentials")
		color.Yellow("Use --force to re-authenticate")
		return
	}

	// If forcing login, clear existing credentials first
	if forceLogin {
		aws.ClearCache(startURL)
	}

	// Start aws sso login --no-browser as a background process
	var awsCmd *exec.Cmd
	if profile != "" {
		awsCmd = exec.Command("aws", "sso", "login", "--no-browser", "--profile", profile)
	} else {
		awsCmd = exec.Command("aws", "sso", "login", "--no-browser")
	}

	stdout, err := awsCmd.StdoutPipe()
	if err != nil {
		log.Fatal("Error creating stdout pipe:", err)
	}

	if err := awsCmd.Start(); err != nil {
		log.Fatal("Error starting aws sso login:", err)
	}

	// Channel to receive the captured URL
	urlChan := make(chan string, 1)

	// Start goroutine to capture URL from stdout
	go captureURLFromOutput(stdout, urlChan)

	// Wait for URL to be captured
	url := <-urlChan
	if url == "" {
		log.Fatal("Failed to capture SSO URL from aws sso login")
	}
	color.Cyan(url)

	// Perform headless authentication
	auth.PerformLogin(url)
	time.Sleep(1 * time.Second)

	// Wait for the aws sso login process to complete
	if err := awsCmd.Wait(); err != nil {
		log.Printf("aws sso login exited with error: %v\n", err)
	} else {
		log.Println("aws sso login completed successfully")
	}
}

func captureURLFromOutput(reader io.ReadCloser, urlChan chan string) {
	scanner := bufio.NewScanner(reader)
	r := regexp.MustCompile("^https.*user_code=([A-Z]{4}-?){2}")

	for scanner.Scan() {
		line := scanner.Text()
		if r.MatchString(line) {
			urlChan <- line
			return
		}
	}

	// If we never found a URL, send empty string
	urlChan <- ""
}
