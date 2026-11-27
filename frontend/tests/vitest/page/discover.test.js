import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { shallowMount, flushPromises } from "@vue/test-utils";
import { createRouter, createMemoryHistory } from "vue-router";
import PPageDiscover from "page/discover.vue";
import PTabDiscoverPast from "page/discover/past.vue";
import PTabDiscoverRandom from "page/discover/random.vue";
import { Photo } from "model/photo";

function mountDiscover(path = "/discover") {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { name: "discover", path: "/discover", component: { template: "<div/>" } },
      { name: "discover_month", path: "/discover/month", component: { template: "<div/>" } },
      { name: "discover_random", path: "/discover/random", component: { template: "<div/>" } },
    ],
  });

  const tab = path.endsWith("/random") ? 2 : path.endsWith("/month") ? 1 : 0;

  return shallowMount(PPageDiscover, {
    props: { tab },
    global: {
      plugins: [router],
      mocks: {
        $view: { enter: vi.fn(), leave: vi.fn() },
        $gettext: (s) => s,
      },
      stubs: {
        "v-tabs": { template: "<div class='v-tabs'><slot /></div>" },
        "v-tab": { template: "<button class='v-tab'><slot /></button>" },
        "v-tabs-window": { template: "<div class='v-tabs-window'><slot /></div>" },
        "v-tabs-window-item": { template: "<div class='v-tabs-window-item'><slot /></div>" },
        "v-toolbar": { template: "<div class='v-toolbar'><slot /></div>" },
        "v-btn-toggle": { template: "<div class='v-btn-toggle'><slot /></div>" },
        "v-btn": { template: "<button class='v-btn'><slot /></button>" },
        "v-spacer": true,
      },
    },
  });
}

describe("page/discover.vue", () => {
  it("shows This Day, This Month, then Random", () => {
    const wrapper = mountDiscover();
    const tabs = wrapper.findAll(".v-tab");
    expect(tabs.map((tab) => tab.text())).toEqual(["This Day in the PastThis Day", "This Month in the PastThis Month", "Random"]);
    expect(wrapper.find("#tab-discover-day").text()).toContain("This Day in the Past");
    expect(wrapper.find("#tab-discover-day").text()).toContain("This Day");
    expect(wrapper.find("#tab-discover-month").text()).toContain("This Month");
    expect(wrapper.find("#tab-discover-random").exists()).toBe(true);
    expect(wrapper.find("#tab-discover-colors").exists()).toBe(false);
    wrapper.unmount();
  });

  it("keeps the view toggle visible on the day tab", () => {
    const wrapper = mountDiscover("/discover");
    expect(wrapper.find(".action-view-cards").exists()).toBe(true);
    expect(wrapper.find(".action-view-mosaic").exists()).toBe(true);
    expect(wrapper.find(".action-shuffle").exists()).toBe(false);
    wrapper.unmount();
  });

  it("shows Shuffle only on the Random tab", () => {
    const wrapper = mountDiscover("/discover/random");
    expect(wrapper.find(".action-shuffle").exists()).toBe(true);
    wrapper.unmount();
  });
});

describe("page/discover/past.vue", () => {
  let search;

  beforeEach(() => {
    search = vi.spyOn(Photo, "search").mockResolvedValue({
      models: [],
      count: 0,
      limit: 156,
      offset: 0,
    });
  });

  afterEach(() => {
    search.mockRestore();
  });

  it("searches by month and day for the day tab and shows an empty state", async () => {
    const wrapper = shallowMount(PTabDiscoverPast, {
      props: { mode: "day", view: "cards" },
      global: {
        mocks: {
          $view: { isHidden: (target) => target !== "PPageDiscover" },
          $event: { subscribe: vi.fn(() => 1), unsubscribe: vi.fn() },
          $clipboard: { selection: [] },
          $gettext: (s) => s,
          $config: {
            getTimeZone: () => "UTC",
            getSettings: () => ({ features: { edit: true } }),
            allow: () => true,
            deny: () => false,
            aclClasses: () => "",
            values: { years: [2024, 2021] },
            feature: () => true,
            get: () => false,
          },
        },
        stubs: {
          "p-loading": true,
          "p-scroll": true,
          "p-photo-clipboard": true,
          "p-photo-view-cards": true,
          "p-photo-view-mosaic": true,
          "v-alert": { template: "<div class='v-alert'><slot /></div>" },
        },
      },
    });

    await wrapper.vm.$nextTick();
    await flushPromises();

    expect(search).toHaveBeenCalled();
    const params = search.mock.calls[0][0];
    expect(params.month).toBeTruthy();
    expect(params.day).toBeTruthy();
    expect(params.order).toBe("newest");
    expect(wrapper.vm.loading).toBe(false);
    expect(wrapper.vm.emptyTitle).toBe("No pictures from this day in past years");
    wrapper.unmount();
  });

  it("omits the day filter on the month tab", async () => {
    const wrapper = shallowMount(PTabDiscoverPast, {
      props: { mode: "month", view: "cards" },
      global: {
        mocks: {
          $view: { isHidden: (target) => target !== "PPageDiscover" },
          $event: { subscribe: vi.fn(() => 1), unsubscribe: vi.fn() },
          $clipboard: { selection: [] },
          $gettext: (s) => s,
          $config: {
            getTimeZone: () => "UTC",
            getSettings: () => ({ features: { edit: true } }),
            allow: () => true,
            deny: () => false,
            aclClasses: () => "",
            values: { years: [2024, 2021] },
            feature: () => true,
            get: () => false,
          },
        },
        stubs: {
          "p-loading": true,
          "p-scroll": true,
          "p-photo-clipboard": true,
          "p-photo-view-cards": true,
          "p-photo-view-mosaic": true,
          "v-alert": { template: "<div class='v-alert'><slot /></div>" },
        },
      },
    });

    await wrapper.vm.$nextTick();
    await flushPromises();

    const params = search.mock.calls[0][0];
    expect(params.month).toBeTruthy();
    expect(params.day).toBeUndefined();
    wrapper.unmount();
  });

  function mountPast(isHidden) {
    return shallowMount(PTabDiscoverPast, {
      props: { mode: "day", view: "cards" },
      global: {
        mocks: {
          $view: { isHidden },
          $event: { subscribe: vi.fn(() => 1), unsubscribe: vi.fn() },
          $clipboard: { selection: [] },
          $gettext: (s) => s,
          $config: {
            getTimeZone: () => "UTC",
            getSettings: () => ({ features: { edit: true } }),
            allow: () => true,
            deny: () => false,
            aclClasses: () => "",
            values: { years: [2024, 2021] },
            feature: () => true,
            get: () => false,
          },
        },
        stubs: {
          "p-loading": true,
          "p-scroll": true,
          "p-photo-clipboard": true,
          "p-photo-view-cards": true,
          "p-photo-view-mosaic": true,
          "v-alert": true,
        },
      },
    });
  }

  it("loads the next page while Discover is the active view", async () => {
    search.mockImplementation(async (params) => {
      if (params.offset > 0) {
        return { models: [{ UID: "b", Year: 2019 }], count: 1, limit: 1, offset: params.offset };
      }

      return { models: [{ UID: "a", Year: 2020 }], count: 1, limit: 1, offset: 0 };
    });

    const isHidden = vi.fn((target) => target !== "PPageDiscover");
    const wrapper = mountPast(isHidden);

    await flushPromises();
    expect(search).toHaveBeenCalledTimes(1);
    expect(wrapper.vm.offset).toBe(1);

    await wrapper.vm.loadMore();
    await flushPromises();

    expect(isHidden).toHaveBeenCalledWith("PPageDiscover");
    expect(search.mock.calls[1][0].offset).toBe(1);
    expect(wrapper.vm.results.map((photo) => photo.UID)).toEqual(["a", "b"]);
    wrapper.unmount();
  });

  it("does not page while Discover is covered by another view", async () => {
    search.mockResolvedValue({
      models: [{ UID: "a", Year: 2020 }],
      count: 1,
      limit: 1,
      offset: 0,
    });

    const wrapper = mountPast(() => true);

    await flushPromises();
    await wrapper.vm.loadMore();
    await flushPromises();

    expect(search).toHaveBeenCalledTimes(1);
    wrapper.unmount();
  });
});

describe("page/discover/random.vue", () => {
  it("requests a random set and Shuffle repeats the query", async () => {
    const search = vi.spyOn(Photo, "search").mockResolvedValue({
      models: [{ UID: "p1" }],
      count: 1,
      limit: 36,
      offset: 0,
    });

    const wrapper = shallowMount(PTabDiscoverRandom, {
      props: { view: "mosaic" },
      global: {
        mocks: {
          $event: { subscribe: vi.fn(() => 1), unsubscribe: vi.fn() },
          $clipboard: { selection: [] },
          $gettext: (s) => s,
          $config: {
            getTimeZone: () => "UTC",
            getSettings: () => ({ features: { edit: true } }),
            allow: () => true,
            deny: () => false,
            aclClasses: () => "",
            values: {},
            feature: () => true,
            get: () => false,
          },
        },
        stubs: {
          "p-loading": true,
          "p-photo-clipboard": true,
          "p-photo-view-cards": true,
          "p-photo-view-mosaic": true,
        },
      },
    });

    await wrapper.vm.$nextTick();
    await flushPromises();

    expect(search.mock.calls[0][0].order).toBe("random");
    await wrapper.vm.shuffle();
    expect(search.mock.calls.length).toBeGreaterThanOrEqual(2);
    expect(search.mock.calls[1][0].order).toBe("random");
    wrapper.unmount();
    search.mockRestore();
  });
});
