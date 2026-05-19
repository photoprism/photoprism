/*

Copyright (c) 2018 - 2025 PhotoPrism UG. All rights reserved.

    This program is free software: you can redistribute it and/or modify
    it under Version 3 of the GNU Affero General Public License (the "AGPL"):
    <https://docs.photoprism.app/license/agpl>

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    The AGPL is supplemented by our Trademark and Brand Guidelines,
    which describe how our Brand Assets may be used:
    <https://www.photoprism.app/trademark>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>

*/

(() => {
  const WorkboxPrecachePrefix = "workbox-precache-";
  const WorkboxPrecacheCacheName = "workbox-precache-v2";

  // cleanupOutdatedCaches() in Workbox uses a broad scope match (`includes`),
  // which can remove instance caches when the Portal runs with root scope.
  // Keep cleanup strict to the exact registration scope instead.
  self.addEventListener("activate", (event) => {
    const scope = self && self.registration ? self.registration.scope : "";

    if (!scope || typeof caches === "undefined") {
      return;
    }

    const currentPrecache = `${WorkboxPrecacheCacheName}-${scope}`;

    event.waitUntil(
      caches.keys().then((cacheNames) =>
        Promise.all(
          cacheNames
            .filter((cacheName) => cacheName.startsWith(WorkboxPrecachePrefix))
            .filter((cacheName) => cacheName.endsWith(scope))
            .filter((cacheName) => cacheName !== currentPrecache)
            .map((cacheName) => caches.delete(cacheName))
        )
      )
    );
  });
})();
