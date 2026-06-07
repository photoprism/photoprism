import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { shallowMount, config as VTUConfig } from "@vue/test-utils";

import PTabDiscoverSeason from "page/discover/season.vue";
import "../fixtures";

describe("page/discover/season.vue", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-06-07T12:00:00Z"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("links to browse photos from today's month and day", () => {
    const wrapper = shallowMount(PTabDiscoverSeason, {
      global: {
        mocks: {
          $config: VTUConfig.global.mocks.$config,
          $gettext: (text) => text,
          $gettextInterpolate: (text, values) => text.replace("%{date}", values.date),
        },
      },
    });

    expect(wrapper.vm.memoriesQuery).toEqual({
      month: 6,
      day: 7,
      order: "oldest",
    });
    expect(wrapper.vm.todayLabel).toBe("June 7");
  });
});
