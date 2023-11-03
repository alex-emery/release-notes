/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/alex-emery/release-notes/pkg/git"
	"github.com/alex-emery/release-notes/pkg/notes"
	jira "github.com/andygrunwald/go-jira/v2/cloud"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func createPrCmd(privatekey *string) *cobra.Command {
	var sourceBranch = new(string)
	var targetBranch = new(string)
	var jiraHost = new(string)

	var prCmd = &cobra.Command{
		Use:   "pr",
		Short: "A brief description of your command",

		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			logger, err := zap.NewDevelopment()
			if err != nil {
				log.Fatal("failed to create logger", err)
			}

			gitAuth := git.New(*privatekey)

			jiraEmail := os.Getenv("JIRA_EMAIL")
			jiraToken := os.Getenv("JIRA_TOKEN")

			tp := jira.BasicAuthTransport{
				Username: jiraEmail,
				APIToken: jiraToken,
			}

			jiraClient, err := jira.NewClient(*jiraHost, tp.Client())
			if err != nil {
				logger.Fatal("failed to create jira client", zap.Error(err))
			}

			repo, err := gitAuth.CloneRepo("k8s-engine")
			if err != nil {
				logger.Fatal("failed to clone repo", zap.Error(err))
			}

			diffs, err := git.GetImagesFromK8s(repo, *sourceBranch, *targetBranch)
			if err != nil {
				logger.Fatal("failed to get images from k8s", zap.Error(err))
			}

			resultChan := make(chan notes.ReleaseNote, len(diffs))

			wg := sync.WaitGroup{}
			for _, diff := range diffs {
				wg.Add(1)
				go func(diff git.ImageDiff) {
					defer func() {
						wg.Done()
					}()

					logger.Info("diff", zap.String("name", diff.Name), zap.String("tag1", diff.Tag1), zap.String("tag2", diff.Tag2))
					if repoName := git.ExtractRepo(diff.Name); repoName != "" {
						resultChan <- notes.CreateReleaseNotesForRepo(ctx, logger, jiraClient, gitAuth, repoName, diff.Tag1, diff.Tag2)
					} else {
						resultChan <- notes.ReleaseNote{}
					}
				}(diff)
			}

			wg.Wait()
			close(resultChan)

			results := []notes.ReleaseNote{}
			for res := range resultChan {
				if res.RepoName == "" {
					continue
				}
				results = append(results, res)
			}

			file, err := os.Create("release-notes.md")
			if err != nil {
				logger.Fatal("failed to create file", zap.Error(err))
			}

			defer file.Close()

			for _, res := range results {
				_, err := file.WriteString(res.String() + "\n")
				if err != nil {
					logger.Fatal("failed to write to file", zap.Error(err))
				}
			}
		},
	}

	prCmd.Flags().StringVarP(sourceBranch, "source", "s", "", "source branch")
	prCmd.Flags().StringVarP(targetBranch, "target", "t", "", "target branch")

	prCmd.Flags().StringVar(jiraHost, "jira-host", "https://adarga.atlassian.net", "the host of the jira instance")
	if err := prCmd.MarkFlagRequired("source"); err != nil {
		log.Fatal(err)
	}

	if err := prCmd.MarkFlagRequired("target"); err != nil {
		log.Fatal(err)
	}

	return prCmd
}
