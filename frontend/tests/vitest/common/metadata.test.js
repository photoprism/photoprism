import { describe, expect, it } from "vitest";
import "../fixtures";
import { defaultMetadataLayout, metadataLayoutRequiresDetails, metadataViewRequiresDetails, MetadataView } from "common/metadata";

describe("common/metadata", () => {
  it("flags keyword layouts as requiring photo details", () => {
    expect(metadataLayoutRequiresDetails(["caption", "keywords"])).toBe(true);
    expect(metadataLayoutRequiresDetails(["caption", "date"])).toBe(false);
  });

  it("detects when a configured cards layout requires details", () => {
    const settings = {
      display: {
        metadata: {
          cards: defaultMetadataLayout(MetadataView.Cards),
        },
      },
    };

    expect(metadataViewRequiresDetails(settings, MetadataView.Cards)).toBe(true);
  });

  it("does not require details for the default list layout", () => {
    const settings = {
      display: {
        metadata: {
          list: defaultMetadataLayout(MetadataView.List),
        },
      },
    };

    expect(metadataViewRequiresDetails(settings, MetadataView.List)).toBe(false);
  });

  it("detects when a configured lightbox layout requires details", () => {
    const settings = {
      display: {
        metadata: {
          lightbox: defaultMetadataLayout(MetadataView.Lightbox),
        },
      },
    };

    expect(metadataViewRequiresDetails(settings, MetadataView.Lightbox)).toBe(true);
  });
});
