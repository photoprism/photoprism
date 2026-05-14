import { describe, expect, it, vi } from "vitest";

import PPagePhotos from "page/photos.vue";

function createConfig() {
  return {
    allow: vi.fn(() => true),
    deny: vi.fn(() => false),
    getSettings: vi.fn(() => ({
      features: {
        edit: true,
        places: true,
        private: false,
        review: false,
      },
    })),
  };
}

describe("page/photos.vue day filters", () => {
  it("reads the day query into the initial filter state", () => {
    const data = PPagePhotos.data.call({
      $clipboard: {
        selection: [],
      },
      $route: {
        name: "browse",
        query: {
          day: "14",
          month: "5",
          year: "2024",
        },
      },
      $config: createConfig(),
      getViewType: () => "cards",
      sortOrder: () => "newest",
      sortReverse: () => false,
    });

    expect(data.filter.day).toBe(14);
    expect(data.filter.month).toBe(5);
    expect(data.filter.year).toBe(2024);
  });

  it("preserves the day filter when building route queries", () => {
    const replace = vi.fn();
    const stub = {
      filter: {
        camera: 0,
        color: "",
        country: "",
        day: 14,
        label: "",
        latlng: "",
        lens: 0,
        month: 5,
        order: "oldest",
        q: "",
        reverse: false,
        year: 2024,
      },
      loading: false,
      settings: {
        view: "cards",
      },
      updateFilter: PPagePhotos.methods.updateFilter,
      $route: {
        query: {},
      },
      $router: {
        replace,
      },
    };

    const changed = PPagePhotos.methods.updateQuery.call(stub);

    expect(changed).toBe(true);
    expect(replace).toHaveBeenCalledWith({
      query: {
        view: "cards",
        year: 2024,
        month: 5,
        day: 14,
        order: "oldest",
      },
    });
  });
});
