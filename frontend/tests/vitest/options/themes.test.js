import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { createVuetify } from "vuetify";

import * as themes from "../../../src/options/themes";

// Custom themes as sites define them in app.js for Vuetify 3, shortened: a dark one in the
// window.__THEMES__ format with its own variables, and a light one without variables.
const customDark = {
  dark: true,
  force: false,
  title: "Custom Dark",
  name: "custom-dark",
  colors: {
    "background": "#2c2d2f",
    "surface": "#161718",
    "on-surface": "#ffffff",
    "surface-variant": "#7E4FE3",
    "primary": "#9E7BEA",
    "highlight": "#5F1DB7",
    "secondary": "#191A1C",
    "accent": "#2D2E2E",
    "navigation": "#141417",
  },
  variables: {
    "btn-height": "34px",
    "border-color": "#FFFFFF",
    "border-opacity": 0.05,
    "hover-opacity": 0.019,
    "overlay-color": "#131313",
    "overlay-opacity": 0.54,
    "icon": "logo.svg",
  },
};

const customLight = {
  dark: false,
  force: false,
  title: "Custom Light",
  name: "custom-light",
  colors: {
    background: "#f9f9f9",
    surface: "#f9f9f9",
    card: "#222222",
    primary: "#788195",
    accent: "#222222",
    navigation: "#ffffff",
  },
};

describe("options/themes", () => {
  beforeEach(() => {
    themes.SetOptions([
      {
        text: "Default",
        value: "default",
        disabled: false,
      },
    ]);

    themes.Set("default", {
      name: "default",
      title: "Default",
      colors: {},
      variables: {},
    });
  });

  it("assigns multiple themes and enforces forced theme", () => {
    const forcedTheme = {
      name: "forced",
      title: "Forced",
      force: true,
      colors: { background: "#000000" },
      variables: {},
    };

    const optionalTheme = {
      name: "optional",
      title: "Optional",
      colors: { background: "#ffffff" },
      variables: {},
    };

    themes.Assign([optionalTheme, forcedTheme]);

    const available = themes.Options();
    expect(available).toHaveLength(1);
    expect(available[0].value).toBe("forced");

    const forced = themes.Get("default", true);
    expect(forced.name).toBe("forced");
    expect(forced.colors.background).toBe("#000000");
  });

  it("does not add duplicate entries when Assign is called multiple times", () => {
    const theme = {
      name: "example",
      title: "Example",
      colors: { background: "#123456" },
      variables: {},
    };

    themes.Assign([theme]);
    themes.Assign([theme]);

    const available = themes.Options();
    expect(available.findIndex((option) => option.value === "example")).toBeGreaterThan(-1);
    const filtered = available.filter((option) => option?.value === "example");
    expect(filtered).toHaveLength(1);
  });

  it("returns all themes as a plain object keyed by name", () => {
    const all = themes.All();
    expect(Array.isArray(all)).toBe(false);
    expect(Object.getPrototypeOf(all)).toBe(Object.prototype);
    expect(all.default).toBeDefined();
    expect(all.default.dark).toBe(false);
    expect(Object.keys(all.default.colors).length).toBeGreaterThan(0);
  });

  afterEach(() => {
    themes.Remove(customDark.name);
    themes.Remove(customLight.name);
  });

  it("fills in what Vuetify 4 adds for custom themes written for Vuetify 3", () => {
    themes.Assign([customDark]);
    themes.Set(customLight.name, customLight);

    const dark = themes.Get("custom-dark");
    expect(dark.dark).toBe(true);
    expect(dark.colors.background).toBe("#2c2d2f");
    expect(dark.variables["btn-height"]).toBe("34px");
    expect(dark.variables["theme-on-dark"]).toBeDefined();
    expect(dark.variables["theme-on-light"]).toBeDefined();

    const light = themes.Get("custom-light");
    expect(light.dark).toBe(false);
    expect(light.colors.primary).toBe("#788195");
    expect(light.variables["theme-on-dark"]).toBeDefined();
    expect(light.variables["theme-on-light"]).toBeDefined();
  });

  it("builds Vuetify 4 themes from custom themes written for Vuetify 3", () => {
    themes.Assign([customDark]);
    themes.Set(customLight.name, customLight);

    const vuetify = createVuetify({ theme: { defaultTheme: "custom-dark", themes: themes.All(), variations: themes.variations } });
    const computed = vuetify.theme.computedThemes.value;

    // Vuetify derives each on-* color a theme leaves out; building the style sheet parses every color.
    expect(computed["custom-dark"].colors["on-primary"]).toMatch(/^#([0-9a-f]{3}|[0-9a-f]{6})$/i);
    expect(computed["custom-light"].colors["on-background"]).toMatch(/^#([0-9a-f]{3}|[0-9a-f]{6})$/i);
    expect(computed["custom-light"].colors["on-card"]).toMatch(/^#([0-9a-f]{3}|[0-9a-f]{6})$/i);
    expect(computed["custom-dark"].colors["on-surface"]).toBe("#ffffff");
    expect(vuetify.theme.styles.value).toContain(".v-theme--custom-light");
  });
});
