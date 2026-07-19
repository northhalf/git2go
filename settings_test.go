package git

import (
	"testing"
	"time"
)

type pathPair struct {
	Level ConfigLevel
	Path  string
}

func TestSearchPath(t *testing.T) {
	paths := []pathPair{
		pathPair{ConfigLevelSystem, "/tmp/system"},
		pathPair{ConfigLevelGlobal, "/tmp/global"},
		pathPair{ConfigLevelXDG, "/tmp/xdg"},
	}

	for _, pair := range paths {
		err := SetSearchPath(pair.Level, pair.Path)
		checkFatal(t, err)

		actual, err := SearchPath(pair.Level)
		checkFatal(t, err)

		if pair.Path != actual {
			t.Fatal("Search paths don't match")
		}
	}
}

func TestMmapSizes(t *testing.T) {
	size := 42 * 1024

	err := SetMwindowSize(size)
	checkFatal(t, err)

	actual, err := MwindowSize()
	if size != actual {
		t.Fatal("Sizes don't match")
	}

	err = SetMwindowMappedLimit(size)
	checkFatal(t, err)

	actual, err = MwindowMappedLimit()
	if size != actual {
		t.Fatal("Sizes don't match")
	}
}

func TestEnableCaching(t *testing.T) {
	err := EnableCaching(false)
	checkFatal(t, err)

	err = EnableCaching(true)
	checkFatal(t, err)
}

func TestEnableStrictHashVerification(t *testing.T) {
	err := EnableStrictHashVerification(false)
	checkFatal(t, err)

	err = EnableStrictHashVerification(true)
	checkFatal(t, err)
}

func TestEnableFsyncGitDir(t *testing.T) {
	err := EnableFsyncGitDir(false)
	checkFatal(t, err)

	err = EnableFsyncGitDir(true)
	checkFatal(t, err)
}

func TestCachedMemory(t *testing.T) {
	current, allowed, err := CachedMemory()
	checkFatal(t, err)

	if current < 0 {
		t.Fatal("current < 0")
	}

	if allowed < 0 {
		t.Fatal("allowed < 0")
	}
}

func TestUserAgent(t *testing.T) {
	original, err := UserAgent()
	checkFatal(t, err)
	defer func() { checkFatal(t, SetUserAgent(original)) }()

	const want = "git2go/v38 test"
	checkFatal(t, SetUserAgent(want))
	got, err := UserAgent()
	checkFatal(t, err)
	if got != want {
		t.Fatalf("UserAgent() = %q, want %q", got, want)
	}
}

func TestServerTimeouts(t *testing.T) {
	originalConnect, err := ServerConnectTimeout()
	checkFatal(t, err)
	defer func() { checkFatal(t, SetServerConnectTimeout(originalConnect)) }()
	originalServer, err := ServerTimeout()
	checkFatal(t, err)
	defer func() { checkFatal(t, SetServerTimeout(originalServer)) }()

	const wantConnect = 2500 * time.Millisecond
	checkFatal(t, SetServerConnectTimeout(wantConnect))
	gotConnect, err := ServerConnectTimeout()
	checkFatal(t, err)
	if gotConnect != wantConnect {
		t.Fatalf("ServerConnectTimeout() = %v, want %v", gotConnect, wantConnect)
	}

	const wantServer = 4500 * time.Millisecond
	checkFatal(t, SetServerTimeout(wantServer))
	gotServer, err := ServerTimeout()
	checkFatal(t, err)
	if gotServer != wantServer {
		t.Fatalf("ServerTimeout() = %v, want %v", gotServer, wantServer)
	}

	if err := SetServerTimeout(-time.Millisecond); err == nil {
		t.Fatal("SetServerTimeout accepted a negative duration")
	}
}

func TestFeatureBackend(t *testing.T) {
	if backend := FeatureBackend(FeatureSHA1); backend == "" {
		t.Fatal("FeatureBackend(FeatureSHA1) is empty")
	}
}

func TestSetCacheMaxSize(t *testing.T) {
	err := SetCacheMaxSize(0)
	checkFatal(t, err)

	err = SetCacheMaxSize(1024 * 1024)
	checkFatal(t, err)

	// revert to default 256MB
	err = SetCacheMaxSize(256 * 1024 * 1024)
	checkFatal(t, err)
}
