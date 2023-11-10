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
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/google/go-github/v56/github"
)

type Auth struct {
	Keys *ssh.PublicKeys
	Path string
}

func (g *Auth) CloneRepo(repo string) (*git.Repository, error) {
	fmt.Println("Cloning repo: ", g.Path+repo)
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
		name := r.Name().Short()
		name = strings.TrimPrefix(name, "v")

		v, err := semver.NewVersion(r.Name().Short())
		if err != nil {
			return nil
		}

		versions = append(versions, v)

		return nil
	})

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

func GetCommitsBetweenTags(r *git.Repository, tag1, tag2 string) ([]object.Commit, error) {
	tagIter, err := r.Tags()
	if err != nil {
		log.Fatal(err)
	}

	startSHA := plumbing.Hash{}
	endSHA := plumbing.Hash{}

	if !strings.HasPrefix(tag1, "v") {
		tag1 = "v" + tag1
	}
	if !strings.HasPrefix(tag2, "v") {
		tag2 = "v" + tag2
	}

	// check if the tag is missing a v prefix
	err = tagIter.ForEach(func(r *plumbing.Reference) error {
		if r.Name().Short() == tag1 {
			startSHA = r.Hash()
		} else if r.Name().Short() == tag2 {
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

	foundSHA := false
	var commits = []object.Commit{}
	err = cIter.ForEach(func(c *object.Commit) error { //TODO: this is actually pretty ineffective
		if foundSHA {
			return nil
		}
		if c.Hash.String() == startSHA.String() {
			foundSHA = true
			return nil
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

// Gets all unique tickets from a list of commits
func CommitsToIssues(commits []object.Commit) []string {
	commitMap := make(map[string]struct{})
	for _, commit := range commits {
		line := strings.Split(commit.Message, "\n")
		ticket := GetTicketFromCommitMessage(line[0])
		if ticket == "" {
			continue
		}

		commitMap[ticket] = struct{}{}
	}

	res := make([]string, 0, len(commitMap))

	for k := range commitMap {
		res = append(res, k)
	}
	return res
}

func ExtractRepo(line string) string {
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
