package git

/*
#include <git2.h>

int _go_git_opts_get_search_path(int level, git_buf *buf)
{
    return git_libgit2_opts(GIT_OPT_GET_SEARCH_PATH, level, buf);
}

int _go_git_opts_set_search_path(int level, const char *path)
{
    return git_libgit2_opts(GIT_OPT_SET_SEARCH_PATH, level, path);
}

int _go_git_opts_set_size_t(int opt, size_t val)
{
    return git_libgit2_opts(opt, val);
}

int _go_git_opts_set_cache_object_limit(git_object_t type, size_t size)
{
    return git_libgit2_opts(GIT_OPT_SET_CACHE_OBJECT_LIMIT, type, size);
}

int _go_git_opts_get_size_t(int opt, size_t *val)
{
    return git_libgit2_opts(opt, val);
}

int _go_git_opts_get_size_t_size_t(int opt, size_t *val1, size_t *val2)
{
    return git_libgit2_opts(opt, val1, val2);
}

int _go_git_opts_get_buf(int opt, git_buf *buf)
{
    return git_libgit2_opts(opt, buf);
}

int _go_git_opts_set_string(int opt, const char *value)
{
    return git_libgit2_opts(opt, value);
}

int _go_git_opts_get_int(int opt, int *value)
{
    return git_libgit2_opts(opt, value);
}

int _go_git_opts_set_int(int opt, int value)
{
    return git_libgit2_opts(opt, value);
}
*/
import "C"
import (
	"errors"
	"math"
	"runtime"
	"time"
	"unsafe"
)

func SearchPath(level ConfigLevel) (string, error) {
	var buf C.git_buf
	defer C.git_buf_dispose(&buf)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := C._go_git_opts_get_search_path(C.int(level), &buf)
	if err < 0 {
		return "", MakeGitError(err)
	}

	return C.GoString(buf.ptr), nil
}

func SetSearchPath(level ConfigLevel, path string) error {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := C._go_git_opts_set_search_path(C.int(level), cpath)
	if err < 0 {
		return MakeGitError(err)
	}

	return nil
}

func UserAgent() (string, error) {
	return getString(C.GIT_OPT_GET_USER_AGENT)
}

func SetUserAgent(userAgent string) error {
	return setString(C.GIT_OPT_SET_USER_AGENT, userAgent)
}

func ServerConnectTimeout() (time.Duration, error) {
	return getDuration(C.GIT_OPT_GET_SERVER_CONNECT_TIMEOUT)
}

func SetServerConnectTimeout(timeout time.Duration) error {
	return setDuration(C.GIT_OPT_SET_SERVER_CONNECT_TIMEOUT, timeout)
}

func ServerTimeout() (time.Duration, error) {
	return getDuration(C.GIT_OPT_GET_SERVER_TIMEOUT)
}

func SetServerTimeout(timeout time.Duration) error {
	return setDuration(C.GIT_OPT_SET_SERVER_TIMEOUT, timeout)
}

func MwindowSize() (int, error) {
	return getSizet(C.GIT_OPT_GET_MWINDOW_SIZE)
}

func SetMwindowSize(size int) error {
	return setSizet(C.GIT_OPT_SET_MWINDOW_SIZE, size)
}

func MwindowMappedLimit() (int, error) {
	return getSizet(C.GIT_OPT_GET_MWINDOW_MAPPED_LIMIT)
}

func SetMwindowMappedLimit(size int) error {
	return setSizet(C.GIT_OPT_SET_MWINDOW_MAPPED_LIMIT, size)
}

func EnableCaching(enabled bool) error {
	if enabled {
		return setSizet(C.GIT_OPT_ENABLE_CACHING, 1)
	} else {
		return setSizet(C.GIT_OPT_ENABLE_CACHING, 0)
	}
}

func EnableStrictHashVerification(enabled bool) error {
	if enabled {
		return setSizet(C.GIT_OPT_ENABLE_STRICT_HASH_VERIFICATION, 1)
	} else {
		return setSizet(C.GIT_OPT_ENABLE_STRICT_HASH_VERIFICATION, 0)
	}
}

func EnableFsyncGitDir(enabled bool) error {
	if enabled {
		return setSizet(C.GIT_OPT_ENABLE_FSYNC_GITDIR, 1)
	} else {
		return setSizet(C.GIT_OPT_ENABLE_FSYNC_GITDIR, 0)
	}
}

func CachedMemory() (current int, allowed int, err error) {
	return getSizetSizet(C.GIT_OPT_GET_CACHED_MEMORY)
}

// deprecated: You should use `CachedMemory()` instead.
func GetCachedMemory() (current int, allowed int, err error) {
	return CachedMemory()
}

func SetCacheMaxSize(maxSize int) error {
	return setSizet(C.GIT_OPT_SET_CACHE_MAX_SIZE, maxSize)
}

func SetCacheObjectLimit(objectType ObjectType, size int) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := C._go_git_opts_set_cache_object_limit(C.git_object_t(objectType), C.size_t(size))
	if err < 0 {
		return MakeGitError(err)
	}

	return nil
}

func getString(opt C.int) (string, error) {
	var buf C.git_buf
	defer C.git_buf_dispose(&buf)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C._go_git_opts_get_buf(opt, &buf)
	if ret < 0 {
		return "", MakeGitError(ret)
	}
	return C.GoString(buf.ptr), nil
}

func setString(opt C.int, value string) error {
	cvalue := C.CString(value)
	defer C.free(unsafe.Pointer(cvalue))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C._go_git_opts_set_string(opt, cvalue)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func getDuration(opt C.int) (time.Duration, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var milliseconds C.int
	ret := C._go_git_opts_get_int(opt, &milliseconds)
	if ret < 0 {
		return 0, MakeGitError(ret)
	}
	return time.Duration(milliseconds) * time.Millisecond, nil
}

func setDuration(opt C.int, value time.Duration) error {
	if value < 0 {
		return errors.New("duration must not be negative")
	}
	milliseconds := value.Milliseconds()
	if milliseconds > math.MaxInt32 {
		return errors.New("duration exceeds libgit2's millisecond range")
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C._go_git_opts_set_int(opt, C.int(milliseconds))
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func getSizet(opt C.int) (int, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var val C.size_t
	err := C._go_git_opts_get_size_t(opt, &val)
	if err < 0 {
		return 0, MakeGitError(err)
	}

	return int(val), nil
}

func getSizetSizet(opt C.int) (int, int, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var val1, val2 C.size_t
	err := C._go_git_opts_get_size_t_size_t(opt, &val1, &val2)
	if err < 0 {
		return 0, 0, MakeGitError(err)
	}

	return int(val1), int(val2), nil
}

func setSizet(opt C.int, val int) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cval := C.size_t(val)
	err := C._go_git_opts_set_size_t(opt, cval)
	if err < 0 {
		return MakeGitError(err)
	}

	return nil
}
