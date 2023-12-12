package notes

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/alex-emery/release-notes/pkg/git"
	jira "github.com/andygrunwald/go-jira/v2/cloud"
	"go.uber.org/zap"
)

func CreateReleaseNotesFromK8sEngine(ctx context.Context, logger *zap.Logger, gitAuth *git.Auth, jiraClient *jira.Client, repoPath string, sourceBranch string, targetBranch *string) (string, error) {
	logger.Info("getting k8s-engine repo")
	repo, err := gitAuth.GetK8sEngineRepo(repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to get k8s-engine repo: %w", err)
	}

	logger.Info("k8s-engine repo opened")

	originalBranch, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	defer func() {
		if err = git.Checkout(repo, originalBranch); err != nil {
			logger.Error("failed to restore to original branch", zap.Error(err))
		}
	}()

	if *targetBranch == "" {
		logger.Debug("target branch not set, getting head ref")

		*targetBranch = originalBranch.Name().Short()
		logger.Info("defaulting target branch", zap.String("target branch", *targetBranch))
	}

	sourceRefs := fmt.Sprintf("refs/heads/%s", sourceBranch)
	targetRefs := fmt.Sprintf("refs/heads/%s", *targetBranch)

	logger.Info("fetching image tags from k8s-engine")
	diffs, err := git.GetImagesFromK8s(repo, sourceRefs, targetRefs)
	if err != nil {
		return "", fmt.Errorf("failed to get images from k8s: %w", err)
	}

	logger.Info("creating release notes")

	resultChan := make(chan ReleaseNote, len(diffs))
	wg := sync.WaitGroup{}
	for _, diff := range diffs {
		wg.Add(1)
		go func(diff git.ImageDiff) {
			defer func() {
				wg.Done()
			}()

			logger.Debug("diff", zap.String("name", diff.Name), zap.String("tag1", diff.Tag1), zap.String("tag2", diff.Tag2))
			if repoName := git.ExtractRepoName(diff.Name); repoName != "" {
				resultChan <- CreateReleaseNotesForRepo(ctx, logger, jiraClient, gitAuth, repoName, diff.Tag1, diff.Tag2)
			} else {
				resultChan <- ReleaseNote{}
			}
		}(diff)
	}

	wg.Wait()
	close(resultChan)

	results := []ReleaseNote{}
	for res := range resultChan {
		if res.RepoName == "" {
			continue
		}
		results = append(results, res)
	}

	return WrapReleaseWithEnvTemplate(ReleaseNoteToString(logger, results...))
}

func ReleaseNoteToString(logger *zap.Logger, notes ...ReleaseNote) string {
	body := strings.Builder{}
	body.Write([]byte("## Release Notes\n\n"))
	for _, note := range notes {
		resString, err := note.String()
		if err != nil {
			logger.Error("failed to get release note for repo", zap.String("repo name", note.RepoName), zap.Error(err))
			continue
		}
		body.WriteString(resString + "\n")
	}

	return body.String()
}
