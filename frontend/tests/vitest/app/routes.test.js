import { afterEach, describe, expect, it, vi } from "vitest";

vi.mock("page/photos.vue", () => ({ default: {} }));
vi.mock("page/albums.vue", () => ({ default: {} }));
vi.mock("page/album/photos.vue", () => ({ default: {} }));
vi.mock("page/places.vue", () => ({ default: {} }));
vi.mock("page/library/browse.vue", () => ({ default: {} }));
vi.mock("page/library/errors.vue", () => ({ default: {} }));
vi.mock("page/labels.vue", () => ({ default: {} }));
vi.mock("page/people.vue", () => ({ default: {} }));
vi.mock("page/library.vue", () => ({ default: {} }));
vi.mock("page/settings.vue", () => ({ default: {} }));
vi.mock("page/admin.vue", () => ({ default: {} }));
vi.mock("page/cluster.vue", () => ({ default: {} }));
vi.mock("page/auth/login.vue", () => ({ default: {} }));
vi.mock("page/discover.vue", () => ({ default: {} }));
vi.mock("page/about/about.vue", () => ({ default: {} }));
vi.mock("page/about/license.vue", () => ({ default: {} }));
vi.mock("page/help.vue", () => ({ default: {} }));
vi.mock("page/connect.vue", () => ({ default: {} }));

vi.mock("app/session", () => ({
  $config: {
    deny: vi.fn(() => false),
  },
  $session: {
    getDefaultRoute: vi.fn(() => "photos"),
    loginRequired: vi.fn(() => false),
  },
}));

describe("app/routes", () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it("configures memories to show photos from this day in previous years", async () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-05-11T12:00:00Z"));

    const { default: routes } = await import("app/routes");
    const route = routes.find((item) => item.name === "memories");
    const props = route.props();

    expect(route.path).toBe("/memories");
    expect(props.staticFilter).toEqual({
      photo: "true",
      month: "5",
      day: "11",
      before: "2025-05-11",
      order: "oldest",
    });
  });
});
