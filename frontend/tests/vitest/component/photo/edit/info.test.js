import { describe, it, expect, vi } from "vitest";
import { shallowMount } from "@vue/test-utils";
import PTabPhotoInfo from "component/photo/edit/info.vue";

// copyText records the values the tab copies to the clipboard.
const copyText = vi.fn();

// makeWrapper mounts the info tab with a photo whose time has the given source.
const makeWrapper = (takenSrc, name = "") => {
  const model = {
    UID: "ps6sg6be2lvl0yh7",
    Name: name,
    Title: "Example",
    TitleSrc: "",
    TakenSrc: takenSrc,
    Albums: [],
    getDateString: () => "April 30, 2024",
    locationInfo: () => "",
    update: vi.fn(),
  };

  return shallowMount(PTabPhotoInfo, {
    props: { uid: model.UID },
    global: {
      mocks: {
        $gettext: (s) => s,
        $util: { sourceName: (s, d) => s || d, copyText },
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
