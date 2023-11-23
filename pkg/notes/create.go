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

func CreateReleaseNotes(ctx context.Context, logger *zap.Logger, gitAuth *git.Auth, jiraClient *jira.Client, repoPath string, sourceBranch, targetBranch string) string {
	logger.Info("getting k8s-engine repo")
	repo, err := gitAuth.GetK8sEngineRepo(repoPath)
	if err != nil {
		logger.Fatal("failed to clone repo", zap.Error(err))
	}
	logger.Info("k8s-engine repo opened")

	sourceRefs := fmt.Sprintf("refs/heads/%s", sourceBranch)
	targetRefs := fmt.Sprintf("refs/heads/%s", targetBranch)

	logger.Info("fetching image tags from k8s-engine")
	diffs, err := git.GetImagesFromK8s(repo, sourceRefs, targetRefs)
	if err != nil {
		logger.Fatal("failed to get images from k8s", zap.Error(err))
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
			if repoName := git.ExtractRepo(diff.Name); repoName != "" {
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

	body := strings.Builder{}
	body.Write([]byte("## Release Notes\n\n"))
	for _, res := range results {
		body.WriteString(res.String() + "\n")
	}

	return body.String()
}
