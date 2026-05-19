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

// Prefix used for namespacing localStorage keys per namespace key.
const namespacePrefix = "pp:";
// Separator between namespace key and the key name.
const namespaceSeparator = ":";

// Legacy keys to clear alongside their namespaced equivalents on logout/reset.
const defaultLegacyClearKeys = new Set([
  "session",
  "session.token",
  "session.id",
  "session.data",
  "session.user",
  "session.provider",
  "session.scope",
  "session.error",
  "clipboard",
  "clipboard.photos",
  "clipboard.albums",
]);

// In-memory storage fallback for environments without Web Storage.
const memoryStorage = () => {
  const data = new Map();
  return {
    get length() {
      return data.size;
    },
    key(index) {
      if (index < 0 || index >= data.size) {
        return null;
      }
      return Array.from(data.keys())[index] || null;
    },
    getItem(key) {
      return data.has(key) ? data.get(key) : null;
    },
    setItem(key, value) {
      data.set(key, String(value));
    },
    removeItem(key) {
      data.delete(key);
    },
    clear() {
      data.clear();
    },
  };
};

// Normalizes a storage implementation or falls back to in-memory storage.
const ensureStorage = (storage) => {
  if (!storage || typeof storage.getItem !== "function") {
    return memoryStorage();
  }
  return storage;
};

// Normalizes a namespace key to a non-empty string.
const normalizeNamespaceKey = (key) => {
  if (key === null || key === undefined) {
    return "";
  }

  let value = typeof key === "string" ? key : String(key);

  if (!value || value === "null" || value === "undefined") {
    return "";
  }

  return value;
};

// Safely access the browser window if it exists.
const getWindow = () => {
  if (typeof window === "undefined") {
    return null;
  }
  return window;
};

// Builds the namespace prefix string for a given namespace key.
export const buildNamespace = (namespaceKey) => {
  const normalized = normalizeNamespaceKey(namespaceKey);
  const key = normalized || "root";
  return `${namespacePrefix}${key}${namespaceSeparator}`;
};

// Parses a namespaced key into its namespace key and suffix components.
export const parseNamespace = (key) => {
  if (!key || typeof key !== "string" || !key.startsWith(namespacePrefix)) {
    return null;
  }

  const rest = key.slice(namespacePrefix.length);
  const sep = rest.indexOf(namespaceSeparator);
  if (sep === -1) {
    return null;
  }

  const encoded = rest.slice(0, sep);
  const suffix = rest.slice(sep + 1);

  if (!encoded) {
    return null;
  }

  return { namespace: encoded, key: suffix };
};

// Storage wrapper that transparently prefixes keys with a namespace key.
export class NamespacedStorage {
  constructor(storage, namespaceKey, options = {}) {
    this.storage = ensureStorage(storage);
    this.namespaceKey = normalizeNamespaceKey(options.namespaceKey) || normalizeNamespaceKey(namespaceKey) || "root";
    this.prefix = buildNamespace(this.namespaceKey);
    this.allowLegacy = options.allowLegacy !== false;
    this.legacyClearKeys = options.legacyClearKeys || defaultLegacyClearKeys;
  }

  // Length of the underlying storage.
  get length() {
    return this.storage.length || 0;
  }

  // Expose the raw storage key method when available.
  key(index) {
    if (typeof this.storage.key !== "function") {
      return null;
    }
    return this.storage.key(index);
  }

  // Computes the fully namespaced key for the current base path.
  namespacedKey(key) {
    if (!key) {
      return this.prefix;
    }
    return this.prefix + key;
  }

  // Reads a value, falling back to legacy keys and migrating if needed.
  getItem(key) {
    if (!key) {
      return null;
    }

    const namespaced = this.storage.getItem(this.namespacedKey(key));

    if (namespaced !== null && namespaced !== undefined) {
      return namespaced;
    }

    if (!this.allowLegacy) {
      return null;
    }

    const legacy = this.storage.getItem(key);

    if (legacy !== null && legacy !== undefined) {
      this.storage.setItem(this.namespacedKey(key), legacy);
      return legacy;
    }

    return null;
  }

  // Stores a value under the namespaced key.
  setItem(key, value) {
    if (!key) {
      return;
    }
    this.storage.setItem(this.namespacedKey(key), value);
  }

  // Removes the namespaced key and optionally its legacy counterpart.
  removeItem(key, options = {}) {
    if (!key) {
      return;
    }

    this.storage.removeItem(this.namespacedKey(key));

    if (options.legacy !== false && this.allowLegacy && this.legacyClearKeys.has(key)) {
      this.storage.removeItem(key);
    }
  }

  // Reads a raw legacy value without namespacing or migration.
  getLegacyItem(key) {
    if (!key) {
      return null;
    }

    return this.storage.getItem(key);
  }

  // Removes a raw legacy key without touching namespaced values.
  removeLegacyItem(key) {
    if (!key) {
      return;
    }

    this.storage.removeItem(key);
  }

  // Clears only keys in the current namespace.
  clear() {
    const keys = [];
    for (let i = 0; i < this.storage.length; i += 1) {
      const key = this.storage.key(i);
      if (key && key.startsWith(this.prefix)) {
        keys.push(key);
      }
    }
    keys.forEach((key) => this.storage.removeItem(key));
  }
}

// Creates a new namespaced storage wrapper for the given namespace key.
export const createNamespacedStorage = (storage, namespaceKey, options = {}) => {
  return new NamespacedStorage(storage, namespaceKey, options);
};

// Enumerates known auth sessions based on stored tokens and namespace keys.
export const listAuthSessions = (storage) => {
  const w = getWindow();
  const store = ensureStorage(storage || w?.localStorage);
  const sessions = [];
  const seen = new Set();

  for (let i = 0; i < store.length; i += 1) {
    const key = store.key(i);
    const parsed = parseNamespace(key);
    if (!parsed || parsed.key !== "session.token") {
      continue;
    }

    const token = store.getItem(key);
    if (!token) {
      continue;
    }

    if (!seen.has(parsed.namespace)) {
      sessions.push({ namespace: parsed.namespace, authToken: token });
      seen.add(parsed.namespace);
    }
  }

  if (!seen.has("/")) {
    const legacy = store.getItem("session.token");
    if (legacy) {
      sessions.push({ namespace: "/", authToken: legacy });
    }
  }

  return sessions;
};

let cachedAppStorage;
let cachedAppSessionStorage;

// Returns the app-local namespaced localStorage wrapper.
export const getAppStorage = () => {
  if (!cachedAppStorage) {
    const w = getWindow();
    cachedAppStorage = createNamespacedStorage(w?.localStorage, w?.__CONFIG__?.storageNamespace);
  }
  return cachedAppStorage;
};

// Returns the app-local namespaced sessionStorage wrapper.
export const getAppSessionStorage = () => {
  if (!cachedAppSessionStorage) {
    const w = getWindow();
    cachedAppSessionStorage = createNamespacedStorage(w?.sessionStorage, w?.__CONFIG__?.storageNamespace);
  }
  return cachedAppSessionStorage;
};
