import { describe, it, expect, vi } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { nextTick } from "vue";
import PTabImport from "page/library/import.vue";
import PTabIndex from "page/library/index.vue";

vi.mock("common/api", () => ({
  default: {
    delete: vi.fn(() => Promise.resolve({})),
    post: vi.fn(() => Promise.resolve({})),
    get: vi.fn(() => Promise.resolve({ data: {} })),
  },
}));

vi.mock("common/notify", () => ({
  default: {
    blockUI: vi.fn(),
    unblockUI: vi.fn(),
    error: vi.fn(),
  },
}));

function gettext(message, params = {}) {
  return String(message).replace(/%\{(\w+)\}/g, (_, key) => (params[key] !== undefined ? String(params[key]) : ""));
}

function buildConfigMock() {
  return {
    loading: vi.fn(() => false),
    load: vi.fn(() => Promise.resolve()),
    feature: vi.fn(() => true),
    insufficientStorage: vi.fn(() => false),
    get: vi.fn(() => false),
    getSettings: vi.fn(() => ({
      disable: { settings: true },
      import: { path: "/", move: false },
      index: { path: "/", rescan: false },
    })),
    values: {
      readonly: false,
      disable: { settings: true },
      count: { hidden: 0 },
    },
  };
}

function mountTab(component) {
  return mount(component, {
    global: {
      mocks: {
        $config: buildConfigMock(),
        $event: {
          subscribe: vi.fn(() => "progress-sub"),
          unsubscribe: vi.fn(),
          publish: vi.fn(),
        },
        $gettext: gettext,
        $isRtl: false,
        $session: { isAdmin: vi.fn(() => true) },
        $util: { truncate: (value) => value },
      },
    },
  });
}

describe("library import and index progress headers", () => {
  it("shows how many import progress events have been received", async () => {
    const wrapper = mountTab(PTabImport);
    await flushPromises();

    wrapper.vm.handleEvent("import.file", { baseName: "first.jpg" });
    await nextTick();

    expect(wrapper.text()).toContain("Importing first.jpg");
    expect(wrapper.text()).toContain("One file processed");

    wrapper.vm.handleEvent("import.file", { baseName: "second.jpg" });
    await nextTick();

    expect(wrapper.text()).toContain("Importing second.jpg");
    expect(wrapper.text()).toContain("2 files processed");

    wrapper.unmount();
  });

  it("shows how many index progress events have been received", async () => {
    const wrapper = mountTab(PTabIndex);
    await flushPromises();

    wrapper.vm.handleEvent("index.indexing", { fileName: "first.jpg" });
    await nextTick();

    expect(wrapper.text()).toContain("Indexing first.jpg");
    expect(wrapper.text()).toContain("One file processed");

    wrapper.vm.handleEvent("index.indexing", { fileName: "second.jpg" });
    await nextTick();

    expect(wrapper.text()).toContain("Indexing second.jpg");
    expect(wrapper.text()).toContain("2 files processed");

    wrapper.unmount();
  });
});
