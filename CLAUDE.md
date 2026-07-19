# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository overview

`git2go` is a Go module (`github.com/northhalf/git2go/v38`) that exposes libgit2 through cgo. This is a maintained fork of the upstream `libgit2/git2go` (unmaintained since 2022) that targets libgit2 1.9.x. The repository contains one Go package, `git`, split by libgit2 subsystem rather than by internal packages. The `third_party/libgit2` Git submodule supplies the bundled C library, pinned to v1.9.6; this version of git2go requires libgit2 1.9.x headers and rejects builds with the experimental SHA-256 API enabled. SHA-256 repositories are not supported (the `Oid` type is a fixed 20-byte array).

## Setup and commands

Bundled builds require Git submodules, CMake, `pkg-config`, and a C compiler. On Debian or Ubuntu:

```sh
sudo apt-get install -y build-essential cmake pkg-config
git submodule update --init
```

Use the bundled static mode for routine development because it builds the expected libgit2 revision locally:

```sh
make test-static                                  # build bundled libgit2 and run all tests
make TEST_ARGS='--count=1 -v' test-static        # verbose test suite
make TEST_ARGS='--count=1 -run ^TestClone$' test-static  # run one test
make build-libgit2-static                         # build only bundled libgit2
make build-libgit2-static && go build -tags static ./...  # build the Go package
make install-static                               # install with bundled static libgit2
```

Other link modes used by CI:

```sh
make test-dynamic                                 # bundled dynamic libgit2
make test                                         # system dynamic libgit2 found by pkg-config
go test --count=1 --tags 'static,system_libgit2' ./...  # system static libgit2
```

Every `make test*` target runs `script/check-MakeGitError-thread-lock.go` before `go test`. Keep that check in the validation path.

Generated enum stringers are committed. Regenerate and check them with:

```sh
go install golang.org/x/tools/cmd/stringer@v0.48.0
make generate
git diff --exit-code
```

The repository has no dedicated lint target. Format changed Go files with `gofmt -w <files>`. Note that `go vet ./...` reports pre-existing `reflect.SliceHeader` / `unsafe.Pointer` warnings in `merge.go`, `message.go`, `rebase.go`, and `remote.go` that the default `go test` vet subset does not flag; these are inherited from upstream and not regressions.

## Architecture

### Go wrappers over libgit2

Files such as `repository.go`, `object.go`, `index.go`, `diff.go`, `remote.go`, and `worktree.go` group APIs by libgit2 subsystem. Their Go structs wrap opaque `C.git_*` pointers. Methods marshal Go values to C, call libgit2, convert results, and retain owning objects when a child pointer depends on them.

`wrapper.c` contains callback and function-pointer adapters that cgo cannot express directly. Build-tag files select one cgo link configuration:

- `Build_system_dynamic.go`: default system dynamic library (`!static`)
- `Build_system_static.go`: system static library (`static,system_libgit2`)
- `Build_bundled_static.go`: locally built bundled library (`static,!system_libgit2`)

All three enforce libgit2 1.9 at compile time and reject `GIT_EXPERIMENTAL_SHA256`.

### Resource ownership

Most C-backed wrappers provide `Free` and install a finalizer as a fallback. Preserve the existing ownership rules when adding wrappers:

- Child objects retain their `*Repository` where needed.
- Borrowed pointers use weak/non-owning wrappers and must not free C-owned memory.
- Calls that pass Go-owned wrappers or buffers into C often need `runtime.KeepAlive` after the C call.
- Temporary C strings, arrays, option payloads, and callback handles require matching cleanup paths.

Package initialization in `git.go` initializes libgit2 and callback registries. `Shutdown` is global and is valid only after all git2go objects have been released.

### Errors, OS threads, and callbacks

libgit2 stores its last error in thread-local state. Code that calls `MakeGitError` must keep the C call and error conversion on one OS thread with `runtime.LockOSThread`/`runtime.UnlockOSThread`. The checker invoked by every Makefile test target enforces this invariant.

Go callback values cannot be passed to C directly. `HandleList` in `handles.go` allocates stable opaque C pointers and maps them back to Go values. Callback setup functions track payloads, `wrapper.c` bridges to cgo-exported Go callbacks, and cleanup functions untrack payloads. Same-stack callbacks preserve the original Go error through an `errorTarget`; callbacks that may run on another stack convert failures through libgit2's callback error state.

Remote transports also use a separate remote-pointer registry so HTTP/SSH transport callbacks can recover the owning Go `Remote`. Bundled libgit2 disables native HTTPS and SSH, causing package initialization to register the Go-managed transports in `http.go`, `ssh.go`, and `transport.go`.

### libgit2 1.9 behavior changes

Two inherited libgit2 1.9 changes affect this binding's semantics:

- **Checkout zero value**: libgit2 1.9 made `GIT_CHECKOUT_SAFE` the zero value. `CheckoutOptions{}` therefore performs a safe checkout, not a dry run. Callers that need a dry run must set `Strategy: CheckoutNone`. The `CheckoutStrategy` constants map directly to the 1.9 C values.
- **`update_refs` callback**: `RemoteCallbacks.UpdateRefsCallback` is preferred over the deprecated `UpdateTipsCallback` and also receives the matching `*Refspec`. `_go_git_populate_remote_callbacks` only installs the `update_refs` C callback when the Go side sets `UpdateRefsCallback`.

Options structs are always initialized with the matching `git_*_options_init` (or a C helper for structs like `git_commit_create_options` that lack an init function) rather than hand-written version literals, so new libgit2 fields pick up official defaults.

### Tests and generated files

Tests live beside the implementation as `*_test.go`. Shared helpers in `git_test.go` create and seed temporary repositories. Some graph and tree tests use the checked-in bare fixture under `testdata/TestGitRepository.git`.

`TestMain` checks both callback registries for leaked handles before calling `Shutdown`; a registry leak therefore fails the suite during teardown. Network tests use a local smart-HTTP fixture (in-process `git-http-backend` CGI), a local file transport, and an in-process SSH server; no test depends on public network services. The CI race job excludes the HTTP/SSH transport tests because the Go-managed HTTP transport in `http.go` has a pre-existing data race (background request goroutine vs. stream reader) inherited from upstream.

`git.go` and `diff.go` contain `go:generate` directives. Their generated `*_string.go` files must remain synchronized with the enum definitions, and CI verifies that `make generate` leaves no diff.
