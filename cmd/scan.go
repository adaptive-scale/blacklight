/*
Copyright © 2025 Debarshi Basak
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/adaptive-scale/blacklight/internal/model"
	"github.com/adaptive-scale/blacklight/internal/scanner"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// printViolation prints a single violation with detailed information
func printViolation(index int, v model.Violation, verbose bool) {
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	
	if verbose {
		fmt.Printf("\n%s #%d\n", yellow("Violation"), index+1)
		fmt.Printf("Rule:        %s\n", cyan(v.Rule.Name))
		fmt.Printf("Severity:    %s\n", getSeverityText(v.Rule.Severity))
		fmt.Printf("Location:    %s", cyan(v.Location))
		if v.LineNum > 0 {
			fmt.Printf(":%d", v.LineNum)
		}
		fmt.Println()
		if v.Rule.Description != "" {
			fmt.Printf("Description: %s\n", v.Rule.Description)
		}
		if v.Match != "" {
			fmt.Printf("Found:       %s\n", red(v.Match))
		}
		if v.Context != "" {
			fmt.Printf("Context:     %s\n", v.Context)
		}
		fmt.Println(strings.Repeat("-", 80))
	} else {
		fmt.Printf("%s: %s in %s", yellow("Found"), red(v.Rule.Name), cyan(v.Location))
		if v.LineNum > 0 {
			fmt.Printf(":%d", v.LineNum)
		}
		fmt.Println()
	}
}

// getSeverityText returns a colored text representation of the severity level
func getSeverityText(level int) string {
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	switch level {
	case 3:
		return red("High")
	case 2:
		return yellow("Medium")
	case 1:
		return green("Low")
	default:
		return fmt.Sprintf("Unknown (%d)", level)
	}
}

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "takes a directory, file, path, database URL or S3 bucket and scans for secrets",
	Long: `takes a directory, file, path, database URL or S3 bucket and scans for secrets

Examples:
  blacklight scan /path/to/dir                     # scan a directory
  blacklight scan --file C:\path\to\file.txt       # scan a single file (Windows)
  blacklight scan --file /path/to/file.txt         # scan a single file (Unix)
  blacklight scan --db "postgres://user:pass@localhost:5432/dbname"
  blacklight scan --db "mysql://user:pass@localhost:3306/dbname" --sample-size 1000
  blacklight scan --s3 "s3://my-bucket/prefix"     # scan an S3 bucket with optional prefix
`,
	Run: func(cmd *cobra.Command, args []string) {
		dbURL, _ := cmd.Flags().GetString("db")
		s3URL, _ := cmd.Flags().GetString("s3")
		cloudURL, _ := cmd.Flags().GetString("drive")
		filePath, _ := cmd.Flags().GetString("file")
		sampleSize, _ := cmd.Flags().GetInt("sample-size")
		ignore, _ := cmd.Flags().GetString("ignore")
		sarif, _ := cmd.Flags().GetBool("sarif")
		verbose, _ := cmd.Flags().GetBool("verbose")
		includeShared, _ := cmd.Flags().GetBool("include-shared")
		days, _ := cmd.Flags().GetInt("days")
		maxSize, _ := cmd.Flags().GetInt64("max-size")

		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error getting home directory:", err)
			return
		}

		scanVerbose := os.Getenv("VERBOSE")

		var allconf []model.Configuration

		// Use filepath.Join for Windows compatibility
		configDir := filepath.Join(home, ".blacklight")
		files, err := os.ReadDir(configDir)
		if err != nil {
			fmt.Println("Error reading directory:", err)
			return
		}

		for _, f := range files {
			var conf []model.Configuration

			if f.IsDir() {
				continue
			}

			if f.Name() == "config.yaml" {
				fmt.Println("=> found configuration " + f.Name())
				configPath := filepath.Join(configDir, f.Name())
				d, err := os.ReadFile(configPath)
				if err != nil {
					fmt.Println("Error reading file:", err)
					continue
				}
				if err := yaml.Unmarshal(d, &conf); err != nil {
					fmt.Println("Error reading configuration in JSON:", err)
					continue
				}

				for _, c := range conf {
					var err error
					c.CompiledRegex, err = regexp.Compile(c.Regex)
					if err != nil {
						fmt.Printf("Error compiling regex for rule %s: %v\n", c.Name, err)
						continue
					}
					allconf = append(allconf, c)
				}
			}
		}

		fmt.Printf("=> Found patterns: %d\n", len(allconf))
		fmt.Println("=> Starting scan...")

		var violations []model.Violation

		switch {
		case dbURL != "":
			// Database scanning mode
			fmt.Printf("🔍 Scanning database: %s\n", dbURL)
			dbScanner := scanner.NewDBScanner(sampleSize)
			dbScanner.AddPattern(allconf...)
			
			dbViolations, err := dbScanner.ScanDatabase(dbURL)
			if err != nil {
				fmt.Printf("Error scanning database: %v\n", err)
				os.Exit(1)
			}
			violations = dbViolations

		case s3URL != "":
			// S3 bucket scanning mode
			fmt.Printf("🔍 Scanning S3 bucket: %s\n", s3URL)
			s3Scanner, err := scanner.NewS3Scanner()
			if err != nil {
				fmt.Printf("Error initializing S3 scanner: %v\n", err)
				os.Exit(1)
			}
			s3Scanner.AddPattern(allconf...)
			
			s3Violations, err := s3Scanner.ScanBucket(s3URL)
			if err != nil {
				fmt.Printf("Error scanning S3 bucket: %v\n", err)
				os.Exit(1)
			}
			violations = s3Violations

		case cloudURL != "":
			// Cloud storage scanning mode
			fmt.Printf("🔍 Scanning cloud storage: %s\n", cloudURL)
			cloudConfig := &scanner.CloudConfig{
				Token:         os.Getenv("CLOUD_TOKEN"),
				DaysToScan:    days,
				IncludeShared: includeShared,
				MaxFileSize:   maxSize,
			}
			cloudScanner, err := scanner.NewCloudScanner(cloudConfig)
			if err != nil {
				fmt.Printf("Error initializing cloud scanner: %v\n", err)
				os.Exit(1)
			}
			cloudScanner.AddPattern(allconf...)

			cloudViolations, err := cloudScanner.ScanStorage(context.Background(), cloudURL)
			if err != nil {
				fmt.Printf("Error scanning cloud storage: %v\n", err)
				os.Exit(1)
			}
			violations = cloudViolations

		default:
			// File system scanning mode
			s := scanner.NewSecretScanner()
			s.AddPattern(allconf...)
			s.SetVerbose(scanVerbose)

			// Handle ignore paths in a cross-platform way
			dirs := strings.Split(ignore, ",")
			for _, d := range dirs {
				d = strings.TrimSpace(d)
				if d != "" {
					// Clean the path to use proper separators
					d = filepath.Clean(d)
					s.AddIgnoreDir(d)
				}
			}

			if filePath != "" {
				// Single file scanning mode
				// Clean the file path to use proper separators
				filePath = filepath.Clean(filePath)
				fmt.Printf("🔍 Scanning file: %s\n", filePath)
				fileInfo, err := os.Stat(filePath)
				if err != nil {
					fmt.Printf("Error accessing file: %v\n", err)
					os.Exit(1)
				}
				if fileInfo.IsDir() {
					fmt.Printf("Error: %s is a directory. Use directory scan mode instead.\n", filePath)
					os.Exit(1)
				}
				violations = s.StartScan(filePath)
			} else {
				// Directory scanning mode
				dir := "."
				if len(args) > 0 {
					// Clean the directory path to use proper separators
					dir = filepath.Clean(args[0])
				}
				fmt.Printf("🔍 Scanning directory: %s\n", dir)
				violations = s.StartScan(dir)
			}
		}

		fmt.Println("\n✅ Scan complete.")

		// Print violations with details based on verbose mode
		if len(violations) > 0 {
			yellow := color.New(color.FgYellow).SprintFunc()
			fmt.Printf("\n%s Found %d potential secret(s):\n", yellow("⚠️"), len(violations))
			
			for i, v := range violations {
				printViolation(i, v, verbose)
			}
		} else {
			green := color.New(color.FgGreen).SprintFunc()
			fmt.Printf("\n%s No secrets found\n", green("✓"))
		}

		if sarif {
			model.GenerateSARIF(violations)
		}
	},
}

func init() {
	scanCmd.Flags().StringP("ignore", "i", "", "ignore directories and files (comma-separated)")
	scanCmd.Flags().StringP("db", "d", "", "database URL to scan (postgres:// or mysql://)")
	scanCmd.Flags().StringP("file", "f", "", "single file to scan (use proper path format for your OS)")
	scanCmd.Flags().StringP("s3", "s", "", "S3 bucket URL to scan (s3://bucket/prefix)")
	scanCmd.Flags().StringP("drive", "r", "", "cloud storage URL to scan (gdrive://, onedrive://, dropbox://, box://)")
	scanCmd.Flags().IntP("sample-size", "n", 100, "number of records to sample from each column (default 100)")
	scanCmd.Flags().Bool("sarif", false, "generate output in SARIF format")
	scanCmd.Flags().BoolP("verbose", "v", false, "show detailed information about violations")
	scanCmd.Flags().Bool("include-shared", false, "include shared files in cloud storage scan")
	scanCmd.Flags().Int("days", 30, "number of days of history to scan for cloud storage")
	scanCmd.Flags().Int64("max-size", 10*1024*1024, "maximum file size to scan in bytes (default 10MB)")

	rootCmd.AddCommand(scanCmd)
}
