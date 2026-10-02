import { describe, it, expect, beforeAll, afterAll } from "vitest";
import "../fixtures";

import countries from "options/countries.json";
import { Countries } from "options/options";
import { createGettext } from "common/gettext";

// useLocale installs a vue3-gettext instance for locale with the given catalog.
const useLocale = (locale, translations) => createGettext({ translations, getLanguageLocale: () => locale });

describe("options/options Countries", () => {
  beforeAll(() => {
    useLocale("de", { de: { Unknown: "Unbekannt" } });
  });
  afterAll(() => {
    useLocale("en", {});
  });
  it("localizes the name of the unknown country", () => {
    expect(Countries().find((c) => c.Code === "zz")).toEqual({ Code: "zz", Name: "Unbekannt" });
  });
  it("keeps the other countries and leaves the source list unchanged", () => {
    const list = Countries();
    expect(list).toHaveLength(countries.length);
    expect(list.find((c) => c.Code === "de")).toEqual({ Code: "de", Name: "Germany" });
    expect(countries.find((c) => c.Code === "zz").Name).toBe("Unknown");
  });
});
