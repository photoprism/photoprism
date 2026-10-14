// Custom themes as sites define them in app.js for Vuetify 3, shortened: a dark one in the
// window.__THEMES__ format with its own variables, and a light one without variables.
export const customDark = {
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

export const customLight = {
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
