import { describe, it, expect } from "vitest";
import Selection, { selectionRange } from "common/selection";

describe("common/selection selectionRange", () => {
  const older = { UID: "older" };
  const newer = { UID: "newer" };
  const all = [newer, older];

  it("keeps the Selection class export", () => {
    const selection = new Selection();
    expect(selection.isEmpty()).toBe(true);
  });

  it("keeps the local index when no wider list is provided", () => {
    expect(selectionRange([older], 0, [])).toEqual({ index: 0, photos: [older] });
  });

  it("maps a section-local photo onto the full result order", () => {
    expect(selectionRange([older], 0, all)).toEqual({ index: 1, photos: all });
  });

  it("returns null when the photo is missing from the full list", () => {
    expect(selectionRange([{ UID: "missing" }], 0, all)).toBeNull();
  });
});
