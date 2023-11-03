package notes

import (
	"context"
	"fmt"
	"log"

	"github.com/alex-emery/release-notes/pkg/git"
	jira "github.com/andygrunwald/go-jira/v2/cloud"
	"go.uber.org/zap"
)

type ReleaseNote struct {
	RepoName string
	Issues   []*jira.Issue
}

func (rn ReleaseNote) String() string {
	result := "### " + rn.RepoName + "\n"
	for _, issue := range rn.Issues {
		line := "- [" + issue.Key + "]" + "(https://adarga.atlassian.net/browse/" + issue.Key + ") - " + issue.Fields.Summary + "\n"
		result += line
	}

	return result
}

func CreateReleaseNotesForRepo(ctx context.Context, logger *zap.Logger, jiraClient *jira.Client, gitAuth *git.Auth, repoName string, tag1 string, tag2 string) ReleaseNote {
	repo, err := gitAuth.CloneRepo(repoName)
	if err != nil {
		logger.Error("failed to clone: skipping", zap.String("repo", repoName), zap.Error(err))
		return ReleaseNote{}
	}

	commits, err := git.GetCommitsBetweenTags(repo, tag1, tag2)
	if err != nil {
		log.Fatal(err)
	}

	foundIssues := make([]*jira.Issue, 0, len(commits))
	uniqueIssues := git.CommitsToIssues(commits)
	for _, issueID := range uniqueIssues {
		fmt.Println("searching for issue ", issueID)
		found, _, err := jiraClient.Issue.Get(ctx, issueID, nil)
		if err != nil {
			log.Println("failed to find issue", issueID, err)
			continue
		}

		foundIssues = append(foundIssues, found)

	}

	return ReleaseNote{
		RepoName: repoName,
		Issues:   foundIssues,
	}
}
