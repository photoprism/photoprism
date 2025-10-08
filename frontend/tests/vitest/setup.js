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
  $isRtl: false,
  $config: {
    feature: (_name) => true,
  },
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
