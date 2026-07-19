package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWorktreeLifecycle(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)
	seedTestRepo(t, repo)

	path := filepath.Join(t.TempDir(), "linked")
	worktree, err := repo.AddWorktree("linked", path, nil)
	checkFatal(t, err)
	defer worktree.Free()

	if worktree.Name() != "linked" {
		t.Fatalf("Name() = %q, want linked", worktree.Name())
	}
	if worktree.Path() != path {
		t.Fatalf("Path() = %q, want %q", worktree.Path(), path)
	}
	checkFatal(t, worktree.Validate())

	names, err := repo.Worktrees()
	checkFatal(t, err)
	if len(names) != 1 || names[0] != "linked" {
		t.Fatalf("Worktrees() = %v, want [linked]", names)
	}

	lookedUp, err := repo.LookupWorktree("linked")
	checkFatal(t, err)
	if lookedUp.Name() != "linked" {
		t.Fatalf("LookupWorktree().Name() = %q, want linked", lookedUp.Name())
	}
	checkFatal(t, lookedUp.Free())

	linkedRepo, err := OpenRepository(path)
	checkFatal(t, err)
	opened, err := linkedRepo.OpenWorktree()
	checkFatal(t, err)
	if opened.Name() != "linked" {
		t.Fatalf("OpenWorktree().Name() = %q, want linked", opened.Name())
	}
	checkFatal(t, opened.Free())
	linkedRepo.Free()

	checkFatal(t, worktree.Lock("maintenance"))
	locked, reason, err := worktree.IsLocked()
	checkFatal(t, err)
	if !locked || reason != "maintenance" {
		t.Fatalf("IsLocked() = %v, %q, want true, maintenance", locked, reason)
	}
	checkFatal(t, worktree.Unlock())

	checkFatal(t, os.RemoveAll(path))
	prunable, err := worktree.IsPrunable(nil)
	checkFatal(t, err)
	if !prunable {
		t.Fatal("removed worktree is not prunable")
	}
	checkFatal(t, worktree.Prune(nil))
}

func TestWorktreeOptionsCompile(t *testing.T) {
	_ = WorktreeAddOptions{
		Lock:             true,
		CheckoutExisting: true,
		CheckoutOptions:  CheckoutOptions{Strategy: CheckoutSafe},
	}
	_ = WorktreePruneOptions{
		Flags: WorktreePruneValid | WorktreePruneLocked | WorktreePruneWorkingTree,
	}
	if ConfigLevelWorktree == ConfigLevelApp {
		t.Fatal("ConfigLevelWorktree and ConfigLevelApp must differ")
	}
}
