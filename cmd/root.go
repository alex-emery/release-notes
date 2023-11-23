/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

func createRootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "release-notes",
		Short: "Creates release notes",
	}

	_ = godotenv.Load()

	rootCmd.AddCommand(createPrCmd())
	rootCmd.AddCommand(createNotesCmd())
	rootCmd.AddCommand(createUpdateCmd())
	return rootCmd

}
func Execute() {
	err := createRootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}
