import { describe, it, expect, vi } from "vitest";

vi.mock("component/photo/timeline.vue", () => ({
  default: { name: "PPhotoTimeline", template: "<div />" },
}));

import PPagePhotos from "page/photos.vue";

// baseFilter returns the Photos page filter shape used by query helpers.
function baseFilter(values = {}) {
  return {
    country: "",
    camera: 0,
    lens: 0,
    label: "",
    latlng: "",
    year: 0,
    month: 0,
    day: 0,
    color: "",
    order: "newest",
    reverse: false,
    q: "",
    ...values,
  };
}

// pageContext returns the minimal Options API context required by data().
function pageContext(query = {}) {
  return {
    $route: { name: "all", query },
    $config: {
      getSettings: () => ({ features: { edit: true, places: true, private: false, review: false } }),
      allow: () => true,
      deny: () => false,
      feature: () => true,
    },
    $clipboard: { selection: [] },
    staticFilter: null,
    getViewType: () => "cards",
    sortOrder: () => "newest",
    sortReverse: () => false,
  };
}

describe("page/photos.vue timeline rail", () => {
  it("is enabled only for calendar-capable non-embedded md-and-up layouts", () => {
    const timelineEnabled = PPagePhotos.computed.timelineEnabled;
    const base = {
      embedded: false,
      $vuetify: { display: { mdAndUp: true } },
      $config: { feature: () => true, allow: () => true },
    };

    expect(timelineEnabled.call(base)).toBe(true);
    expect(timelineEnabled.call({ ...base, embedded: true })).toBe(false);
    expect(timelineEnabled.call({ ...base, $vuetify: { display: { mdAndUp: false } } })).toBe(false);
    expect(timelineEnabled.call({ ...base, $config: { feature: () => false, allow: () => true } })).toBe(false);
    expect(timelineEnabled.call({ ...base, $config: { feature: () => true, allow: () => false } })).toBe(false);
  });

  it("reserves the timeline rail only while the mounted child reports visible content", () => {
    const timelineVisible = PPagePhotos.computed.timelineVisible;
    const stub = { timelineEnabled: true, timelineRailVisible: true };

    expect(timelineVisible.call(stub)).toBe(true);
    expect(timelineVisible.call({ ...stub, timelineRailVisible: false })).toBe(false);
    expect(timelineVisible.call({ ...stub, timelineEnabled: false })).toBe(false);
  });

  it("builds timeline params from active filters without result-shaping keys", () => {
    const params = PPagePhotos.computed.timelineParams.call({
      filter: baseFilter({
        country: "de",
        year: 2024,
        month: 5,
        day: 21,
        public: "true",
        quality: "3",
        q: "cats",
        reverse: true,
      }),
      staticFilter: {
        favorite: true,
        private: "true",
        public: "",
        hidden: false,
        offset: 100,
      },
    });

    expect(params).toEqual({
      country: "de",
      year: 2024,
      month: 5,
      day: 21,
      public: "",
      quality: "3",
      q: "cats",
      favorite: true,
      private: "true",
      hidden: false,
    });
  });

  it("initializes the day filter from the route query", () => {
    const data = PPagePhotos.data.call(pageContext({ year: "2024", month: "5", day: "21" }));

    expect(data.filter.year).toBe(2024);
    expect(data.filter.month).toBe(5);
    expect(data.filter.day).toBe(21);
  });

  it("syncs the day filter when the route changes", () => {
    const stub = {
      $view: { isActive: () => true, focus: vi.fn() },
      $refs: { page: {} },
      $route: { name: "all", query: { year: "2024", month: "5", day: "21" } },
      $config: {
        getSettings: () => ({ features: { private: false, review: false } }),
      },
      staticFilter: null,
      filter: baseFilter(),
      settings: { view: "cards" },
      routeName: "all",
      lastFilter: {},
      getViewType: () => "cards",
      sortOrder: () => "newest",
      sortReverse: () => false,
      search: vi.fn(),
    };

    PPagePhotos.watch.$route.call(stub);

    expect(stub.filter.year).toBe(2024);
    expect(stub.filter.month).toBe(5);
    expect(stub.filter.day).toBe(21);
    expect(stub.search).toHaveBeenCalledTimes(1);
  });

  it("lets the timeline clear day through updateQuery", () => {
    const replace = vi.fn();
    const stub = {
      filter: baseFilter({ q: "cats", day: 21 }),
      settings: { view: "cards" },
      loading: false,
      $route: { query: { view: "cards", q: "cats", order: "newest", day: "21" } },
      $router: { replace },
      updateFilter: PPagePhotos.methods.updateFilter,
    };

    const changed = PPagePhotos.methods.updateQuery.call(stub, { day: 0 });

    expect(changed).toBe(true);
    expect(stub.filter.day).toBe(0);
    expect(replace).toHaveBeenCalledWith({ query: { view: "cards", order: "newest", q: "cats" } });
  });

  it("updates timeline rail visibility from the child component", () => {
    const stub = { timelineRailVisible: true };

    PPagePhotos.methods.setTimelineVisible.call(stub, false);
    expect(stub.timelineRailVisible).toBe(false);

    PPagePhotos.methods.setTimelineVisible.call(stub, true);
    expect(stub.timelineRailVisible).toBe(true);
  });

  it("clears hidden day filters when year or month changes", () => {
    const replace = vi.fn();
    const stub = {
      filter: baseFilter({ year: 2024, month: 5, day: 21 }),
      settings: { view: "cards" },
      loading: false,
      $route: { query: { view: "cards", order: "newest", year: "2024", month: "5", day: "21" } },
      $router: { replace },
      updateFilter: PPagePhotos.methods.updateFilter,
    };

    const changed = PPagePhotos.methods.updateQuery.call(stub, { year: 2025 });

    expect(changed).toBe(true);
    expect(stub.filter.year).toBe(2025);
    expect(stub.filter.day).toBe(0);
    expect(replace).toHaveBeenCalledWith({ query: { view: "cards", order: "newest", year: 2025, month: 5 } });
  });

  it("includes day in photo search params without adding thumbnail sources", () => {
    const params = PPagePhotos.methods.searchParams.call({
      searchCount: () => 156,
      offset: 0,
      filter: baseFilter({ year: 2024, month: 5, day: 21 }),
      staticFilter: null,
    });

    expect(params).toMatchObject({
      count: 156,
      offset: 0,
      merged: true,
      year: 2024,
      month: 5,
      day: 21,
    });
  });
});
