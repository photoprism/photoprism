import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { nextTick } from "vue";
import PLocationInput from "component/location/input.vue";

describe("PLocationInput", () => {
  let wrapper;

  const defaultProps = {
    latlng: [null, null],
    disabled: false,
    hideDetails: true,
    label: "Location",
    placeholder: "37.75267, -122.543",
    density: "comfortable",
    validateOn: "input",
    showMapButton: false,
    icon: "mdi-map-marker",
    mapButtonTitle: "Open Map",
    mapButtonDisabled: false,
    enableUndo: false,
    autoApply: true,
    debounceDelay: 1000,
  };

  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount();
    }
    vi.useRealTimers();
    vi.clearAllTimers();
  });

  const createWrapper = (props = {}) => {
    return mount(PLocationInput, {
      props: { ...defaultProps, ...props },
    });
  };

  describe("Component Rendering", () => {
    it("should render input field with correct placeholder", () => {
      const placeholder = "Custom placeholder";
      wrapper = createWrapper({ placeholder });

      const input = wrapper.find("input");
      expect(input.exists()).toBe(true);
      expect(input.attributes("placeholder")).toBe(placeholder);
    });

    it("should disable input when disabled prop is true", () => {
      wrapper = createWrapper({ disabled: true });

      const input = wrapper.find("input");
      expect(input.attributes("disabled")).toBeDefined();
    });

    it("should show map button when showMapButton is true", () => {
      wrapper = createWrapper({ showMapButton: true });

      const mapButton = wrapper.find(".action-map");
      expect(mapButton.exists()).toBe(true);
    });

    it("should display existing coordinates in input field", async () => {
      wrapper = createWrapper({ latlng: [37.7749, -122.4194] });

      await nextTick();
      const input = wrapper.find("input");
      expect(input.element.value).toBe("37.7749, -122.4194");
    });
  });

  describe("User Input and Validation", () => {
    beforeEach(() => {
      wrapper = createWrapper();
    });

    it("should emit coordinates when valid input is entered and Enter is pressed", async () => {
      const input = wrapper.find("input");

      await input.setValue("37.7749, -122.4194");
      await input.trigger("keydown.enter");

      expect(wrapper.emitted("update:latlng")).toEqual([[[37.7749, -122.4194]]]);
      expect(wrapper.emitted("changed")).toEqual([[{ lat: 37.7749, lng: -122.4194 }]]);
    });

    it("should not emit coordinates for invalid input", async () => {
      const input = wrapper.find("input");

      await input.setValue("invalid coordinates");
      await input.trigger("keydown.enter");

      expect(wrapper.emitted("update:latlng")).toBeFalsy();
      expect(wrapper.emitted("changed")).toBeFalsy();
    });

    it("should handle various valid coordinate formats", async () => {
      const input = wrapper.find("input");

      // Test with spaces around comma
      await input.setValue("90, 180");
      await input.trigger("keydown.enter");

      expect(wrapper.emitted("update:latlng")[0]).toEqual([[90, 180]]);
    });
  });

  describe("Button Interactions", () => {
    it("should emit open-map event when map button is clicked", async () => {
      wrapper = createWrapper({ showMapButton: true });

      const mapButton = wrapper.find(".action-map");
      await mapButton.trigger("click");

      expect(wrapper.emitted("open-map")).toBeTruthy();
    });

    it("should clear coordinates when clear button is clicked", async () => {
      wrapper = createWrapper({ latlng: [37.7749, -122.4194] });

      // Wait for component to initialize and coordinateInput to be set
      await nextTick();

      const clearButton = wrapper.find(".action-clear");
      expect(clearButton.exists()).toBe(true);

      await clearButton.trigger("click");

      expect(wrapper.emitted("update:latlng")).toEqual([[[0, 0]]]);
      expect(wrapper.emitted("changed")).toEqual([[{ lat: 0, lng: 0 }]]);
      expect(wrapper.emitted("cleared")).toBeTruthy();
    });

    it("should show and work with undo button when enabled", async () => {
      wrapper = createWrapper({ enableUndo: true, latlng: [37.7749, -122.4194] });

      // Wait for component to initialize and coordinateInput to be set
      await nextTick();

      // Clear coordinates first
      const clearButton = wrapper.find(".action-clear");
      expect(clearButton.exists()).toBe(true);
      await clearButton.trigger("click");
      await nextTick();

      // Undo button should appear
      const undoButton = wrapper.find(".action-undo");
      expect(undoButton.exists()).toBe(true);

      // Click undo to restore coordinates
      await undoButton.trigger("click");

      const latlngEmits = wrapper.emitted("update:latlng");

      // Last emit should restore original coordinates
      expect(latlngEmits[latlngEmits.length - 1]).toEqual([[37.7749, -122.4194]]);
    });
  });

  describe("Auto Apply Feature", () => {
    it("should auto apply valid coordinates after debounce delay", async () => {
      wrapper = createWrapper({ autoApply: true, debounceDelay: 500 });

      const input = wrapper.find("input");
      await input.setValue("37.7749, -122.4194");

      // Should not emit immediately
      expect(wrapper.emitted("update:latlng")).toBeFalsy();

      // Fast forward timer
      vi.advanceTimersByTime(500);
      await nextTick();

      expect(wrapper.emitted("update:latlng")).toEqual([[[37.7749, -122.4194]]]);
    });

    it("should not auto apply when autoApply is disabled", async () => {
      wrapper = createWrapper({ autoApply: false });

      const input = wrapper.find("input");
      await input.setValue("37.7749, -122.4194");

      vi.advanceTimersByTime(1000);
      await nextTick();

      expect(wrapper.emitted("update:latlng")).toBeFalsy();
    });
  });

  describe("Paste Functionality", () => {
    beforeEach(() => {
      wrapper = createWrapper();
    });

    it("should handle paste with valid coordinates", async () => {
      const input = wrapper.find("input");

      const pasteEvent = new Event("paste");
      pasteEvent.clipboardData = {
        getData: vi.fn().mockReturnValue("40.7128, -74.0060"),
      };

      await input.trigger("paste", { clipboardData: pasteEvent.clipboardData });

      expect(wrapper.emitted("update:latlng")).toEqual([[[40.7128, -74.006]]]);
      expect(wrapper.emitted("changed")).toEqual([[{ lat: 40.7128, lng: -74.006 }]]);
    });

    it("should handle paste with space-separated coordinates", async () => {
      const input = wrapper.find("input");

      const pasteEvent = new Event("paste");
      pasteEvent.clipboardData = {
        getData: vi.fn().mockReturnValue("40.7128 -74.0060"),
      };

      await input.trigger("paste", { clipboardData: pasteEvent.clipboardData });

      expect(wrapper.emitted("update:latlng")).toEqual([[[40.7128, -74.006]]]);
    });
  });

  describe("Props Updates", () => {
    it("should update input field when lat/lng props change", async () => {
      wrapper = createWrapper();

      await wrapper.setProps({ latlng: [40.7128, -74.006] });

      const input = wrapper.find("input");
      expect(input.element.value).toBe("40.7128, -74.006");
    });

    it("should clear input field when coordinates are invalid", async () => {
      wrapper = createWrapper({ latlng: [0, 0] });

      await nextTick();
      const input = wrapper.find("input");
      expect(input.element.value).toBe("");
    });
  });
});
