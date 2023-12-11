/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/alex-emery/release-notes/internal/model/wizard"
	"github.com/spf13/cobra"
)

func createUpdateCmd() *cobra.Command {
	var repoPath = new(string)
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
			app := wizard.New(*repoPath)

			if err := app.Run(); err != nil {
				panic(err)
			}
		},
	}

	updateCmd.Flags().StringVar(repoPath, "path", ".", "path to the local k8s-engine repo")

	return updateCmd
}
