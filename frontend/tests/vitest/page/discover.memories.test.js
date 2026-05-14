import { describe, expect, it } from "vitest";

import PMemories from "page/discover/memories.vue";

describe("page/discover/memories.vue", () => {
  it("groups past memories by year and limits previews", () => {
    const groups = PMemories.methods.buildGroups.call(
      {
        memoryDate: { year: 2026 },
      },
      [
        { UID: "a", Year: 2025 },
        { UID: "b", Year: 2024 },
        { UID: "c", Year: 2025 },
        { UID: "d", Year: 2026 },
        { UID: "e", Year: 2025 },
        { UID: "f", Year: 0 },
        { UID: "g", Year: 2025 },
        { UID: "h", Year: 2025 },
      ]
    );

    expect(groups.map((group) => group.year)).toEqual([2025, 2024]);
    expect(groups[0].count).toBe(5);
    expect(groups[0].preview.map((photo) => photo.UID)).toEqual(["a", "c", "e", "g"]);
  });

  it("builds browse queries with month, day, and optional year", () => {
    const stub = {
      memoryDate: { month: 5, day: 14 },
    };

    expect(PMemories.methods.browseQuery.call(stub)).toEqual({
      month: 5,
      day: 14,
      order: "oldest",
    });

    expect(PMemories.methods.browseQuery.call(stub, 2024)).toEqual({
      month: 5,
      day: 14,
      order: "oldest",
      year: 2024,
    });
  });
});
