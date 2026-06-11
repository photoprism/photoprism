import { describe, it, expect, beforeEach, afterEach } from "vitest";
import "../fixtures";
import Config from "common/config";
import StorageShim from "node-storage-shim";
import * as themes from "options/themes";

const defaultConfig = new Config(new StorageShim(), window.__CONFIG__);

const createTestConfig = () => {
  const values = JSON.parse(JSON.stringify(window.__CONFIG__));
  return new Config(new StorageShim(), values);
};

const resetThemesToDefault = () => {
  themes.SetOptions([
    {
      text: "Default",
      value: "default",
      disabled: false,
    },
  ]);

  themes.Set("default", {
    name: "default",
    title: "Default",
    colors: {},
    variables: {},
  });
};

describe("common/config", () => {
  beforeEach(() => {
    resetThemesToDefault();
  });

  afterEach(() => {
    resetThemesToDefault();
  });
  it("should get all config values", () => {
    const storage = new StorageShim();
    const values = { siteTitle: "Foo", name: "testConfig", year: "2300" };

    const cfg = new Config(storage, values);
    const result = cfg.getValues();
    expect(result.name).toBe("testConfig");
  });

  it("should set multiple config values", () => {
    const storage = new StorageShim();
    const values = {
      siteTitle: "Foo",
      country: "Germany",
      city: "Hamburg",
      settings: { ui: { language: "de", theme: "default" } },
    };
    const newValues = {
      siteTitle: "Foo",
      new: "xxx",
      city: "Berlin",
      debug: true,
      settings: { ui: { language: "en", theme: "lavender" } },
    };
    const cfg = new Config(storage, values);
    expect(cfg.values.settings.ui.theme).toBe("default");
    expect(cfg.values.settings.ui.language).toBe("de");
    expect(cfg.values.new).toBeUndefined();
    expect(cfg.values.city).toBe("Hamburg");
    cfg.setValues();
    expect(cfg.values.new).toBeUndefined();
    expect(cfg.values.city).toBe("Hamburg");
    cfg.setValues(newValues);
    const result = cfg.getValues();
    expect(result.city).toBe("Berlin");
    expect(result.new).toBe("xxx");
    expect(result.country).toBe("Germany");
    expect(cfg.values.settings.ui.theme).toBe("lavender");
    expect(cfg.values.settings.ui.language).toBe("en");
  });

  it("should test constructor with empty values", () => {
    const storage = new StorageShim();
    const values = {};
    const config = new Config(storage, values);
    expect(config.debug).toBe(true);
    expect(config.demo).toBe(false);
    expect(config.frontendUri).toBe("/library");
    expect(config.loginUri).toBe("/library/login");
    expect(config.apiUri).toBe("/api/v1");
  });

  it("derives login uri from configured frontend uri", () => {
    const storage = new StorageShim();
    const values = {
      siteTitle: "Foo",
      baseUri: "/portal",
      frontendUri: "/portal/admin/",
    };

    const config = new Config(storage, values);
    expect(config.frontendUri).toBe("/portal/admin");
    expect(config.loginUri).toBe("/portal/admin/login");
  });

  it("uses base uri fallback when frontend uri is missing", () => {
    const storage = new StorageShim();
    const values = {
      siteTitle: "Foo",
      baseUri: "/portal",
    };

    const config = new Config(storage, values);
    expect(config.frontendUri).toBe("/portal/library");
    expect(config.loginUri).toBe("/portal/library/login");
  });

  it("themeAssetUri prefixes a root-relative theme path with the base uri", () => {
    const config = new Config(new StorageShim(), { siteTitle: "Foo", baseUri: "/i/pro-1" });
    expect(config.themeAssetUri("/_theme/logo.svg")).toBe("/i/pro-1/_theme/logo.svg");
  });

  it("themeAssetUri returns the path unchanged without a base uri", () => {
    const config = new Config(new StorageShim(), { siteTitle: "Foo" });
    expect(config.themeAssetUri("/_theme/logo.svg")).toBe("/_theme/logo.svg");
  });

  it("themeAssetUri leaves absolute, protocol-relative, already-prefixed, and empty paths alone", () => {
    const config = new Config(new StorageShim(), { siteTitle: "Foo", baseUri: "/i/pro-1" });
    expect(config.themeAssetUri("https://cdn.example.com/logo.svg")).toBe("https://cdn.example.com/logo.svg");
    expect(config.themeAssetUri("//cdn.example.com/logo.svg")).toBe("//cdn.example.com/logo.svg");
    expect(config.themeAssetUri("/i/pro-1/_theme/logo.svg")).toBe("/i/pro-1/_theme/logo.svg");
    expect(config.themeAssetUri("")).toBe("");
  });

  it("getIcon prefixes a theme-provided icon with the base uri on path-prefixed deployments", () => {
    const config = new Config(new StorageShim(), {
      siteTitle: "Foo",
      baseUri: "/i/pro-1",
      settings: { ui: { theme: "branded" } },
    });
    themes.Set("branded", { name: "branded", title: "Branded", colors: {}, variables: { icon: "/_theme/logo.svg" } });
    config.setTheme("branded");
    expect(config.getIcon()).toBe("/i/pro-1/_theme/logo.svg");
  });

  it("should store values", () => {
    const storage = new StorageShim();
    const values = { siteTitle: "Foo", country: "Germany", city: "Hamburg" };
    const config = new Config(storage, values);
    expect(config.storage["config"]).toBe(null);
    config.storeValues();
    const expected = '{"siteTitle":"Foo","country":"Germany","city":"Hamburg"}';
    expect(config.storage["config"]).toBe(expected);
  });

  it("should return the develop feature flag value", () => {
    expect(defaultConfig.featDevelop()).toBe(true);
  });

  it("should return the experimental feature flag value", () => {
    expect(defaultConfig.featExperimental()).toBe(true);
  });

  it("should return the preview feature flag value", () => {
    expect(defaultConfig.featPreview()).toBe(true);
  });

  it("should set and get single config value", () => {
    const storage = new StorageShim();
    const values = { siteTitle: "Foo", country: "Germany", city: "Hamburg" };

    const config = new Config(storage, values);
    config.set("city", "Berlin");
    const result = config.get("city");
    expect(result).toBe("Berlin");
  });

  it("should return app about", () => {
    expect(defaultConfig.getAbout()).toBe("PhotoPrism® CE");
  });

  it("honors forced themes when setting theme", () => {
    const storage = new StorageShim();
    const cfg = new Config(storage, {
      settings: {
        ui: {
          theme: "default",
        },
      },
    });

    const forcedTheme = {
      name: "portal-forced",
      title: "Portal Forced",
      force: true,
      colors: { background: "#111111" },
      variables: {},
    };

    themes.Assign([forcedTheme]);

    cfg.setTheme("default");

    expect(cfg.themeName).toBe("portal-forced");
    expect(cfg.theme.colors.background).toBe("#111111");
  });

  it("should return app edition", () => {
    expect(defaultConfig.getEdition()).toBe("ce");
  });

  it("should return settings", () => {
    const result = defaultConfig.getSettings();
    expect(result.ui.theme).toBe("default");
    expect(result.ui.language).toBe("en");
  });

  it("should return feature", () => {
    expect(defaultConfig.feature("places")).toBe(true);
    expect(defaultConfig.feature("download")).toBe(true);
  });

  it("featAppPasswords mirrors the appPasswords feature flag", () => {
    const cfg = createTestConfig();
    const settings = JSON.parse(JSON.stringify(cfg.getSettings()));
    settings.features = { ...settings.features, appPasswords: true };
    cfg.set("settings", settings);
    expect(cfg.featAppPasswords()).toBe(true);
    settings.features = { ...settings.features, appPasswords: false };
    cfg.set("settings", settings);
    expect(cfg.featAppPasswords()).toBe(false);
  });

  it("returns albums when library access is restricted", () => {
    const cfg = createTestConfig();
    const settings = JSON.parse(JSON.stringify(cfg.getSettings()));
    settings.features = {
      ...settings.features,
      search: true,
      albums: true,
      settings: true,
    };
    cfg.set("settings", settings);
    cfg.set("acl", {
      photos: { full_access: false, access_library: false },
      albums: { full_access: false, view: true },
      settings: { full_access: false, update: false },
    });

    expect(cfg.getDefaultRoute()).toBe("albums");
  });

  it("returns settings when library and albums are unavailable", () => {
    const cfg = createTestConfig();
    const settings = JSON.parse(JSON.stringify(cfg.getSettings()));
    settings.features = {
      ...settings.features,
      search: true,
      albums: false,
      settings: true,
    };
    cfg.set("settings", settings);
    cfg.set("acl", {
      photos: { full_access: false, access_library: false },
      albums: { full_access: false, view: false },
      settings: { full_access: false, update: false },
    });

    expect(cfg.getDefaultRoute()).toBe("settings");
  });

  it("honors settings start page when permitted", () => {
    const cfg = createTestConfig();
    const settings = JSON.parse(JSON.stringify(cfg.getSettings()));
    settings.ui = {
      ...settings.ui,
      startPage: "settings",
    };
    settings.features = {
      ...settings.features,
      search: true,
      settings: true,
    };
    cfg.set("settings", settings);
    cfg.set("acl", {
      photos: { full_access: false, access_library: true },
      settings: { full_access: false, update: true },
    });

    expect(cfg.getDefaultRoute()).toBe("settings");
  });

  it("falls back to default route when settings feature is disabled", () => {
    const cfg = createTestConfig();
    const settings = JSON.parse(JSON.stringify(cfg.getSettings()));
    settings.ui = {
      ...settings.ui,
      startPage: "settings",
    };
    settings.features = {
      ...settings.features,
      search: true,
      settings: false,
    };
    cfg.set("settings", settings);
    cfg.set("acl", {
      photos: { full_access: false, access_library: true },
      settings: { full_access: false, update: true },
    });

    expect(cfg.getDefaultRoute()).toBe("browse");
  });

  it("should test get name", () => {
    const result = defaultConfig.getPerson("a");
    expect(result).toBeNull();

    const result2 = defaultConfig.getPerson("Andrea Sander");
    expect(result2.UID).toBe("jr0jgyx2viicdnf7");

    const result3 = defaultConfig.getPerson("Otto Sander");
    expect(result3.UID).toBe("jr0jgyx2viicdn88");
  });

  it("should create, update and delete people", () => {
    const storage = new StorageShim();
    const values = { Debug: true, siteTitle: "Foo", country: "Germany", city: "Hamburg" };

    const cfg = new Config(storage, values);
    cfg.onPeople("people.created", { entities: {} });
    expect(cfg.values.people).toEqual([]);
    cfg.onPeople("people.created", {
      entities: [
        {
          UID: "abc123",
          Name: "Test Name",
          Keywords: ["Test", "Name"],
        },
      ],
    });
    expect(cfg.values.people[0].Name).toBe("Test Name");
    cfg.onPeople("people.updated", {
      entities: [
        {
          UID: "abc123",
          Name: "New Name",
          Keywords: ["New", "Name"],
        },
      ],
    });
    expect(cfg.values.people[0].Name).toBe("New Name");
    cfg.onPeople("people.deleted", {
      entities: ["abc123"],
    });
    expect(cfg.values.people).toEqual([]);
  });

  it("should return language locale", () => {
    const cfg = new Config(new StorageShim(), Object.assign({}, window.__CONFIG__));
    expect(cfg.getLanguageLocale()).toBe("en");
  });

  it("should return user time zone", () => {
    const cfg = new Config(new StorageShim(), Object.assign({}, window.__CONFIG__));
    expect(cfg.getTimeZone()).toBe("Local");
  });

  it("should return if language is rtl", () => {
    const cfg = new Config(new StorageShim(), Object.assign({}, window.__CONFIG__));
    const result = cfg.isRtl();
    expect(result).toBe(false);
    const newValues = {
      Debug: true,
      siteTitle: "Foo",
      country: "Germany",
      city: "Hamburg",
      settings: {
        ui: {
          language: "he",
        },
      },
    };
    cfg.setValues(newValues);
    const result2 = cfg.isRtl();
    expect(result2).toBe(true);
    const values2 = { siteTitle: "Foo" };
    const storage = new StorageShim();
    const config3 = new Config(storage, values2);
    const result3 = config3.isRtl();
    expect(result3).toBe(false);
    cfg.setLanguage("en");
  });

  it("should return album categories", () => {
    const cfg = new Config(new StorageShim(), Object.assign({}, window.__CONFIG__));
    const result = cfg.albumCategories();
    expect(result[0]).toBe("Animal");
    const newValues = {
      albumCategories: ["Mouse"],
    };
    cfg.setValues(newValues);
    const result2 = cfg.albumCategories();
    expect(result2[0]).toBe("Mouse");
  });

  it("should update counts", () => {
    const cfg = new Config(new StorageShim(), Object.assign({}, window.__CONFIG__));
    expect(cfg.values.count.all).toBe(133);
    expect(cfg.values.count.photos).toBe(132);
    cfg.onCount("add.photos", {
      count: 2,
    });
    expect(cfg.values.count.all).toBe(135);
    expect(cfg.values.count.photos).toBe(134);
    expect(cfg.values.count.videos).toBe(1);
    cfg.onCount("add.videos", {
      count: 1,
    });
    expect(cfg.values.count.all).toBe(136);
    expect(cfg.values.count.videos).toBe(2);
    expect(cfg.values.count.cameras).toBe(6);
    cfg.onCount("add.cameras", {
      count: 3,
    });
    expect(cfg.values.count.all).toBe(136);
    expect(cfg.values.count.cameras).toBe(9);
    expect(cfg.values.count.lenses).toBe(5);
    cfg.onCount("add.lenses", {
      count: 1,
    });
    expect(cfg.values.count.lenses).toBe(6);
    expect(cfg.values.count.countries).toBe(6);
    cfg.onCount("add.countries", {
      count: 2,
    });
    expect(cfg.values.count.countries).toBe(8);
    expect(cfg.values.count.states).toBe(8);
    cfg.onCount("add.states", {
      count: 1,
    });
    expect(cfg.values.count.states).toBe(9);
    expect(cfg.values.count.people).toBe(5);
    cfg.onCount("add.people", {
      count: 4,
    });
    expect(cfg.values.count.people).toBe(9);
    expect(cfg.values.count.places).toBe(17);
    cfg.onCount("add.places", {
      count: 1,
    });
    expect(cfg.values.count.places).toBe(18);
    expect(cfg.values.count.labels).toBe(22);
    cfg.onCount("add.labels", {
      count: 2,
    });
    expect(cfg.values.count.labels).toBe(24);
    expect(cfg.values.count.albums).toBe(2);
    cfg.onCount("add.albums", {
      count: 3,
    });
    expect(cfg.values.count.albums).toBe(5);
    expect(cfg.values.count.moments).toBe(4);
    cfg.onCount("add.moments", {
      count: 1,
    });
    expect(cfg.values.count.moments).toBe(5);
    expect(cfg.values.count.months).toBe(27);
    cfg.onCount("add.months", {
      count: 4,
    });
    expect(cfg.values.count.months).toBe(31);
    expect(cfg.values.count.folders).toBe(23);
    cfg.onCount("add.folders", {
      count: 2,
    });
    expect(cfg.values.count.folders).toBe(25);
    expect(cfg.values.count.files).toBe(136);
    cfg.onCount("add.files", {
      count: 14,
    });
    expect(cfg.values.count.files).toBe(150);
    expect(cfg.values.count.favorites).toBe(1);
    cfg.onCount("add.favorites", {
      count: 4,
    });
    expect(cfg.values.count.favorites).toBe(5);
    expect(cfg.values.count.review).toBe(22);
    cfg.onCount("add.review", {
      count: 1,
    });
    expect(cfg.values.count.all).toBe(135);
    expect(cfg.values.count.review).toBe(23);
    expect(cfg.values.count.private).toBe(0);
    cfg.onCount("add.private", {
      count: 3,
    });
    expect(cfg.values.count.private).toBe(3);
    expect(cfg.values.count.all).toBe(135);
    cfg.onCount("add.photos", {
      count: 4,
    });
    expect(cfg.values.count.all).toBe(139);
  });

  it("should return user interface direction string", async () => {
    const cfg = new Config(new StorageShim(), Object.assign({}, window.__CONFIG__));
    await cfg.setLanguage("en", true);
    expect(document.dir).toBe("ltr");
    expect(cfg.dir()).toBe("ltr");
    expect(cfg.dir(true)).toBe("rtl");
    expect(cfg.dir(false)).toBe("ltr");
    await cfg.setLanguage("he", false);
    expect(document.dir).toBe("ltr");
    await cfg.setLanguage("he", true);
    expect(cfg.dir()).toBe("rtl");
    expect(document.dir).toBe("rtl");
    expect(cfg.dir()).toBe("rtl");
    expect(cfg.dir(true)).toBe("rtl");
    expect(cfg.dir(false)).toBe("ltr");
    await cfg.setLanguage("en", true);
    expect(document.dir).toBe("ltr");
    expect(cfg.dir()).toBe("ltr");
  });

  // A10 contract: isPublic / isDemo / isPortal must return a Boolean for every
  // input shape so that bindings like `:disabled="isDemo"` never pass undefined
  // to a Vuetify Boolean prop. See #4966.
  describe("isPublic / isDemo / isPortal Boolean contract", () => {
    const make = (overrides) => new Config(new StorageShim(), { ...window.__CONFIG__, ...overrides });

    it("return Boolean false when the underlying flag is missing", () => {
      const cfg = make({ public: undefined, demo: undefined, portal: undefined });
      for (const fn of ["isPublic", "isDemo", "isPortal"]) {
        const result = cfg[fn]();
        expect(typeof result, fn).toBe("boolean");
        expect(result, fn).toBe(false);
      }
    });
    it("return Boolean true when the underlying flag is true", () => {
      const cfg = make({ public: true, demo: true, portal: true });
      for (const fn of ["isPublic", "isDemo", "isPortal"]) {
        const result = cfg[fn]();
        expect(typeof result, fn).toBe("boolean");
        expect(result, fn).toBe(true);
      }
    });
    it("return Boolean false when `values` itself is missing", () => {
      const cfg = new Config(new StorageShim(), null);
      for (const fn of ["isPublic", "isDemo", "isPortal"]) {
        const result = cfg[fn]();
        expect(typeof result, fn).toBe("boolean");
        expect(result, fn).toBe(false);
      }
    });
  });

  describe("cluster OIDC accessors", () => {
    const make = (oidc) => new Config(new StorageShim(), { ...window.__CONFIG__, ext: { oidc } });

    it("isClusterOidc reflects the ext.oidc.cluster flag as a Boolean", () => {
      expect(make({ cluster: true }).isClusterOidc()).toBe(true);
      expect(make({ cluster: false }).isClusterOidc()).toBe(false);
      expect(make(undefined).isClusterOidc()).toBe(false);
      expect(new Config(new StorageShim(), null).isClusterOidc()).toBe(false);
    });

    it("oidcLoginUri returns the ext.oidc.loginUri or an empty string", () => {
      expect(make({ loginUri: "/library/api/v1/oidc/login" }).oidcLoginUri()).toBe("/library/api/v1/oidc/login");
      expect(make({}).oidcLoginUri()).toBe("");
      expect(new Config(new StorageShim(), null).oidcLoginUri()).toBe("");
    });
  });

  describe("storage availability", () => {
    const make = (usage) => new Config(new StorageShim(), { ...window.__CONFIG__, usage });

    it("filesQuotaReached returns true only when the files quota is reached or exceeded", () => {
      expect(make({ filesUsedPct: 99 }).filesQuotaReached()).toBe(false);
      expect(make({ filesUsedPct: 100 }).filesQuotaReached()).toBe(true);
      expect(make({ filesUsedPct: 101 }).filesQuotaReached()).toBe(true);
    });
    it("storageLow mirrors the usage.storageLow flag as a Boolean", () => {
      expect(make({ storageLow: true }).storageLow()).toBe(true);
      expect(make({ storageLow: false }).storageLow()).toBe(false);
      expect(make({}).storageLow()).toBe(false);
    });
    it("insufficientStorage is true when either the quota is reached or storage is low", () => {
      expect(make({ filesUsedPct: 50, storageLow: false }).insufficientStorage()).toBe(false);
      expect(make({ filesUsedPct: 100, storageLow: false }).insufficientStorage()).toBe(true);
      expect(make({ filesUsedPct: 50, storageLow: true }).insufficientStorage()).toBe(true);
    });
    it("return Boolean false when usage info is missing", () => {
      const cfg = new Config(new StorageShim(), null);
      for (const fn of ["filesQuotaReached", "storageLow", "insufficientStorage"]) {
        const result = cfg[fn]();
        expect(typeof result, fn).toBe("boolean");
        expect(result, fn).toBe(false);
      }
    });
  });
});
