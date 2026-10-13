import { describe, it, expect, vi } from "vitest";
import { shallowMount } from "@vue/test-utils";
import PTabPhotoInfo from "component/photo/edit/info.vue";
import $util from "common/util";

// copyText records the values the tab copies to the clipboard.
const copyText = vi.fn();

// makeWrapper mounts the info tab with a photo whose time has the given source.
const makeWrapper = (takenSrc, name = "", albums = []) => {
  const model = {
    UID: "ps6sg6be2lvl0yh7",
    Name: name,
    Title: "Example",
    TitleSrc: "",
    TakenSrc: takenSrc,
    Albums: albums,
    getDateString: () => "April 30, 2024",
    locationInfo: () => "",
    update: vi.fn(),
  };

  return shallowMount(PTabPhotoInfo, {
    props: { uid: model.UID },
    global: {
      mocks: {
        $gettext: (s) => s,
        $util: { sourceName: (s, d) => s || d, copyText, balanceBidi: (s) => $util.balanceBidi(s) },
        $router: { resolve: () => ({ href: "/library/albums/view" }) },
        $config: {
          values: {},
          getTimeZone: () => "UTC",
          get: () => false,
          feature: () => false,
        },
        $view: { getData: () => ({ model }) },
      },
      directives: { tooltip: {} },
      stubs: {
        VForm: { template: "<form><slot /></form>" },
        VTable: { template: "<table><slot /></table>" },
      },
    },
  });
};

// takenIcons returns the source icons, which only the time shows while TitleSrc and PlaceSrc are empty.
const takenIcons = (wrapper) => wrapper.findAll("v-icon-stub.src").map((icon) => icon.attributes("icon"));

describe("component/photo/edit/info", () => {
  it("removes bidi controls from displayed names", async () => {
    const wrapper = makeWrapper("meta", "2026/photo\u202Eabc.jpg");
    expect(wrapper.vm.visibleName("صورة.jpg")).toBe("صورة.jpg");

    const name = wrapper.findAll("span.cursor-copy").find((s) => s.text().includes("abc"));
    expect(name.text()).toBe("2026/photoabc.jpg");
    expect(name.attributes("dir")).toBe("ltr");

    await name.trigger("click");
    expect(copyText).toHaveBeenCalledWith("2026/photo\u202Eabc.jpg");
  });

  it("isolates album titles and balances their isolates", () => {
    const wrapper = makeWrapper("meta", "", [
      { UID: "as6sg6bxpogaaba7", Title: "رحلة 2026" },
      { UID: "as6sg6bxpogaaba8", Title: "Summer\u2069\u200F 2026" },
    ]);
    const titles = wrapper.findAll("a.text-link bdi");
    expect(titles.map((t) => t.text())).toEqual(["رحلة 2026", "Summer\u200F 2026"]);
  });

  it("shows an icon for a time taken from the modify time", () => {
    expect(takenIcons(makeWrapper("modified"))).toEqual(["mdi-clock-edit-outline"]);
  });

  it("shows an icon for a time taken from the file name", () => {
    expect(takenIcons(makeWrapper("name"))).toEqual(["mdi-file-tree-outline"]);
  });

  it("shows no icon for a capture time from metadata", () => {
    expect(takenIcons(makeWrapper("meta"))).toEqual([]);
  });
});
