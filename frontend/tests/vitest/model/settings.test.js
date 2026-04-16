import { describe, it, expect } from "vitest";
import "../fixtures";
import { defaultMetadataLayout, MetadataView } from "common/metadata";
import Settings from "model/settings";

describe("model/settings", () => {
  it("should default metadata layouts for cards, list, and lightbox", () => {
    const model = new Settings({});

    expect(model.display.metadata.cards).toEqual(defaultMetadataLayout(MetadataView.Cards));
    expect(model.display.metadata.list).toEqual(defaultMetadataLayout(MetadataView.List));
    expect(model.display.metadata.lightbox).toEqual(defaultMetadataLayout(MetadataView.Lightbox));
  });

  it("should backfill missing display settings", () => {
    const model = new Settings({ display: { originals: true } });

    expect(model.display.originals).toBe(true);
    expect(model.display.retinaLightbox).toBe(false);
    expect(model.display.retinaThumbnails).toBe(false);
    expect(model.display.metadata.cards).toEqual(defaultMetadataLayout(MetadataView.Cards));
    expect(model.display.metadata.list).toEqual(defaultMetadataLayout(MetadataView.List));
    expect(model.display.metadata.lightbox).toEqual(defaultMetadataLayout(MetadataView.Lightbox));
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
