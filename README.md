git2go
======

Go bindings for [libgit2](http://libgit2.github.com/).

This is a maintained fork that tracks libgit2 1.9.x. The upstream
`libgit2/git2go` repository has not been updated since 2022 and does not support
libgit2 beyond 1.5. This fork publishes a new Go major version, `/v38`, so
existing `/v34` consumers can opt in without a forced upgrade.

### Version mapping

| libgit2 | git2go (this fork) |
|---------|--------------------|
| 1.9     | v38                |

Import it with the major version suffix:

```sh
go get github.com/northhalf/git2go/v38
```
```go
import "github.com/northhalf/git2go/v38"
```

The `main` branch tracks libgit2 1.9 and only supports statically linking
against the bundled libgit2 for routine development.

### Requirements

- Go 1.25 or newer (tested against Go 1.25 and 1.26).
- libgit2 1.9.x. The vendored submodule is pinned to v1.9.6.
- For bundled builds: CMake, `pkg-config`, a C compiler, and Git submodules.
- SHA-256 repositories are **not** supported. This build uses a fixed 20-byte
  `Oid` and rejects libgit2 builds with the experimental SHA-256 API enabled.

### Migrating from `/v34`

This fork keeps the existing git2go API source-compatible where possible, with
two behavior changes to be aware of:

- **Checkout zero value**: libgit2 1.9 made `GIT_CHECKOUT_SAFE` the zero value.
  In `/v38`, `CheckoutOptions{}` performs a safe checkout instead of a dry run.
  Code that relied on the zero value being a dry run must set
  `Strategy: CheckoutNone` explicitly.
- **`UpdateTipsCallback` is deprecated**: libgit2 1.9 prefers the new
  `update_refs` callback, exposed as `UpdateRefsCallback` (which also receives
  the matching refspec). `UpdateTipsCallback` still works.

New APIs cover worktrees, staged commits, commit parents, remote push options,
fetch depth, redirect policy, remote OID type, User-Agent, and server timeouts.

Installing
----------

This project wraps the functionality provided by libgit2. It thus needs it in order to perform the work.

This project wraps the functionality provided by libgit2. If you're using a versioned branch, install it to your system via your system's package manager and then install git2go.


### Versioned branch, dynamic linking

When linking dynamically against a released version of libgit2, install it via your system's package manager. CGo will take care of finding its pkg-config file and set up the linking. Import via Go modules:

```go
import "github.com/northhalf/git2go/v38"
```

### Versioned branch, static linking

Follow the instructions for [Versioned branch, dynamic linking](#versioned-branch-dynamic-linking), but pass the `-tags static,system_libgit2` flag to all `go` commands that build any binaries. For instance:

    go build -tags static,system_libgit2 github.com/my/project/...
    go test -tags static,system_libgit2 github.com/my/project/...
    go install -tags static,system_libgit2 github.com/my/project/...

### `main` branch, or vendored static linking

If using `main` or building a branch with the vendored libgit2 statically, we need to build libgit2 first. In order to build it, you need `cmake`, `pkg-config` and a C compiler. You will also need the development packages for OpenSSL (outside of Windows or macOS) and LibSSH2 installed if you want libgit2 to support HTTPS and SSH respectively. Note that even if libgit2 is included in the resulting binary, its dependencies will not be.

Run `go get -d github.com/northhalf/git2go/v38` to download the code. From there, we need to build the C code and put it into the resulting go binary.

    git submodule update --init # get libgit2
    make install-static

will compile libgit2, link it into git2go and install it. The `main` branch is set up to follow the specific libgit2 version that is vendored, so trying dynamic linking may or may not work depending on the exact versions involved.

In order to let Go pass the correct flags to `pkg-config`, `-tags static` needs to be passed to all `go` commands that build any binaries. For instance:

    go build -tags static github.com/my/project/...
    go test -tags static github.com/my/project/...
    go install -tags static github.com/my/project/...

One thing to take into account is that since Go expects the `pkg-config` file to be within the same directory where `make install-static` was called, so the `go.mod` file may need to have a [`replace` directive](https://github.com/golang/go/wiki/Modules#when-should-i-use-the-replace-directive) so that the correct setup is achieved. So if `git2go` is checked out at `$GOPATH/src/github.com/northhalf/git2go` and your project at `$GOPATH/src/github.com/my/project`, the `go.mod` file of `github.com/my/project` might need to have a line like

    replace github.com/northhalf/git2go/v38 => ../../northhalf/git2go

Parallelism and network operations
----------------------------------

libgit2 may use OpenSSL and LibSSH2 for performing encrypted network connections. For now, git2go asks libgit2 to set locking for OpenSSL. This makes HTTPS connections thread-safe, but it is fragile and will likely stop doing it soon. This may also make SSH connections thread-safe if your copy of libssh2 is linked against OpenSSL. Check libgit2's `THREADSAFE.md` for more information.

Running the tests
-----------------

For the stable version, `go test` will work as usual. For the `main` branch, similarly to installing, running the tests requires building a local libgit2 library, so the Makefile provides a wrapper that makes sure it's built

    make test-static

Alternatively, you can build the library manually first and then run the tests

    make install-static
    go test -v -tags static ./...

License
-------

M to the I to the T. See the LICENSE file if you've never seen an MIT license before.

Authors
-------

- Carlos Martín (@carlosmn)
- Vicent Martí (@vmg)

