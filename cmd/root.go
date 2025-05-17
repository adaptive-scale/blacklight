/*
Copyright © 2025 Debarshi Basak
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/adaptive-scale/blacklight/internal/model"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "blacklight",
	Short: "A pluggable secret scanner",
	Long: `Allows users to scan code, files for secrets.
version: 0.1.0
`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.blacklight.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// loadRules loads and compiles rules from the default rules file
func loadRules() ([]model.Configuration, error) {
	// Get rules file path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	rulesFile := filepath.Join(homeDir, ".blacklight", "rules.yaml")
	if _, err := os.Stat(rulesFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("rules file not found at %s", rulesFile)
	}

	// Read rules file
	data, err := os.ReadFile(rulesFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read rules file: %w", err)
	}

	// Parse rules
	var rules []model.Configuration
	if err := yaml.Unmarshal(data, &rules); err != nil {
		return nil, fmt.Errorf("failed to parse rules file: %w", err)
	}

	// Compile regexes
	for i := range rules {
		rules[i].CompiledRegex, err = regexp.Compile(rules[i].Regex)
		if err != nil {
			return nil, fmt.Errorf("failed to compile regex for rule %s: %w", rules[i].Name, err)
		}
	}

	return rules, nil
}

// printViolations prints the violations in a formatted way
func printViolations(violations []model.Violation) {
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	for _, v := range violations {
		// Color based on severity
		var severityColor func(a ...interface{}) string
		switch v.Rule.Severity {
		case 3:
			severityColor = red
		case 2:
			severityColor = yellow
		default:
			severityColor = green
		}

		fmt.Printf("%s: %s\n", severityColor(fmt.Sprintf("[Severity %d]", v.Rule.Severity)), v.Rule.Name)
		fmt.Printf("Location: %s\n", v.Location)
		if v.Context != "" {
			fmt.Printf("Context: %s\n", v.Context)
		}
		fmt.Printf("Match: %s\n", v.Match)
		fmt.Println(strings.Repeat("-", 80))
	}
}
