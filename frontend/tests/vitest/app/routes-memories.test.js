import { describe, it, expect, vi, afterEach } from "vitest";
import "../fixtures";
import routes from "app/routes";

const memoriesRoute = routes.find((r) => r.name === "memories");

describe("app/routes /memories", () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it("filters photos by the current month and day", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date(2026, 4, 25, 12, 0, 0));

    expect(memoriesRoute.path).toBe("/memories");
    expect(memoriesRoute.props()).toEqual({
      staticFilter: {
        quality: "3",
        month: "5",
        day: "25",
      },
    });
  });
});
