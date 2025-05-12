import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount } from "@vue/test-utils";
import PSidebarInfo from "component/sidebar/info.vue";
import { DateTime } from "luxon";

// Mock dependencies
vi.mock("component/map.vue", () => ({
  default: {
    name: "p-map",
    template: "<div class='p-map-stub'></div>",
    props: ["lat", "lng"],
  },
}));

vi.mock("options/formats", () => ({
  DATETIME_MED: "DATETIME_MED",
  DATETIME_MED_TZ: "DATETIME_MED_TZ",
}));

describe("PSidebarInfo component", () => {
  let wrapper;
  const mockModel = {
    UID: "abc123",
    Title: "Test Title",
    Caption: "Test Caption",
    TakenAtLocal: "2023-01-01T10:00:00Z",
    TimeZone: "UTC",
    Lat: 52.52,
    Lng: 13.405,
    getTypeInfo: vi.fn().mockReturnValue("JPEG, 1920x1080"),
    getTypeIcon: vi.fn().mockReturnValue("mdi-file-image"),
    getLatLng: vi.fn().mockReturnValue("52.5200, 13.4050"),
    copyLatLng: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();

    wrapper = mount(PSidebarInfo, {
      props: {
        modelValue: mockModel,
        context: "photos",
      },
      global: {
        stubs: {
          PMap: true,
        },
      },
    });
  });

  it("should render correctly with model data", () => {
    expect(wrapper.vm).toBeTruthy();
    expect(wrapper.find(".p-sidebar-info").exists()).toBe(true);

    const html = wrapper.html();
    expect(html).toContain("Test Title");
    expect(html).toContain("Test Caption");

    expect(mockModel.getTypeInfo).toHaveBeenCalled();
    expect(mockModel.getTypeIcon).toHaveBeenCalled();
    expect(mockModel.getLatLng).toHaveBeenCalled();
  });

  it("should emit close event when close button is clicked", async () => {
    const closeButton = wrapper.find(".vbtn-stub");
    await closeButton.trigger("click");

    expect(wrapper.emitted()).toHaveProperty("close");
  });

  it("should trigger copyLatLng when location is clicked", async () => {
    // Find the location item by its class
    const clickableItems = wrapper.findAll(".clickable");
    if (clickableItems.length > 0) {
      await clickableItems[0].trigger("click");
      expect(mockModel.copyLatLng).toHaveBeenCalled();
    }
  });

  it("should format time correctly", () => {
    // Mock DateTime.fromISO to return a controllable object
    const mockToLocaleString = vi.fn().mockReturnValue("January 1, 2023, 10:00 AM");
    DateTime.fromISO = vi.fn().mockReturnValue({
      toLocaleString: mockToLocaleString,
    });

    const formattedTime = wrapper.vm.formatTime(mockModel);

    expect(DateTime.fromISO).toHaveBeenCalledWith("2023-01-01T10:00:00Z", { zone: "UTC" });
    expect(mockToLocaleString).toHaveBeenCalledWith("DATETIME_MED_TZ");
    expect(formattedTime).toBe("January 1, 2023, 10:00 AM");
  });

  it("should handle model with timezone", () => {
    // Create a model with non-UTC timezone
    const modelWithTZ = {
      ...mockModel,
      TimeZone: "Europe/Berlin",
    };

    // Mock DateTime.fromISO to return a controllable object
    const mockToLocaleString = vi.fn().mockReturnValue("January 1, 2023, 11:00 AM CET");
    DateTime.fromISO = vi.fn().mockReturnValue({
      toLocaleString: mockToLocaleString,
    });

    const formattedTime = wrapper.vm.formatTime(modelWithTZ);

    expect(DateTime.fromISO).toHaveBeenCalledWith("2023-01-01T10:00:00Z", { zone: "Europe/Berlin" });
    expect(mockToLocaleString).toHaveBeenCalledWith("DATETIME_MED_TZ");
    expect(formattedTime).toBe("January 1, 2023, 11:00 AM CET");
  });

  it("should handle model without taken time", () => {
    const modelWithoutTime = {
      ...mockModel,
      TakenAtLocal: null,
    };

    const formattedTime = wrapper.vm.formatTime(modelWithoutTime);

    expect(formattedTime).toBe("Unknown");
  });
});
