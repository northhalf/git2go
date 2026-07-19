package git

/*
#include <git2.h>
*/
import "C"
import (
	"runtime"
	"unsafe"
)

type Worktree struct {
	doNotCompare
	ptr  *C.git_worktree
	repo *Repository
}

type WorktreeAddOptions struct {
	Lock             bool
	CheckoutExisting bool
	Reference        *Reference
	CheckoutOptions  CheckoutOptions
}

type WorktreePruneFlag uint32

const (
	WorktreePruneValid       WorktreePruneFlag = C.GIT_WORKTREE_PRUNE_VALID
	WorktreePruneLocked      WorktreePruneFlag = C.GIT_WORKTREE_PRUNE_LOCKED
	WorktreePruneWorkingTree WorktreePruneFlag = C.GIT_WORKTREE_PRUNE_WORKING_TREE
)

type WorktreePruneOptions struct {
	Flags WorktreePruneFlag
}

func newWorktree(ptr *C.git_worktree, repo *Repository) *Worktree {
	worktree := &Worktree{ptr: ptr, repo: repo}
	runtime.SetFinalizer(worktree, (*Worktree).Free)
	return worktree
}

func (repo *Repository) Worktrees() ([]string, error) {
	var names C.git_strarray

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_worktree_list(&names, repo.ptr)
	runtime.KeepAlive(repo)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	defer C.git_strarray_dispose(&names)
	return makeStringsFromCStrings(names.strings, int(names.count)), nil
}

func (repo *Repository) LookupWorktree(name string) (*Worktree, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	var ptr *C.git_worktree
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_worktree_lookup(&ptr, repo.ptr, cname)
	runtime.KeepAlive(repo)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	return newWorktree(ptr, repo), nil
}

func (repo *Repository) OpenWorktree() (*Worktree, error) {
	var ptr *C.git_worktree

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_worktree_open_from_repository(&ptr, repo.ptr)
	runtime.KeepAlive(repo)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	return newWorktree(ptr, repo), nil
}

func populateWorktreeAddOptions(copts *C.git_worktree_add_options, opts *WorktreeAddOptions, errorTarget *error) *C.git_worktree_add_options {
	if opts == nil {
		return nil
	}
	C.git_worktree_add_options_init(copts, C.GIT_WORKTREE_ADD_OPTIONS_VERSION)
	copts.lock = cbool(opts.Lock)
	copts.checkout_existing = cbool(opts.CheckoutExisting)
	if opts.Reference != nil {
		copts.ref = opts.Reference.ptr
	}
	populateCheckoutOptions(&copts.checkout_options, &opts.CheckoutOptions, errorTarget)
	return copts
}

func freeWorktreeAddOptions(copts *C.git_worktree_add_options) {
	if copts == nil {
		return
	}
	freeCheckoutOptions(&copts.checkout_options)
}

func (repo *Repository) AddWorktree(name, path string, opts *WorktreeAddOptions) (*Worktree, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	var callbackErr error
	copts := populateWorktreeAddOptions(&C.git_worktree_add_options{}, opts, &callbackErr)
	defer freeWorktreeAddOptions(copts)

	var ptr *C.git_worktree
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_worktree_add(&ptr, repo.ptr, cname, cpath, copts)
	runtime.KeepAlive(repo)
	if ret == C.int(ErrorCodeUser) && callbackErr != nil {
		return nil, callbackErr
	}
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	return newWorktree(ptr, repo), nil
}

func (worktree *Worktree) Free() error {
	if worktree.ptr == nil {
		return ErrInvalid
	}
	runtime.SetFinalizer(worktree, nil)
	C.git_worktree_free(worktree.ptr)
	worktree.ptr = nil
	worktree.repo = nil
	return nil
}

func (worktree *Worktree) Name() string {
	name := C.git_worktree_name(worktree.ptr)
	runtime.KeepAlive(worktree)
	if name == nil {
		return ""
	}
	return C.GoString(name)
}

func (worktree *Worktree) Path() string {
	path := C.git_worktree_path(worktree.ptr)
	runtime.KeepAlive(worktree)
	if path == nil {
		return ""
	}
	return C.GoString(path)
}

func (worktree *Worktree) Validate() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_worktree_validate(worktree.ptr)
	runtime.KeepAlive(worktree)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (worktree *Worktree) Lock(reason string) error {
	creason := C.CString(reason)
	defer C.free(unsafe.Pointer(creason))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_worktree_lock(worktree.ptr, creason)
	runtime.KeepAlive(worktree)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (worktree *Worktree) Unlock() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_worktree_unlock(worktree.ptr)
	runtime.KeepAlive(worktree)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (worktree *Worktree) IsLocked() (bool, string, error) {
	var reason C.git_buf
	defer C.git_buf_dispose(&reason)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_worktree_is_locked(&reason, worktree.ptr)
	runtime.KeepAlive(worktree)
	if ret < 0 {
		return false, "", MakeGitError(ret)
	}
	return ret > 0, C.GoString(reason.ptr), nil
}

func populateWorktreePruneOptions(copts *C.git_worktree_prune_options, opts *WorktreePruneOptions) *C.git_worktree_prune_options {
	if opts == nil {
		return nil
	}
	C.git_worktree_prune_options_init(copts, C.GIT_WORKTREE_PRUNE_OPTIONS_VERSION)
	copts.flags = C.uint32_t(opts.Flags)
	return copts
}

func (worktree *Worktree) IsPrunable(opts *WorktreePruneOptions) (bool, error) {
	copts := populateWorktreePruneOptions(&C.git_worktree_prune_options{}, opts)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_worktree_is_prunable(worktree.ptr, copts)
	runtime.KeepAlive(worktree)
	if ret < 0 {
		return false, MakeGitError(ret)
	}
	return ret > 0, nil
}

func (worktree *Worktree) Prune(opts *WorktreePruneOptions) error {
	copts := populateWorktreePruneOptions(&C.git_worktree_prune_options{}, opts)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_worktree_prune(worktree.ptr, copts)
	runtime.KeepAlive(worktree)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}
