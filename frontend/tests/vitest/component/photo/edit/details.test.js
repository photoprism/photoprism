import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { shallowMount, mount } from "@vue/test-utils";
import { nextTick } from "vue";
import PTabPhotoDetails from "component/photo/edit/details.vue";

// Mock model factory
const createMockModel = (overrides = {}) => ({
  hasId: () => true,
  wasChanged: () => false,
  getDateTime: () => ({ toFormat: () => "12:34:56" }),
  localDate: vi.fn().mockReturnValue({
    isValid: true,
    toFormat: (fmt) => {
      if (fmt === "d") return "1";
      if (fmt === "L") return "4";
      if (fmt === "y") return "2024";
      return "2024-04-01T11:29:54";
    },
    toISO: () => "2024-04-01T11:29:54",
  }),
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
            if (f === "edit") return true;
            if (f === "review") return true;
            if (f === "places") return true;
            return false;
          }),
          get: vi.fn().mockImplementation((k) => {
            if (k === "clip") return 255;
            if (k === "readonly") return false;
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
        PLocationInput: { template: '<div class="p-location-input"></div>' },
        PLocationDialog: { template: '<div class="p-location-dialog"></div>' },
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
    if (wrapper) wrapper.unmount();
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
          if (fmt === "d") return "15";
          if (fmt === "L") return "6";
          if (fmt === "y") return "2023";
          return "2023-06-15";
        }),
        toISO: () => "2023-06-15T14:30:00",
      };

      wrapper.vm.view.model.localDate.mockReturnValue(mockLocalDate);
      wrapper.vm.view.model.Day = 0;
      wrapper.vm.view.model.Month = 0;
      wrapper.vm.view.model.Year = 0;

      wrapper.vm.updateModel();

      expect(wrapper.vm.view.model.Day).toBe(15);
      expect(wrapper.vm.view.model.Month).toBe(6);
      expect(wrapper.vm.view.model.Year).toBe(2023);
      expect(wrapper.vm.view.model.TakenAtLocal).toBe("2023-06-15T14:30:00Z");
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

    it("sets Country VAutocomplete readonly prop when Lat/Lng are present", async () => {
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
            PLocationInput: { template: '<div class="p-location-input"></div>' },
            PLocationDialog: { template: '<div class="p-location-dialog"></div>' },
          },
        },
      });

      await nextTick();
      const allAutocompletes = w.findAllComponents({ name: "VAutocomplete" });
      const countryAuto = allAutocompletes.find((c) => c.classes().includes("input-country"));
      expect(countryAuto).toBeTruthy();
      expect(countryAuto.props("readonly")).toBe(false);

      // Set Lat/Lng to non-zero → readonly=true
      w.vm.view.model.Lat = 37.5;
      w.vm.view.model.Lng = -122.4;
      await nextTick();
      expect(countryAuto.props("readonly")).toBe(true);

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

    it("validates text length with textRule", () => {
      const longText = "a".repeat(300); // Longer than clip limit of 255
      const result = wrapper.vm.textRule(longText);
      expect(result).toBe("Text too long");

      const shortText = "Valid text";
      const validResult = wrapper.vm.textRule(shortText);
      expect(validResult).toBe(true);
    });
  });

  describe("photo viewing", () => {
    it("opens photo in lightbox", () => {
      wrapper.vm.openPhoto();
      expect(wrapper.vm.$lightbox.openModels).toHaveBeenCalled();
    });
  });

  describe("timezone handling", () => {
    it("syncs time when timezone changes", () => {
      const syncTimeSpy = vi.spyOn(wrapper.vm, "syncTime");
      wrapper.vm.view.model.TimeZone = "America/New_York";
      wrapper.vm.syncTime();
      expect(syncTimeSpy).toHaveBeenCalled();
    });

    it("updates TakenAt when in UTC timezone", () => {
      wrapper.vm.view.model.currentTimeZoneUTC = vi.fn().mockReturnValue(true);
      const mockLocalDate = {
        isValid: true,
        toFormat: () => "2023-06-15",
        toISO: () => "2023-06-15T14:30:00",
      };

      wrapper.vm.view.model.localDate.mockReturnValue(mockLocalDate);
      wrapper.vm.updateModel();

      expect(wrapper.vm.view.model.TakenAt).toBe("2023-06-15T14:30:00Z");
    });
  });
});
