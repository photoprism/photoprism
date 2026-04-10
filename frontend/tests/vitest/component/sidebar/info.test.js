import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount } from "@vue/test-utils";
import PSidebarInfo from "component/sidebar/info.vue";
import * as contexts from "options/contexts";
import { DateTime } from "luxon";
import $util from "common/util";
import { Album } from "model/album";

// Max name length used by the validation pipeline (matches the production
// "clip" client-config value). Override the global $config.get mock so the
// length-check branch can be exercised in tests.
const CLIP_LEN = 160;
const validationConfig = {
  feature: () => true,
  get: (key) => (key === "clip" ? CLIP_LEN : false),
  getSettings: () => ({ features: { edit: true, favorites: true, download: true, archive: true } }),
  allow: () => true,
  featExperimental: () => false,
  featDevelop: () => false,
  values: {},
  dir: () => "ltr",
};
// Mounted with the real $util.normalizeLabelTitle so the validation
// pipeline runs against the same normalization the component uses at
// runtime. Other $util methods needed at render time are stubbed inline.
const validationUtil = {
  normalizeLabelTitle: (s) => $util.normalizeLabelTitle(s),
  formatCamera: (camera, id, make, model, long) => $util.formatCamera(camera, id, make, model, long),
  encodeHTML: (s) => s,
  sanitizeHtml: (s) => s,
  hasTouch: () => false,
  formatSeconds: (n) => String(n),
  formatRemainingSeconds: () => "0",
  videoFormat: () => "avc",
  videoFormatUrl: () => "/v.mp4",
  thumb: () => ({ src: "/t.jpg", w: 100, h: 100 }),
};
function mountInfoForChips(props) {
  return mount(PSidebarInfo, {
    props: { canEdit: true, context: contexts.Photos, ...props },
    global: {
      stubs: { PMap: true },
      mocks: {
        $config: validationConfig,
        $util: validationUtil,
      },
    },
  });
}

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
      getLatLng: vi.fn().mockReturnValue("52.5200, 13.4050"),
      copyLatLng: vi.fn(),
    };

    mockPhoto = {
      Type: "image",
      CameraID: 2,
      CameraMake: "Canon",
      CameraModel: "EOS R5",
      LensID: 2,
      LensMake: "Canon",
      LensModel: "RF 50mm F1.2L",
      getCameraInfo: vi.fn().mockReturnValue("Canon EOS R5"),
      getLensInfo: vi.fn().mockReturnValue("RF 50mm F1.2L"),
      getImageInfo: vi.fn().mockReturnValue("JPEG, 1920 × 1080, 4.2 MB"),
      getVideoInfo: vi.fn().mockReturnValue(""),
      getVectorInfo: vi.fn().mockReturnValue(""),
      getExifInfo: vi.fn().mockReturnValue("50mm \u2022 \u0192/1.2 \u2022 ISO 400 \u2022 1/125"),
      locationInfo: vi.fn().mockReturnValue("Berlin, Germany"),
      getMarkers: vi.fn().mockReturnValue([
        { UID: "m1", CropID: "crop1", Name: "Jane Doe", SubjUID: "subj1", thumbnailUrl: () => "/t/thumb1/public/tile_160" },
        { UID: "m2", CropID: "crop2", Name: "", SubjUID: "", thumbnailUrl: () => "/svg/portrait" },
      ]),
      Labels: [
        { Uncertainty: 0, Label: { ID: 1, UID: "lbl1", Name: "Nature", Slug: "nature", CustomSlug: "" } },
        { Uncertainty: 0, Label: { ID: 2, UID: "lbl2", Name: "Landscape", Slug: "landscape", CustomSlug: "custom-landscape" } },
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
    expect(html).toContain("photos/2023/IMG_001.jpg");
    expect(html).toContain("JPEG, 1920 × 1080, 4.2 MB");

    expect(mockPhoto.getImageInfo).toHaveBeenCalled();
    expect(mockModel.getLatLng).toHaveBeenCalled();
  });

  it("should not render an icon or pencil next to the filename", () => {
    const fileRow = wrapper.find(".metadata__file");
    expect(fileRow.exists()).toBe(true);
    expect(fileRow.find(".meta-inline-pencil").exists()).toBe(false);
    const filename = fileRow.find(".meta-filename");
    expect(filename.exists()).toBe(true);
    expect(filename.find(".v-icon").exists()).toBe(false);
  });

  it("should render file info row with a prepend icon like Taken/Camera", () => {
    const html = wrapper.html();
    expect(html).toContain("JPEG, 1920 × 1080, 4.2 MB");
    expect(html).toContain("mdi-image-outline");
  });

  it("should emit close event when close button is clicked", async () => {
    const onClose = vi.fn();
    const w = mount(PSidebarInfo, {
      props: { modelValue: mockModel, photo: mockPhoto, context: contexts.Photos, onClose },
      global: { stubs: { PMap: true } },
    });
    const closeButton = w.findAll("button")[0];
    await closeButton.trigger("click");
    expect(onClose).toHaveBeenCalled();
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

  it("should hide camera row in read-only mode when only ISO/exposure are set", () => {
    const photo = {
      ...mockPhoto,
      CameraID: 1,
      CameraMake: "",
      CameraModel: "",
      Iso: 100,
      Exposure: "1/125",
      getCameraInfo: vi.fn().mockReturnValue("Unknown, ISO 100, 1/125"),
      getMarkers: vi.fn().mockReturnValue([]),
    };
    const w = mount(PSidebarInfo, {
      props: { modelValue: mockModel, photo, canEdit: false, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(w.vm.cameraInfo).toBe("");
    expect(w.html()).not.toContain("mdi-camera ");
  });

  it("should hide lens row when only FNumber/FocalLength are set without a real lens", () => {
    const photo = {
      ...mockPhoto,
      LensID: 1,
      LensMake: "",
      LensModel: "",
      Lens: null,
      FNumber: 1.8,
      FocalLength: 50,
      getLensInfo: vi.fn().mockReturnValue("50mm, ƒ/1.8"),
      getMarkers: vi.fn().mockReturnValue([]),
    };
    const w = mount(PSidebarInfo, {
      props: { modelValue: mockModel, photo, canEdit: false, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(w.vm.lensInfo).toBe("");
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
    expect(w.vm.fileInfo).toBe("");
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

  // Face marker editing (approval required for every change)
  it("should not hit the server on eject until the user confirms", async () => {
    const marker = {
      UID: "mE",
      CropID: "cropE",
      Name: "Alice",
      SubjUID: "subjA",
      clearSubject: vi.fn().mockResolvedValue(true),
      setName: vi.fn().mockResolvedValue(true),
    };
    wrapper.vm.onEjectPerson(marker);
    expect(marker.clearSubject).not.toHaveBeenCalled();
    expect(marker.Name).toBe("");
    expect(marker.SubjUID).toBe("");
    expect(wrapper.vm.isEditingPerson(marker)).toBe(true);

    wrapper.vm.confirmField();
    expect(marker.clearSubject).toHaveBeenCalledTimes(1);
    expect(marker.setName).not.toHaveBeenCalled();
  });

  it("should restore the marker on cancel after eject", () => {
    const marker = {
      UID: "mF",
      CropID: "cropF",
      Name: "Bob",
      SubjUID: "subjB",
      clearSubject: vi.fn(),
      setName: vi.fn(),
    };
    wrapper.vm.onEjectPerson(marker);
    wrapper.vm._editStartedAt = 0;
    wrapper.vm.cancelEditing();
    expect(marker.Name).toBe("Bob");
    expect(marker.SubjUID).toBe("subjB");
    expect(marker.clearSubject).not.toHaveBeenCalled();
    expect(marker.setName).not.toHaveBeenCalled();
  });

  it("should require confirmation when selecting an existing person from the dropdown", () => {
    const marker = {
      UID: "mG",
      CropID: "cropG",
      Name: "",
      SubjUID: "",
      clearSubject: vi.fn(),
      setName: vi.fn().mockResolvedValue(true),
    };
    wrapper.vm.startEditingPerson(marker);
    wrapper.vm.onSelectPerson(marker, { Name: "Carol", UID: "subjC" });
    expect(marker.Name).toBe("Carol");
    expect(marker.SubjUID).toBe("subjC");
    expect(marker.setName).not.toHaveBeenCalled();
    expect(wrapper.vm.isEditingPerson(marker)).toBe(true);

    wrapper.vm.confirmField();
    expect(marker.setName).toHaveBeenCalledTimes(1);
  });

  it("should not call setName when confirming an unchanged marker", () => {
    const marker = {
      UID: "mH",
      CropID: "cropH",
      Name: "Dave",
      SubjUID: "subjD",
      clearSubject: vi.fn(),
      setName: vi.fn(),
    };
    wrapper.vm.startEditingPerson(marker);
    wrapper.vm.confirmField();
    expect(marker.setName).not.toHaveBeenCalled();
    expect(marker.clearSubject).not.toHaveBeenCalled();
  });

  // Labels
  it("should return labels from photo prop", () => {
    expect(wrapper.vm.labels).toHaveLength(2);
    expect(wrapper.vm.labels[0].Label.Name).toBe("Nature");
  });

  // Albums
  it("should return albums from photo prop", () => {
    expect(wrapper.vm.albums).toHaveLength(2);
    expect(wrapper.vm.albums[0].Title).toBe("Vacation 2023");
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
    expect(wrapper.vm.fileInfo).toBe("JPEG, 1920 × 1080, 4.2 MB");
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

  // isEditable
  it("should not be editable without canEdit prop", () => {
    expect(wrapper.vm.isEditable).toBeFalsy();
  });

  it("should be editable when canEdit is true with valid photo", () => {
    const w = mount(PSidebarInfo, {
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(w.vm.isEditable).toBeTruthy();
  });

  // Altitude
  it("should return altitude when photo has Altitude", () => {
    const photo = { ...mockPhoto, Altitude: 340 };
    const w = mount(PSidebarInfo, {
      props: { modelValue: mockModel, photo, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(w.vm.altitude).toBe("340 m");
  });

  // Labels Uncertainty filter
  it("should hide labels with Uncertainty 100", () => {
    const photo = {
      ...mockPhoto,
      Labels: [
        { Uncertainty: 0, Label: { ID: 1, UID: "lbl1", Name: "Nature", Slug: "nature", CustomSlug: "" } },
        { Uncertainty: 100, Label: { ID: 3, UID: "lbl3", Name: "Hidden", Slug: "hidden", CustomSlug: "" } },
      ],
    };
    const w = mount(PSidebarInfo, {
      props: { modelValue: mockModel, photo, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(w.vm.labels).toHaveLength(1);
    expect(w.vm.labels[0].Label.Name).toBe("Nature");
  });

  // Inline editing: startEditing / cancelEditing
  it("should set editingField and store original value on startEditing", () => {
    const photo = { ...mockPhoto, Title: "Test Title", Caption: "Test Caption" };
    const w = mount(PSidebarInfo, {
      props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    w.vm.startEditing("title");
    expect(w.vm.editingField).toBe("title");
    expect(w.vm.editOriginal).toBe("Test Title");
  });

  it("should restore original value on cancelEditing", async () => {
    const photo = { ...mockPhoto, Title: "Test Title", wasChanged: vi.fn().mockReturnValue(false) };
    const w = mount(PSidebarInfo, {
      props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    w.vm.startEditing("title");
    photo.Title = "Modified";
    // Wait past the 200ms blur guard
    w.vm._editStartedAt = Date.now() - 300;
    w.vm.cancelEditing();
    expect(photo.Title).toBe("Test Title");
    expect(w.vm.editingField).toBeNull();
  });

  // getFieldValue / setFieldValue
  it("should get and set field values for all fields", () => {
    const photo = { ...mockPhoto, Title: "Test Title", Caption: "Test Caption" };
    const w = mount(PSidebarInfo, {
      props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(w.vm.getFieldValue("title")).toBe("Test Title");
    expect(w.vm.getFieldValue("caption")).toBe("Test Caption");
    expect(w.vm.getFieldValue("subject")).toBe("Mountains");
    expect(w.vm.getFieldValue("notes")).toBe("Some notes about this photo");
    expect(w.vm.getFieldValue("unknown")).toBe("");

    w.vm.setFieldValue("title", "New Title");
    expect(photo.Title).toBe("New Title");
  });

  // Pending label operations
  it("should toggle label pending removal", () => {
    const w = mount(PSidebarInfo, {
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    const label = { Label: { ID: 1, UID: "lbl1", Name: "Nature" } };

    expect(w.vm.isLabelPendingRemoval(label)).toBe(false);
    w.vm.toggleLabelRemoval(label);
    expect(w.vm.isLabelPendingRemoval(label)).toBe(true);
    w.vm.toggleLabelRemoval(label);
    expect(w.vm.isLabelPendingRemoval(label)).toBe(false);
  });

  it("should add and remove pending label additions", () => {
    const w = mount(PSidebarInfo, {
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    w.vm.pendingLabelAdditions.push("Sunset");
    expect(w.vm.pendingLabelAdditions).toContain("Sunset");

    w.vm.removePendingLabelAdd("Sunset");
    expect(w.vm.pendingLabelAdditions).not.toContain("Sunset");
  });

  it("should ignore duplicate pending label additions via onLabelSelected", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.editingField = "labels";
    w.vm.onLabelSelected({ Name: "Sunset", UID: "lbl-new" });
    w.vm.onLabelSelected({ Name: "Sunset", UID: "lbl-new" });
    expect(w.vm.pendingLabelAdditions).toHaveLength(1);
  });

  it("should ignore non-object values in onLabelSelected", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.onLabelSelected("string-value");
    w.vm.onLabelSelected(null);
    expect(w.vm.pendingLabelAdditions).toHaveLength(0);
  });

  it("should skip labels already on the photo in onLabelSelected", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.editingField = "labels";
    w.vm.onLabelSelected({ Name: "Nature", UID: "lbl1" });
    expect(w.vm.pendingLabelAdditions).toHaveLength(0);
  });

  // Label validation parity with batch edit + labels tab.
  it("should dedupe pending label additions case-insensitively in onLabelSelected", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.editingField = "labels";
    w.vm.onLabelSelected({ Name: "cat" });
    w.vm.onLabelSelected({ Name: "CAT" });
    expect(w.vm.pendingLabelAdditions).toEqual(["cat"]);
  });

  it("should skip labels already on the photo case-insensitively in onLabelSelected", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.editingField = "labels";
    w.vm.onLabelSelected({ Name: "nature" });
    expect(w.vm.pendingLabelAdditions).toHaveLength(0);
  });

  it("should dedupe pending label additions case-insensitively in onLabelEnter", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.editingField = "labels";
    w.vm.pendingLabelAdditions.push("cat");
    w.vm.chipSearch = "CAT";
    w.vm.onLabelEnter();
    expect(w.vm.pendingLabelAdditions).toEqual(["cat"]);
  });

  it("should trim whitespace in onLabelEnter", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.editingField = "labels";
    w.vm.chipSearch = "  dog  ";
    w.vm.onLabelEnter();
    expect(w.vm.pendingLabelAdditions).toEqual(["dog"]);
  });

  it("should silently reject empty or whitespace-only label input in onLabelEnter", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.editingField = "labels";
    w.vm.chipSearch = "   ";
    w.vm.onLabelEnter();
    expect(w.vm.pendingLabelAdditions).toHaveLength(0);
    expect(w.vm.$notify.error).not.toHaveBeenCalled();
  });

  it("should reject labels longer than the configured clip length and notify", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.editingField = "labels";
    w.vm.chipSearch = "a".repeat(CLIP_LEN + 10);
    w.vm.onLabelEnter();
    expect(w.vm.pendingLabelAdditions).toHaveLength(0);
    expect(w.vm.$notify.error).toHaveBeenCalledWith("Name too long");
  });

  it("should match existing labels through normalization (punctuation stripped)", () => {
    const photo = {
      ...mockPhoto,
      Labels: [{ Uncertainty: 0, Label: { ID: 99, UID: "lbl99", Name: "Cat!", Slug: "cat", CustomSlug: "" } }],
    };
    const w = mountInfoForChips({ modelValue: mockModel, photo });
    w.vm.editingField = "labels";
    w.vm.chipSearch = "cat";
    w.vm.onLabelEnter();
    expect(w.vm.pendingLabelAdditions).toHaveLength(0);
  });

  it("should match existing labels through normalization (& vs and)", () => {
    const photo = {
      ...mockPhoto,
      Labels: [{ Uncertainty: 0, Label: { ID: 99, UID: "lbl99", Name: "Rock & Roll", Slug: "rock-and-roll", CustomSlug: "" } }],
    };
    const w = mountInfoForChips({ modelValue: mockModel, photo });
    w.vm.editingField = "labels";
    w.vm.chipSearch = "rock and roll";
    w.vm.onLabelEnter();
    expect(w.vm.pendingLabelAdditions).toHaveLength(0);
  });

  it("should silently reject punctuation-only label input", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.editingField = "labels";
    w.vm.chipSearch = "!!!";
    w.vm.onLabelEnter();
    expect(w.vm.pendingLabelAdditions).toHaveLength(0);
    expect(w.vm.$notify.error).not.toHaveBeenCalled();
  });

  // Pending album operations
  it("should toggle album pending removal", () => {
    const w = mount(PSidebarInfo, {
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    const album = { UID: "alb1", Title: "Vacation 2023" };

    expect(w.vm.isAlbumPendingRemoval(album)).toBe(false);
    w.vm.toggleAlbumRemoval(album);
    expect(w.vm.isAlbumPendingRemoval(album)).toBe(true);
    w.vm.toggleAlbumRemoval(album);
    expect(w.vm.isAlbumPendingRemoval(album)).toBe(false);
  });

  it("should add and remove pending album additions", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    const album = { UID: "alb-new", Title: "New Album" };

    w.vm.onAlbumSelected(album);
    expect(w.vm.pendingAlbumAdditions).toHaveLength(1);
    expect(w.vm.pendingAlbumAdditions[0].UID).toBe("alb-new");

    w.vm.removePendingAlbumAdd(album);
    expect(w.vm.pendingAlbumAdditions).toHaveLength(0);
  });

  it("should ignore non-object values in onAlbumSelected", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.onAlbumSelected("string-value");
    w.vm.onAlbumSelected(null);
    expect(w.vm.pendingAlbumAdditions).toHaveLength(0);
  });

  it("should skip albums already on the photo in onAlbumSelected", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.onAlbumSelected({ UID: "alb1", Title: "Vacation 2023" });
    expect(w.vm.pendingAlbumAdditions).toHaveLength(0);
  });

  // Album validation parity with batch edit + labels tab.
  it("should dedupe albums by normalized title even when UIDs differ", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.onAlbumSelected({ UID: "alb-other", Title: "vacation 2023" });
    expect(w.vm.pendingAlbumAdditions).toHaveLength(0);
  });

  it("should dedupe pending album additions by normalized title", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.pendingAlbumAdditions.push({ UID: "alb-a", Title: "Trip" });
    w.vm.onAlbumSelected({ UID: "alb-b", Title: "trip" });
    expect(w.vm.pendingAlbumAdditions).toHaveLength(1);
  });

  it("should reject overlong album titles in onAlbumEnter and not call save", () => {
    const saveSpy = vi.spyOn(Album.prototype, "save").mockResolvedValue();
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.editingField = "albums";
    w.vm.chipSearch = "a".repeat(CLIP_LEN + 10);
    w.vm.onAlbumEnter();
    expect(saveSpy).not.toHaveBeenCalled();
    expect(w.vm.pendingAlbumAdditions).toHaveLength(0);
    expect(w.vm.$notify.error).toHaveBeenCalledWith("Name too long");
    saveSpy.mockRestore();
  });

  it("should ignore empty/whitespace input in onAlbumEnter and not call save", () => {
    const saveSpy = vi.spyOn(Album.prototype, "save").mockResolvedValue();
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.editingField = "albums";
    w.vm.chipSearch = "   ";
    w.vm.onAlbumEnter();
    expect(saveSpy).not.toHaveBeenCalled();
    expect(w.vm.pendingAlbumAdditions).toHaveLength(0);
    saveSpy.mockRestore();
  });

  it("should skip onAlbumEnter when title matches existing album case-insensitively", () => {
    const saveSpy = vi.spyOn(Album.prototype, "save").mockResolvedValue();
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.editingField = "albums";
    w.vm.chipSearch = "VACATION 2023";
    w.vm.onAlbumEnter();
    expect(saveSpy).not.toHaveBeenCalled();
    expect(w.vm.pendingAlbumAdditions).toHaveLength(0);
    saveSpy.mockRestore();
  });

  it("should skip onAlbumEnter when title matches a pending addition case-insensitively", () => {
    const saveSpy = vi.spyOn(Album.prototype, "save").mockResolvedValue();
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.editingField = "albums";
    w.vm.pendingAlbumAdditions.push({ UID: "alb-pending", Title: "Trip" });
    w.vm.chipSearch = "trip";
    w.vm.onAlbumEnter();
    expect(saveSpy).not.toHaveBeenCalled();
    expect(w.vm.pendingAlbumAdditions).toHaveLength(1);
    saveSpy.mockRestore();
  });

  it("should create a new album in onAlbumEnter and add it to pending", async () => {
    const saveSpy = vi.spyOn(Album.prototype, "save").mockImplementation(function () {
      this.UID = "alb-created";
      return Promise.resolve(this);
    });
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.editingField = "albums";
    w.vm.albumOptions = [];
    w.vm.chipSearch = "Brand New Trip";
    w.vm.onAlbumEnter();
    await new Promise((r) => setTimeout(r, 0));
    expect(saveSpy).toHaveBeenCalledTimes(1);
    expect(w.vm.pendingAlbumAdditions).toHaveLength(1);
    expect(w.vm.pendingAlbumAdditions[0].Title).toBe("Brand New Trip");
    expect(w.vm.albumOptions.some((a) => a.UID === "alb-created")).toBe(true);
    saveSpy.mockRestore();
  });

  // cancelEditing clears all pending state
  it("should clear all pending state on cancelEditing", () => {
    const w = mount(PSidebarInfo, {
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    w.vm.editingField = "labels";
    w.vm.pendingLabelRemovals = [1];
    w.vm.pendingLabelAdditions = ["Sunset"];
    w.vm.pendingAlbumRemovals = ["alb1"];
    w.vm.pendingAlbumAdditions = [{ UID: "alb-new", Title: "New" }];

    w.vm._editStartedAt = Date.now() - 300;
    w.vm.cancelEditing();

    expect(w.vm.editingField).toBeNull();
    expect(w.vm.pendingLabelRemovals).toHaveLength(0);
    expect(w.vm.pendingLabelAdditions).toHaveLength(0);
    expect(w.vm.pendingAlbumRemovals).toHaveLength(0);
    expect(w.vm.pendingAlbumAdditions).toHaveLength(0);
  });

  // Photo watcher
  it("should cancel editing when photo changes", async () => {
    const w = mount(PSidebarInfo, {
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    w.vm.editingField = "title";
    w.vm._editStartedAt = Date.now() - 300;

    await w.setProps({ photo: { ...mockPhoto, Title: "Other" } });
    expect(w.vm.editingField).toBeNull();
  });

  // clearChipInput
  it("should reset chip state on clearChipInput", () => {
    const w = mount(PSidebarInfo, {
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    w.vm.chipInput = { Name: "test" };
    w.vm.chipSearch = "test";
    const prevKey = w.vm.chipKey;

    w.vm.clearChipInput();

    expect(w.vm.chipInput).toBeNull();
    expect(w.vm.chipSearch).toBe("");
    expect(w.vm.chipKey).toBe(prevKey + 1);
  });
});
