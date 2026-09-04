// Pins the suggestion-list gating in page/people/new.vue: only the tile being
// edited receives the people list, so the item state a mounted grid holds stays
// proportional to the list length instead of list length times tile count.
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";

import PPageFaces from "page/people/new.vue";
import Face from "model/face";
import typeaheadCache from "common/typeahead-cache";

const people = [
  { UID: "js6sg6b2h8njw0sx", Name: "John Doe" },
  { UID: "js6sg6b1h1njaaaa", Name: "Jane Roe" },
];

// VComboboxStub exposes the items prop without rendering Vuetify's menu.
const VComboboxStub = {
  name: "VCombobox",
  props: ["items"],
  template: "<div class='combobox-stub'></div>",
};

describe("PPageFaces suggestion-list gating", () => {
  let wrapper;

  beforeEach(async () => {
    typeaheadCache.clear();
    vi.spyOn(typeaheadCache, "getPeople").mockResolvedValue(people);
    vi.spyOn(Face, "search").mockResolvedValue({
      models: [
        new Face({ ID: "FACE1", SubjUID: "", Name: "" }),
        new Face({ ID: "FACE2", SubjUID: "", Name: "" }),
        new Face({ ID: "FACE3", SubjUID: "", Name: "" }),
      ],
      count: 3,
      limit: 999,
      offset: 0,
    });

    wrapper = mount(PPageFaces, {
      props: { staticFilter: { markers: true, unknown: true }, active: true },
      global: {
        mocks: {
          $gettext: (msg) => msg,
          $gettextInterpolate: (msg) => msg,
          $notify: { info: vi.fn(), warn: vi.fn(), error: vi.fn(), blockUI: vi.fn(), unblockUI: vi.fn() },
          $config: { values: {}, get: vi.fn(() => false), feature: vi.fn(() => true) },
          $route: { query: {}, name: "people_faces" },
          $router: { replace: vi.fn(), push: vi.fn() },
          $event: { subscribe: vi.fn(() => 1), unsubscribe: vi.fn() },
        },
        stubs: {
          VCombobox: VComboboxStub,
          VTextField: true,
          VImg: true,
          PScroll: true,
          PLoading: true,
        },
      },
    });

    await flushPromises();
  });
  afterEach(() => {
    vi.restoreAllMocks();
    if (wrapper) {
      wrapper.unmount();
    }
  });

  it("hands every unfocused tile an empty list", () => {
    const boxes = wrapper.findAllComponents(VComboboxStub);
    expect(boxes).toHaveLength(3);
    boxes.forEach((box) => expect(box.props("items")).toHaveLength(0));
  });
  it("shares one array instance across the unfocused tiles", () => {
    const boxes = wrapper.findAllComponents(VComboboxStub);
    expect(boxes[0].props("items")).toBe(boxes[1].props("items"));
    expect(boxes[1].props("items")).toBe(boxes[2].props("items"));
  });
  it("hands the list to the focused tile only", async () => {
    await wrapper.vm.onFocusName({ ID: "FACE2" });
    await flushPromises();

    const boxes = wrapper.findAllComponents(VComboboxStub);
    expect(boxes[1].props("items")).toEqual(people);
    expect(boxes[0].props("items")).toHaveLength(0);
    expect(boxes[2].props("items")).toHaveLength(0);
  });
  it("moves the list when another tile is focused", async () => {
    await wrapper.vm.onFocusName({ ID: "FACE2" });
    await flushPromises();
    await wrapper.vm.onFocusName({ ID: "FACE3" });
    await flushPromises();

    const boxes = wrapper.findAllComponents(VComboboxStub);
    expect(boxes[2].props("items")).toEqual(people);
    expect(boxes[0].props("items")).toHaveLength(0);
    expect(boxes[1].props("items")).toHaveLength(0);
  });
});
