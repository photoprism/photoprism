import { mount, flushPromises } from "@vue/test-utils";
import { reactive } from "vue";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import Dialog from "component/photo/export/dialog.vue";
import Photo from "model/photo";
import File from "model/file";
import Config from "common/config";
import Util from "common/util";
import Axios from "axios";
import { customDark, customLight } from "../../../helpers/themes";

const thumbs = [
  { size: "fit_720", usage: "SD TV, Mobile", w: 720, h: 720 },
  { size: "fit_1920", usage: "Full HD", w: 1920, h: 1200 },
];

// file creates a fully loaded image file for dialog tests.
function file(id, values = {}) {
  return { UID: `fs6sg6bqhhin${String(id.charCodeAt(0)).padStart(4, "0")}`, Hash: id.charCodeAt(0).toString(16).padStart(40, "0"), Name: `${id}.jpg`, Root: "/", MediaType: "image", FileType: "jpg", Width: 4000, Height: 3000, Size: 10000, ...values };
}

const initialWidth = window.innerWidth;
let wrapper;
let config;
let session;
let view;

// open mounts the export dialog with live ACL and setting state and full-detail loading.
async function open(files = [file("a"), file("b")], options = {}) {
  config = {
    themeName: "light", apiUri: "/api/v1", contentUri: "/api/v1", downloadToken: "old", previewToken: "preview",
    values: reactive({ settings: { features: { export: true, download: false } }, acl: { files: { access_all: true } }, thumbs }),
    allow: Config.prototype.allow,
  };
  session = { refresh: vi.fn(async () => {
    config.downloadToken = "fresh";
  }) };
  view = { enter: vi.fn(), leave: vi.fn(), focus: vi.fn() };
  vi.spyOn(Photo.prototype, "find").mockResolvedValue({ Files: files });
  Photo.prototype.find.mockClear();
  wrapper = mount(Dialog, {
    props: { visible: true, uid: "ps6sg6be2lvl0yh7", ...options },
    attachTo: document.body,
    global: {
      stubs: { "v-img": { props: ["src", "width", "height"], template: '<img :src="src" :width="width" :height="height" alt="" />' }, "v-dialog": { name: "ExportDialogStub", props: ["modelValue", "fullscreen", "theme", "maxWidth"], template: '<div v-if="modelValue" class="dialog-stub"><slot /></div>' } },
      mocks: { $config: config, $session: session, $view: view, $util: Util },
    },
  });
  await flushPromises();
  session.refresh.mockClear();
  return wrapper;
}

beforeEach(() => {
  vi.spyOn(Axios, "get").mockResolvedValue({ data: {} });
  vi.spyOn(HTMLAnchorElement.prototype, "click").mockImplementation(() => {});
});
afterEach(() => {
  if (wrapper) {
    wrapper.vm.$vuetify.theme.global.name = "light";
    window.innerWidth = initialWidth;
    wrapper.vm.$vuetify.display.update();
    wrapper.unmount();
  }
  wrapper = null; vi.restoreAllMocks(); vi.useRealTimers(); document.body.innerHTML = "";
});

describe("photo export dialog", () => {
  it.each([720, 1920, 7680])("uses the server's thumbnail limit %s when loading export choices", async (limit) => {
    await open([file("p", { Name: "large.png", FileType: "png" })]);
    config.values.thumbLimit = limit;
    config.values.thumbs = [...thumbs, { size: "fit_4096", usage: "DCI 4K", w: 4096, h: 4096 }];
    await wrapper.vm.loadFiles();
    await flushPromises();
    const options = wrapper.vm.groups[0].options;
    expect(options.at(-1)).toMatchObject(limit === 7680
      ? { label: "Full Size", size: "fit_4096", width: 4000, height: 3000 }
      : limit === 1920 ? { size: "fit_1920", width: 1600, height: 1200 } : { size: "fit_720", width: 720, height: 540 });
    expect(wrapper.text().includes("Full Size")).toBe(limit === 7680);
  });
  it.each([customDark, customLight])("uses the lightbox theme with custom application theme $name", async (theme) => {
    await open();
    config.themeName = theme.name;
    wrapper.vm.$vuetify.theme.themes[theme.name] = theme;
    wrapper.vm.$vuetify.theme.global.name = theme.name;
    await flushPromises();
    expect(wrapper.findComponent({ name: "ExportDialogStub" }).props("theme")).toBe("lightbox");
  });
  it("keeps mobile fullscreen with a top close control and full-width primary action", async () => {
    await open();
    window.innerWidth = 375;
    wrapper.vm.$vuetify.display.update();
    await flushPromises();
    expect(wrapper.findComponent({ name: "ExportDialogStub" }).props("fullscreen")).toBe(true);
    expect(wrapper.find(".action-cancel").exists()).toBe(false);
    expect(wrapper.find(".export-actions .action-export").classes()).toContain("v-btn--block");
    await wrapper.find(".action-dismiss").trigger("click");
    expect(wrapper.emitted("close")).toHaveLength(1);
  });
  it("binds the lightbox theme and desktop width", async () => {
    await open();
    expect(wrapper.findComponent({ name: "ExportDialogStub" }).props()).toMatchObject({ theme: "lightbox", maxWidth: "560" });
    expect(wrapper.find(".action-export").text()).toBe("Export 0 Files");
    expect(wrapper.find(".export-hint").exists()).toBe(false);
    await wrapper.find(".action-dismiss").trigger("click");
    expect(wrapper.emitted("close")).toHaveLength(1);
  });
  it("uses the matched raster model for a long Unicode original and its own metadata", async () => {
    const original = file("r", { Name: "folder/旅行." + "long".repeat(30) + ".dng", MediaType: "raw", FileType: "dng", Codec: "raw", Duration: 1500000000 });
    const preview = file("p", { Name: original.Name + ".jpg", Root: "sidecar", Width: 2000, Height: 1000, Size: 5000 });
    const thumbnail = vi.spyOn(File.prototype, "thumbnailUrl").mockImplementation(function (size) {
      return `/matched/${this.Hash}/${size}`;
    });
    await open([original, preview]);
    expect(wrapper.find(".export-thumbnail").attributes("src")).toBe(`/matched/${preview.Hash}/tile_224`);
    expect(thumbnail).toHaveBeenCalledWith("tile_224");
    expect(wrapper.find(".export-filename").text()).toContain("旅行.");
    expect(wrapper.find(".export-file-type").text()).toBe("DNG · RAW");
    expect(wrapper.find(".export-file-metadata").text()).toBe(`4000 × 3000 · raw · ${Util.formatDuration(1500000000)} · 10 KB`);
    expect(wrapper.vm.optionBytes(wrapper.vm.groups[0].options[1])).toBe("5 KB");
    expect(wrapper.vm.optionBytes(wrapper.vm.groups[0].options[2])).toBe("");
  });
  it("omits unknown metadata and supplies a media icon without a matched preview", async () => {
    await open([file("p", { Name: "document.pdf", MediaType: "document", FileType: "pdf", Width: 0, Height: 0, Size: undefined })]);
    expect(wrapper.find(".export-file-metadata").exists()).toBe(false);
    expect(wrapper.find(".export-option-size").exists()).toBe(false);
    expect(wrapper.find(".export-resized-jpegs").exists()).toBe(false);
    expect(wrapper.vm.optionMetadata(wrapper.vm.groups[0].options[0])).toBe("PDF");
    expect(wrapper.vm.mediaIcon(wrapper.vm.groups[0])).toBe("mdi-file-document");
  });
  it("orders compatible video before the preview and JPEG sizes without changing queue order", async () => {
    const animation = file("a", { Name: "animation.gif", FileType: "gif", Frames: 2 });
    const preview = file("p", { Name: "animation.gif.jpg", Root: "sidecar" });
    Axios.get.mockResolvedValue({ data: { codec: "gif", duration: 2, avc: { ready: true, format: "mp4", codec: "avc1", width: 4000, height: 3000, bytes: 50000 } } });
    await open([animation, preview]);
    expect(wrapper.findAll(".export-option").map((row) => row.attributes("data-option-id").split(":")[0])).toEqual(["original", "avc", "preview", "thumbnail", "thumbnail"]);
    expect(wrapper.vm.groups[0].options.map((option) => option.kind)).toEqual(["original", "preview", "thumbnail", "thumbnail", "avc"]);
    const compatible = wrapper.find('.export-option[data-option-id^="avc"]');
    expect(compatible.text()).toContain("Compatible Video");
    expect(compatible.text()).toContain("4000 × 3000 · MP4 · avc1");
    expect(compatible.find(".export-option-size").text()).toBe("49 KB");
    expect(compatible.find(".export-conversion-note").exists()).toBe(false);
    expect(wrapper.vm.groups[0].original).toMatchObject({ duration: 2000000000, codec: "gif" });
  });
  it("groups resized JPEGs under one heading per original while keeping originals and previews separate", async () => {
    const raw = file("r", { Name: "photo.dng", FileType: "dng", MediaType: "raw" });
    const preview = file("p", { Name: "photo.dng.jpg", Root: "sidecar" });
    await open([raw, preview, file("a")]);
    await wrapper.setData({ expanded: wrapper.vm.groups.map((group) => group.id) });
    const sections = wrapper.findAll(".export-resized-jpegs");
    expect(sections).toHaveLength(2);
    for (const section of sections) {
      expect(section.attributes("role")).toBe("group");
      expect(section.attributes("aria-label")).toBe("Resized JPEGs");
      expect(section.find(".export-section-title").text()).toBe("Resized JPEGs");
      expect(section.findAll(".export-option").map((row) => row.attributes("data-option-id").split(":")[0])).toEqual(["thumbnail", "thumbnail"]);
      expect(section.findAll(".v-divider")).toHaveLength(1);
    }
    expect(wrapper.findAll(".export-options-section:not(.export-resized-jpegs) .export-option").map((row) => row.attributes("data-option-id").split(":")[0])).toEqual(["original", "preview", "original"]);
    await sections[0].find(`input[value="thumbnail:${preview.UID}:fit_1920"]`).setValue(true);
    await sections[0].find(`input[value="thumbnail:${preview.UID}:fit_720"]`).setValue(true);
    expect(wrapper.vm.selected).toContain(`thumbnail:${preview.UID}:fit_720`);
    expect(wrapper.find(".action-export").text()).toBe("Export 2 Files");
    expect(wrapper.find(".export-hint").exists()).toBe(true);
  });
  it("loads full Files with Download disabled and leaves every option unselected", async () => {
    await open([file("v", { MediaType: "video", Video: true }), file("a"), file("b")]);
    expect(Photo.prototype.find).toHaveBeenCalledWith("ps6sg6be2lvl0yh7");
    expect(wrapper.vm.groups.map((group) => group.id)).toEqual([file("v").UID, file("a").UID, file("b").UID]);
    expect(wrapper.vm.selected).toEqual([]);
    expect(wrapper.vm.expanded).toEqual([file("v").UID]);
    expect(wrapper.find(".action-export").attributes("disabled")).toBeDefined();
    expect(wrapper.find(".action-export").text()).toBe("Export 0 Files");
    expect(wrapper.find(".export-selected").exists()).toBe(false);
    expect(wrapper.vm.selectedCount(wrapper.vm.groups[0])).toBe(0);
    expect(wrapper.vm.optionLabel(wrapper.vm.groups[1].options[0])).toBe("Original");
    expect(wrapper.vm.optionMetadata(wrapper.vm.groups[1].options[0])).toBe("4000 × 3000 · JPG");
    expect(wrapper.vm.optionBytes(wrapper.vm.groups[1].options[0])).toBe("10 KB");
  });
  it("supports multiple open accordion groups and multiple outputs per original", async () => {
    await open();
    await wrapper.findAll(".v-expansion-panel-title")[1].trigger("click");
    expect(wrapper.vm.expanded).toEqual([file("a").UID, file("b").UID]);
    const original = wrapper.find('input[value="original:fs6sg6bqhhin0097"]');
    await wrapper.find('input[value="thumbnail:fs6sg6bqhhin0097:fit_1920"]').setValue(true);
    await original.setValue(true);
    expect(wrapper.vm.selected).toEqual(["thumbnail:fs6sg6bqhhin0097:fit_1920", "original:fs6sg6bqhhin0097"]);
    expect(wrapper.find(".action-export").text()).toBe("Export 2 Files");
    expect(wrapper.find(".export-selected").text()).toBe("2 Selected");
    expect(wrapper.vm.selectedCount(wrapper.vm.groups[0])).toBe(2);
    await wrapper.setData({ selected: [] });
    expect(wrapper.find(".action-export").attributes("disabled")).toBeDefined();
  });
  it("explains a missing or ambiguous preview and leaves Original unselected", async () => {
    await open([file("r", { Name: "a.cr2", MediaType: "raw", FileType: "cr2" })]);
    expect(wrapper.text()).toContain("A unique still preview could not be found");
    expect(wrapper.vm.groups[0].options).toHaveLength(1);
    expect(wrapper.vm.selected).toEqual([]);
  });
  it("shows a retryable error and then an empty state", async () => {
    await open([], { visible: false });
    Photo.prototype.find.mockRejectedValueOnce(new Error("offline"));
    await wrapper.setProps({ visible: true });
    await flushPromises();
    expect(wrapper.text()).toContain("Failed to load files");
    await wrapper.find(".action-retry").trigger("click");
    await flushPromises();
    expect(wrapper.text()).toContain("No supported files");
    expect(wrapper.find(".action-export").attributes("disabled")).toBeDefined();
  });
  it("shows loading and ignores closed-dialog responses, including after reopening", async () => {
    await open([], { visible: false });
    let finish;
    Photo.prototype.find.mockImplementationOnce(() => new Promise((resolve) => {
      finish = resolve;
    }));
    await wrapper.setProps({ visible: true });
    expect(wrapper.vm.loading).toBe(true);
    expect(wrapper.text()).toContain("Loading");
    wrapper.vm.close();
    await wrapper.setProps({ visible: false });
    await wrapper.setProps({ visible: true });
    await flushPromises();
    finish({ Files: [file("late")] });
    await flushPromises();
    expect(wrapper.vm.groups).toEqual([]);
    expect(wrapper.vm.loading).toBe(false);
  });
  it("resets initial selections on reopening", async () => {
    await open();
    await wrapper.setData({ selected: ["original:fs6sg6bqhhin0098"], expanded: [file("b").UID] });
    await wrapper.setProps({ visible: false });
    await wrapper.setProps({ visible: true });
    await flushPromises();
    expect(wrapper.vm.selected).toEqual([]);
    expect(wrapper.vm.expanded).toEqual([file("a").UID]);
  });
  it("rechecks settings and ACL and closes when either is lost", async () => {
    await open();
    config.values.acl.files.access_all = false;
    await flushPromises();
    expect(wrapper.emitted("close")).toHaveLength(1);
    expect(wrapper.find(".action-export").attributes("disabled")).toBeDefined();
    config.values.acl.files.access_all = true;
    config.values.settings.features.export = false;
    await flushPromises();
    expect(wrapper.vm.permitted).toBe(false);
  });
  it("rejects opening with disabled Export and stops on photo changes", async () => {
    await open();
    await wrapper.setProps({ uid: "ps6sg6be2lvl0yh8" });
    expect(wrapper.emitted("close")).toHaveLength(1);
    wrapper.unmount();
    wrapper = null;
    await open([], { visible: false });
    config.values.settings.features.export = false;
    await wrapper.setProps({ visible: true });
    expect(Photo.prototype.find).not.toHaveBeenCalled();
    expect(wrapper.emitted("close")).toHaveLength(1);
  });
  it("closes only after all frozen selections are handed off without showing download links", async () => {
    await open();
    vi.useFakeTimers();
    await wrapper.setData({ selected: ["original:fs6sg6bqhhin0098", "thumbnail:fs6sg6bqhhin0097:fit_720", "original:fs6sg6bqhhin0097"] });
    const running = wrapper.vm.confirm();
    await flushPromises();
    expect(wrapper.vm.handed).toBe(1);
    expect(wrapper.emitted("close")).toBeUndefined();
    expect(wrapper.vm.activeOutputs).toHaveLength(2);
    expect(wrapper.findAll("input")).toHaveLength(0);
    expect(wrapper.find(".action-export").exists()).toBe(false);
    expect(wrapper.findAll(".export-output")).toHaveLength(2);
    expect(view.focus).toHaveBeenCalledWith(wrapper.vm.$refs.content, ".action-cancel");
    expect(wrapper.findAll(".export-output a")).toHaveLength(0);
    expect(wrapper.vm.downloads.map((item) => item.id)).toEqual(["original:fs6sg6bqhhin0097", "thumbnail:fs6sg6bqhhin0097:fit_720", "original:fs6sg6bqhhin0098"]);
    await wrapper.vm.confirm();
    expect(session.refresh).toHaveBeenCalledOnce();
    await vi.advanceTimersByTimeAsync(500);
    await running;
    expect(wrapper.vm.complete).toBe(true);
    expect(wrapper.emitted("close")).toHaveLength(1);
    expect(wrapper.vm.activeOutputs).toHaveLength(0);
    expect(wrapper.text()).not.toContain("Sent to Browser");
    expect(wrapper.text()).not.toContain("Files have been sent to your browser.");
    expect(wrapper.find(".export-error").exists()).toBe(false);
    expect(HTMLAnchorElement.prototype.click).toHaveBeenCalledTimes(3);
  });
  it("keeps a failed export open and closes after retrying it successfully", async () => {
    await open();
    await wrapper.setData({ selected: [`original:${file("a").UID}`] });
    await wrapper.setData({ selected: [`original:${file("a").UID}`] });
    session.refresh.mockRejectedValueOnce(new Error("expired"));
    await wrapper.vm.confirm();
    expect(wrapper.vm.complete).toBe(false);
    expect(wrapper.vm.transferError).toBe(true);
    expect(wrapper.emitted("close")).toBeUndefined();
    expect(wrapper.find('.export-error[role="alert"]').text()).toBe("Export stopped. Please try again.");
    expect(wrapper.find(".action-close").text()).toBe("Close");
    expect(HTMLAnchorElement.prototype.click).not.toHaveBeenCalled();
    await wrapper.find(".action-retry-output").trigger("click");
    await flushPromises();
    expect(wrapper.vm.handed).toBe(1);
    expect(wrapper.vm.complete).toBe(true);
    expect(wrapper.vm.transferError).toBe(false);
    expect(wrapper.emitted("close")).toHaveLength(1);
    expect(HTMLAnchorElement.prototype.click).toHaveBeenCalledOnce();
  });
  it.each(["cancel", "escape", "close", "uid", "revoke"])("stops pending handoffs on %s", async (action) => {
    await open();
    vi.useFakeTimers();
    await wrapper.setData({ selected: ["original:fs6sg6bqhhin0097", "original:fs6sg6bqhhin0098"] });
    const running = wrapper.vm.confirm();
    await Promise.resolve();
    if (action === "cancel") {
      await wrapper.find(".action-close").trigger("click");
    }
    if (action === "escape") {
      await wrapper.find(".dialog-stub").trigger("keydown", { key: "Escape", code: "Escape" });
    }
    if (action === "close") {
      await wrapper.setProps({ visible: false });
    }
    if (action === "uid") {
      await wrapper.setProps({ uid: "ps6sg6be2lvl0yh8" });
    }
    if (action === "revoke") {
      config.values.acl.files.access_all = false; await Promise.resolve();
    }
    await vi.runAllTimersAsync();
    await running;
    expect(HTMLAnchorElement.prototype.click).toHaveBeenCalledOnce();
    expect(wrapper.vm.handed).toBe(1);
    if (["cancel", "escape", "uid", "revoke"].includes(action)) {
      expect(wrapper.emitted("close")).toHaveLength(1);
    }
  });
  it("reports session refresh failures without handing a file to the browser", async () => {
    await open();
    await wrapper.setData({ selected: [`original:${file("a").UID}`] });
    session.refresh.mockRejectedValueOnce(new Error("expired"));
    await wrapper.vm.confirm();
    expect(wrapper.vm.transferError).toBe(true);
    expect(wrapper.vm.handed).toBe(0);
    expect(HTMLAnchorElement.prototype.click).not.toHaveBeenCalled();
  });
  it("uses the shared focus scope and lets native Tab, Space, and Enter stay in the dialog", async () => {
    await open();
    wrapper.vm.afterEnter();
    expect(view.enter).toHaveBeenCalledWith(wrapper.vm);
    const outside = vi.fn();
    document.addEventListener("keydown", outside);
    for (const key of ["Tab", " ", "Enter"]) {
      const event = new KeyboardEvent("keydown", { key, bubbles: true, cancelable: true });
      wrapper.find(".dialog-stub").element.dispatchEvent(event);
      expect(event.defaultPrevented).toBe(false);
    }
    expect(outside).not.toHaveBeenCalled();
    document.removeEventListener("keydown", outside);
    wrapper.vm.afterLeave();
    expect(view.leave).toHaveBeenCalledWith(wrapper.vm);
    expect(wrapper.emitted("leave")).toHaveLength(1);
  });
});

describe("video export dialog", () => {
  it("does not label native motion-video extraction as conversion", async () => {
    const motion = file("m", { Name: "motion.jpg", MediaType: "live", Video: true });
    Axios.get.mockResolvedValue({ data: { videoPart: { preparable: true, ready: false, format: "mp4", codec: "hvc1", bytes: 0 } } });
    await open([motion]);
    expect(wrapper.find('.export-option[data-option-id^="video-part"]').exists()).toBe(true);
    expect(wrapper.find(".export-conversion-note").exists()).toBe(false);
  });
  it("shows separate Apple groups and counts two Original selections as two outputs", async () => {
    const image = file("a", { Name: "live.jpg" });
    const movie = file("b", { Name: "live.mov", FileType: "mov", MediaType: "video", Video: true });
    Axios.get.mockImplementation(async (url) => ({ data: url.includes(movie.Hash) ? { appleKey: "live-pair", codec: "avc1", duration: 3 } : { appleKey: "live-pair" } }));
    await open([movie, image]);
    expect(wrapper.vm.groups).toHaveLength(2);
    expect(wrapper.vm.groups.map((group) => group.id)).toEqual([image.UID, movie.UID]);
    expect(wrapper.vm.groups[1].options.map((option) => option.kind)).toEqual(["original"]);
    expect(wrapper.vm.groups[0].apple).toBe(true);
    expect(wrapper.text()).toContain("Live Photo");
    await wrapper.setData({ selected: [`original:${image.UID}`, `original:${movie.UID}`] });
    expect(wrapper.vm.selectedCount(wrapper.vm.groups[0])).toBe(1);
    expect(wrapper.vm.selectedCount(wrapper.vm.groups[1])).toBe(1);
    expect(wrapper.find(".action-export").text()).toBe("Export 2 Files");
    expect(wrapper.findAll(".export-filename").map((item) => item.text())).toEqual(["live.jpg", "live.mov"]);
  });
  it("shows preparing without blocking an Original and retries only the failed AVC", async () => {
    vi.useFakeTimers();
    const movie = file("v", { Name: "movie.mov", FileType: "mov", MediaType: "video", Video: true });
    let finish;
    const requests = [];
    Axios.get.mockImplementation(async (url) => {
      requests.push(url);
      if (url.includes("export=probe")) {
        return { data: { codec: "hvc1", avc: { preparable: true, ready: false, format: "mp4", codec: "avc1", bytes: 0 } } };
      }
      if (requests.filter((request) => request.includes("export=prepare")).length === 1) {
        return new Promise((_resolve, reject) => {
          finish = reject;
        });
      }
      return { data: { ready: true, codec: "avc1", format: "mp4", width: 1280, height: 720, bytes: 1000 } };
    });
    await open([movie]);
    expect(wrapper.find('.export-option[data-option-id^="avc"] .export-option-size').exists()).toBe(false);
    expect(wrapper.find('.export-option[data-option-id^="avc"] .export-conversion-note').text()).toBe("Needs Conversion");
    await wrapper.setData({ selected: wrapper.vm.groups[0].options.map((option) => option.id) });
    const running = wrapper.vm.confirm();
    await flushPromises();
    expect(wrapper.vm.handed).toBe(1);
    expect(wrapper.vm.preparing).toBe(1);
    expect(wrapper.text()).toContain("Preparing");
    expect(wrapper.find('.export-option[data-option-id^="avc"] .export-option-size').exists()).toBe(false);
    expect(wrapper.find(".export-output[data-option-id^=avc] .v-progress-circular").exists()).toBe(true);
    expect(wrapper.findAll(".export-output a")).toHaveLength(0);
    expect(wrapper.findAll(".export-output")).toHaveLength(1);
    expect(wrapper.emitted("close")).toBeUndefined();
    expect(wrapper.vm.downloads.map((item) => item.filename)).toEqual(["movie.mov", "movie_avc.mp4"]);
    finish(new Error("encoder failed"));
    await running;
    await flushPromises();
    expect(wrapper.vm.errors).toBe(1);
    expect(wrapper.vm.transferError).toBe(true);
    expect(wrapper.findAll(".action-retry-output")).toHaveLength(1);
    expect(wrapper.find('.export-output[data-option-id^="avc"]').text()).toContain("Failed");
    session.refresh.mockImplementation(async () => {
      config.downloadToken = "rotated";
    });
    await wrapper.find(".action-retry-output").trigger("click");
    await vi.advanceTimersByTimeAsync(250);
    expect(requests.at(-1)).toContain("/rotated/export?export=prepare&kind=avc");
    expect(wrapper.vm.handed).toBe(2);
    expect(wrapper.vm.errors).toBe(0);
    expect(wrapper.vm.complete).toBe(true);
    expect(wrapper.vm.transferError).toBe(false);
    expect(HTMLAnchorElement.prototype.click).toHaveBeenCalledTimes(2);
    expect(wrapper.emitted("close")).toHaveLength(1);
    expect(wrapper.findAll(".export-output a")).toHaveLength(0);
    const avc = wrapper.vm.downloads.find((item) => item.kind === "avc");
    expect(wrapper.vm.optionMetadata(avc)).toBe("1280 × 720 · MP4 · avc1");
    expect(wrapper.vm.optionBytes(avc)).toBe("1 KB");
    expect(wrapper.vm.needsConversion(avc)).toBe(false);
  });
  it.each(["cancel", "escape", "photo", "lightbox", "revoke", "unmount"])("ignores a late preparation after %s", async (action) => {
    let finish;
    const movie = file("v", { Name: "movie.mov", FileType: "mov", MediaType: "video", Video: true });
    Axios.get.mockImplementation(async (url) => {
      if (url.includes("export=probe")) {
        return { data: { avc: { preparable: true, ready: false, format: "mp4" } } };
      }
      return new Promise((resolve) => {
        finish = resolve;
      });
    });
    await open([movie]);
    const dialog = wrapper.vm;
    await wrapper.setData({ selected: [`avc:${movie.UID}`] });
    const running = wrapper.vm.confirm();
    await flushPromises();
    expect(wrapper.vm.preparing).toBe(1);
    if (action === "cancel") {
      wrapper.vm.close();
    } else if (action === "escape") {
      await wrapper.find(".dialog-stub").trigger("keydown", { key: "Escape", code: "Escape" });
    } else if (action === "photo") {
      await wrapper.setProps({ uid: "ps6sg6be2lvl0yh8" });
    } else if (action === "lightbox") {
      await wrapper.setProps({ visible: false });
    } else if (action === "unmount") {
      wrapper.unmount();
      wrapper = null;
    } else {
      config.values.acl.files.access_all = false;
      await flushPromises();
    }
    await running;
    finish({ data: { ready: true, format: "mp4", codec: "avc1" } });
    await flushPromises();
    expect(dialog.handed).toBe(0);
    expect(dialog.states[`avc:${movie.UID}`].state).toBe("preparing");
    expect(HTMLAnchorElement.prototype.click).not.toHaveBeenCalled();
  });
  it("rediscovers ready cache on reopening and drops a late probe from the old opening", async () => {
    const movie = file("v", { Name: "movie.mov", FileType: "mov", MediaType: "video", Video: true });
    let finish;
    await open([movie], { visible: false });
    Axios.get.mockImplementationOnce(() => new Promise((resolve) => {
      finish = resolve;
    }));
    await wrapper.setProps({ visible: true });
    await flushPromises();
    expect(wrapper.vm.loading).toBe(true);
    wrapper.vm.close();
    await wrapper.setProps({ visible: false });
    Axios.get.mockResolvedValue({ data: { avc: { ready: true, preparable: false, format: "mp4", codec: "avc1" } } });
    await wrapper.setProps({ visible: true });
    await flushPromises();
    expect(wrapper.vm.groups[0].options[1].ready).toBe(true);
    expect(wrapper.vm.selected).toEqual([]);
    finish({ data: { avc: { ready: false, preparable: false, format: "mp4" } } });
    await flushPromises();
    expect(wrapper.vm.groups[0].options[1].ready).toBe(true);
  });
});

describe("indexed animated-image probing", () => {
  it("probes image PNG/WebP animations using file Frames and Duration", async () => {
    const png = file("a", { Name: "animated.png", FileType: "png", Frames: 2 });
    const webp = file("b", { Name: "animated.webp", FileType: "webp", Duration: 1000000000 });
    Axios.get.mockResolvedValue({ data: { avc: { ready: false, preparable: true, format: "mp4", codec: "avc1" } } });
    await open([png, webp]);
    expect(Axios.get).toHaveBeenCalledTimes(2);
    expect(wrapper.vm.groups.map((group) => group.original.mediaType)).toEqual(["animated", "animated"]);
    expect(wrapper.vm.groups.every((group) => group.options.map((option) => option.kind).join() === "original,avc")).toBe(true);
  });
});


describe("export feedback regressions", () => {
  it("keeps counts local to JPEG original groups and suppresses misleading RAW errors", async () => {
    const raw = file("r", { Name: "capture.cr2", MediaType: "raw", FileType: "cr2" });
    const jpeg = file("j", { Name: "capture.jpg" });
    await open([raw, jpeg]);
    await wrapper.setData({ expanded: [raw.UID, jpeg.UID] });
    expect(wrapper.text()).not.toContain("unique still preview");
    expect(wrapper.vm.groups[0].options).toHaveLength(1);
    await wrapper.find(`input[value="original:${jpeg.UID}"]`).setValue(true);
    expect(wrapper.vm.selectedCount(wrapper.vm.groups[0])).toBe(0);
    expect(wrapper.vm.selectedCount(wrapper.vm.groups[1])).toBe(1);
    expect(wrapper.findAll(".export-selected").map((chip) => chip.text())).toEqual(["1 Selected"]);
    expect(wrapper.find(".action-export").text()).toBe("Export 1 File");
  });
  it("renders real preview errors in the error color and labels the PDF first page", async () => {
    const pdf = file("d", { Name: "document.pdf", MediaType: "document", FileType: "pdf" });
    await open([pdf]);
    expect(wrapper.find(".text-error").text()).toContain("unique still preview");
    Photo.prototype.find.mockResolvedValue({ Files: [pdf, file("p", { Name: "document.pdf.jpg", Root: "sidecar" })] });
    await wrapper.vm.loadFiles();
    await flushPromises();
    expect(wrapper.text()).toContain("Full Size Preview (First Page)");
    expect(wrapper.text()).not.toContain("unique still preview");
  });
  it("surfaces failed probes and retries one group without losing another group's selection", async () => {
    const movie = file("m", { Name: "movie.mov", FileType: "mov", MediaType: "video" });
    const still = file("s");
    Axios.get.mockImplementation((url) => url.includes(movie.Hash) ? Promise.reject({ response: { status: 503 } }) : Promise.resolve({ data: {} }));
    await open([movie, still]);
    expect(wrapper.find(".export-probe-error").text()).toContain("Failed to check video options");
    expect(wrapper.vm.groups[0].options.map((option) => option.kind)).toEqual(["original"]);
    await wrapper.setData({ selected: [`original:${still.UID}`] });
    Axios.get.mockResolvedValue({ data: { codec: "hevc", avc: { format: "mp4", codec: "avc1", preparable: true } } });
    await wrapper.find(".action-retry-probe").trigger("click");
    await flushPromises();
    expect(wrapper.find(".export-probe-error").exists()).toBe(false);
    expect(wrapper.vm.groups[0].options.some((option) => option.kind === "avc")).toBe(true);
    expect(wrapper.vm.selected).toEqual([`original:${still.UID}`]);
  });
  it("ignores a manual probe retry that completes after reopening", async () => {
    const movie = file("m", { Name: "movie.mov", FileType: "mov", MediaType: "video" });
    Axios.get.mockRejectedValue({ response: { status: 503 } });
    await open([movie]);
    let resolve;
    Axios.get.mockImplementationOnce(() => new Promise((done) => {
      resolve = done;
    }));
    const pending = wrapper.vm.retryProbe(wrapper.vm.groups[0]);
    await Promise.resolve();
    await wrapper.setProps({ visible: false });
    Axios.get.mockResolvedValue({ data: {} });
    await wrapper.setProps({ visible: true });
    await flushPromises();
    resolve({ data: { avc: { format: "mp4", preparable: true } } });
    await pending;
    expect(wrapper.vm.groups[0].options.map((option) => option.kind)).toEqual(["original"]);
    expect(wrapper.vm.selected).toEqual([]);
  });
});
