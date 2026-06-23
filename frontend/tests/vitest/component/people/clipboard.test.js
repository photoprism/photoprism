import { describe, it, expect, vi } from "vitest";
import { shallowMount } from "@vue/test-utils";
import PPeopleClipboard from "component/people/clipboard.vue";

function mountClipboard({ selection = ["subj1", "subj2"], allow = vi.fn(() => true), feature = vi.fn(() => true) } = {}) {
  const router = {
    push: vi.fn(),
  };

  const clearSelection = vi.fn();

  const wrapper = shallowMount(PPeopleClipboard, {
    props: {
      selection,
      clearSelection,
    },
    global: {
      mocks: {
        $config: {
          allow,
          feature,
          getSettings: () => ({ features: { albums: true, download: true, search: true } }),
        },
        $gettext: (msg) => msg,
        $isRtl: false,
        $router: router,
      },
      stubs: {
        "v-speed-dial": { template: "<div><slot></slot></div>" },
        "v-btn-toggle": { template: '<div v-bind="$attrs"><slot></slot></div>' },
        "v-btn": { template: '<button v-bind="$attrs"><slot></slot></button>' },
        "p-photo-album-dialog": true,
      },
    },
  });

  return { wrapper, router, clearSelection };
}

describe("component/people/clipboard", () => {
  it("shows the search mode toggle only for multi-person selections", () => {
    const { wrapper } = mountClipboard();
    const { wrapper: singleWrapper } = mountClipboard({ selection: ["subj1"] });

    expect(wrapper.find(".action-search-mode").exists()).toBe(true);
    expect(singleWrapper.find(".action-search-mode").exists()).toBe(false);
  });

  it("searches selected people with a subject OR query", () => {
    const { wrapper, router, clearSelection } = mountClipboard();

    wrapper.vm.search();

    expect(wrapper.vm.searchMode).toBe("or");
    expect(router.push).toHaveBeenCalledWith({ name: "all", query: { q: "subject:subj1|subj2" } });
    expect(clearSelection).toHaveBeenCalledTimes(1);
    expect(wrapper.vm.expanded).toBe(false);
  });

  it("searches selected people with a subject AND query", () => {
    const { wrapper, router, clearSelection } = mountClipboard();

    wrapper.vm.searchMode = "and";
    wrapper.vm.search();

    expect(router.push).toHaveBeenCalledWith({ name: "all", query: { q: "subject:subj1&subj2" } });
    expect(clearSelection).toHaveBeenCalledTimes(1);
  });

  it("does not search when no people are selected", () => {
    const { wrapper, router, clearSelection } = mountClipboard({ selection: [] });

    wrapper.vm.search();

    expect(router.push).not.toHaveBeenCalled();
    expect(clearSelection).not.toHaveBeenCalled();
  });

  it("does not search when photo search permission is denied", () => {
    const allow = vi.fn((resource, action) => !(resource === "photos" && action === "search"));
    const { wrapper, router, clearSelection } = mountClipboard({ allow });

    wrapper.vm.search();

    expect(router.push).not.toHaveBeenCalled();
    expect(clearSelection).not.toHaveBeenCalled();
  });
});
