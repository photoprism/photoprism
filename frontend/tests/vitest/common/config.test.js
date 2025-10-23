import { describe, it, expect } from "vitest";
import "../fixtures";
import Config from "common/config";
import StorageShim from "node-storage-shim";
import * as themes from "options/themes";

const defaultConfig = new Config(new StorageShim(), window.__CONFIG__);

const createTestConfig = () => {
  const values = JSON.parse(JSON.stringify(window.__CONFIG__));
  return new Config(new StorageShim(), values);
};

describe("common/config", () => {
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
    expect(config.apiUri).toBe("/api/v1");
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

    themes.Assign([forcedTheme]);

    cfg.setTheme("default");

    expect(cfg.themeName).toBe("portal-forced");
    expect(cfg.theme.colors.background).toBe("#111111");

    themes.Remove("portal-forced");
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
});
