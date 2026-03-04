// Package github provides a client for reading and writing Buildbarn jsonnet
// config files stored in the Hermetiq/bb-config GitHub repository.
package github

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	gogithub "github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"
)

const (
	DefaultOwner = "Hermetiq"
	DefaultRepo  = "bb-config"
)

// Client wraps the GitHub API for reading/writing Buildbarn config files.
type Client struct {
	gh    *gogithub.Client
	owner string
	repo  string
}

// NewClient creates a GitHub client authenticated via GITHUB_TOKEN env var.
func NewClient() (*Client, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable is required")
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(context.Background(), ts)
	ghClient := gogithub.NewClient(httpClient)
	return &Client{
		gh:    ghClient,
		owner: DefaultOwner,
		repo:  DefaultRepo,
	}, nil
}

// ListCollections returns all top-level directory names (collection keys) in
// the repo.
func (c *Client) ListCollections(ctx context.Context) ([]string, error) {
	_, dirContents, _, err := c.gh.Repositories.GetContents(ctx, c.owner, c.repo, "", nil)
	if err != nil {
		return nil, fmt.Errorf("listing repo root: %w", err)
	}
	var collections []string
	for _, item := range dirContents {
		if item.GetType() == "dir" {
			collections = append(collections, item.GetName())
		}
	}
	return collections, nil
}

// ListFiles returns all .jsonnet filenames inside a collection directory.
func (c *Client) ListFiles(ctx context.Context, collectionKey string) ([]string, error) {
	_, dirContents, _, err := c.gh.Repositories.GetContents(ctx, c.owner, c.repo, collectionKey, nil)
	if err != nil {
		return nil, fmt.Errorf("listing collection %q: %w", collectionKey, err)
	}
	var files []string
	for _, item := range dirContents {
		if item.GetType() == "file" && strings.HasSuffix(item.GetName(), ".jsonnet") {
			files = append(files, item.GetName())
		}
	}
	return files, nil
}

// ReadFile returns the raw jsonnet source of a config file.
func (c *Client) ReadFile(ctx context.Context, collectionKey, fileName string) (string, string, error) {
	path := collectionKey + "/" + fileName
	fileContent, _, resp, err := c.gh.Repositories.GetContents(ctx, c.owner, c.repo, path, nil)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return "", "", fmt.Errorf("file %q not found in collection %q", fileName, collectionKey)
		}
		return "", "", fmt.Errorf("reading file %q: %w", path, err)
	}
	if fileContent == nil {
		return "", "", fmt.Errorf("path %q is a directory, not a file", path)
	}
	decoded, err := fileContent.GetContent()
	if err != nil {
		return "", "", fmt.Errorf("decoding file content: %w", err)
	}
	return decoded, fileContent.GetSHA(), nil
}

// WriteFile commits updated jsonnet content back to the repo.
// sha is the blob SHA returned by ReadFile (required by the GitHub API).
func (c *Client) WriteFile(ctx context.Context, collectionKey, fileName, content, sha, message string) error {
	path := collectionKey + "/" + fileName
	opts := &gogithub.RepositoryContentFileOptions{
		Message: gogithub.String(message),
		Content: []byte(content),
		SHA:     gogithub.String(sha),
	}
	_, _, err := c.gh.Repositories.UpdateFile(ctx, c.owner, c.repo, path, opts)
	if err != nil {
		return fmt.Errorf("writing file %q: %w", path, err)
	}
	return nil
}
