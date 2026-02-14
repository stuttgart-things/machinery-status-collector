package git

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// GitHubClient interacts with the GitHub REST API to manage files, branches, and pull requests.
type GitHubClient struct {
	token      string
	owner      string
	repo       string
	httpClient *http.Client
	baseURL    string
}

// NewGitHubClient creates a GitHubClient configured for the given repository.
func NewGitHubClient(token, owner, repo string) *GitHubClient {
	return &GitHubClient{
		token:      token,
		owner:      owner,
		repo:       repo,
		httpClient: &http.Client{},
		baseURL:    "https://api.github.com",
	}
}

// FetchFile retrieves a file's content and SHA from the given ref.
func (c *GitHubClient) FetchFile(path, ref string) ([]byte, string, error) {
	apiPath := fmt.Sprintf("/repos/%s/%s/contents/%s?ref=%s",
		url.PathEscape(c.owner), url.PathEscape(c.repo), path, url.QueryEscape(ref))

	resp, err := c.doRequest(http.MethodGet, apiPath, nil)
	if err != nil {
		return nil, "", fmt.Errorf("fetch file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("fetch file: unexpected status %d", resp.StatusCode)
	}

	var result struct {
		Content  string `json:"content"`
		SHA      string `json:"sha"`
		Encoding string `json:"encoding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, "", fmt.Errorf("fetch file: decode response: %w", err)
	}

	decoded, err := base64.StdEncoding.DecodeString(result.Content)
	if err != nil {
		return nil, "", fmt.Errorf("fetch file: decode base64: %w", err)
	}

	return decoded, result.SHA, nil
}

// CreateBranch creates a new branch pointing at the given SHA.
func (c *GitHubClient) CreateBranch(baseSHA, branchName string) error {
	apiPath := fmt.Sprintf("/repos/%s/%s/git/refs",
		url.PathEscape(c.owner), url.PathEscape(c.repo))

	body := map[string]string{
		"ref": "refs/heads/" + branchName,
		"sha": baseSHA,
	}

	resp, err := c.doRequest(http.MethodPost, apiPath, body)
	if err != nil {
		return fmt.Errorf("create branch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("create branch: unexpected status %d", resp.StatusCode)
	}

	return nil
}

// UpdateFile commits an update to a file on the given branch.
func (c *GitHubClient) UpdateFile(path, branchName, message string, content []byte, sha string) error {
	apiPath := fmt.Sprintf("/repos/%s/%s/contents/%s",
		url.PathEscape(c.owner), url.PathEscape(c.repo), path)

	body := map[string]string{
		"message": message,
		"content": base64.StdEncoding.EncodeToString(content),
		"sha":     sha,
		"branch":  branchName,
	}

	resp, err := c.doRequest(http.MethodPut, apiPath, body)
	if err != nil {
		return fmt.Errorf("update file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("update file: unexpected status %d", resp.StatusCode)
	}

	return nil
}

// CreatePR opens a pull request and returns the PR number.
func (c *GitHubClient) CreatePR(title, body, head, base string) (int, error) {
	apiPath := fmt.Sprintf("/repos/%s/%s/pulls",
		url.PathEscape(c.owner), url.PathEscape(c.repo))

	reqBody := map[string]string{
		"title": title,
		"body":  body,
		"head":  head,
		"base":  base,
	}

	resp, err := c.doRequest(http.MethodPost, apiPath, reqBody)
	if err != nil {
		return 0, fmt.Errorf("create PR: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return 0, fmt.Errorf("create PR: unexpected status %d", resp.StatusCode)
	}

	var result struct {
		Number int `json:"number"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("create PR: decode response: %w", err)
	}

	return result.Number, nil
}

// ListOpenPRs returns PR numbers for open PRs from the given head branch.
func (c *GitHubClient) ListOpenPRs(head string) ([]int, error) {
	apiPath := fmt.Sprintf("/repos/%s/%s/pulls?state=open&head=%s:%s",
		url.PathEscape(c.owner), url.PathEscape(c.repo),
		url.QueryEscape(c.owner), url.QueryEscape(head))

	resp, err := c.doRequest(http.MethodGet, apiPath, nil)
	if err != nil {
		return nil, fmt.Errorf("list open PRs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list open PRs: unexpected status %d", resp.StatusCode)
	}

	var prs []struct {
		Number int `json:"number"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
		return nil, fmt.Errorf("list open PRs: decode response: %w", err)
	}

	numbers := make([]int, len(prs))
	for i, pr := range prs {
		numbers[i] = pr.Number
	}
	return numbers, nil
}

func (c *GitHubClient) doRequest(method, path string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github+json")

	return c.httpClient.Do(req)
}
