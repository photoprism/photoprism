package server

import _ "embed"

// fallbackScopeCleanupScript is served when the generated service worker helper
// is unavailable in local development or test builds.
//
//go:embed sw_scope_cleanup_fallback.js
var fallbackScopeCleanupScript []byte
