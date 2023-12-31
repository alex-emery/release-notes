package git

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/google/go-github/v56/github"
	"go.uber.org/zap"
)

type Auth struct {
	Keys   *ssh.PublicKeys
	Path   string
	logger *zap.Logger
}

// GetK8sEngineRepo either clones the repo if the path is empty or opens an existing repo.
func (g *Auth) GetK8sEngineRepo(path string) (*git.Repository, error) {
	if path == "" {
		return g.CloneRepo("k8s-engine")
	}

	return g.OpenExisting(path)
}

func (g *Auth) OpenExisting(repo string) (*git.Repository, error) {
	return git.PlainOpen(repo)
}

func (g *Auth) CloneRepo(repo string) (*git.Repository, error) {
	g.logger.Debug(fmt.Sprintf("Cloning repo: %s%s", g.Path, repo))
	return git.Clone(memory.NewStorage(), memfs.New(), &git.CloneOptions{
		Auth: g.Keys,
		URL:  g.Path + repo,
	})
}

func GetTagsBetweenTags(r *git.Repository, tag1, tag2 string) ([]string, error) {
	iter, err := r.Tags()
	if err != nil {
		return nil, err
	}

	versions := []*semver.Version{}

	err = iter.ForEach(func(r *plumbing.Reference) error {
		name := strings.TrimPrefix(r.Name().Short(), "v")
		name = strings.TrimPrefix(name, "v")

		v, err := semver.NewVersion(name)
		if err != nil {
			return nil
		}

		versions = append(versions, v)

		return nil
	})
	if err != nil {
		return nil, err
	}

	tag1 = strings.TrimPrefix(tag1, "v")
	tag2 = strings.TrimPrefix(tag2, "v")

	end, _ := semver.NewVersion(tag1)
	start, _ := semver.NewVersion(tag2)

	res := make([]string, 0)
	for _, v := range versions {
		if v.GreaterThan(start) && v.LessThan(end) || v.Equal(end) {
			res = append(res, "v"+v.String())
		}
	}

	return res, nil
}

func GetReleaseForTags(client *github.Client, repo, tag string) string {
	tagResp, _, err := client.Repositories.GetReleaseByTag(context.Background(), "Adarga-Ltd", repo, tag)
	if err != nil {
		log.Fatal(err)
	}

	return tagResp.GetBody()
}

func Checkout(r *git.Repository, branch *plumbing.Reference) error {
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	err = w.Checkout(&git.CheckoutOptions{
		Branch: branch.Name(),
	})

	if err != nil {
		return fmt.Errorf("failed to checkout branch %s: %w", branch.Name(), err)
	}
	return nil
}
func CompareTags(tag1, tag2 string) (int, error) {
	// strip off any v prefix
	tag1 = strings.TrimPrefix(tag1, "v")
	tag2 = strings.TrimPrefix(tag2, "v")

	v1, err := semver.NewVersion(tag1)
	if err != nil {
		return 0, err
	}

	v2, err := semver.NewVersion(tag2)
	if err != nil {
		return 0, err
	}

	return v1.Compare(v2), nil
}

func GetCommitsBetweenTags(r *git.Repository, tag1, tag2 string) ([]object.Commit, error) {
	tagIter, err := r.Tags()
	if err != nil {
		log.Fatal(err)
	}

	startSHA := plumbing.Hash{}
	endSHA := plumbing.Hash{}

	// only get commits if tag1 < tag2
	val, err := CompareTags(tag1, tag2)
	if err != nil {
		return nil, err
	}

	if val != -1 {
		return nil, fmt.Errorf("tag1: %s must be less than tag2: %s", tag1, tag2)
	}

	tag1 = strings.TrimLeft(tag1, "v")
	tag2 = strings.TrimLeft(tag2, "v")

	// check if the tag is missing a v prefix
	err = tagIter.ForEach(func(r *plumbing.Reference) error {
		name := r.Name().Short()
		name = strings.TrimPrefix(name, "v")
		name = strings.TrimPrefix(name, "deploy-")
		if name == tag1 {
			startSHA = r.Hash()
		} else if name == tag2 {
			endSHA = r.Hash()
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if startSHA.IsZero() {
		return nil, fmt.Errorf("failed to find start SHA: %s", tag1)
	}

	if endSHA.IsZero() {
		return nil, fmt.Errorf("failed to find end SHA: %s", tag2)
	}

	cIter, err := r.Log(&git.LogOptions{
		From: endSHA,
	})

	if err != nil {
		return nil, err
	}

	var commits = []object.Commit{}
	err = cIter.ForEach(func(c *object.Commit) error {
		if c.Hash.String() == startSHA.String() {
			return storer.ErrStop
		}

		commits = append(commits, *c)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return commits, nil
}

// uses regex to extract the ticket from the string.
// the ticket is formatted as APP-\w{4}
func GetTicketFromCommitMessage(message string) string {
	re := regexp.MustCompile(`APP-\d+`)
	return re.FindString(message)
}

// Bundles commits to the respective ticket
type IssueCommitMap map[string][]object.Commit

// Gets all unique tickets from a list of commits
func CommitsToIssues(commits []object.Commit) IssueCommitMap {
	commitMap := make(IssueCommitMap)
	for _, commit := range commits {
		line := strings.Split(commit.Message, "\n")
		ticket := GetTicketFromCommitMessage(line[0])
		if ticket == "" {
			continue
		}

		if _, ok := commitMap[ticket]; !ok {
			commitMap[ticket] = []object.Commit{}
		}
		commitMap[ticket] = append(commitMap[ticket], commit)
	}

	return commitMap
}

func ExtractRepoName(line string) string {
	var re = regexp.MustCompile(`(?m)adarga/(?P<repo>.+)`)

	match := re.FindStringSubmatch(line)
	if len(match) == 0 {
		return ""
	}

	result := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}

	return result["repo"]
}

func ExtractPR(s string) string {
	re := regexp.MustCompile(`\(#(\d+)\)`)
	match := re.FindStringSubmatch(s)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}
