/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/alex-emery/release-notes/pkg/git"
	"github.com/alex-emery/release-notes/pkg/notes"
	jira "github.com/andygrunwald/go-jira/v2/cloud"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func createNotesCmd(verbose *bool) *cobra.Command {
	var jiraHost = new(string)
	var privateKey = new(string)
	var notesCmd = &cobra.Command{
		Use:   "notes",
		Short: "Creates release notes for a repo",
		Long: `Creates release notes for a repo, by fetching all commits between the two given tags.
Jira tickets are extracted from the commit messages.
These Jira tickets are then used to provide additional information in the generated notes..`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 3 {
				log.Fatal("not enough arguments: expected repo tag1 tag2")
			}

			ctx := cmd.Context()
			var logger *zap.Logger
			var err error
			if *verbose {
				logger, err = zap.NewDevelopment()
			} else {
				logger, err = zap.NewProduction()
			}

			if err != nil {
				log.Fatal("failed to create logger", err)
			}

			jiraEmail := os.Getenv("JIRA_EMAIL")
			if jiraEmail == "" {
				logger.Fatal("JIRA_EMAIL not set")
			}

			jiraToken := os.Getenv("JIRA_TOKEN")
			if jiraToken == "" {
				logger.Fatal("JIRA_TOKEN not set")
			}

			gitAuth, err := git.New(logger, *privateKey)
			if err != nil {
				logger.Fatal("failed to create git auth", zap.Error(err))
			}

			repoName := args[0]
			tag1 := args[1]
			tag2 := args[2]

			githubToken := os.Getenv("GITHUB_TOKEN")
			if githubToken == "" {
				log.Fatal("GITHUB_TOKEN not set")
			}

			tp := jira.BasicAuthTransport{
				Username: jiraEmail,
				APIToken: jiraToken,
			}

			jiraClient, err := jira.NewClient(*jiraHost, tp.Client())
			if err != nil {
				logger.Fatal("failed to create jira client", zap.Error(err))
			}

			releaseNote := notes.CreateReleaseNotesForRepo(ctx, logger, jiraClient, gitAuth, repoName, tag1, tag2)
			if err != nil {
				logger.Fatal("failed to create release notes", zap.Error(err))
			}

			if releaseNote.Issues == nil {
				logger.Fatal("no issues found")
			}

			releaseNoteString := notes.ReleaseNoteToString(logger, releaseNote)

			fmt.Println(releaseNoteString)
		},
	}

	notesCmd.Flags().StringVar(jiraHost, "jira-host", "https://adarga.atlassian.net", "the host of the jira instance")
	notesCmd.Flags().StringVar(privateKey, "private-key", "", "the path to the private key to use for git authentication")

	return notesCmd
}
