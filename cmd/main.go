// Package main provides the CLI entry point for sonarqube-cli.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "sonarqube-cli",
	Short: "SonarQube resource management CLI",
}

func init() {
	rootCmd.AddCommand(resourcesCmd, projectCmd, userCmd, groupCmd, tokenCmd)
}
