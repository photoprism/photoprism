import { describe, it, expect, vi } from "vitest";
import { shallowMount } from "@vue/test-utils";
import PTabPhotoInfo from "component/photo/edit/info.vue";

// makeWrapper mounts the info tab with a photo whose time has the given source.
const makeWrapper = (takenSrc) => {
  const model = {
    UID: "ps6sg6be2lvl0yh7",
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
        $util: { sourceName: (s, d) => s || d, copyText: vi.fn() },
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
