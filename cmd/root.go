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
	var verbose = new(bool)

	var rootCmd = &cobra.Command{
		Use:   "release-notes",
		Short: "Creates release notes",
	}

	rootCmd.PersistentFlags().BoolVar(verbose, "verbose", false, "enable verbose logging")

	_ = godotenv.Load()

	rootCmd.AddCommand(createPrCmd(verbose))
	rootCmd.AddCommand(createNotesCmd(verbose))
	rootCmd.AddCommand(createUpdateCmd())
	return rootCmd

}
func Execute() {
	err := createRootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}
