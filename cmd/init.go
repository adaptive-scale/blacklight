/*
Copyright © 2025 Debarshi Basak
*/
package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/adaptive-scale/blacklight/internal/scanner"
	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "generates the initial configuration",
	Long: `
Run:

blacklight init

Should generate all configuration and scanner regexes.
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("initializing configuration")

		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error getting home directory:", err)
			return
		}

		if err := os.MkdirAll(path.Join(home, ".blacklight"), os.FileMode(0755)); err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}

		join := path.Join(home, ".blacklight", "config.yaml")

		d, err := yaml.Marshal(scanner.Regex)
		if err != nil {
			fmt.Println("Error marshalling configuration:", err)
			return
		}

		err = os.WriteFile(join, d, os.FileMode(0755))
		if err != nil {
			fmt.Println("Error reading directory:", err)
			return
		}

		fmt.Println("=> created configuration " + join)

		parser := path.Join(home, ".blacklight", "parser.yaml")

		d1, err := yaml.Marshal(scanner.Parser)
		if err != nil {
			fmt.Println("Error marshalling configuration:", err)
			return
		}

		err = os.WriteFile(parser, d1, os.FileMode(0755))
		if err != nil {
			fmt.Println("Error reading directory:", err)
			return
		}

		fmt.Println("=> created configuration " + parser)

	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
