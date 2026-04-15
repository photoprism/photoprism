import { describe, expect, it } from "vitest";
import { choosePackedThumbSize, layoutPackedRows } from "common/packed";

describe("common/packed", () => {
  it("returns empty rows when input is invalid", () => {
    expect(layoutPackedRows([], 1000)).toEqual([]);
    expect(layoutPackedRows(null, 1000)).toEqual([]);
    expect(layoutPackedRows([{ Width: 10, Height: 10 }], 0)).toEqual([]);
  });

  it("keeps justified rows aligned to the measured container width", () => {
    const rows = layoutPackedRows(
      [
        { Width: 400, Height: 200 },
        { Width: 400, Height: 200 },
        { Width: 400, Height: 200 },
      ],
      1000,
      { gutter: 10, targetRowHeight: 200 }
    );

    expect(rows).toHaveLength(1);
    expect(rows[0].items.map((item) => item.index)).toEqual([0, 1, 2]);
    expect(rows[0].items.reduce((sum, item) => sum + item.width, 0) + 20).toBe(1000);
  });

  it("does not stretch the last row to fill the container", () => {
    const rows = layoutPackedRows(
      [
        { Width: 400, Height: 200 },
        { Width: 400, Height: 200 },
      ],
      1000,
      { gutter: 10, targetRowHeight: 200 }
    );

    expect(rows).toHaveLength(1);
    expect(rows[0].items.reduce((sum, item) => sum + item.width, 0) + 10).toBeLessThan(1000);
    expect(rows[0].height).toBe(200);
  });

  it("chooses fit thumbnails based on rendered size", () => {
    expect(choosePackedThumbSize(700, 300, false)).toBe("fit_720");
    expect(choosePackedThumbSize(900, 300, false)).toBe("fit_1280");
  });
});
