//go:build static && system_libgit2
// +build static,system_libgit2

package git

/*
#cgo pkg-config: libgit2 --static
#cgo CFLAGS: -DLIBGIT2_STATIC
#include <git2.h>

#if LIBGIT2_VER_MAJOR != 1 || LIBGIT2_VER_MINOR != 9
# error "Invalid libgit2 version; this git2go supports libgit2 v1.9.x"
#endif

#ifdef GIT_EXPERIMENTAL_SHA256
# error "git2go does not support libgit2's experimental SHA-256 API"
#endif
*/
import "C"
