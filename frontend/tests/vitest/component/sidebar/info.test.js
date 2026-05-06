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
// mountSidebar accepts the same option shape that the suite used to pass
// directly to `mount(PSidebarInfo, ...)`. PSidebarInfo no longer takes the
// legacy props (modelValue/photo/canEdit/context/markersVisible/...); they
// live on the parent lightbox's $data and are surfaced through $view.getData()
// (see lightbox.vue). This helper extracts the legacy keys from `options.props`
// and exposes them as a $view mock so existing tests keep working with only
// `mount(PSidebarInfo, ` -> `mountSidebar(` renamed.
function mountSidebar(options = {}) {
  const props = { ...(options.props || {}) };
  const legacy = {
    model: props.modelValue,
    photo: props.photo,
    canEdit: props.canEdit,
    contextAllowsEdit: true,
    collection: props.collection,
    context: props.context,
    markersVisible: props.markersVisible,
    addingMarker: props.addingMarker,
    markersBusy: props.markersBusy,
    pendingNameMarkerUid: props.newMarkerUid,
  };

  // Strip legacy keys so Vue does not warn about unknown props.
  delete props.modelValue;
  delete props.photo;
  delete props.canEdit;
  delete props.collection;
  delete props.context;
  delete props.markersVisible;
  delete props.addingMarker;
  delete props.markersBusy;
  delete props.newMarkerUid;

  if (!props.uid && legacy.model && legacy.model.UID) {
    props.uid = legacy.model.UID;
  }

  const global = options.global || {};
  const mocks = global.mocks || {};
  const stubs = global.stubs || {};

  return mount(PSidebarInfo, {
    ...options,
    props,
    global: {
      ...global,
      stubs: { PMap: true, ...stubs },
      mocks: {
        $view: { getData: () => legacy },
        ...mocks,
      },
    },
  });
}

function mountInfoForChips(props) {
  return mountSidebar({
    props: { canEdit: true, context: contexts.Photos, ...props },
    global: {
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

    wrapper = mountSidebar({
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
    if (wrapper) {
      wrapper.unmount();
      wrapper = null;
    }
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
    const w = mountSidebar({
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
    const w = mountSidebar({
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
    const w = mountSidebar({
      props: { modelValue: mockModel, photo, canEdit: false, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(w.vm.cameraInfo).toBe("");
    expect(w.html()).not.toContain("mdi-camera ");
  });

  // Backend hydrates every photo with the "Unknown" placeholder camera
  // (CameraID=1, Camera={Make:"", Model:"Unknown"}). The read-only sidebar
  // must not surface that as an empty " Unknown" row.
  it("should hide camera row when backend returns the Unknown placeholder camera", () => {
    const photo = {
      ...mockPhoto,
      CameraID: 1,
      Camera: { ID: 1, Make: "", Model: "Unknown", Slug: "zz" },
      CameraMake: "",
      CameraModel: "Unknown",
      Iso: 0,
      Exposure: "",
      getCameraInfo: vi.fn().mockReturnValue(" Unknown"),
      getMarkers: vi.fn().mockReturnValue([]),
    };
    const w = mountSidebar({
      props: { modelValue: mockModel, photo, canEdit: false, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(w.vm.cameraInfo).toBe("");
    expect(w.html()).not.toContain("mdi-camera ");
  });

  // Editable users (admin) need an actionable camera row even when the
  // photo has no camera set — hide the cameraInfo text would leave the row
  // blank with just a pencil icon. Fall back to "Unknown" so the intent is
  // clear and the click target is discoverable.
  it("should show 'Unknown' in the camera row for editable users when no camera info is set", () => {
    const photo = {
      ...mockPhoto,
      CameraID: 1,
      Camera: { ID: 1, Make: "", Model: "Unknown", Slug: "zz" },
      CameraMake: "",
      CameraModel: "Unknown",
      Iso: 0,
      Exposure: "",
      getCameraInfo: vi.fn().mockReturnValue(" Unknown"),
      getMarkers: vi.fn().mockReturnValue([]),
    };
    const w = mountSidebar({
      props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    // cameraInfo itself stays suppressed; the template falls back to the
    // localized "Unknown" placeholder only when the row is visible.
    expect(w.vm.cameraInfo).toBe("");
    const cameraRow = w.find(".v-list-item [class*='mdi-camera']")?.element?.closest(".v-list-item");
    expect(cameraRow).toBeTruthy();
    expect(cameraRow.textContent).toContain("Unknown");
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
    const w = mountSidebar({
      props: { modelValue: mockModel, photo, canEdit: false, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(w.vm.lensInfo).toBe("");
  });

  it("should return empty strings when photo prop is null", () => {
    const w = mountSidebar({
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

  it("should make the avatar of named people clickable for navigation", () => {
    const personRows = wrapper.findAll(".metadata__person-row");
    const namedAvatar = personRows[0].find(".meta-person__avatar");
    const unnamedAvatar = personRows[1].find(".meta-person__avatar");
    expect(namedAvatar.classes()).toContain("clickable");
    expect(unnamedAvatar.classes()).not.toContain("clickable");
  });

  // People section: face marker buttons (show/hide + add) and per-row remove.
  // The sidebar emits events; the parent lightbox owns the actual state.
  it("should render the People header with show/hide and + buttons when editable", () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(w.html()).toContain("People");
    expect(w.find(".meta-markers-toggle").exists()).toBe(true);
    expect(w.find(".meta-marker-add").exists()).toBe(true);
  });

  it("should render the People header even when there are no markers, when editable", () => {
    const photo = { ...mockPhoto, getMarkers: vi.fn().mockReturnValue([]) };
    const w = mountSidebar({
      props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(w.html()).toContain("People");
    expect(w.find(".meta-markers-toggle").exists()).toBe(true);
    expect(w.find(".meta-marker-add").exists()).toBe(true);
  });

  it("should not render the show/hide or + icons when not editable", () => {
    // The default wrapper is mounted without canEdit → isEditable false.
    expect(wrapper.find(".meta-markers-toggle").exists()).toBe(false);
    expect(wrapper.find(".meta-marker-add").exists()).toBe(false);
  });

  it("should still list named people when not editable but the photo has markers", () => {
    expect(wrapper.find(".metadata__person-row").exists()).toBe(true);
    expect(wrapper.html()).toContain("Jane Doe");
  });

  it("should hide the People section entirely when not editable and there are no markers", () => {
    const photo = { ...mockPhoto, getMarkers: vi.fn().mockReturnValue([]) };
    const w = mountSidebar({
      props: { modelValue: mockModel, photo, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(w.html()).not.toContain("People");
  });

  it("should reflect the markersVisible prop on the toggle icon", () => {
    const w = mountSidebar({
      props: {
        modelValue: mockModel,
        photo: mockPhoto,
        canEdit: true,
        context: contexts.Photos,
        markersVisible: true,
      },
      global: { stubs: { PMap: true } },
    });
    const toggle = w.find(".meta-markers-toggle");
    expect(toggle.classes()).toContain("is-active");
  });

  it("should emit toggle-markers-visible when the eye icon is clicked", async () => {
    const onToggle = vi.fn();
    const w = mountSidebar({
      props: {
        "modelValue": mockModel,
        "photo": mockPhoto,
        "canEdit": true,
        "context": contexts.Photos,
        "onToggle-markers-visible": onToggle,
      },
      global: { stubs: { PMap: true } },
    });
    await w.find(".meta-markers-toggle").trigger("click");
    expect(onToggle).toHaveBeenCalled();
  });

  it("should emit toggle-adding-marker when the + icon is clicked", async () => {
    const onToggle = vi.fn();
    const w = mountSidebar({
      props: {
        "modelValue": mockModel,
        "photo": mockPhoto,
        "canEdit": true,
        "context": contexts.Photos,
        "onToggle-adding-marker": onToggle,
      },
      global: { stubs: { PMap: true } },
    });
    const icon = w.find(".meta-marker-add");
    expect(icon.classes()).toContain("mdi-plus");
    await icon.trigger("click");
    expect(onToggle).toHaveBeenCalled();
  });

  it("should swap the add icon to a check while addingMarker is true", () => {
    const w = mountSidebar({
      props: {
        modelValue: mockModel,
        photo: mockPhoto,
        canEdit: true,
        context: contexts.Photos,
        addingMarker: true,
      },
      global: { stubs: { PMap: true } },
    });
    const icon = w.find(".meta-marker-add");
    expect(icon.classes()).toContain("is-active");
    expect(icon.classes()).toContain("mdi-check");
  });

  it("should still emit toggle-adding-marker when addingMarker is true (so the user can exit)", async () => {
    const onToggle = vi.fn();
    const w = mountSidebar({
      props: {
        "modelValue": mockModel,
        "photo": mockPhoto,
        "canEdit": true,
        "context": contexts.Photos,
        "addingMarker": true,
        "onToggle-adding-marker": onToggle,
      },
      global: { stubs: { PMap: true } },
    });
    await w.find(".meta-marker-add").trigger("click");
    expect(onToggle).toHaveBeenCalled();
  });

  it("should not render the per-row remove icon on a named marker", () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    const personRows = w.findAll(".metadata__person-row");
    // First row is "Jane Doe" with SubjUID — no remove icon.
    expect(personRows[0].find(".meta-marker-remove").exists()).toBe(false);
  });

  it("should render the per-row remove icon on an unnamed marker when editable", () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    const personRows = w.findAll(".metadata__person-row");
    // Second row is the unnamed marker.
    expect(personRows[1].find(".meta-marker-remove").exists()).toBe(true);
  });

  it("should not render the per-row remove icon when not editable", () => {
    // wrapper is mounted without canEdit.
    const personRows = wrapper.findAll(".metadata__person-row");
    expect(personRows[1].find(".meta-marker-remove").exists()).toBe(false);
  });

  it("should emit remove-marker with the marker when the remove icon is clicked", async () => {
    const onRemove = vi.fn();
    const w = mountSidebar({
      props: {
        "modelValue": mockModel,
        "photo": mockPhoto,
        "canEdit": true,
        "context": contexts.Photos,
        "onRemove-marker": onRemove,
      },
      global: { stubs: { PMap: true } },
    });
    const personRows = w.findAll(".metadata__person-row");
    await personRows[1].find(".meta-marker-remove").trigger("click");
    expect(onRemove).toHaveBeenCalledTimes(1);
    expect(onRemove.mock.calls[0][0].UID).toBe("m2");
  });

  it("should refuse to emit remove-marker on a marker that has a SubjUID", () => {
    const onRemove = vi.fn();
    const w = mountSidebar({
      props: {
        "modelValue": mockModel,
        "photo": mockPhoto,
        "canEdit": true,
        "context": contexts.Photos,
        "onRemove-marker": onRemove,
      },
      global: { stubs: { PMap: true } },
    });
    w.vm.onRemoveMarker({ UID: "mNamed", SubjUID: "subjX", Name: "Alice" });
    expect(onRemove).not.toHaveBeenCalled();
  });

  it("should refuse to emit toggle-markers-visible / toggle-adding-marker / remove-marker while markersBusy is true", () => {
    const onToggleVisible = vi.fn();
    const onToggleAdd = vi.fn();
    const onRemove = vi.fn();
    const w = mountSidebar({
      props: {
        "modelValue": mockModel,
        "photo": mockPhoto,
        "canEdit": true,
        "context": contexts.Photos,
        "markersBusy": true,
        "onToggle-markers-visible": onToggleVisible,
        "onToggle-adding-marker": onToggleAdd,
        "onRemove-marker": onRemove,
      },
      global: { stubs: { PMap: true } },
    });
    w.vm.onToggleMarkersVisible();
    w.vm.onToggleAddingMarker();
    w.vm.onRemoveMarker({ UID: "mX", SubjUID: "" });
    expect(onToggleVisible).not.toHaveBeenCalled();
    expect(onToggleAdd).not.toHaveBeenCalled();
    expect(onRemove).not.toHaveBeenCalled();
  });

  it("should render an inline name input for every marker when canEdit is true", () => {
    const w = mountSidebar({
      props: {
        modelValue: mockModel,
        photo: mockPhoto,
        canEdit: true,
        context: contexts.Photos,
        markersVisible: true,
        addingMarker: true,
      },
      global: { stubs: { PMap: true } },
    });
    // One input per marker row — no pencil click required to edit.
    const inputs = w.findAll(".meta-inline-marker");
    expect(inputs.length).toBe(2);
    expect(w.find(".meta-inline-pencil.meta-person-pencil").exists()).toBe(false);
  });

  // Feature flag gate (Task 4): when $config.feature('people') is false,
  // the entire People section is hidden regardless of marker presence.
  it("should hide the People section when the people feature is disabled", () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: {
        stubs: { PMap: true },
        mocks: {
          $config: { feature: (key) => key !== "people", get: () => false, allow: () => true, values: {}, dir: () => "ltr" },
        },
      },
    });
    expect(w.html()).not.toContain("People");
    expect(w.find(".meta-markers-toggle").exists()).toBe(false);
    expect(w.find(".meta-marker-add").exists()).toBe(false);
    expect(w.find(".metadata__person-row").exists()).toBe(false);
  });

  // Eject button (Task 2): named markers expose mdi-eject.
  it("should render the eject icon on a named marker and emit eject-marker on click", async () => {
    const onEject = vi.fn();
    const w = mountSidebar({
      props: {
        "modelValue": mockModel,
        "photo": mockPhoto,
        "canEdit": true,
        "context": contexts.Photos,
        "onEject-marker": onEject,
      },
      global: { stubs: { PMap: true } },
    });
    const personRows = w.findAll(".metadata__person-row");
    // First row: Jane Doe (SubjUID set).
    const ejectIcon = personRows[0].find(".meta-marker-eject");
    expect(ejectIcon.exists()).toBe(true);
    // Unnamed marker should NOT have an eject icon.
    expect(personRows[1].find(".meta-marker-eject").exists()).toBe(false);
    await ejectIcon.trigger("click");
    expect(onEject).toHaveBeenCalledTimes(1);
    expect(onEject.mock.calls[0][0].UID).toBe("m1");
  });

  // Named markers keep the combobox (so named and unnamed rows look
  // identical) but render it readonly — renaming a named face requires
  // ejecting first. Mirrors the edit dialog's People tab gate.
  it("should render the named marker combobox as readonly", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    const personRows = w.findAll(".metadata__person-row");
    const namedInput = personRows[0].find(".meta-inline-marker--named input");
    expect(namedInput.exists()).toBe(true);
    expect(namedInput.element.readOnly).toBe(true);
    const unnamedInput = personRows[1].find(".meta-inline-marker input");
    expect(unnamedInput.exists()).toBe(true);
    expect(unnamedInput.element.readOnly).toBe(false);
  });

  it("should refuse to emit eject-marker on a marker without SubjUID", () => {
    const onEject = vi.fn();
    const w = mountSidebar({
      props: {
        "modelValue": mockModel,
        "photo": mockPhoto,
        "canEdit": true,
        "context": contexts.Photos,
        "onEject-marker": onEject,
      },
      global: { stubs: { PMap: true } },
    });
    w.vm.onEjectMarker({ UID: "mX", SubjUID: "", Name: "" });
    expect(onEject).not.toHaveBeenCalled();
  });

  it("should refuse to emit eject-marker while markersBusy is true", () => {
    const onEject = vi.fn();
    const w = mountSidebar({
      props: {
        "modelValue": mockModel,
        "photo": mockPhoto,
        "canEdit": true,
        "context": contexts.Photos,
        "markersBusy": true,
        "onEject-marker": onEject,
      },
      global: { stubs: { PMap: true } },
    });
    w.vm.onEjectMarker({ UID: "m1", SubjUID: "subj1", Name: "Jane Doe" });
    expect(onEject).not.toHaveBeenCalled();
  });

  // pendingNameMarkerUid focuses the input on the freshly-created marker and
  // emits naming-started so the parent (lightbox) can clear its own state.
  // The value lives on the parent's $view-data; mutating it through the
  // captured reference triggers the `newMarkerUid` computed and its watcher.
  it("should focus the matching marker input when pendingNameMarkerUid changes", async () => {
    const onNamingStarted = vi.fn();
    const w = mountSidebar({
      props: {
        "modelValue": mockModel,
        "photo": mockPhoto,
        "canEdit": true,
        "context": contexts.Photos,
        "newMarkerUid": null,
        "onNaming-started": onNamingStarted,
      },
      global: { stubs: { PMap: true } },
      attachTo: document.body,
    });
    w.vm.view.pendingNameMarkerUid = "m2";
    await w.vm.$nextTick();
    await w.vm.$nextTick();
    expect(onNamingStarted).toHaveBeenCalled();
    w.unmount();
  });

  it("should call marker.setName and emit reload-markers when confirming an inline name", async () => {
    const setName = vi.fn().mockResolvedValue(undefined);
    const namedMarker = {
      UID: "m2",
      CropID: "crop2",
      Name: "",
      SubjUID: "",
      thumbnailUrl: () => "/t/thumb2/public/tile_160",
      setName,
    };
    const photo = { ...mockPhoto, getMarkers: vi.fn().mockReturnValue([namedMarker]) };
    const onReload = vi.fn();
    const w = mountSidebar({
      props: {
        "modelValue": mockModel,
        "photo": photo,
        "canEdit": true,
        "context": contexts.Photos,
        "onReload-markers": onReload,
      },
      global: { stubs: { PMap: true } },
    });
    w.vm.setMarkerInputValue("m2", "Alice");
    w.vm.confirmMarkerName(namedMarker);
    expect(namedMarker.Name).toBe("Alice");
    expect(setName).toHaveBeenCalled();
    await new Promise((r) => setTimeout(r, 0));
    expect(onReload).toHaveBeenCalled();
    expect(w.vm.hasPendingEdit()).toBe(false);
  });

  it("should revert the draft to the marker's saved name when cancelMarkerName runs", async () => {
    const markers = mockPhoto.getMarkers();
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    await w.vm.$nextTick();
    w.vm.setMarkerInputValue("m1", "Changed");
    expect(w.vm.hasPendingEdit()).toBe(true);
    w.vm.cancelMarkerName(markers[0]);
    expect(w.vm.markerInputValue("m1")).toBe("Jane Doe");
    expect(w.vm.hasPendingEdit()).toBe(false);
  });

  it("should link an existing subject when the typed name matches a known person", async () => {
    const setName = vi.fn().mockResolvedValue(undefined);
    const namedMarker = {
      UID: "m2",
      Name: "",
      SubjUID: "",
      thumbnailUrl: () => "/t/thumb2/public/tile_160",
      setName,
    };
    const photo = { ...mockPhoto, getMarkers: vi.fn().mockReturnValue([namedMarker]) };
    const knownPersonConfig = {
      feature: () => true,
      get: () => false,
      getSettings: () => ({ features: { edit: true } }),
      allow: () => true,
      featExperimental: () => false,
      featDevelop: () => false,
      values: { people: [{ UID: "sXYZ", Name: "Alice Smith" }] },
      dir: () => "ltr",
    };
    const w = mountSidebar({
      props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
      global: {
        stubs: { PMap: true },
        mocks: { $config: knownPersonConfig },
      },
    });
    await w.vm.$nextTick();
    w.vm.setMarkerInputValue("m2", "alice smith");
    w.vm.confirmMarkerName(namedMarker);
    expect(namedMarker.Name).toBe("Alice Smith");
    expect(namedMarker.SubjUID).toBe("sXYZ");
    expect(setName).toHaveBeenCalled();
  });

  it("should commit immediately when onPickPerson selects a dropdown item", async () => {
    const setName = vi.fn().mockResolvedValue(undefined);
    const namedMarker = {
      UID: "m2",
      Name: "",
      SubjUID: "",
      thumbnailUrl: () => "/t/thumb2/public/tile_160",
      setName,
    };
    const photo = { ...mockPhoto, getMarkers: vi.fn().mockReturnValue([namedMarker]) };
    const knownPersonConfig = {
      feature: () => true,
      get: () => false,
      getSettings: () => ({ features: { edit: true } }),
      allow: () => true,
      featExperimental: () => false,
      featDevelop: () => false,
      values: { people: [{ UID: "sBOB", Name: "Bob" }] },
      dir: () => "ltr",
    };
    const w = mountSidebar({
      props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
      global: {
        stubs: { PMap: true },
        mocks: { $config: knownPersonConfig },
      },
    });
    await w.vm.$nextTick();
    w.vm.onPickPerson(namedMarker, { UID: "sBOB", Name: "Bob" });
    expect(namedMarker.Name).toBe("Bob");
    expect(namedMarker.SubjUID).toBe("sBOB");
    expect(setName).toHaveBeenCalled();
  });

  it("should expose knownPeople from $config.values.people", () => {
    const knownPersonConfig = {
      feature: () => true,
      get: () => false,
      getSettings: () => ({ features: { edit: true } }),
      allow: () => true,
      featExperimental: () => false,
      featDevelop: () => false,
      values: {
        people: [
          { UID: "sA", Name: "Alice" },
          { UID: "sB", Name: "Bob" },
        ],
      },
      dir: () => "ltr",
    };
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: {
        stubs: { PMap: true },
        mocks: { $config: knownPersonConfig },
      },
    });
    expect(w.vm.knownPeople).toHaveLength(2);
    expect(w.vm.knownPeople[0].Name).toBe("Alice");
  });

  it("should fall back to an empty knownPeople list when values.people is missing", () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(w.vm.knownPeople).toEqual([]);
  });

  it("should not confirm an empty name (treats it as cancel)", async () => {
    const setName = vi.fn();
    const namedMarker = {
      UID: "m2",
      Name: "",
      SubjUID: "",
      thumbnailUrl: () => "/t/thumb2/public/tile_160",
      setName,
    };
    const photo = { ...mockPhoto, getMarkers: vi.fn().mockReturnValue([namedMarker]) };
    const w = mountSidebar({
      props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    await w.vm.$nextTick();
    w.vm.setMarkerInputValue("m2", "   ");
    w.vm.confirmMarkerName(namedMarker);
    expect(setName).not.toHaveBeenCalled();
  });

  // Blur on an unnamed marker with a new name (and no matching known person)
  // must prompt the user before persisting, mirroring the people-tab pattern.
  it("should open the Add-name dialog on blur for an unnamed marker with a novel name", async () => {
    const setName = vi.fn().mockResolvedValue(undefined);
    const marker = {
      UID: "m2",
      Name: "",
      SubjUID: "",
      thumbnailUrl: () => "/t/thumb2/public/tile_160",
      setName,
    };
    const photo = { ...mockPhoto, getMarkers: vi.fn().mockReturnValue([marker]) };
    const w = mountInfoForChips({ modelValue: mockModel, photo });
    await w.vm.$nextTick();
    w.vm.setMarkerInputValue("m2", "Plane Port");
    w.vm.confirmMarkerName(marker, "blur");
    expect(setName).not.toHaveBeenCalled();
    expect(w.vm.addNameDialog.visible).toBe(true);
    expect(w.vm.addNameDialog.marker?.UID).toBe("m2");
    expect(w.vm.addNameDialog.name).toBe("Plane Port");
  });

  it("should persist when the Add-name dialog is confirmed", async () => {
    const setName = vi.fn().mockResolvedValue(undefined);
    const marker = {
      UID: "m2",
      Name: "",
      SubjUID: "",
      thumbnailUrl: () => "/t/thumb2/public/tile_160",
      setName,
    };
    const photo = { ...mockPhoto, getMarkers: vi.fn().mockReturnValue([marker]) };
    const w = mountInfoForChips({ modelValue: mockModel, photo });
    await w.vm.$nextTick();
    w.vm.setMarkerInputValue("m2", "Plane Port");
    w.vm.confirmMarkerName(marker, "blur");
    expect(w.vm.addNameDialog.visible).toBe(true);
    w.vm.onAddNameConfirm();
    expect(marker.Name).toBe("Plane Port");
    expect(setName).toHaveBeenCalledTimes(1);
    expect(w.vm.addNameDialog.visible).toBe(false);
  });

  it("should revert the input and skip save when the Add-name dialog is canceled", async () => {
    const setName = vi.fn();
    const marker = {
      UID: "m2",
      Name: "",
      SubjUID: "",
      thumbnailUrl: () => "/t/thumb2/public/tile_160",
      setName,
    };
    const photo = { ...mockPhoto, getMarkers: vi.fn().mockReturnValue([marker]) };
    const w = mountInfoForChips({ modelValue: mockModel, photo });
    await w.vm.$nextTick();
    w.vm.setMarkerInputValue("m2", "Plane Port");
    w.vm.confirmMarkerName(marker, "blur");
    w.vm.onAddNameCancel();
    expect(setName).not.toHaveBeenCalled();
    expect(w.vm.addNameDialog.visible).toBe(false);
    expect(w.vm.markerInputValue("m2")).toBe("");
  });

  // Skip the dialog when the typed name already resolves to a known subject —
  // there's nothing ambiguous to ask about.
  it("should skip the Add-name dialog and save immediately on blur when the name matches a known person", async () => {
    const setName = vi.fn().mockResolvedValue(undefined);
    const marker = {
      UID: "m2",
      Name: "",
      SubjUID: "",
      thumbnailUrl: () => "/t/thumb2/public/tile_160",
      setName,
    };
    const photo = { ...mockPhoto, getMarkers: vi.fn().mockReturnValue([marker]) };
    const knownPersonConfig = {
      ...validationConfig,
      values: { people: [{ UID: "sALC", Name: "Alice Smith" }] },
    };
    const w = mountSidebar({
      props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
      global: {
        stubs: { PMap: true },
        mocks: { $config: knownPersonConfig, $util: validationUtil },
      },
    });
    await w.vm.$nextTick();
    w.vm.setMarkerInputValue("m2", "alice smith");
    w.vm.confirmMarkerName(marker, "blur");
    expect(w.vm.addNameDialog.visible).toBe(false);
    expect(setName).toHaveBeenCalled();
    expect(marker.Name).toBe("Alice Smith");
    expect(marker.SubjUID).toBe("sALC");
  });

  // Inline blur now commits the edit instead of silently reverting — this
  // was the bug where typing in an inline field and clicking away (or
  // navigating) would quietly lose the change.
  it("should commit the edit on blur (onInlineFieldBlur calls confirmField)", async () => {
    const update = vi.fn().mockResolvedValue(undefined);
    const photo = {
      ...mockPhoto,
      Title: "Original",
      update,
      wasChanged: function () {
        return this.Title !== "Original";
      },
    };
    const w = mountSidebar({
      props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    w.vm.startEditing("title");
    await w.vm.$nextTick();
    photo.Title = "Changed";
    w.vm._editStartedAt = null;
    w.vm.onInlineFieldBlur();
    expect(update).toHaveBeenCalled();
    expect(w.vm.editingField).toBeNull();
    // Value must NOT be reverted to "Original".
    expect(photo.Title).toBe("Changed");
  });

  it("should respect the 200ms debounce guard on blur", async () => {
    const update = vi.fn().mockResolvedValue(undefined);
    const photo = {
      ...mockPhoto,
      Title: "Original",
      update,
      wasChanged: () => true,
    };
    const w = mountSidebar({
      props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    w.vm.startEditing("title");
    await w.vm.$nextTick();
    // _editStartedAt was just set by startEditing; blur should be a no-op.
    w.vm.onInlineFieldBlur();
    expect(update).not.toHaveBeenCalled();
    expect(w.vm.editingField).toBe("title");
  });

  it("should still cancel on Escape via cancelEditing", async () => {
    const photo = {
      ...mockPhoto,
      Title: "Original",
      update: vi.fn(),
      wasChanged: () => true,
    };
    const w = mountSidebar({
      props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    w.vm.startEditing("title");
    await w.vm.$nextTick();
    photo.Title = "Changed";
    w.vm._editStartedAt = null;
    w.vm.cancelEditing();
    // Escape/cancelEditing reverts to the stored original.
    expect(photo.Title).toBe("Original");
    expect(w.vm.editingField).toBeNull();
  });

  // Inline text fields (title/caption/subject/...) are intentionally NOT
  // tracked by hasPendingEdit: onInlineFieldBlur() auto-commits them before
  // any nav source can fire, so they can never be pending at nav time.
  it("should NOT report hasPendingEdit for a dirty inline text field (auto-commits on blur)", () => {
    const photo = {
      ...mockPhoto,
      Title: "Changed",
      wasChanged: () => true,
    };
    const w = mountSidebar({
      props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    w.vm.editingField = "title";
    expect(w.vm.hasPendingEdit()).toBe(false);
  });

  it("should report hasPendingEdit when labels have pending additions or removals", () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(w.vm.hasPendingEdit()).toBe(false);
    w.vm.chipState.labels.additions = ["New Label"];
    expect(w.vm.hasPendingEdit()).toBe(true);
    w.vm.chipState.labels.additions = [];
    w.vm.chipState.labels.removals = [{ Label: { UID: "lbl1" } }];
    expect(w.vm.hasPendingEdit()).toBe(true);
  });

  it("should report hasPendingEdit when albums have pending additions or removals", () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(w.vm.hasPendingEdit()).toBe(false);
    w.vm.chipState.albums.additions = [{ UID: "alb-new", Title: "New" }];
    expect(w.vm.hasPendingEdit()).toBe(true);
    w.vm.chipState.albums.additions = [];
    w.vm.chipState.albums.removals = [{ UID: "alb1" }];
    expect(w.vm.hasPendingEdit()).toBe(true);
  });

  it("should report hasPendingEdit while an inline marker name is dirty", async () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    await w.vm.$nextTick();
    expect(w.vm.hasPendingEdit()).toBe(false);
    w.vm.setMarkerInputValue("m2", "Alice");
    expect(w.vm.hasPendingEdit()).toBe(true);
  });

  it("should resolve confirmDiscardPending to true immediately when there is no pending edit", async () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    await expect(w.vm.confirmDiscardPending()).resolves.toBe(true);
    expect(w.vm.discardDialog.visible).toBe(false);
  });

  it("should open the discard dialog when confirmDiscardPending is called with a pending edit", async () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    await w.vm.$nextTick();
    w.vm.setMarkerInputValue("m2", "Alice");
    const promise = w.vm.confirmDiscardPending();
    expect(w.vm.discardDialog.visible).toBe(true);
    expect(typeof w.vm.discardDialog.resolver).toBe("function");
    // Resolve the dialog via the confirm handler so the test doesn't hang.
    w.vm.onDiscardConfirm();
    await expect(promise).resolves.toBe(true);
    expect(w.vm.discardDialog.visible).toBe(false);
    expect(w.vm.hasPendingEdit()).toBe(false);
  });

  it("should clear pending edits when onDiscardConfirm runs", async () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    await w.vm.$nextTick();
    w.vm.setMarkerInputValue("m2", "Alice");
    const promise = w.vm.confirmDiscardPending();
    w.vm.onDiscardConfirm();
    await expect(promise).resolves.toBe(true);
    expect(w.vm.hasPendingEdit()).toBe(false);
    expect(w.vm.markerInputValue("m2")).toBe("");
  });

  it("should keep pending edits when onDiscardCancel runs", async () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    await w.vm.$nextTick();
    w.vm.setMarkerInputValue("m2", "Alice");
    const promise = w.vm.confirmDiscardPending();
    w.vm.onDiscardCancel();
    await expect(promise).resolves.toBe(false);
    expect(w.vm.markerInputValue("m2")).toBe("Alice");
    expect(w.vm.hasPendingEdit()).toBe(true);
  });

  // confirmCamera rolls back the optimistic mutation if the API call fails
  // and surfaces the failure via $notify.error so the user knows the save
  // didn't take effect. Uses the model's __originalValues snapshot via the
  // originalValue() accessor.
  describe("confirmCamera failure rollback", () => {
    function buildCameraPhoto(overrides = {}) {
      const photo = {
        UID: "ps6sg6be2lvl0yh7",
        CameraID: 1,
        LensID: 1,
        Iso: 100,
        Exposure: "1/200",
        FNumber: 2.8,
        FocalLength: 50,
        Files: [],
        Labels: [],
        Albums: [],
        Details: {},
        getMarkers: () => [],
        getCameraInfo: () => "",
        getLensInfo: () => "",
        getImageInfo: () => "",
        getVideoInfo: () => "",
        getVectorInfo: () => "",
        getExifInfo: () => "",
        locationInfo: () => "",
        update: vi.fn(),
        ...overrides,
      };
      // Mirror the Model.__originalValues contract: originalValue(key) reads
      // the snapshot taken when the photo was last loaded/saved, and
      // rollback() restores every tracked field from that snapshot.
      photo.__originalValues = {
        CameraID: photo.CameraID,
        LensID: photo.LensID,
        Iso: photo.Iso,
        Exposure: photo.Exposure,
        FNumber: photo.FNumber,
        FocalLength: photo.FocalLength,
      };
      photo.originalValue = (key) => photo.__originalValues[key];
      photo.rollback = () => {
        Object.keys(photo.__originalValues).forEach((key) => {
          photo[key] = photo.__originalValues[key];
        });
        return photo;
      };
      return photo;
    }

    const newCameraData = {
      CameraID: 5,
      LensID: 7,
      Iso: 800,
      Exposure: "1/60",
      FNumber: 1.8,
      FocalLength: 35,
    };

    it("rolls back camera fields and notifies the user when update() rejects", async () => {
      const photo = buildCameraPhoto();
      photo.update.mockRejectedValueOnce(new Error("boom"));

      const w = mountSidebar({
        props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
        global: { stubs: { PMap: true } },
      });

      w.vm.confirmCamera(newCameraData);

      // Optimistic write applies before the rejection lands.
      expect(photo.CameraID).toBe(5);
      expect(photo.update).toHaveBeenCalledTimes(1);

      await w.vm.$nextTick();
      await w.vm.$nextTick();

      // Rolled back to __originalValues — must match the pre-confirmCamera state.
      expect(photo.CameraID).toBe(1);
      expect(photo.LensID).toBe(1);
      expect(photo.Iso).toBe(100);
      expect(photo.Exposure).toBe("1/200");
      expect(photo.FNumber).toBe(2.8);
      expect(photo.FocalLength).toBe(50);

      expect(w.vm.$notify.error).toHaveBeenCalledTimes(1);
    });

    it("keeps the optimistic mutation when update() resolves", async () => {
      const photo = buildCameraPhoto();
      photo.update.mockResolvedValueOnce({});

      const w = mountSidebar({
        props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
        global: { stubs: { PMap: true } },
      });

      w.vm.confirmCamera(newCameraData);

      await w.vm.$nextTick();
      await w.vm.$nextTick();

      expect(photo.CameraID).toBe(5);
      expect(photo.LensID).toBe(7);
      expect(photo.Iso).toBe(800);
      expect(w.vm.$notify.error).not.toHaveBeenCalled();
    });

    it("does nothing when the photo has no UID", () => {
      const photo = buildCameraPhoto({ UID: "" });

      const w = mountSidebar({
        props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
        global: { stubs: { PMap: true } },
      });

      w.vm.confirmCamera(newCameraData);

      expect(photo.update).not.toHaveBeenCalled();
      expect(photo.CameraID).toBe(1);
    });
  });

  // Symmetric coverage for the other three save paths. Each handler mirrors
  // confirmCamera: mutate the photo optimistically, call photo.update(), on
  // success sync the matching Thumb fields, on failure roll back and notify.
  // The Photo mock fakes the Model.__originalValues + rollback() contract so
  // these tests stay decoupled from the real model implementation.
  function attachRollbackContract(photo, snapshot) {
    photo.__originalValues = { ...snapshot };
    photo.originalValue = (key) => photo.__originalValues[key];
    photo.rollback = () => {
      Object.keys(photo.__originalValues).forEach((key) => {
        photo[key] = photo.__originalValues[key];
      });
      return photo;
    };
    return photo;
  }

  function buildSidebarPhoto(initial, overrides = {}) {
    const photo = {
      UID: "ps6sg6be2lvl0yh7",
      Files: [],
      Labels: [],
      Albums: [],
      Details: {},
      getMarkers: () => [],
      getCameraInfo: () => "",
      getLensInfo: () => "",
      getImageInfo: () => "",
      getVideoInfo: () => "",
      getVectorInfo: () => "",
      getExifInfo: () => "",
      locationInfo: () => "",
      wasChanged: vi.fn().mockReturnValue(true),
      ...initial,
      update: vi.fn(),
      ...overrides,
    };
    return attachRollbackContract(photo, initial);
  }

  describe("confirmDateTime failure rollback", () => {
    function buildDateTimePhoto(overrides = {}) {
      const photo = buildSidebarPhoto(
        {
          Day: 1,
          Month: 1,
          Year: 2020,
          TimeZone: "UTC",
          TakenAtLocal: "2020-01-01T00:00:00Z",
          TakenAt: "2020-01-01T00:00:00Z",
        },
        overrides
      );
      photo.localDate = vi.fn().mockReturnValue({
        isValid: true,
        toISO: () => "2022-07-15T13:45:30",
      });
      photo.currentTimeZoneUTC = vi.fn().mockReturnValue(true);
      return photo;
    }

    // Use UTC so the success-path test stays inside the suite-wide
    // DateTime.fromISO mock surface (which only stubs toLocaleString,
    // not setZone — see beforeEach above). The rollback-path test
    // never re-renders formatTime() because the mutations get reverted
    // before Vue flushes, so any timezone works there.
    const newDateTimeData = {
      Day: 15,
      Month: 7,
      Year: 2022,
      TimeZone: "UTC",
      time: "13:45:30",
    };

    it("rolls back date/time fields and notifies the user when update() rejects", async () => {
      const photo = buildDateTimePhoto();
      photo.update.mockRejectedValueOnce(new Error("boom"));

      const w = mountSidebar({
        props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
        global: { stubs: { PMap: true } },
      });

      w.vm.confirmDateTime(newDateTimeData);
      // Optimistic write applied before the rejection lands.
      expect(photo.Year).toBe(2022);

      await w.vm.$nextTick();
      await w.vm.$nextTick();

      expect(photo.Day).toBe(1);
      expect(photo.Month).toBe(1);
      expect(photo.Year).toBe(2020);
      expect(photo.TimeZone).toBe("UTC");
      expect(w.vm.$notify.error).toHaveBeenCalledTimes(1);
    });

    it("keeps the optimistic mutation when update() resolves", async () => {
      const photo = buildDateTimePhoto();
      photo.update.mockResolvedValueOnce({});

      const w = mountSidebar({
        props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
        global: { stubs: { PMap: true } },
      });

      w.vm.confirmDateTime(newDateTimeData);

      await w.vm.$nextTick();
      await w.vm.$nextTick();

      expect(photo.Year).toBe(2022);
      expect(photo.TimeZone).toBe("UTC");
      expect(w.vm.$notify.error).not.toHaveBeenCalled();
    });
  });

  describe("confirmLocation failure rollback", () => {
    function buildLocationPhoto(overrides = {}) {
      return buildSidebarPhoto(
        {
          Lat: 0,
          Lng: 0,
          PlaceSrc: "",
          Country: "zz",
        },
        overrides
      );
    }

    const newLocationData = {
      lat: 52.52,
      lng: 13.405,
      location: { country: "de" },
    };

    it("rolls back location fields and notifies the user when update() rejects", async () => {
      const photo = buildLocationPhoto();
      photo.update.mockRejectedValueOnce(new Error("boom"));

      const w = mountSidebar({
        props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
        global: { stubs: { PMap: true } },
      });

      w.vm.confirmLocation(newLocationData);
      // Optimistic write applied before the rejection lands.
      expect(photo.Lat).toBe(52.52);

      await w.vm.$nextTick();
      await w.vm.$nextTick();

      expect(photo.Lat).toBe(0);
      expect(photo.Lng).toBe(0);
      expect(photo.PlaceSrc).toBe("");
      expect(photo.Country).toBe("zz");
      expect(w.vm.$notify.error).toHaveBeenCalledTimes(1);
    });

    it("keeps the optimistic mutation when update() resolves", async () => {
      const photo = buildLocationPhoto();
      photo.update.mockResolvedValueOnce({});

      const w = mountSidebar({
        props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
        global: { stubs: { PMap: true } },
      });

      w.vm.confirmLocation(newLocationData);

      await w.vm.$nextTick();
      await w.vm.$nextTick();

      expect(photo.Lat).toBe(52.52);
      expect(photo.Lng).toBe(13.405);
      expect(photo.PlaceSrc).toBe("manual");
      expect(photo.Country).toBe("de");
      expect(w.vm.$notify.error).not.toHaveBeenCalled();
    });
  });

  describe("confirmField (inline edit) failure rollback", () => {
    function buildInlineEditPhoto(overrides = {}) {
      // The inline-edit binding is v-model="p.Title", so the user's keystrokes
      // mutate photo.Title directly before confirmField() runs.
      return buildSidebarPhoto({ Title: "Original", Caption: "Original caption" }, overrides);
    }

    it("rolls back the inline edit and notifies the user when update() rejects", async () => {
      const photo = buildInlineEditPhoto();
      photo.update.mockRejectedValueOnce(new Error("boom"));

      const w = mountSidebar({
        props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
        global: { stubs: { PMap: true } },
      });

      // Simulate the v-model edit having already mutated photo.Title.
      photo.Title = "Edited";
      w.vm.editingField = "title";

      w.vm.confirmField();
      // editingField is cleared synchronously before the update resolves.
      expect(w.vm.editingField).toBeNull();

      await w.vm.$nextTick();
      await w.vm.$nextTick();

      expect(photo.Title).toBe("Original");
      expect(w.vm.$notify.error).toHaveBeenCalledTimes(1);
    });

    it("keeps the inline edit when update() resolves", async () => {
      const photo = buildInlineEditPhoto();
      photo.update.mockResolvedValueOnce({});

      const w = mountSidebar({
        props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
        global: { stubs: { PMap: true } },
      });

      photo.Title = "Edited";
      w.vm.editingField = "title";

      w.vm.confirmField();

      await w.vm.$nextTick();
      await w.vm.$nextTick();

      expect(photo.Title).toBe("Edited");
      expect(w.vm.$notify.error).not.toHaveBeenCalled();
    });

    it("does not call update() when the field is unchanged", () => {
      const photo = buildInlineEditPhoto();
      photo.wasChanged.mockReturnValue(false);

      const w = mountSidebar({
        props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
        global: { stubs: { PMap: true } },
      });

      w.vm.editingField = "title";
      w.vm.confirmField();

      expect(photo.update).not.toHaveBeenCalled();
      expect(w.vm.$notify.error).not.toHaveBeenCalled();
    });
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

  it("should hide private albums from sidebar cross-links", () => {
    const photoWithPrivateAlbum = {
      ...mockPhoto,
      Albums: [
        { UID: "alb1", Title: "Vacation 2023", Slug: "vacation-2023" },
        { UID: "alb2", Title: "Favorites", Slug: "favorites" },
        { UID: "alb3", Title: "Secret", Slug: "secret", Private: true },
      ],
    };
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: photoWithPrivateAlbum, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(w.vm.albums.map((a) => a.UID)).toEqual(["alb1", "alb2"]);
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
    const w = mountSidebar({
      props: { modelValue: { ...mockModel, Caption: "" }, photo: mockPhoto, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(w.vm.captionHtml).toBe("");
  });

  // Cross-link navigation — label, album, and named-person avatars must open
  // in a new browser tab so the current lightbox edit context is preserved.
  function mountWithMockRouter(resolveHref) {
    return mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, context: contexts.Photos },
      global: {
        stubs: { PMap: true },
        mocks: {
          $router: { resolve: vi.fn().mockReturnValue({ href: resolveHref }) },
        },
      },
    });
  }

  it("should open a new tab with a label filter for label clicks", () => {
    const openSpy = vi.spyOn(window, "open").mockImplementation(() => null);
    const w = mountWithMockRouter("/library/browse?q=label%3Anature");
    w.vm.navigateToLabel({ UID: "lbl1", Name: "Nature", Slug: "nature", CustomSlug: "" });
    expect(w.vm.$router.resolve).toHaveBeenCalledWith({ name: "browse", query: { q: "label:nature" } });
    expect(openSpy).toHaveBeenCalledWith("/library/browse?q=label%3Anature", "_blank", "noopener,noreferrer");
    openSpy.mockRestore();
  });

  it("should prefer CustomSlug for label navigation", () => {
    const openSpy = vi.spyOn(window, "open").mockImplementation(() => null);
    const w = mountWithMockRouter("/library/browse?q=label%3Acustom-landscape");
    w.vm.navigateToLabel({ UID: "lbl2", Name: "Landscape", Slug: "landscape", CustomSlug: "custom-landscape" });
    expect(w.vm.$router.resolve).toHaveBeenCalledWith({ name: "browse", query: { q: "label:custom-landscape" } });
    expect(openSpy).toHaveBeenCalled();
    openSpy.mockRestore();
  });

  it("should open a new tab for album clicks", () => {
    const openSpy = vi.spyOn(window, "open").mockImplementation(() => null);
    const w = mountWithMockRouter("/library/albums/alb1/view");
    w.vm.navigateToAlbum({ UID: "alb1", Title: "Vacation 2023" });
    expect(w.vm.$router.resolve).toHaveBeenCalledWith({ name: "album", params: { album: "alb1", slug: "view" } });
    expect(openSpy).toHaveBeenCalledWith("/library/albums/alb1/view", "_blank", "noopener,noreferrer");
    openSpy.mockRestore();
  });

  it("should open a new tab with a subject filter for person avatars that have SubjUID", () => {
    const openSpy = vi.spyOn(window, "open").mockImplementation(() => null);
    const w = mountWithMockRouter("/library/browse?q=subject%3Asubj1");
    w.vm.navigateToPerson({ UID: "m1", Name: "Jane Doe", SubjUID: "subj1" });
    expect(w.vm.$router.resolve).toHaveBeenCalledWith({ name: "browse", query: { q: "subject:subj1" } });
    expect(openSpy).toHaveBeenCalled();
    openSpy.mockRestore();
  });

  it("should open a new tab with a person filter when only Name is available", () => {
    const openSpy = vi.spyOn(window, "open").mockImplementation(() => null);
    const w = mountWithMockRouter("/library/browse?q=person%3AUnknown%20Person");
    w.vm.navigateToPerson({ UID: "m3", Name: "Unknown Person", SubjUID: "" });
    expect(w.vm.$router.resolve).toHaveBeenCalledWith({ name: "browse", query: { q: "person:Unknown Person" } });
    expect(openSpy).toHaveBeenCalled();
    openSpy.mockRestore();
  });

  it("should not open a tab for a person without name or SubjUID", () => {
    const openSpy = vi.spyOn(window, "open").mockImplementation(() => null);
    const w = mountWithMockRouter("/library/browse");
    w.vm.navigateToPerson({ UID: "m4", Name: "", SubjUID: "" });
    expect(openSpy).not.toHaveBeenCalled();
    openSpy.mockRestore();
  });

  // isEditable
  it("should not be editable without canEdit prop", () => {
    expect(wrapper.vm.isEditable).toBeFalsy();
  });

  it("should be editable when canEdit is true with valid photo", () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(w.vm.isEditable).toBeTruthy();
  });

  // Altitude
  it("should return altitude when photo has Altitude", () => {
    const photo = { ...mockPhoto, Altitude: 340 };
    const w = mountSidebar({
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
    const w = mountSidebar({
      props: { modelValue: mockModel, photo, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(w.vm.labels).toHaveLength(1);
    expect(w.vm.labels[0].Label.Name).toBe("Nature");
  });

  // Inline editing: startEditing / cancelEditing
  it("should set editingField and store original value on startEditing", () => {
    const photo = { ...mockPhoto, Title: "Test Title", Caption: "Test Caption" };
    const w = mountSidebar({
      props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    w.vm.startEditing("title");
    expect(w.vm.editingField).toBe("title");
    expect(w.vm.editOriginal).toBe("Test Title");
  });

  it("should restore original value on cancelEditing", async () => {
    const photo = { ...mockPhoto, Title: "Test Title", wasChanged: vi.fn().mockReturnValue(false) };
    const w = mountSidebar({
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
    const w = mountSidebar({
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
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    const id = 1;

    expect(w.vm.isChipPendingRemoval("labels", id)).toBe(false);
    w.vm.togglePendingChipRemoval("labels", id);
    expect(w.vm.isChipPendingRemoval("labels", id)).toBe(true);
    w.vm.togglePendingChipRemoval("labels", id);
    expect(w.vm.isChipPendingRemoval("labels", id)).toBe(false);
  });

  it("should add and remove pending label additions", () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    w.vm.chipState.labels.additions.push("Sunset");
    expect(w.vm.chipState.labels.additions).toContain("Sunset");

    w.vm.removePendingChipAdd("labels", "Sunset");
    expect(w.vm.chipState.labels.additions).not.toContain("Sunset");
  });

  it("should ignore duplicate pending label additions via onLabelSelected", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.editingField = "labels";
    w.vm.onLabelSelected({ Name: "Sunset", UID: "lbl-new" });
    w.vm.onLabelSelected({ Name: "Sunset", UID: "lbl-new" });
    expect(w.vm.chipState.labels.additions).toHaveLength(1);
  });

  it("should ignore non-object values in onLabelSelected", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.onLabelSelected("string-value");
    w.vm.onLabelSelected(null);
    expect(w.vm.chipState.labels.additions).toHaveLength(0);
  });

  it("should skip labels already on the photo in onLabelSelected", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.editingField = "labels";
    w.vm.onLabelSelected({ Name: "Nature", UID: "lbl1" });
    expect(w.vm.chipState.labels.additions).toHaveLength(0);
  });

  // Label validation parity with batch edit + labels tab.
  it("should dedupe pending label additions case-insensitively in onLabelSelected", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.editingField = "labels";
    w.vm.onLabelSelected({ Name: "cat" });
    w.vm.onLabelSelected({ Name: "CAT" });
    expect(w.vm.chipState.labels.additions).toEqual(["cat"]);
  });

  it("should skip labels already on the photo case-insensitively in onLabelSelected", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.editingField = "labels";
    w.vm.onLabelSelected({ Name: "nature" });
    expect(w.vm.chipState.labels.additions).toHaveLength(0);
  });

  it("should dedupe pending label additions case-insensitively in onLabelEnter", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.editingField = "labels";
    w.vm.chipState.labels.additions.push("cat");
    w.vm.chipSearch = "CAT";
    w.vm.onLabelEnter();
    expect(w.vm.chipState.labels.additions).toEqual(["cat"]);
  });

  it("should trim whitespace in onLabelEnter", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.editingField = "labels";
    w.vm.chipSearch = "  dog  ";
    w.vm.onLabelEnter();
    expect(w.vm.chipState.labels.additions).toEqual(["dog"]);
  });

  it("should silently reject empty or whitespace-only label input in onLabelEnter", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.editingField = "labels";
    w.vm.chipSearch = "   ";
    w.vm.onLabelEnter();
    expect(w.vm.chipState.labels.additions).toHaveLength(0);
    expect(w.vm.$notify.error).not.toHaveBeenCalled();
  });

  it("should reject labels longer than the configured clip length and notify", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.editingField = "labels";
    w.vm.chipSearch = "a".repeat(CLIP_LEN + 10);
    w.vm.onLabelEnter();
    expect(w.vm.chipState.labels.additions).toHaveLength(0);
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
    expect(w.vm.chipState.labels.additions).toHaveLength(0);
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
    expect(w.vm.chipState.labels.additions).toHaveLength(0);
  });

  it("should silently reject punctuation-only label input", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.editingField = "labels";
    w.vm.chipSearch = "!!!";
    w.vm.onLabelEnter();
    expect(w.vm.chipState.labels.additions).toHaveLength(0);
    expect(w.vm.$notify.error).not.toHaveBeenCalled();
  });

  it("should accept emoji-only label input", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.editingField = "labels";
    w.vm.chipSearch = "🌅";
    w.vm.onLabelEnter();
    expect(w.vm.chipState.labels.additions).toEqual(["🌅"]);
  });

  // Pending album operations
  it("should toggle album pending removal", () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    const uid = "alb1";

    expect(w.vm.isChipPendingRemoval("albums", uid)).toBe(false);
    w.vm.togglePendingChipRemoval("albums", uid);
    expect(w.vm.isChipPendingRemoval("albums", uid)).toBe(true);
    w.vm.togglePendingChipRemoval("albums", uid);
    expect(w.vm.isChipPendingRemoval("albums", uid)).toBe(false);
  });

  it("should add and remove pending album additions", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    const album = { UID: "alb-new", Title: "New Album" };

    w.vm.onAlbumSelected(album);
    expect(w.vm.chipState.albums.additions).toHaveLength(1);
    expect(w.vm.chipState.albums.additions[0].UID).toBe("alb-new");

    w.vm.removePendingChipAdd("albums", album.UID);
    expect(w.vm.chipState.albums.additions).toHaveLength(0);
  });

  it("should ignore non-object values in onAlbumSelected", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.onAlbumSelected("string-value");
    w.vm.onAlbumSelected(null);
    expect(w.vm.chipState.albums.additions).toHaveLength(0);
  });

  it("should skip albums already on the photo in onAlbumSelected", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.onAlbumSelected({ UID: "alb1", Title: "Vacation 2023" });
    expect(w.vm.chipState.albums.additions).toHaveLength(0);
  });

  // Album validation parity with batch edit + labels tab.
  it("should dedupe albums by normalized title even when UIDs differ", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.onAlbumSelected({ UID: "alb-other", Title: "vacation 2023" });
    expect(w.vm.chipState.albums.additions).toHaveLength(0);
  });

  it("should dedupe pending album additions by normalized title", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.chipState.albums.additions.push({ UID: "alb-a", Title: "Trip" });
    w.vm.onAlbumSelected({ UID: "alb-b", Title: "trip" });
    expect(w.vm.chipState.albums.additions).toHaveLength(1);
  });

  it("should reject overlong album titles in onAlbumEnter and not call save", () => {
    const saveSpy = vi.spyOn(Album.prototype, "save").mockResolvedValue();
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.editingField = "albums";
    w.vm.chipSearch = "a".repeat(CLIP_LEN + 10);
    w.vm.onAlbumEnter();
    expect(saveSpy).not.toHaveBeenCalled();
    expect(w.vm.chipState.albums.additions).toHaveLength(0);
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
    expect(w.vm.chipState.albums.additions).toHaveLength(0);
    saveSpy.mockRestore();
  });

  it("should skip onAlbumEnter when title matches existing album case-insensitively", () => {
    const saveSpy = vi.spyOn(Album.prototype, "save").mockResolvedValue();
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.editingField = "albums";
    w.vm.chipSearch = "VACATION 2023";
    w.vm.onAlbumEnter();
    expect(saveSpy).not.toHaveBeenCalled();
    expect(w.vm.chipState.albums.additions).toHaveLength(0);
    saveSpy.mockRestore();
  });

  it("should skip onAlbumEnter when title matches a pending addition case-insensitively", () => {
    const saveSpy = vi.spyOn(Album.prototype, "save").mockResolvedValue();
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.editingField = "albums";
    w.vm.chipState.albums.additions.push({ UID: "alb-pending", Title: "Trip" });
    w.vm.chipSearch = "trip";
    w.vm.onAlbumEnter();
    expect(saveSpy).not.toHaveBeenCalled();
    expect(w.vm.chipState.albums.additions).toHaveLength(1);
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
    expect(w.vm.chipState.albums.additions).toHaveLength(1);
    expect(w.vm.chipState.albums.additions[0].Title).toBe("Brand New Trip");
    expect(w.vm.albumOptions.some((a) => a.UID === "alb-created")).toBe(true);
    saveSpy.mockRestore();
  });

  // cancelEditing clears all pending state
  it("should clear all pending state on cancelEditing", () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    w.vm.editingField = "labels";
    w.vm.chipState.labels.removals = [1];
    w.vm.chipState.labels.additions = ["Sunset"];
    w.vm.chipState.albums.removals = ["alb1"];
    w.vm.chipState.albums.additions = [{ UID: "alb-new", Title: "New" }];

    w.vm._editStartedAt = Date.now() - 300;
    w.vm.cancelEditing();

    expect(w.vm.editingField).toBeNull();
    expect(w.vm.chipState.labels.removals).toHaveLength(0);
    expect(w.vm.chipState.labels.additions).toHaveLength(0);
    expect(w.vm.chipState.albums.removals).toHaveLength(0);
    expect(w.vm.chipState.albums.additions).toHaveLength(0);
  });

  // Photo watcher: the parent lightbox owns the unsaved-changes guard, so
  // the sidebar no longer silently cancels inline edits when photo changes.
  it("should preserve inline edit state across photo changes (parent guards navigation)", async () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    w.vm.editingField = "title";
    w.vm._editStartedAt = Date.now() - 300;

    await w.setProps({ photo: { ...mockPhoto, Title: "Other" } });
    expect(w.vm.editingField).toBe("title");
  });

  // clearChipInput
  it("should reset chip state on clearChipInput", () => {
    const w = mountSidebar({
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

  describe("restricted-role view", () => {
    const mountRestricted = (isRestricted) =>
      mountSidebar({
        props: {
          modelValue: mockModel,
          photo: mockPhoto,
          canEdit: true,
          context: contexts.Photos,
        },
        global: {
          stubs: { PMap: true },
          mocks: {
            $session: {
              isSidebarRestricted: () => isRestricted,
            },
          },
        },
      });

    it("renders permitted fields for restricted sessions", () => {
      const w = mountRestricted(true);
      const html = w.html();

      expect(w.vm.restrictedRole).toBe(true);
      expect(html).toContain("Test Title");
      expect(html).toContain("Test Caption");
      expect(html).toContain("JPEG, 1920 × 1080, 4.2 MB");
      expect(html).toContain("52.5200, 13.4050");
    });

    it("hides every restricted sidebar section for restricted sessions", () => {
      const w = mountRestricted(true);
      const html = w.html();

      expect(w.find(".metadata__file").exists()).toBe(false);
      expect(html).not.toContain("photos/2023/IMG_001.jpg");

      expect(html).not.toContain("Canon EOS R5");
      expect(html).not.toContain("RF 50mm F1.2L");
      expect(html).not.toContain("mdi-camera");
      expect(html).not.toContain("mdi-camera-iris");

      // Place name; coordinates + map remain visible.
      expect(html).not.toContain("Berlin, Germany");

      expect(html).not.toContain(">People<");
      expect(html).not.toContain(">Labels<");
      expect(html).not.toContain(">Albums<");
      expect(html).not.toContain(">Keywords<");
      expect(html).not.toContain(">Notes<");

      expect(html).not.toContain("Jane Doe");
      expect(html).not.toContain("Nature");
      expect(html).not.toContain("Vacation 2023");
      expect(html).not.toContain("Mountains"); // Subject
      expect(html).not.toContain("John Photographer"); // Artist
      expect(html).not.toContain("2023 John"); // Copyright
      expect(html).not.toContain("CC BY 4.0"); // License
      expect(html).not.toContain("Some notes about this photo");
      expect(html).not.toContain("nature, mountains, sunset"); // Keywords

      expect(w.vm.isEditable).toBe(false);
      expect(w.find(".meta-inline-pencil").exists()).toBe(false);
    });

    it("shows the full sidebar when the session is not restricted", () => {
      const w = mountRestricted(false);
      const html = w.html();

      expect(w.vm.restrictedRole).toBe(false);
      expect(html).toContain("Canon EOS R5");
      expect(html).toContain("RF 50mm F1.2L");
      expect(html).toContain("photos/2023/IMG_001.jpg");
      expect(html).toContain("Berlin, Germany");
    });

    it("falls back to model.getTypeInfo for fileInfo when photo is null", () => {
      const thumbModel = {
        ...mockModel,
        getTypeInfo: vi.fn().mockReturnValue("JPEG\u20034.0MP\u20034032\u2009\u00d7\u20093024"),
      };
      const w = mountSidebar({
        props: {
          modelValue: thumbModel,
          photo: null,
          canEdit: false,
          context: contexts.Photos,
        },
        global: {
          stubs: { PMap: true },
          mocks: {
            $session: { isSidebarRestricted: () => true },
          },
        },
      });
      expect(w.vm.fileInfo).toBe("JPEG\u20034.0MP\u20034032\u2009\u00d7\u20093024");
      expect(thumbModel.getTypeInfo).toHaveBeenCalled();
    });
  });

  // Exhaustive matrix: for every role (admin, user, guest, visitor,
  // contributor, and anonymous share-link sessions) assert which fields
  // the sidebar renders and which edit affordances it exposes, against
  // both a fully-populated photo and a photo without any metadata.
  //
  // The matrix mirrors SidebarRestrictedRoles in
  // frontend/src/model/user.js and the isSidebarRestricted() contract
  // in frontend/src/common/session.js. Role restriction is mocked
  // directly on $session because the component reads it from
  // $session.isSidebarRestricted(); the real role-to-restriction
  // mapping is covered by the User/Session model unit tests.
  describe("role x field visibility matrix", () => {
    // Text fragments the tests look for. Keeping them here so a single
    // change to the fixture (e.g. a label rename) stays localized.
    const TEXT = {
      title: "Matrix Title",
      caption: "Matrix Caption",
      filename: "photos/matrix/IMG_9001.jpg",
      camera: "Canon EOS R5",
      lens: "RF 50mm F1.2L",
      placeName: "Berlin, Germany",
      altitude: "128 m",
      peopleHeader: ">People<",
      labelsHeader: ">Labels<",
      albumsHeader: ">Albums<",
      keywordsHeader: ">Keywords<",
      notesHeader: ">Notes<",
      namedMarker: "Jane Doe",
      labelName: "Nature",
      albumTitle: "Vacation 2024",
      keywords: "sunset, mountains",
      notes: "A short note on this photo",
      subject: "Mountains",
      artist: "John Photographer",
      copyright: "2024 John",
      license: "CC BY 4.0",
    };

    // Build a model with just enough for the sidebar header to render.
    function buildModel({ withMetadata }) {
      const base = {
        UID: "matrix-photo",
        TakenAtLocal: "2024-05-01T10:00:00Z",
        TimeZone: "UTC",
        getLatLng: vi.fn().mockReturnValue("52.5200, 13.4050"),
        copyLatLng: vi.fn(),
      };
      if (withMetadata) {
        return {
          ...base,
          Title: TEXT.title,
          Caption: TEXT.caption,
          Lat: 52.52,
          Lng: 13.405,
          Altitude: 128,
        };
      }
      // Empty-metadata photo: no title, no caption, no coordinates.
      return { ...base, Title: "", Caption: "", Lat: 0, Lng: 0 };
    }

    function buildPhoto({ withMetadata }) {
      if (!withMetadata) {
        return {
          Type: "image",
          // No CameraID, no Lens, no markers, no labels, no albums, no
          // Details; the component must suppress the corresponding rows.
          getCameraInfo: vi.fn().mockReturnValue(""),
          getLensInfo: vi.fn().mockReturnValue(""),
          getImageInfo: vi.fn().mockReturnValue(""),
          getVideoInfo: vi.fn().mockReturnValue(""),
          getVectorInfo: vi.fn().mockReturnValue(""),
          getExifInfo: vi.fn().mockReturnValue(""),
          locationInfo: vi.fn().mockReturnValue(""),
          getMarkers: vi.fn().mockReturnValue([]),
          Labels: [],
          Albums: [],
          Details: null,
          FileName: "",
        };
      }
      return {
        Type: "image",
        Lat: 52.52,
        Lng: 13.405,
        Altitude: 128,
        CameraID: 2,
        CameraMake: "Canon",
        CameraModel: "EOS R5",
        LensID: 2,
        LensMake: "Canon",
        LensModel: "RF 50mm F1.2L",
        Iso: 400,
        Exposure: "1/125",
        FNumber: 1.2,
        FocalLength: 50,
        getCameraInfo: vi.fn().mockReturnValue(TEXT.camera),
        getLensInfo: vi.fn().mockReturnValue(TEXT.lens),
        getImageInfo: vi.fn().mockReturnValue("JPEG, 1920 x 1080, 4.2 MB"),
        getVideoInfo: vi.fn().mockReturnValue(""),
        getVectorInfo: vi.fn().mockReturnValue(""),
        getExifInfo: vi.fn().mockReturnValue("50mm \u2022 f/1.2 \u2022 ISO 400 \u2022 1/125"),
        locationInfo: vi.fn().mockReturnValue(TEXT.placeName),
        getMarkers: vi.fn().mockReturnValue([
          { UID: "m1", CropID: "crop1", Name: TEXT.namedMarker, SubjUID: "subj1", thumbnailUrl: () => "/t/thumb1/public/tile_160" },
          { UID: "m2", CropID: "crop2", Name: "", SubjUID: "", thumbnailUrl: () => "/svg/portrait" },
        ]),
        Labels: [
          { Uncertainty: 0, Label: { ID: 1, UID: "lbl1", Name: TEXT.labelName, Slug: "nature", CustomSlug: "" } },
          // One high-uncertainty label the component must suppress.
          { Uncertainty: 100, Label: { ID: 9, UID: "lbl9", Name: "HiddenLabel", Slug: "hidden", CustomSlug: "" } },
        ],
        Albums: [{ UID: "alb1", Title: TEXT.albumTitle, Slug: "vacation-2024" }],
        Details: {
          Keywords: TEXT.keywords,
          Notes: TEXT.notes,
          Subject: TEXT.subject,
          Artist: TEXT.artist,
          Copyright: TEXT.copyright,
          License: TEXT.license,
        },
        FileName: TEXT.filename,
      };
    }

    function mountFor({ anonymous, restricted, editable }, shape) {
      return mountSidebar({
        props: {
          modelValue: buildModel(shape),
          photo: buildPhoto(shape),
          canEdit: editable,
          context: contexts.Photos,
        },
        global: {
          stubs: { PMap: true },
          mocks: {
            $session: {
              isAnonymous: () => !!anonymous,
              isSidebarRestricted: () => !!restricted,
            },
          },
        },
      });
    }

    // Role-string mapping: the component reads the contract through
    // $session.isSidebarRestricted(), so the visibility matrix below
    // only needs the two equivalence classes. The concrete role list
    // lives in `frontend/src/model/user.js` and is covered by the
    // user/session model's own tests; this guard catches regressions
    // where a new restricted role is added to the product but not to
    // the documented list.
    it("keeps the restricted-role contract in sync with SidebarRestrictedRoles", async () => {
      const mod = await import("model/user");
      expect(mod.SidebarRestrictedRoles).toEqual(expect.arrayContaining(["guest", "visitor", "contributor"]));
    });

    describe("editable session (admin/user)", () => {
      let w;
      beforeEach(() => {
        w = mountFor({ anonymous: false, restricted: false, editable: true }, { withMetadata: true });
      });

      it("renders every metadata section and its content", () => {
        expect(w.vm.restrictedRole).toBe(false);
        expect(w.vm.isEditable).toBe(true);
        const html = w.html();
        for (const needle of [
          TEXT.title,
          TEXT.caption,
          "JPEG, 1920 x 1080, 4.2 MB",
          "52.5200, 13.4050",
          TEXT.filename,
          TEXT.camera,
          TEXT.lens,
          TEXT.placeName,
          TEXT.peopleHeader,
          TEXT.labelsHeader,
          TEXT.albumsHeader,
          TEXT.keywordsHeader,
          TEXT.notesHeader,
          TEXT.namedMarker,
          TEXT.labelName,
          TEXT.albumTitle,
          TEXT.subject,
          TEXT.artist,
          TEXT.copyright,
          TEXT.license,
          TEXT.keywords,
          TEXT.notes,
        ]) {
          expect(html).toContain(needle);
        }
        expect(w.find(".metadata__file").exists()).toBe(true);
      });

      it("renders pencil icons and face-marker controls", () => {
        expect(w.findAll(".meta-inline-pencil").length).toBeGreaterThanOrEqual(10);
        expect(w.find(".meta-markers-toggle").exists()).toBe(true);
        expect(w.find(".meta-marker-add").exists()).toBe(true);
      });
    });

    describe("restricted session (guest/visitor/contributor/share-link)", () => {
      let w;
      beforeEach(() => {
        w = mountFor({ anonymous: false, restricted: true, editable: false }, { withMetadata: true });
      });

      it("renders only the shared fields and hides every restricted section", () => {
        expect(w.vm.restrictedRole).toBe(true);
        expect(w.vm.isEditable).toBe(false);
        const html = w.html();
        // Allow-list.
        expect(html).toContain(TEXT.title);
        expect(html).toContain(TEXT.caption);
        expect(html).toContain("JPEG, 1920 x 1080, 4.2 MB");
        expect(html).toContain("52.5200, 13.4050");
        // Deny-list.
        expect(w.find(".metadata__file").exists()).toBe(false);
        for (const needle of [
          TEXT.filename,
          TEXT.camera,
          TEXT.lens,
          TEXT.placeName,
          TEXT.altitude,
          TEXT.peopleHeader,
          TEXT.labelsHeader,
          TEXT.albumsHeader,
          TEXT.keywordsHeader,
          TEXT.notesHeader,
          TEXT.namedMarker,
          TEXT.labelName,
          TEXT.albumTitle,
          TEXT.subject,
          TEXT.artist,
          TEXT.copyright,
          TEXT.license,
          TEXT.keywords,
          TEXT.notes,
        ]) {
          expect(html).not.toContain(needle);
        }
      });

      it("exposes no edit affordances or face-marker controls", () => {
        expect(w.find(".meta-inline-pencil").exists()).toBe(false);
        expect(w.find(".meta-inline-edit").exists()).toBe(false);
        expect(w.find(".meta-add-prompt").exists()).toBe(false);
        expect(w.find(".meta-markers-toggle").exists()).toBe(false);
        expect(w.find(".meta-marker-add").exists()).toBe(false);
        expect(w.find(".meta-marker-remove").exists()).toBe(false);
      });
    });

    describe("empty metadata", () => {
      it("editable session still suppresses rows whose values are empty", () => {
        const w = mountFor({ anonymous: false, restricted: false, editable: true }, { withMetadata: false });
        // Details is null for this fixture, which forces isEditable to
        // be falsy even for an otherwise-editable session.
        expect(w.vm.isEditable).toBeFalsy();
        const html = w.html();
        for (const needle of [TEXT.camera, TEXT.lens, TEXT.placeName, TEXT.peopleHeader, TEXT.labelsHeader, TEXT.keywordsHeader, TEXT.notesHeader]) {
          expect(html).not.toContain(needle);
        }
        expect(w.find(".meta-inline-pencil").exists()).toBe(false);
      });

      it("restricted session renders the minimal shell only", () => {
        const w = mountFor({ anonymous: false, restricted: true, editable: false }, { withMetadata: false });
        expect(w.vm.restrictedRole).toBe(true);
        expect(w.find(".meta-inline-pencil").exists()).toBe(false);
        expect(w.find(".meta-markers-toggle").exists()).toBe(false);
      });
    });

    // Parent-driven editor (canEdit=true, Details present, not
    // restricted, empty fields) must expose "add prompt" affordances
    // so admins can populate metadata from scratch.
    it("shows add-prompt affordances to admins editing an empty Details object", () => {
      const w = mountSidebar({
        props: {
          modelValue: buildModel({ withMetadata: false }),
          photo: {
            ...buildPhoto({ withMetadata: false }),
            // A fresh, empty Details object is enough to unblock isEditable.
            Details: { Keywords: "", Notes: "", Subject: "", Artist: "", Copyright: "", License: "" },
          },
          canEdit: true,
          context: contexts.Photos,
        },
        global: {
          stubs: { PMap: true },
          mocks: {
            $session: { isAnonymous: () => false, isSidebarRestricted: () => false },
          },
        },
      });
      expect(w.vm.isEditable).toBe(true);
      // Add-prompt spans are the "click to start editing" placeholders.
      const prompts = w.findAll(".meta-add-prompt");
      expect(prompts.length).toBeGreaterThanOrEqual(5);
      // At least title, caption, keywords, notes, subject are all expected.
      const texts = prompts.map((p) => p.text());
      expect(texts).toContain("Title");
      expect(texts).toContain("Caption");
      expect(texts).toContain("Keywords");
      expect(texts).toContain("Notes");
      expect(texts).toContain("Subject");
    });

    // Explicit share-link (anonymous) case: the sidebar must behave
    // identically to a restricted-role session even though the backing
    // User record is empty.
    it("treats anonymous share-link sessions as restricted even with canEdit=true", () => {
      const w = mountSidebar({
        props: {
          modelValue: buildModel({ withMetadata: true }),
          photo: buildPhoto({ withMetadata: true }),
          // Simulate a lightbox that forgot to flip canEdit off: the
          // sidebar must still drop all edit affordances, driven by
          // restrictedRole alone.
          canEdit: true,
          context: contexts.Photos,
        },
        global: {
          stubs: { PMap: true },
          mocks: {
            $session: { isAnonymous: () => true, isSidebarRestricted: () => true },
          },
        },
      });
      expect(w.vm.restrictedRole).toBe(true);
      expect(w.vm.isEditable).toBe(false);
      expect(w.find(".meta-inline-pencil").exists()).toBe(false);
      expect(w.find(".meta-markers-toggle").exists()).toBe(false);
      expect(w.find(".meta-marker-add").exists()).toBe(false);
    });

    // featPeople = false must hide the People section even for an
    // admin on a photo that has markers.
    it("hides the People section when featPeople is disabled, regardless of role", () => {
      const w = mountSidebar({
        props: {
          modelValue: buildModel({ withMetadata: true }),
          photo: buildPhoto({ withMetadata: true }),
          canEdit: true,
          context: contexts.Photos,
        },
        global: {
          stubs: { PMap: true },
          mocks: {
            $config: { ...validationConfig, feature: (k) => k !== "people" },
            $session: { isAnonymous: () => false, isSidebarRestricted: () => false },
          },
        },
      });
      const html = w.html();
      expect(html).not.toContain(TEXT.peopleHeader);
      expect(w.find(".meta-markers-toggle").exists()).toBe(false);
      expect(w.find(".meta-marker-add").exists()).toBe(false);
    });
  });

  // Pins the template bindings on <p-date-time-dialog>, <p-camera-dialog>,
  // and <p-location-dialog>. Commit 7a611b6d6 removed the legacy `photo`
  // prop on PSidebarInfo and routed parent state through `$view.getData()`
  // (exposed via the `photo` computed). The three dialog tags in the sidebar
  // template kept referencing the now-undefined `photo` identifier, so the
  // dialogs received `:photo="undefined"` / `:latlng="[0, 0]"` and their
  // loadFromPhoto() hooks bailed out — opening every sidebar dialog with
  // empty inputs even when the photo had values. These tests fail if the
  // bindings ever regress to a name that does not exist on the component.
  describe("dialog photo prop bindings", () => {
    const dialogStubs = {
      PDateTimeDialog: { name: "PDateTimeDialog", template: "<div />", props: ["visible", "photo"] },
      PCameraDialog: { name: "PCameraDialog", template: "<div />", props: ["visible", "photo"] },
      PLocationDialog: { name: "PLocationDialog", template: "<div />", props: ["visible", "latlng"] },
    };

    function mountWithDialogStubs(photo) {
      return mountSidebar({
        props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
        global: { stubs: { PMap: true, ...dialogStubs } },
      });
    }

    it("passes the current photo to <p-date-time-dialog>", () => {
      const w = mountWithDialogStubs(mockPhoto);
      const dialog = w.findComponent({ name: "PDateTimeDialog" });
      expect(dialog.exists()).toBe(true);
      // Vue 3 wraps `view` (data()) in a reactive proxy, so the dialog
      // sees a proxy of mockPhoto rather than the raw reference. A
      // deep-equality check still pins the regression: the broken
      // binding produced `undefined`, not a structurally-equal proxy.
      expect(dialog.props("photo")).toEqual(mockPhoto);
    });

    it("passes the current photo to <p-camera-dialog>", () => {
      const w = mountWithDialogStubs(mockPhoto);
      const dialog = w.findComponent({ name: "PCameraDialog" });
      expect(dialog.exists()).toBe(true);
      expect(dialog.props("photo")).toEqual(mockPhoto);
    });

    it("derives <p-location-dialog> latlng from the current photo", () => {
      const photo = { ...mockPhoto, Lat: 52.52, Lng: 13.405 };
      const w = mountWithDialogStubs(photo);
      const dialog = w.findComponent({ name: "PLocationDialog" });
      expect(dialog.exists()).toBe(true);
      expect(dialog.props("latlng")).toEqual([52.52, 13.405]);
    });

    it("falls back to [0, 0] for <p-location-dialog> when the photo has no coordinates", () => {
      const w = mountWithDialogStubs(mockPhoto);
      const dialog = w.findComponent({ name: "PLocationDialog" });
      expect(dialog.exists()).toBe(true);
      expect(dialog.props("latlng")).toEqual([0, 0]);
    });
  });
});
