/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/alex-emery/release-notes/internal/wizard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

func createUpdateCmd() *cobra.Command {
	var repoPath = new(string)
	// updateCmd represents the update command
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Interactively update images in the k8s engine repo",
		RunE: func(cmd *cobra.Command, args []string) error {
			m := wizard.NewModel(*repoPath)
			p := tea.NewProgram(m)
			if _, err := p.Run(); err != nil {
				return err
			}
			return nil
		},
	}

	updateCmd.Flags().StringVar(repoPath, "path", ".", "path to the local k8s-engine repo")

	return updateCmd
}
