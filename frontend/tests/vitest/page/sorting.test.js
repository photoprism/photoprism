import { describe, it, expect, vi, beforeEach } from "vitest";
import Photos from "page/photos.vue";
import Albums from "page/albums.vue";
import Photo from "model/photo";
import Album from "model/album";
import { shallowMount, flushPromises } from "@vue/test-utils";
import { getAppStorage } from "common/storage";

beforeEach(() => getAppStorage().clear());

describe("page sort integration", () => {
  it.each([[Photos, "browse", "newest"], [Photos, "archive", "archived"], [Photos, "review", "added"], [Photos, "hidden", "added"]])("uses photo route defaults", (component, name, order) => {
    expect(component.methods.resolveSort.call({ $route: { name, query: {} } })).toEqual({ order, reverse: false });
  });
  it("shows album sorting without management access or filter-panel access", () => {
    const wrapper = shallowMount({ ...Albums, created() {}, mounted() {} }, {
      props: { staticFilter: { type: "album" }, defaultOrder: "favorites" },
      global: {
        mocks: {
          $route: { name: "albums", query: {} },
          $session: { isSuperAdmin: () => false },
          $view: { leave: vi.fn(), clearRestoreState: vi.fn() },
          $config: {
            allow: () => false, get: () => false, aclClasses: () => "", albumCategories: () => [],
            getSettings: () => ({ features: {} }), values: { count: { albums: 0 } },
          },
        },
        stubs: {
          VForm: { template: "<form><slot /></form>" },
          VToolbar: { template: "<div><slot /></div>" },
        },
      },
    });
    expect(wrapper.vm.canExpand).toBe(false);
    expect(wrapper.findComponent({ name: "PSortMenu" }).exists()).toBe(true);
    wrapper.unmount();
  });
  it("keeps embedded photos newest first without changing storage", () => {
    const keysBefore = getAppStorage().getItem("photos.order");
    expect(Photos.methods.resolveSort.call({ embedded: true, $route: { name: "browse", query: { order: "oldest" } } })).toEqual({ order: "newest", reverse: false });
    expect(getAppStorage().getItem("photos.order")).toBe(keysBefore);
  });
  it.each([["album", "favorites"], ["folder", "name"], ["state", "place"], ["moment", "newest"], ["month", "newest"]])("uses the album type default for %s", (type, defaultOrder) => {
    expect(Albums.methods.resolveSort.call({ staticFilter: { type }, defaultOrder, $route: { query: {} } })).toEqual({ order: defaultOrder, reverse: false });
  });
  it("persists album direction independently of other album types", () => {
    const folder = { staticFilter: { type: "folder" }, defaultOrder: "name", $route: { query: { order: "added", reverse: "true" } } };
    Albums.methods.resolveSort.call(folder);
    const moment = { staticFilter: { type: "moment" }, defaultOrder: "newest", $route: { query: {} } };
    expect(Albums.methods.resolveSort.call(moment)).toEqual({ order: "newest", reverse: false });
    folder.$route.query = {};
    expect(Albums.methods.resolveSort.call(folder)).toEqual({ order: "added", reverse: true });
  });
  it.each([Photos, Albums])("keeps filters intact and drops false reverse from API requests", (component) => {
    const state = { filter: { order: "size", reverse: false, year: 2015, type: "raw", q: "cat" }, staticFilter: { favorite: "true" }, offset: 0, searchCount: () => 100 };
    const params = component.methods.searchParams.call(state);
    expect(params).toMatchObject({ order: "size", year: 2015, type: "raw", q: "cat", favorite: "true" });
    expect(params).not.toHaveProperty("reverse");
    expect(state.filter.reverse).toBe(false);
  });
  it.each([Photos, Albums])("does not partially mutate a loading page", (component) => {
    const updateFilter = vi.fn();
    expect(component.methods.updateQuery.call({ loading: true, updateFilter }, { order: "name" })).toBe(false);
    expect(updateFilter).not.toHaveBeenCalled();
  });
  it.each([Photos, Albums])("updates the route once, preserving unrelated filters", (component) => {
    const state = {
      loading: false,
      filter: { order: "size", reverse: false, year: 2015, type: "raw", q: "cat" },
      settings: { view: "cards" },
      $route: { query: {} },
      $router: { replace: vi.fn() },
    };
    state.updateFilter = component.methods.updateFilter.bind(state);
    expect(component.methods.updateQuery.call(state, { order: "name", reverse: true })).toBe(true);
    expect(state.$router.replace).toHaveBeenCalledExactlyOnceWith({ query: { view: "cards", order: "name", reverse: true, year: 2015, type: "raw", q: "cat" } });
  });
  it.each([[Photos, Photo], [Albums, Album]])("normalizes pagination and refresh requests as well as initial searches", async (component, model) => {
    for (const [order, reverse, expected] of [["newest", false, undefined], ["oldest", false, undefined], ["name", true, true], ["name", false, undefined]]) {
      const search = vi.spyOn(model, "search").mockResolvedValue({ models: [], count: 0, limit: 1, offset: 0 });
      const state = {
        scrollDisabled: false, $view: { isHidden: () => false },
        lastFilter: { order, reverse, type: "raw" }, staticFilter: { favorite: "true" },
        dirty: false, page: 0, batchSize: 10, offset: 0, results: [], lightbox: { open: false },
        restorePending: 0, setOffset: vi.fn(), $nextTick: vi.fn(), finishRestore: vi.fn(),
      };
      component.methods.loadMore.call(state);
      const params = search.mock.calls[0][0];
      expect(params).toMatchObject({ order, type: "raw", favorite: "true" });
      if (expected) {
        expect(params.reverse).toBe(true);
      } else {
        expect(params).not.toHaveProperty("reverse");
      }
      await flushPromises();
      search.mockRestore();
    }
  });
  it.each([Photos, Albums])("does not change filters while a pagination request is pending", (component) => {
    const updateFilter = vi.fn();
    expect(component.methods.updateQuery.call({ loading: false, listen: false, updateFilter }, { order: "name" })).toBe(false);
    expect(updateFilter).not.toHaveBeenCalled();
  });
  it("restores media types only on unrestricted photo pages", () => {
    expect(Photos.methods.mediaType.call({ staticFilter: {} }, { type: "raw" })).toBe("raw");
    expect(Photos.methods.mediaType.call({ staticFilter: {} }, {})).toBe("");
    for (const key of ["video", "live", "audio", "animated"]) {
      expect(Photos.methods.mediaType.call({ staticFilter: { [key]: "true" } }, { type: "raw" })).toBe("");
    }
  });
});
