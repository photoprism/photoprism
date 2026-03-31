import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount } from "@vue/test-utils";
import PSidebarInfo from "component/sidebar/info.vue";
import * as contexts from "options/contexts";
import { DateTime } from "luxon";

// Mock dependencies
vi.mock("component/map.vue", () => ({
  default: {
    name: "p-map",
    template: "<div class='p-map-stub'></div>",
    props: ["latlng", "animateDuration"],
  },
}));

vi.mock("options/formats", () => ({
  DATETIME_MED: "DATETIME_MED",
  DATETIME_MED_TZ: "DATETIME_MED_TZ",
}));

describe("PSidebarInfo component", () => {
  let wrapper;
  let originalFromISO;
  let mockModel;
  let mockPhoto;

  function createMocks() {
    mockModel = {
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

    mockPhoto = {
      getCameraInfo: vi.fn().mockReturnValue("Canon EOS R5"),
      getLensInfo: vi.fn().mockReturnValue("RF 50mm F1.2L"),
      getExifInfo: vi.fn().mockReturnValue("50mm \u2022 \u0192/1.2 \u2022 ISO 400 \u2022 1/125"),
      getMarkers: vi.fn().mockReturnValue([
        { UID: "m1", CropID: "crop1", Name: "Jane Doe", SubjUID: "subj1", thumbnailUrl: () => "/t/thumb1/public/tile_160" },
        { UID: "m2", CropID: "crop2", Name: "", SubjUID: "", thumbnailUrl: () => "/svg/portrait" },
      ]),
      Labels: [
        { Label: { UID: "lbl1", Name: "Nature", Slug: "nature", CustomSlug: "" } },
        { Label: { UID: "lbl2", Name: "Landscape", Slug: "landscape", CustomSlug: "custom-landscape" } },
      ],
      Albums: [
        { UID: "alb1", Title: "Vacation 2023", Slug: "vacation-2023" },
        { UID: "alb2", Title: "Favorites", Slug: "favorites" },
      ],
      Details: {
        Notes: "Some notes about this photo",
        Subject: "Mountains",
        Artist: "John Photographer",
        Copyright: "2023 John",
        License: "CC BY 4.0",
        Keywords: "nature, mountains, sunset",
      },
      PlaceLabel: "Berlin, Germany",
      FileName: "photos/2023/IMG_001.jpg",
      OriginalName: "IMG_001_original.jpg",
    };
  }

  beforeEach(() => {
    createMocks();

    originalFromISO = DateTime.fromISO;
    DateTime.fromISO = vi.fn().mockImplementation(() => ({
      toLocaleString: () => "January 1, 2023, 10:00 AM",
    }));

    wrapper = mount(PSidebarInfo, {
      props: {
        modelValue: mockModel,
        photo: mockPhoto,
        context: contexts.Photos,
      },
      global: {
        stubs: {
          PMap: true,
        },
      },
    });
  });

  afterEach(() => {
    DateTime.fromISO = originalFromISO;
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
    const onClose = vi.fn();
    const w = mount(PSidebarInfo, {
      props: { modelValue: mockModel, photo: mockPhoto, context: contexts.Photos, onClose },
      global: { stubs: { PMap: true } },
    });
    const allButtons = w.findAll("button");
    if (allButtons.length > 0) {
      await allButtons[0].trigger("click");
      expect(onClose).toHaveBeenCalled();
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
    const formattedTime = wrapper.vm.formatTime({ ...mockModel, TakenAtLocal: null });
    expect(formattedTime).toBe("Unknown");
  });

  // Camera, lens, and EXIF info
  it("should display camera info from photo prop", () => {
    expect(wrapper.vm.cameraInfo).toBe("Canon EOS R5");
  });

  it("should display lens info from photo prop", () => {
    expect(wrapper.vm.lensInfo).toBe("RF 50mm F1.2L");
  });

  it("should display EXIF info from photo prop", () => {
    expect(wrapper.vm.exifInfo).toBe("50mm \u2022 \u0192/1.2 \u2022 ISO 400 \u2022 1/125");
  });

  it("should hide camera info when value is Unknown", () => {
    const photo = { ...mockPhoto, getCameraInfo: vi.fn().mockReturnValue("Unknown"), getMarkers: vi.fn().mockReturnValue([]) };
    const w = mount(PSidebarInfo, {
      props: { modelValue: mockModel, photo, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(w.vm.cameraInfo).toBe("");
  });

  it("should return empty strings when photo prop is null", () => {
    const w = mount(PSidebarInfo, {
      props: { modelValue: mockModel, photo: null, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(w.vm.cameraInfo).toBe("");
    expect(w.vm.lensInfo).toBe("");
    expect(w.vm.exifInfo).toBe("");
    expect(w.vm.people).toEqual([]);
    expect(w.vm.labels).toEqual([]);
    expect(w.vm.albums).toEqual([]);
    expect(w.vm.placeName).toBe("");
    expect(w.vm.fileName).toBe("");
    expect(w.vm.originalName).toBe("");
    expect(w.vm.subject).toBe("");
    expect(w.vm.artist).toBe("");
    expect(w.vm.copyright).toBe("");
    expect(w.vm.license).toBe("");
    expect(w.vm.keywords).toBe("");
    expect(w.vm.notesHtml).toBe("");
  });

  // People
  it("should return all markers including unnamed", () => {
    expect(wrapper.vm.people).toHaveLength(2);
    expect(wrapper.vm.people[0].Name).toBe("Jane Doe");
    expect(wrapper.vm.people[1].Name).toBe("");
  });

  it("should render person rows with avatars", () => {
    const personRows = wrapper.findAll(".metadata__person-row");
    expect(personRows.length).toBe(2);

    const avatars = wrapper.findAll(".meta-person__avatar");
    expect(avatars.length).toBe(2);
  });

  it("should make named people clickable", () => {
    const personRows = wrapper.findAll(".metadata__person-row");
    expect(personRows[0].classes()).toContain("clickable");
    expect(personRows[1].classes()).not.toContain("clickable");
  });

  // Labels
  it("should return labels from photo prop", () => {
    expect(wrapper.vm.labels).toHaveLength(2);
    expect(wrapper.vm.labels[0].Label.Name).toBe("Nature");
  });

  it("should render label chips", () => {
    const html = wrapper.html();
    expect(html).toContain("Nature");
    expect(html).toContain("Landscape");
  });

  // Albums
  it("should return albums from photo prop", () => {
    expect(wrapper.vm.albums).toHaveLength(2);
    expect(wrapper.vm.albums[0].Title).toBe("Vacation 2023");
  });

  it("should render album chips", () => {
    const html = wrapper.html();
    expect(html).toContain("Vacation 2023");
    expect(html).toContain("Favorites");
  });

  // Metadata details
  it("should return metadata details from photo prop", () => {
    expect(wrapper.vm.subject).toBe("Mountains");
    expect(wrapper.vm.artist).toBe("John Photographer");
    expect(wrapper.vm.copyright).toBe("2023 John");
    expect(wrapper.vm.license).toBe("CC BY 4.0");
    expect(wrapper.vm.keywords).toBe("nature, mountains, sunset");
  });

  it("should return place and file info from photo prop", () => {
    expect(wrapper.vm.placeName).toBe("Berlin, Germany");
    expect(wrapper.vm.fileName).toBe("photos/2023/IMG_001.jpg");
    expect(wrapper.vm.originalName).toBe("IMG_001_original.jpg");
  });

  // Caption and notes HTML
  it("should produce caption and notes HTML via sanitize pipeline", () => {
    expect(wrapper.vm.captionHtml).toBe("Test Caption");
    expect(wrapper.vm.notesHtml).toBe("Some notes about this photo");
  });

  it("should return empty caption HTML when no caption", () => {
    const w = mount(PSidebarInfo, {
      props: { modelValue: { ...mockModel, Caption: "" }, photo: mockPhoto, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(w.vm.captionHtml).toBe("");
  });

  // Navigation events
  it("should emit navigate event for label click with slug", () => {
    const onNavigate = vi.fn();
    const w = mount(PSidebarInfo, {
      props: { modelValue: mockModel, photo: mockPhoto, context: contexts.Photos, onNavigate },
      global: { stubs: { PMap: true } },
    });
    w.vm.navigateToLabel({ UID: "lbl1", Name: "Nature", Slug: "nature", CustomSlug: "" });
    expect(onNavigate).toHaveBeenCalledWith({ name: "browse", query: { q: "label:nature" } });
  });

  it("should prefer CustomSlug for label navigation", () => {
    const onNavigate = vi.fn();
    const w = mount(PSidebarInfo, {
      props: { modelValue: mockModel, photo: mockPhoto, context: contexts.Photos, onNavigate },
      global: { stubs: { PMap: true } },
    });
    w.vm.navigateToLabel({ UID: "lbl2", Name: "Landscape", Slug: "landscape", CustomSlug: "custom-landscape" });
    expect(onNavigate).toHaveBeenCalledWith({ name: "browse", query: { q: "label:custom-landscape" } });
  });

  it("should emit navigate event for album click", () => {
    const onNavigate = vi.fn();
    const w = mount(PSidebarInfo, {
      props: { modelValue: mockModel, photo: mockPhoto, context: contexts.Photos, onNavigate },
      global: { stubs: { PMap: true } },
    });
    w.vm.navigateToAlbum({ UID: "alb1", Title: "Vacation 2023" });
    expect(onNavigate).toHaveBeenCalledWith({ name: "album", params: { album: "alb1", slug: "view" } });
  });

  it("should emit navigate with subject filter for person with SubjUID", () => {
    const onNavigate = vi.fn();
    const w = mount(PSidebarInfo, {
      props: { modelValue: mockModel, photo: mockPhoto, context: contexts.Photos, onNavigate },
      global: { stubs: { PMap: true } },
    });
    w.vm.navigateToPerson({ UID: "m1", Name: "Jane Doe", SubjUID: "subj1" });
    expect(onNavigate).toHaveBeenCalledWith({ name: "browse", query: { q: "subject:subj1" } });
  });

  it("should emit navigate with person filter when only Name available", () => {
    const onNavigate = vi.fn();
    const w = mount(PSidebarInfo, {
      props: { modelValue: mockModel, photo: mockPhoto, context: contexts.Photos, onNavigate },
      global: { stubs: { PMap: true } },
    });
    w.vm.navigateToPerson({ UID: "m3", Name: "Unknown Person", SubjUID: "" });
    expect(onNavigate).toHaveBeenCalledWith({ name: "browse", query: { q: "person:Unknown Person" } });
  });

  it("should not emit navigate for person without name or SubjUID", () => {
    const onNavigate = vi.fn();
    const w = mount(PSidebarInfo, {
      props: { modelValue: mockModel, photo: mockPhoto, context: contexts.Photos, onNavigate },
      global: { stubs: { PMap: true } },
    });
    w.vm.navigateToPerson({ UID: "m4", Name: "", SubjUID: "" });
    expect(onNavigate).not.toHaveBeenCalled();
  });
});
