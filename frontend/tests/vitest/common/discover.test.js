import { describe, it, expect } from "vitest";
import {
  groupPhotosByYear,
  isPastYearPhoto,
  discoverSearchParams,
  pageAdvance,
  pastYears,
  randomSearchParams,
  startOfCurrentYear,
  viewerToday,
} from "common/discover";

describe("common/discover", () => {
  describe("viewerToday", () => {
    it("uses the IANA zone from Settings rather than UTC", () => {
      const now = new Date("2026-01-01T04:30:00.000Z");
      expect(viewerToday("UTC", now)).toEqual({ year: 2026, month: 1, day: 1 });
      expect(viewerToday("America/Los_Angeles", now)).toEqual({ year: 2025, month: 12, day: 31 });
    });

    it("falls back to the browser zone for Local and invalid zones", () => {
      const now = new Date("2026-06-15T12:00:00.000Z");
      const local = viewerToday("Local", now);
      expect(local.year).toBeGreaterThan(2000);
      expect(local.month).toBeGreaterThanOrEqual(1);
      expect(local.month).toBeLessThanOrEqual(12);
      expect(viewerToday("Not/AZone", now)).toEqual(viewerToday("Local", now));
    });
  });

  describe("pastYears", () => {
    it("drops the current year, unknown years, and sorts newest first", () => {
      expect(pastYears([2026, 2024, "2019", 0, -1, 2025], 2026)).toEqual([2025, 2024, 2019]);
    });

    it("returns an empty list when no indexed years are available", () => {
      expect(pastYears([], 2026)).toEqual([]);
      expect(pastYears(undefined, 2026)).toEqual([]);
    });
  });

  describe("discoverSearchParams", () => {
    it("filters this day in past years and excludes the current year", () => {
      const now = new Date("2026-10-02T15:00:00.000Z");
      expect(
        discoverSearchParams({
          mode: "day",
          timeZone: "UTC",
          years: [2026, 2024, 2021],
          now,
          count: 50,
        })
      ).toEqual({
        count: 50,
        offset: 0,
        merged: true,
        order: "newest",
        month: "10",
        day: "2",
        year: "2024|2021",
      });
    });

    it("omits the day filter for the month tab", () => {
      const now = new Date("2026-10-02T15:00:00.000Z");
      const params = discoverSearchParams({
        mode: "month",
        timeZone: "UTC",
        years: [2024],
        now,
      });
      expect(params.month).toBe("10");
      expect(params.day).toBeUndefined();
      expect(params.year).toBe("2024");
    });

    it("falls back to a before filter when no past years are indexed", () => {
      const now = new Date("2026-03-01T00:00:00.000Z");
      const params = discoverSearchParams({
        mode: "day",
        timeZone: "UTC",
        years: [2026],
        now,
      });
      expect(params.year).toBeUndefined();
      expect(params.before).toBe(startOfCurrentYear(2026));
      expect(params.before).toBe("2026-01-01");
    });

    it("keeps February 29 as an exact calendar day", () => {
      const now = new Date("2024-02-29T12:00:00.000Z");
      const params = discoverSearchParams({
        mode: "day",
        timeZone: "UTC",
        years: [2020, 2016],
        now,
      });
      expect(params.month).toBe("2");
      expect(params.day).toBe("29");
    });
  });

  describe("groupPhotosByYear", () => {
    it("groups newest first, skips empty and current years, and counts photos", () => {
      const groups = groupPhotosByYear(
        [
          { UID: "a", Year: 2024 },
          { UID: "b", Year: 2021 },
          { UID: "c", Year: 2024 },
          { UID: "now", Year: 2026 },
          { UID: "unknown", Year: -1 },
        ],
        2026
      );

      expect(groups.map((group) => group.year)).toEqual([2024, 2021]);
      expect(groups[0].count).toBe(2);
      expect(groups[0].yearsAgo).toBe(2);
      expect(groups[1].yearsAgo).toBe(5);
    });

    it("returns no sections when nothing matches a past year", () => {
      expect(groupPhotosByYear([{ Year: 2026 }, { Year: 0 }], 2026)).toEqual([]);
      expect(groupPhotosByYear([], 2026)).toEqual([]);
    });
  });

  describe("isPastYearPhoto", () => {
    it("rejects unknown, missing, and current-year dates", () => {
      expect(isPastYearPhoto({ Year: 2020 }, 2026)).toBe(true);
      expect(isPastYearPhoto({ Year: 2026 }, 2026)).toBe(false);
      expect(isPastYearPhoto({ Year: -1 }, 2026)).toBe(false);
      expect(isPastYearPhoto({}, 2026)).toBe(false);
    });
  });

  describe("pageAdvance", () => {
    it("advances by the returned page size and stops on a short page", () => {
      expect(pageAdvance({ count: 100, limit: 100, offset: 0 }, 0, 156)).toEqual({ complete: false, offset: 100 });
      expect(pageAdvance({ count: 40, limit: 100, offset: 100 }, 100, 156)).toEqual({ complete: true, offset: 100 });
      expect(pageAdvance({ count: 0, limit: 0, offset: 0 }, 0, 156)).toEqual({ complete: true, offset: 0 });
    });
  });

  describe("randomSearchParams", () => {
    it("requests a finite random set", () => {
      expect(randomSearchParams({ count: 24, quality: "3", public: "true" })).toEqual({
        count: 24,
        offset: 0,
        merged: true,
        order: "random",
        quality: "3",
        public: "true",
      });
    });
  });
});
