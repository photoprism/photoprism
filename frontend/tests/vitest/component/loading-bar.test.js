import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount } from "@vue/test-utils";
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
  });

  it("should render correctly", () => {
    expect(wrapper.vm).toBeTruthy();
    expect(wrapper.find("#p-loading-bar").exists()).toBe(true);
    expect(wrapper.find(".top-progress").exists()).toBe(false); // Initially not visible

    // Check computed properties
    expect(wrapper.vm.progressColor).toBe("#29d"); // Default color
    expect(wrapper.vm.isStarted).toBe(false);
  });

  it("should subscribe to ajax events on mount", () => {
    expect(mockSubscribe).toHaveBeenCalledTimes(2);
    expect(mockSubscribe.mock.calls[0][0]).toBe("ajax.start");
    expect(mockSubscribe.mock.calls[1][0]).toBe("ajax.end");
  });

  it("should start the loading bar", async () => {
    wrapper.vm.start();

    await wrapper.vm.$nextTick();
    expect(wrapper.vm.visible).toBe(true);

    // After transition, the bar should be displayed
    wrapper.vm.afterEnter();
    expect(wrapper.vm.status).not.toBeNull();
  });

  it("should make progress visible when started", async () => {
    expect(wrapper.vm.visible).toBe(false);

    // Start the bar
    wrapper.vm.start();
    await wrapper.vm.$nextTick();

    // Should be visible now
    expect(wrapper.vm.visible).toBe(true);
  });

  it("should handle error state", async () => {
    wrapper.vm.fail();
    await wrapper.vm.$nextTick();

    expect(wrapper.vm.error).toBe(true);
    expect(wrapper.vm.progressColor).toBe("#f44336"); // Error color
  });

  it("should pause the loading bar", () => {
    wrapper.vm.start();
    wrapper.vm.pause();

    expect(wrapper.vm.isPaused).toBe(true);
  });

  it("should initialize progress to zero", () => {
    expect(wrapper.vm.progress).toBe(0);

    expect(wrapper.vm.getProgress()).toBe(0);
  });
});
