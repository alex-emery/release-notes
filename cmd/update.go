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
		Short: "Interactively update images in the k8s engine repo",
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
