import { $gettext, T } from "common/gettext";
import { style, colors, variables } from "ui";

/* Theme Definitions */

// TODO: Make sure all themes have a button and table color.
let themes = {
  /* Default user interface theme */
  default: {
    dark: true,
    title: "Default",
    name: "default",
    colors: {
      background: "#2c2d2f",
      surface: "#161718",
      "on-surface": "#ffffff",
      "surface-bright": "#333333",
      "surface-variant": "#7852cd",
      "on-surface-variant": "#f6f7e8",
      card: "#171718",
      selected: "#5e319b",
      table: "#242426", // Variations: 242628, 212325, 1E2022, 1C1D1F, 191A1C, 161718, 131415, 111112
      button: "#1D1E1F",
      primary: "#9E7BEA",
      highlight: "#5e319b",
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
      "on-album": "#ffffff",
      download: "#00bfa5",
      private: "#00b8d4",
      edit: "#2196F3",
      share: "#3F51B5",
      love: "#ef5350",
      terminal: "#4A464F",
      navigation: "#141417",
      "navigation-home": "#0e0f10",
    },
  },

  /* Optional themes that the user can choose from in Settings > General */
  abyss: {
    title: "Abyss",
    name: "abyss",
    dark: true,
    colors: {
      background: "#202020",
      surface: "#202020",
      card: "#242424",
      primary: "#814fd9",
      highlight: "#7e57c2",
      "surface-variant": "#814fd9",
      "on-surface-variant": "#1a1a1a",
      secondary: "#111111",
      "secondary-light": "#1a1a1a",
      accent: "#090c10",
      error: "#e57373",
      info: "#00acc1",
      success: "#26A69A",
      warning: "#ffd740",
      remove: "#9575cd",
      restore: "#64b5f6",
      album: "#7e57c2",
      download: "#673ab7",
      private: "#512da8",
      edit: "#4527a0",
      share: "#311b92",
      love: "#ef5350",
      terminal: "#333333",
      navigation: "#0d0d0d",
      "navigation-home": "#000000",
    },
  },
  carbon: {
    dark: true,
    title: "Carbon",
    name: "carbon",
    colors: {
      background: "#16141c",
      surface: "#16141c",
      card: "#292732",
      primary: "#8a6eff",
      highlight: "#53478a",
      "surface-variant": "#7f63fd",
      secondary: "#0E0D12",
      "secondary-light": "#292733",
      accent: "#262238",
      error: "#e57373",
      info: "#00acc1",
      success: "#26A69A",
      warning: "#ffd740",
      remove: "#e57373",
      restore: "#64b5f6",
      album: "#ffab00",
      download: "#00bfa5",
      private: "#00b8d4",
      edit: "#0AA9FF",
      share: "#9575cd",
      love: "#ef5350",
      terminal: "#292733",
      navigation: "#0E0D12",
      "navigation-home": "#0E0D12",
    },
  },
  chrome: {
    dark: true,
    title: "Chrome",
    name: "chrome",
    colors: {
      background: "#1d1d1d",
      surface: "#1d1d1d",
      card: "#1f1f1f",
      primary: "#ffffff",
      highlight: "#393939",
      "surface-variant": "#ffffff",
      secondary: "#1f1f1f",
      "secondary-light": "#292929",
      accent: "#727272",
      error: "#d36161",
      info: "#0696a7",
      success: "#3da097",
      warning: "#e5c036",
      remove: "#d35442",
      restore: "#3bbeaf",
      album: "#e39c0b",
      download: "#06a590",
      private: "#0AA9C2",
      edit: "#009FF5",
      share: "#9575cd",
      love: "#dd3f3e",
      terminal: "#2f3131",
      navigation: "#1e2122",
      "navigation-home": "#1e2122",
    },
  },
  gemstone: {
    title: "Gemstone",
    name: "gemstone",
    dark: true,
    colors: {
      background: "#2f2f31",
      surface: "#2f2f31",
      card: "#2b2b2d",
      primary: "#AFB4D4",
      highlight: "#545465",
      "surface-variant": "#9BA0C5",
      secondary: "#272727",
      "secondary-light": "#37373a",
      accent: "#333",
      error: "#e57373",
      info: "#00acc1",
      success: "#26A69A",
      warning: "#ffd740",
      remove: "#e57373",
      restore: "#64b5f6",
      album: "#ffab00",
      download: "#00bfa5",
      private: "#00b8d4",
      edit: "#0AA9FF",
      share: "#9575cd",
      love: "#ef5350",
      terminal: "#4A464F",
      navigation: "#1C1C21",
      "navigation-home": "#131316",
    },
  },
  grayscale: {
    title: "Grayscale",
    name: "grayscale",
    dark: true,
    colors: {
      background: "#525252",
      surface: "#525252",
      card: "#5e5e5e",
      primary: "#c8bdb1",
      highlight: "#726e69",
      "surface-variant": "#c8bdb1",
      secondary: "#444",
      "secondary-light": "#5E5E5E",
      accent: "#333",
      error: "#e57373",
      info: "#5a94dd",
      success: "#26A69A",
      warning: "#e3d181",
      love: "#ef5350",
      remove: "#e35333",
      restore: "#64b5f6",
      album: "#ffab40",
      download: "#07bd9f",
      private: "#48bcd6",
      edit: "#0AA9FF",
      share: "#0070a0",
      terminal: "#333333",
      navigation: "#353839",
      "navigation-home": "#212121",
    },
  },
  lavender: {
    title: "Lavender",
    name: "lavender",
    dark: false,
    colors: {
      background: "#fafafa",
      surface: "#FAFBFF",
      card: "#DFE0E8",
      primary: "#9ca2c9",
      highlight: "#6c6f84",
      "surface-variant": "#475185",
      secondary: "#E2E5F3",
      "secondary-light": "#eef0f6",
      accent: "#EAEAF3",
      error: "#e57373",
      info: "#00acc1",
      success: "#26A69A",
      warning: "#ffd740",
      remove: "#e57373",
      restore: "#64b5f6",
      album: "#ffab00",
      download: "#00bfa5",
      private: "#00b8d4",
      edit: "#0AA9FF",
      share: "#9575cd",
      love: "#ef5350",
      terminal: "#333333",
      navigation: "#1b1e32",
      "navigation-home": "#121421",
    },
  },
  legacy: {
    title: "Legacy",
    name: "legacy",
    dark: false,
    colors: {
      background: "#F5F5F5",
      surface: "#F5F5F5",
      card: "#e0e0e0",
      primary: "#FFCA28",
      highlight: "#212121",
      "surface-variant": "#212121",
      secondary: "#bdbdbd",
      "secondary-light": "#e0e0e0",
      accent: "#757575",
      error: "#e57373",
      info: "#00acc1",
      success: "#26A69A",
      warning: "#ffd740",
      remove: "#e57373",
      restore: "#64b5f6",
      album: "#ffab00",
      download: "#00bfa5",
      private: "#00b8d4",
      edit: "#00b8d4",
      share: "#9575cd",
      love: "#ef5350",
      terminal: "#333333",
      navigation: "#212121",
      "navigation-home": "#000000",
    },
  },
  mint: {
    dark: true,
    title: "Mint",
    name: "mint",
    colors: {
      background: "#121212",
      surface: "#121212",
      card: "#1e1e1e",
      primary: "#2bb14c",
      highlight: "#22903d",
      "surface-variant": "#2bb14c",
      secondary: "#181818",
      "secondary-light": "#1f1f1f",
      accent: "#727272",
      error: "#d36161",
      info: "#0696a7",
      success: "#3da097",
      warning: "#e5c036",
      remove: "#d35442",
      restore: "#3bbeaf",
      album: "#e39c0b",
      download: "#06a590",
      private: "#0bb1ca",
      edit: "#0AA9FF",
      share: "#9575cd",
      love: "#ef5350",
      terminal: "#181818",
      navigation: "#181818",
      "navigation-home": "#181818",
    },
  },
  neon: {
    title: "Neon",
    name: "neon",
    dark: true,
    colors: {
      background: "#242326",
      surface: "#242326",
      card: "#1b1a1c",
      primary: "#f44abf",
      highlight: "#890664",
      "surface-variant": "#cc0d99",
      secondary: "#111111",
      "secondary-light": "#1a1a1a",
      accent: "#090c10",
      error: "#e57373",
      info: "#00acc1",
      success: "#26A69A",
      warning: "#fece3e",
      love: "#fb4483",
      remove: "#9100a0",
      restore: "#5e33f8",
      album: "#6234b5",
      download: "#8d56eb",
      private: "#4749c8",
      edit: "#5658eb",
      share: "#5692eb",
      terminal: "#333333",
      navigation: "#0e0d0f",
      "navigation-home": "#000000",
    },
  },
  nordic: {
    dark: false,
    title: "Nordic",
    name: "nordic",
    colors: {
      background: "#f7f8fa",
      "on-background": "#4c566a",
      surface: "#ECEFF4",
      "on-surface": "#3e4757",
      "surface-bright": "#cbced6",
      "surface-variant": "#8590A7",
      "on-surface-variant": "#f6f7e8",
      card: "#eceff4",
      table: "#f2f3f7",
      button: "#E4E6EB",
      "on-button": "#3e4757",
      primary: "#4ca0b8",
      highlight: "#D8DCE3",
      "on-highlight": "#3e4757",
      selected: "#d8dee9",
      secondary: "#E2E7EE",
      "on-secondary": "#4c566a",
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
      navigation: "#E5E9F0",
      "on-navigation": "#3e4757",
      "navigation-home": "#dde3eb",
      "on-navigation-home": "#3e4757",
    },
    variables: {
      "overlay-color": "#f2f2f2",
      "border-color": "#ffffff",
      "border-opacity": 0.08,
      "high-emphasis-opacity": 0.96,
      "medium-emphasis-opacity": 0.7,
      "hover-opacity": 0.06,
      "focus-opacity": 0.08,
    },
  },
  onyx: {
    title: "Onyx",
    name: "onyx",
    dark: false,
    colors: {
      background: "#e5e4e2",
      surface: "#e5e4e2",
      "on-surface": "#000000",
      card: "#a8a8a8",
      button: "#505557",
      table: "#dddcda",
      primary: "#c8bdb1",
      highlight: "#393c3d",
      "surface-variant": "#a39a90",
      "on-surface-variant": "#656565",
      secondary: "#a8a8a8",
      "on-secondary": "#000000",
      "secondary-light": "#cdccca",
      accent: "#656565",
      error: "#e57373",
      info: "#5a94dd",
      success: "#26A69A",
      warning: "#e3d181",
      love: "#ef5350",
      remove: "#e35333",
      restore: "#64b5f6",
      album: "#ffab40",
      download: "#07bd9f",
      private: "#48bcd6",
      edit: "#0AA9FF",
      share: "#0070a0",
      terminal: "#333333",
      navigation: "#353839",
      "navigation-home": "#212121",
    },
  },
  shadow: {
    title: "Shadow",
    name: "shadow",
    dark: true,
    colors: {
      background: "#444",
      surface: "#444",
      card: "#666666",
      primary: "#c4f1e5",
      highlight: "#74817d",
      "surface-variant": "#c8e3e7",
      secondary: "#585858",
      "secondary-light": "#666",
      accent: "#333",
      error: "#e57373",
      info: "#00acc1",
      success: "#26A69A",
      warning: "#ffd740",
      remove: "#e57373",
      restore: "#64b5f6",
      album: "#ffab00",
      download: "#00bfa5",
      private: "#00b8d4",
      edit: "#0AA9FF",
      share: "#9575cd",
      love: "#ef5350",
      terminal: "#333333",
      navigation: "#212121",
      "navigation-home": "#000000",
    },
  },
  vanta: {
    title: "Vanta",
    name: "vanta",
    dark: true,
    colors: {
      background: "#212121",
      surface: "#212121",
      card: "#1d1d1d",
      primary: "#04acaf",
      highlight: "#444444",
      "surface-variant": "#04acaf",
      secondary: "#111111",
      "secondary-light": "#1a1a1a",
      accent: "#090c10",
      error: "#e57373",
      info: "#00acc1",
      success: "#26A69A",
      warning: "#ffd740",
      remove: "#e57373",
      restore: "#64b5f6",
      album: "#ffab00",
      download: "#00bfa5",
      private: "#00b8d4",
      edit: "#0AA9FF",
      share: "#9575cd",
      love: "#ef5350",
      terminal: "#333333",
      navigation: "#0d0d0d",
      "navigation-home": "#000000",
    },
  },
  yellowstone: {
    title: "Yellowstone",
    name: "yellowstone",
    dark: true,
    colors: {
      background: "#32312f",
      surface: "#32312f",
      card: "#262524",
      primary: "#ffb700",
      highlight: "#54524e",
      "surface-variant": "#ffb700",
      secondary: "#21201f",
      "secondary-light": "#262523",
      accent: "#333",
      error: "#e57373",
      info: "#00acc1",
      success: "#26A69A",
      warning: "#ffd740",
      remove: "#e57373",
      restore: "#64b5f6",
      album: "#ffab00",
      download: "#00bfa5",
      private: "#00b8d4",
      edit: "#0AA9FF",
      share: "#9575cd",
      love: "#ef5350",
      terminal: "#464544",
      navigation: "#191817",
      "navigation-home": "#0c0c0b",
    },
  },

  /* Special theme used on the login page */
  login: {
    dark: false,
    title: "Login",
    name: "login",
    colors: {
      background: "#2f3031",
      surface: "#fafafa",
      "on-surface": "#333333",
      "surface-bright": "#fafafa",
      "surface-variant": "#00a6a9",
      "on-surface-variant": "#c8e3e7",
      card: "#505050",
      table: "#505050",
      button: "#c8e3e7",
      primary: "#05dde1",
      highlight: "#00a6a9",
      secondary: "#c8e3e7",
      "secondary-light": "#2a2b2c",
      accent: "#05dde1",
      error: "#e57373",
      info: "#00acc1",
      success: "#4db6ac",
      warning: "#ffd740",
      remove: "#DF5353",
      restore: "#3EA2F4",
      album: "#ffab00",
      download: "#00bfa5",
      private: "#00b8d4",
      edit: "#0AA9FF",
      share: "#9575cd",
      love: "#ef5350",
      terminal: "#4A464F",
      navigation: "#141417",
      "navigation-home": "#0e0f10",
    },
    variables: {
      "border-color": "#ffffff",
      "border-opacity": 0.08,
      "high-emphasis-opacity": 0.96,
      "medium-emphasis-opacity": 0.7,
      "hover-opacity": 0.08,
      "focus-opacity": 0.1,
    },
  },

  /* Special light theme, e.g. used for map controls in Places */
  light: {
    dark: false,
    title: "Light",
    name: "light",
    colors: {
      "on-surface": "#000000",
      "surface-bright": "#FFFFFF",
      "surface-light": "#EEEEEE",
      "surface-variant": "#1e1e1f",
      "on-surface-variant": "#EEEEEE",
    },
    variables: {
      "focus-opacity": 0.0,
    },
  },
};

/* Automatically Generated Theme Color variations */

export const variations = {
  colors: ["primary", "highlight", "secondary", "surface", "navigation"],
  lighten: 2,
  darken: 1,
};

/* Themes Available for Selection in Settings > General */

let options = [
  {
    text: $gettext("Default"),
    value: "default",
    disabled: false,
  },
  {
    text: "Abyss",
    value: "abyss",
    disabled: false,
  },
  {
    text: "Carbon",
    value: "carbon",
    disabled: false,
  },
  {
    text: "Chrome",
    value: "chrome",
    disabled: false,
  },
  {
    text: "Gemstone",
    value: "gemstone",
    disabled: false,
  },
  {
    text: "Grayscale",
    value: "grayscale",
    disabled: false,
  },
  {
    text: "Lavender",
    value: "lavender",
    disabled: false,
  },
  {
    text: "Legacy",
    value: "legacy",
    disabled: false,
  },
  {
    text: "Mint",
    value: "mint",
    disabled: false,
  },
  {
    text: "Neon",
    value: "neon",
    disabled: false,
  },
  {
    text: "Nordic",
    value: "nordic",
    disabled: false,
  },
  {
    text: "Onyx",
    value: "onyx",
    disabled: false,
  },
  {
    text: "Shadow",
    value: "shadow",
    disabled: false,
  },
  {
    text: "Vanta",
    value: "vanta",
    disabled: false,
  },
  {
    text: "Yellowstone",
    value: "yellowstone",
    disabled: false,
  },
];

/* Theme Helper Functions */

// All returns an object containing all defined themes for use with Vuetify.
export const All = () => {
  let result = [];

  for (let k in themes) {
    if (themes.hasOwnProperty(k)) {
      // Get theme definition.
      const theme = themes[k];

      // Skip themes without a name.
      if (!theme["name"]) {
        continue;
      }

      // Get theme style (dark, light).
      const s = style(theme);

      // Add theme definition with presets.
      result[theme.name] = {
        dark: !!theme.dark,
        colors: theme.colors ? { ...colors[s], ...theme.colors } : colors[s],
        variables: theme.variables ? { ...variables[s], ...theme.variables } : variables[s],
      };
    }
  }

  // Return all themes with dark/light presets applied.
  return result;
};

// Get returns a theme by name.
export const Get = (name) => {
  if (typeof themes[name] === "undefined") {
    name = options[0].value;
  }

  // Get theme definition.
  const theme = themes[name];

  // Get theme style (dark, light).
  const s = style(theme);

  // Return theme definition with dark/light presets applied.
  return {
    dark: !!theme.dark,
    title: theme.title ? theme.title : theme.name,
    name: theme.name,
    colors: theme.colors ? { ...colors[s], ...theme.colors } : colors[s],
    variables: theme.variables ? { ...variables[s], ...theme.variables } : variables[s],
  };
};

// Set adds or replaces a theme by name.
export const Set = (name, val) => {
  if (typeof themes[name] === "undefined") {
    options.push({
      text: val.title,
      value: val.name,
      disabled: false,
    });
  }

  themes[name] = val;
};

// Remove deletes a theme by name.
export const Remove = (name) => {
  delete themes[name];
  const i = options.findIndex((el) => el.value === name);
  if (i > -1) {
    options.splice(i, 1);
  }
};

// Translated returns theme selection options with the current locale.
export const Translated = () => {
  return options.map((v) => {
    if (v.disabled) {
      return null;
    }

    return {
      text: T(v.text),
      value: v.value,
    };
  });
};

export const Options = () => options;

export const SetOptions = (v) => (options = v);
