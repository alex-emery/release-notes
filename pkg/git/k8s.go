package git

import (
	"fmt"
	"io"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

// grabs images and tags from the k8s-engine repo and only
// from kustomization.yaml files.
func GetImagesFromK8s(r *git.Repository, sourceRefs, targetRefs string) ([]ImageDiff, error) {
	// Get the worktree
	w, err := r.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	// check out source
	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName(sourceRefs),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to checkout branch %s: %w", sourceRefs, err)
	}

	changedFiles, err := GetChangedKustomizations(r, targetRefs)
	if err != nil {
		return nil, fmt.Errorf("failed to get files: %w", err)
	}

	sourceYaml, err := ParseYamls(w, changedFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source yaml: %w", err)
	}

	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName(targetRefs),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to checkout branch %s: %w", targetRefs, err)
	}

	targetYaml, err := ParseYamls(w, changedFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to parse target yaml: %w", err)
	}

	return DiffKustomizations(sourceYaml, targetYaml), nil
}

type ImageDiff struct {
	Name string
	Tag1 string
	Tag2 string
}

func DiffKustomizations(original, dest []*types.Kustomization) []ImageDiff {
	results := []ImageDiff{}
	for i := range original {
		// assuming they're in the same order...
		for _, oimg := range original[i].Images {
			for _, dimg := range dest[i].Images {
				if oimg.Name == dimg.Name && oimg.NewTag != dimg.NewTag {
					results = append(results, ImageDiff{
						Name: oimg.Name,
						Tag1: oimg.NewTag,
						Tag2: dimg.NewTag,
					})
				}
			}
		}
	}

	return results
}

func ParseYamls(w *git.Worktree, filepath []string) ([]*types.Kustomization, error) {
	yamlFiles := []*types.Kustomization{}
	for _, file := range filepath {
		k, err := ParseYaml(w, file)
		if err != nil {
			return nil, err
		}
		yamlFiles = append(yamlFiles, k)
	}

	return yamlFiles, nil
}

func ParseYaml(w *git.Worktree, filepath string) (*types.Kustomization, error) {
	file, err := w.Filesystem.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filepath, err)
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filepath, err)
	}

	k := &types.Kustomization{}
	err = yaml.Unmarshal(data, k)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal file %s: %w", filepath, err)
	}

	return k, nil
}

func GetChangedKustomizations(r *git.Repository, branch string) ([]string, error) {
	// Get the HEAD reference
	headRef, err := r.Head()
	if err != nil {
		return nil, err
	}

	// Get the commit for the HEAD
	headCommit, err := r.CommitObject(headRef.Hash())
	if err != nil {
		return nil, err
	}

	// Get the branch reference
	branchRef, err := r.Reference(plumbing.ReferenceName(branch), true)
	if err != nil {
		return nil, err
	}

	// Get the commit for the branch
	branchCommit, err := r.CommitObject(branchRef.Hash())
	if err != nil {
		return nil, err
	}

	// Get the diff between the HEAD and the branch
	changes, err := headCommit.Patch(branchCommit)
	if err != nil {
		return nil, err
	}

	changedFiles := []string{}
	for _, change := range changes.FilePatches() {
		from, to := change.Files()
		if from != nil {
			if strings.HasSuffix(from.Path(), "kustomization.yaml") {
				changedFiles = append(changedFiles, from.Path())
			}
		} else {
			if strings.HasSuffix(to.Path(), "kustomization.yaml") {
				changedFiles = append(changedFiles, to.Path())
			}
		}
	}

	return changedFiles, nil
}
