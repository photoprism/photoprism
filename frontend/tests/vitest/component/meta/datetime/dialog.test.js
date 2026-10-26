import { describe, it, expect, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { DateTime } from "luxon";
import PMetaDatetimeDialog from "component/meta/datetime/dialog.vue";

describe("PMetaDatetimeDialog component", () => {
  function mockPhoto(overrides = {}) {
    const base = {
      Day: 15,
      Month: 6,
      Year: 2023,
      TimeZone: "Europe/Berlin",
      TakenAtLocal: "2023-06-15T14:30:00",
      TakenAt: "2023-06-15T12:30:00Z",
      getDateTime() {
        return DateTime.fromISO(this.TakenAtLocal).toUTC();
      },
      ...overrides,
    };
    return base;
  }

  it("should load values from photo via loadFromPhoto", () => {
    const photo = mockPhoto();
    const w = mount(PMetaDatetimeDialog, {
      props: { visible: false, photo },
    });

    expect(w.vm.day).toBe(0);
    expect(w.vm.year).toBe(0);

    w.vm.loadFromPhoto();

    expect(w.vm.day).toBe(15);
    expect(w.vm.month).toBe(6);
    expect(w.vm.year).toBe(2023);
    expect(w.vm.timeZone).toBe("Europe/Berlin");
    expect(w.vm.time).toBe("14:30:00");
    expect(w.vm.invalidDate).toBe(false);
  });

  it("should handle null photo gracefully", () => {
    const w = mount(PMetaDatetimeDialog, {
      props: { visible: false, photo: null },
    });

    w.vm.loadFromPhoto();

    expect(w.vm.day).toBe(0);
    expect(w.vm.month).toBe(0);
    expect(w.vm.year).toBe(0);
    expect(w.vm.time).toBe("");
  });

  it("should emit close event", () => {
    const onClose = vi.fn();
    const w = mount(PMetaDatetimeDialog, {
      props: { visible: false, photo: mockPhoto(), onClose },
    });

    w.vm.close();
    expect(onClose).toHaveBeenCalledOnce();
  });

  it("should emit confirm with edited values", () => {
    const onConfirm = vi.fn();
    const w = mount(PMetaDatetimeDialog, {
      props: { visible: false, photo: mockPhoto(), onConfirm },
    });

    w.vm.loadFromPhoto();

    w.vm.setDay(20);
    w.vm.setMonth(3);
    w.vm.setYear(2024);
    w.vm.time = "09:15:00";
    w.vm.setTime();
    w.vm.timeZone = "America/New_York";

    w.vm.confirm();

    expect(onConfirm).toHaveBeenCalledOnce();
    expect(onConfirm).toHaveBeenCalledWith({
      Day: 20,
      Month: 3,
      Year: 2024,
      TimeZone: "America/New_York",
      time: "09:15:00",
    });
  });

  it("should pass the stored timestamps when nothing was edited", () => {
    const onConfirm = vi.fn();
    const w = mount(PMetaDatetimeDialog, {
      props: { visible: false, photo: mockPhoto(), onConfirm },
    });

    w.vm.loadFromPhoto();
    w.vm.confirm();

    expect(onConfirm).toHaveBeenCalledWith({
      Day: 15,
      Month: 6,
      Year: 2023,
      TimeZone: "Europe/Berlin",
      time: "14:30:00",
      TakenAt: "2023-06-15T12:30:00Z",
      TakenAtLocal: "2023-06-15T14:30:00",
    });
  });

  describe("time zone changes", () => {
    // mountUTC mounts the dialog for a UTC photo taken late on June 30.
    const mountUTC = (overrides = {}) => {
      const onConfirm = vi.fn();
      const photo = mockPhoto({
        Day: 30,
        Month: 6,
        Year: 2025,
        TimeZone: "UTC",
        TakenAt: "2025-06-30T23:30:00Z",
        TakenAtLocal: "2025-06-30T23:30:00Z",
        ...overrides,
      });
      const w = mount(PMetaDatetimeDialog, { props: { visible: false, photo, onConfirm } });
      w.vm.loadFromPhoto();
      return { w, onConfirm };
    };

    it("moves the shown date and time and passes the timestamps on", () => {
      const { w, onConfirm } = mountUTC();
      w.vm.setTimeZone("Europe/Berlin");
      expect([w.vm.year, w.vm.month, w.vm.day, w.vm.time]).toEqual([2025, 7, 1, "01:30:00"]);
      expect(w.vm.timeLabel).toContain("Local");
      w.vm.confirm();
      expect(onConfirm).toHaveBeenCalledWith({
        Day: 1,
        Month: 7,
        Year: 2025,
        TimeZone: "Europe/Berlin",
        time: "01:30:00",
        TakenAt: "2025-06-30T23:30:00Z",
        TakenAtLocal: "2025-07-01T01:30:00Z",
      });
    });

    it("applies a pending time edit before moving it", () => {
      const { w, onConfirm } = mountUTC();
      w.vm.time = "10:00:00";
      w.vm.setTime();
      w.vm.setTimeZone("Europe/Berlin");
      expect([w.vm.day, w.vm.time]).toEqual([30, "12:00:00"]);
      w.vm.confirm();
      expect(onConfirm.mock.calls[0][0]).toMatchObject({ TakenAt: "2025-06-30T10:00:00Z", TakenAtLocal: "2025-06-30T12:00:00Z" });
    });

    it("sends manual edits after a zone change as wall time", () => {
      const { w, onConfirm } = mountUTC();
      w.vm.setTimeZone("Europe/Berlin");
      w.vm.time = "03:45:00";
      w.vm.setTime();
      w.vm.confirm();
      expect(onConfirm.mock.calls[0][0]).toEqual({ Day: 1, Month: 7, Year: 2025, TimeZone: "Europe/Berlin", time: "03:45:00" });
    });

    it("keeps a known month and year when the day is unknown", () => {
      const { w } = mountUTC({ Day: -1, Month: 1, Year: 2020, TakenAt: "2020-01-01T03:00:00Z", TakenAtLocal: "2020-01-01T03:00:00Z" });
      w.vm.setTimeZone("America/New_York");
      expect([w.vm.year, w.vm.month, w.vm.day, w.vm.time]).toEqual([2020, 1, -1, "22:00:00"]);
    });

    it("converts from the saved UTC time through repeated selections", () => {
      const { w } = mountUTC();
      w.vm.setTimeZone("Europe/Berlin");
      w.vm.setTimeZone("America/New_York");
      expect([w.vm.year, w.vm.month, w.vm.day, w.vm.time]).toEqual([2025, 6, 30, "19:30:00"]);
    });

    it("keeps the local time of a named zone, also through UTC", () => {
      const onConfirm = vi.fn();
      const w = mount(PMetaDatetimeDialog, { props: { visible: false, photo: mockPhoto(), onConfirm } });
      w.vm.loadFromPhoto();
      w.vm.setTimeZone("UTC");
      expect([w.vm.time, w.vm.timeLabel]).toEqual(["14:30:00", "UTC"]);
      w.vm.setTimeZone("America/New_York");
      expect(w.vm.time).toBe("14:30:00");
      w.vm.confirm();
      expect(onConfirm.mock.calls[0][0]).toMatchObject({ TimeZone: "America/New_York", TakenAt: "2023-06-15T18:30:00Z", TakenAtLocal: "2023-06-15T14:30:00Z" });
    });

    it("does not shift the time when the selection passes through Local or is cleared", () => {
      const { w, onConfirm } = mountUTC();
      const key = w.vm.timeZoneKey;
      w.vm.setTimeZone("Europe/Berlin");
      w.vm.setTimeZone("Local");
      w.vm.setTimeZone(null);
      expect([w.vm.timeZone, w.vm.time, w.vm.timeZoneKey]).toEqual(["Local", "23:30:00", key + 1]);
      w.vm.setTimeZone("America/New_York");
      w.vm.confirm();
      expect(onConfirm.mock.calls[0][0]).toMatchObject({ TakenAt: "2025-06-30T23:30:00Z", TakenAtLocal: "2025-06-30T19:30:00Z", time: "19:30:00" });
    });

    it("ignores an invalid zone and an invalid pending time", () => {
      const { w } = mountUTC();
      const key = w.vm.timeZoneKey;
      w.vm.setTimeZone("Invalid/Zone");
      expect([w.vm.timeZone, w.vm.timeZoneKey]).toEqual(["UTC", key + 1]);
      w.vm.time = "25:99";
      w.vm.setTime();
      w.vm.setTimeZone("Europe/Berlin");
      expect([w.vm.timeZone, w.vm.time, w.vm.timeZoneKey]).toEqual(["UTC", "25:99", key + 2]);
    });
  });

  it("should not emit confirm when date is invalid", () => {
    const onConfirm = vi.fn();
    const w = mount(PMetaDatetimeDialog, {
      props: { visible: false, photo: mockPhoto(), onConfirm },
    });

    w.vm.loadFromPhoto();
    w.vm.invalidDate = true;

    w.vm.confirm();

    expect(onConfirm).not.toHaveBeenCalled();
  });

  // Regression: setTime() previously short-circuited on malformed input
  // (skipping updateLocalDate), leaving invalidDate stale. Without this
  // branch the Confirm button stayed enabled against bad time strings
  // and the parent received "25:99:99" as the new time.
  it("should mark the date invalid when the time field contains a malformed value", () => {
    const onConfirm = vi.fn();
    const w = mount(PMetaDatetimeDialog, {
      props: { visible: false, photo: mockPhoto(), onConfirm },
    });

    w.vm.loadFromPhoto();
    expect(w.vm.invalidDate).toBe(false);

    // Malformed time → setTime flips invalidDate and bails before update.
    w.vm.time = "25:99:99";
    w.vm.setTime();
    expect(w.vm.invalidDate).toBe(true);

    // confirm() then refuses to emit.
    w.vm.confirm();
    expect(onConfirm).not.toHaveBeenCalled();

    // Recovering to a valid time clears invalidDate via updateLocalDate.
    w.vm.time = "09:15:00";
    w.vm.setTime();
    expect(w.vm.invalidDate).toBe(false);
  });

  it("should show UTC label when the selected zone is UTC", () => {
    const photo = mockPhoto({ TimeZone: "UTC" });
    const w = mount(PMetaDatetimeDialog, {
      props: { visible: false, photo },
    });

    w.vm.loadFromPhoto();

    expect(w.vm.timeLabel).toContain("UTC");
  });

  it("should show Local Time label when photo time is not UTC", () => {
    const w = mount(PMetaDatetimeDialog, {
      props: { visible: false, photo: mockPhoto() },
    });

    expect(w.vm.timeLabel).toContain("Local");
  });

  it("should clamp day when month changes to shorter month", () => {
    const photo = mockPhoto({ Day: 31, Month: 1, Year: 2023 });
    const w = mount(PMetaDatetimeDialog, {
      props: { visible: false, photo },
    });

    w.vm.loadFromPhoto();
    expect(w.vm.day).toBe(31);

    // February has 28 days in 2023
    w.vm.month = 2;
    w.vm.clampDayToValidRange();
    expect(w.vm.day).toBe(28);
  });

  it("should handle leap year day clamping", () => {
    const photo = mockPhoto({ Day: 31, Month: 1, Year: 2024 });
    const w = mount(PMetaDatetimeDialog, {
      props: { visible: false, photo },
    });

    w.vm.loadFromPhoto();

    w.vm.month = 2;
    w.vm.clampDayToValidRange();
    expect(w.vm.day).toBe(29);
  });

  it("should build correct local date strings", () => {
    const w = mount(PMetaDatetimeDialog, {
      props: { visible: false, photo: mockPhoto() },
    });

    w.vm.loadFromPhoto();

    expect(w.vm.localYearString()).toBe("2023");
    expect(w.vm.localMonthString()).toBe("06");
    expect(w.vm.localDayString()).toBe("15");
  });

  it("should pad single-digit values in date strings", () => {
    const photo = mockPhoto({ Day: 5, Month: 3, Year: 800 });
    const w = mount(PMetaDatetimeDialog, {
      props: { visible: false, photo },
    });

    w.vm.loadFromPhoto();

    expect(w.vm.localYearString()).toBe("0800");
    expect(w.vm.localMonthString()).toBe("03");
    expect(w.vm.localDayString()).toBe("05");
  });

  it("should use defaults for missing photo fields", () => {
    const photo = mockPhoto({ Day: 0, Month: 0, Year: 2023, TimeZone: "" });
    const w = mount(PMetaDatetimeDialog, {
      props: { visible: false, photo },
    });

    w.vm.loadFromPhoto();

    expect(w.vm.day).toBe(0);
    expect(w.vm.month).toBe(0);
    expect(w.vm.year).toBe(2023);
    expect(w.vm.timeZone).toBe("");
  });
});
