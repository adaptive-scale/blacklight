/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// rulesCmd represents the rules command
var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("rules called")
	},
}

func init() {

	rulesCmd.Flags().StringP("list", "l", "", "List all rules")
	rulesCmd.Flags().StringP("add", "a", "", "add rules")
	rulesCmd.Flags().StringP("remove", "r", "", "remove rules")
	rulesCmd.Flags().StringP("add-parser", "p", "", "add parser")

	rootCmd.AddCommand(rulesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// rulesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// rulesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
