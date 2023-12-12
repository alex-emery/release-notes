package git_test

import (
	"testing"

	"github.com/alex-emery/release-notes/pkg/git"
	"github.com/stretchr/testify/assert"
)

func TestTicketExtraction(t *testing.T) {
	testCases := []struct {
		line     string
		expected string
	}{
		{
			line:     "3eb443b feat: [APP-12431] increase max recv size for GetLeafVectors gRPC call (#138)",
			expected: "APP-12431",
		},
		{
			line:     "3eb443b feat: APP-2117_fix_meta_parsing increase max recv size for GetLeafVectors gRPC call (#138)",
			expected: "APP-2117",
		},
	}

	for _, tc := range testCases {
		actual := git.GetTicketFromCommitMessage(tc.line)
		assert.Equal(t, tc.expected, actual)
	}
}

func TestExtractRepo(t *testing.T) {
	testCases := []struct {
		repo     string
		expected string
	}{
		{
			repo:     "adarga/k8s-engine",
			expected: "k8s-engine",
		},
		{
			repo:     "975704811528.dkr.ecr.eu-west-2.amazonaws.com/adarga/engine-health-metrics",
			expected: "engine-health-metrics",
		},
		{
			repo:     "",
			expected: "",
		},
		{
			repo:     "adarga",
			expected: "",
		},
		{
			repo:     "unrelated/repo",
			expected: "",
		},
	}

	for _, tc := range testCases {
		actual := git.ExtractRepoName(tc.repo)
		assert.Equal(t, tc.expected, actual)
	}

}

func TestExtractPR(t *testing.T) {
	testCases := []struct {
		line     string
		expected string
	}{
		{
			line:     "feat(APP-12505): support answer rating (#174)",
			expected: "174",
		},
		{
			line:     "feat(APP-12505): support answer rating",
			expected: "",
		},
	}

	for _, tc := range testCases {
		actual := git.ExtractPR(tc.line)
		assert.Equal(t, tc.expected, actual)
	}
}
