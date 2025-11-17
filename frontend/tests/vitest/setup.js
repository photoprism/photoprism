import { afterEach, vi } from "vitest";
import "@testing-library/jest-dom";
import { config } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import * as components from "vuetify/components";
import * as directives from "vuetify/directives";
import "vuetify/styles";

import clientConfig from "./config";
import { $config } from "app/session";

$config.setValues(clientConfig);

// Make config available in browser environment
window.__CONFIG__ = clientConfig;

// Create a proper Vuetify instance with all components and styles
const vuetify = createVuetify({
  components,
  directives,
  theme: {
    defaultTheme: "light",
  },
});

// Configure Vue Test Utils global configuration
config.global.mocks = {
  $gettext: (text) => text,
  $pgettext: (_ctx, text) => text,
  $isRtl: false,
  $config: {
    feature: () => true,
    get: () => false,
    getSettings: () => ({ features: { edit: true, favorites: true, download: true, archive: true } }),
    allow: () => true,
    featExperimental: () => false,
    featDevelop: () => false,
    values: {},
    dir: () => "ltr",
  },
  $event: {
    subscribe: () => "sub-id",
    subscribeOnce: () => "sub-id-once",
    unsubscribe: () => {},
    publish: () => {},
  },
  $view: {
    enter: () => {},
    leave: () => {},
    isActive: () => true,
  },
  $notify: { success: () => {}, error: () => {}, warn: () => {} },
  $fullscreen: {
    isSupported: () => true,
    isEnabled: () => false,
    request: () => Promise.resolve(),
    exit: () => Promise.resolve(),
  },
  $clipboard: { selection: [], has: () => false, toggle: () => {} },
  $util: {
    hasTouch: () => false,
    encodeHTML: (s) => s,
    sanitizeHtml: (s) => s,
    formatSeconds: (n) => String(n),
    formatRemainingSeconds: () => "0",
    videoFormat: () => "avc",
    videoFormatUrl: () => "/v.mp4",
    thumb: () => ({ src: "/t.jpg", w: 100, h: 100 }),
  },
  $api: { post: vi.fn(), delete: vi.fn(), get: vi.fn() },
};

config.global.plugins = [vuetify];

config.global.stubs = {
  transition: false,
};

config.global.directives = {
  tooltip: {
    mounted(el, binding) {
      el.setAttribute("data-tooltip", binding.value);
    },
  },
};

const originalMount = config.global.mount;
config.global.mount = function (component, options = {}) {
  options.global = options.global || {};
  options.global.config = options.global.config || {};
  options.global.config.globalProperties = options.global.config.globalProperties || {};
  options.global.config.globalProperties.$emit = vi.fn();

  // Add vuetify to all mount calls
  if (!options.global.plugins) {
    options.global.plugins = [vuetify];
  } else if (Array.isArray(options.global.plugins)) {
    options.global.plugins.push(vuetify);
  }

  return originalMount(component, options);
};

// Clean up after each test
afterEach(() => {
  vi.resetAllMocks();
});

// Export shared configuration
export { clientConfig };

export default {
  vuetify,
};
