package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/rhyas/aws-mfa-sso/internal/aws"
	"github.com/spf13/cobra"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

var roleArn string

var credentialProcessCmd = &cobra.Command{
	Use:   "credential-process",
	Short: "Output credentials in AWS credential_process format",
	Long: `Assumes a role using SSO credentials and outputs in the format expected by AWS credential_process.
This can be used in ~/.aws/config with the credential_process setting.`,
	Run: runCredentialProcess,
}

func init() {
	credentialProcessCmd.Flags().StringVarP(&roleArn, "role-arn", "r", "", "Role ARN to assume (required)")
	credentialProcessCmd.MarkFlagRequired("role-arn")
	rootCmd.AddCommand(credentialProcessCmd)
}

// CredentialProcessOutput is the JSON format expected by AWS credential_process
type CredentialProcessOutput struct {
	Version         int    `json:"Version"`
	AccessKeyId     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	SessionToken    string `json:"SessionToken"`
	Expiration      string `json:"Expiration"`
}

func runCredentialProcess(cmd *cobra.Command, args []string) {
	// When AWS calls credential_process, AWS_PROFILE is set to the downstream profile
	// We need to read the source_profile to find the SSO profile
	currentProfile := os.Getenv("AWS_PROFILE")
	if currentProfile == "" {
		currentProfile = "default"
	}

	// Get the source_profile (SSO profile) from the current profile
	ssoProfile := aws.GetSourceProfile(currentProfile)
	if ssoProfile == "" {
		// If there's no source_profile, assume the current profile is the SSO profile
		ssoProfile = currentProfile
	}

	// Get SSO credentials from the SSO profile
	startURL := aws.GetProfileStartURL(ssoProfile)

	if !aws.HasValidCredentials(startURL) {
		log.Fatalf("No valid SSO credentials found for profile '%s'. Please run 'aws-mfa-sso --profile %s' first to login.", ssoProfile, ssoProfile)
	}

	ctx := context.Background()

	// Load AWS config using the SSO profile credentials
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigProfile(ssoProfile),
		config.WithRegion("us-east-1"), // Default region if not set
	)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	// Create STS client
	stsClient := sts.NewFromConfig(cfg)

	// Assume the role
	sessionName := getSessionName()
	result, err := stsClient.AssumeRole(ctx, &sts.AssumeRoleInput{
		RoleArn:         &roleArn,
		RoleSessionName: sessionName,
	})
	if err != nil {
		log.Fatalf("Failed to assume role: %v", err)
	}

	// Format output for credential_process
	output := CredentialProcessOutput{
		Version:         1,
		AccessKeyId:     *result.Credentials.AccessKeyId,
		SecretAccessKey: *result.Credentials.SecretAccessKey,
		SessionToken:    *result.Credentials.SessionToken,
		Expiration:      result.Credentials.Expiration.Format(time.RFC3339),
	}

	// Output JSON to stdout
	encoder := json.NewEncoder(os.Stdout)
	if err := encoder.Encode(output); err != nil {
		log.Fatalf("Failed to encode credentials: %v", err)
	}
}

func getSessionName() *string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	sessionName := fmt.Sprintf("aws-mfa-sso-%s-%d", hostname, time.Now().Unix())
	return &sessionName
}
