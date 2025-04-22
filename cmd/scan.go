/*
Copyright © 2025 Debarshi Basak <debarshi@adaptive.live>

*/
package cmd

import (
	"blacklight/internal/model"
	"blacklight/internal/scanner"
	"fmt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
	"path"
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
		sarif, _ := cmd.Flags().GetBool("sarif")

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

			if f.Name() == "config.yaml" {

				fmt.Println("=> found configuration " + f.Name())
				d, err := os.ReadFile(path.Join(home, ".blacklight", f.Name()))
				if err != nil {
					fmt.Println("Error reading file:", err)
					continue
				}
				if err := yaml.Unmarshal(d, &conf); err != nil {
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
		violations := s.StartScan(dir)

		fmt.Printf("🔍 Scanning directory: %s\n", dir)
		fmt.Println("✅ Scan complete.")
		fmt.Printf("=> Found %d violations\n", len(violations))

		if sarif {
			model.GenerateSARIF(violations)
		}
	},
}

func init() {

	scanCmd.Flags().StringP("ignore", "i", "", "ignore directories and files")
	//scanCmd.Flags().StringP("ignore-selective", "s", "", "ignore files and regex combination")
	scanCmd.Flags().BoolP("sarif", "f", false, "generate output in SARIF format")

	rootCmd.AddCommand(scanCmd)
}
