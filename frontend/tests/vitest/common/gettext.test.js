// Targets the positional-interpolation bridge in common/gettext.js (Tp), which renders backend
// notification messages — addressed by their English source id with Go printf placeholders — in
// the current UI language. The vue3-gettext runtime itself is third-party and not retested here.
import { describe, it, expect } from "vitest";

import { Tp } from "common/gettext";

describe("common/gettext", () => {
  describe("Tp", () => {
    it("substitutes a single %s placeholder", () => {
      expect(Tp("Indexing files in %s", ["/photos"])).toBe("Indexing files in /photos");
    });
    it("substitutes a single %d placeholder", () => {
      expect(Tp("Indexing completed in %d s", [5])).toBe("Indexing completed in 5 s");
    });
    it("substitutes multiple ordered placeholders of mixed type", () => {
      expect(Tp("%d entries added to %s", [3, "Holiday"])).toBe("3 entries added to Holiday");
    });
    it("substitutes repeated same-type placeholders positionally", () => {
      expect(Tp("Removed %d files and %d photos", [10, 4])).toBe("Removed 10 files and 4 photos");
    });
    it("returns the message unchanged when there are no params", () => {
      expect(Tp("Album created", [])).toBe("Album created");
      expect(Tp("Album created")).toBe("Album created");
    });
    it("un-escapes %% to a literal percent and does not consume a param", () => {
      expect(Tp("100%% complete in %d s", [5])).toBe("100% complete in 5 s");
    });
    it("leaves a literal percent in translated text intact (no space flag, real verbs only)", () => {
      // A translator may write a bare "100% ..." in a localized backend message; "% " must not be
      // mistaken for a printf verb and splice a param mid-word.
      expect(Tp("100% done in %d s", [5])).toBe("100% done in 5 s");
      expect(Tp("Zu 50% fertig", [])).toBe("Zu 50% fertig");
    });
    it("leaves a placeholder literal when its param is missing", () => {
      expect(Tp("Indexing files in %s", [])).toBe("Indexing files in %s");
    });
    it("renders an empty string for null or undefined param values", () => {
      expect(Tp("Indexing files in %s", [null])).toBe("Indexing files in ");
    });
    it("returns an empty string for a null message", () => {
      expect(Tp(null, [1])).toBe("");
    });
  });
});
