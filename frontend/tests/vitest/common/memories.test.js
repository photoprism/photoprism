import { describe, expect, it } from "vitest";
import { todayMemoriesFilter } from "common/memories";

describe("common/memories", () => {
  it("builds a photo search filter for the same calendar day", () => {
    expect(todayMemoriesFilter(new Date("2026-05-21T14:30:00Z"))).toEqual({
      photo: "true",
      month: "5",
      day: "21",
      before: "2025-12-31",
      order: "newest",
    });
  });
});
