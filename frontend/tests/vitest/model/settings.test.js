import { describe, it, expect } from "vitest";
import "../fixtures";
import Settings from "model/settings";

describe("model/settings", () => {
  it("should default image packing to false", () => {
    const model = new Settings({});

    expect(model.display.imagePacking).toBe(false);
  });

  it("should backfill missing image packing settings", () => {
    const model = new Settings({ display: { originals: true } });

    expect(model.display.imagePacking).toBe(false);
    expect(model.display.originals).toBe(true);
  });

  it("should return if key was changed", () => {
    const model = new Settings({ ui: { language: "de", scrollbar: false } });
    expect(model.changed("ui", "scrollbar")).toBe(false);
    expect(model.changed("ui", "language")).toBe(false);
  });

  it("should load settings", async () => {
    const model = new Settings({ ui: { language: "de", scrollbar: false } });
    const response = await model.load();
    expect(response["ui"]["scrollbar"]).toBe(false);
    expect(response["ui"]["language"]).toBe("de");
  });

  it("should save settings", async () => {
    const model = new Settings({ ui: { language: "de", scrollbar: false } });
    const response = await model.save();
    expect(response["ui"]["scrollbar"]).toBe(false);
    expect(response["ui"]["language"]).toBe("de");
  });
});
