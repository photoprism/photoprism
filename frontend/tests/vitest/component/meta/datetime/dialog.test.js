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
      timeIsUTC() {
        return false;
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

    w.vm.day = 20;
    w.vm.month = 3;
    w.vm.year = 2024;
    w.vm.time = "09:15:00";
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

  it("should show UTC label when photo time is UTC", () => {
    const photo = mockPhoto({
      timeIsUTC() {
        return true;
      },
    });
    const w = mount(PMetaDatetimeDialog, {
      props: { visible: false, photo },
    });

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
