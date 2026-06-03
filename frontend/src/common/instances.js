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
    <https://www.photoprism.app/trademark>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>

*/

import { listAuthSessions, buildNamespace } from "common/storage";

// Storage suffixes under which each instance records its public identity, so
// peers on the same shared-domain origin can render a navigation switcher entry.
// The namespace is a SHA-256 hash of the SiteUrl and is not reversible, so the
// URL and title must be persisted explicitly under the instance's namespace.
const InstanceUrlKey = "instance.url";
const InstanceTitleKey = "instance.title";

// InstanceIdentityKeys lists the suffix keys written by persistInstanceIdentity,
// so callers (e.g. session logout) can clear them across storage backends.
export const InstanceIdentityKeys = [InstanceUrlKey, InstanceTitleKey];

// safeWindow returns the browser window if available, else null.
const safeWindow = () => (typeof window === "undefined" ? null : window);

// basePathSegment returns the last non-empty path segment of a URL (e.g.
// "https://app.example.com/i/pro-1/" -> "pro-1"), or "" when the URL is invalid
// or root-pathed.
function basePathSegment(url) {
  if (typeof url !== "string" || url === "") {
    return "";
  }
  let pathname;
  try {
    pathname = new URL(url, safeWindow()?.location?.origin || "http://localhost/").pathname;
  } catch {
    return "";
  }
  const segments = pathname.split("/").filter((s) => s !== "");
  return segments.length ? segments[segments.length - 1] : "";
}

// instanceLabel derives a distinctive switcher label for an instance: its base
// path segment when present (same-origin instances differ only by path and
// commonly share a generic title like "PhotoPrism"), else the recorded title,
// else the URL.
export function instanceLabel(url, title) {
  return basePathSegment(url) || title || url || "";
}

// persistInstanceIdentity records this instance's SiteUrl and display title in
// the given (namespaced) store, so other instances on the same origin can list
// it in the navigation instance switcher. No-op without a URL or usable store.
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
}

// listReachableInstances returns the instances (other than currentNamespace) that
// have a live session token and a recorded identity in shared browser storage,
// for the navigation instance switcher. Both localStorage (persistent sessions)
// and sessionStorage (ephemeral sessions, shared across same-tab navigations) are
// scanned, since recordInstanceIdentity writes to whichever the instance uses.
// Returns an empty array on standalone or subdomain-isolated deployments where no
// peer sessions are discoverable.
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
      if (!url) {
        return;
      }

      seen.add(prefix);
      instances.push({
        namespace,
        url,
        title: instanceLabel(url, store.getItem(prefix + InstanceTitleKey)),
      });
    });
  });

  return instances;
}
