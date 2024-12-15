/* Dark Theme Presets */

export const dark = "dark";

export const colorsDark = {
  background: "#2c2d2f",
  surface: "#161718",
  "on-surface": "#ffffff",
  "surface-bright": "#333333",
  "surface-variant": "#7E4FE3",
  "on-surface-variant": "#f6f7e8",
  card: "#171718",
  table: "#1F2022", // Variations: 242628, 212325, 1E2022, 1C1D1F, 191A1C, 161718, 131415, 111112
  button: "#1D1E1F",
  primary: "#9E7BEA",
  highlight: "#5F1DB7",
  selected: "#4d4d4e",
  secondary: "#191A1C",
  "secondary-light": "#1E2022",
  accent: "#2D2E2E",
  error: "#e57373",
  info: "#00acc1",
  success: "#4db6ac",
  warning: "#ffd740",
  favorite: "#FFD600",
  remove: "#da4e4c",
  restore: "#00d48a",
  album: "#ed9e00",
  download: "#00bfa5",
  private: "#00b8d4",
  edit: "#2196F3",
  share: "#3F51B5",
  love: "#ef5350",
  terminal: "#4A464F",
  navigation: "#141417",
  "navigation-home": "#0e0f10",
};

export const variablesDark = {
  "btn-height": "32px",
  "table-row-height": "44px",
  "table-header-height": "44px",
  "border-color": "#FFFFFF",
  "border-opacity": 0.05,
  "high-emphasis-opacity": 0.96,
  "medium-emphasis-opacity": 0.88,
  "label-opacity": 0.67,
  "disabled-opacity": 0.75,
  "idle-opacity": 0.1,
  "hover-opacity": 0.019,
  "focus-opacity": 0.022,
  "selected-opacity": 0.08,
  "activated-opacity": 0,
  "pressed-opacity": 0.16,
  "dragged-opacity": 0.08,
  "overlay-color": "#121212",
  "overlay-opacity": 0.42,
  "theme-kbd": "#212529",
  "theme-on-kbd": "#FFFFFF",
  "theme-code": "#343434",
  "theme-on-code": "#CCCCCC",
};

/* Light Theme Presets */

export const light = "light";

export const colorsLight = {
  background: "#f7f8fa",
  surface: "#e4e8f0",
  "on-surface": "#1e1e1f",
  "surface-bright": "#cbced6",
  "surface-variant": "#878787",
  "on-surface-variant": "#f6f7e8",
  card: "#a8a8a8",
  button: "#474b4d",
  table: "#dddcda",
  primary: "#353839",
  highlight: "#3d3f40",
  selected: "#c3c3c3",
  secondary: "#e2e7ee",
  "secondary-light": "#eceff4",
  accent: "#F2F5FA",
  error: "#BF616A",
  info: "#88C0D0",
  success: "#8FBCBB",
  warning: "#f0d8a8",
  favorite: "#EBCB8B",
  remove: "#BF616A",
  restore: "#81A1C1",
  album: "#EBCB8B",
  download: "#8FBCBB",
  private: "#88C0D0",
  edit: "#88C0D0",
  share: "#B48EAD",
  love: "#ef5350",
  terminal: "#4C566A",
  navigation: "#e7ebf1",
  "navigation-home": "#dde3eb",
};

export const variablesLight = {
  ...variablesDark,
  ...{
    "overlay-color": "#e4e8f0",
    "border-color": "#ffffff",
    "border-opacity": 0.08,
    "high-emphasis-opacity": 0.96,
    "medium-emphasis-opacity": 0.7,
    "hover-opacity": 0.08,
    "focus-opacity": 0.1,
  },
};

/* Export Theme Style Presets */

export function style(theme) {
  if (typeof theme !== "object") {
    return dark;
  }

  return theme.dark ? dark : light;
}

export const colors = {
  dark: colorsDark,
  light: colorsLight,
};

export const variables = {
  dark: variablesDark,
  light: variablesLight,
};
