import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount } from "@vue/test-utils";
import PLightboxSidebar from "component/lightbox/sidebar.vue";
import * as contexts from "options/contexts";
import { DateTime } from "luxon";
import $util from "common/util";
import { Album } from "model/album";
import typeaheadCache from "common/typeahead-cache";
import { $faceMarkers } from "common/face-markers";

// Max name length used by the validation pipeline (matches the production
// "clip" client-config value). Override the global $config.get mock so the
// length-check branch can be exercised in tests.
const CLIP_LEN = 160;
const validationConfig = {
  feature: () => true,
  get: (key) => (key === "clip" ? CLIP_LEN : false),
  getSettings: () => ({ features: { edit: true, favorites: true, download: true, archive: true } }),
  allow: () => true,
  deny: () => false,
  featExperimental: () => false,
  featDevelop: () => false,
  values: {},
  dir: () => "ltr",
};
// Mounted with the real $util.normalizeTitle so the validation pipeline
// runs against the same normalization the component uses at runtime.
// Other $util methods needed at render time are stubbed inline.
const validationUtil = {
  normalizeTitle: (s) => $util.normalizeTitle(s),
  formatCamera: (camera, id, make, model, long) => $util.formatCamera(camera, id, make, model, long),
  typeName: (type, defaultValue) => $util.typeName(type, defaultValue),
  encodeHTML: (s) => s,
  sanitizeHtml: (s) => s,
  copyText: vi.fn(),
  hasTouch: () => false,
  formatSeconds: (n) => String(n),
  formatRemainingSeconds: () => "0",
  videoFormat: () => "avc",
  videoFormatUrl: () => "/v.mp4",
  thumb: () => ({ src: "/t.jpg", w: 100, h: 100 }),
};
// mountSidebar accepts the same option shape that the suite used to pass
// directly to `mount(PLightboxSidebar, ...)`. PLightboxSidebar no longer takes the
// legacy props (modelValue/photo/canEdit/context/markersVisible/...); they
// live on the parent lightbox's $data and are surfaced through $view.getData()
// (see lightbox.vue). This helper extracts the legacy keys from `options.props`
// and exposes them as a $view mock so existing tests keep working with only
// `mount(PLightboxSidebar, ` -> `mountSidebar(` renamed.
function mountSidebar(options = {}) {
  const props = { ...(options.props || {}) };
  // Translate the legacy boolean props (`markersVisible` / `markersEdit`)
  // into the new state-machine value (F2-1: `faceMarkerMode`). Tests
  // pre-dating the enum still set the booleans; new tests can pass
  // `faceMarkerMode` directly.
  let faceMarkerMode = props.faceMarkerMode;
  if (faceMarkerMode === undefined) {
    if (props.markersEdit) {
      faceMarkerMode = "edit";
    } else if (props.markersVisible) {
      faceMarkerMode = "display";
    } else {
      faceMarkerMode = null;
    }
  }
  const legacy = {
    model: props.modelValue,
    photo: props.photo,
    canEdit: props.canEdit,
    contextAllowsEdit: true,
    collection: props.collection,
    context: props.context,
  };

  // Bridge legacy props that used to live on the parent lightbox $data
  // into the shared face-markers singleton. The sidebar's computeds now
  // read directly from `$faceMarkers` (mode / busy / pendingNameMarkerUid)
  // so individual tests can still drive them via the friendly props
  // (`markersVisible`, `markersEdit`, `markersBusy`, `newMarkerUid`).
  // Each `mountSidebar()` call is treated as a fresh test scope.
  $faceMarkers.reset();
  if (faceMarkerMode) {
    $faceMarkers.setMode(faceMarkerMode);
  }
  if (props.markersBusy) {
    $faceMarkers.setBusy(true);
  }
  if (props.newMarkerUid) {
    $faceMarkers.setPendingNameMarkerUid(props.newMarkerUid);
  }

  // Strip legacy keys so Vue does not warn about unknown props.
  delete props.modelValue;
  delete props.photo;
  delete props.canEdit;
  delete props.collection;
  delete props.context;
  delete props.markersVisible;
  delete props.markersEdit;
  delete props.faceMarkerMode;
  delete props.markersBusy;
  delete props.newMarkerUid;

  if (!props.uid && legacy.model && legacy.model.UID) {
    props.uid = legacy.model.UID;
  }

  const global = options.global || {};
  const mocks = global.mocks || {};
  const stubs = global.stubs || {};

  return mount(PLightboxSidebar, {
    ...options,
    props,
    global: {
      ...global,
      stubs: { PMap: true, ...stubs },
      mocks: {
        $view: { getData: () => legacy, enter: () => {}, leave: () => {}, isActive: () => true },
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

describe("PLightboxSidebar component", () => {
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
      getLatLng: vi.fn().mockReturnValue("52.5200°N 13.4050°E"),
      getLatLngShort: vi.fn().mockReturnValue("52.5200°N 13.4050°E"),
      copyLatLng: vi.fn(),
    };

    mockPhoto = {
      UID: "ps6sg6be2lvl0y12",
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
      PlaceLabel: "Berlin, Germany",
      Country: "de",
      placeName: vi.fn().mockReturnValue("Berlin, Germany"),
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
      // Album/label membership writes — the sidebar's instant-save additions
      // path (addLabelImmediate / addAlbumImmediate) and the batched removal
      // path (confirmLabels / confirmAlbums) all route through these.
      addLabel: vi.fn().mockResolvedValue({}),
      removeLabel: vi.fn().mockResolvedValue({}),
      addToAlbum: vi.fn().mockResolvedValue({}),
      removeFromAlbum: vi.fn().mockResolvedValue({}),
    };
  }

  beforeEach(() => {
    // typeaheadCache is module-scope; clear it before each test so a
    // cached labels/albums list from a previous test doesn't bleed
    // into the next one's network-spy assertions.
    typeaheadCache.clear();
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
    expect(wrapper.find(".p-lightbox-sidebar").exists()).toBe(true);

    const html = wrapper.html();
    expect(html).toContain("Test Title");
    expect(html).toContain("Test Caption");
    expect(html).toContain("photos/2023/IMG_001.jpg");
    expect(html).toContain("JPEG, 1920 × 1080, 4.2 MB");

    expect(mockPhoto.getImageInfo).toHaveBeenCalled();
    expect(mockModel.getLatLngShort).toHaveBeenCalled();
  });

  it("should not render an inline pencil next to the file row", () => {
    const fileRow = wrapper.find(".meta-file");
    expect(fileRow.exists()).toBe(true);
    // The file row is intentionally non-editable: it merges the type/
    // size title with the filename subtitle and uses click-to-copy
    // (which doesn't need a pencil affordance). The mdi-* prepend icon
    // is intentional and decorative.
    expect(fileRow.find(".meta-inline-pencil").exists()).toBe(false);
    expect(fileRow.text()).toContain(mockPhoto.FileName);
  });

  it("should render file info row with the file metadata text", () => {
    const html = wrapper.html();
    expect(html).toContain("JPEG, 1920 × 1080, 4.2 MB");
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

  it("should trigger copyLatLng when the coordinates row is clicked", async () => {
    const coordinatesRow = wrapper.find(".meta-coordinates");
    expect(coordinatesRow.exists()).toBe(true);
    await coordinatesRow.trigger("click");
    expect(mockModel.copyLatLng).toHaveBeenCalled();
  });

  it("should handle model without taken time", () => {
    const formattedTime = wrapper.vm.formatTime({ ...mockModel, TakenAtLocal: null });
    expect(formattedTime).toBe("Unknown");
  });

  // Regression: formatTime() used to derive Day=1 / Jan 1 for every
  // Year/Month/Day=-1 (Unknown) photo because it formatted the padded
  // TakenAtLocal directly. The Thumb model from viewer.Result carries
  // no Year/Month/Day, so the right fix is to delegate to the full
  // Photo's getDateString() — which already handles the Unknown
  // fallback — whenever the full Photo is loaded.
  it("delegates to photo.getDateString(true) when the full Photo is loaded", () => {
    const getDateString = vi.fn().mockReturnValue("June 2024");
    const photo = { ...mockPhoto, getDateString };
    const w = mountSidebar({
      props: { modelValue: mockModel, photo, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });

    const result = w.vm.formatTime(mockModel);

    expect(getDateString).toHaveBeenCalledWith(true);
    expect(result).toBe("June 2024");
  });

  // Limited-access sessions get an empty Photo placeholder whose default
  // getDateString() renders "Unknown" against empty Year/Month/Day. The UID
  // gate makes formatTime fall through to the Thumb's TakenAtLocal instead.
  it("skips photo.getDateString and uses Thumb.TakenAtLocal for the empty Photo placeholder", () => {
    const getDateString = vi.fn().mockReturnValue("Unknown");
    const emptyPhoto = { UID: "", getDateString, getMarkers: vi.fn().mockReturnValue([]) };
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: emptyPhoto, canEdit: false, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });

    const result = w.vm.formatTime(mockModel);

    expect(getDateString).not.toHaveBeenCalled();
    expect(result).toBe("January 1, 2023, 10:00 AM");
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
    // fileName returns null (not "") so the merged file row's
    // `:subtitle` binding skips rendering an empty subtitle element —
    // Vuetify gates the subtitle on `props.subtitle != null`, so "" would
    // still render an empty slot.
    expect(w.vm.fileName).toBeNull();
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

  // People section: per-role toggle in the header (#append slot of the
  // section row). Editable users see ONLY the pencil/pencil-off "Edit
  // Faces" toggle — draw mode is a strict superset of display mode for
  // them, so a separate eye toggle adds no value. Non-editable users
  // see ONLY the eye/eye-off "Show face markers" toggle and only when
  // there is at least one marker to view. The state machine
  // (`$faceMarkers`) still has both FaceMarkerDisplay and FaceMarkerEdit;
  // each role just exposes one of the two via its button.
  it("should render the People header with only the pencil toggle when editable", () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(w.html()).toContain("People");
    expect(w.find(".meta-faces-edit").exists()).toBe(true);
    // `.meta-markers-toggle` is shared by both toggle variants (the eye
    // for read-only display and the pencil for edit, since edit-mode also
    // toggles visibility). `.meta-faces-display` is the eye-only
    // discriminator that should be absent in editable contexts.
    expect(w.find(".meta-faces-display").exists()).toBe(false);
  });

  it("should render the pencil toggle even when the photo has no face markers (so editable users can add the first one)", () => {
    const photo = { ...mockPhoto, getMarkers: vi.fn().mockReturnValue([]) };
    const w = mountSidebar({
      props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(w.html()).toContain("People");
    expect(w.find(".meta-faces-edit").exists()).toBe(true);
    expect(w.find(".meta-faces-display").exists()).toBe(false);
  });

  it("should render only the eye toggle when not editable, on a photo with face markers", () => {
    // The default wrapper is mounted without canEdit → isEditable false.
    // The eye toggle is the only marker control non-editable users see;
    // the pencil toggle is editable-only.
    expect(wrapper.find(".meta-markers-toggle").exists()).toBe(true);
    expect(wrapper.find(".meta-faces-edit").exists()).toBe(false);
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

  it("should reflect the markersVisible state on the eye toggle (non-editable role)", () => {
    // Eye toggle is non-editable-only; markersVisible = true paints the
    // is-active class so the icon flips to `mdi-eye-off-outline`.
    const w = mountSidebar({
      props: {
        modelValue: mockModel,
        photo: mockPhoto,
        context: contexts.Photos,
        markersVisible: true,
      },
      global: { stubs: { PMap: true } },
    });
    const toggle = w.find(".meta-markers-toggle");
    expect(toggle.exists()).toBe(true);
    expect(toggle.classes()).toContain("is-active");
    expect(toggle.find("i.mdi-eye-off-outline").exists()).toBe(true);
  });

  it("should emit toggle-face-marker-mode when the eye icon is clicked (non-editable role)", async () => {
    const onToggle = vi.fn();
    const w = mountSidebar({
      props: {
        "modelValue": mockModel,
        "photo": mockPhoto,
        "context": contexts.Photos,
        "onToggleFaceMarkerMode": onToggle,
      },
      global: { stubs: { PMap: true } },
    });
    await w.find(".meta-markers-toggle").trigger("click");
    expect(onToggle).toHaveBeenCalled();
  });

  it("should emit toggle-face-marker-edit when the + icon is clicked", async () => {
    const onToggle = vi.fn();
    const w = mountSidebar({
      props: {
        "modelValue": mockModel,
        "photo": mockPhoto,
        "canEdit": true,
        "context": contexts.Photos,
        "onToggleFaceMarkerEdit": onToggle,
      },
      global: { stubs: { PMap: true } },
    });
    const btn = w.find(".meta-faces-edit");
    // After the v-icon → v-btn refactor the mdi-* class lives on the
    // nested <i> inside the button, not on the button itself. The
    // "Edit Faces" toggle uses pencil / pencil-off icons (was +/check
    // when the toggle's role was "Add face"; now it enters edit mode
    // which covers add AND remove via the overlay).
    expect(btn.find("i.mdi-pencil-outline").exists()).toBe(true);
    await btn.trigger("click");
    expect(onToggle).toHaveBeenCalled();
  });

  it("should swap the edit icon to pencil-off while markersEdit is true", () => {
    const w = mountSidebar({
      props: {
        modelValue: mockModel,
        photo: mockPhoto,
        canEdit: true,
        context: contexts.Photos,
        markersEdit: true,
      },
      global: { stubs: { PMap: true } },
    });
    const btn = w.find(".meta-faces-edit");
    expect(btn.classes()).toContain("is-active");
    expect(btn.find("i.mdi-pencil-off-outline").exists()).toBe(true);
  });

  it("should still emit toggle-face-marker-edit when markersEdit is true (so the user can exit)", async () => {
    const onToggle = vi.fn();
    const w = mountSidebar({
      props: {
        "modelValue": mockModel,
        "photo": mockPhoto,
        "canEdit": true,
        "context": contexts.Photos,
        "markersEdit": true,
        "onToggleFaceMarkerEdit": onToggle,
      },
      global: { stubs: { PMap: true } },
    });
    await w.find(".meta-faces-edit").trigger("click");
    expect(onToggle).toHaveBeenCalled();
  });

  // The per-row mdi-close button (formerly .meta-marker-remove inside
  // the combobox's #append-inner slot) was retired: it looked like a
  // "clear input" affordance but rejected the marker on the backend
  // without confirmation. Removal now lives on the face-marker overlay
  // — see tests/vitest/component/meta/face/markers.test.js for
  // the click-to-remove + confirm-pill flow.
  it("should not render a per-row remove icon on any person row", () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    const personRows = w.findAll(".metadata__person-row");
    for (const row of personRows) {
      expect(row.find(".meta-marker-remove").exists()).toBe(false);
    }
  });

  it("should refuse to emit toggle-face-marker-mode / toggle-face-marker-edit while markersBusy is true", () => {
    const onToggleMode = vi.fn();
    const onToggleDraw = vi.fn();
    const w = mountSidebar({
      props: {
        "modelValue": mockModel,
        "photo": mockPhoto,
        "canEdit": true,
        "context": contexts.Photos,
        "markersBusy": true,
        "onToggleFaceMarkerMode": onToggleMode,
        "onToggleFaceMarkerEdit": onToggleDraw,
      },
      global: { stubs: { PMap: true } },
    });
    w.vm.onToggleFaceMarkerMode();
    w.vm.onToggleFaceMarkerEdit();
    expect(onToggleMode).not.toHaveBeenCalled();
    expect(onToggleDraw).not.toHaveBeenCalled();
  });

  it("should render an inline name input for every marker when canEdit is true", () => {
    const w = mountSidebar({
      props: {
        modelValue: mockModel,
        photo: mockPhoto,
        canEdit: true,
        context: contexts.Photos,
        markersVisible: true,
        markersEdit: true,
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
    expect(w.find(".meta-faces-edit").exists()).toBe(false);
    expect(w.find(".metadata__person-row").exists()).toBe(false);
  });

  // Unassign button: named markers expose mdi-eject inside the
  // combobox's #append-inner slot, but only while the People section
  // is in edit mode (pencil on) — gated so a casual hover can't
  // accidentally clear a face's subject assignment.
  it("should render the Unassign icon on a named marker in edit mode and emit clear-subject on click", async () => {
    const onClearSubject = vi.fn();
    const w = mountSidebar({
      props: {
        "modelValue": mockModel,
        "photo": mockPhoto,
        "canEdit": true,
        "context": contexts.Photos,
        "markersEdit": true,
        "onClearSubject": onClearSubject,
      },
      global: { stubs: { PMap: true } },
    });
    const personRows = w.findAll(".metadata__person-row");
    // First row: Jane Doe (SubjUID set).
    const clearSubjectIcon = personRows[0].find(".meta-marker-clear-subject");
    expect(clearSubjectIcon.exists()).toBe(true);
    // Unnamed marker should NOT have an Unassign icon.
    expect(personRows[1].find(".meta-marker-clear-subject").exists()).toBe(false);
    await clearSubjectIcon.trigger("click");
    expect(onClearSubject).toHaveBeenCalledTimes(1);
    expect(onClearSubject.mock.calls[0][0].UID).toBe("m1");
  });

  it("should hide the Unassign icon on a named marker when not in edit mode", () => {
    const w = mountSidebar({
      props: {
        modelValue: mockModel,
        photo: mockPhoto,
        canEdit: true,
        context: contexts.Photos,
      },
      global: { stubs: { PMap: true } },
    });
    const personRows = w.findAll(".metadata__person-row");
    // Even though the first row is named, the Unassign affordance is
    // suppressed outside edit mode.
    expect(personRows[0].find(".meta-marker-clear-subject").exists()).toBe(false);
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

  // The default dropdown chevron (`.v-combobox__menu-icon`) is suppressed
  // unconditionally on every marker combobox via `:menu-icon="null"` —
  // both named (readonly) and unnamed rows. Auto-open-on-focus is the
  // discovery affordance for the knownPeople dropdown; the chevron is
  // redundant with the inline edit affordances (× / ⏏ buttons) and was
  // misleading on named, readonly rows where ejecting is the only path
  // to a new name. Mock $config supplies one known person because
  // Vuetify additionally gates the chevron on `items.length > 0`, and
  // we want a state in which Vuetify *would* render it if it were not
  // suppressed by the prop.
  it("should hide the v-combobox menu-icon on every marker row", () => {
    const config = {
      ...validationConfig,
      values: { people: [{ UID: "sXYZ", Name: "Alice Smith" }] },
    };
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { mocks: { $config: config, $util: validationUtil } },
    });
    const personRows = w.findAll(".metadata__person-row");
    // First row: Jane Doe (SubjUID set) — chevron absent.
    expect(personRows[0].find(".v-combobox__menu-icon").exists()).toBe(false);
    // Second row: unnamed marker — chevron also absent.
    expect(personRows[1].find(".v-combobox__menu-icon").exists()).toBe(false);
  });

  it("should refuse to emit clear-subject on a marker without SubjUID", () => {
    const onClearSubject = vi.fn();
    const w = mountSidebar({
      props: {
        "modelValue": mockModel,
        "photo": mockPhoto,
        "canEdit": true,
        "context": contexts.Photos,
        "onClearSubject": onClearSubject,
      },
      global: { stubs: { PMap: true } },
    });
    w.vm.onClearSubject({ UID: "mX", SubjUID: "", Name: "" });
    expect(onClearSubject).not.toHaveBeenCalled();
  });

  it("should refuse to emit clear-subject while markersBusy is true", () => {
    const onClearSubject = vi.fn();
    const w = mountSidebar({
      props: {
        "modelValue": mockModel,
        "photo": mockPhoto,
        "canEdit": true,
        "context": contexts.Photos,
        "markersBusy": true,
        "onClearSubject": onClearSubject,
      },
      global: { stubs: { PMap: true } },
    });
    w.vm.onClearSubject({ UID: "m1", SubjUID: "subj1", Name: "Jane Doe" });
    expect(onClearSubject).not.toHaveBeenCalled();
  });

  // pendingNameMarkerUid focuses the input on the freshly-created marker and
  // emits naming-started so the parent (lightbox) can clear its own state.
  // The value lives on the shared face-markers singleton; mutating it
  // triggers the sidebar's `newMarkerUid` computed and its watcher.
  it("should focus the matching marker input when pendingNameMarkerUid changes", async () => {
    const onNamingStarted = vi.fn();
    const w = mountSidebar({
      props: {
        "modelValue": mockModel,
        "photo": mockPhoto,
        "canEdit": true,
        "context": contexts.Photos,
        "newMarkerUid": null,
        "onNamingStarted": onNamingStarted,
      },
      global: { stubs: { PMap: true } },
      attachTo: document.body,
    });
    $faceMarkers.setPendingNameMarkerUid("m2");
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
        "onReloadMarkers": onReload,
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
    w.vm.chipState.people.options = [{ UID: "sXYZ", Name: "Alice Smith" }];
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
    w.vm.chipState.people.options = [{ UID: "sBOB", Name: "Bob" }];
    w.vm.onPickPerson(namedMarker, { UID: "sBOB", Name: "Bob" });
    expect(namedMarker.Name).toBe("Bob");
    expect(namedMarker.SubjUID).toBe("sBOB");
    expect(setName).toHaveBeenCalled();
  });

  it("should populate knownPeople from typeaheadCache.getPeople", async () => {
    typeaheadCache.clear();
    const cacheSpy = vi.spyOn(typeaheadCache, "getPeople").mockResolvedValue([
      { UID: "sA", Name: "Alice" },
      { UID: "sB", Name: "Bob" },
    ]);
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    w.vm.loadChipOptions("people");
    await Promise.resolve();
    await Promise.resolve();
    expect(cacheSpy).toHaveBeenCalled();
    expect(w.vm.knownPeople).toHaveLength(2);
    expect(w.vm.knownPeople[0].Name).toBe("Alice");
    cacheSpy.mockRestore();
  });

  it("should start with an empty knownPeople list before the cache loads", () => {
    typeaheadCache.clear();
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(w.vm.knownPeople).toEqual([]);
  });

  // Locks in the focus-before-resolve flow: loadChipOptions (combobox @focus)
  // may run while the people fetch is still in flight; knownPeople stays empty
  // until it resolves, then populates without any synchronous read hazard.
  it("knownPeople stays empty until the people fetch resolves, then populates", async () => {
    typeaheadCache.clear();
    let resolveFn;
    const deferred = new Promise((resolve) => (resolveFn = resolve));
    const spy = vi.spyOn(typeaheadCache, "getPeople").mockReturnValue(deferred);
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    w.vm.chipState.people.options = [];
    w.vm.loadChipOptions("people");
    await Promise.resolve();
    expect(w.vm.knownPeople).toEqual([]);
    resolveFn([{ UID: "sA", Name: "Alice" }]);
    await Promise.resolve();
    await Promise.resolve();
    expect(w.vm.knownPeople).toEqual([{ Name: "Alice", UID: "sA" }]);
    spy.mockRestore();
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
    expect(w.vm.addNameDialog.markerUid).toBe("m2");
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
    w.vm.chipState.people.options = [{ UID: "sALC", Name: "Alice Smith" }];
    w.vm.setMarkerInputValue("m2", "alice smith");
    w.vm.confirmMarkerName(marker, "blur");
    expect(w.vm.addNameDialog.visible).toBe(false);
    expect(setName).toHaveBeenCalled();
    expect(marker.Name).toBe("Alice Smith");
    expect(marker.SubjUID).toBe("sALC");
  });

  // P1-1 — locale-aware case-insensitive match. The previous "en" locale
  // missed matches under Turkish dotted/dotless i and other non-ASCII case
  // folding rules; we now use `undefined` to defer to the active locale.
  // The standard ECMAScript Intl implementation handles Cyrillic case
  // collation under base sensitivity regardless of locale, so we can
  // exercise the path without relying on a specific JS locale.
  it("findKnownPerson resolves Cyrillic case differences via locale-aware match", () => {
    const knownPersonConfig = {
      ...validationConfig,
      values: { people: [{ UID: "sIvan", Name: "Иван Петров" }] },
    };
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: {
        stubs: { PMap: true },
        mocks: { $config: knownPersonConfig, $util: validationUtil },
      },
    });
    w.vm.chipState.people.options = [{ UID: "sIvan", Name: "Иван Петров" }];
    expect(w.vm.findKnownPerson("иван петров")).toEqual({ UID: "sIvan", Name: "Иван Петров" });
    expect(w.vm.findKnownPerson("Unknown Person")).toBeNull();
  });

  // P1-2 — knownPeople is sorted locale-aware in loadChipOptions("people");
  // insertion order from people.created WS events would otherwise put new
  // subjects at the top. Blank/null entries are filtered out.
  it("knownPeople returns a locale-aware sorted list from the people cache", async () => {
    typeaheadCache.clear();
    const cacheSpy = vi.spyOn(typeaheadCache, "getPeople").mockResolvedValue([
      { UID: "sZara", Name: "Zara" },
      { UID: "sAndre", Name: "André" },
      { UID: "sBob", Name: "Bob" },
      { UID: "sAlice", Name: "alice" },
      { UID: "sBlank", Name: "" },
      null,
    ]);
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    w.vm.loadChipOptions("people");
    await Promise.resolve();
    await Promise.resolve();
    const names = w.vm.knownPeople.map((p) => p.Name);
    expect(names).toEqual(["alice", "André", "Bob", "Zara"]);
    cacheSpy.mockRestore();
  });

  // P1-3 — typed-but-uncommitted text must survive a WS-driven sync. The
  // `editing` flag on the draft entry tells syncMarkerDrafts to leave the
  // user's typing alone even when the backend value moves.
  it("syncMarkerDrafts does not clobber typed text while editing is in progress", async () => {
    const marker = {
      UID: "m2",
      Name: "",
      SubjUID: "",
      thumbnailUrl: () => "/t/thumb2/public/tile_160",
      setName: vi.fn(),
    };
    const photo = { ...mockPhoto, getMarkers: vi.fn().mockReturnValue([marker]) };
    const w = mountInfoForChips({ modelValue: mockModel, photo });
    await w.vm.$nextTick();
    w.vm.setMarkerInputValue("m2", "Plane Port");
    expect(w.vm.markerDrafts.m2.editing).toBe(true);
    // Simulate an unrelated WS update that re-fires the people watcher with
    // a renamed backend Name. The user's typing must NOT be overwritten.
    w.vm.syncMarkerDrafts([{ UID: "m2", Name: "Renamed By Worker" }]);
    expect(w.vm.markerInputValue("m2")).toBe("Plane Port");
    // The captured `original` does update — that's the backend's authoritative
    // value the draft will roll back to on cancel.
    expect(w.vm.markerDrafts.m2.original).toBe("Renamed By Worker");
  });

  // P1-3 — once the edit is settled (commit or cancel), the editing flag
  // clears and a subsequent WS update can re-sync the input again.
  it("syncMarkerDrafts re-syncs after the edit settles", async () => {
    const marker = {
      UID: "m2",
      Name: "",
      SubjUID: "",
      thumbnailUrl: () => "/t/thumb2/public/tile_160",
      setName: vi.fn().mockResolvedValue(undefined),
    };
    const photo = { ...mockPhoto, getMarkers: vi.fn().mockReturnValue([marker]) };
    const w = mountInfoForChips({ modelValue: mockModel, photo });
    await w.vm.$nextTick();
    w.vm.setMarkerInputValue("m2", "Plane Port");
    w.vm.cancelMarkerName(marker);
    expect(w.vm.markerDrafts.m2.editing).toBe(false);
    w.vm.syncMarkerDrafts([{ UID: "m2", Name: "Server Side Name" }]);
    expect(w.vm.markerInputValue("m2")).toBe("Server Side Name");
  });

  // P1-6 — confirmMarkerName is a no-op when markersBusy. Prevents a queued
  // rename from racing with an in-flight reject/eject on the same marker.
  it("confirmMarkerName bails when markersBusy is true", async () => {
    const setName = vi.fn().mockResolvedValue(undefined);
    const marker = {
      UID: "m2",
      Name: "",
      SubjUID: "",
      thumbnailUrl: () => "/t/thumb2/public/tile_160",
      setName,
    };
    const photo = { ...mockPhoto, getMarkers: vi.fn().mockReturnValue([marker]) };
    const w = mountSidebar({
      props: { modelValue: mockModel, photo, canEdit: true, markersBusy: true, context: contexts.Photos },
      global: { stubs: { PMap: true }, mocks: { $util: validationUtil } },
    });
    await w.vm.$nextTick();
    w.vm.setMarkerInputValue("m2", "Plane Port");
    w.vm.confirmMarkerName(marker, "enter");
    expect(setName).not.toHaveBeenCalled();
    expect(w.vm.addNameDialog.visible).toBe(false);
  });

  // P1-6 — invalid (rejected) marker is filtered out of getMarkers(true)
  // but the row's @blur can still fire during the unmount tick. Make sure
  // confirmMarkerName refuses to write through to a rejected marker.
  it("confirmMarkerName bails when the marker is invalid (rejected)", async () => {
    const setName = vi.fn().mockResolvedValue(undefined);
    const marker = {
      UID: "m2",
      Name: "",
      SubjUID: "",
      Invalid: true,
      thumbnailUrl: () => "/t/thumb2/public/tile_160",
      setName,
    };
    const photo = { ...mockPhoto, getMarkers: vi.fn().mockReturnValue([marker]) };
    const w = mountInfoForChips({ modelValue: mockModel, photo });
    await w.vm.$nextTick();
    w.vm.setMarkerInputValue("m2", "Plane Port");
    w.vm.confirmMarkerName(marker, "enter");
    expect(setName).not.toHaveBeenCalled();
  });

  // P1-7 — onRemoveMarker / onClearSubject stamp _lastDestructiveMarkerActionAt;
  // a follow-on @blur within 200ms must be ignored to prevent the destructive
  // action from racing the inline-commit path.
  it("confirmMarkerName bails when an icon click was registered <200ms ago", async () => {
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
    w.vm._lastDestructiveMarkerActionAt = Date.now();
    w.vm.confirmMarkerName(marker, "blur");
    expect(setName).not.toHaveBeenCalled();
    // After the 200ms window expires the commit is allowed again.
    w.vm._lastDestructiveMarkerActionAt = Date.now() - 500;
    w.vm.confirmMarkerName(marker, "enter");
    expect(setName).toHaveBeenCalledTimes(1);
  });

  // P1-7 — onClearSubject arms the destructive-action stamp even when
  // emitting upward; the parent lightbox owns the actual mutation. The
  // sibling onRemoveMarker path was retired when the combobox's
  // mdi-close hazard moved to the overlay's edit-mode click-to-remove.
  it("onClearSubject arms _lastDestructiveMarkerActionAt before emitting", () => {
    const marker = { UID: "m1", Name: "Jane Doe", SubjUID: "subj1", thumbnailUrl: () => "/t/x/p/tile_160" };
    const photo = { ...mockPhoto, getMarkers: vi.fn().mockReturnValue([marker]) };
    const w = mountInfoForChips({ modelValue: mockModel, photo });
    expect(w.vm._lastDestructiveMarkerActionAt).toBeFalsy();
    w.vm.onClearSubject(marker);
    expect(w.vm._lastDestructiveMarkerActionAt).toBeGreaterThan(0);
  });

  // P1-8 — the Add-name dialog stores a UID (not the transient Marker
  // instance returned by getMarkers). On confirm, the live Marker is
  // re-derived; a missing UID resolves to a silent no-op.
  it("onAddNameConfirm resolves the marker by UID at commit time", async () => {
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
    expect(w.vm.addNameDialog.markerUid).toBe("m2");
    w.vm.onAddNameConfirm();
    expect(setName).toHaveBeenCalledTimes(1);
  });

  it("onAddNameConfirm is a silent no-op when the stored UID no longer resolves", async () => {
    const setName = vi.fn();
    const marker = {
      UID: "m2",
      Name: "",
      SubjUID: "",
      thumbnailUrl: () => "/t/thumb2/public/tile_160",
      setName,
    };
    // First render has the marker; after blur opens the dialog the photo
    // navigates away and the marker disappears from getMarkers().
    const getMarkers = vi.fn().mockReturnValueOnce([marker]).mockReturnValue([]);
    const photo = { ...mockPhoto, getMarkers };
    const w = mountInfoForChips({ modelValue: mockModel, photo });
    await w.vm.$nextTick();
    w.vm.setMarkerInputValue("m2", "Plane Port");
    w.vm.confirmMarkerName(marker, "blur");
    expect(w.vm.addNameDialog.visible).toBe(true);
    w.vm.onAddNameConfirm();
    expect(setName).not.toHaveBeenCalled();
    expect(w.vm.addNameDialog.visible).toBe(false);
  });

  // P1-9 — cancelMarkerName must blur the marker's own input, not whatever
  // happens to be document.activeElement at the moment Esc was pressed.
  it("cancelMarkerName scopes the blur to the marker's own input, not document.activeElement", async () => {
    const marker = { UID: "m2", Name: "", SubjUID: "", thumbnailUrl: () => "/svg/portrait", setName: vi.fn() };
    const photo = { ...mockPhoto, getMarkers: vi.fn().mockReturnValue([marker]) };
    const w = mountInfoForChips({ modelValue: mockModel, photo });
    await w.vm.$nextTick();
    w.vm.setMarkerInputValue("m2", "Plane Port");
    // The previous implementation called `document.activeElement.blur()`
    // blindly; with an unrelated focused element this would incorrectly
    // blur it. Verify the scoped query path leaves the outsider alone.
    const outsider = document.createElement("input");
    document.body.appendChild(outsider);
    outsider.focus();
    expect(document.activeElement).toBe(outsider);
    const outsiderBlur = vi.spyOn(outsider, "blur");
    // Spy on the marker row's own input so we also confirm the scoped path
    // ran (not just that the blind path didn't).
    const markerInput = w.vm.$el.querySelector(`[data-marker-uid="m2"] input`);
    expect(markerInput).toBeTruthy();
    const markerBlur = vi.spyOn(markerInput, "blur");
    w.vm.cancelMarkerName(marker);
    expect(markerBlur).toHaveBeenCalled();
    expect(outsiderBlur).not.toHaveBeenCalled();
    document.body.removeChild(outsider);
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

  // onInlineEnter — Enter commits on single-line fields (commitOnEnter:
  // true) and falls through (no preventDefault) for free-form fields
  // like Notes / Caption so the textarea can insert a newline. Shift+
  // Enter always falls through, even on commitOnEnter fields.
  describe("onInlineEnter (commit-on-Enter for single-line fields)", () => {
    let w;
    beforeEach(() => {
      w = mountSidebar({
        props: {
          modelValue: mockModel,
          photo: { ...mockPhoto, Details: { Subject: "", Notes: "", Keywords: "" }, update: vi.fn(), wasChanged: () => false },
          canEdit: true,
          context: contexts.Photos,
        },
        global: { stubs: { PMap: true } },
      });
    });

    function makeEvent(shiftKey) {
      return { shiftKey, preventDefault: vi.fn(), stopPropagation: vi.fn() };
    }

    it("commits and suppresses the newline for commitOnEnter fields", () => {
      const subject = w.vm.fieldRegistry.subject;
      expect(subject.commitOnEnter).toBe(true);
      const ev = makeEvent(false);
      const confirmSpy = vi.spyOn(w.vm, "confirmField").mockImplementation(() => {});
      w.vm.onInlineEnter(ev, subject);
      expect(ev.preventDefault).toHaveBeenCalledTimes(1);
      expect(ev.stopPropagation).toHaveBeenCalledTimes(1);
      expect(confirmSpy).toHaveBeenCalledTimes(1);
    });

    it("falls through (no commit) for fields without commitOnEnter (Notes)", () => {
      const notes = w.vm.fieldRegistry.notes;
      expect(notes.commitOnEnter).toBeUndefined();
      const ev = makeEvent(false);
      const confirmSpy = vi.spyOn(w.vm, "confirmField").mockImplementation(() => {});
      w.vm.onInlineEnter(ev, notes);
      expect(ev.preventDefault).not.toHaveBeenCalled();
      expect(ev.stopPropagation).not.toHaveBeenCalled();
      expect(confirmSpy).not.toHaveBeenCalled();
    });

    it("falls through on Shift+Enter even for commitOnEnter fields", () => {
      const keywords = w.vm.fieldRegistry.keywords;
      expect(keywords.commitOnEnter).toBe(true);
      const ev = makeEvent(true);
      const confirmSpy = vi.spyOn(w.vm, "confirmField").mockImplementation(() => {});
      w.vm.onInlineEnter(ev, keywords);
      expect(ev.preventDefault).not.toHaveBeenCalled();
      expect(confirmSpy).not.toHaveBeenCalled();
    });
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

  // Inline text fields (title/caption/subject/...) are usually NOT tracked
  // by hasPendingEdit because onInlineFieldBlur() auto-commits them before
  // any nav source can fire. The carve-out is hasPendingInlineOverflow:
  // confirmField() short-circuits on a value above the field's maxLength,
  // leaving the editor open with invalid text — that case (covered below)
  // intentionally DOES report pending so the discard dialog can fire.
  it("should NOT report hasPendingEdit for a dirty inline text field within the cap (auto-commits on blur)", () => {
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

  // Navigation guard: if the user typed past the per-field cap (e.g. 5000-
  // char Caption against the 4096 cap), confirmField() keeps the editor
  // open instead of persisting the value. Without the overflow branch in
  // hasPendingEdit, the next/prev arrow would silently swap the photo
  // out from under the invalid edit; with it, confirmDiscardPending() can
  // surface the discard dialog and the "Discard invalid changes?" label
  // tells the user which kind of edit they're abandoning.
  it("should report hasPendingEdit when an inline editor's value exceeds the field cap", () => {
    const photo = {
      ...mockPhoto,
      Caption: "x".repeat(5000), // PhotoMaxLength.Caption is 4096
      wasChanged: () => true,
    };
    const w = mountSidebar({
      props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });

    // No active editor → not pending.
    expect(w.vm.hasPendingEdit()).toBe(false);

    // Editing the overlength field → pending; discard label flips.
    w.vm.editingField = "caption";
    expect(w.vm.hasPendingInlineOverflow()).toBe(true);
    expect(w.vm.hasPendingEdit()).toBe(true);
    expect(w.vm.discardDialogText).toBe("Discard invalid changes?");
  });

  it("should keep the generic discard text when overflow coexists with another pending edit", () => {
    const photo = {
      ...mockPhoto,
      Caption: "x".repeat(5000),
      wasChanged: () => true,
    };
    const w = mountSidebar({
      props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });

    w.vm.editingField = "caption";
    w.vm.chipState.labels.removals = [{ Label: { UID: "lbl1" } }];

    expect(w.vm.hasPendingInlineOverflow()).toBe(true);
    expect(w.vm.hasPendingNonOverflowEdit()).toBe(true);
    expect(w.vm.discardDialogText).toBe("Discard unsaved changes?");
  });

  // Additions go through the instant-save path (addLabelImmediate /
  // addAlbumImmediate), so they never enter chipState — only batched
  // removals can leave the sidebar in a pending state.
  it("should report hasPendingEdit when labels have pending removals", () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(w.vm.hasPendingEdit()).toBe(false);
    w.vm.chipState.labels.removals = [{ Label: { UID: "lbl1" } }];
    expect(w.vm.hasPendingEdit()).toBe(true);
  });

  it("should report hasPendingEdit when albums have pending removals", () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(w.vm.hasPendingEdit()).toBe(false);
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

  it("should report hasPendingEdit for typed-but-uncommitted text in the labels combobox", () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(w.vm.hasPendingEdit()).toBe(false);
    w.vm.chipState.labels.search = "alpha";
    expect(w.vm.hasPendingEdit()).toBe(true);
    // Trimmed-empty input is not pending.
    w.vm.chipState.labels.search = "   ";
    expect(w.vm.hasPendingEdit()).toBe(false);
  });

  it("should report hasPendingEdit for typed-but-uncommitted text in the albums autocomplete", () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    w.vm.chipState.albums.search = "vacation";
    expect(w.vm.hasPendingEdit()).toBe(true);
  });

  it("should report hasPendingEdit while the Add-name confirmation dialog is visible", () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(w.vm.hasPendingEdit()).toBe(false);
    w.vm.addNameDialog = { visible: true, marker: { UID: "m2" }, name: "Alice" };
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

    it("trims surrounding whitespace before checking the field cap so a value that fits after trim still saves", () => {
      // Mirrors the backend trim performed by txt.Clip; without it a user
      // typing "x"*1024 + trailing spaces would trip the cap and the
      // editor would stay open even though the trimmed value fits the
      // column.
      const photo = buildInlineEditPhoto();
      photo.update.mockResolvedValueOnce({});
      const w = mountSidebar({
        props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
        global: { stubs: { PMap: true } },
      });

      // Copyright cap is 1024; "x"*1024 + trailing spaces is fine after trim.
      photo.Details = { Copyright: "x".repeat(1024) + "   " };
      w.vm.editingField = "copyright";

      w.vm.confirmField();

      // Trimmed value persisted on the photo and the editor closed.
      expect(photo.Details.Copyright).toBe("x".repeat(1024));
      expect(w.vm.editingField).toBeNull();
      expect(w.vm.$notify.error).not.toHaveBeenCalled();
    });

    it("blocks save and keeps the editor open when the value exceeds the field cap", () => {
      // Inline editor has no parent v-form; the rendered `:rules` only
      // paint an inline error. The imperative length check inside
      // confirmField is the actual gate — without it photo.update()
      // would post overlength input and the "successfully saved" toast
      // would fire while the server silently drops the value.
      const photo = buildInlineEditPhoto();

      const w = mountSidebar({
        props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
        global: { stubs: { PMap: true } },
      });

      // Copyright cap is 1024 (PhotoMaxLength.Copyright); push past it.
      photo.Details = { Copyright: "x".repeat(1025) };
      w.vm.editingField = "copyright";

      w.vm.confirmField();

      // Editor stays open so the user can fix the overlong input.
      expect(w.vm.editingField).toBe("copyright");
      expect(photo.update).not.toHaveBeenCalled();
      expect(w.vm.$notify.error).toHaveBeenCalledTimes(1);
    });
  });

  // Labels
  it("should return labels from photo prop sorted alphabetically", () => {
    // Backend preload orders by uncertainty/topicality but the sidebar
    // doesn't surface those scores; the chips are sorted by name.
    expect(wrapper.vm.labels).toHaveLength(2);
    expect(wrapper.vm.labels.map((l) => l.Label.Name)).toEqual(["Landscape", "Nature"]);
  });

  // Albums
  it("should return albums from photo prop sorted alphabetically", () => {
    expect(wrapper.vm.albums).toHaveLength(2);
    expect(wrapper.vm.albums.map((a) => a.Title)).toEqual(["Favorites", "Vacation 2023"]);
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
    // Alphabetical sort: Favorites < Vacation 2023.
    expect(w.vm.albums.map((a) => a.UID)).toEqual(["alb2", "alb1"]);
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

  // fileName resolution prefers the underlying media file (video / Live)
  // over the generated JPEG cover that primaryFile() returns. Previously
  // the sidebar showed "...mp4.jpg" for video photos because the JPG was
  // marked Primary for indexing — the user-facing name should be the
  // original media file.
  it("fileName surfaces the original media file for video photos, not the JPEG cover", () => {
    const videoPhoto = {
      ...mockPhoto,
      Type: "video",
      FileName: "video/tagesschau.mp4.jpg",
      originalFile: () => ({ Name: "video/tagesschau.mp4", OriginalName: "tagesschau.mp4" }),
    };
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: videoPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true }, mocks: { $util: validationUtil } },
    });
    expect(w.vm.fileName).toBe("video/tagesschau.mp4");
  });

  // For image photos the originalFile() result equals primaryFile() — both
  // return `this` when Files has fewer than 2 entries — so we fall through
  // to the existing photo.FileName branch.
  it("fileName falls back to photo.FileName when originalFile is the photo itself", () => {
    const photo = { ...mockPhoto, originalFile: function () {
      return this; 
    }, FileName: "photos/2023/IMG_001.jpg" };
    const w = mountSidebar({
      props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true }, mocks: { $util: validationUtil } },
    });
    expect(w.vm.fileName).toBe("photos/2023/IMG_001.jpg");
  });

  // Caption and notes HTML
  it("should produce caption and notes HTML via sanitize pipeline", () => {
    expect(wrapper.vm.captionHtml).toBe("Test Caption");
    expect(wrapper.vm.notesHtml).toBe("Some notes about this photo");
  });

  // The merged single-row detailsFields template must route Notes
  // through the sanitized v-html branch (display: "html" + htmlValue:
  // "notesHtml") rather than plain-text interpolation. Catches any
  // future regression where the html-aware ladder is dropped from the
  // merged loop and Notes silently falls back to the text branch.
  it("renders Notes through the sanitized v-html branch in the merged details row", () => {
    const notesRow = wrapper.find(".meta-notes");
    expect(notesRow.exists()).toBe(true);
    const innerNotes = notesRow.find(".meta-scrollable.meta-notes");
    expect(innerNotes.exists()).toBe(true);
    expect(innerNotes.html()).toContain("Some notes about this photo");
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

  // inlineEditDirty + undoInlineEdit — Undo affordance for inline-text
  // editors. The Undo button is always shown while editing, but
  // disabled until inlineEditDirty flips true; clicking it reverts to
  // the editOriginal captured at startEditing time without exiting
  // edit mode (Escape is the revert-AND-exit gesture).
  describe("inline edit Undo", () => {
    let w;
    let photo;
    beforeEach(() => {
      photo = {
        ...mockPhoto,
        Title: "Original Title",
        Details: { Subject: "Original Subject", Notes: "", Keywords: "" },
        update: vi.fn(),
        wasChanged: () => true,
      };
      w = mountSidebar({
        props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
        global: { stubs: { PMap: true } },
      });
    });

    it("inlineEditDirty is false when no field is being edited", () => {
      expect(w.vm.editingField).toBeNull();
      expect(w.vm.inlineEditDirty).toBe(false);
    });

    it("inlineEditDirty stays false right after startEditing (value unchanged)", async () => {
      w.vm.startEditing("subject");
      await w.vm.$nextTick();
      expect(w.vm.editingField).toBe("subject");
      expect(w.vm.inlineEditDirty).toBe(false);
    });

    it("inlineEditDirty flips true once the field value diverges from the original", async () => {
      w.vm.startEditing("subject");
      await w.vm.$nextTick();
      // Route the mutation through the reactive write helper that the
      // textarea's @update:model-value handler uses, so Vue's computed
      // dependency tracking picks it up.
      w.vm.setFieldValue("subject", "Edited");
      await w.vm.$nextTick();
      expect(w.vm.getFieldValue("subject")).toBe("Edited");
      expect(w.vm.inlineEditDirty).toBe(true);
    });

    it("undoInlineEdit reverts to editOriginal without exiting edit mode", async () => {
      w.vm.startEditing("subject");
      await w.vm.$nextTick();
      w.vm.setFieldValue("subject", "Edited");
      await w.vm.$nextTick();
      expect(w.vm.editingField).toBe("subject");

      w.vm.undoInlineEdit();
      await w.vm.$nextTick();
      expect(w.vm.getFieldValue("subject")).toBe("Original Subject");
      expect(w.vm.editingField).toBe("subject");
      expect(w.vm.inlineEditDirty).toBe(false);
    });

    it("undoInlineEdit is a no-op when no field is being edited", () => {
      expect(w.vm.editingField).toBeNull();
      w.vm.undoInlineEdit();
      expect(photo.Details.Subject).toBe("Original Subject");
    });

    // Defaults-agnostic binding check: the test sets each flag to a
    // known value before asserting, so flipping the data() default
    // (e.g., for an A/B variant that hides Save by default) doesn't
    // break the test. What we care about is that the root :class
    // binding actually mirrors the flags in both directions.
    it("hide-edit-undo / hide-edit-save root classes mirror the data flags in both directions", async () => {
      const root = w.find(".p-lightbox-sidebar");

      w.vm.hideEditUndo = false;
      w.vm.hideEditSave = false;
      await w.vm.$nextTick();
      expect(root.classes()).not.toContain("hide-edit-undo");
      expect(root.classes()).not.toContain("hide-edit-save");

      w.vm.hideEditUndo = true;
      w.vm.hideEditSave = true;
      await w.vm.$nextTick();
      expect(root.classes()).toContain("hide-edit-undo");
      expect(root.classes()).toContain("hide-edit-save");
    });
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

  // L8: instant-save additions — onLabelSelected / onLabelEnter call
  // photo.addLabel(name) immediately. The chip appears as a real primary
  // chip via the model's setValues(r.data) chain (mocked here as a
  // resolved Promise). chipState.labels.additions does not exist anymore.
  it("should call photo.addLabel immediately on onLabelSelected", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.onLabelSelected({ Name: "Sunset", UID: "lbl-new" });
    expect(mockPhoto.addLabel).toHaveBeenCalledWith("Sunset");
    expect(mockPhoto.addLabel).toHaveBeenCalledTimes(1);
  });

  it("should ignore non-object values in onLabelSelected", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.onLabelSelected("string-value");
    w.vm.onLabelSelected(null);
    expect(mockPhoto.addLabel).not.toHaveBeenCalled();
  });

  it("should skip labels already on the photo in onLabelSelected", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.onLabelSelected({ Name: "Nature", UID: "lbl1" });
    expect(mockPhoto.addLabel).not.toHaveBeenCalled();
  });

  it("should skip labels already on the photo case-insensitively in onLabelSelected", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.onLabelSelected({ Name: "nature" });
    expect(mockPhoto.addLabel).not.toHaveBeenCalled();
  });

  it("should trim whitespace in onLabelEnter before calling photo.addLabel", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.chipState.labels.search = "  dog  ";
    w.vm.onLabelEnter();
    expect(mockPhoto.addLabel).toHaveBeenCalledWith("dog");
  });

  it("should silently reject empty or whitespace-only label input in onLabelEnter", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.chipState.labels.search = "   ";
    w.vm.onLabelEnter();
    expect(mockPhoto.addLabel).not.toHaveBeenCalled();
    expect(w.vm.$notify.error).not.toHaveBeenCalled();
  });

  it("should reject labels longer than the configured clip length and notify", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.chipState.labels.search = "a".repeat(CLIP_LEN + 10);
    w.vm.onLabelEnter();
    expect(mockPhoto.addLabel).not.toHaveBeenCalled();
    expect(w.vm.$notify.error).toHaveBeenCalledWith("Name too long");
  });

  it("should match existing labels through normalization (punctuation stripped)", () => {
    const photo = {
      ...mockPhoto,
      Labels: [{ Uncertainty: 0, Label: { ID: 99, UID: "lbl99", Name: "Cat!", Slug: "cat", CustomSlug: "" } }],
    };
    const w = mountInfoForChips({ modelValue: mockModel, photo });
    w.vm.chipState.labels.search = "cat";
    w.vm.onLabelEnter();
    expect(photo.addLabel).not.toHaveBeenCalled();
  });

  it("should match existing labels through normalization (& vs and)", () => {
    const photo = {
      ...mockPhoto,
      Labels: [{ Uncertainty: 0, Label: { ID: 99, UID: "lbl99", Name: "Rock & Roll", Slug: "rock-and-roll", CustomSlug: "" } }],
    };
    const w = mountInfoForChips({ modelValue: mockModel, photo });
    w.vm.chipState.labels.search = "rock and roll";
    w.vm.onLabelEnter();
    expect(photo.addLabel).not.toHaveBeenCalled();
  });

  it("should silently reject punctuation-only label input", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.chipState.labels.search = "!!!";
    w.vm.onLabelEnter();
    expect(mockPhoto.addLabel).not.toHaveBeenCalled();
    expect(w.vm.$notify.error).not.toHaveBeenCalled();
  });

  it("should accept emoji-only label input and call photo.addLabel", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.chipState.labels.search = "🌅";
    w.vm.onLabelEnter();
    expect(mockPhoto.addLabel).toHaveBeenCalledWith("🌅");
  });

  it("should notify on photo.addLabel rejection (no transient chip)", async () => {
    mockPhoto.addLabel = vi.fn().mockRejectedValue(new Error("boom"));
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.chipState.labels.search = "Sunset";
    w.vm.onLabelEnter();
    // Wait two microtasks for the promise rejection + .catch handler.
    await Promise.resolve();
    await Promise.resolve();
    expect(w.vm.$notify.error).toHaveBeenCalledWith("Failed to save changes");
  });

  // L3: addLabelImmediate cross-checks labelOptions via $util.normalizeTitle
  // so typed variants like `Hello Cat` / `hello-cat` collapse onto the
  // canonical existing label (preserving server-side casing/punctuation).
  it("should canonicalize the typed name to an existing labelOption when normalized-equal", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.chipState.labels.options = [{ UID: "lbl-canonical", Name: "Hello Cat" }];
    w.vm.chipState.labels.search = "hello-cat";
    w.vm.onLabelEnter();
    // Backend gets the canonical existing-label name, NOT the typed variant.
    expect(mockPhoto.addLabel).toHaveBeenCalledWith("Hello Cat");
  });

  it("should pass the typed name through when no labelOption matches", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.chipState.labels.options = [{ UID: "lbl-canonical", Name: "Hello Cat" }];
    w.vm.chipState.labels.search = "Sunset";
    w.vm.onLabelEnter();
    // Sunset has no match in labelOptions → typed name is sent verbatim.
    expect(mockPhoto.addLabel).toHaveBeenCalledWith("Sunset");
  });

  it("should canonicalize across punctuation and case variants", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.chipState.labels.options = [{ UID: "lbl-canonical", Name: "Rock & Roll" }];
    // & expands to "and" in normalization, so "Rock and Roll" maps to the
    // same canonical label and the saved name picks up the & spelling.
    w.vm.chipState.labels.search = "Rock and Roll";
    w.vm.onLabelEnter();
    expect(mockPhoto.addLabel).toHaveBeenCalledWith("Rock & Roll");
  });

  // Parity with onAlbumEnter: pressing Enter on non-empty Labels input
  // ALWAYS clears the input and bumps the key (force-remount closes the
  // menu), regardless of whether addLabelImmediate fired an API call.
  // The pre-fix code only cleared on success, which left the input
  // populated and the menu open when the typed label was already on the
  // photo — felt unresolved next to Albums.
  it("clears the input and bumps the key when Enter fires a real add", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    const startKey = w.vm.chipState.labels.key;
    w.vm.chipState.labels.search = "Brand New Label";
    w.vm.onLabelEnter();
    expect(mockPhoto.addLabel).toHaveBeenCalledWith("Brand New Label");
    expect(w.vm.chipState.labels.search).toBe("");
    expect(w.vm.chipState.labels.input).toBe(null);
    expect(w.vm.chipState.labels.key).toBe(startKey + 1);
  });

  it("clears the input and bumps the key when Enter is pressed on a label already on the photo", () => {
    const photo = {
      ...mockPhoto,
      Labels: [{ Uncertainty: 0, Label: { ID: 99, UID: "lbl99", Name: "Nature", Slug: "nature", CustomSlug: "" } }],
    };
    const w = mountInfoForChips({ modelValue: mockModel, photo });
    const startKey = w.vm.chipState.labels.key;
    w.vm.chipState.labels.search = "Nature";
    w.vm.onLabelEnter();
    // The label is already on the photo — addLabelImmediate short-circuits
    // and no API call is made.
    expect(photo.addLabel).not.toHaveBeenCalled();
    // …but the input must still clear and the menu close so the gesture
    // feels resolved (matches onAlbumEnter).
    expect(w.vm.chipState.labels.search).toBe("");
    expect(w.vm.chipState.labels.input).toBe(null);
    expect(w.vm.chipState.labels.key).toBe(startKey + 1);
  });

  it("does not clear or bump the key on empty Enter", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    const startKey = w.vm.chipState.labels.key;
    w.vm.chipState.labels.search = "   ";
    w.vm.onLabelEnter();
    expect(mockPhoto.addLabel).not.toHaveBeenCalled();
    // Empty/whitespace Enter is a no-op — leave the input alone so the
    // user can keep typing without losing focus mid-keystroke.
    expect(w.vm.chipState.labels.key).toBe(startKey);
  });

  it("does not clear the input when the typed name is too long", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    const startKey = w.vm.chipState.labels.key;
    const typed = "a".repeat(CLIP_LEN + 10);
    w.vm.chipState.labels.search = typed;
    w.vm.onLabelEnter();
    expect(mockPhoto.addLabel).not.toHaveBeenCalled();
    expect(w.vm.$notify.error).toHaveBeenCalledWith("Name too long");
    // Length-validation failures keep the typed text so the user can
    // shorten it instead of losing the whole entry.
    expect(w.vm.chipState.labels.search).toBe(typed);
    expect(w.vm.chipState.labels.key).toBe(startKey);
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

  // L8: instant-save additions — onAlbumSelected calls photo.addToAlbum
  // with the album UID immediately. The chip appears as a real primary
  // chip via the model's evict+refind chain (mocked here as a resolved
  // Promise). chipState.albums.additions does not exist anymore.
  it("should call photo.addToAlbum immediately on onAlbumSelected", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    const album = { UID: "alb-new", Title: "New Album" };

    w.vm.onAlbumSelected(album);
    expect(mockPhoto.addToAlbum).toHaveBeenCalledWith("alb-new");
    expect(mockPhoto.addToAlbum).toHaveBeenCalledTimes(1);
  });

  // Chip keyboard accessibility (proposal item L6) — onChipActivate covers
  // click + Enter (always navigates, in both editable and read-only modes),
  // onChipDelete covers Delete + Backspace and the × icon click (toggles
  // pending removal only when editable).
  it("onChipActivate navigates a label chip when the section is read-only", () => {
    const openSpy = vi.spyOn(window, "open").mockImplementation(() => null);
    // mountWithMockRouter omits canEdit → isEditable false.
    const w = mountWithMockRouter("/library/browse");
    w.vm.onChipActivate("labels", { Label: { ID: 1, UID: "lbl1", Name: "Nature", Slug: "nature", CustomSlug: "" } });
    expect(w.vm.$router.resolve).toHaveBeenCalledWith({ name: "browse", query: { q: "label:nature" } });
    expect(openSpy).toHaveBeenCalled();
    openSpy.mockRestore();
  });

  // Chip click is always a link — even in edit mode. Removal happens via
  // the × icon (separate click handler on the v-icon) or the keyboard
  // Delete / Backspace path. The earlier "click anywhere on the chip to
  // remove" behavior was a regression flagged by the user on May 11.
  it("onChipActivate navigates a label chip even when the section is editable", () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    const navSpy = vi.spyOn(w.vm, "navigateToLabel").mockImplementation(() => null);
    const label = { Label: { ID: 7, UID: "lbl7", Name: "Beach" } };
    w.vm.onChipActivate("labels", label);

    expect(navSpy).toHaveBeenCalledWith(label.Label);
    expect(w.vm.isChipPendingRemoval("labels", 7)).toBe(false);
    navSpy.mockRestore();
  });

  it("onChipActivate navigates an album chip when the section is read-only", () => {
    const openSpy = vi.spyOn(window, "open").mockImplementation(() => null);
    const w = mountWithMockRouter("/library/albums/alb1/view");
    w.vm.onChipActivate("albums", { UID: "alb1", Title: "Vacation 2023" });
    expect(w.vm.$router.resolve).toHaveBeenCalledWith({ name: "album", params: { album: "alb1", slug: "view" } });
    expect(openSpy).toHaveBeenCalled();
    openSpy.mockRestore();
  });

  it("onChipActivate navigates an album chip even when the section is editable", () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    const navSpy = vi.spyOn(w.vm, "navigateToAlbum").mockImplementation(() => null);
    const album = { UID: "alb-x", Title: "Trip" };
    w.vm.onChipActivate("albums", album);

    expect(navSpy).toHaveBeenCalledWith(album);
    expect(w.vm.isChipPendingRemoval("albums", "alb-x")).toBe(false);
    navSpy.mockRestore();
  });

  it("onChipDelete is a no-op when the section is not editable", () => {
    // mountWithMockRouter omits canEdit → isEditable false. onChipDelete
    // should refuse to stage a removal in a read-only context.
    const w = mountWithMockRouter("/library/browse");
    w.vm.onChipDelete("labels", { Label: { ID: 1 } });
    w.vm.onChipDelete("albums", { UID: "alb1" });
    expect(w.vm.chipState.labels.removals).toHaveLength(0);
    expect(w.vm.chipState.albums.removals).toHaveLength(0);
  });

  it("onChipDelete toggles removal independently per field when editable", () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    // Both fields' chips are deletable simultaneously now that the chip
    // sections share a single isEditable gate (no per-section edit-mode).
    w.vm.onChipDelete("labels", { Label: { ID: 9 } });
    w.vm.onChipDelete("albums", { UID: "alb-y" });
    expect(w.vm.isChipPendingRemoval("labels", 9)).toBe(true);
    expect(w.vm.isChipPendingRemoval("albums", "alb-y")).toBe(true);
  });

  it("onChipActivate / onChipDelete tolerate falsy items", () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    expect(() => w.vm.onChipActivate("labels", null)).not.toThrow();
    expect(() => w.vm.onChipDelete("labels", undefined)).not.toThrow();
  });

  it("should ignore non-object values in onAlbumSelected", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.onAlbumSelected("string-value");
    w.vm.onAlbumSelected(null);
    expect(mockPhoto.addToAlbum).not.toHaveBeenCalled();
  });

  // v-combobox emits transient update:model-value events while the user
  // types free text. Clearing the input here would bump chipState.albums.key,
  // force-remount the combobox, and kill focus mid-keystroke. The non-object
  // path must be a silent no-op so typing into the album combobox stays
  // usable. Mirrors onLabelSelected.
  it("should not clear chipState.albums on transient non-object onAlbumSelected", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.chipState.albums.search = "test";
    w.vm.chipState.albums.input = "test";
    const initialKey = w.vm.chipState.albums.key;
    w.vm.onAlbumSelected("test");
    w.vm.onAlbumSelected(null);
    w.vm.onAlbumSelected({ Title: "test" }); // free-text stub without UID
    expect(w.vm.chipState.albums.search).toBe("test");
    expect(w.vm.chipState.albums.input).toBe("test");
    expect(w.vm.chipState.albums.key).toBe(initialKey);
  });

  it("should skip albums already on the photo in onAlbumSelected", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.onAlbumSelected({ UID: "alb1", Title: "Vacation 2023" });
    expect(mockPhoto.addToAlbum).not.toHaveBeenCalled();
  });

  // Album validation parity with batch edit + labels tab.
  it("should dedupe albums by normalized title even when UIDs differ", () => {
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.onAlbumSelected({ UID: "alb-other", Title: "vacation 2023" });
    expect(mockPhoto.addToAlbum).not.toHaveBeenCalled();
  });

  it("should reject overlong album titles in onAlbumEnter and not call save", () => {
    const saveSpy = vi.spyOn(Album.prototype, "save").mockResolvedValue();
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.chipState.albums.search = "a".repeat(CLIP_LEN + 10);
    w.vm.onAlbumEnter();
    expect(saveSpy).not.toHaveBeenCalled();
    expect(mockPhoto.addToAlbum).not.toHaveBeenCalled();
    expect(w.vm.$notify.error).toHaveBeenCalledWith("Name too long");
    saveSpy.mockRestore();
  });

  it("should ignore empty/whitespace input in onAlbumEnter and not call save", () => {
    const saveSpy = vi.spyOn(Album.prototype, "save").mockResolvedValue();
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.chipState.albums.search = "   ";
    w.vm.onAlbumEnter();
    expect(saveSpy).not.toHaveBeenCalled();
    expect(mockPhoto.addToAlbum).not.toHaveBeenCalled();
    saveSpy.mockRestore();
  });

  it("should skip onAlbumEnter when title matches existing album case-insensitively", () => {
    const saveSpy = vi.spyOn(Album.prototype, "save").mockResolvedValue();
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.chipState.albums.search = "VACATION 2023";
    w.vm.onAlbumEnter();
    expect(saveSpy).not.toHaveBeenCalled();
    expect(mockPhoto.addToAlbum).not.toHaveBeenCalled();
    saveSpy.mockRestore();
  });

  it("should create a new album in onAlbumEnter and add the photo to it", async () => {
    const saveSpy = vi.spyOn(Album.prototype, "save").mockImplementation(function () {
      this.UID = "alb-created";
      return Promise.resolve(this);
    });
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.chipState.albums.options = [];
    w.vm.chipState.albums.search = "Brand New Trip";
    w.vm.onAlbumEnter();
    await new Promise((r) => setTimeout(r, 0));
    expect(saveSpy).toHaveBeenCalledTimes(1);
    // After Album.save resolves with a UID, the sidebar fires the
    // instant-save addAlbumImmediate path → photo.addToAlbum(newUid).
    expect(mockPhoto.addToAlbum).toHaveBeenCalledWith("alb-created");
    expect(w.vm.chipState.albums.options.some((a) => a.UID === "alb-created")).toBe(true);
    saveSpy.mockRestore();
  });

  it("should notify on photo.addToAlbum rejection (no transient chip)", async () => {
    mockPhoto.addToAlbum = vi.fn().mockRejectedValue(new Error("boom"));
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.onAlbumSelected({ UID: "alb-new", Title: "New Album" });
    // Wait two microtasks for the promise rejection + .catch handler.
    await Promise.resolve();
    await Promise.resolve();
    expect(w.vm.$notify.error).toHaveBeenCalledWith("Failed to save changes");
  });

  // onAlbumEnter resolves typed text to an existing album only via a
  // normalized exact match (case + punctuation + `+`/`_`/`-` → space).
  // Substring fuzzy matching is intentionally NOT applied — typing `test`
  // must not silently merge into an unrelated `LRUTEST-ALBUM-…`. Users pick
  // partial matches via the dropdown (click or arrow + Enter on a
  // highlighted item, which fires `onAlbumSelected`).
  it("onAlbumEnter creates a new album when typed text only fuzzy-matches existing options", async () => {
    const saveSpy = vi.spyOn(Album.prototype, "save").mockImplementation(function () {
      this.UID = "alb-new-ar";
      return Promise.resolve(this);
    });
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.chipState.albums.options = [
      { UID: "alb-berlin", Title: "Berlin" },
      { UID: "alb-archive", Title: "Archive" },
      { UID: "alb-arctic", Title: "Arctic" },
    ];
    // `ar` is a substring of "Archive"/"Arctic"/"Berlin" but NOT a
    // normalized exact match for any of them. The legacy fuzzy fallback
    // would have silently merged into "Archive"; the new behavior creates
    // a fresh "ar" album to mirror the labels combobox's free-text path.
    w.vm.chipState.albums.search = "ar";
    w.vm.onAlbumEnter();
    await new Promise((r) => setTimeout(r, 0));
    expect(saveSpy).toHaveBeenCalledTimes(1);
    expect(mockPhoto.addToAlbum).toHaveBeenCalledWith("alb-new-ar");
    saveSpy.mockRestore();
  });

  it("onAlbumEnter normalizes punctuation and case for exact-match resolution", () => {
    const saveSpy = vi.spyOn(Album.prototype, "save").mockResolvedValue();
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.chipState.albums.options = [{ UID: "alb-hello-cat", Title: "Hello Cat" }];
    // `hello-cat`, `Hello+Cat`, `hello.CAT`, `HELLO CAT` all normalize to
    // "hello cat" (every punctuation character is converted to whitespace
    // by normalizeTitle), so they should resolve to the existing
    // "Hello Cat" album.
    for (const typed of ["hello-cat", "Hello+Cat", "hello.CAT", "HELLO CAT"]) {
      mockPhoto.addToAlbum.mockClear();
      w.vm.chipState.albums.search = typed;
      w.vm.onAlbumEnter();
      expect(mockPhoto.addToAlbum).toHaveBeenCalledWith("alb-hello-cat");
      // Pulled out the existing album — no Album.save round-trip.
      expect(saveSpy).not.toHaveBeenCalled();
    }
    saveSpy.mockRestore();
  });

  it("onAlbumEnter creates a new album for substring-only matches instead of merging", async () => {
    const saveSpy = vi.spyOn(Album.prototype, "save").mockImplementation(function () {
      this.UID = "alb-new-summer";
      return Promise.resolve(this);
    });
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.chipState.albums.options = [{ UID: "alb-summer-2023", Title: "Summer 2023" }];
    // `summer` is a substring of "Summer 2023" but not a normalized exact
    // match — typing it and pressing Enter must create a fresh "summer"
    // album, not silently add the photo to "Summer 2023".
    w.vm.chipState.albums.search = "summer";
    w.vm.onAlbumEnter();
    await new Promise((r) => setTimeout(r, 0));
    expect(saveSpy).toHaveBeenCalledTimes(1);
    expect(mockPhoto.addToAlbum).toHaveBeenCalledWith("alb-new-summer");
    saveSpy.mockRestore();
  });

  // Regression test for the reported bug: typing "test" must not match an
  // existing album whose title contains "test" as a substring.
  it("onAlbumEnter creates a new album when typed text appears as a substring of an existing title", async () => {
    const saveSpy = vi.spyOn(Album.prototype, "save").mockImplementation(function () {
      this.UID = "alb-new-test";
      return Promise.resolve(this);
    });
    const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
    w.vm.chipState.albums.options = [{ UID: "alb-lrutest", Title: "LRUTEST-ALBUM-1777990015505" }];
    w.vm.chipState.albums.search = "test";
    w.vm.onAlbumEnter();
    await new Promise((r) => setTimeout(r, 0));
    expect(saveSpy).toHaveBeenCalledTimes(1);
    expect(mockPhoto.addToAlbum).toHaveBeenCalledWith("alb-new-test");
    saveSpy.mockRestore();
  });

  // cancelEditing only resets inline-text edit state. Pending chip
  // removals are independent (committed via the toolbar ✓) and survive
  // an inline-text Esc — chip state is cleared by resetInlineEdits()
  // when the discard-pending dialog confirms.
  it("should leave chip removals intact when cancelEditing fires", () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    w.vm.editingField = "title";
    w.vm.editOriginal = "Original Title";
    w.vm.chipState.labels.removals = [1];
    w.vm.chipState.albums.removals = ["alb1"];

    w.vm._editStartedAt = Date.now() - 300;
    w.vm.cancelEditing();

    expect(w.vm.editingField).toBeNull();
    expect(w.vm.chipState.labels.removals).toEqual([1]);
    expect(w.vm.chipState.albums.removals).toEqual(["alb1"]);
  });

  it("should clear all pending state on resetInlineEdits", () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    w.vm.chipState.labels.removals = [1];
    w.vm.chipState.albums.removals = ["alb1"];
    w.vm.chipState.labels.search = "typed-but-not-saved";

    w.vm.resetInlineEdits();

    expect(w.vm.chipState.labels.removals).toHaveLength(0);
    expect(w.vm.chipState.albums.removals).toHaveLength(0);
    expect(w.vm.chipState.labels.search).toBe("");
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

  // L10: loadChipOptions reads from the shared module-scope typeahead
  // cache. The cache itself owns the cap warning + de-dup contract
  // (see common/typeahead-cache.test.js); these tests only pin that
  // the sidebar populates chipState.<field>.options from getLabels /
  // getAlbums and maps to the consumer-friendly shape.
  describe("loadChipOptions cache integration", () => {
    it("populates chipState.labels.options from typeaheadCache.getLabels", async () => {
      typeaheadCache.clear();
      const models = [{ Name: "Cat", UID: "lbl-cat", Slug: "cat" }];
      const cacheSpy = vi.spyOn(typeaheadCache, "getLabels").mockResolvedValueOnce(models);
      const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
      w.vm.loadChipOptions("labels");
      await Promise.resolve();
      await Promise.resolve();
      expect(cacheSpy).toHaveBeenCalled();
      // Sidebar maps to the {Name, UID} shape its combobox needs.
      expect(w.vm.chipState.labels.options).toEqual([{ Name: "Cat", UID: "lbl-cat" }]);
      cacheSpy.mockRestore();
    });

    it("populates chipState.albums.options from typeaheadCache.getAlbums", async () => {
      typeaheadCache.clear();
      const models = [{ Title: "Trip", UID: "alb-trip" }];
      const cacheSpy = vi.spyOn(typeaheadCache, "getAlbums").mockResolvedValueOnce(models);
      const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
      w.vm.loadChipOptions("albums");
      await Promise.resolve();
      await Promise.resolve();
      expect(cacheSpy).toHaveBeenCalled();
      // Sidebar passes album models through unchanged.
      expect(w.vm.chipState.albums.options).toEqual(models);
      cacheSpy.mockRestore();
    });

    it("swallows cache errors so a transient fetch failure does not block the editor", async () => {
      const cacheSpy = vi.spyOn(typeaheadCache, "getLabels").mockRejectedValueOnce(new Error("boom"));
      const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
      expect(() => w.vm.loadChipOptions("labels")).not.toThrow();
      await Promise.resolve();
      await Promise.resolve();
      expect(w.vm.chipState.labels.options).toEqual([]);
      cacheSpy.mockRestore();
    });

    // The backend's order=name doesn't always return a clean
    // alphabetical list (and the cap-bounded fetch may interleave).
    // Sort client-side via locale-aware comparison so the dropdown
    // reads naturally for the user.
    it("sorts label options alphabetically (case-insensitive)", async () => {
      typeaheadCache.clear();
      const cacheSpy = vi
        .spyOn(typeaheadCache, "getLabels")
        .mockResolvedValueOnce([
          { Name: "Mountain", UID: "1" },
          { Name: "apple", UID: "2" },
          { Name: "Beach", UID: "3" },
        ]);
      const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
      w.vm.loadChipOptions("labels");
      await Promise.resolve();
      await Promise.resolve();
      expect(w.vm.chipState.labels.options.map((l) => l.Name)).toEqual(["apple", "Beach", "Mountain"]);
      cacheSpy.mockRestore();
    });

    it("sorts album options alphabetically by title", async () => {
      typeaheadCache.clear();
      const cacheSpy = vi
        .spyOn(typeaheadCache, "getAlbums")
        .mockResolvedValueOnce([
          { Title: "Zebra", UID: "z" },
          { Title: "alpha", UID: "a" },
          { Title: "Mango", UID: "m" },
        ]);
      const w = mountInfoForChips({ modelValue: mockModel, photo: mockPhoto });
      w.vm.loadChipOptions("albums");
      await Promise.resolve();
      await Promise.resolve();
      expect(w.vm.chipState.albums.options.map((a) => a.Title)).toEqual(["alpha", "Mango", "Zebra"]);
      cacheSpy.mockRestore();
    });
  });

  // clearChipInput now takes a field argument; the no-arg form clears
  // every field (used by resetInlineEdits during a discard-pending).
  it("should reset per-field chip state on clearChipInput(field)", () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    w.vm.chipState.labels.input = { Name: "test" };
    w.vm.chipState.labels.search = "test";
    const prevKey = w.vm.chipState.labels.key;

    w.vm.clearChipInput("labels");

    expect(w.vm.chipState.labels.input).toBeNull();
    expect(w.vm.chipState.labels.search).toBe("");
    expect(w.vm.chipState.labels.key).toBe(prevKey + 1);
  });

  it("should reset both fields when clearChipInput is called without arguments", () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    w.vm.chipState.labels.search = "type-l";
    w.vm.chipState.albums.search = "type-a";

    w.vm.clearChipInput();

    expect(w.vm.chipState.labels.search).toBe("");
    expect(w.vm.chipState.albums.search).toBe("");
  });

  // `visibleLabels` / `visibleAlbums` filter out chips marked for removal
  // so the chip-row `v-list-item` disappears once every chip in the
  // section has been soft-removed. Without these computeds the wrapper
  // would render as an empty box above the combobox row.
  describe("visibleLabels / visibleAlbums computeds", () => {
    it("hides soft-removed labels from the visible list", () => {
      const w = mountSidebar({
        props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
        global: { stubs: { PMap: true } },
      });
      const total = w.vm.labels.length;
      expect(total).toBeGreaterThan(0);
      const firstId = w.vm.labels[0].Label.ID;
      w.vm.togglePendingChipRemoval("labels", firstId);

      expect(w.vm.visibleLabels).toHaveLength(total - 1);
      expect(w.vm.visibleLabels.some((l) => l?.Label?.ID === firstId)).toBe(false);
    });

    it("repopulates after undoChipRemovals so the wrapper comes back", () => {
      const w = mountSidebar({
        props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
        global: { stubs: { PMap: true } },
      });
      const total = w.vm.labels.length;
      w.vm.labels.forEach((l) => w.vm.togglePendingChipRemoval("labels", l?.Label?.ID));
      expect(w.vm.visibleLabels).toHaveLength(0);

      w.vm.undoChipRemovals("labels");

      expect(w.vm.visibleLabels).toHaveLength(total);
    });

    it("hides soft-removed albums from the visible list", () => {
      const w = mountSidebar({
        props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
        global: { stubs: { PMap: true } },
      });
      const total = w.vm.albums.length;
      if (total === 0) {
        return;
      } // fixture has no albums — nothing to assert
      const firstUid = w.vm.albums[0].UID;
      w.vm.togglePendingChipRemoval("albums", firstUid);

      expect(w.vm.visibleAlbums).toHaveLength(total - 1);
      expect(w.vm.visibleAlbums.some((a) => a?.UID === firstUid)).toBe(false);
    });
  });

  // The Undo icon in each chip section toolbar clears that section's
  // `chipState.<field>.removals` in a single click. Soft-removed chips
  // are filtered out by `visibleLabels` / `visibleAlbums`, so clearing
  // the removals array makes the chips reappear reactively in the v-for.
  describe("undoChipRemovals(field)", () => {
    it("clears pending labels removals but leaves albums untouched", () => {
      const w = mountSidebar({
        props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
        global: { stubs: { PMap: true } },
      });
      w.vm.chipState.labels.removals = [1, 2];
      w.vm.chipState.albums.removals = ["alb-x"];

      w.vm.undoChipRemovals("labels");

      expect(w.vm.chipState.labels.removals).toHaveLength(0);
      expect(w.vm.chipState.albums.removals).toEqual(["alb-x"]);
    });

    it("clears pending albums removals but leaves labels untouched", () => {
      const w = mountSidebar({
        props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
        global: { stubs: { PMap: true } },
      });
      w.vm.chipState.labels.removals = [1];
      w.vm.chipState.albums.removals = ["alb-x", "alb-y"];

      w.vm.undoChipRemovals("albums");

      expect(w.vm.chipState.labels.removals).toEqual([1]);
      expect(w.vm.chipState.albums.removals).toHaveLength(0);
    });

    it("is a silent no-op for an unknown field key", () => {
      const w = mountSidebar({
        props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
        global: { stubs: { PMap: true } },
      });
      w.vm.chipState.labels.removals = [1];
      expect(() => w.vm.undoChipRemovals("nope")).not.toThrow();
      expect(w.vm.chipState.labels.removals).toEqual([1]);
    });
  });

  // Auto-commit on navigation/close: when the lightbox calls
  // confirmDiscardPending() (slide change, close, keyboard navigation),
  // pending chip removals commit silently first — mirroring the inline-
  // text auto-commit on blur. The discard dialog only fires for state
  // the user could still want to keep (marker drafts, typed combobox text,
  // open Add-name confirmation).
  describe("auto-commit chip removals on confirmDiscardPending", () => {
    it("fires confirmLabels when labels have pending removals", async () => {
      const w = mountSidebar({
        props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
        global: { stubs: { PMap: true } },
      });
      const spy = vi.spyOn(w.vm, "confirmLabels");
      w.vm.chipState.labels.removals = [42];

      const result = w.vm.confirmDiscardPending();

      expect(spy).toHaveBeenCalledTimes(1);
      expect(w.vm.discardDialog.visible).toBe(false);
      await expect(result).resolves.toBe(true);
    });

    it("fires confirmAlbums when albums have pending removals", async () => {
      const w = mountSidebar({
        props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
        global: { stubs: { PMap: true } },
      });
      const spy = vi.spyOn(w.vm, "confirmAlbums");
      w.vm.chipState.albums.removals = ["alb-x"];

      const result = w.vm.confirmDiscardPending();

      expect(spy).toHaveBeenCalledTimes(1);
      expect(w.vm.discardDialog.visible).toBe(false);
      await expect(result).resolves.toBe(true);
    });

    it("skips confirmLabels / confirmAlbums when there are no pending removals", () => {
      const w = mountSidebar({
        props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
        global: { stubs: { PMap: true } },
      });
      const labelsSpy = vi.spyOn(w.vm, "confirmLabels");
      const albumsSpy = vi.spyOn(w.vm, "confirmAlbums");

      w.vm.confirmDiscardPending();

      expect(labelsSpy).not.toHaveBeenCalled();
      expect(albumsSpy).not.toHaveBeenCalled();
    });

    it("still opens the discard dialog when typed combobox text remains after auto-commit", () => {
      const w = mountSidebar({
        props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
        global: { stubs: { PMap: true } },
      });
      w.vm.chipState.labels.removals = [42];
      w.vm.chipState.labels.search = "alpha";

      w.vm.confirmDiscardPending();

      // Removals were auto-committed (cleared synchronously by confirmLabels),
      // but the typed text remained, so the dialog had to open.
      expect(w.vm.chipState.labels.removals).toHaveLength(0);
      expect(w.vm.discardDialog.visible).toBe(true);
      // Resolve so the test doesn't hang.
      w.vm.onDiscardCancel();
    });

    // Settling matched drafts up front avoids the @blur vs discard-dialog
    // race that used to persist the change even when the user clicked Discard.
    it("auto-commits a matched marker draft so the discard dialog does not appear", async () => {
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
        global: { stubs: { PMap: true }, mocks: { $config: knownPersonConfig, $util: validationUtil } },
      });
      await w.vm.$nextTick();
      w.vm.chipState.people.options = [{ UID: "sALC", Name: "Alice Smith" }];
      w.vm.setMarkerInputValue("m2", "Alice Smith");

      const result = w.vm.confirmDiscardPending();

      expect(setName).toHaveBeenCalledTimes(1);
      expect(marker.Name).toBe("Alice Smith");
      expect(marker.SubjUID).toBe("sALC");
      expect(w.vm.discardDialog.visible).toBe(false);
      await expect(result).resolves.toBe(true);
    });

    // Stacking a discard dialog on top of the Add-name dialog would
    // double-prompt for the same decision; defer instead.
    it("defers to the Add-name dialog instead of stacking a discard prompt", async () => {
      const marker = {
        UID: "m2",
        Name: "",
        SubjUID: "",
        thumbnailUrl: () => "/t/thumb2/public/tile_160",
        setName: vi.fn().mockResolvedValue(undefined),
      };
      const photo = { ...mockPhoto, getMarkers: vi.fn().mockReturnValue([marker]) };
      const w = mountSidebar({
        props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
        global: { stubs: { PMap: true }, mocks: { $config: validationConfig, $util: validationUtil } },
      });
      await w.vm.$nextTick();
      w.vm.setMarkerInputValue("m2", "Brand New Name");
      w.vm.confirmMarkerName(marker, "blur");
      expect(w.vm.addNameDialog.visible).toBe(true);

      const result = w.vm.confirmDiscardPending();

      // No discard dialog stacked on top of Add-name.
      expect(w.vm.discardDialog.visible).toBe(false);
      // Cancelling Add-name resolves navigation without a second prompt.
      w.vm.onAddNameCancel();
      await expect(result).resolves.toBe(true);
    });

    it("confirming the Add-name dialog also resolves the deferred navigation", async () => {
      const setName = vi.fn().mockResolvedValue(undefined);
      const marker = {
        UID: "m2",
        Name: "",
        SubjUID: "",
        thumbnailUrl: () => "/t/thumb2/public/tile_160",
        setName,
      };
      const photo = { ...mockPhoto, getMarkers: vi.fn().mockReturnValue([marker]) };
      const w = mountSidebar({
        props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
        global: { stubs: { PMap: true }, mocks: { $config: validationConfig, $util: validationUtil } },
      });
      await w.vm.$nextTick();
      w.vm.setMarkerInputValue("m2", "Brand New Name");
      w.vm.confirmMarkerName(marker, "blur");
      const result = w.vm.confirmDiscardPending();

      w.vm.onAddNameConfirm();

      expect(setName).toHaveBeenCalledTimes(1);
      expect(marker.Name).toBe("Brand New Name");
      await expect(result).resolves.toBe(true);
    });
  });

  // L9: onChipEscape clears the typed text and pending removals for one
  // field — independent of any inline-text editingField that might be
  // active in the sidebar.
  it("should clear search and removals on onChipEscape(field)", () => {
    const w = mountSidebar({
      props: { modelValue: mockModel, photo: mockPhoto, canEdit: true, context: contexts.Photos },
      global: { stubs: { PMap: true } },
    });
    w.vm.chipState.labels.search = "summer";
    w.vm.chipState.labels.removals = [42];
    w.vm.chipState.albums.removals = ["alb-x"];

    w.vm.onChipEscape("labels");

    expect(w.vm.chipState.labels.search).toBe("");
    expect(w.vm.chipState.labels.removals).toHaveLength(0);
    // Albums removals untouched — Esc is per-field.
    expect(w.vm.chipState.albums.removals).toEqual(["alb-x"]);
  });

  describe("restricted-role view", () => {
    // Returns a $config mock whose ACL grants deny everything when
    // `restricted` is true (simulating a share-link / visitor session
    // without library access), or allows everything when false. This
    // is the new substrate that replaces the static-array session check
    // — the sidebar now reads each gate through $config.allow().
    const restrictedConfig = (restricted) => ({
      feature: () => true,
      get: () => false,
      getSettings: () => ({ features: { edit: true, favorites: true, download: true, archive: true } }),
      allow: () => !restricted,
      deny: () => !!restricted,
      featExperimental: () => false,
      featDevelop: () => false,
      values: {},
      dir: () => "ltr",
    });

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
            $config: restrictedConfig(isRestricted),
          },
        },
      });

    it("renders permitted fields for restricted sessions", () => {
      const w = mountRestricted(true);
      const html = w.html();

      expect(w.vm.canViewLibrary).toBe(false);
      expect(html).toContain("Test Title");
      expect(html).toContain("Test Caption");
      expect(html).toContain("JPEG, 1920 × 1080, 4.2 MB");
      expect(html).toContain("52.5200°N 13.4050°E");
    });

    it("hides every restricted sidebar section for restricted sessions", () => {
      const w = mountRestricted(true);
      const html = w.html();

      // The merged file row (type + dimensions + size as the title)
      // stays visible for restricted sessions per the parallel
      // "renders permitted fields" test above; the filename subtitle
      // is gated off and the row collapses to a single line.
      expect(w.find(".meta-file").exists()).toBe(true);
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

      expect(w.vm.canViewLibrary).toBe(true);
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
            $config: restrictedConfig(true),
          },
        },
      });
      expect(w.vm.fileInfo).toBe("JPEG\u20034.0MP\u20034032\u2009\u00d7\u20093024");
      expect(thumbModel.getTypeInfo).toHaveBeenCalled();
    });

    // fileTypeName drives the file-row tooltip \u2014 Image, Raw, Video, etc.
    // The setup.js mock for $util.typeName echoes the type value for known
    // strings and falls back to defaultValue when type is empty.
    it("fileTypeName resolves the photo Type through $util.typeName", () => {
      const w = mountSidebar({
        props: {
          modelValue: mockModel,
          photo: { ...mockPhoto, Type: "raw" },
          canEdit: false,
          context: contexts.Photos,
        },
        global: { stubs: { PMap: true } },
      });
      expect(w.vm.fileTypeName).toBe("raw");
    });

    it("fileTypeName falls back to the generic File label when Type is empty", () => {
      const thumbModel = { ...mockModel, Type: "" };
      const w = mountSidebar({
        props: {
          modelValue: thumbModel,
          photo: null,
          canEdit: false,
          context: contexts.Photos,
        },
        global: { stubs: { PMap: true } },
      });
      expect(w.vm.fileTypeName).toBe("File");
    });

    // Limited-access sessions never load the full Photo; the empty
    // placeholder Photo.getDefaults() returns is seeded Type="image" and
    // would mask the Thumb's real type (e.g. live) without the UID gate.
    it("mediaType / fileTypeName / fileIcon fall through to Thumb.Type when Photo is the empty placeholder", () => {
      const thumbModel = { ...mockModel, Type: "live" };
      const emptyPhoto = { UID: "", Type: "image", getMarkers: vi.fn().mockReturnValue([]) };
      const w = mountSidebar({
        props: {
          modelValue: thumbModel,
          photo: emptyPhoto,
          canEdit: false,
          context: contexts.Photos,
        },
        global: {
          stubs: { PMap: true },
          mocks: { $config: restrictedConfig(true) },
        },
      });
      expect(w.vm.mediaType).toBe("live");
      expect(w.vm.fileTypeName).toBe("live");
      expect(w.vm.fileIcon).toBe("mdi-play-circle-outline");
    });
  });

  // Exhaustive matrix: against a fully-populated photo and a photo
  // without any metadata, assert which fields the sidebar renders and
  // which edit affordances it exposes for each of the two equivalence
  // classes that drive the gating today: "library access granted" vs
  // "library access denied" (share-link visitors, guests, etc.).
  //
  // Each gate now flows through $config.allow(resource, perm). The
  // matrix mocks the ACL by injecting a $config whose grants flip with
  // the `restricted` flag, mirroring what the backend would send in
  // the client config payload (internal/config/client_config.go).
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
        getLatLng: vi.fn().mockReturnValue("52.5200°N 13.4050°E"),
        getLatLngShort: vi.fn().mockReturnValue("52.5200°N 13.4050°E"),
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
        UID: "ps6sg6be2lvl0y12",
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
        PlaceLabel: TEXT.placeName,
        Country: "de",
        placeName: vi.fn().mockReturnValue(TEXT.placeName),
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

    // matrixConfig builds a $config mock whose ACL grants follow the
    // restricted flag. allow() returns true for every (resource, perm)
    // pair when unrestricted; deny() is the boolean inverse so the
    // fetchPhoto / preloadNextPhoto gates also resolve correctly. This
    // is the substrate that replaced the static-array session check —
    // it lets the visibility matrix below pivot on one boolean without
    // re-implementing the role table.
    const matrixConfig = (restricted) => ({
      feature: () => true,
      get: () => false,
      getSettings: () => ({ features: { edit: true, favorites: true, download: true, archive: true } }),
      allow: () => !restricted,
      deny: () => !!restricted,
      featExperimental: () => false,
      featDevelop: () => false,
      values: {},
      dir: () => "ltr",
    });

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
            $config: matrixConfig(restricted),
            $session: {
              isAnonymous: () => !!anonymous,
            },
          },
        },
      });
    }

    describe("editable session (admin/user)", () => {
      let w;
      beforeEach(() => {
        w = mountFor({ anonymous: false, restricted: false, editable: true }, { withMetadata: true });
      });

      it("renders every metadata section and its content", () => {
        expect(w.vm.canViewLibrary).toBe(true);
        expect(w.vm.isEditable).toBe(true);
        const html = w.html();
        for (const needle of [
          TEXT.title,
          TEXT.caption,
          "JPEG, 1920 x 1080, 4.2 MB",
          "52.5200°N 13.4050°E",
          TEXT.filename,
          TEXT.camera,
          TEXT.lens,
          TEXT.placeName,
          TEXT.peopleHeader,
          TEXT.labelsHeader,
          TEXT.albumsHeader,
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
        // Editable users see the merged file row with both the file
        // info (title: type + dimensions + size) and the filename
        // (subtitle: path).
        const fileRow = w.find(".meta-file");
        expect(fileRow.exists()).toBe(true);
        expect(fileRow.text()).toContain(TEXT.filename);
        // Keywords/Notes are now part of the single-row icon-prepended
        // detailsFields layout (no more text-subtitle-2 header rows).
        // Confirm the rows render with their prepend icons and value
        // content under the canonical .meta-{key} selector.
        const keywordsRow = w.find(".meta-keywords");
        expect(keywordsRow.exists()).toBe(true);
        expect(keywordsRow.text()).toContain(TEXT.keywords);
        const notesRow = w.find(".meta-notes");
        expect(notesRow.exists()).toBe(true);
        expect(notesRow.html()).toContain(TEXT.notes);
      });

      it("renders inline pencils and only the Edit Faces toggle (no eye toggle for editable users)", () => {
        expect(w.findAll(".meta-inline-pencil").length).toBeGreaterThanOrEqual(10);
        expect(w.find(".meta-faces-edit").exists()).toBe(true);
        // Editable users see only the pencil toggle — edit mode is a
        // strict superset of display mode for them. The pencil and the eye
        // share `.meta-markers-toggle` (the visibility-toggle role);
        // `.meta-faces-display` is the eye-only discriminator.
        expect(w.find(".meta-faces-display").exists()).toBe(false);
      });
    });

    describe("restricted session (guest/visitor/contributor/share-link)", () => {
      let w;
      beforeEach(() => {
        w = mountFor({ anonymous: false, restricted: true, editable: false }, { withMetadata: true });
      });

      it("renders only the shared fields and hides every restricted section", () => {
        expect(w.vm.canViewLibrary).toBe(false);
        expect(w.vm.isEditable).toBe(false);
        const html = w.html();
        // Allow-list.
        expect(html).toContain(TEXT.title);
        expect(html).toContain(TEXT.caption);
        expect(html).toContain("JPEG, 1920 x 1080, 4.2 MB");
        expect(html).toContain("52.5200°N 13.4050°E");
        // The merged file row's title (type + dimensions + size) is
        // shared with restricted sessions; the filename subtitle is
        // gated behind `canViewLibrary` and must not appear here.
        const fileRow = w.find(".meta-file");
        expect(fileRow.exists()).toBe(true);
        // Deny-list.
        expect(fileRow.text()).not.toContain(TEXT.filename);
        for (const needle of [
          TEXT.filename,
          TEXT.camera,
          TEXT.lens,
          TEXT.placeName,
          TEXT.altitude,
          TEXT.peopleHeader,
          TEXT.labelsHeader,
          TEXT.albumsHeader,
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
        expect(w.find(".meta-faces-edit").exists()).toBe(false);
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
        for (const needle of [TEXT.camera, TEXT.lens, TEXT.placeName, TEXT.peopleHeader, TEXT.labelsHeader]) {
          expect(html).not.toContain(needle);
        }
        expect(w.find(".meta-inline-pencil").exists()).toBe(false);
      });

      it("restricted session renders the minimal shell only", () => {
        const w = mountFor({ anonymous: false, restricted: true, editable: false }, { withMetadata: false });
        expect(w.vm.canViewLibrary).toBe(false);
        expect(w.find(".meta-inline-pencil").exists()).toBe(false);
        expect(w.find(".meta-markers-toggle").exists()).toBe(false);
      });
    });

    // Parent-driven editor (canEdit=true, Details present, library
    // access granted, empty fields) must expose "add prompt"
    // affordances so admins can populate metadata from scratch.
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
            $config: matrixConfig(false),
            $session: { isAnonymous: () => false },
          },
        },
      });
      expect(w.vm.isEditable).toBe(true);
      // Add-prompt spans are the "click to start editing" placeholders.
      const prompts = w.findAll(".meta-add-prompt");
      expect(prompts.length).toBeGreaterThanOrEqual(5);
      // Structural assertion for the merged details rows: each empty
      // field renders exactly one add-prompt under its canonical
      // .meta-{key} row. Avoids coupling to the exact label copy
      // (which UX iterates on per field).
      for (const key of ["subject", "copyright", "artist", "license", "keywords", "notes"]) {
        const row = w.find(`.meta-${key}`);
        expect(row.exists(), `missing .meta-${key} row`).toBe(true);
        expect(row.find(".meta-add-prompt").exists(), `missing add-prompt in .meta-${key}`).toBe(true);
      }
      // Title and Caption rows don't carry a per-key class on the row
      // itself (legacy hand-coded structure); fall back to label match
      // for those, since "Add a Title" / "Add a Caption" are stable.
      const texts = prompts.map((p) => p.text());
      expect(texts).toContain("Add a Title");
      expect(texts).toContain("Add a Caption");
    });

    // Explicit share-link (anonymous) case: the sidebar must behave
    // identically to any other library-access-denied session even
    // though the backing User record is empty. The backend ships an
    // ACL with no library-access grant for anonymous sessions, which
    // is what matrixConfig(true) simulates here.
    it("treats anonymous share-link sessions as library-denied even with canEdit=true", () => {
      const w = mountSidebar({
        props: {
          modelValue: buildModel({ withMetadata: true }),
          photo: buildPhoto({ withMetadata: true }),
          // Simulate a lightbox that forgot to flip canEdit off: the
          // sidebar must still drop all edit affordances, driven by
          // canViewLibrary alone.
          canEdit: true,
          context: contexts.Photos,
        },
        global: {
          stubs: { PMap: true },
          mocks: {
            $config: matrixConfig(true),
            $session: { isAnonymous: () => true },
          },
        },
      });
      expect(w.vm.canViewLibrary).toBe(false);
      expect(w.vm.isEditable).toBe(false);
      expect(w.find(".meta-inline-pencil").exists()).toBe(false);
      expect(w.find(".meta-markers-toggle").exists()).toBe(false);
      expect(w.find(".meta-faces-edit").exists()).toBe(false);
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
            $session: { isAnonymous: () => false },
          },
        },
      });
      const html = w.html();
      expect(html).not.toContain(TEXT.peopleHeader);
      expect(w.find(".meta-markers-toggle").exists()).toBe(false);
      expect(w.find(".meta-faces-edit").exists()).toBe(false);
    });
  });

  // Pins the template bindings on <p-meta-datetime-dialog>, <p-meta-camera-dialog>,
  // and <p-meta-location-dialog>. Commit 7a611b6d6 removed the legacy `photo`
  // prop on PLightboxSidebar and routed parent state through `$view.getData()`
  // (exposed via the `photo` computed). The three dialog tags in the sidebar
  // template kept referencing the now-undefined `photo` identifier, so the
  // dialogs received `:photo="undefined"` / `:latlng="[0, 0]"` and their
  // loadFromPhoto() hooks bailed out — opening every sidebar dialog with
  // empty inputs even when the photo had values. These tests fail if the
  // bindings ever regress to a name that does not exist on the component.
  describe("dialog photo prop bindings", () => {
    const dialogStubs = {
      PMetaDatetimeDialog: { name: "PMetaDatetimeDialog", template: "<div />", props: ["visible", "photo"] },
      PMetaCameraDialog: { name: "PMetaCameraDialog", template: "<div />", props: ["visible", "photo"] },
      PMetaLocationDialog: { name: "PMetaLocationDialog", template: "<div />", props: ["visible", "latlng"] },
    };

    function mountWithDialogStubs(photo) {
      return mountSidebar({
        props: { modelValue: mockModel, photo, canEdit: true, context: contexts.Photos },
        global: { stubs: { PMap: true, ...dialogStubs } },
      });
    }

    it("passes the current photo to <p-meta-datetime-dialog>", () => {
      const w = mountWithDialogStubs(mockPhoto);
      const dialog = w.findComponent({ name: "PMetaDatetimeDialog" });
      expect(dialog.exists()).toBe(true);
      // Vue 3 wraps `view` (data()) in a reactive proxy, so the dialog
      // sees a proxy of mockPhoto rather than the raw reference. A
      // deep-equality check still pins the regression: the broken
      // binding produced `undefined`, not a structurally-equal proxy.
      expect(dialog.props("photo")).toEqual(mockPhoto);
    });

    it("passes the current photo to <p-meta-camera-dialog>", () => {
      const w = mountWithDialogStubs(mockPhoto);
      const dialog = w.findComponent({ name: "PMetaCameraDialog" });
      expect(dialog.exists()).toBe(true);
      expect(dialog.props("photo")).toEqual(mockPhoto);
    });

    it("derives <p-meta-location-dialog> latlng from the current photo", () => {
      const photo = { ...mockPhoto, Lat: 52.52, Lng: 13.405 };
      const w = mountWithDialogStubs(photo);
      const dialog = w.findComponent({ name: "PMetaLocationDialog" });
      expect(dialog.exists()).toBe(true);
      expect(dialog.props("latlng")).toEqual([52.52, 13.405]);
    });

    it("falls back to [0, 0] for <p-meta-location-dialog> when the photo has no coordinates", () => {
      const w = mountWithDialogStubs(mockPhoto);
      const dialog = w.findComponent({ name: "PMetaLocationDialog" });
      expect(dialog.exists()).toBe(true);
      expect(dialog.props("latlng")).toEqual([0, 0]);
    });
  });
});
