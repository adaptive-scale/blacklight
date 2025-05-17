package cmd

import (
	"fmt"
	"time"

	"github.com/adaptive-scale/blacklight/internal/scanner"
	"github.com/spf13/cobra"
)

var slackCmd = &cobra.Command{
	Use:   "slack",
	Short: "Scan Slack messages and files for secrets",
	Long: `Scan Slack workspace messages and files for secrets and sensitive information.
	
This command requires a Slack Bot User OAuth Token with the following scopes:
- channels:history
- channels:read
- files:read
- groups:history
- groups:read
- im:history
- im:read
- mpim:history
- mpim:read

Example:
  # Scan all accessible channels
  blacklight slack --token xoxb-your-token

  # Scan specific channels
  blacklight slack --token xoxb-your-token --channels C01234567,C89012345

  # Scan last 7 days of messages
  blacklight slack --token xoxb-your-token --days 7

  # Include message threads and files
  blacklight slack --token xoxb-your-token --include-threads --include-files`,
	RunE: runSlackCmd,
}

var slackConfig = &scanner.SlackConfig{}

func init() {
	rootCmd.AddCommand(slackCmd)

	slackCmd.Flags().StringVar(&slackConfig.Token, "token", "", "Slack Bot User OAuth Token (required)")
	slackCmd.Flags().StringSliceVar(&slackConfig.Channels, "channels", nil, "Comma-separated list of channel IDs to scan (optional, defaults to all accessible channels)")
	slackCmd.Flags().IntVar(&slackConfig.DaysToScan, "days", 30, "Number of days of history to scan")
	slackCmd.Flags().BoolVar(&slackConfig.IncludeThreads, "include-threads", false, "Whether to scan message threads")
	slackCmd.Flags().BoolVar(&slackConfig.IncludeFiles, "include-files", false, "Whether to scan file contents")
	slackCmd.Flags().BoolVar(&slackConfig.ExcludeArchived, "exclude-archived", true, "Whether to exclude archived channels")

	slackCmd.MarkFlagRequired("token")
}

func runSlackCmd(cmd *cobra.Command, args []string) error {
	// Create Slack scanner
	slackScanner, err := scanner.NewSlackScanner(slackConfig)
	if err != nil {
		return fmt.Errorf("failed to create slack scanner: %w", err)
	}

	// Load rules
	rules, err := loadRules()
	if err != nil {
		return fmt.Errorf("failed to load rules: %w", err)
	}

	// Start scanning
	fmt.Printf("Starting Slack scan...\n")
	startTime := time.Now()

	violations, err := slackScanner.Scan(cmd.Context(), rules)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	// Print results
	duration := time.Since(startTime)
	fmt.Printf("\nScan completed in %s\n", duration)
	fmt.Printf("Found %d violations\n\n", len(violations))

	if len(violations) > 0 {
		printViolations(violations)
	}

	return nil
} 