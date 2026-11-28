import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { shallowMount, mount } from "@vue/test-utils";
import { nextTick } from "vue";
import PTabPhotoDetails from "component/photo/edit/details.vue";
import { Photo } from "model/photo";
import { Mock } from "../../../fixtures";

// Mock model factory
const createMockModel = (overrides = {}) => ({
  hasId: () => true,
  wasChanged: () => false,
  getDateTime: () => ({ toFormat: () => "12:34:56" }),
  localDate: vi.fn().mockReturnValue({
    isValid: true,
    toFormat: (fmt) => {
      if (fmt === "d") {
        return "1";
      }
      if (fmt === "L") {
        return "4";
      }
      if (fmt === "y") {
        return "2024";
      }
      return "2024-04-01T11:29:54";
    },
    toISO: () => "2024-04-01T11:29:54",
    toUTC: () => ({ toISO: () => "2024-04-01T11:29:54Z" }),
    day: 1,
    month: 4,
    year: 2024,
  }),
  // setLocalTime mirrors Photo.setLocalTime closely enough for the component, which only reads its result.
  setLocalTime: vi.fn(function (time) {
    return this.localDate(time);
  }),
  changeTimeZone: vi.fn(),
  currentTimeZoneUTC: () => false,
  timeIsUTC: () => false,
  Title: "Example",
  TitleSrc: "manual",
  Caption: "",
  CaptionSrc: "",
  Day: 30,
  Month: 4,
  Year: 2024,
  TakenAtLocal: "2024-04-30T11:29:54Z",
  TakenAt: "2024-04-30T11:29:54Z",
  TakenSrc: "",
  TimeZone: "UTC",
  Lat: 0,
  Lng: 0,
  Country: "",
  PlaceSrc: "",
  Place: { PlaceID: "", Label: "" },
  CameraID: 1,
  CameraSrc: "",
  LensID: 1,
  Iso: "",
  Exposure: "",
  FNumber: "",
  FocalLength: "",
  Altitude: "",
  Details: {
    Subject: "",
    SubjectSrc: "",
    Keywords: "",
    KeywordsSrc: "",
    Notes: "",
    NotesSrc: "",
    License: "",
    LicenseSrc: "",
    Artist: "",
    ArtistSrc: "",
    Copyright: "",
    CopyrightSrc: "",
  },
  Quality: 3,
  update: vi.fn().mockResolvedValue({}),
  thumbnailUrl: () => "/thumb/tile_500/foo.jpg",
  ...overrides,
});

// Stubs and global mocks
const makeWrapper = (overrides = {}) => {
  const mockModel = createMockModel(overrides.model || {});

  return shallowMount(PTabPhotoDetails, {
    props: { uid: "p123" },
    global: {
      mocks: {
        $gettext: (s) => s,
        $pgettext: (_c, s) => s,
        $isRtl: false,
        $notify: { error: vi.fn(), success: vi.fn() },
        $lightbox: { openModels: vi.fn() },
        $vuetify: { display: { xs: false } },
        $config: {
          feature: vi.fn().mockImplementation((f) => {
            if (f === "edit") {
              return true;
            }
            if (f === "review") {
              return true;
            }
            if (f === "places") {
              return true;
            }
            return false;
          }),
          get: vi.fn().mockImplementation((k) => {
            if (k === "clip") {
              return 255;
            }
            if (k === "readonly") {
              return false;
            }
            return "";
          }),
          values: {
            cameras: [{ ID: 1, Name: "Canon EOS R5" }],
            lenses: [{ ID: 1, Name: "Canon RF 24-70mm" }],
          },
        },
        $view: {
          getData: () => ({ model: mockModel }),
        },
      },
      stubs: {
        VForm: { template: "<form><slot /></form>" },
        VRow: { template: "<div><slot /></div>" },
        VCol: { template: "<div><slot /></div>" },
        VTextField: { template: "<input />", props: ["modelValue", "appendInnerIcon", "disabled", "rules", "label"] },
        VTextarea: { template: "<textarea></textarea>" },
        VAutocomplete: { template: "<select></select>" },
        VSelect: { template: "<select></select>" },
        VBtn: { template: "<button><slot /></button>" },
        PMetaLocationInput: { template: '<div class="p-meta-location-input"></div>' },
        PMetaLocationDialog: { template: '<div class="p-meta-location-dialog"></div>' },
      },
    },
    ...overrides,
  });
};

describe("component/photo/edit/details", () => {
  let wrapper;

  beforeEach(() => {
    wrapper = makeWrapper();
  });

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount();
    }
  });

  describe("initialization and data sync", () => {
    it("renders and syncs initial time", () => {
      expect(wrapper.vm.time).toBe("12:34:56");
    });

    it("syncs location label from Place data", () => {
      const wrapperWithPlace = makeWrapper({
        model: {
          Place: { PlaceID: "abc123", Label: "San Francisco, CA" },
        },
      });

      wrapperWithPlace.vm.syncLocation();
      expect(wrapperWithPlace.vm.locationLabel).toBe("San Francisco, CA");
      wrapperWithPlace.unmount();
    });

    it("uses default location label when no place data", () => {
      wrapper.vm.syncLocation();
      expect(wrapper.vm.locationLabel).toBe("Location");
    });

    it("initializes computed properties correctly", () => {
      expect(wrapper.vm.cameraOptions).toEqual([{ ID: 1, Name: "Canon EOS R5" }]);
      expect(wrapper.vm.lensOptions).toEqual([{ ID: 1, Name: "Canon RF 24-70mm" }]);
    });

    it("detects review mode correctly", () => {
      const lowQualityWrapper = makeWrapper({ model: { Quality: 2 } });
      expect(lowQualityWrapper.vm.inReview).toBe(true);
      lowQualityWrapper.unmount();

      const highQualityWrapper = makeWrapper({ model: { Quality: 4 } });
      expect(highQualityWrapper.vm.inReview).toBe(false);
      highQualityWrapper.unmount();
    });
  });

  describe("date handling", () => {
    it("clamps day when out of range for effective month/year", () => {
      // Set February with Day=30 should clamp to 29 (leap year 2024)
      wrapper.vm.view.model.Year = 2024;
      wrapper.vm.view.model.Month = 2;
      wrapper.vm.view.model.Day = 30;
      wrapper.vm.clampDayToValidRange();
      expect(wrapper.vm.view.model.Day).toBe(29);
    });

    it("handles setDay with object value", () => {
      const syncTimeSpy = vi.spyOn(wrapper.vm, "syncTime");

      wrapper.vm.setDay({ value: 15 });
      expect(wrapper.vm.view.model.Day).toBe(15);
      expect(syncTimeSpy).toHaveBeenCalled();
    });

    it("handles setDay with null (unknown)", () => {
      wrapper.vm.setDay(null);
      expect(wrapper.vm.view.model.Day).toBe(-1);
      expect(wrapper.vm.view.model.Year).toBe(-1);
    });

    it("handles setDay with numeric string", () => {
      vi.spyOn(wrapper.vm.rules, "isNumberRange").mockReturnValue(true);
      wrapper.vm.setDay("25");
      expect(wrapper.vm.view.model.Day).toBe(25);
    });

    it("handles setMonth with object value", () => {
      wrapper.vm.setMonth({ value: 12 });
      expect(wrapper.vm.view.model.Month).toBe(12);
    });

    it("handles setMonth with null (unknown)", () => {
      wrapper.vm.setMonth(null);
      expect(wrapper.vm.view.model.Month).toBe(-1);
      expect(wrapper.vm.view.model.Year).toBe(-1);
    });

    it("handles setYear with object value", () => {
      wrapper.vm.setYear({ value: 2023 });
      expect(wrapper.vm.view.model.Year).toBe(2023);
    });

    it("handles setYear with null (unknown)", () => {
      wrapper.vm.setYear(null);
      expect(wrapper.vm.view.model.Year).toBe(-1);
    });

    it("calculates effective year from model or TakenAtLocal", () => {
      // With explicit Year
      wrapper.vm.view.model.Year = 2023;
      expect(wrapper.vm.effectiveYear()).toBe(2023);

      // Without explicit Year, should extract from TakenAtLocal
      wrapper.vm.view.model.Year = -1;
      wrapper.vm.view.model.TakenAtLocal = "2022-06-15T10:30:00Z";
      expect(wrapper.vm.effectiveYear()).toBe(2022);
    });

    it("calculates effective month from model or TakenAtLocal", () => {
      // With explicit Month
      wrapper.vm.view.model.Month = 8;
      expect(wrapper.vm.effectiveMonth()).toBe(8);

      // Without explicit Month, should extract from TakenAtLocal
      wrapper.vm.view.model.Month = -1;
      wrapper.vm.view.model.TakenAtLocal = "2022-06-15T10:30:00Z";
      expect(wrapper.vm.effectiveMonth()).toBe(6);
    });
  });

  describe("time handling", () => {
    it("validates and sets time", () => {
      vi.spyOn(wrapper.vm.rules, "isTime").mockReturnValue(true);
      const updateModelSpy = vi.spyOn(wrapper.vm, "updateModel");

      wrapper.vm.time = "14:30:00";
      wrapper.vm.setTime();
      expect(updateModelSpy).toHaveBeenCalled();
    });

    it("does not update model with invalid time", () => {
      vi.spyOn(wrapper.vm.rules, "isTime").mockReturnValue(false);
      const updateModelSpy = vi.spyOn(wrapper.vm, "updateModel");

      wrapper.vm.time = "25:99:99";
      wrapper.vm.setTime();
      expect(updateModelSpy).not.toHaveBeenCalled();
    });

    it("updates model with valid local date", () => {
      const mockLocalDate = {
        isValid: true,
        toFormat: vi.fn().mockImplementation((fmt) => {
          if (fmt === "d") {
            return "15";
          }
          if (fmt === "L") {
            return "6";
          }
          if (fmt === "y") {
            return "2023";
          }
          return "2023-06-15";
        }),
        toISO: () => "2023-06-15T14:30:00",
        toUTC: () => ({ toISO: () => "2023-06-15T14:30:00Z" }),
        day: 15,
        month: 6,
        year: 2023,
      };

      wrapper.vm.view.model.localDate.mockReturnValue(mockLocalDate);
      wrapper.vm.view.model.Day = 0;
      wrapper.vm.view.model.Month = 0;
      wrapper.vm.view.model.Year = 0;

      wrapper.vm.updateModel();

      expect(wrapper.vm.view.model.setLocalTime).toHaveBeenCalledWith(wrapper.vm.time);
      expect(wrapper.vm.view.model.Day).toBe(15);
      expect(wrapper.vm.view.model.Month).toBe(6);
      expect(wrapper.vm.view.model.Year).toBe(2023);
    });

    it("sets invalidDate flag with invalid date", () => {
      const mockInvalidDate = { isValid: false };
      wrapper.vm.view.model.localDate.mockReturnValue(mockInvalidDate);

      wrapper.vm.updateModel();
      expect(wrapper.vm.invalidDate).toBe(true);
    });
  });

  describe("location handling", () => {
    it("updates latitude and longitude", () => {
      wrapper.vm.updateLatLng([37.7749, -122.4194]);

      expect(wrapper.vm.view.model.Lat).toBe(37.7749);
      expect(wrapper.vm.view.model.Lng).toBe(-122.4194);
      expect(wrapper.vm.view.model.PlaceSrc).toBe("manual");
    });

    it("handles location change with country data", () => {
      const locationData = {
        location: {
          country: "US",
          place: { label: "San Francisco, California" },
        },
      };

      wrapper.vm.onLocationChanged(locationData);

      expect(wrapper.vm.view.model.Country).toBe("US");
      expect(wrapper.vm.locationLabel).toBe("San Francisco, California");
    });

    it("resets location label when no place data", () => {
      wrapper.vm.onLocationChanged({});
      expect(wrapper.vm.locationLabel).toBe("Location");
    });

    it("opens and closes location dialog", () => {
      expect(wrapper.vm.locationDialog).toBe(false);

      wrapper.vm.adjustLocation();
      expect(wrapper.vm.locationDialog).toBe(true);
    });

    it("confirms location from dialog", () => {
      const updateLatLngSpy = vi.spyOn(wrapper.vm, "updateLatLng");
      const onLocationChangedSpy = vi.spyOn(wrapper.vm, "onLocationChanged");

      const locationData = { lat: 40.7128, lng: -74.006 };
      wrapper.vm.confirmLocation(locationData);

      expect(updateLatLngSpy).toHaveBeenCalledWith([40.7128, -74.006]);
      expect(onLocationChangedSpy).toHaveBeenCalledWith(locationData);
      expect(wrapper.vm.locationDialog).toBe(false);
    });

    it("keeps Country VAutocomplete editable when Lat/Lng are present", async () => {
      // Use full mount with real VAutocomplete (not stubbed) so we can assert props
      const w = mount(PTabPhotoDetails, {
        props: { uid: "p123" },
        global: {
          mocks: {
            $gettext: (s) => s,
            $pgettext: (_c, s) => s,
            $isRtl: false,
            $notify: { error: vi.fn(), success: vi.fn() },
            $lightbox: { openModels: vi.fn() },
            $vuetify: { display: { xs: false } },
            $config: {
              feature: vi.fn().mockImplementation((f) => (f === "edit" ? true : false)),
              get: vi.fn().mockImplementation((k) => (k === "clip" ? 255 : "")),
              values: { cameras: [], lenses: [] },
            },
            $view: {
              getData: () => ({ model: createMockModel() }),
            },
          },
          stubs: {
            VAutocomplete: false,
            PMetaLocationInput: { template: '<div class="p-meta-location-input"></div>' },
            PMetaLocationDialog: { template: '<div class="p-meta-location-dialog"></div>' },
          },
        },
      });

      await nextTick();
      const allAutocompletes = w.findAllComponents({ name: "VAutocomplete" });
      const countryAuto = allAutocompletes.find((c) => c.classes().includes("input-country"));
      expect(countryAuto).toBeTruthy();
      expect(countryAuto.props("readonly")).not.toBe(true);

      // Country can be corrected manually without clearing the coordinates.
      w.vm.view.model.Lat = 37.5;
      w.vm.view.model.Lng = -122.4;
      await nextTick();
      expect(countryAuto.props("readonly")).not.toBe(true);

      w.unmount();
    });
  });

  describe("saving and validation", () => {
    it("prevents save with invalid date", async () => {
      wrapper.vm.invalidDate = true;
      await wrapper.vm.save(false);
      expect(wrapper.vm.$notify.error).toHaveBeenCalledWith("Invalid date");
      expect(wrapper.vm.view.model.update).not.toHaveBeenCalled();
    });

    it("saves successfully with valid data", async () => {
      wrapper.vm.invalidDate = false;
      wrapper.vm.dateChanged = true;
      const updateModelSpy = vi.spyOn(wrapper.vm, "updateModel");
      const syncDataSpy = vi.spyOn(wrapper.vm, "syncData");

      await wrapper.vm.save(false);

      expect(updateModelSpy).toHaveBeenCalled();
      expect(wrapper.vm.view.model.update).toHaveBeenCalled();
      expect(syncDataSpy).toHaveBeenCalled();
    });

    it("emits close event when saving with close flag", async () => {
      wrapper.vm.invalidDate = false;
      await wrapper.vm.save(true);
      expect(wrapper.emitted("close")).toBeTruthy();
    });

    it("emits close event when close method called", () => {
      wrapper.vm.close();
      expect(wrapper.emitted("close")).toBeTruthy();
    });

    // Vue 3's component proxy intercepts $refs reads; standard assignment
    // doesn't stick. The internal instance (vm.$) exposes the underlying
    // `refs` object that the proxy reads from, so injecting the mock
    // there lets `this.$refs.form.validate` resolve to the spy when
    // save() runs.
    const overrideFormRef = (vm, validate) => {
      vm.$.refs.form = { validate };
    };

    it("blocks save and notifies when form validation fails", async () => {
      wrapper.vm.invalidDate = false;
      const validate = vi.fn().mockResolvedValue({ valid: false });
      overrideFormRef(wrapper.vm, validate);

      await wrapper.vm.save(false);

      expect(validate).toHaveBeenCalled();
      expect(wrapper.vm.view.model.update).not.toHaveBeenCalled();
      expect(wrapper.vm.$notify.error).toHaveBeenCalledWith("Changes could not be saved");
    });

    it("proceeds with save when form validation passes", async () => {
      wrapper.vm.invalidDate = false;
      const validate = vi.fn().mockResolvedValue({ valid: true });
      overrideFormRef(wrapper.vm, validate);

      await wrapper.vm.save(false);

      expect(validate).toHaveBeenCalled();
      expect(wrapper.vm.view.model.update).toHaveBeenCalled();
    });

    it("validates text length via the centralized rules.text factory", () => {
      // After migrating the per-component textRule to the shared
      // common/form rules.text(...) factory, each :rules attribute
      // calls the factory inline. The component exposes `rules` on
      // its instance so call sites can invoke it from the template.
      const [, maxLenRule] = wrapper.vm.rules.text(false, 0, 255, "Title");

      // Too long → label-specific localized message.
      expect(maxLenRule("a".repeat(300))).toBe("Title is too long");

      // Short input → passes.
      expect(maxLenRule("Valid text")).toBe(true);

      // Defensive: null / undefined / object inputs don't crash, the
      // factory's maxLen short-circuits on non-string input.
      expect(maxLenRule(null)).toBe(true);
      expect(maxLenRule(undefined)).toBe(true);
      expect(maxLenRule({ Name: "obj" })).toBe(true);
    });

    // Per-field caps must come from PhotoMaxLength (backend VARCHAR), not
    // the historical $config.get('clip') = 160 ceiling. These cases lock
    // each field's exposure to its real backend cap so a future bare
    // `clip` regression fails loudly here.
    it("exposes PhotoMaxLength and wires the per-field caps", () => {
      const m = wrapper.vm.PhotoMaxLength;
      expect(m).toEqual({
        Title: 200,
        Caption: 4096,
        Subject: 1024,
        Artist: 1024,
        Copyright: 1024,
        License: 1024,
        Keywords: 2048,
        Notes: 2048,
        Exposure: 64,
      });

      const cases = [
        ["Title", "Title", m.Title],
        ["Caption", "Caption", m.Caption],
        ["Subject", "Subject", m.Subject],
        ["Copyright", "Copyright", m.Copyright],
        ["Artist", "Artist", m.Artist],
        ["License", "License", m.License],
        ["Keywords", "Keywords", m.Keywords],
        ["Notes", "Notes", m.Notes],
      ];

      for (const [label, errorLabel, cap] of cases) {
        const [, rule] = wrapper.vm.rules.text(false, 0, cap, label);
        // At the cap → passes (200 valid for Title, 4096 for Caption, …).
        expect(rule("a".repeat(cap))).toBe(true);
        // One char beyond → label-specific error.
        expect(rule("a".repeat(cap + 1))).toBe(`${errorLabel} is too long`);
      }
    });
  });

  describe("photo viewing", () => {
    it("opens photo in lightbox", () => {
      wrapper.vm.openPhoto();
      expect(wrapper.vm.$lightbox.openModels).toHaveBeenCalled();
    });
  });

  describe("time zone changes", () => {
    // loadCapture binds a real photo model to the editor; the PUT mock merges each request into the last saved values.
    const loadCapture = (values = {}) => {
      const photo = new Photo({
        UID: "ps6sg6be2lvl0yh7",
        Caption: "",
        CaptionSrc: "",
        TakenAt: "2025-06-30T23:30:00Z",
        TakenAtLocal: "2025-06-30T23:30:00Z",
        TakenSrc: "manual",
        TimeZone: "UTC",
        Year: 2025,
        Month: 6,
        Day: 30,
        Details: {},
        ...values,
      });
      let saved = photo.getValues();
      wrapper.vm.view.model = photo;
      wrapper.vm.syncData();
      Mock.resetHistory();
      Mock.onPut(photo.getEntityResource()).reply((request) => {
        saved = { ...saved, ...JSON.parse(request.data) };
        return [200, saved];
      });
      return photo;
    };

    // lastBody returns the JSON body of the most recent PUT request.
    const lastBody = () => JSON.parse(Mock.history.put.at(-1).data);

    it("applies a pending manual date before converting from UTC", async () => {
      const photo = loadCapture();
      wrapper.vm.setDay(29);
      wrapper.vm.setTimeZone("Europe/Berlin");
      await wrapper.vm.save(false);
      expect(photo.TakenAt).toBe("2025-06-29T23:30:00Z");
      expect(photo.TakenAtLocal).toBe("2025-06-30T01:30:00Z");
      expect([photo.Year, photo.Month, photo.Day]).toEqual([2025, 6, 30]);
    });

    it("keeps a manual local time when the zone changes again", () => {
      const photo = loadCapture();
      wrapper.vm.setTimeZone("Europe/Berlin");
      wrapper.vm.time = "03:45:00";
      wrapper.vm.setTime();
      wrapper.vm.setTimeZone("America/New_York");
      expect(photo.TakenAtLocal).toBe("2025-07-01T03:45:00Z");
      expect(photo.TakenAt).toBe("2025-07-01T07:45:00Z");
      expect(wrapper.vm.time).toBe("03:45:00");
    });

    it("applies manual local date and time after a zone change", async () => {
      const photo = loadCapture();
      wrapper.vm.setTimeZone("Europe/Berlin");
      wrapper.vm.setDay(2);
      wrapper.vm.time = "03:45:00";
      wrapper.vm.setTime();
      await wrapper.vm.save(false);
      expect(photo.TakenAt).toBe("2025-07-02T01:45:00Z");
      expect(photo.TakenAtLocal).toBe("2025-07-02T03:45:00Z");
      expect([photo.Year, photo.Month, photo.Day]).toEqual([2025, 7, 2]);
    });

    it("keeps unknown components and the timestamp on a caption save", async () => {
      const photo = loadCapture({ Year: -1, Month: -1, Day: -1 });
      photo.Caption = "Updated caption";
      await wrapper.vm.save(false);
      expect(lastBody()).toEqual({ Caption: "Updated caption", CaptionSrc: "manual" });
      expect(photo.TakenAt).toBe("2025-06-30T23:30:00Z");
      expect(photo.TakenAtLocal).toBe("2025-06-30T23:30:00Z");
      expect([photo.Year, photo.Month, photo.Day]).toEqual([-1, -1, -1]);
    });

    it("keeps a known month and year when the day is unknown", async () => {
      const photo = loadCapture({ TakenAt: "2020-01-01T03:00:00Z", TakenAtLocal: "2020-01-01T03:00:00Z", Year: 2020, Month: 1, Day: -1 });
      wrapper.vm.setTimeZone("America/New_York");
      await wrapper.vm.save(false);
      expect(lastBody()).toEqual({ TakenAtLocal: "2019-12-31T22:00:00Z", TimeZone: "America/New_York", TakenSrc: "manual" });
      expect([photo.Year, photo.Month, photo.Day]).toEqual([2020, 1, -1]);
    });

    const cases = [
      // zone, target, taken, local, request body, saved TakenAt, saved TakenAtLocal, Y/M/D
      [
        "UTC",
        "Europe/Berlin",
        "2025-06-30T23:30:00Z",
        "2025-06-30T23:30:00Z",
        { TakenAtLocal: "2025-07-01T01:30:00Z", TimeZone: "Europe/Berlin", Month: 7, Day: 1 },
        "2025-06-30T23:30:00Z",
        "2025-07-01T01:30:00Z",
        [2025, 7, 1],
      ],
      [
        "UTC",
        "UTC-12",
        "2020-01-01T00:30:00Z",
        "2020-01-01T00:30:00Z",
        { TakenAtLocal: "2019-12-31T12:30:00Z", TimeZone: "UTC-12", Year: 2019, Month: 12, Day: 31 },
        "2020-01-01T00:30:00Z",
        "2019-12-31T12:30:00Z",
        [2019, 12, 31],
      ],
      [
        "UTC-12",
        "UTC",
        "2020-01-01T11:45:00Z",
        "2019-12-31T23:45:00Z",
        { TakenAt: "2019-12-31T23:45:00Z", TimeZone: "UTC" },
        "2019-12-31T23:45:00Z",
        "2019-12-31T23:45:00Z",
        [2019, 12, 31],
      ],
      [
        "Europe/Berlin",
        "America/New_York",
        "2025-06-30T12:00:00Z",
        "2025-06-30T14:00:00Z",
        { TakenAt: "2025-06-30T18:00:00Z", TimeZone: "America/New_York" },
        "2025-06-30T18:00:00Z",
        "2025-06-30T14:00:00Z",
        [2025, 6, 30],
      ],
      [
        "Local",
        "Europe/Berlin",
        "2025-06-30T23:30:00Z",
        "2025-06-30T23:30:00Z",
        { TakenAt: "2025-06-30T21:30:00Z", TimeZone: "Europe/Berlin" },
        "2025-06-30T21:30:00Z",
        "2025-06-30T23:30:00Z",
        [2025, 6, 30],
      ],
      [
        "Europe/Berlin",
        "Local",
        "2025-06-30T21:30:00Z",
        "2025-06-30T23:30:00Z",
        { TakenAt: "2025-06-30T23:30:00Z", TimeZone: "Local" },
        "2025-06-30T23:30:00Z",
        "2025-06-30T23:30:00Z",
        [2025, 6, 30],
      ],
    ];

    it.each(cases)("changes %s to %s and keeps the result across saves", async (zone, target, taken, local, body, savedTaken, savedLocal, ymd) => {
      const photo = loadCapture({
        TakenAt: taken,
        TakenAtLocal: local,
        TimeZone: zone,
        Year: Number(local.slice(0, 4)),
        Month: Number(local.slice(5, 7)),
        Day: Number(local.slice(8, 10)),
      });

      await wrapper.findComponent(".input-timezone").vm.$emit("update:modelValue", target);
      expect(wrapper.vm.time).toBe(savedLocal.slice(11, 19));
      await wrapper.vm.save(false);

      expect(lastBody()).toEqual({ ...body, TakenSrc: "manual" });
      expect(photo.TakenAt).toBe(savedTaken);
      expect(photo.TakenAtLocal).toBe(savedLocal);
      expect([photo.Year, photo.Month, photo.Day]).toEqual(ymd);

      photo.Caption = "Updated caption";
      await wrapper.vm.save(false);
      expect(lastBody()).toEqual({ Caption: "Updated caption", CaptionSrc: "manual" });
      expect(photo.TakenAt).toBe(savedTaken);
      expect(photo.TakenAtLocal).toBe(savedLocal);
    });

    it("does not shift the time when the selection passes through Local or is cleared", () => {
      const photo = loadCapture();
      const key = wrapper.vm.timeZoneKey;
      wrapper.vm.setTimeZone("Europe/Berlin");
      wrapper.vm.setTimeZone("Local");
      wrapper.vm.setTimeZone(null);
      expect([photo.TimeZone, wrapper.vm.timeZoneKey]).toEqual(["Local", key + 1]);
      wrapper.vm.setTimeZone("Europe/Berlin");
      expect([photo.TakenAt, photo.TakenAtLocal, wrapper.vm.time]).toEqual(["2025-06-30T23:30:00Z", "2025-07-01T01:30:00Z", "01:30:00"]);
    });

    it("keeps the zone and resets the field when the zone is invalid", () => {
      const photo = loadCapture();
      const key = wrapper.vm.timeZoneKey;
      wrapper.vm.setTimeZone("Invalid/Zone");
      expect([photo.TimeZone, photo.TakenAtLocal, wrapper.vm.timeZoneKey]).toEqual(["UTC", "2025-06-30T23:30:00Z", key + 1]);
    });

    it("resyncs the time field when the saved values are replaced", async () => {
      const photo = loadCapture();
      wrapper.vm.setTimeZone("Europe/Berlin");
      expect(wrapper.vm.time).toBe("01:30:00");
      photo.setValues({ TimeZone: "UTC", TakenAt: "2025-06-30T23:30:00Z", TakenAtLocal: "2025-06-30T23:30:00Z", Day: 30, Month: 6 });
      await nextTick();
      expect(wrapper.vm.time).toBe("23:30:00");
    });

    it("keeps a pending time edit when the saved values are replaced", async () => {
      const photo = loadCapture();
      wrapper.vm.time = "10:15:00";
      wrapper.vm.dateChanged = true;
      photo.setValues({ TakenAt: "2025-06-30T22:00:00Z", TakenAtLocal: "2025-06-30T22:00:00Z" });
      await nextTick();
      expect(wrapper.vm.time).toBe("10:15:00");
    });

    it("keeps the zone and resets the field when a pending time is invalid", () => {
      const photo = loadCapture();
      const key = wrapper.vm.timeZoneKey;
      wrapper.vm.time = "25:99:00";
      wrapper.vm.setTime();
      wrapper.vm.setTimeZone("America/New_York");
      expect([photo.TimeZone, wrapper.vm.timeZoneKey, wrapper.vm.invalidDate]).toEqual(["UTC", key + 1, true]);
    });

    it("does not shift the time when the selection changes repeatedly before saving", () => {
      const utc = loadCapture();
      for (const zone of ["Europe/Berlin", "America/New_York", "UTC-12", "Europe/Berlin"]) {
        wrapper.vm.setTimeZone(zone);
      }
      expect([utc.TakenAt, utc.TakenAtLocal]).toEqual(["2025-06-30T23:30:00Z", "2025-07-01T01:30:00Z"]);

      const berlin = loadCapture({ TimeZone: "Europe/Berlin", TakenAt: "2025-06-30T12:00:00Z", TakenAtLocal: "2025-06-30T14:00:00Z" });
      for (const zone of ["UTC", "America/New_York"]) {
        wrapper.vm.setTimeZone(zone);
      }
      expect([berlin.TakenAt, berlin.TakenAtLocal]).toEqual(["2025-06-30T18:00:00Z", "2025-06-30T14:00:00Z"]);
    });
  });

  describe("updateModel", () => {
    // loadPhoto binds a real photo model in the given zone to the editor.
    const loadPhoto = (values) => {
      const photo = new Photo({
        UID: "ps6sg6be2lvl0yh7",
        TakenAt: "2023-06-15T10:00:00Z",
        TakenAtLocal: "2023-06-15T12:00:00Z",
        Year: 2023,
        Month: 6,
        Day: 15,
        Details: {},
        ...values,
      });
      wrapper.vm.view.model = photo;
      wrapper.vm.syncData();
      return photo;
    };

    it("derives TakenAt from the wall time in a named zone", () => {
      const photo = loadPhoto({ TimeZone: "Europe/Berlin" });
      wrapper.vm.time = "14:30:00";
      expect(wrapper.vm.updateModel()).toBe(true);
      expect(photo.TakenAtLocal).toBe("2023-06-15T14:30:00Z");
      expect(photo.TakenAt).toBe("2023-06-15T12:30:00Z");
    });

    it("keeps TakenAt when the zone is Local", () => {
      const photo = loadPhoto({ TimeZone: "Local" });
      wrapper.vm.time = "14:30:00";
      expect(wrapper.vm.updateModel()).toBe(true);
      expect(photo.TakenAtLocal).toBe("2023-06-15T14:30:00Z");
      expect(photo.TakenAt).toBe("2023-06-15T10:00:00Z");
    });

    it("fills unset date parts and rejects an invalid date", () => {
      const photo = loadPhoto({ TimeZone: "UTC", Day: 0, Month: 0, Year: 0 });
      wrapper.vm.time = "14:30:00";
      expect(wrapper.vm.updateModel()).toBe(true);
      expect([photo.Year, photo.Month, photo.Day]).toEqual([2023, 6, 15]);
      wrapper.vm.time = "25:99:00";
      expect(wrapper.vm.updateModel()).toBe(false);
      expect(wrapper.vm.invalidDate).toBe(true);
    });
  });
});
