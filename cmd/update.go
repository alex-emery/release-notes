/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/charmbracelet/gum/choose"
	"github.com/spf13/cobra"
)

func createUpdateCmd() *cobra.Command {
	// updateCmd represents the update command
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {
			choose.Options{
				Options: []string{"1", "2", "3"},
				Header:  "Choose a release",
				Height:  10,
			}.Run()
		},
	}

	return updateCmd
}
