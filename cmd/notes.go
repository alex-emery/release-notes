/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/alex-emery/release-notes/pkg/git"
	"github.com/google/go-github/v56/github"
	"github.com/spf13/cobra"
)

func createNotesCmd(privatekey *string) *cobra.Command {

	var jiraHost string
	var notesCmd = &cobra.Command{
		Use:   "notes",
		Short: "creates release notes for a repo",
		Long:  "creates release notes for a repo, fetching additional fields from Jira.",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 3 {
				log.Fatal("not enough arguments: expected repo tag1 tag2")
			}

			gitAuth := git.New(*privatekey)
			repoName := args[0]
			tag1 := args[1]
			tag2 := args[2]

			githubToken := os.Getenv("GITHUB_TOKEN")
			if githubToken == "" {
				log.Fatal("GITHUB_TOKEN not set")
			}

			ghClient := github.NewClient(nil).WithAuthToken(githubToken)

			repo, err := gitAuth.CloneRepo(repoName)
			if err != nil {
				log.Fatal(err)
			}

			tags, err := git.GetTagsBetweenTags(repo, tag1, tag2)
			if err != nil {
				log.Fatal(err)
			}

			for _, tag := range tags {
				fmt.Println(git.GetReleaseForTags(ghClient, repoName, tag))
			}

		},
	}

	notesCmd.Flags().StringVar(&jiraHost, "jira-host", "https://adarga.atlassian.net", "the host of the jira instance")

	return notesCmd
}
