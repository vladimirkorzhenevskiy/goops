package gitlab

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
)

type Config struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

func (c Config) Validate() error {
	if c.BaseURL == "" {
		return errors.New("`baseURL` undefined")
	}

	if c.BaseURL == "" {
		return errors.New("`apiKey` undefined")
	}

	return nil
}

type Client struct {
	cfg Config
	sdk *gitlab.Client
}

func New(cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Timeout: cfg.Timeout,
	}

	sdk, err := gitlab.NewClient(
		cfg.APIKey,
		gitlab.WithBaseURL(cfg.BaseURL),
		gitlab.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, err
	}

	client := &Client{
		cfg: cfg,
		sdk: sdk,
	}

	return client, nil
}

func (c *Client) GetMergeRequest(ctx context.Context, projectID string, mergeRequestID int64) (*MergeRequest, error) {
	res, _, err := c.sdk.MergeRequests.GetMergeRequest(
		projectID,
		mergeRequestID,
		nil,
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("fetch raw diff from gitlab: %w", err)
	}

	return &MergeRequest{
		IID: res.IID,
		Refs: MergeRequestRefs{
			BaseSHA:  res.DiffRefs.BaseSha,
			HeadSHA:  res.DiffRefs.HeadSha,
			StartSHA: res.DiffRefs.StartSha,
		},
	}, nil
}

func (c *Client) FetchMergeRequestDiff(ctx context.Context, projectID string, mergeRequestID int64) ([]*MergeRequestDiff, error) {
	res, _, err := c.sdk.MergeRequests.ShowMergeRequestRawDiffs(
		projectID,
		mergeRequestID,
		nil,
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("fetch raw diff from gitlab: %w", err)
	}

	return ParseMergeRequestDiff(res), nil
}

func (c *Client) CreateMergeRequestDiscussion(
	ctx context.Context,
	projectID string,
	mergeRequestID int64,
	refs MergeRequestRefs,
	filePath string,
	line int64,
	text string,
) error {
	opts := &gitlab.CreateMergeRequestDiscussionOptions{
		Body: gitlab.Ptr(text),
		Position: &gitlab.PositionOptions{
			BaseSHA:      gitlab.Ptr(refs.BaseSHA),
			StartSHA:     gitlab.Ptr(refs.StartSHA),
			HeadSHA:      gitlab.Ptr(refs.HeadSHA),
			PositionType: gitlab.Ptr("text"),
			NewPath:      gitlab.Ptr(filePath),
			NewLine:      gitlab.Ptr(line),
		},
	}

	_, _, err := c.sdk.Discussions.CreateMergeRequestDiscussion(projectID, mergeRequestID, opts, gitlab.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("create discussion on %s:%d: %w", filePath, line, err)
	}

	return nil
}
