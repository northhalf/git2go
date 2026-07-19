package git

/*
#include <git2.h>
*/
import "C"

type Feature int

const (
	// libgit2 was built with threading support
	FeatureThreads Feature = C.GIT_FEATURE_THREADS

	// libgit2 was built with HTTPS support built-in
	FeatureHTTPS Feature = C.GIT_FEATURE_HTTPS

	// libgit2 was build with SSH support built-in
	FeatureSSH Feature = C.GIT_FEATURE_SSH

	// libgit2 was built with nanosecond support for files
	FeatureNSec Feature = C.GIT_FEATURE_NSEC

	FeatureHTTPParser  Feature = C.GIT_FEATURE_HTTP_PARSER
	FeatureRegex       Feature = C.GIT_FEATURE_REGEX
	FeatureI18N        Feature = C.GIT_FEATURE_I18N
	FeatureAuthNTLM    Feature = C.GIT_FEATURE_AUTH_NTLM
	FeatureNegotiate   Feature = C.GIT_FEATURE_AUTH_NEGOTIATE
	FeatureCompression Feature = C.GIT_FEATURE_COMPRESSION
	FeatureSHA1        Feature = C.GIT_FEATURE_SHA1
)

// Features returns a bit-flag of Feature values indicating which features the
// loaded libgit2 library has.
func Features() Feature {
	features := C.git_libgit2_features()

	return Feature(features)
}

// FeatureBackend returns the backend name for a compiled libgit2 feature.
func FeatureBackend(feature Feature) string {
	backend := C.git_libgit2_feature_backend(C.git_feature_t(feature))
	if backend == nil {
		return ""
	}
	return C.GoString(backend)
}
