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

// serviceWorkerScopeBase returns the registration scope derived from the app base URI.
export const serviceWorkerScopeBase = (baseUri) => {
  return baseUri ? baseUri.replace(/\/+$/, "") + "/" : "/";
};

// serviceWorkerUrl returns the registration URL for the current scope.
export const serviceWorkerUrl = (scopeBase) => {
  return `${scopeBase}sw.js`.replace(/\/\/+/g, "/");
};

// shouldCleanupRootScopeServiceWorker indicates if legacy root-scope workers should be removed.
export const shouldCleanupRootScopeServiceWorker = (scopeBase) => {
  if (typeof scopeBase !== "string" || !scopeBase.startsWith("/") || scopeBase === "/") {
    return false;
  }

  // Proxy-routed instance paths include at least two path segments (prefix + tenant),
  // e.g. "/<prefix>/<tenant>/" such as "/i/pro-1/" or "/instance/pro-1/".
  return scopeBase.split("/").filter(Boolean).length >= 2;
};

// isRootScopeRegistration checks whether a service worker registration controls the root scope.
export const isRootScopeRegistration = (registration) => {
  if (!registration || typeof registration.scope !== "string" || registration.scope === "") {
    return false;
  }

  if (registration.scope === "/") {
    return true;
  }

  try {
    return new URL(registration.scope).pathname === "/";
  } catch {
    return false;
  }
};

// cleanupLegacyRootScopeServiceWorkers unregisters root-scope workers for instance paths.
export const cleanupLegacyRootScopeServiceWorkers = (nav, scopeBase, log = console) => {
  if (!nav || !("serviceWorker" in nav)) {
    return Promise.resolve(false);
  }

  if (!shouldCleanupRootScopeServiceWorker(scopeBase)) {
    return Promise.resolve(false);
  }

  if (typeof nav.serviceWorker.getRegistrations !== "function") {
    return Promise.resolve(false);
  }

  return nav.serviceWorker
    .getRegistrations()
    .then((registrations) =>
      Promise.all(
        registrations
          .filter((registration) => isRootScopeRegistration(registration))
          .map((registration) =>
            registration.unregister().catch((err) => {
              if (typeof log?.warn === "function") {
                log.warn("service worker: root scope unregister failed", err);
              }

              return false;
            })
          )
      )
    )
    .then((results) => results.some(Boolean))
    .catch((err) => {
      if (typeof log?.warn === "function") {
        log.warn("service worker: root scope cleanup failed", err);
      }

      return false;
    });
};

// shouldRegisterServiceWorker determines whether service worker registration is safe.
export const shouldRegisterServiceWorker = (config) => {
  const scopeBase = serviceWorkerScopeBase(config?.baseUri);

  // Avoid root-scope service workers for the portal UI on shared domains.
  // Instances still register workers at /<prefix>/<tenant>/ where cache scopes stay isolated.
  return !(config?.values?.portal && scopeBase === "/");
};

// registerServiceWorker registers the PWA service worker when supported.
export const registerServiceWorker = (nav, config, log = console) => {
  if (!nav || !("serviceWorker" in nav)) {
    return Promise.resolve(false);
  }

  const scopeBase = serviceWorkerScopeBase(config?.baseUri);

  if (!shouldRegisterServiceWorker(config)) {
    if (typeof log?.debug === "function") {
      log.debug("service worker: skipped for portal root scope");
    }

    return Promise.resolve(false);
  }

  return cleanupLegacyRootScopeServiceWorkers(nav, scopeBase, log)
    .then(() => nav.serviceWorker.register(serviceWorkerUrl(scopeBase), { scope: scopeBase }))
    .then(() => true)
    .catch((err) => {
      if (typeof log?.warn === "function") {
        log.warn("service worker: register failed", err);
      }

      return false;
    });
};
