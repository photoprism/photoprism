import { describe, it, expect, vi, beforeEach } from "vitest";
import { shallowMount } from "@vue/test-utils";
import PPhotoViewCards from "component/photo/view/cards.vue";
import PPhotoViewList from "component/photo/view/list.vue";
import * as contexts from "options/contexts";
import "../../fixtures";

function createPhotoStub(reason = "unsupported raw format") {
  return {
    ID: 1,
    UID: "ps123",
    Title: "Hidden Sample",
    Caption: "",
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
    getOriginalName: () => "IMG_0001.DNG",
    getDateString: () => "Sunday, July 8, 2012",
    shortDateString: () => "Jul 8, 2012",
    getImageInfo: () => "10 MP",
    getVideoInfo: () => "00:03",
    getVectorInfo: () => "PDF",
    locationInfo: () => "Unknown",
    getHiddenReason: () => reason,
  };
}

function createConfigMock() {
  return {
    getSettings: () => ({
      features: {
        places: true,
        private: true,
        download: true,
      },
      search: {
        showTitles: true,
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

function mountCards(context) {
  return shallowMount(PPhotoViewCards, {
    props: {
      photos: [createPhotoStub()],
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
        $config: createConfigMock(),
      },
      stubs: {
        IconLivePhoto: true,
      },
    },
  });
}

function mountList(context) {
  return shallowMount(PPhotoViewList, {
    props: {
      photos: [createPhotoStub()],
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
        $config: createConfigMock(),
      },
      stubs: {
        IconLivePhoto: true,
      },
    },
  });
}

describe("component/photo/view hidden errors", () => {
  beforeEach(() => {
    global.IntersectionObserver = class IntersectionObserver {
      observe() {}
      unobserve() {}
      disconnect() {}
    };
  });

  it("renders hidden reason in cards view for hidden context", () => {
    const wrapper = mountCards(contexts.Hidden);
    expect(wrapper.find(".meta-error").exists()).toBe(true);
    expect(wrapper.html()).toContain("unsupported raw format");
  });

  it("does not render hidden reason in cards view outside hidden context", () => {
    const wrapper = mountCards(contexts.Photos);
    expect(wrapper.find(".meta-error").exists()).toBe(false);
  });

  it("renders hidden reason in list view for hidden context", () => {
    const wrapper = mountList(contexts.Hidden);
    expect(wrapper.find(".meta-error").exists()).toBe(true);
    expect(wrapper.html()).toContain("unsupported raw format");
  });

  it("does not render hidden reason in list view outside hidden context", () => {
    const wrapper = mountList(contexts.Photos);
    expect(wrapper.find(".meta-error").exists()).toBe(false);
  });
});
