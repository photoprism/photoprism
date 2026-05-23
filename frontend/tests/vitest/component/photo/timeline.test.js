import { describe, it, expect, vi, beforeEach } from "vitest";
import { shallowMount, config as VTUConfig } from "@vue/test-utils";

import PPhotoViewTimeline from "component/photo/view/timeline.vue";
import { buildTimelineSections, photoLocalDateParts } from "component/photo/view/timeline";
import "../../fixtures";

function photo(values) {
  return {
    ID: values.ID,
    UID: values.UID || `uid-${values.ID}`,
    Type: values.Type || "image",
    TakenAt: values.TakenAt || "",
    TakenAtLocal: values.TakenAtLocal || "",
    Year: values.Year || -1,
    Month: values.Month || -1,
    Day: values.Day || -1,
    Private: false,
    Favorite: false,
    Title: values.Title || "",
    getOriginalName: vi.fn(() => values.Title || `photo-${values.ID}.jpg`),
    thumbnailUrl: vi.fn(() => `/thumb-${values.ID}.jpg`),
    classes: vi.fn(() => ["is-photo", "type-image"]),
    isStack: vi.fn(() => false),
    videoContentType: vi.fn(() => "video/mp4"),
    videoUrl: vi.fn(() => `/video-${values.ID}.mp4`),
    getDurationInfo: vi.fn(() => ""),
    toggleLike: vi.fn(),
  };
}

function mountTimeline(props = {}) {
  const configMock = {
    ...VTUConfig.global.mocks.$config,
    get: vi.fn(() => false),
    getSettings: vi.fn(() => ({
      features: { private: true },
      search: { showTitles: true, showCaptions: false },
    })),
  };

  return shallowMount(PPhotoViewTimeline, {
    props: {
      photos: [],
      filter: { order: "newest" },
      selectMode: false,
      isSharedView: false,
      openPhoto: vi.fn(),
      editPhoto: vi.fn(),
      ...props,
    },
    global: {
      mocks: {
        $config: configMock,
        $clipboard: { toggle: vi.fn(), addRange: vi.fn() },
        $isMobile: false,
      },
      stubs: {
        IconLivePhoto: true,
      },
    },
  });
}

describe("component/photo/view/timeline", () => {
  beforeEach(() => {
    global.IntersectionObserver = class IntersectionObserver {
      observe() {}
      unobserve() {}
      disconnect() {}
    };
  });

  it("uses TakenAtLocal for grouping without timezone conversion", () => {
    const parts = photoLocalDateParts({
      TakenAt: "2024-07-01T01:00:00Z",
      TakenAtLocal: "2024-06-30T23:00:00",
      Year: 2024,
      Month: 7,
      Day: 1,
    });

    expect(parts).toEqual({
      known: true,
      year: 2024,
      month: 6,
      day: 30,
    });
  });

  it("groups photos by local month and day while preserving result indexes", () => {
    const photos = [
      photo({ ID: 1, TakenAtLocal: "2024-06-30T23:00:00" }),
      photo({ ID: 2, TakenAtLocal: "2024-06-30T09:30:00" }),
      photo({ ID: 3, TakenAtLocal: "2024-05-03T12:00:00" }),
      photo({ ID: 4 }),
    ];
    const sections = buildTimelineSections(
      photos,
      [
        { value: 5, text: "May" },
        { value: 6, text: "June" },
      ],
      (s) => s
    );

    expect(sections.map((section) => section.key)).toEqual(["2024-06", "2024-05", "unknown"]);
    expect(sections[0].title).toBe("June 2024");
    expect(sections[0].countLabel).toBe("2 pictures");
    expect(sections[0].days[0].title).toBe("30");
    expect(sections[0].days[0].entries.map((entry) => entry.index)).toEqual([0, 1]);
    expect(sections[2].title).toBe("Unknown date");
    expect(sections[2].days[0].entries[0].index).toBe(3);
  });

  it("renders a separate scrollable timeline view", () => {
    const wrapper = mountTimeline({
      photos: [photo({ ID: 1, TakenAtLocal: "2024-06-30T23:00:00" }), photo({ ID: 2, TakenAtLocal: "2024-05-03T12:00:00" })],
    });

    expect(wrapper.find(".timeline-view").exists()).toBe(true);
    expect(wrapper.findAll(".timeline-section")).toHaveLength(2);
    expect(wrapper.find(".timeline-section__header h2").text()).toBe("June 2024");
    expect(wrapper.find(".timeline-day__label").text()).toBe("30");
  });

  it("keeps the original result index when opening a photo", () => {
    const openPhoto = vi.fn();
    const wrapper = mountTimeline({
      openPhoto,
      photos: [
        photo({ ID: 1, TakenAtLocal: "2024-06-30T23:00:00" }),
        photo({ ID: 2, TakenAtLocal: "2024-05-03T12:00:00" }),
        photo({ ID: 3, TakenAtLocal: "2024-05-02T12:00:00" }),
      ],
    });

    wrapper.vm.input.mouseDown({ timeStamp: 100 }, 2);
    wrapper.vm.onClick({ timeStamp: 200, shiftKey: false }, 2);

    expect(openPhoto).toHaveBeenCalledWith(2);
  });
});
