/* Dark Theme Presets */

export const dark = "dark";

export const colorsDark = {
  "background": "#2c2d2f", // Page canvas; the lowest elevation level (cards and data tables fall back to this).
  "surface": "#161718", // Default container surface for sheets, dialogs, and list items.
  "on-surface": "#ffffff", // Body text and icon color on `surface`, `surface-bright`, and `surface-light`.
  "surface-bright": "#333333", // Lifted variant of `surface`; raised tiles, hover states, inline editors.
  "surface-variant": "#7E4FE3", // Foreground tone consumed by Vuetify defaults for active dropdown rows, focus rings, and `color="surface-variant"` props.
  "on-surface-variant": "#f6f7e8", // Text/icon color that contrasts with `surface-variant` when used as a background.
  "card": "#171718", // Dedicated card-container background; a third tier between `surface` and `background`.
  "table": "#1F2022", // VDataTable row and header background.
  "button": "#1D1E1F", // Secondary button color (Cancel, dismiss) — the neutral companion to `highlight`.
  "switch": "#101112", // VSwitch track background when off.
  "primary": "#9E7BEA", // Brand/identity accent (icons, active-tab text, chip focus ring).
  "highlight": "#5F1DB7", // Primary call-to-action accent (Confirm/Save/Apply buttons, chips, active toggle).
  "selected": "#4d4d4e", // Active list-item background; pairs with `on-selected` for text inside the row.
  "secondary": "#191A1C", // Background for secondary panels (tab strips, expansion-panel headers, nav drawer sections).
  "secondary-light": "#1E2022", // Lifted variant of `secondary`; used for raised surfaces inside secondary panels.
  "accent": "#2D2E2E", // Small-decoration tint for hover/focus subtleties; not a primary action color.
  "error": "#e57373", // Error state for banners, validation errors, and error toasts.
  "danger": "#e57373", // Legacy alias for `error`; kept for backward compatibility — new code should bind `error`.
  "info": "#00acc1", // Informational notification color (snackbars, info-level log icons, neutral badges).
  "success": "#4db6ac", // Successful-outcome notification color (saved toast, completed-job indicator).
  "warning": "#bc9714", // Caution / recoverable-concern notification color (warning banners, paused indicators).
  "favorite": "#FFD600", // Favorite-star color.
  "remove": "#da4e4c", // Destructive "remove from collection" action color (distinct from `error`, which is a fault state).
  "restore": "#00d48a", // Restore-from-trash / undo-remove action color.
  "album": "#ed9e00", // Album identity color (album icons, chips, thumbnail accents).
  "on-album": "#ffffff", // Text/icon color on album-tinted backgrounds.
  "download": "#00bfa5", // Download affordance (buttons, progress).
  "private": "#00b8d4", // Private badge / lock indicator.
  "edit": "#2196F3", // Edit affordance (usually muted; rarely tinted).
  "share": "#3F51B5", // Share affordance (share links, share dialog accents).
  "love": "#ef5350", // Love / heart indicator (emotional emphasis only).
  "terminal": "#4A464F", // Background for terminal / code blocks (log views, code samples).
  "navigation": "#141417", // App bar / top toolbar background.
  "navigation-home": "#0e0f10", // "Home" navigation-state background (typically a touch darker than `navigation`).
};

export const variablesDark = {
  "btn-height": "34px", // Fixed pixel height for `VBtn`.
  "table-row-height": "44px", // Fixed pixel height for `VDataTable` body rows.
  "table-header-height": "44px", // Fixed pixel height for `VDataTable` header rows.
  "border-color": "#FFFFFF", // Divider and outlined-variant border color (text fields, cards, alerts).
  "border-opacity": 0.05, // Alpha applied to `border-color`.
  "high-emphasis-opacity": 0.96, // Body-text alpha (primary text on `surface`).
  "medium-emphasis-opacity": 0.88, // Secondary-text alpha (captions, helper text).
  "label-opacity": 0.67, // Floating-label and helper-text alpha on inputs.
  "disabled-opacity": 0.75, // Alpha applied to disabled controls.
  "idle-opacity": 0.1, // Idle-state alpha for toggle controls (off-state `VSwitch`, etc.).
  "fill-opacity": 0.04, // Background fill alpha for `solo-filled` inputs.
  "hover-opacity": 0.019, // Hover-overlay alpha on list items, buttons, and rows.
  "focus-opacity": 0.022, // Keyboard-focus overlay alpha on inputs and rows.
  "selected-opacity": 0.08, // Active/selected-state overlay alpha on list items and rows.
  "activated-opacity": 0, // Activated-state overlay alpha (currently no overlay).
  "pressed-opacity": 0.16, // Press/click overlay alpha on buttons and rows.
  "dragged-opacity": 0.08, // Dragged-state overlay alpha for sortable / draggable rows.
  "overlay-color": "#131313", // Scrim color behind dialogs, menus, and the navigation drawer.
  "overlay-opacity": 0.54, // Scrim alpha applied to `overlay-color`.
  "theme-kbd": "#212529", // `<kbd>` block background.
  "theme-on-kbd": "#FFFFFF", // `<kbd>` block foreground (text/icon color).
  "theme-code": "#343434", // Inline `<code>` background (distinct from the `terminal` color token, which paints code-block surfaces).
  "theme-on-code": "#CCCCCC", // Inline `<code>` foreground.
};

export const light = "light";

export const colorsLight = {
  "background": "#FFFFFF",
  "surface": "#FFFFFF",
  "on-surface": "#1e1e1f",
  "surface-bright": "#FFFFFF",
  "surface-light": "#EEEEEE",
  "surface-variant": "#424242",
  "on-surface-variant": "#EEEEEE",
  "card": "#a8a8a8",
  "button": "#474b4d",
  "switch": "#727272",
  "table": "#dddcda",
  "primary": "#1867C0",
  "highlight": "#3d3f40",
  "selected": "#c3c3c3",
  "secondary": "#e2e7ee",
  "secondary-light": "#eceff4",
  "accent": "#F2F5FA",
  "error": "#BF616A",
  "danger": "#BF616A",
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
  "terminal": "#4C566A",
  "navigation": "#e7ebf1",
  "navigation-home": "#dde3eb",
};

export const variablesLight = {
  ...variablesDark,
  ...{
    "border-color": "#000000",
    "border-opacity": 0.12,
    "high-emphasis-opacity": 0.95,
    "medium-emphasis-opacity": 0.75,
    "label-opacity": 0.8,
    "disabled-opacity": 0.75,
    "idle-opacity": 0.1,
    "fill-opacity": 0.04,
    "hover-opacity": 0.019,
    "focus-opacity": 0.022,
    "selected-opacity": 0.08,
    "activated-opacity": 0,
    "pressed-opacity": 0.16,
    "dragged-opacity": 0.08,
    "overlay-color": "#f2f2f2",
    "overlay-opacity": 0.42,
    "theme-kbd": "#212529",
    "theme-on-kbd": "#FFFFFF",
    "theme-code": "#F5F5F5",
    "theme-on-code": "#000000",
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
