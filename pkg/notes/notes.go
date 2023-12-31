package notes

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"

	"github.com/alex-emery/release-notes/pkg/git"
	jira "github.com/andygrunwald/go-jira/v2/cloud"
	"github.com/go-git/go-git/v5/plumbing/object"
	"go.uber.org/zap"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type IssueCommitMap map[*jira.Issue][]object.Commit
type ReleaseNote struct {
	RepoName string
	Issues   IssueCommitMap
}

// all the fields for printing the template.
type PRTemplate struct {
	RepoName string
	RepoURL  string
	Issues   []IssueTemplate
}

type IssueTemplate struct {
	ID      string
	Summary string
	Status  string
	Labels  []string
	PRs     []string
}

// Just wraps the relase with the env template.
func WrapReleaseWithEnvTemplate(content string) (string, error) {
	tmpl, err := template.ParseFS(templateFS, "env.template")
	if err != nil {
		return "", fmt.Errorf("failed to parse template : %v", err)
	}

	// execute the struct against the template
	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, content)
	if err != nil {
		return "", fmt.Errorf("failed to execute template : %v", err)
	}

	return tpl.String(), nil
}

// Print issue.
func (rn ReleaseNote) String() (string, error) {
	pr := PRTemplate{
		RepoURL:  "https://github.com/Adarga-Ltd/" + rn.RepoName,
		RepoName: formatRepoName(rn.RepoName),
		Issues:   make([]IssueTemplate, 0, len(rn.Issues)),
	}

	for issue, commits := range rn.Issues {
		currentIssue := IssueTemplate{
			ID:      issue.Key,
			Labels:  issue.Fields.Labels,
			Summary: issue.Fields.Summary,
			Status:  issue.Fields.Status.Name,
			PRs:     []string{},
		}

		for _, commit := range commits {

			// get PR number from commit message
			prNumber := git.ExtractPR(strings.Split(commit.Message, "\n")[0])
			currentIssue.PRs = append(currentIssue.PRs, prNumber)
		}

		pr.Issues = append(pr.Issues, currentIssue)
	}

	// read in the template
	tmpl, err := template.ParseFS(templateFS, "notes.template")
	if err != nil {
		return "", fmt.Errorf("failed to parse template : %v", err)
	}

	// execute the struct against the template
	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, pr)
	if err != nil {
		return "", fmt.Errorf("failed to execute template : %v", err)
	}

	return tpl.String(), nil
}

// takes string-like-this and make it
// String Like This
func formatRepoName(s string) string {
	caser := cases.Title(language.English)

	words := strings.Split(s, "-")
	for i, word := range words {
		words[i] = caser.String(word)
	}

	return strings.Join(words, " ")
}

func CreateReleaseNotesForRepo(ctx context.Context, logger *zap.Logger, jiraClient *jira.Client, gitAuth *git.Auth, repoName string, tag1 string, tag2 string) ReleaseNote {
	logger = logger.With(zap.String("repo", repoName))
	repo, err := gitAuth.CloneRepo(repoName)
	if err != nil {
		logger.Error("failed to clone: skipping", zap.Error(err))
		return ReleaseNote{}
	}

	logger.Debug("getting commits between tags", zap.String("tag1", tag1), zap.String("tag2", tag2))
	commits, err := git.GetCommitsBetweenTags(repo, tag1, tag2)
	if err != nil {
		logger.Error("failed to get commits between tags", zap.String("tag1", tag1), zap.String("tag2", tag2), zap.Error(err))
		return ReleaseNote{}
	}

	issueCommitMap := make(IssueCommitMap)
	uniqueIssues := git.CommitsToIssues(commits)

	logger.Debug("unique issues", zap.Int("count", len(uniqueIssues)))

	for issueID, commits := range uniqueIssues {
		logger.Debug("searching for issue ", zap.String("issueID", issueID))

		found, _, err := jiraClient.Issue.Get(ctx, issueID, nil)
		if err != nil {
			logger.Error("failed to find issue", zap.String("issueID", issueID), zap.Error(err))
			continue
		}

		issueCommitMap[found] = commits

	}

	return ReleaseNote{
		RepoName: repoName,
		Issues:   issueCommitMap,
	}
}
