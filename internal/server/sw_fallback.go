package server

import _ "embed"

// fallbackServiceWorker is a tiny service worker embedded in the binary so the
// server can respond to /sw.js even when the frontend assets (and thus the
// Workbox-generated service worker) have not been built yet. It keeps tests and
// development environments functional without affecting production builds.
//
//go:embed sw_fallback.js
var fallbackServiceWorker []byte
