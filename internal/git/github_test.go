package git

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(url, token string) *GitHubClient {
	return &GitHubClient{
		token:      token,
		owner:      "test-owner",
		repo:       "test-repo",
		httpClient: &http.Client{},
		baseURL:    url,
	}
}

func TestFetchFile(t *testing.T) {
	want := []byte("hello world")
	wantSHA := "abc123sha"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"content":  base64.StdEncoding.EncodeToString(want),
			"sha":      wantSHA,
			"encoding": "base64",
		})
	}))
	defer srv.Close()

	client := newTestClient(srv.URL, "test-token")
	content, sha, err := client.FetchFile("path/to/file.yaml", "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(content) != string(want) {
		t.Fatalf("expected content %q, got %q", want, content)
	}
	if sha != wantSHA {
		t.Fatalf("expected sha %q, got %q", wantSHA, sha)
	}
}

func TestFetchFile_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Not Found"})
	}))
	defer srv.Close()

	client := newTestClient(srv.URL, "test-token")
	_, _, err := client.FetchFile("nonexistent.yaml", "main")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}

func TestCreateBranch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if body["ref"] != "refs/heads/my-branch" {
			t.Fatalf("expected ref 'refs/heads/my-branch', got %q", body["ref"])
		}
		if body["sha"] != "deadbeef" {
			t.Fatalf("expected sha 'deadbeef', got %q", body["sha"])
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"ref": "refs/heads/my-branch"})
	}))
	defer srv.Close()

	client := newTestClient(srv.URL, "test-token")
	if err := client.CreateBranch("deadbeef", "my-branch"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateBranch_AlreadyExists(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{"message": "Reference already exists"})
	}))
	defer srv.Close()

	client := newTestClient(srv.URL, "test-token")
	if err := client.CreateBranch("deadbeef", "existing-branch"); err == nil {
		t.Fatal("expected error for 422 response")
	}
}

func TestUpdateFile(t *testing.T) {
	fileContent := []byte("updated content")
	fileSHA := "oldsha123"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}

		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		expectedContent := base64.StdEncoding.EncodeToString(fileContent)
		if body["content"] != expectedContent {
			t.Fatalf("expected base64 content %q, got %q", expectedContent, body["content"])
		}
		if body["sha"] != fileSHA {
			t.Fatalf("expected sha %q, got %q", fileSHA, body["sha"])
		}
		if body["branch"] != "update-branch" {
			t.Fatalf("expected branch 'update-branch', got %q", body["branch"])
		}
		if body["message"] != "update file" {
			t.Fatalf("expected message 'update file', got %q", body["message"])
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"content": map[string]string{"sha": "newsha456"}})
	}))
	defer srv.Close()

	client := newTestClient(srv.URL, "test-token")
	if err := client.UpdateFile("path/to/file.yaml", "update-branch", "update file", fileContent, fileSHA); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreatePR(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if body["title"] != "Update status" {
			t.Fatalf("expected title 'Update status', got %q", body["title"])
		}
		if body["body"] != "Status update PR" {
			t.Fatalf("expected body 'Status update PR', got %q", body["body"])
		}
		if body["head"] != "feature-branch" {
			t.Fatalf("expected head 'feature-branch', got %q", body["head"])
		}
		if body["base"] != "main" {
			t.Fatalf("expected base 'main', got %q", body["base"])
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"number": 42})
	}))
	defer srv.Close()

	client := newTestClient(srv.URL, "test-token")
	num, err := client.CreatePR("Update status", "Status update PR", "feature-branch", "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if num != 42 {
		t.Fatalf("expected PR number 42, got %d", num)
	}
}

func TestListOpenPRs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"number": 10, "title": "PR 10"},
			{"number": 20, "title": "PR 20"},
			{"number": 30, "title": "PR 30"},
		})
	}))
	defer srv.Close()

	client := newTestClient(srv.URL, "test-token")
	numbers, err := client.ListOpenPRs("my-branch")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(numbers) != 3 {
		t.Fatalf("expected 3 PRs, got %d", len(numbers))
	}
	expected := []int{10, 20, 30}
	for i, n := range numbers {
		if n != expected[i] {
			t.Fatalf("expected PR number %d at index %d, got %d", expected[i], i, n)
		}
	}
}

func TestListOpenPRs_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{})
	}))
	defer srv.Close()

	client := newTestClient(srv.URL, "test-token")
	numbers, err := client.ListOpenPRs("no-prs-branch")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(numbers) != 0 {
		t.Fatalf("expected 0 PRs, got %d", len(numbers))
	}
}

func TestAuthHeader(t *testing.T) {
	wantToken := "my-secret-token"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer "+wantToken {
			t.Fatalf("expected Authorization 'Bearer %s', got %q", wantToken, auth)
		}

		// Return a valid response so the method doesn't error on decode.
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"content":  base64.StdEncoding.EncodeToString([]byte("data")),
			"sha":      "sha1",
			"encoding": "base64",
		})
	}))
	defer srv.Close()

	client := newTestClient(srv.URL, wantToken)

	// Exercise FetchFile which uses doRequest under the hood.
	_, _, err := client.FetchFile("any/file", "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
