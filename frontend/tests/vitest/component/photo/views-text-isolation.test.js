import { describe, it, expect, vi, beforeEach } from "vitest";
import { shallowMount } from "@vue/test-utils";
import PPhotoViewCards from "component/photo/view/cards.vue";
import PPhotoViewList from "component/photo/view/list.vue";
import * as contexts from "options/contexts";
import "../../fixtures";

function createPhotoStub(reason = "", overrides = {}) {
  return {
    ID: 1,
    UID: "ps123",
    Title: "صورة من العائلة",
    Caption: "شاطئ البحر",
    Quality: 5,
    Type: "image",
    Year: 0,
    CameraID: 0,
    Iso: 0,
    LensID: 0,
    FocalLength: 0,
    Country: "zz",
    Favorite: false,
    Private: false,
    CameraMake: "",
    CameraModel: "",
    classes: () => "",
    thumbnailUrl: () => "/thumb.jpg",
    isStack: () => false,
    getOriginalName: () => "صورة.dng",
    getDateString: () => "Sunday, July 8, 2012",
    shortDateString: () => "Jul 8, 2012",
    getImageInfo: () => "10 MP",
    getVideoInfo: () => "00:03",
    getVectorInfo: () => "PDF",
    locationInfo: () => "Unknown",
    getHiddenReason: () => reason,
    ...overrides,
  };
}

function createConfigMock(showTitles = true) {
  return {
    getSettings: () => ({
      features: {
        places: true,
        private: true,
        download: true,
      },
      search: {
        showTitles,
        showCaptions: true,
      },
    }),
    feature: () => true,
    get: () => false,
    values: {
      settings: {
        features: {
          places: true,
          private: true,
        },
      },
    },
  };
}

function mountCards(context, showTitles = true, photo = createPhotoStub()) {
  return shallowMount(PPhotoViewCards, {
    props: {
      photos: [photo],
      context,
      filter: { order: "newest" },
      selectMode: false,
      isSharedView: false,
      openPhoto: vi.fn(),
      editPhoto: vi.fn(),
      openDate: vi.fn(),
      openLocation: vi.fn(),
    },
    global: {
      mocks: {
        $config: createConfigMock(showTitles),
      },
      stubs: {
        IconLivePhoto: true,
      },
    },
  });
}

function mountList(context, showTitles = true, photo = createPhotoStub()) {
  return shallowMount(PPhotoViewList, {
    props: {
      photos: [photo],
      context,
      filter: { order: "newest" },
      selectMode: false,
      isSharedView: false,
      openPhoto: vi.fn(),
      editPhoto: vi.fn(),
      openDate: vi.fn(),
      openLocation: vi.fn(),
    },
    global: {
      mocks: {
        $config: createConfigMock(showTitles),
      },
      stubs: {
        IconLivePhoto: true,
      },
    },
  });
}

describe("component/photo/view text isolation", () => {
  beforeEach(() => {
    global.IntersectionObserver = class IntersectionObserver {
      observe() {}
      unobserve() {}
      disconnect() {}
    };
  });

  it("isolates titles, captions and file names in cards view", async () => {
    const wrapper = mountCards(contexts.Photos);
    const title = wrapper.find(".result:not(.placeholder) .meta-title bdi");
    expect(title.text()).toBe("صورة من العائلة");
    expect(wrapper.find(".result:not(.placeholder) .meta-title").attributes("dir")).toBeUndefined();
    expect(wrapper.find(".result:not(.placeholder) .meta-caption bdi").text()).toBe("شاطئ البحر");
    expect(wrapper.find(".meta-filename bdi").text()).toBe("صورة.dng");
    expect(wrapper.find(".meta-filename bdi").attributes("dir")).toBe("ltr");
    expect(wrapper.find(".result:not(.placeholder) .meta-title bdi").attributes("dir")).toBe("auto");

    await wrapper.setData({ firstVisibleElementIndex: 1, lastVisibleElementIndex: 1 });
    expect(wrapper.find(".placeholder .meta-title bdi").text()).toBe("صورة من العائلة");
    expect(wrapper.find(".placeholder .meta-caption bdi").text()).toBe("شاطئ البحر");
  });

  it("isolates the title in list view", () => {
    const wrapper = mountList(contexts.Photos);
    expect(wrapper.find(".text-truncate bdi").text()).toBe("صورة من العائلة");
  });

  it("balances isolates in titles and captions", () => {
    const photo = createPhotoStub("", { Title: "رحلة\u2069\u200F 2026", Caption: "a\u2029b\u2067c" });
    const cards = mountCards(contexts.Photos, true, photo);
    expect(cards.find(".result:not(.placeholder) .meta-title bdi").text()).toBe("رحلة\u200F 2026");
    expect(cards.find(".result:not(.placeholder) .meta-caption bdi").text()).toBe("a b\u2067c\u2069");
    expect(mountList(contexts.Photos, true, photo).find(".text-truncate bdi").text()).toBe("رحلة\u200F 2026");
  });

  it("shows file names left to right when titles are hidden", () => {
    const cards = mountCards(contexts.Photos, false);
    expect(cards.find(".result:not(.placeholder) .meta-title bdi").attributes("dir")).toBe("ltr");
    expect(cards.find(".result:not(.placeholder) .meta-title bdi").text()).toBe("صورة.dng");

    const list = mountList(contexts.Photos, false);
    expect(list.find(".text-truncate bdi").attributes("dir")).toBe("ltr");
    expect(mountList(contexts.Photos).find(".text-truncate bdi").attributes("dir")).toBe("auto");
  });
});
