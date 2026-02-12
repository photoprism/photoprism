// No-op fallback for service worker scope cleanup helper.
// Production builds ship a generated helper from frontend/src/sw-scope-cleanup.js.
self.addEventListener("activate", () => {});
