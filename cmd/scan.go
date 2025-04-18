/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"blacklight/internal/model"
	"blacklight/internal/scanner"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

//var secretPatterns = map[string]*regexp.Regexp{
//	"AWS Access Key":  regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
//	"AWS Secret Key":  regexp.MustCompile(`(?i)aws(.{0,20})?['"][0-9a-zA-Z/+]{40}['"]`),
//	"Generic API Key": regexp.MustCompile(`(?i)api[_-]?key(.{0,20})?['"][0-9a-zA-Z]{32,45}['"]`),
//	"Slack Token":     regexp.MustCompile(`xox[baprs]-[0-9a-zA-Z]{10,48}`),
//	"Private Key":     regexp.MustCompile(`-----BEGIN (RSA|DSA|EC|PGP|OPENSSH) PRIVATE KEY-----`),
//	"JWT":             regexp.MustCompile(`eyJ[A-Za-z0-9-_]+?\.[A-Za-z0-9-_]+?\.[A-Za-z0-9-_]+`),
//}

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "takes a directory, file, path and scans for secrets",
	Long: `takes a directory, file, path and scans for secrets

blacklight scan /path/to/dir

`,
	Run: func(cmd *cobra.Command, args []string) {

		ignore, _ := cmd.Flags().GetString("ignore")
		//ignoreSelective, _ := cmd.Flags().GetString("ignore-selective")

		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error getting home directory:", err)
			return
		}

		verbose := os.Getenv("VERBOSE")

		var allconf []model.Configuration

		files, err := os.ReadDir(path.Join(home, ".blacklight"))
		if err != nil {
			fmt.Println("Error reading directory:", err)
			return
		}

		for _, f := range files {

			var conf []model.Configuration

			if f.IsDir() {
				continue
			}

			if filepath.Ext(f.Name()) == ".json" {

				fmt.Println("=> found configuration " + f.Name())
				d, err := os.ReadFile(path.Join(home, ".blacklight", f.Name()))
				if err != nil {
					fmt.Println("Error reading file:", err)
					continue
				}
				if err := json.Unmarshal(d, &conf); err != nil {
					fmt.Println("Error reading configuration in JSON:", err)
					continue
				}

				for _, c := range conf {
					c.RegexVal = regexp.MustCompile(c.Regex)
					allconf = append(allconf, c)
				}
			}
		}

		fmt.Printf("=> Found patterns: %d\n", len(allconf))
		fmt.Println("=> Starting scan...")

		dir := "."
		if len(os.Args) > 1 {
			dir = os.Args[2]
		}

		s := scanner.NewSecretScanner()

		s.AddPattern(allconf...)

		dirs := strings.Split(ignore, ",")

		for _, d := range dirs {
			d = strings.TrimSpace(d)
			if d != "" {
				s.AddIgnoreDir(d)
			}
		}

		s.SetVerbose(verbose)
		//s.AddIgnoreSelective(ignoreSelective)
		s.StartScan(dir)

		fmt.Printf("🔍 Scanning directory: %s\n", dir)
		fmt.Println("✅ Scan complete.")
	},
}

func init() {

	scanCmd.Flags().StringP("ignore", "i", "", "ignore directories and files")
	scanCmd.Flags().StringP("ignore-selective", "s", "", "ignore files and regex combination")

	rootCmd.AddCommand(scanCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// scanCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// scanCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
