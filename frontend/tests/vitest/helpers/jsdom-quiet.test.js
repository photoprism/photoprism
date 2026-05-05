import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";

// Imported for its side effect: installs the jsdomError filter on
// window._virtualConsole. Already loaded transitively by setup.js, but
// we re-import here for clarity.
import "./jsdom-quiet";

describe("helpers/jsdom-quiet", () => {
  let errorSpy;

  beforeEach(() => {
    errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
  });

  afterEach(() => {
    errorSpy.mockRestore();
  });

  function emit(error) {
    window._virtualConsole.emit("jsdomError", error);
  }

  function makeError({ type, message, sheetText, cause } = {}) {
    const e = new Error(message ?? "boom");
    if (type !== undefined) e.type = type;
    if (sheetText !== undefined) e.sheetText = sheetText;
    if (cause !== undefined) e.cause = cause;
    return e;
  }

  it("drops css-parsing errors when the stylesheet contains Vuetify-flavored markers", () => {
    // Stylesheets on the Vuetify surface — whether shipped by Vuetify
    // itself or authored in PhotoPrism — are recognized by their
    // selectors / custom properties and treated as the same noise.
    emit(
      makeError({
        type: "css-parsing",
        message: "Could not parse CSS stylesheet",
        sheetText: ".v-application { --v-theme-primary: 0,0,0; }",
        cause: new Error("rule parse failed"),
      })
    );
    expect(errorSpy).not.toHaveBeenCalled();
  });

  it("logs css-parsing errors when the stylesheet has no Vuetify markers, including the underlying cause", () => {
    const cause = new Error("rule parse failed");
    emit(
      makeError({
        type: "css-parsing",
        message: "Could not parse CSS stylesheet",
        sheetText: ".photoprism-foo { color: red; }",
        cause,
      })
    );
    expect(errorSpy).toHaveBeenCalledTimes(1);
    const arg = errorSpy.mock.calls[0][0];
    expect(arg).toContain("Could not parse CSS stylesheet");
    expect(arg).toContain(cause.stack);
  });

  it("logs unhandled-exception jsdomErrors via the cause stack", () => {
    const cause = new Error("inner");
    emit(makeError({ type: "unhandled-exception", cause }));
    expect(errorSpy).toHaveBeenCalledWith(cause.stack);
  });

  it("logs other jsdomErrors via err.message", () => {
    emit(makeError({ message: "something else" }));
    expect(errorSpy).toHaveBeenCalledWith("something else");
  });
});
