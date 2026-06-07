import { describe, it, expect, vi } from "vitest";

import PPagePhotos from "page/photos.vue";

function configMock() {
  return {
    getSettings: vi.fn(() => ({
      features: {
        edit: true,
        places: true,
        private: false,
        review: false,
      },
    })),
    allow: vi.fn(() => true),
    deny: vi.fn(() => false),
  };
}

function newStub(query = {}) {
  return {
    $route: {
      name: "browse",
      query,
    },
    $config: configMock(),
    $clipboard: {
      selection: [],
    },
    staticFilter: {},
    embedded: false,
    getViewType: vi.fn(() => "cards"),
    sortOrder: vi.fn(() => "newest"),
    sortReverse: vi.fn(() => false),
  };
}

describe("page/photos.vue date filters", () => {
  it("initializes month and day from route query parameters", () => {
    const state = PPagePhotos.data.call(newStub({ month: "6", day: "7" }));

    expect(state.filter.month).toBe(6);
    expect(state.filter.day).toBe(7);
  });

  it("defaults the day filter to all days when no query parameter is set", () => {
    const state = PPagePhotos.data.call(newStub({ month: "6" }));

    expect(state.filter.month).toBe(6);
    expect(state.filter.day).toBe(0);
  });
});

