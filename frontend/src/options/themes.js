import { $gettext, T } from "common/gettext";
import { style, colors, variables } from "ui";

/* Theme Color Variations */

export const variations = {
  colors: ["primary", "highlight", "secondary", "surface", "surface-variant", "table", "navigation", "add", "remove"],
  lighten: 2,
  darken: 2,
};

/* User Interface Themes */

let themes = {
  /* Default user interface theme */
  default: {
    dark: true,
    title: "Default",
    name: "default",
    colors: {
      "background": "#19191a", // Page canvas; the lowest elevation level (cards and data tables fall back to this).
      "on-background": "#f9fafb", // Text and icon color readable against `background`.
      "surface": "#262628", // Default container surface for sheets, dialogs, and list items (sits above `background`).
      "on-surface": "#f9fafb", // Body text and icon color on `surface`, `surface-bright`, and `surface-light`.
      "surface-bright": "#333333", // Lifted variant of `surface`; raised tiles, hover states, inline editors.
      "surface-variant": "#7852cd", // Muted brand-purple foreground tone consumed by Vuetify defaults for active dropdown rows, focus rings, and `color="surface-variant"` props.
      "on-surface-variant": "#f6f7e8", // Text/icon color that contrasts with `surface-variant` when used as a background.
      "primary": "#9E7BEA", // Brand/identity accent (icons, active-tab text, chip focus ring).
      "secondary": "#19191a", // Background for secondary panels (tab strips, expansion-panel headers, nav drawer sections).
      "secondary-light": "#202022", // Lifted variant of `secondary`; used for raised surfaces inside secondary panels.
      "accent": "#303136", // Small-decoration tint for hover/focus subtleties; not a primary action color.
      "card": "#27272a", // Dedicated card-container background; a third tier between `surface` and `background`.
      "selected": "#5e319b", // Active list-item background; pairs with `on-selected` for text inside the row.
      "highlight": "#5e319b", // Primary call-to-action accent (Confirm/Save/Apply buttons, chips, active toggle).
      "switch": "#101112", // VSwitch track background when off.
      "button": "#262628", // Secondary button color (Cancel, dismiss) — the neutral companion to `highlight`.
      "table": "#202021", // VDataTable row and header background.
      "on-table": "#f9fafb", // Text and icon color on `table` rows.
      "error": "#e57373", // Error state for banners, validation errors, and error toasts.
      "info": "#00acc1", // Informational notification color (snackbars, info-level log icons, neutral badges).
      "success": "#4db6ac", // Successful-outcome notification color (saved toast, completed-job indicator).
      "warning": "#bc9714", // Caution / recoverable-concern notification color (warning banners, paused indicators).
      "favorite": "#FFD600", // Favorite-star color.
      "add": "#00897B", // User-initiated "add to collection" action color (paired with `remove`; distinct from `download` and `success`).
      "remove": "#d35150", // Destructive "remove from collection" action color (distinct from `error`, which is a fault state).
      "restore": "#00d48a", // Restore-from-trash / undo-remove action color.
      "album": "#ed9e00", // Album identity color (album icons, chips, thumbnail accents).
      "on-album": "#ffffff", // Text/icon color on album-tinted backgrounds.
      "download": "#00bfa5", // Download affordance (buttons, progress).
      "private": "#00b8d4", // Private badge / lock indicator.
      "edit": "#2196F3", // Edit affordance (usually muted; rarely tinted).
      "share": "#3F51B5", // Share affordance (share links, share dialog accents).
      "love": "#ef5350", // Love / heart indicator (emotional emphasis only).
      "terminal": "#282730", // Background for terminal / code blocks (log views, code samples).
      "navigation": "#19191a", // App bar / top toolbar background; matches `background` for a flat look in this theme.
      "navigation-home": "#19191a", // "Home" navigation-state background; matches `navigation` in this theme.
    },
    variables: {
      "border-color": "#363636", // Divider and outlined-variant border color.
      "border-opacity": 0.46, // Alpha applied to `border-color`.
      "fill-opacity": 0, // Background fill alpha for `solo-filled` inputs (0 = no fill in this theme).
      "hover-opacity": 0.03, // Hover-overlay alpha on list items, buttons, and rows.
      "disabled-opacity": 0.65, // Alpha applied to disabled controls.
      "focus-opacity": 0.05, // Keyboard-focus overlay alpha on inputs and rows.
    },
  },

  /* Special theme used for the photo/video viewer */
  lightbox: {
    dark: true,
    title: "Lightbox",
    name: "lightbox",
    colors: {
      "background": "#0c0d0d", // Page canvas; near-black so the photo dominates. Also painted onto `navigation` rows.
      "on-background": "#ffffff", // Text and icon color readable against `background`.
      "surface": "#181818", // Default container surface (sidebar, dropdown menus, dialog v-card); one step above `background`.
      "on-surface": "#f9fafb", // Body text and icon color on `surface`, `surface-bright`, `surface-light`, and `card`.
      "surface-bright": "#1c1c1c", // Lifted variant of `surface`; raised tiles, hover backgrounds.
      "surface-variant": "#A19D9F", // Mid-grey foreground consumed by Vuetify defaults for active dropdown rows, focus rings, and `color="surface-variant"` props.
      "on-surface-variant": "#1c1c1c", // Text/icon color that contrasts with `surface-variant` when used as a background.
      "card": "#171717", // Dedicated card-container background; raised dialog v-cards inherit this via the VCard default.
      "selected": "#3d3f40", // Active list-item background; neutral grey one step above `highlight`.
      "table": "#242424", // VDataTable row and header background.
      "button": "#242424", // Secondary button color (e.g. dialog "Cancel"); the neutral companion to `highlight`.
      "highlight": "#3c3c3c", // Primary button color (Save / Confirm); muted dark grey to keep the action calm.
      "switch": "#101112", // VSwitch track background when off.
      "primary": "#9E8FC9", // Brand/identity accent and base color for face-marker `__rect--named` / `__handle` strokes; muted purple chosen to match the lightbox aesthetic.
      "secondary": "#191919", // Background for secondary panels (tab strips, expansion-panel headers, nav drawer sections).
      "secondary-light": "#1e1e1e", // Lifted variant of `secondary`; backs `.meta-chip` and the chip-selector chips.
      "accent": "#BDAFE4", // Light-lavender accent; consumed by face-marker `__rect--draft` (drag-to-create state).
      "error": "#e57373", // Error state for banners, validation errors, and error toasts.
      "info": "#9E7BEA", // Informational notification color (info toasts); purple matches the `primary` family.
      "success": "#8763d5", // Successful-outcome notification color (matches the `info` purple cast).
      "warning": "#ecc434", // Caution / recoverable-concern notification color.
      "favorite": "#FFD600", // Favorite-star color.
      "remove": "#cd4645", // Destructive "remove from collection" action color (also consumed by face-marker `__rect--removing` and the Remove pill; distinct from `error`, a fault state).
      "restore": "#00d48a", // Restore-from-trash / undo-remove action color.
      "album": "#ed9e00", // Album identity color (album icons, chips, thumbnail accents).
      "on-album": "#ffffff", // Text/icon color on album-tinted backgrounds.
      "download": "#00bfa5", // Download affordance (buttons, progress).
      "private": "#00b8d4", // Private badge / lock indicator.
      "edit": "#2196F3", // Edit affordance (usually muted; rarely tinted).
      "share": "#3F51B5", // Share affordance (share links, share dialog accents).
      "love": "#ef5350", // Love / heart indicator (emotional emphasis only).
      "terminal": "#4A464F", // Background for terminal / code blocks (log views, code samples).
      "navigation": "#0c0d0d", // App bar / top toolbar background; matches `background` for a flat lightbox look.
      "navigation-home": "#0c0d0d", // "Home" navigation-state background; matches `navigation` / `background` in this theme.
    },
    variables: {
      "border-color": "#ffffff", // Divider and outlined-variant border color (rendered against `border-opacity`).
      "border-opacity": 0.1, // Alpha on `border-color`; reads as a barely-there divider against the near-black canvas.
      "hover-opacity": 0.08, // Hover-overlay alpha; bumped above the Vuetify dark default (~0.04) to read on near-black.
      "focus-opacity": 0.06, // Keyboard-focus overlay alpha; intentionally below `hover-opacity` so hover+focus don't compound.
      "fill-opacity": 0.04, // Background-fill alpha for `solo-filled` inputs (the project-wide default; see `defaults.js`).
      "overlay-color": "#141417", // v-overlay scrim color behind v-dialog / v-menu modals.
      "overlay-opacity": 0.6, // Alpha on `overlay-color`; the photo stays visible but dimmed behind sidebar dialogs.
      "theme-overlay-multiplier": 0.16, // Vuetify elevation-overlay multiplier; kept low to keep raised surfaces near-black.
      "high-emphasis-opacity": 0.96, // Body-text alpha on `surface`; under 1.0 so pure-white doesn't read as harsh.
      "medium-emphasis-opacity": 0.88, // Secondary-text alpha (captions, helpers); above the 0.7 dark default for legibility.
      "label-opacity": 0.67, // Floating-label and helper-text alpha on inputs.
      "disabled-opacity": 0.75, // Alpha applied to disabled controls.
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
      "warning": "#bc9714",
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
      "button": "#161616",
      "selected": "#64459b",
      "accent": "#090c10",
      "error": "#e57373",
      "info": "#00acc1",
      "success": "#26A69A",
      "warning": "#bc9714",
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
      "overlay-opacity": 0.62,
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
      "warning": "#bc9714",
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
      "add": "#94d5c4",
      "remove": "#e45d6a",
      "on-remove": "#F3F3F5",
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
      "warning": "#bc9714",
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
      "background": "#181818",
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
      "selected": "#286c29",
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
      "button": "#1a191a",
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
      "error": "#d5303f",
      "on-error": "#cd1b2b",
      "danger": "#9f2727",
      "info": "#4aa2bc",
      "on-info": "#323742",
      "success": "#89d1cf",
      "on-success": "#323742",
      "warning": "#d88a0b",
      "on-warning": "#b87d16",
      "favorite": "#EBCB8B",
      "add": "#b2ddd2",
      "on-add": "#323742",
      "remove": "#e49ca4",
      "on-remove": "#323742",
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
      "warning": "#bc9714",
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
  thinking: {
    title: "Thinking",
    name: "thinking",
    dark: true,
    colors: {
      "background": "#191a1a",
      "on-background": "#f4f6fc",
      "surface": "#252727",
      "on-surface": "#f6f6fa",
      "surface-bright": "#262626",
      "surface-variant": "#999999",
      "on-surface-variant": "#1f2121",
      "primary": "#906fe9",
      "highlight": "#683daf",
      "secondary": "#191a1a",
      "on-secondary": "#f2f2f4",
      "secondary-light": "#1b1d1d",
      "accent": "#232323",
      "card": "#242424",
      "on-card": "#fafafa",
      "selected": "#424242",
      "on-selected": "#ffffff",
      "switch": "#6c6c6c",
      "button": "#303232",
      "on-button": "#f4f4f5",
      "table": "#212222",
      "on-table": "#f2f2f4",
      "error": "#f87171",
      "info": "#60a5fa",
      "success": "#10b981",
      "warning": "#fbbf24",
      "on-warning": "#ffffff",
      "favorite": "#facc15",
      "remove": "#f87171",
      "restore": "#45b8e6",
      "album": "#8d71e6",
      "on-album": "#ffffff",
      "download": "#34c6dc",
      "private": "#47b4e7",
      "edit": "#7f8df0",
      "share": "#a586f2",
      "love": "#fb7185",
      "terminal": "#111415",
      "navigation": "#191a1a",
      "navigation-home": "#191a1a",
    },
    variables: {
      "border-color": "#383838",
      "border-opacity": 0.18,
      "disabled-opacity": 0.58,
      "hover-opacity": 0.045,
      "focus-opacity": 0.06,
      "fill-opacity": 0,
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
      "warning": "#bc9714",
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
      "warning": "#bc9714",
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
    text: "Thinking",
    value: "thinking",
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
export const Get = (name, preferForced = true) => {
  if (Array.isArray(options) && preferForced) {
    const forced = options.find((t) => t.force && t.value);
    if (forced) {
      name = forced.value;
    }
  }

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
        force: true,
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

// Assign adds or replaces multiple themes at once.
export const Assign = (t) => {
  for (const theme of t) {
    if (theme?.name && theme?.colors) {
      Set(theme.name, theme);
    }
  }
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
