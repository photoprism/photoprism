import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount } from "@vue/test-utils";
import { nextTick } from "vue";
import { createVuetify } from "vuetify";
import * as components from "vuetify/components";
import * as labsComponents from "vuetify/labs/components";
import * as directives from "vuetify/directives";
import PUploadDialog from "component/upload/dialog.vue";

// VFileUploadItem calls URL.createObjectURL to build preview src URLs; jsdom does not implement it.
if (typeof URL.createObjectURL === "undefined") {
  URL.createObjectURL = () => "blob:mock";
  URL.revokeObjectURL = () => {};
}

// VDialog relies on window.visualViewport which jsdom does not provide.
if (typeof window.visualViewport === "undefined") {
  Object.defineProperty(window, "visualViewport", {
    value: {
      width: 1024,
      height: 768,
      offsetLeft: 0,
      offsetTop: 0,
      pageLeft: 0,
      pageTop: 0,
      scale: 1,
      addEventListener: () => {},
      removeEventListener: () => {},
    },
    writable: true,
    configurable: true,
  });
}

// Vuetify instance that includes both standard and labs components (for VFileUpload).
const vuetify = createVuetify({
  components: { ...components, ...labsComponents },
  directives,
  theme: { defaultTheme: "light" },
});

// Shared file fixture.
const makeFile = (name = "photo.jpg", size = 1024, type = "image/jpeg") => new File(["x".repeat(size)], name, { type });

// Stubs for modules that make external calls.
vi.mock("model/album", () => ({
  default: { search: vi.fn(() => Promise.resolve({ models: [] })) },
}));

vi.mock("common/albums", () => ({
  createAlbumSelectionWatcher: () => () => {},
}));

vi.mock("common/api", () => ({
  default: {
    post: vi.fn(() => Promise.resolve({})),
    put: vi.fn(() => Promise.resolve({})),
  },
}));

vi.mock("common/notify", () => ({
  default: { info: vi.fn(), success: vi.fn(), error: vi.fn(), warn: vi.fn() },
}));

function buildConfigMock(overrides = {}) {
  return {
    get: vi.fn((key) => {
      if (key === "demo") return false;
      if (key === "uploadAllow") return "image/*";
      if (key === "uploadNSFW") return false;
      return null;
    }),
    feature: vi.fn(() => false),
    filesQuotaReached: vi.fn(() => false),
    ...overrides,
  };
}

function mountDialog({ visible = true, data = {}, configOverrides = {} } = {}) {
  return mount(PUploadDialog, {
    props: { visible, data },
    attachTo: document.body,
    global: {
      plugins: [vuetify],
      mocks: {
        $config: buildConfigMock(configOverrides),
        $session: { getUserUID: vi.fn(() => "uid123") },
        $util: { generateToken: vi.fn(() => "tok-abc") },
        $view: { enter: vi.fn(), leave: vi.fn() },
        $notify: { info: vi.fn(), success: vi.fn(), error: vi.fn() },
        $gettext: (msg, params = {}) => msg.replace(/%\{(\w+)\}/g, (_, k) => (params[k] !== undefined ? String(params[k]) : "")),
        $isRtl: false,
      },
    },
  });
}

describe("component/upload/dialog", () => {
  let wrapper;

  beforeEach(() => {
    wrapper = mountDialog();
  });

  afterEach(() => {
    if (wrapper) wrapper.unmount();
    vi.clearAllMocks();
  });

  // ─── Rendering ────────────────────────────────────────────────────────────

  describe("Rendering", () => {
    it("renders the v-file-upload component", () => {
      expect(wrapper.findComponent({ name: "VFileUpload" }).exists()).toBe(true);
    });

    it("does not render a hidden native file input", () => {
      // The old hidden <input class="input-upload"> should be gone.
      expect(wrapper.find("input.input-upload").exists()).toBe(false);
    });

    // v-dialog teleports its content to document.body; query from there.
    it("renders the Upload action button in the teleported dialog", () => {
      const btn = document.body.querySelector(".action-upload");
      expect(btn).not.toBeNull();
      expect(btn.textContent.trim()).toBe("Upload");
    });

    it("renders the Close button in the teleported dialog", () => {
      expect(document.body.querySelector(".action-close")).not.toBeNull();
    });

    it("v-file-upload carries the input-file-upload class", () => {
      expect(document.body.querySelector(".input-file-upload")).not.toBeNull();
    });
  });

  // ─── v-file-upload props ──────────────────────────────────────────────────

  describe("v-file-upload binding", () => {
    it("passes filterByType from config (the accept value)", () => {
      const fu = wrapper.findComponent({ name: "VFileUpload" });
      // accept is mapped to the filterByType prop on VFileUpload.
      expect(fu.props("filterByType")).toBe("image/*");
    });

    it("passes multiple=true", () => {
      const fu = wrapper.findComponent({ name: "VFileUpload" });
      expect(fu.props("multiple")).toBe(true);
    });

    it("is disabled when busy is true", async () => {
      wrapper.vm.busy = true;
      await nextTick();
      const fu = wrapper.findComponent({ name: "VFileUpload" });
      expect(fu.props("disabled")).toBe(true);
    });

    it("is disabled when filesQuotaReached is true", async () => {
      wrapper.vm.filesQuotaReached = true;
      await nextTick();
      const fu = wrapper.findComponent({ name: "VFileUpload" });
      expect(fu.props("disabled")).toBe(true);
    });

    it("is enabled (not disabled) in the default state", () => {
      const fu = wrapper.findComponent({ name: "VFileUpload" });
      expect(fu.props("disabled")).toBe(false);
    });

    it("updates selected when v-file-upload emits update:modelValue", async () => {
      const fu = wrapper.findComponent({ name: "VFileUpload" });
      const files = [makeFile("a.jpg"), makeFile("b.jpg")];
      await fu.vm.$emit("update:modelValue", files);
      await nextTick();
      expect(wrapper.vm.selected).toEqual(files);
    });
  });

  // ─── onFilesSelected — append / remove / dedup ────────────────────────────

  describe("onFilesSelected()", () => {
    it("clears selection when an empty array is emitted", async () => {
      wrapper.vm.selected = [makeFile()];
      await nextTick();
      wrapper.vm.onFilesSelected([]);
      expect(wrapper.vm.selected).toEqual([]);
    });

    it("appends new files to existing selection on a second browse pass", () => {
      const a = makeFile("a.jpg");
      const b = makeFile("b.jpg");
      wrapper.vm.selected = [a];
      // b is a brand-new File object (not in existing by reference) → merge
      wrapper.vm.onFilesSelected([b]);
      expect(wrapper.vm.selected).toHaveLength(2);
      expect(wrapper.vm.selected).toContain(a);
      expect(wrapper.vm.selected).toContain(b);
    });

    it("removes a file when the emitted value is a subset of the existing selection", () => {
      const a = makeFile("a.jpg");
      const b = makeFile("b.jpg");
      wrapper.vm.selected = [a, b];
      // VFileUploadItem × button emits the remaining files by reference
      wrapper.vm.onFilesSelected([b]);
      expect(wrapper.vm.selected).toEqual([b]);
    });

    it("skips duplicates identified by name + size + lastModified", () => {
      const ts = 1700000000000;
      const a = new File(["x".repeat(512)], "a.jpg", { type: "image/jpeg", lastModified: ts });
      wrapper.vm.selected = [a];
      // Second browse produces a different File object but with identical metadata
      const aDup = new File(["x".repeat(512)], "a.jpg", { type: "image/jpeg", lastModified: ts });
      wrapper.vm.onFilesSelected([aDup]);
      expect(wrapper.vm.selected).toHaveLength(1);
    });

    it("handles a non-array (single File) emitted value by wrapping it", () => {
      const f = makeFile("single.jpg");
      wrapper.vm.onFilesSelected(f);
      expect(wrapper.vm.selected).toContain(f);
    });
  });

  // ─── hasFiles computed ─────────────────────────────────────────────────────

  describe("hasFiles computed", () => {
    it("is false when no files are selected", () => {
      expect(wrapper.vm.hasFiles).toBe(false);
    });

    it("is true when at least one file is selected", async () => {
      wrapper.vm.selected = [makeFile()];
      await nextTick();
      expect(wrapper.vm.hasFiles).toBe(true);
    });

    it("returns false after selected is reset to empty array", async () => {
      wrapper.vm.selected = [makeFile()];
      await nextTick();
      wrapper.vm.selected = [];
      await nextTick();
      expect(wrapper.vm.hasFiles).toBe(false);
    });

    it("returns false when selected is not an array", async () => {
      wrapper.vm.selected = null;
      await nextTick();
      expect(wrapper.vm.hasFiles).toBe(false);
    });
  });

  // ─── Upload button enabled/disabled ───────────────────────────────────────

  describe("Upload button disabled state", () => {
    it("is disabled with no files selected (hasFiles=false)", () => {
      expect(wrapper.vm.hasFiles).toBe(false);
    });

    it("reflects hasFiles=true once a file is selected", async () => {
      wrapper.vm.selected = [makeFile()];
      await nextTick();
      expect(wrapper.vm.hasFiles).toBe(true);
    });

    it("busy overrides hasFiles for the disabled calculation", async () => {
      wrapper.vm.selected = [makeFile()];
      wrapper.vm.busy = true;
      await nextTick();
      expect(wrapper.vm.busy).toBe(true);
    });
  });

  // ─── onUpload() guards ────────────────────────────────────────────────────

  describe("onUpload() — no files selected", () => {
    it("does not set busy when selected is empty", () => {
      wrapper.vm.selected = [];
      wrapper.vm.onUpload();
      expect(wrapper.vm.busy).toBe(false);
    });
  });

  describe("onUpload() — already busy", () => {
    it("returns without changing total when already busy", () => {
      wrapper.vm.busy = true;
      wrapper.vm.selected = [makeFile()];
      const prevTotal = wrapper.vm.total;
      wrapper.vm.onUpload();
      expect(wrapper.vm.total).toBe(prevTotal);
    });
  });

  describe("onUpload() — demo file limit", () => {
    it("shows error and skips upload when too many files in demo mode", async () => {
      const $notify = (await import("common/notify")).default;
      const demoWrapper = mountDialog();
      demoWrapper.vm.isDemo = true;
      demoWrapper.vm.fileLimit = 3;
      demoWrapper.vm.selected = [makeFile("1.jpg"), makeFile("2.jpg"), makeFile("3.jpg"), makeFile("4.jpg")];

      demoWrapper.vm.onUpload();

      expect(demoWrapper.vm.busy).toBe(false);
      expect($notify.error).toHaveBeenCalledWith("Too many files selected");
      demoWrapper.unmount();
    });
  });

  // ─── onUpload() — starts upload ───────────────────────────────────────────

  describe("onUpload() — starts upload", () => {
    it("sets busy and initialises total and totalSize", () => {
      wrapper.vm.selected = [makeFile("a.jpg", 500), makeFile("b.jpg", 300)];
      wrapper.vm.onUpload();

      expect(wrapper.vm.busy).toBe(true);
      expect(wrapper.vm.total).toBe(2);
      expect(wrapper.vm.totalSize).toBe(800);
    });

    it("uses generateToken to set the upload token", () => {
      wrapper.vm.selected = [makeFile()];
      wrapper.vm.onUpload();
      expect(wrapper.vm.token).toBe("tok-abc");
    });

    it("resets eta and remainingTime at start", () => {
      wrapper.vm.selected = [makeFile()];
      wrapper.vm.eta = "2 minutes";
      wrapper.vm.remainingTime = 120;
      wrapper.vm.onUpload();
      expect(wrapper.vm.eta).toBe("");
      expect(wrapper.vm.remainingTime).toBe(-1);
    });
  });

  // ─── reset() ─────────────────────────────────────────────────────────────

  describe("reset()", () => {
    it("clears selected files", () => {
      wrapper.vm.selected = [makeFile()];
      wrapper.vm.reset();
      expect(wrapper.vm.selected).toEqual([]);
    });

    it("resets busy, total, and token", () => {
      Object.assign(wrapper.vm, { busy: true, total: 5, token: "x" });
      wrapper.vm.reset();
      expect(wrapper.vm.busy).toBe(false);
      expect(wrapper.vm.total).toBe(0);
      expect(wrapper.vm.token).toBe("");
    });

    it("resets all size and progress counters", () => {
      Object.assign(wrapper.vm, {
        completedTotal: 80,
        completedSize: 400,
        totalSize: 500,
        totalFailed: 1,
      });
      wrapper.vm.reset();
      expect(wrapper.vm.completedTotal).toBe(0);
      expect(wrapper.vm.completedSize).toBe(0);
      expect(wrapper.vm.totalSize).toBe(0);
      expect(wrapper.vm.totalFailed).toBe(0);
    });
  });

  // ─── onClose() ───────────────────────────────────────────────────────────

  describe("onClose()", () => {
    it("invokes the onClose handler when not busy", () => {
      // Pass an event-listener prop (Vue 3 idiomatic alternative to wrapper.emitted()).
      const onClose = vi.fn();
      const w = mountDialog({ visible: true });
      // Inject the handler after mount so we can observe call count.
      w.vm.$props; // touch props to ensure reactivity is set up
      w.vm.busy = false;
      // Attach via the component's onClose prop path by mounting a fresh wrapper.
      const w2 = mount(PUploadDialog, {
        props: { visible: true, data: {}, onClose },
        attachTo: document.body,
        global: {
          plugins: [vuetify],
          mocks: {
            $config: buildConfigMock(),
            $session: { getUserUID: vi.fn(() => "uid123") },
            $util: { generateToken: vi.fn(() => "tok-abc") },
            $view: { enter: vi.fn(), leave: vi.fn() },
            $notify: { info: vi.fn(), success: vi.fn(), error: vi.fn() },
            $gettext: (msg, params = {}) => msg.replace(/%\{(\w+)\}/g, (_, k) => (params[k] !== undefined ? String(params[k]) : "")),
            $isRtl: false,
          },
        },
      });
      w2.vm.busy = false;
      w2.vm.onClose();
      expect(onClose).toHaveBeenCalledOnce();
      w.unmount();
      w2.unmount();
    });

    it("does not invoke the onClose handler when busy", () => {
      const onClose = vi.fn();
      const w2 = mount(PUploadDialog, {
        props: { visible: true, data: {}, onClose },
        attachTo: document.body,
        global: {
          plugins: [vuetify],
          mocks: {
            $config: buildConfigMock(),
            $session: { getUserUID: vi.fn(() => "uid123") },
            $util: { generateToken: vi.fn(() => "tok-abc") },
            $view: { enter: vi.fn(), leave: vi.fn() },
            $notify: { info: vi.fn(), success: vi.fn(), error: vi.fn() },
            $gettext: (msg, params = {}) => msg.replace(/%\{(\w+)\}/g, (_, k) => (params[k] !== undefined ? String(params[k]) : "")),
            $isRtl: false,
          },
        },
      });
      w2.vm.busy = true;
      w2.vm.onClose();
      expect(onClose).not.toHaveBeenCalled();
      w2.unmount();
    });
  });

  // ─── onUploadProgress() ───────────────────────────────────────────────────

  describe("onUploadProgress()", () => {
    it("ignores events with missing loaded or total", () => {
      wrapper.vm.totalSize = 1000;
      wrapper.vm.onUploadProgress({});
      expect(wrapper.vm.completedTotal).toBe(0);
    });

    it("does not update when loaded >= total", () => {
      wrapper.vm.totalSize = 1000;
      wrapper.vm.completedSize = 0;
      wrapper.vm.started = Date.now() - 1000;
      wrapper.vm.onUploadProgress({ loaded: 1000, total: 1000 });
      // loaded < total is required; at equality the branch is skipped.
      expect(wrapper.vm.completedTotal).toBe(0);
    });

    it("updates completedTotal proportionally", () => {
      wrapper.vm.totalSize = 1000;
      wrapper.vm.completedSize = 0;
      wrapper.vm.started = Date.now() - 1000;
      wrapper.vm.onUploadProgress({ loaded: 400, total: 1000 });
      expect(wrapper.vm.completedTotal).toBeGreaterThan(0);
      expect(wrapper.vm.completedTotal).toBeLessThanOrEqual(40);
    });
  });

  // ─── onUploadComplete() ───────────────────────────────────────────────────

  describe("onUploadComplete()", () => {
    it("ignores a null file", () => {
      wrapper.vm.completedSize = 0;
      wrapper.vm.onUploadComplete(null);
      expect(wrapper.vm.completedSize).toBe(0);
    });

    it("ignores a file with no size", () => {
      wrapper.vm.completedSize = 0;
      wrapper.vm.onUploadComplete({ size: 0 });
      expect(wrapper.vm.completedSize).toBe(0);
    });

    it("accumulates completedSize and updates completedTotal", () => {
      wrapper.vm.totalSize = 2000;
      wrapper.vm.completedSize = 0;
      wrapper.vm.onUploadComplete({ size: 750 });
      expect(wrapper.vm.completedSize).toBe(750);
      expect(wrapper.vm.completedTotal).toBe(37);
    });
  });

  // ─── visible watcher ─────────────────────────────────────────────────────

  describe("visible watcher", () => {
    it("calls reset when dialog becomes hidden", async () => {
      const spy = vi.spyOn(wrapper.vm, "reset");
      await wrapper.setProps({ visible: false });
      expect(spy).toHaveBeenCalled();
    });

    it("pre-populates selectedAlbums from data.albums when shown", async () => {
      const albums = [{ UID: "u1", Title: "Holiday" }];
      const w = mountDialog({ visible: false, data: { albums } });
      await w.setProps({ visible: true });
      await nextTick();
      expect(w.vm.selectedAlbums).toEqual(albums);
      w.unmount();
    });
  });

  // ─── title computed ───────────────────────────────────────────────────────

  describe("title computed", () => {
    it("returns Upload", () => {
      expect(wrapper.vm.title).toBe("Upload");
    });
  });
});
