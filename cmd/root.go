/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

func createRootCmd() *cobra.Command {
	var privateKey string
	var rootCmd = &cobra.Command{
		Use:   "release-notes",
		Short: "Creates release notes",
	}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	rootCmd.PersistentFlags().StringVar(&privateKey, "private-key", "", "path to the private key for git")
	if err := rootCmd.MarkPersistentFlagRequired("private-key"); err != nil {
		log.Fatal(err)
	}

	rootCmd.AddCommand(createPrCmd(&privateKey))
	rootCmd.AddCommand(createNotesCmd(&privateKey))
	return rootCmd

}
func Execute() {
	err := createRootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}
