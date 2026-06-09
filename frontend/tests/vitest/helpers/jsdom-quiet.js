// Quiet jsdom's known false-positive CSS-parser warnings on
// Vuetify-flavored stylesheets, regardless of where the stylesheet
// is authored. Vuetify components, the Vuetify base styles, and
// PhotoPrism's own CSS all share the same surface (`.v-application`,
// `.v-overlay`, `--v-theme-*`, ...), and jsdom's parser rejects the
// modern rules they use (e.g. @layer, container queries, @scope) the
// same way in all of them. This module treats those warnings as a
// single category of noise and suppresses them across that surface.
//
// jsdom 29 emits a `jsdomError` of `type: "css-parsing"` whenever its
// CSS parser rejects a rule. The default
// `(new VirtualConsole()).forwardTo(console)` forwarder writes a bare
// "Could not parse CSS stylesheet" line per occurrence, with no
// origin, no offending text, and no parser stack — so it cannot point
// to a real regression even when one exists.
//
// This module replaces that default forwarder with one that:
//   - Drops "css-parsing" errors whose stylesheet text contains any
//     Vuetify-flavored marker (the surface jsdom is known to reject
//     non-actionably). The check is heuristic and intentionally broad
//     enough to cover Vuetify, Vuetify-derived components, and our
//     own Vuetify-themed styles.
//   - For every other "css-parsing" error (i.e. CSS that doesn't even
//     touch the Vuetify surface), prints the message AND the
//     underlying parser cause, so an unrelated regression surfaces
//     with strictly more detail than the default forwarder produced.
//   - Forwards non-CSS jsdomErrors using the same shape as jsdom 29's
//     default forwardTo handler.
//
// Why this reaches into `window._virtualConsole`: Vitest does not
// expose a hook for the JSDOM virtualConsole. The documented config
// path (`environmentOptions.jsdom.virtualConsole`) is unusable under
// `pool: "vmForks"` because options are sent to the worker via
// structured clone and a VirtualConsole instance is not cloneable
// (its EventEmitter listeners are functions). Mutating
// `window._virtualConsole` after JSDOM initialization is a test-only
// workaround and is acknowledged here as such.
//
// MUST be imported before any module that loads CSS so the filter is
// installed when vitest injects stylesheets at import-evaluation time.

// Substrings whose presence in a stylesheet's text means the warning
// is from the Vuetify-flavored surface jsdom rejects non-actionably.
// Matching is by inclusion, not authorship — PhotoPrism CSS that
// extends Vuetify will (correctly) match.
const VUETIFY_FLAVORED_MARKERS = ["--v-theme-", "--v-medium-emphasis-opacity", ".v-application", ".v-locale-provider", ".v-overlay"];

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

if (typeof window !== "undefined" && window._virtualConsole) {
  const vc = window._virtualConsole;
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
