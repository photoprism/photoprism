import { $gettext, T } from "common/gettext";
import { style, colors, variables } from "ui";

/* Theme Color Variations */

export const variations = {
  colors: ["primary", "highlight", "secondary", "surface", "surface-variant", "table", "navigation"],
  lighten: 2,
  darken: 1,
};

/* User Interface Themes */

let themes = {
  /* Default user interface theme */
  default: {
    dark: true,
    title: "Default",
    name: "default",
    colors: {
      "background": "#2c2d2f",
      "surface": "#161718",
      "on-surface": "#ffffff",
      "surface-bright": "#333333",
      "surface-variant": "#7852cd",
      "on-surface-variant": "#f6f7e8",
      "card": "#171718",
      "selected": "#5e319b",
      "table": "#242426", // Variations: 242628, 212325, 1E2022, 1C1D1F, 191A1C, 161718, 131415, 111112
      "button": "#1D1E1F",
      "switch": "#101112",
      "primary": "#9E7BEA",
      "highlight": "#5e319b",
      "secondary": "#191A1C",
      "secondary-light": "#1E2022",
      "accent": "#2D2E2E",
      "error": "#e57373",
      "info": "#9E7BEA",
      "success": "#8763d5",
      "warning": "#ecc434",
      "favorite": "#FFD600",
      "remove": "#da4e4c",
      "restore": "#00d48a",
      "album": "#ed9e00",
      "on-album": "#ffffff",
      "download": "#00bfa5",
      "private": "#00b8d4",
      "edit": "#2196F3",
      "share": "#3F51B5",
      "love": "#ef5350",
      "terminal": "#4A464F",
      "navigation": "#141417",
      "navigation-home": "#0e0f10",
    },
  },

  /* Optional themes that the user can choose from in Settings > General */
  abyss: {
    title: "Abyss",
    name: "abyss",
    dark: true,
    colors: {
      "background": "#202020",
      "surface": "#0f0f0f",
      "card": "#242424",
      "primary": "#814fd9",
      "highlight": "#7e57c2",
      "surface-variant": "#814fd9",
      "on-surface-variant": "#1a1a1a",
      "secondary": "#111111",
      "secondary-light": "#1a1a1a",
      "table": "#242424",
      "button": "1a1a1a",
      "selected": "#64459b",
      "accent": "#090c10",
      "error": "#e57373",
      "info": "#00acc1",
      "success": "#26A69A",
      "warning": "#ffd740",
      "remove": "#9575cd",
      "restore": "#64b5f6",
      "album": "#7e57c2",
      "download": "#673ab7",
      "private": "#512da8",
      "edit": "#4527a0",
      "share": "#311b92",
      "love": "#ef5350",
      "terminal": "#333333",
      "navigation": "#0d0d0d",
      "navigation-home": "#000000",
    },
    variables: {
      "disabled-opacity": 0.6,
      "hover-opacity": 0.03,
    },
  },
  carbon: {
    dark: true,
    title: "Carbon",
    name: "carbon",
    colors: {
      "background": "#16141c",
      "surface": "#24212E",
      "card": "#292732",
      "primary": "#8a6eff",
      "highlight": "#53478a",
      "surface-variant": "#7f63fd",
      "secondary": "#0E0D12",
      "secondary-light": "#292733",
      "table": "#1D1B26",
      "button": "#2C273E",
      "switch": "#707070",
      "selected": "#53478a",
      "accent": "#262238",
      "error": "#e57373",
      "info": "#00acc1",
      "success": "#26A69A",
      "warning": "#ffd740",
      "remove": "#e57373",
      "restore": "#64b5f6",
      "album": "#ffab00",
      "download": "#00bfa5",
      "private": "#00b8d4",
      "edit": "#0AA9FF",
      "share": "#9575cd",
      "love": "#ef5350",
      "terminal": "#292733",
      "navigation": "#0E0D12",
      "navigation-home": "#0E0D12",
    },
    variables: {
      "overlay-color": "#24212E",
      "disabled-opacity": 0.6,
      "hover-opacity": 0.03,
    },
  },
  chrome: {
    dark: true,
    title: "Chrome",
    name: "chrome",
    colors: {
      "background": "#1e1f20",
      "on-background": "#ffffff",
      "surface": "#202121",
      "card": "#1a1b1c",
      "primary": "#ffffff",
      "highlight": "#393a3b",
      "surface-variant": "#e8e9eb",
      "on-surface-variant": "#262728",
      "secondary": "#1a1b1c",
      "secondary-light": "#292929",
      "button": "#2c2d2e",
      "table": "#262728",
      "switch": "#707070",
      "selected": "#48494b",
      "accent": "#727272",
      "error": "#d36161",
      "info": "#0696a7",
      "success": "#3da097",
      "warning": "#e5c036",
      "remove": "#d35442",
      "restore": "#3bbeaf",
      "album": "#e39c0b",
      "download": "#06a590",
      "private": "#0AA9C2",
      "edit": "#009FF5",
      "share": "#9575cd",
      "love": "#dd3f3e",
      "terminal": "#333333",
      "navigation": "#1a1b1c",
      "navigation-home": "#1a1b1c",
    },
    variables: {
      "overlay-color": "#424242",
      "disabled-opacity": 0.55,
      "hover-opacity": 0.03,
      "border-opacity": 0.16,
    },
  },
  gemstone: {
    title: "Gemstone",
    name: "gemstone",
    dark: true,
    colors: {
      "background": "#2b2c31",
      "surface": "#1D1E24",
      "card": "#26272C",
      "primary": "#AFB4D4",
      "highlight": "#45455c",
      "switch": "#474a60",
      "surface-variant": "#6e74a1",
      "on-surface-variant": "#f6f7e8",
      "secondary": "#222228",
      "secondary-light": "#37373a",
      "table": "#27282d",
      "button": "#202126",
      "selected": "#424771",
      "accent": "#333",
      "error": "#dc7171",
      "info": "#9aa6f4",
      "success": "#858cc3",
      "warning": "#d38e03",
      "remove": "#e57373",
      "restore": "#64b5f6",
      "album": "#ffab00",
      "download": "#00bfa5",
      "private": "#00b8d4",
      "edit": "#0AA9FF",
      "share": "#9575cd",
      "love": "#ef5350",
      "terminal": "#393A41",
      "navigation": "#222228",
      "navigation-home": "#1C1C20",
    },
    variables: {
      "overlay-color": "#1E1F24",
      "disabled-opacity": 0.62,
      "focus-opacity": 0.056,
      "hover-opacity": 0.01,
      "border-opacity": 0.14,
      "fill-opacity": 0.03,
    },
  },
  grayscale: {
    title: "Grayscale",
    name: "grayscale",
    dark: true,
    colors: {
      "background": "#525252",
      "surface": "#424242",
      "card": "#5e5e5e",
      "primary": "#c8bdb1",
      "highlight": "#726e69",
      "surface-variant": "#c8bdb1",
      "on-surface-variant": "#252525",
      "secondary": "#444",
      "secondary-light": "#5E5E5E",
      "button": "#343434",
      "table": "#4e4e4e",
      "selected": "#252525",
      "accent": "#333",
      "error": "#e57373",
      "info": "#5a94dd",
      "success": "#26A69A",
      "warning": "#e3d181",
      "love": "#ef5350",
      "remove": "#e35333",
      "restore": "#64b5f6",
      "album": "#ffab40",
      "download": "#07bd9f",
      "private": "#48bcd6",
      "edit": "#0AA9FF",
      "share": "#0070a0",
      "terminal": "#333333",
      "navigation": "#353839",
      "navigation-home": "#212121",
    },
    variables: {
      "disabled-opacity": 0.6,
      "hover-opacity": 0.03,
    },
  },
  lavender: {
    title: "Lavender",
    name: "lavender",
    dark: false,
    colors: {
      "background": "#F3F3F5",
      "surface": "#dadbe6",
      "on-surface": "#17171b",
      "card": "#DFE0E8",
      "primary": "#9ca2c9",
      "highlight": "#6E7189",
      "surface-variant": "#53557a",
      "secondary": "#c4c4cf",
      "secondary-light": "#eef0f6",
      "selected": "#797ea3",
      "table": "#E5E5EB",
      "button": "#8A8CA8",
      "switch": "#707070",
      "accent": "#EAEAF3",
      "error": "#e57373",
      "info": "#7887df",
      "success": "#26A69A",
      "warning": "#bfa965",
      "remove": "#e57373",
      "restore": "#64b5f6",
      "album": "#ffab00",
      "download": "#00bfa5",
      "private": "#00b8d4",
      "edit": "#0AA9FF",
      "share": "#9575cd",
      "love": "#ef5350",
      "terminal": "#e7e8f2",
      "navigation": "#393B4B",
      "navigation-home": "#2C2D3B",
    },
    variables: {
      "overlay-color": "#2C2D3B",
      "overlay-opacity": 0.26,
      "high-emphasis-opacity": 0.99,
      "medium-emphasis-opacity": 0.92,
      "disabled-opacity": 0.66,
      "hover-opacity": 0.01,
      "focus-opacity": 0.02,
      "fill-opacity": 0.01,
    },
  },
  legacy: {
    title: "Legacy",
    name: "legacy",
    dark: false,
    colors: {
      "background": "#F5F5F5",
      "surface": "#E7E7E7",
      "card": "#e0e0e0",
      "primary": "#FFCA28",
      "highlight": "#212121",
      "surface-variant": "#212121",
      "secondary": "#bdbdbd",
      "secondary-light": "#e0e0e0",
      "button": "#E0E0E0",
      "table": "#FFFFFF",
      "selected": "#212121",
      "accent": "#757575",
      "error": "#e57373",
      "info": "#00acc1",
      "success": "#26A69A",
      "warning": "#ffd740",
      "remove": "#e57373",
      "restore": "#64b5f6",
      "album": "#ffab00",
      "download": "#00bfa5",
      "private": "#00b8d4",
      "edit": "#00b8d4",
      "share": "#9575cd",
      "love": "#ef5350",
      "terminal": "#bdbdbd",
      "navigation": "#212121",
      "navigation-home": "#000000",
    },
    variables: {
      "disabled-opacity": 0.55,
      "hover-opacity": 0.06,
    },
  },
  mint: {
    dark: true,
    title: "Mint",
    name: "mint",
    colors: {
      "background": "#121212",
      "surface": "#191919",
      "card": "#1e1e1e",
      "primary": "#2bb14c",
      "highlight": "#22903d",
      "surface-variant": "#2bb14c",
      "secondary": "#181818",
      "secondary-light": "#1f1f1f",
      "table": "#1a1a1a",
      "button": "#1F1F1F",
      "switch": "#ffffff",
      "selected": "228d3c",
      "accent": "#727272",
      "error": "#d36161",
      "info": "#0696a7",
      "success": "#3da097",
      "warning": "#e5c036",
      "remove": "#d35442",
      "restore": "#3bbeaf",
      "album": "#e39c0b",
      "download": "#06a590",
      "private": "#0bb1ca",
      "edit": "#0AA9FF",
      "share": "#9575cd",
      "love": "#ef5350",
      "terminal": "#181818",
      "navigation": "#181818",
      "navigation-home": "#181818",
    },
    variables: {
      "disabled-opacity": 0.55,
      "hover-opacity": 0.02,
      "focus-opacity": 0.02,
      "fill-opacity": 0.02,
    },
  },
  neon: {
    title: "Neon",
    name: "neon",
    dark: true,
    colors: {
      "background": "#242326",
      "surface": "#0f0f0f",
      "card": "#1b1a1c",
      "primary": "#f44abf",
      "highlight": "#890664",
      "surface-variant": "#cc0d99",
      "secondary": "#111111",
      "secondary-light": "#1a1a1a",
      "button": "1a1a1a",
      "table": "#302E32",
      "selected": "#a30a7a",
      "accent": "#090c10",
      "error": "#e57373",
      "info": "#00acc1",
      "success": "#26A69A",
      "warning": "#fece3e",
      "love": "#fb4483",
      "remove": "#9100a0",
      "restore": "#5e33f8",
      "album": "#6234b5",
      "download": "#8d56eb",
      "private": "#4749c8",
      "edit": "#5658eb",
      "share": "#5692eb",
      "terminal": "#333333",
      "navigation": "#0e0d0f",
      "navigation-home": "#000000",
    },
    variables: {
      "disabled-opacity": 0.65,
      "hover-opacity": 0.03,
    },
  },
  nordic: {
    dark: false,
    title: "Nordic",
    name: "nordic",
    colors: {
      "background": "#f7f8fa",
      "on-background": "#4c566a",
      "surface": "#edf0f6",
      "on-surface": "#3e4757",
      "surface-bright": "#cbced6",
      "surface-variant": "#4ca0b8",
      "on-surface-variant": "#f6f7e8",
      "card": "#eceff4",
      "on-card": "#3e4757",
      "table": "#f2f3f7",
      "button": "#ECEFF4",
      "switch": "#333333",
      "on-button": "#3e4757",
      "primary": "#4ca0b8",
      "highlight": "#519FB6",
      "on-highlight": "#ffffff",
      "selected": "#bae4fa",
      "on-selected": "#3e4757",
      "secondary": "#e4e9f1",
      "on-secondary": "#3e4757",
      "secondary-light": "#f3f5f8",
      "accent": "#F2F5FA",
      "error": "#BF616A",
      "info": "#88C0D0",
      "success": "#8FBCBB",
      "warning": "#f0d8a8",
      "favorite": "#EBCB8B",
      "remove": "#BF616A",
      "restore": "#81A1C1",
      "album": "#EBCB8B",
      "download": "#8FBCBB",
      "private": "#88C0D0",
      "edit": "#88C0D0",
      "share": "#B48EAD",
      "love": "#ef5350",
      "terminal": "#e5e9f0",
      "navigation": "#E5E9F0",
      "on-navigation": "#3e4757",
      "navigation-home": "#dde3eb",
      "on-navigation-home": "#3e4757",
    },
    variables: {
      "overlay-color": "#f2f2f2",
      "border-color": "#555556",
      "border-opacity": 0.08,
      "high-emphasis-opacity": 0.96,
      "medium-emphasis-opacity": 0.7,
      "hover-opacity": 0.01,
      "focus-opacity": 0.02,
      "fill-opacity": 0.01,
      "shadow-key-umbra-opacity": "#cbced630",
      "shadow-key-penumbra-opacity": "#cbced624",
      "shadow-key-ambient-opacity": "#cbced61f",
    },
  },
  onyx: {
    title: "Onyx",
    name: "onyx",
    dark: false,
    colors: {
      "background": "#e5e4e2",
      "surface": "#d8d7d5",
      "card": "#CDCCCA",
      "on-card": "#1b1a1a",
      "button": "#CDCCCA",
      "switch": "#a7a7a7",
      "table": "#eeeeee",
      "selected": "#807870",
      "primary": "#a0978d",
      "highlight": "#393c3d",
      "surface-variant": "#57595A",
      "on-surface-variant": "#ffffff",
      "secondary": "#a8a8a8",
      "secondary-light": "#cdccca",
      "on-secondary": "#000000",
      "accent": "#656565",
      "error": "#e57373",
      "info": "#5a94dd",
      "success": "#26A69A",
      "warning": "#e3d181",
      "love": "#ef5350",
      "remove": "#e35333",
      "restore": "#64b5f6",
      "album": "#ffab40",
      "download": "#07bd9f",
      "private": "#48bcd6",
      "edit": "#0AA9FF",
      "share": "#0070a0",
      "terminal": "#a8a8a8",
      "navigation": "#353839",
      "navigation-home": "#212121",
    },
    variables: {
      "disabled-opacity": 0.55,
      "hover-opacity": 0.06,
    },
  },
  shadow: {
    title: "Shadow",
    name: "shadow",
    dark: true,
    colors: {
      "background": "#444",
      "surface": "#555555",
      "button": "#555555",
      "card": "#666666",
      "primary": "#c4f1e5",
      "highlight": "#759089",
      "surface-variant": "#c8e3e7",
      "on-surface-variant": "#222222",
      "secondary": "#343434",
      "secondary-light": "#666",
      "table": "#3e3e3e",
      "selected": "#5d736d",
      "accent": "#333",
      "error": "#e57373",
      "info": "#00acc1",
      "success": "#26A69A",
      "warning": "#ffd740",
      "remove": "#e57373",
      "restore": "#64b5f6",
      "album": "#ffab00",
      "download": "#00bfa5",
      "private": "#00b8d4",
      "edit": "#0AA9FF",
      "share": "#9575cd",
      "love": "#ef5350",
      "terminal": "#333333",
      "navigation": "#212121",
      "navigation-home": "#000000",
    },
    variables: {
      "border-color": "#2f2f2f",
      "border-opacity": 0.25,
      "disabled-opacity": 0.65,
      "hover-opacity": 0.05,
    },
  },
  vanta: {
    title: "Vanta",
    name: "vanta",
    dark: true,
    colors: {
      "background": "#212121",
      "button": "#1A1A1A",
      "surface": "#0d0d0d",
      "card": "#1d1d1d",
      "primary": "#04acaf",
      "highlight": "#03898c",
      "surface-variant": "#04acaf",
      "on-surface-variant": "#21201F",
      "secondary": "#111111",
      "secondary-light": "#1a1a1a",
      "table": "#262626",
      "selected": "#026769",
      "accent": "#090c10",
      "error": "#e57373",
      "info": "#00acc1",
      "success": "#26A69A",
      "warning": "#ffd740",
      "remove": "#e57373",
      "restore": "#64b5f6",
      "album": "#ffab00",
      "download": "#00bfa5",
      "private": "#00b8d4",
      "edit": "#0AA9FF",
      "share": "#9575cd",
      "love": "#ef5350",
      "terminal": "#333333",
      "navigation": "#0d0d0d",
      "navigation-home": "#000000",
    },
    variables: {
      "border-color": "#021212",
      "border-opacity": 0.25,
      "disabled-opacity": 0.65,
      "hover-opacity": 0.03,
    },
  },
  yellowstone: {
    title: "Yellowstone",
    name: "yellowstone",
    dark: true,
    colors: {
      "background": "#32312f",
      "surface": "#161615",
      "surface-variant": "#ffb700",
      "on-surface-variant": "#21201F",
      "card": "#262524",
      "selected": "#ffc430",
      "table": "#373532",
      "button": "#262523",
      "primary": "#ffb700",
      "highlight": "#ffb700",
      "secondary": "#1a1918",
      "secondary-light": "#262523",
      "accent": "#333",
      "error": "#e57373",
      "info": "#00acc1",
      "success": "#26A69A",
      "warning": "#ffd740",
      "remove": "#e57373",
      "restore": "#64b5f6",
      "album": "#ffab00",
      "download": "#00bfa5",
      "private": "#00b8d4",
      "edit": "#0AA9FF",
      "share": "#9575cd",
      "love": "#ef5350",
      "terminal": "#464544",
      "navigation": "#191817",
      "navigation-home": "#0c0c0b",
    },
    variables: {
      "border-color": "#282725",
      "border-opacity": 0.35,
      "disabled-opacity": 0.65,
      "hover-opacity": 0.06,
    },
  },

  /* Special theme used for the photo/video viewer */
  lightbox: {
    dark: true,
    title: "Lightbox",
    name: "lightbox",
    colors: {
      "background": "#131313",
      "surface": "#131313",
      "on-surface": "#ffffff",
      "surface-bright": "#333333",
      "surface-variant": "#cccccc",
      "on-surface-variant": "#f6f6f6",
      "card": "#171717",
      "selected": "#9c9c9c",
      "table": "#242424", // Variations: 242628, 212325, 1E2022, 1C1D1F, 191A1C, 161718, 131415, 111112
      "button": "#1D1E1F",
      "switch": "#101112",
      "primary": "#ebebeb",
      "highlight": "#9c9c9c",
      "secondary": "#191919",
      "secondary-light": "#1e1e1e",
      "accent": "#2D2E2E",
      "error": "#e57373",
      "info": "#9E7BEA",
      "success": "#8763d5",
      "warning": "#ecc434",
      "favorite": "#FFD600",
      "remove": "#da4e4c",
      "restore": "#00d48a",
      "album": "#ed9e00",
      "on-album": "#ffffff",
      "download": "#00bfa5",
      "private": "#00b8d4",
      "edit": "#2196F3",
      "share": "#3F51B5",
      "love": "#ef5350",
      "terminal": "#4A464F",
      "navigation": "#121212",
      "navigation-home": "#121212",
    },
  },

  /* Special theme used on the login page */
  login: {
    dark: false,
    title: "Login",
    name: "login",
    colors: {
      "background": "#2f3031",
      "surface": "#fafafa",
      "on-surface": "#333333",
      "surface-bright": "#fafafa",
      "surface-variant": "#00a6a9",
      "on-surface-variant": "#c8e3e7",
      "card": "#505050",
      "table": "#505050",
      "button": "#c8e3e7",
      "primary": "#05dde1",
      "highlight": "#00a6a9",
      "secondary": "#c8e3e7",
      "secondary-light": "#2a2b2c",
      "accent": "#05dde1",
      "error": "#e57373",
      "info": "#00acc1",
      "success": "#4db6ac",
      "warning": "#ffd740",
      "remove": "#DF5353",
      "restore": "#3EA2F4",
      "album": "#ffab00",
      "download": "#00bfa5",
      "private": "#00b8d4",
      "edit": "#0AA9FF",
      "share": "#9575cd",
      "love": "#ef5350",
      "terminal": "#4A464F",
      "navigation": "#141417",
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
      "background": "#ffffff",
      "surface": "#ffffff",
      "on-surface": "#000000",
      "surface-bright": "#FFFFFF",
      "surface-light": "#EEEEEE",
      "surface-variant": "#1e1e1f",
      "on-surface-variant": "#EEEEEE",
    },
    variables: {
      "border-color": "#1e1e1f",
      "high-emphasis-opacity": 1.0,
      "medium-emphasis-opacity": 0.8,
      "label-opacity": 0.96,
      "focus-opacity": 0.0,
      "hover-opacity": 0.05,
    },
  },
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
export const Set = (name, theme) => {
  if (!theme) {
    return;
  }

  if (!name) {
    name = theme.name;
  }

  const force = theme?.force;

  if (force) {
    // If the force flag is set, make this theme the only available option.
    options = [
      {
        text: theme.title ? theme.title : $gettext("Custom"),
        value: name,
        disabled: false,
      },
    ];
  } else if (typeof themes[name] === "undefined") {
    // Otherwise, add it to the list of available themes,
    // unless a theme with the same name already exists.
    options.push({
      text: theme.title ? theme.title : $gettext("Custom"),
      value: name,
      disabled: false,
    });
  }

  themes[name] = theme;
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
