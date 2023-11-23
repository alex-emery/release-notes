package github

import (
	"context"

	"github.com/google/go-github/v56/github"
	"go.uber.org/zap"
)

type Client struct {
	client *github.Client
	logger *zap.Logger
}

func New(logger *zap.Logger, token string) *Client {
	return &Client{
		logger: logger,
		client: github.NewClient(nil).WithAuthToken(token),
	}
}

func (c *Client) CreatePR(ctx context.Context, head, base, title, body string) error {
	c.logger.Debug("creating PR", zap.String("head", head), zap.String("base", base), zap.String("title", title), zap.String("body", body))
	_, _, err := c.client.PullRequests.Create(ctx, "Adarga-Ltd", "k8s-engine", &github.NewPullRequest{
		Title: github.String(title),
		Body:  github.String(body),
		Head:  github.String(head),
		Base:  github.String(base),
	})

	return err
}
