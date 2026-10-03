// jsdom-quiet replaces jsdom's default jsdomError forwarder: it drops "css-parsing" errors for
// Vuetify-flavored stylesheets, which jsdom rejects non-actionably (@layer, container queries),
// and logs other CSS errors with their parser cause. It uses Vitest's global JSDOM instance, so it
// MUST be imported before any module that loads CSS.

// Substrings whose presence in a stylesheet's text means the warning
// is from the Vuetify-flavored surface jsdom rejects non-actionably.
// Matching is by inclusion, not authorship — PhotoPrism CSS that
// extends Vuetify will (correctly) match.
const VUETIFY_FLAVORED_MARKERS = ["--v-theme-", "--v-medium-emphasis-opacity", ".v-application", ".v-locale-provider", ".v-overlay"];

// hasVuetifyFlavoredMarker reports whether a stylesheet's text contains any Vuetify-flavored marker.
function hasVuetifyFlavoredMarker(text) {
  if (typeof text !== "string" || text.length === 0) {
    return false;
  }
  for (const marker of VUETIFY_FLAVORED_MARKERS) {
    if (text.includes(marker)) {
      return true;
    }
  }
  return false;
}

// virtualConsole returns the virtual console of the JSDOM instance Vitest exposes as a global.
export function virtualConsole() {
  return globalThis.jsdom?.virtualConsole;
}

const vc = virtualConsole();

if (vc) {
  vc.removeAllListeners("jsdomError");
  vc.on("jsdomError", (err) => {
    if (!err) {
      return;
    }
    if (err.type === "css-parsing") {
      if (hasVuetifyFlavoredMarker(err.sheetText)) {
        return;
      }
      const cause = err.cause && err.cause.stack ? `\n${err.cause.stack}` : "";
      console.error(`${err.message}${cause}`);
      return;
    }
    if (err.type === "unhandled-exception" && err.cause) {
      console.error(err.cause.stack);
      return;
    }
    console.error(err.message);
  });
}
