import { describe, it, expect } from "vitest";
import "../fixtures";

import Model from "model/model";

// Option is a minimal model with -1/0/1 option fields.
class Option extends Model {
  getDefaults() {
    return { On: 0, Off: 0 };
  }
}

describe("model/model", () => {
  describe("flagEnabled", () => {
    it("follows the default for 0 and a missing value", () => {
      const m = new Option();
      expect(m.flagEnabled("On")).toBe(true);
      expect(m.flagEnabled("Off", false)).toBe(false);
      expect(m.flagEnabled("Missing")).toBe(true);
      expect(m.flagEnabled("Missing", false)).toBe(false);
    });

    it("reports explicit values regardless of the default", () => {
      const m = new Option({ On: -1, Off: 1 });
      expect(m.flagEnabled("On")).toBe(false);
      expect(m.flagEnabled("Off", false)).toBe(true);
    });
  });

  describe("setFlag", () => {
    it("keeps an untouched default when switched back", () => {
      const m = new Option();
      m.setFlag("On", false);
      expect(m.On).toBe(-1);
      m.setFlag("On", true);
      expect(m.On).toBe(0);
      m.setFlag("Off", true, false);
      expect(m.Off).toBe(1);
      m.setFlag("Off", false, false);
      expect(m.Off).toBe(0);
      expect(m.getValues(true)).toEqual({});
    });

    it("stores an explicit value when the saved state differs", () => {
      const m = new Option({ On: -1, Off: 1 });
      m.setFlag("On", true);
      expect(m.On).toBe(1);
      m.setFlag("Off", false, false);
      expect(m.Off).toBe(-1);
      expect(m.getValues(true)).toEqual({ On: 1, Off: -1 });
    });
  });
});
