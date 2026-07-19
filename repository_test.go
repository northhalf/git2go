package git

import (
	"os"
	"testing"
	"time"
)

func TestCreateCommitBuffer(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	loc, err := time.LoadLocation("Europe/Berlin")
	checkFatal(t, err)
	sig := &Signature{
		Name:  "Rand Om Hacker",
		Email: "random@hacker.com",
		When:  time.Date(2013, 03, 06, 14, 30, 0, 0, loc),
	}

	idx, err := repo.Index()
	checkFatal(t, err)
	err = idx.AddByPath("README")
	checkFatal(t, err)
	err = idx.Write()
	checkFatal(t, err)
	treeId, err := idx.WriteTree()
	checkFatal(t, err)

	message := "This is a commit\n"
	tree, err := repo.LookupTree(treeId)
	checkFatal(t, err)

	for encoding, expected := range map[MessageEncoding]string{
		MessageEncodingUTF8: `tree b7119b11e8ef7a1a5a34d3ac87f5b075228ac81e
author Rand Om Hacker <random@hacker.com> 1362576600 +0100
committer Rand Om Hacker <random@hacker.com> 1362576600 +0100

This is a commit
`,
		MessageEncoding("ASCII"): `tree b7119b11e8ef7a1a5a34d3ac87f5b075228ac81e
author Rand Om Hacker <random@hacker.com> 1362576600 +0100
committer Rand Om Hacker <random@hacker.com> 1362576600 +0100
encoding ASCII

This is a commit
`,
	} {
		encoding := encoding
		expected := expected
		t.Run(string(encoding), func(t *testing.T) {
			buf, err := repo.CreateCommitBuffer(sig, sig, encoding, message, tree)
			checkFatal(t, err)

			if expected != string(buf) {
				t.Errorf("mismatched commit buffer, expected %v, got %v", expected, string(buf))
			}
		})
	}
}

func TestCreateCommitFromIds(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	loc, err := time.LoadLocation("Europe/Berlin")
	checkFatal(t, err)
	sig := &Signature{
		Name:  "Rand Om Hacker",
		Email: "random@hacker.com",
		When:  time.Date(2013, 03, 06, 14, 30, 0, 0, loc),
	}

	idx, err := repo.Index()
	checkFatal(t, err)
	err = idx.AddByPath("README")
	checkFatal(t, err)
	err = idx.Write()
	checkFatal(t, err)
	treeId, err := idx.WriteTree()
	checkFatal(t, err)

	message := "This is a commit\n"
	tree, err := repo.LookupTree(treeId)
	checkFatal(t, err)
	expectedCommitId, err := repo.CreateCommit("HEAD", sig, sig, message, tree)
	checkFatal(t, err)

	commitId, err := repo.CreateCommitFromIds("", sig, sig, message, treeId)
	checkFatal(t, err)

	if !expectedCommitId.Equal(commitId) {
		t.Errorf("mismatched commit ids, expected %v, got %v", expectedCommitId.String(), commitId.String())
	}
}

func TestCreateCommitFromStageAndParents(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)
	seedTestRepo(t, repo)

	checkFatal(t, os.WriteFile(repo.Workdir()+"README", []byte("staged change\n"), 0644))
	index, err := repo.Index()
	checkFatal(t, err)
	defer index.Free()
	checkFatal(t, index.AddByPath("README"))
	checkFatal(t, index.Write())

	sig := signature()
	oid, err := repo.CreateCommitFromStage("staged commit", &CommitCreateOptions{
		Author:    sig,
		Committer: sig,
	})
	checkFatal(t, err)

	commit, err := repo.LookupCommit(oid)
	checkFatal(t, err)
	defer commit.Free()
	if commit.Summary() != "staged commit" {
		t.Fatalf("commit summary = %q, want staged commit", commit.Summary())
	}

	parents, err := repo.CommitParents()
	checkFatal(t, err)
	defer func() {
		for _, parent := range parents {
			parent.Free()
		}
	}()
	if len(parents) != 1 || !parents[0].Id().Equal(oid) {
		t.Fatalf("CommitParents() returned %d unexpected parents", len(parents))
	}

	oidType := repo.OidType()
	if oidType != OidTypeSHA1 {
		t.Fatalf("Repository.OidType() = %v, want %v", oidType, OidTypeSHA1)
	}
}

func TestDefaultSignaturesFromEnv(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	t.Setenv("GIT_AUTHOR_NAME", "Author Name")
	t.Setenv("GIT_AUTHOR_EMAIL", "author@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "Committer Name")
	t.Setenv("GIT_COMMITTER_EMAIL", "committer@example.com")

	author, committer, err := repo.DefaultSignaturesFromEnv()
	checkFatal(t, err)
	if author.Name != "Author Name" || author.Email != "author@example.com" {
		t.Fatalf("author = %+v", author)
	}
	if committer.Name != "Committer Name" || committer.Email != "committer@example.com" {
		t.Fatalf("committer = %+v", committer)
	}
}

func TestRepositorySetConfig(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	loc, err := time.LoadLocation("Europe/Berlin")
	checkFatal(t, err)
	sig := &Signature{
		Name:  "Rand Om Hacker",
		Email: "random@hacker.com",
		When:  time.Date(2013, 03, 06, 14, 30, 0, 0, loc),
	}

	idx, err := repo.Index()
	checkFatal(t, err)
	err = idx.AddByPath("README")

	treeId, err := idx.WriteTree()
	checkFatal(t, err)

	message := "This is a commit\n"
	tree, err := repo.LookupTree(treeId)
	checkFatal(t, err)
	_, err = repo.CreateCommit("HEAD", sig, sig, message, tree)
	checkFatal(t, err)

	repoConfig, err := repo.Config()
	checkFatal(t, err)

	temp := Config{}
	localConfig, err := temp.OpenLevel(repoConfig, ConfigLevelLocal)
	checkFatal(t, err)
	repoConfig = nil

	err = repo.SetConfig(localConfig)
	checkFatal(t, err)

	configFieldName := "core.filemode"
	err = localConfig.SetBool(configFieldName, true)
	checkFatal(t, err)

	localConfig = nil

	repoConfig, err = repo.Config()
	checkFatal(t, err)

	result, err := repoConfig.LookupBool(configFieldName)
	checkFatal(t, err)
	if result != true {
		t.Fatal("result must be true")
	}
}

func TestRepositoryItemPath(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	gitDir, err := repo.ItemPath(RepositoryItemGitDir)
	checkFatal(t, err)
	if gitDir == "" {
		t.Error("expected not empty gitDir")
	}
}
