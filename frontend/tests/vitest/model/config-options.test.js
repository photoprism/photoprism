import { describe, it, expect } from "vitest";
import "../fixtures";
import ConfigOptions from "model/config-options";

describe("model/config-options", () => {
  it("should get options defaults", () => {
    const values = {};
    const options = new ConfigOptions(values);
    const result = options.getDefaults();
    expect(result.Debug).toBe(false);
    expect(result.ReadOnly).toBe(false);
    expect(result.ThumbSize).toBe(0);
  });

  it("should test changed", () => {
    const values = {};
    const options = new ConfigOptions(values);
    expect(options.changed()).toBe(false);
  });

  it("should load options", async () => {
    const values = {};
    const options = new ConfigOptions(values);
    try {
      const response = await options.load();
      expect(response.success).toBe("ok");
    } catch (error) {
      // Vitest will fail the test if a promise rejects
      throw error;
    }
    expect(options.changed()).toBe(false);
  });

  it("should save options", async () => {
    const values = { Debug: true };
    const options = new ConfigOptions(values);
    try {
      const response = await options.save();
      expect(response.success).toBe("ok");
    } catch (error) {
      throw error;
    }
    expect(options.changed()).toBe(false);
  });
});
