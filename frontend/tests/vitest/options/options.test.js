import { describe, it, expect } from "vitest";
import "../fixtures";
import * as options from "../../../src/options/options";
import {
  AccountTypes,
  Colors,
  DefaultLocale,
  Expires,
  FallbackLocale,
  FeedbackCategories,
  FindLanguage,
  FindLocale,
  Gender,
  Intervals,
  ItemsPerPage,
  MapsAnimate,
  MapsStyle,
  Orientations,
  PhotoTypes,
  RetryLimits,
  SetDefaultLocale,
  StartPages,
  ThumbFilters,
  ThumbSizes,
  Timeouts,
} from "../../../src/options/options";

describe("options/options", () => {
  it("should get timezones", () => {
    const timezones = options.TimeZones();
    expect(timezones[0].ID).toBe("Local");
    expect(timezones[0].Name).toBe("Local");
    expect(timezones[1].ID).toBe("UTC");
    expect(timezones[1].Name).toBe("UTC");
  });

  it("should get days", () => {
    const Days = options.Days();
    expect(Days[0].text).toBe("01");
    expect(Days[30].text).toBe("31");
  });

  it("should get years", () => {
    const Years = options.Years();
    const currentYear = new Date().getUTCFullYear();
    expect(Years[0].text).toBe(currentYear.toString());
  });

  it("should get indexed years", () => {
    const IndexedYears = options.IndexedYears();
    expect(IndexedYears[0].text).toBe("2021");
  });

  it("should get months", () => {
    const Months = options.Months();
    expect(Months[5].text).toBe("June");
  });

  it("should get short months", () => {
    const MonthsShort = options.MonthsShort();
    expect(MonthsShort[5].text).toBe("06");
  });

  it("should get languages", () => {
    const Languages = options.Languages();
    expect(Languages[0].value).toBe("en");
  });

  it("should set default locale", () => {
    // Assuming DefaultLocale is exported and mutable for testing purposes
    // Initial state check might depend on test execution order, so we control it here.
    SetDefaultLocale("en"); // Ensure starting state
    expect(options.DefaultLocale).toBe("en");
    SetDefaultLocale("de");
    expect(options.DefaultLocale).toBe("de");
    SetDefaultLocale("en"); // Reset for other tests
  });

  it("should return default when no locale is provided", () => {
    expect(FindLanguage("").value).toBe("en");
  });

  it("should return default if locale is smaller than 2", () => {
    expect(FindLanguage("d").value).toBe("en");
  });

  it("should return default for unknown locale", () => {
    expect(FindLanguage("xx").value).toBe("en");
  });

  it("should return correct locale", () => {
    expect(FindLanguage("de").value).toBe("de");
    expect(FindLanguage("de").text).toBe("Deutsch");
    expect(FindLanguage("de_AT").value).toBe("de");
    expect(FindLanguage("de_AT").text).toBe("Deutsch");
    expect(FindLanguage("zh-tw").value).toBe("zh_TW");
    expect(FindLanguage("zh-tw").text).toBe("繁體中文");
    expect(FindLanguage("zh+tw").value).toBe("zh_TW");
    expect(FindLanguage("zh+tw").text).toBe("繁體中文");
    expect(FindLanguage("zh_AT").value).toBe("zh");
    expect(FindLanguage("zh_AT").text).toBe("简体中文");
    expect(FindLanguage("ZH_TW").value).toBe("zh_TW");
    expect(FindLanguage("ZH_TW").text).toBe("繁體中文");
    expect(FindLanguage("zH-tW").value).toBe("zh_TW");
    expect(FindLanguage("zH-tW").text).toBe("繁體中文");
  });

  it("should return default locale", () => {
    expect(FindLocale("xx")).toBe("en");
    expect(FindLocale("")).toBe("en");
  });

  it("should return fallback locale", () => {
    expect(FallbackLocale()).toBe("en");
  });

  it("should return items per page", () => {
    expect(ItemsPerPage()[0].value).toBe(10);
  });

  it("should return start page options", () => {
    let features = {
      account: true,
      albums: true,
      archive: true,
      delete: true,
      download: true,
      edit: true,
      estimates: true,
      favorites: true,
      files: true,
      folders: true,
      import: true,
      labels: true,
      library: true,
      logs: true,
      calendar: true,
      moments: true,
      people: true,
      places: true,
      private: true,
      ratings: true,
      reactions: true,
      review: true,
      search: true,
      services: true,
      settings: true,
      share: true,
      upload: true,
      videos: true,
    };
    let pages = StartPages(features);
    expect(pages.length).toBe(13);
    expect(pages[5].value).toBe("people");
    expect(pages[5].props.disabled).toBe(false);
    expect(pages[pages.length - 1].value).toBe("settings");
    expect(pages[pages.length - 1].props.disabled).toBe(false);
    features = {
      ...features, // copy previous settings
      calendar: false,
      people: false,
      settings: false,
    };
    pages = StartPages(features);
    expect(pages.length).toBe(13);
    expect(pages[5].value).toBe("people");
    expect(pages[5].props.disabled).toBe(true);
    expect(pages[pages.length - 1].value).toBe("settings");
    expect(pages[pages.length - 1].props.disabled).toBe(true);
  });

  it("should return animation options", () => {
    expect(MapsAnimate()[1].value).toBe(2500);
  });

  it("should return photo types", () => {
    expect(PhotoTypes()[0].value).toBe("image");
    expect(PhotoTypes()[1].value).toBe("raw");
  });

  it("should return map styles", () => {
    let styles = MapsStyle(true);
    expect(styles[styles.length - 1].value).toContain("low-resolution");
    styles = MapsStyle(false);
    expect(styles[styles.length - 1].value).not.toContain("low-resolution");
  });

  it("should return timeouts", () => {
    expect(Timeouts()[1].value).toBe("high");
  });

  it("should return retry limits", () => {
    expect(RetryLimits()[1].value).toBe(1);
  });

  it("should return intervals", () => {
    expect(Intervals()[0].text).toBe("Never");
    expect(Intervals()[1].text).toBe("1 hour");
  });

  it("should return expiry options", () => {
    expect(Expires()[0].text).toBe("Never");
    expect(Expires()[1].text).toBe("After 1 day");
  });

  it("should return colors", () => {
    expect(Colors()[0].Slug).toBe("purple");
  });

  it("should return feedback categories", () => {
    expect(FeedbackCategories()[0].value).toBe("feedback");
  });

  it("should return thumb sizes", () => {
    expect(ThumbSizes()[1].value).toBe("fit_720");
  });

  it("should return thumb filters", () => {
    expect(ThumbFilters()[0].value).toBe("blackman");
  });

  it("should return gender", () => {
    expect(Gender()[2].value).toBe("other");
  });

  it("should return orientations", () => {
    expect(Orientations()[1].text).toBe("90°");
  });

  it("should return service account type options", () => {
    expect(AccountTypes()[0].value).toBe("webdav");
    expect(AccountTypes().length).toBe(1);
  });
});
