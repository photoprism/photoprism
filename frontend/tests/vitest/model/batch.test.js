import { describe, it, expect } from "vitest";
import "../fixtures";
import { Batch } from "model/batch";
import { Photo } from "model/photo";

describe("model/batch", () => {
  it("should return defaults", () => {
    const b = new Batch();
    const d = b.getDefaults();
    expect(Array.isArray(d.models)).toBe(true);
    expect(d.values).toEqual({});
    expect(Array.isArray(d.selection)).toBe(true);
  });

  it("should return default form data", () => {
    const b = new Batch();
    const f = b.getDefaultFormData();
    const expectedKeys = [
      "Title",
      "DetailsSubject",
      "Caption",
      "Day",
      "Month",
      "Year",
      "TimeZone",
      "Country",
      "Altitude",
      "Lat",
      "Lng",
      "DetailsArtist",
      "DetailsCopyright",
      "DetailsLicense",
      "DetailsKeywords",
      "Type",
      "Scan",
      "Private",
      "Favorite",
      "Panorama",
      "Iso",
      "FocalLength",
      "FNumber",
      "Exposure",
      "CameraID",
      "LensID",
      "Albums",
      "Labels",
    ];

    expect(Object.keys(f).sort()).toEqual(expectedKeys.sort());
    expect(f.Albums).toEqual({ action: "none", mixed: false, items: [] });
    expect(f.Labels).toEqual({ action: "none", mixed: false, items: [] });
  });

  it("should set selections", () => {
    const b = new Batch();
    b.setSelections(["pt5y3865st5p3k5l", "pt5y3863oyip9a2d", "pt5y38631t2s9p0a"]);
    expect(b.selection).toEqual([]);
  });

  it("should report selection state for a given id", () => {
    const b = new Batch();
    b.models = [new Photo({ UID: "pt5y3865st5p3k5l" }), new Photo({ UID: "pt5y3863oyip9a2d" })];
    b.setSelections(["pt5y3865st5p3k5l", "pt5y3863oyip9a2d"]);
    expect(b.isSelected("pt5y3865st5p3k5l")).toBe(true);
    // toggle one and check again
    b.toggle("pt5y3865st5p3k5l");
    expect(b.isSelected("pt5y3865st5p3k5l")).toBe(false);
    // unknown id returns null per implementation
    expect(b.isSelected("pt00000000000000")).toBeNull();
  });

  it("should toggle and toggleAll", () => {
    const b = new Batch();
    b.models = [
      new Photo({ UID: "pt5y3865st5p3k5l" }),
      new Photo({ UID: "pt5y3863oyip9a2d" }),
      new Photo({ UID: "pt5y38631t2s9p0a" }),
    ];
    b.setSelections(["pt5y3865st5p3k5l", "pt5y3863oyip9a2d", "pt5y38631t2s9p0a"]);
    expect(b.getLengthOfAllSelected()).toBe(3);
    b.toggle("pt5y3863oyip9a2d");
    expect(b.isSelected("pt5y3863oyip9a2d")).toBe(false);
    expect(b.getLengthOfAllSelected()).toBe(2);

    b.toggleAll(false);
    expect(b.getLengthOfAllSelected()).toBe(0);

    b.toggleAll(true);
    expect(b.getLengthOfAllSelected()).toBe(3);
  });

  it("should call save and update values from response", async () => {
    const b = new Batch();
    const selection = ["pt20fg34bbwdm2ld", "pt20fg2qikiy7zax"];
    const values = { Title: { value: "New" } };

    const existing = new Photo({ UID: "pt20fg34bbwdm2ld", Title: "Old" });
    b.models = [existing];

    const { Mock } = await import("../fixtures");
    Mock.onPost("api/v1/batch/photos/edit", { photos: selection, values }).reply(200, {
      models: [
        { UID: "pt20fg34bbwdm2ld", Title: "Updated" },
        { UID: "pt20fg2qikiy7zb0", Title: "New" },
      ],
      values: { Title: { value: "Saved" } },
    });

    const result = await b.save(selection, values);
    expect(result).toBe(b);
    expect(b.values).toEqual({ Title: { value: "Saved" } });
    expect(b.models.find((m) => m.UID === "pt20fg34bbwdm2ld").Title).toBe("Updated");
    expect(b.models.some((m) => m.UID === "pt20fg2qikiy7zb0")).toBe(true);
  });

  it("should load data (models and values) via load", async () => {
    const b = new Batch();
    const selection = ["pt5y3865st5p3k5l", "pt5y3863oyip9a2d", "pt5y38631t2s9p0a"];

    // Response should include models and values
    const { Mock } = await import("../fixtures");
    Mock.onPost("api/v1/batch/photos/edit", { photos: selection }).reply(
      200,
      {
        models: [
          { ID: 1, UID: "pt5y3865st5p3k5l", Title: "A" },
          { ID: 2, UID: "pt5y3863oyip9a2d", Title: "B" },
        ],
        values: { Title: { mixed: true } },
      },
      { "Content-Type": "application/json; charset=utf-8" }
    );

    const result = await b.load(selection);
    expect(result).toBe(b);

    expect(Array.isArray(b.models)).toBe(true);
    expect(b.models.length).toBe(2);
    expect(b.models[0]).toBeInstanceOf(Photo);
    expect(b.models[1]).toBeInstanceOf(Photo);
    expect(b.values).toEqual({ Title: { mixed: true } });
    expect(b.selection).toEqual([
      { id: "pt5y3865st5p3k5l", selected: true },
      { id: "pt5y3863oyip9a2d", selected: true },
    ]);
  });

  it("should drop selection ids that no longer have editable models", () => {
    const b = new Batch();
    b.models = [new Photo({ UID: "pt5y3865st5p3k5l" }), new Photo({ UID: "pt5y3863oyip9a2d" })];

    b.setSelections([
      "pt5y3865st5p3k5l",
      "pt5y38631t2s9p0a",
      "pt5y3863oyip9a2d",
      "pt5y3863kb9amo1x",
    ]);

    expect(b.selection).toEqual([
      { id: "pt5y3865st5p3k5l", selected: true },
      { id: "pt5y3863oyip9a2d", selected: true },
    ]);
    expect(b.selectionById.get("pt5y38631t2s9p0a")).toBeUndefined();
    expect(b.getLengthOfAllSelected()).toBe(2);
  });
});
