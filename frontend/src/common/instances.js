/*

Copyright (c) 2018 - 2026 PhotoPrism UG. All rights reserved.

    This program is free software: you can redistribute it and/or modify
    it under Version 3 of the GNU Affero General Public License (the "AGPL"):
    <https://docs.photoprism.app/license/agpl>

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    The AGPL is supplemented by our Trademark and Brand Guidelines,
    which describe how our Brand Assets may be used:
    <https://www.photoprism.app/trademark/>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>

*/

import { listAuthSessions, buildNamespace, createNamespacedStorage } from "common/storage";

// Storage suffixes under which each instance records its public identity, so
// peers on the same shared-domain origin can render a navigation switcher entry.
// The namespace is a SHA-256 hash of the SiteUrl and is not reversible, so the
// URL and title must be persisted explicitly under the instance's namespace.
const InstanceUrlKey = "instance.url";
const InstanceTitleKey = "instance.title";
const InstanceIconKey = "instance.icon";
const InstanceRouteKey = "instance.route";
const InstancePortalKey = "instance.portal";

// InstanceIdentityKeys lists the suffix keys written by persistInstanceIdentity,
// so callers (e.g. session logout) can clear them across storage backends.
export const InstanceIdentityKeys = [InstanceUrlKey, InstanceTitleKey, InstanceIconKey, InstanceRouteKey, InstancePortalKey];

// safeWindow returns the browser window if available, else null.
const safeWindow = () => (typeof window === "undefined" ? null : window);

// isHttpUrl reports whether url is an absolute http(s) URL. Instance URLs are
// read from shared same-origin storage written by peer instances, so the
// switcher rejects any other scheme (javascript:, data:, …) before listing or
// navigating to them.
function isHttpUrl(url) {
  if (!url || typeof url !== "string") {
    return false;
  }
  try {
    const protocol = new URL(url).protocol;
    return protocol === "http:" || protocol === "https:";
  } catch {
    return false;
  }
}

// instanceLabel derives a short, distinctive display name from a SiteUrl — the
// last base-path segment (e.g. "pro-1" for ".../i/pro-1/"). The switcher can
// only surface same-origin instances, which always differ by path, so the path
// segment is more distinctive than the frequently-generic site caption/title
// (multiple instances commonly share the default "PhotoPrism" caption). Returns
// "" for a root-path or unparseable URL so the caller falls back to the title.
export function instanceLabel(siteUrl) {
  if (!siteUrl || typeof siteUrl !== "string") {
    return "";
  }
  try {
    const segments = new URL(siteUrl).pathname.split("/").filter(Boolean);
    return segments.length ? segments[segments.length - 1] : "";
  } catch {
    return "";
  }
}

// instanceTitle derives the switcher label from an instance's client config values,
// preferring the operator-configured site name so a distinctively branded instance
// shows its real name (e.g. "ACME") instead of the lowercase base-path slug. The
// backend siteName (Config.SiteName) resolves SITE_NAME → AppName → SiteTitle and is
// empty for an unbranded instance, so we then fall back to the distinctive base-path
// segment (instanceLabel) — which the menu still shows as a subtitle to keep peers
// that share a generic name distinguishable. Returns "" only when no field is usable.
export function instanceTitle(values) {
  if (!values || typeof values !== "object") {
    return "";
  }
  return values.siteName || instanceLabel(values.siteUrl) || values.siteTitle || values.name || values.siteUrl || "";
}

// instancePath returns the base path of a SiteUrl (e.g. "/i/pro-1") so the
// switcher can show how same-origin peers differ without repeating the shared
// origin. Returns "/" for a root install and "" for an unparseable URL.
export function instancePath(siteUrl) {
  if (!siteUrl || typeof siteUrl !== "string") {
    return "";
  }
  try {
    const path = new URL(siteUrl).pathname.replace(/\/+$/, "");
    return path || "/";
  } catch {
    return "";
  }
}

// persistInstanceIdentity records this instance's SiteUrl, display title, app icon,
// and frontend route in the given (namespaced) store, so other instances on the same
// origin can list it in the navigation instance switcher and open it at its app entry
// point. The route is the frontend URI (e.g. "/portal" or "/i/pro-1/library");
// the switcher opens it so a web-overlay landing page at the site root is bypassed.
// No-op without a URL or usable store.
export function persistInstanceIdentity(store, identity) {
  if (!store || typeof store.setItem !== "function" || !identity || !identity.url) {
    return;
  }

  store.setItem(InstanceUrlKey, identity.url);

  if (identity.title) {
    store.setItem(InstanceTitleKey, identity.title);
  } else {
    store.removeItem(InstanceTitleKey);
  }

  if (identity.icon) {
    store.setItem(InstanceIconKey, identity.icon);
  } else {
    store.removeItem(InstanceIconKey);
  }

  if (identity.route) {
    store.setItem(InstanceRouteKey, identity.route);
  } else {
    store.removeItem(InstanceRouteKey);
  }

  // Mark the Portal's own session; Sign-Out delegates it to the Portal end-session endpoint.
  if (identity.portal) {
    store.setItem(InstancePortalKey, "true");
  } else {
    store.removeItem(InstancePortalKey);
  }
}

// listReachableInstances returns the instances (other than currentNamespace) that
// have a live session token and a recorded identity in shared browser storage,
// for the navigation instance switcher. Both localStorage (persistent sessions)
// and sessionStorage (ephemeral sessions, shared across same-tab navigations) are
// scanned, since recordInstanceIdentity writes to whichever the instance uses.
// Results are ordered so the Portal (origin root path) leads, then co-located peers
// by ascending base-path. Returns an empty array on standalone or subdomain-isolated
// deployments where no peer sessions are discoverable.
export function listReachableInstances(options) {
  const opts = options || {};

  let stores;
  if (opts.storage) {
    stores = [opts.storage];
  } else if (Array.isArray(opts.stores)) {
    stores = opts.stores;
  } else {
    const w = safeWindow();
    stores = [w?.localStorage, w?.sessionStorage];
  }

  const currentPrefix = buildNamespace(opts.currentNamespace);
  const seen = new Set();
  const instances = [];

  stores.forEach((store) => {
    if (!store || typeof store.getItem !== "function") {
      return;
    }

    listAuthSessions(store).forEach((session) => {
      const namespace = session && session.namespace;
      if (!namespace) {
        return;
      }

      const prefix = buildNamespace(namespace);
      if (prefix === currentPrefix || seen.has(prefix)) {
        return;
      }

      const url = store.getItem(prefix + InstanceUrlKey);
      if (!isHttpUrl(url)) {
        return;
      }

      // route is the app-entry URL to navigate to (the SiteUrl origin + the peer's
      // frontend URI). Resolve the stored route against the SiteUrl and reject any
      // non-http(s) result; fall back to the SiteUrl for legacy/empty entries.
      const storedRoute = store.getItem(prefix + InstanceRouteKey);
      let route = url;
      if (storedRoute) {
        try {
          const resolved = new URL(storedRoute, url).href;
          if (isHttpUrl(resolved)) {
            route = resolved;
          }
        } catch {
          // keep the SiteUrl when the stored route can't be resolved.
        }
      }

      seen.add(prefix);
      instances.push({
        namespace,
        url,
        route,
        title: store.getItem(prefix + InstanceTitleKey) || url,
        icon: store.getItem(prefix + InstanceIconKey) || "",
      });
    });
  });

  // Order the switcher so the Portal (origin root path "/") leads, then peers by
  // ascending base-path length: co-located instances live under distinct base paths
  // (e.g. "/i/pro-1"), so the shortest path is always the Portal at the origin root.
  instances.sort((a, b) => {
    const pa = instancePath(a.url) || "/";
    const pb = instancePath(b.url) || "/";
    return pa.length - pb.length || pa.localeCompare(pb);
  });

  return instances;
}

// instanceSessionUrl returns the absolute DELETE-session endpoint for a peer,
// derived from its recorded SiteUrl (<SiteUrl>api/v1/session), or "" when the
// SiteUrl is missing or not http(s). Same-origin by design (shared-domain proxy).
export function instanceSessionUrl(siteUrl) {
  if (!isHttpUrl(siteUrl)) {
    return "";
  }
  try {
    const base = siteUrl.endsWith("/") ? siteUrl : siteUrl + "/";
    const resolved = new URL("api/v1/session", base).href;
    return isHttpUrl(resolved) ? resolved : "";
  } catch {
    return "";
  }
}

// listLogoutTargets returns the peer sessions a cluster-wide Sign-Out should revoke
// server-side — {namespace, authToken, url} each, excluding currentNamespace (signed
// out via the normal path). Scans both storage backends. A peer with an unresolvable
// URL is still returned (url === "") so its local storage is still cleared.
export function listLogoutTargets(options) {
  const opts = options || {};

  let stores;
  if (opts.storage) {
    stores = [opts.storage];
  } else if (Array.isArray(opts.stores)) {
    stores = opts.stores;
  } else {
    const w = safeWindow();
    stores = [w?.localStorage, w?.sessionStorage];
  }

  const currentPrefix = buildNamespace(opts.currentNamespace);
  const seen = new Set();
  const targets = [];

  stores.forEach((store) => {
    if (!store || typeof store.getItem !== "function") {
      return;
    }

    listAuthSessions(store).forEach((session) => {
      const namespace = session && session.namespace;
      const authToken = session && session.authToken;
      if (!namespace || !authToken) {
        return;
      }

      const prefix = buildNamespace(namespace);
      if (prefix === currentPrefix || seen.has(prefix)) {
        return;
      }

      seen.add(prefix);
      targets.push({
        namespace,
        authToken,
        url: instanceSessionUrl(store.getItem(prefix + InstanceUrlKey)),
        // True for the cluster Portal's own session (delegated to its end-session endpoint).
        portal: store.getItem(prefix + InstancePortalKey) === "true",
      });
    });
  });

  return targets;
}

// signOutInstances best-effort revokes each target's session via a DELETE
// authenticated with the target's own token. Same-origin only; a target without a
// URL is skipped and failures (unreachable, 401) are swallowed so one bad peer
// can't block Sign-Out. Resolves once all settle; never rejects.
export function signOutInstances(targets, fetchImpl) {
  const doFetch = fetchImpl || (typeof fetch === "function" ? fetch.bind(safeWindow() || undefined) : null);

  if (!doFetch || !Array.isArray(targets) || targets.length === 0) {
    return Promise.resolve([]);
  }

  return Promise.allSettled(
    targets
      .filter((t) => t && t.url && t.authToken)
      .map((t) =>
        doFetch(t.url, {
          method: "DELETE",
          // Authenticate with the peer's own token; same-origin so the response
          // Set-Cookie (clearing the Portal OP cookie) is honored.
          headers: { "X-Auth-Token": t.authToken },
          credentials: "same-origin",
          cache: "no-store",
        }).catch(() => {})
      )
  );
}

// clearInstanceStorage removes every namespaced key for each of namespaces from the
// given raw storage backends, so a cluster-wide Sign-Out leaves no peer tokens or
// stale switcher entries. Each namespace is cleared via a NamespacedStorage wrapper.
export function clearInstanceStorage(namespaces, stores) {
  if (!Array.isArray(namespaces) || namespaces.length === 0) {
    return;
  }

  const backends = (Array.isArray(stores) ? stores : [stores]).filter((s) => s && typeof s.removeItem === "function");

  namespaces.forEach((namespace) => {
    if (!namespace) {
      return;
    }
    backends.forEach((store) => createNamespacedStorage(store, namespace).clear());
  });
}
