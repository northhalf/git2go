package git

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cgi"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	REMOTENAME = "testremote"
)

func TestClone(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	seedTestRepo(t, repo)

	path, err := ioutil.TempDir("", "git2go")
	checkFatal(t, err)

	ref, err := repo.References.Lookup("refs/heads/master")
	checkFatal(t, err)

	repo2, err := Clone(repo.Path(), path, &CloneOptions{Bare: true})
	defer cleanupTestRepo(t, repo2)

	checkFatal(t, err)

	ref2, err := repo2.References.Lookup("refs/heads/master")
	checkFatal(t, err)

	if ref.Cmp(ref2) != 0 {
		t.Fatal("reference in clone does not match original ref")
	}
}

func TestCloneWithCallback(t *testing.T) {
	t.Parallel()
	testPayload := 0

	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	seedTestRepo(t, repo)

	path, err := ioutil.TempDir("", "git2go")
	checkFatal(t, err)

	opts := CloneOptions{
		Bare: true,
		RemoteCreateCallback: func(r *Repository, name, url string) (*Remote, error) {
			testPayload += 1
			return r.Remotes.Create(REMOTENAME, url)
		},
	}

	repo2, err := Clone(repo.Path(), path, &opts)
	defer cleanupTestRepo(t, repo2)

	checkFatal(t, err)

	if testPayload != 1 {
		t.Fatal("Payload's value has not been changed")
	}

	remote, err := repo2.Remotes.Lookup(REMOTENAME)
	if err != nil || remote == nil {
		t.Fatal("Remote was not created properly")
	}
	defer remote.Free()
}

func startGitHTTPServer(t *testing.T) *httptest.Server {
	t.Helper()

	execPath, err := exec.Command("git", "--exec-path").Output()
	checkFatal(t, err)
	backend := filepath.Join(strings.TrimSpace(string(execPath)), "git-http-backend")
	projectRoot, err := filepath.Abs("testdata")
	checkFatal(t, err)

	cgiHandler := &cgi.Handler{
		Path: backend,
		Dir:  projectRoot,
		Env: []string{
			"GIT_PROJECT_ROOT=" + projectRoot,
			"GIT_HTTP_EXPORT_ALL=1",
		},
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.TransferEncoding) > 0 {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(body))
			r.ContentLength = int64(len(body))
			r.TransferEncoding = nil
		}
		cgiHandler.ServeHTTP(w, r)
	})
	return httptest.NewServer(handler)
}

func TestCloneWithHTTPUrl(t *testing.T) {
	server := startGitHTTPServer(t)
	defer server.Close()

	path, err := ioutil.TempDir("", "git2go")
	checkFatal(t, err)
	defer os.RemoveAll(path)

	repo, err := Clone(server.URL+"/TestGitRepository.git", path, &CloneOptions{})
	checkFatal(t, err)
	defer repo.Free()
}
