/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/alex-emery/release-notes/pkg/git"
	"github.com/alex-emery/release-notes/pkg/github"
	"github.com/alex-emery/release-notes/pkg/input"
	"github.com/alex-emery/release-notes/pkg/notes"
	jira "github.com/andygrunwald/go-jira/v2/cloud"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func createPrCmd() *cobra.Command {
	var sourceBranch = new(string)
	var targetBranch = new(string)
	var repoPath = new(string)
	var jiraHost = new(string)
	var privateKey = new(string)
	var dryRun = new(bool)
	var verbose = new(bool)

	var prCmd = &cobra.Command{
		Use:   "pr",
		Short: "A brief description of your command",

		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

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

			gitAuth, err := git.New(logger, *privateKey)
			if err != nil {
				logger.Fatal("failed to create git auth", zap.Error(err))
			}

			jiraEmail := os.Getenv("JIRA_EMAIL")
			if jiraEmail == "" {
				logger.Fatal("JIRA_EMAIL not set")
			}

			jiraToken := os.Getenv("JIRA_TOKEN")
			if jiraToken == "" {
				logger.Fatal("JIRA_TOKEN not set")
			}

			ghToken := os.Getenv("GITHUB_TOKEN")
			if ghToken == "" {
				logger.Fatal("GITHUB_TOKEN not set")
			}

			tp := jira.BasicAuthTransport{
				Username: jiraEmail,
				APIToken: jiraToken,
			}

			jiraClient, err := jira.NewClient(*jiraHost, tp.Client())
			if err != nil {
				logger.Fatal("failed to create jira client", zap.Error(err))
			}

			notes, err := notes.CreateReleaseNotes(ctx, logger, gitAuth, jiraClient, *repoPath, *sourceBranch, *targetBranch)
			if err != nil {
				logger.Fatal("failed to create release notes", zap.Error(err))
			}

			ghClient := github.New(logger, ghToken)

			// ask the user to enter a title
			title, err := input.Run("Enter a title for the PR: ")
			if err != nil {
				logger.Fatal("failed to read title", zap.Error(err))
			}

			if title == "" {
				logger.Fatal("title cannot be empty")
			}
			if *dryRun {
				fmt.Println(notes)
				return
			}

			if err = ghClient.CreatePR(ctx, *targetBranch, *sourceBranch, title, notes); err != nil {
				logger.Fatal("failed to create PR", zap.Error(err))
			}
		},
	}

	prCmd.Flags().StringVarP(sourceBranch, "source", "s", "main", "source branch")
	prCmd.Flags().StringVarP(targetBranch, "target", "t", "", "target branch")
	prCmd.Flags().StringVar(repoPath, "path", ".", "path to the local k8s-engine repo")
	prCmd.Flags().StringVar(jiraHost, "jira-host", "https://adarga.atlassian.net", "the host of the jira instance")
	prCmd.Flags().BoolVar(dryRun, "dry-run", false, "enable to not create the PR")
	prCmd.Flags().BoolVar(verbose, "verbose", false, "enable verbose logging")

	prCmd.Flags().StringVar(privateKey, "private-key", "", "path to the private key")

	return prCmd
}
