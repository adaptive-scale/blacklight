/*
Copyright © 2025 Debarshi Basak <debarshi@adaptive.live>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adaptive-scale/blacklight/internal/model"
	"github.com/adaptive-scale/blacklight/internal/scanner"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// rulesCmd represents the rules command
var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "Manage scanning rules",
	Long: `Manage scanning rules for blacklight.
	
Examples:
  # List all rules
  blacklight rules list

  # Add a new rule
  blacklight rules add --name "Custom API Key" --regex "api_key_[a-zA-Z0-9]{32}" --severity 2 --type "secret"

  # Add rules from a YAML file
  blacklight rules add -f rules.yaml`,
}

// addRuleCmd represents the rule add command
var addRuleCmd = &cobra.Command{
	Use:   "add",
	Short: "Add new scanning rules",
	Long: `Add new scanning rules to blacklight. Rules can be added either through command line arguments
or through a YAML configuration file.

Examples:
  # Add a single rule via CLI
  blacklight rules add --name "Custom API Key" --regex "api_key_[a-zA-Z0-9]{32}" --severity 2 --type "secret"

  # Add multiple rules from a YAML file
  blacklight rules add -f rules.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		file, _ := cmd.Flags().GetString("file")
		name, _ := cmd.Flags().GetString("name")
		regex, _ := cmd.Flags().GetString("regex")
		severity, _ := cmd.Flags().GetInt("severity")
		ruleType, _ := cmd.Flags().GetString("type")

		// Get config directory
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Error getting home directory: %v\n", err)
			return
		}
		configDir := filepath.Join(home, ".blacklight")

		if file != "" {
			// Add rules from file
			if err := addRulesFromFile(file, configDir); err != nil {
				fmt.Printf("Error adding rules from file: %v\n", err)
				return
			}
		} else if name != "" && regex != "" {
			// Add single rule from CLI arguments
			rule := model.Configuration{
				Name:     name,
				Regex:    regex,
				Severity: severity,
				Type:     ruleType,
			}

			if err := addSingleRule(rule, configDir); err != nil {
				fmt.Printf("Error adding rule: %v\n", err)
				return
			}
		} else {
			fmt.Println("Error: either --file or --name and --regex must be provided")
			return
		}
	},
}

// addRulesFromFile adds rules from a YAML file
func addRulesFromFile(filePath, configDir string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	var newRules []model.Configuration
	if err := yaml.Unmarshal(data, &newRules); err != nil {
		return fmt.Errorf("error parsing YAML: %v", err)
	}

	// Load existing rules
	configFile := filepath.Join(configDir, "config.yaml")
	var existingRules []model.Configuration
	if data, err := os.ReadFile(configFile); err == nil {
		yaml.Unmarshal(data, &existingRules)
	}

	// Add new rules
	existingRules = append(existingRules, newRules...)

	// Save updated rules
	data, err = yaml.Marshal(existingRules)
	if err != nil {
		return fmt.Errorf("error marshaling rules: %v", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("error writing rules file: %v", err)
	}

	fmt.Printf("Successfully added %d rules\n", len(newRules))
	return nil
}

// addSingleRule adds a single rule to the config file
func addSingleRule(rule model.Configuration, configDir string) error {
	configFile := filepath.Join(configDir, "config.yaml")

	// Load existing rules
	var rules []model.Configuration
	if data, err := os.ReadFile(configFile); err == nil {
		yaml.Unmarshal(data, &rules)
	}

	// Add new rule
	rules = append(rules, rule)

	// Save updated rules
	data, err := yaml.Marshal(rules)
	if err != nil {
		return fmt.Errorf("error marshaling rules: %v", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("error writing rules file: %v", err)
	}

	fmt.Printf("Successfully added rule: %s\n", rule.Name)
	return nil
}

// listRulesCmd represents the rules list command
var listRulesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured rules",
	Long: `List all configured rules, including both built-in and custom rules.
	
Example:
  blacklight rules list
  blacklight rules list --type secret
  blacklight rules list --severity 3`,
	Run: func(cmd *cobra.Command, args []string) {
		ruleType, _ := cmd.Flags().GetString("type")
		severity, _ := cmd.Flags().GetInt("severity")

		// Get config directory
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Error getting home directory: %v\n", err)
			return
		}
		configDir := filepath.Join(home, ".blacklight")
		configFile := filepath.Join(configDir, "config.yaml")

		// Load user-defined rules
		var rules []model.Configuration
		if data, err := os.ReadFile(configFile); err == nil {
			if err := yaml.Unmarshal(data, &rules); err != nil {
				fmt.Printf("Error parsing config file: %v\n", err)
				return
			}
		}

		// Add built-in rules if no user rules found
		if len(rules) == 0 {
			rules = append(rules, scanner.Regex...)
		}

		// Filter rules based on flags
		var filteredRules []model.Configuration
		for _, rule := range rules {
			if ruleType != "" && rule.Type != ruleType {
				continue
			}
			if severity != 0 && rule.Severity != severity {
				continue
			}
			filteredRules = append(filteredRules, rule)
		}

		// Create a new table writer
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetStyle(table.StyleRounded)

		// Add header
		t.AppendHeader(table.Row{"Name", "Type", "Severity", "Status", "Pattern"})

		// Setup colors
		yellow := color.New(color.FgYellow).SprintFunc()
		red := color.New(color.FgRed).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()

		// Add rows
		for _, rule := range filteredRules {
			severityStr := fmt.Sprint(rule.Severity)
			switch rule.Severity {
			case 1:
				severityStr = green(severityStr)
			case 2:
				severityStr = yellow(severityStr)
			case 3:
				severityStr = red(severityStr)
			}

			status := "Enabled"
			if rule.Disabled {
				status = yellow("Disabled")
			}

			pattern := rule.Regex
			if len(pattern) > 50 {
				pattern = pattern[:47] + "..."
			}

			t.AppendRow(table.Row{
				rule.Name,
				rule.Type,
				severityStr,
				status,
				pattern,
			})
		}

		// Set column configurations
		t.SetColumnConfigs([]table.ColumnConfig{
			{Number: 1, WidthMax: 40},
			{Number: 2, WidthMax: 20},
			{Number: 3, Align: text.AlignCenter},
			{Number: 4, Align: text.AlignCenter},
			{Number: 5, WidthMax: 60},
		})

		// Print summary and render table
		fmt.Printf("\nFound %d rules:\n\n", len(filteredRules))
		t.Render()
	},
}

func init() {
	rootCmd.AddCommand(rulesCmd)

	// Add subcommands
	rulesCmd.AddCommand(addRuleCmd)
	rulesCmd.AddCommand(listRulesCmd)

	// Add flags for the add command
	addRuleCmd.Flags().StringP("file", "f", "", "YAML file containing rules to add")
	addRuleCmd.Flags().StringP("name", "n", "", "Name of the rule")
	addRuleCmd.Flags().String("regex", "", "Regular expression pattern for the rule")
	addRuleCmd.Flags().IntP("severity", "s", 2, "Severity level (1-3, where 1 is highest)")
	addRuleCmd.Flags().StringP("type", "t", "secret", "Type of the rule (e.g., secret, pci, pii)")

	// Add flags for the list command
	listRulesCmd.Flags().StringP("type", "t", "", "Filter rules by type (e.g., secret, pci, pii)")
	listRulesCmd.Flags().IntP("severity", "s", 0, "Filter rules by severity level (1-3)")
}
