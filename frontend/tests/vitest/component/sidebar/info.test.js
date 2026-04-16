import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, mount } from "@vue/test-utils";
import { DateTime } from "luxon";
import PSidebarInfo from "component/sidebar/info.vue";
import { Photo } from "model/photo";
import * as contexts from "options/contexts";

vi.mock("component/map.vue", () => ({
  default: {
    name: "p-map",
    template: "<div class='p-map-stub'></div>",
    props: ["lat", "lng"],
  },
}));

vi.mock("options/formats", () => ({
  DATE_MED: "DATE_MED",
  DATETIME_MED: "DATETIME_MED",
  DATETIME_MED_TZ: "DATETIME_MED_TZ",
}));

describe("PSidebarInfo component", () => {
  let originalFromISO;
  let findSpy;
  let copyLatLng;
  let wrapper;

  const baseModel = {
    UID: "abc123",
    Title: "Test Title",
    Caption: "Test Caption",
    TakenAt: "2023-01-01T10:00:00Z",
    TakenAtLocal: "2023-01-01T10:00:00Z",
    TimeZone: "UTC",
    Year: 2023,
    Month: 1,
    Day: 1,
    Lat: 52.52,
    Lng: 13.405,
    FileName: "test-title.jpg",
    Type: "image",
    Width: 1920,
    Height: 1080,
    getLatLng: vi.fn().mockReturnValue("52.5200, 13.4050"),
    copyLatLng: vi.fn(),
  };

  async function mountSidebar(modelValue = baseModel, details = baseModel) {
    copyLatLng = vi.fn();

    findSpy = vi.spyOn(Photo.prototype, "find").mockResolvedValue(
      new Photo({
        ...details,
        copyLatLng,
        getLatLng: vi.fn().mockReturnValue("52.5200, 13.4050"),
      })
    );

    wrapper = mount(PSidebarInfo, {
      props: {
        modelValue,
        context: contexts.Photos,
      },
      global: {
        stubs: {
          PMap: true,
        },
      },
    });

    await flushPromises();

    return wrapper;
  }

  beforeEach(async () => {
    vi.clearAllMocks();
    originalFromISO = DateTime.fromISO;
    DateTime.fromISO = vi.fn().mockImplementation(() => ({
      toLocaleString: () => "January 1, 2023",
    }));

    await mountSidebar();
  });

  afterEach(() => {
    DateTime.fromISO = originalFromISO;
    findSpy?.mockRestore();
    wrapper?.unmount();
  });

  it("renders the configured metadata fields from fetched details", () => {
    expect(wrapper.vm).toBeTruthy();
    expect(wrapper.find(".p-sidebar-info").exists()).toBe(true);
    expect(findSpy).toHaveBeenCalledWith("abc123");
    expect(wrapper.vm.metadataItems.some((item) => item.key.startsWith("caption-") && item.text === "Test Caption")).toBe(true);
    expect(wrapper.vm.metadataItems.some((item) => item.key.startsWith("filename-") && item.text.includes("test-title"))).toBe(true);
    expect(wrapper.html()).toContain("52.5200, 13.4050");
  });

  it("emits close when the toolbar button is clicked", async () => {
    const closeButton = wrapper.find("button[title='Close']");

    expect(closeButton.exists()).toBe(true);

    await closeButton.trigger("click");

    expect(wrapper.emitted()).toHaveProperty("close");
  });

  it("copies the location when the location row is clicked", async () => {
    const clickableItems = wrapper.findAll(".clickable");

    expect(clickableItems.length).toBeGreaterThan(0);

    await clickableItems[0].trigger("click");

    expect(copyLatLng).toHaveBeenCalled();
  });

  it("hides missing dates", async () => {
    wrapper.unmount();
    findSpy.mockRestore();

    await mountSidebar(
      { ...baseModel, UID: "missing-time", TakenAt: "", TakenAtLocal: "", Year: -1, Month: -1, Day: -1, Lat: 0, Lng: 0 },
      { ...baseModel, UID: "missing-time", TakenAt: "", TakenAtLocal: "", Year: -1, Month: -1, Day: -1, Lat: 0, Lng: 0 }
    );

    const dateItem = wrapper.vm.metadataItems.find((item) => item.key.startsWith("date-"));

    expect(dateItem).toBeUndefined();
  });
});
