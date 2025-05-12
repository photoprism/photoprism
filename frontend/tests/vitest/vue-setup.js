import { config } from "@vue/test-utils";
import { vi } from "vitest";

// Mock Vuetify components
const vuetifyComponents = [
  "VBtn",
  "VToolbar",
  "VToolbarTitle",
  "VList",
  "VListItem",
  "VDivider",
  "VProgressCircular",
  "VIcon",
  "VRow",
  "VCol",
  "VCard",
  "VCardTitle",
  "VCardText",
  "VCardActions",
  "VTextField",
  "VTextarea",
  "VSheet",
];

// Create stubs for Vuetify components
const vuetifyStubs = vuetifyComponents.reduce((acc, component) => {
  acc[component] = {
    name: component.toLowerCase(),
    template: `<div data-testid="${component.toLowerCase()}" class="${component.toLowerCase()}-stub"><slot></slot></div>`,
  };
  acc[component.toLowerCase()] = {
    name: component.toLowerCase(),
    template: `<div data-testid="${component.toLowerCase()}" class="${component.toLowerCase()}-stub"><slot></slot></div>`,
  };
  return acc;
}, {});

// Configure Vue Test Utils global configuration
config.global.mocks = {
  $gettext: (text) => text,
  $isRtl: false,
  $config: {
    feature: (_name) => true,
  },
};

config.global.stubs = {
  ...vuetifyStubs,
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

  return originalMount(component, options);
};

export default {
  vuetifyStubs,
};
