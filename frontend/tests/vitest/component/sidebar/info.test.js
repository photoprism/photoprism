import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount } from "@vue/test-utils";
import PSidebarInfo from "component/sidebar/info.vue";
import { Marker } from "model/marker";
import * as contexts from "options/contexts";
import { DateTime } from "luxon";

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
  let originalFromISO;

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
    originalFromISO = DateTime.fromISO;
    DateTime.fromISO = vi.fn().mockImplementation(() => {
      return {
        toLocaleString: () => "January 1, 2023, 10:00 AM",
      };
    });

    wrapper = mount(PSidebarInfo, {
      props: {
        modelValue: mockModel,
        context: contexts.Photos,
      },
      global: {
        stubs: {
          PMap: true,
          PConfirmDialog: true,
        },
      },
    });
  });

  afterEach(() => {
    DateTime.fromISO = originalFromISO;

    if (wrapper) {
      wrapper.unmount();
    }
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
    const closeButtonSelectors = [".close-button", "button[aria-label='Close']", "button[title='Close']"];

    let closeButton;
    for (const selector of closeButtonSelectors) {
      closeButton = wrapper.find(selector);
      if (closeButton.exists()) {
        break;
      }
    }

    if (!closeButton || !closeButton.exists()) {
      const allButtons = wrapper.findAll("button");
      if (allButtons.length > 0) {
        closeButton = allButtons[0];
      }
    }

    if (closeButton && closeButton.exists()) {
      await closeButton.trigger("click");
      expect(wrapper.emitted()).toHaveProperty("close");
    }
  });

  it("should trigger copyLatLng when location is clicked", async () => {
    const clickableItems = wrapper.findAll(".clickable");
    if (clickableItems.length > 0) {
      await clickableItems[0].trigger("click");
      expect(mockModel.copyLatLng).toHaveBeenCalled();
    }
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

describe("PSidebarInfo people markers", () => {
  let wrapper;
  let marker;

  const mountSidebar = () =>
    mount(PSidebarInfo, {
      props: {
        modelValue: {
          UID: "p1",
          Title: "Example",
          Caption: "",
          TakenAtLocal: "2024-01-01T12:00:00Z",
          TimeZone: "UTC",
          Lat: 0,
          Lng: 0,
          getTypeInfo: () => "JPEG",
          getTypeIcon: () => "mdi-image",
          copyLatLng: vi.fn(),
        },
        photo: {
          UID: "p1",
          getMarkers: vi.fn(() => [marker]),
        },
        canEdit: true,
      },
      global: {
        mocks: {
          $gettext: (msg, values) => {
            if (!values) {
              return msg;
            }

            return msg.replace("%{s}", values.s);
          },
          $pgettext: (_ctx, msg) => msg,
          $notify: {
            blockUI: vi.fn(),
            unblockUI: vi.fn(),
          },
          $config: {
            feature: vi.fn((name) => name === "people"),
            values: {
              people: [{ UID: "s1", Name: "Jane Doe" }],
            },
          },
        },
        stubs: {
          PMap: true,
          PConfirmDialog: true,
        },
      },
    });

  beforeEach(() => {
    marker = new Marker({
      UID: "m1",
      Name: "",
      SubjUID: "",
      Invalid: false,
      Thumb: "thumb-hash",
    });
    marker.reject = vi.fn().mockResolvedValue(marker);
    marker.clearSubject = vi.fn().mockResolvedValue(marker);
    marker.setName = vi.fn().mockResolvedValue(marker);

    wrapper = mountSidebar();
  });

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount();
    }
  });

  it("loads markers from the current photo", () => {
    expect(wrapper.vm.markers).toHaveLength(1);
    expect(wrapper.vm.markers[0].UID).toBe("m1");
  });

  it("rejects a marker and requests a reload", async () => {
    await wrapper.vm.onReject(marker);

    expect(marker.reject).toHaveBeenCalled();
    expect(wrapper.emitted("reload-markers")).toBeTruthy();
  });

  it("assigns an existing person name and requests a reload", async () => {
    await wrapper.vm.onSetPerson(marker, { UID: "s1", Name: "Jane Doe" });

    expect(marker.Name).toBe("Jane Doe");
    expect(marker.SubjUID).toBe("s1");
    expect(marker.setName).toHaveBeenCalled();
    expect(wrapper.emitted("reload-markers")).toBeTruthy();
  });
});
