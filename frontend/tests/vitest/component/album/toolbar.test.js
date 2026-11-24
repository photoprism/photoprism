import { describe, it, expect, vi, afterEach } from "vitest";
import { shallowMount, config } from "@vue/test-utils";
import PAlbumToolbar from "component/album/toolbar.vue";
import { $session } from "app/session";

let wrapper;
afterEach(() => {
  wrapper?.unmount();
  vi.restoreAllMocks();
});

// mountToolbar supplies the session and route state used by album controls.
function mountToolbar(role = "admin") {
  vi.spyOn($session, "getUser").mockReturnValue({ Role: role });
  const updateQuery = vi.fn();
  const replace = vi.fn();
  wrapper = shallowMount(PAlbumToolbar, {
    props: { album: { UID: "as6sg6bxpogaaba8", Title: "Album", Order: "newest", getId: () => "as6sg6bxpogaaba8" }, filter: { q: "cat" }, settings: { view: "cards" }, updateQuery },
    global: { renderStubDefaultSlot: true, stubs: { VMenu: { template: '<div><slot name="activator" :props="{}" /><slot /></div>' } }, mocks: {
      $config: { ...config.global.mocks.$config, albumCategories: () => [] },
      $route: { query: { q: "cat", order: "size" }, meta: {} }, $router: { replace },
    } },
  });
  return { updateQuery, replace };
}

describe("album toolbar controls", () => {
  it("keeps submitted search when reversing the current order", () => {
    const { replace } = mountToolbar();
    wrapper.vm.selectSort("size");
    expect(replace).toHaveBeenCalledWith({ query: { q: "cat", order: "size", reverse: "true" } });
  });
  it("uses the cover palette until the toolbar is compact or search is open", async () => {
    mountToolbar();
    await wrapper.setData({ searchOpen: false });
    await wrapper.setProps({ compact: false, collapseProgress: 0, coverStyle: { "--album-cover-ink": "#000", "--album-toolbar-shade": "0.2" } });
    expect(wrapper.vm.toolbarStyle["--album-toolbar-opacity"]).toBe(0);
    expect(wrapper.vm.toolbarStyle["--album-cover-ink"]).toBe("#000");
    expect(wrapper.vm.toolbarStyle["--album-toolbar-shade"]).toBe("0.2");
    await wrapper.setProps({ collapseProgress: 0.5 });
    expect(wrapper.vm.toolbarStyle["--album-toolbar-opacity"]).toBe(0.5);
    await wrapper.setData({ searchOpen: true });
    expect(wrapper.vm.toolbarStyle["--album-toolbar-opacity"]).toBe(1);
    expect(wrapper.vm.toolbarStyle["--album-toolbar-color"]).toBe("rgb(var(--v-theme-on-secondary))");
    expect(wrapper.findComponent({ name: "VTextField" }).props("bgColor")).toBe("surface");
  });
  it("retains the selected view when sorting after a search", () => {
    const { replace } = mountToolbar();
    wrapper.vm.setView("mosaic");
    expect(replace).toHaveBeenLastCalledWith({ query: { q: "cat", order: "size", view: "mosaic" } });
    wrapper.vm.$route.query.view = "mosaic";
    wrapper.vm.selectSort("title");
    expect(replace.mock.calls.at(-1)[0].query.view).toBe("mosaic");
  });
  it("clears search without changing order and submits a draft", () => {
    const { updateQuery } = mountToolbar();
    wrapper.vm.clearSearch();
    expect(updateQuery).toHaveBeenLastCalledWith({ q: "" });
    wrapper.vm.searchText = "zebra";
    wrapper.vm.submitSearch();
    expect(updateQuery).toHaveBeenLastCalledWith({ q: "zebra" });
  });
  it("restores the album toolbar when focus leaves an unsubmitted search", async () => {
    const { updateQuery } = mountToolbar();
    await wrapper.setProps({ filter: { q: "" }, compact: false, collapseProgress: 0 });
    await wrapper.setData({ searchOpen: true, searchText: "zebra" });
    await wrapper.find(".album-search-field").trigger("focusout");
    expect(wrapper.find(".album-search-field").exists()).toBe(false);
    expect(wrapper.find(".action-album-search").exists()).toBe(true);
    expect(wrapper.vm.searchText).toBe("");
    expect(updateQuery).not.toHaveBeenCalled();
  });
  it("keeps the submitted search visible when focus leaves the field", async () => {
    const { updateQuery } = mountToolbar();
    await wrapper.setData({ searchText: "unsubmitted" });
    await wrapper.find(".album-search-field").trigger("focusout");
    expect(wrapper.find(".album-search-field").exists()).toBe(true);
    expect(wrapper.vm.searchText).toBe("cat");
    expect(updateQuery).not.toHaveBeenCalled();
  });
  it("allows a guest to search and hides search for a share-link visitor", () => {
    mountToolbar("guest");
    expect(wrapper.vm.canSearch).toBe(true);
    wrapper.unmount();
    vi.restoreAllMocks();
    mountToolbar("visitor");
    expect(wrapper.vm.canSearch).toBe(false);
    expect(wrapper.find(".album-search-field").exists()).toBe(false);
    expect(wrapper.find(".action-album-sort").exists()).toBe(true);
  });
});
