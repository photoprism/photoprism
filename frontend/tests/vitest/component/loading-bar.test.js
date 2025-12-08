import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount } from "@vue/test-utils";
import { nextTick } from "vue";
import PLoadingBar from "component/loading-bar.vue";

// Mock $event subscription
const mockSubscribe = vi.fn();

// Mock queue function to execute callbacks immediately
vi.mock("component/loading-bar.vue", async (importOriginal) => {
  const actual = await importOriginal();
  return {
    ...actual,
    queue: (fn) => {
      fn((next) => {
        if (next) next();
      });
    },
  };
});

describe("PLoadingBar component", () => {
  let wrapper;

  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers();

    wrapper = mount(PLoadingBar, {
      global: {
        mocks: {
          $event: {
            subscribe: mockSubscribe,
          },
        },
        stubs: {
          transition: false,
        },
      },
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    if (wrapper) {
      wrapper.unmount();
    }
  });

  describe("Component Initialization", () => {
    it("should render container element", () => {
      const container = wrapper.find("#p-loading-bar");
      expect(container.exists()).toBe(true);
    });

    it("should not show progress bar initially", () => {
      const progressBar = wrapper.find(".top-progress");
      expect(progressBar.exists()).toBe(false);
    });

    it("should subscribe to AJAX events on mount", () => {
      expect(mockSubscribe).toHaveBeenCalledTimes(2);
      expect(mockSubscribe.mock.calls[0][0]).toBe("ajax.start");
      expect(mockSubscribe.mock.calls[1][0]).toBe("ajax.end");
    });
  });

  describe("Progress Bar Visibility", () => {
    it("should show progress bar when started", async () => {
      // Start the loading bar
      wrapper.vm.start();
      await nextTick();

      const progressBar = wrapper.find(".top-progress");
      expect(progressBar.exists()).toBe(true);
    });

    it("should hide progress bar when completed", async () => {
      // Start and complete the loading bar
      wrapper.vm.start();
      await nextTick();

      wrapper.vm.done();
      await nextTick();

      // Fast forward through completion animation
      vi.advanceTimersByTime(1000);
      await nextTick();

      const progressBar = wrapper.find(".top-progress");
      expect(progressBar.exists()).toBe(false);
    });
  });

  describe("Progress Bar Appearance", () => {
    beforeEach(async () => {
      wrapper.vm.start();
      await nextTick();
    });

    it("should display with default color", async () => {
      const progressBar = wrapper.find(".top-progress");
      const barStyle = progressBar.attributes("style");

      expect(barStyle).toContain("background-color: rgb(34, 153, 221)"); // #29d
    });

    it("should display with error color when failed", async () => {
      wrapper.vm.fail();
      await nextTick();

      const progressBar = wrapper.find(".top-progress");
      const barStyle = progressBar.attributes("style");

      expect(barStyle).toContain("background-color: rgb(244, 67, 54)"); // #f44336
    });

    it("should have peg element for styling", () => {
      const peg = wrapper.find(".peg");
      expect(peg.exists()).toBe(true);
    });
  });

  describe("Progress Bar Behavior", () => {
    it("should show progress bar when started", async () => {
      wrapper.vm.start();
      await nextTick();

      const progressBar = wrapper.find(".top-progress");
      expect(progressBar.exists()).toBe(true);
      expect(progressBar.attributes("style")).toContain("width:");
    });

    it("should pause progress when paused", async () => {
      wrapper.vm.start();
      wrapper.vm.pause();
      await nextTick();

      // When paused, progress bar should still exist but not advance
      const progressBar = wrapper.find(".top-progress");
      expect(progressBar.exists()).toBe(true);
    });

    it("should complete progress when done is called", async () => {
      wrapper.vm.start();
      await nextTick();

      const progressBar = wrapper.find(".top-progress");
      expect(progressBar.exists()).toBe(true);

      wrapper.vm.done();

      // Allow time for completion animation
      vi.advanceTimersByTime(1000);
      await nextTick();

      // Progress bar should disappear after completion
      expect(wrapper.find(".top-progress").exists()).toBe(false);
    });
  });
});
