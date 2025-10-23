// Minimal fallback service worker served during development and tests.
// Production builds replace this handler with the Workbox-generated sw.js.
self.addEventListener("install", (event) => {
  self.skipWaiting();
});

self.addEventListener("activate", (event) => {
  event.waitUntil(self.clients.claim());
});

self.addEventListener("fetch", () => {
  // Default network-first behaviour; caching is delegated to the
  // Workbox-generated service worker in production builds.
});
