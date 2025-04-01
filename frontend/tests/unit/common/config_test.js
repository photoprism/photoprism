import "../fixtures";
import Config from "common/config";
import StorageShim from "node-storage-shim";

let chai = require("chai/chai");
let assert = chai.assert;

const defaultConfig = new Config(new StorageShim(), window.__CONFIG__);

describe("common/config", () => {
  it("should get all config values", () => {
    const storage = new StorageShim();
    const values = { siteTitle: "Foo", name: "testConfig", year: "2300" };

    const cfg = new Config(storage, values);
    const result = cfg.getValues();
    assert.equal(result.name, "testConfig");
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
    assert.equal(cfg.values.settings.ui.theme, "default");
    assert.equal(cfg.values.settings.ui.language, "de");
    assert.equal(cfg.values.new, undefined);
    assert.equal(cfg.values.city, "Hamburg");
    cfg.setValues();
    assert.equal(cfg.values.new, undefined);
    assert.equal(cfg.values.city, "Hamburg");
    cfg.setValues(newValues);
    const result = cfg.getValues();
    assert.equal(result.city, "Berlin");
    assert.equal(result.new, "xxx");
    assert.equal(result.country, "Germany");
    assert.equal(cfg.values.settings.ui.theme, "lavender");
    assert.equal(cfg.values.settings.ui.language, "en");
  });

  it("should test constructor with empty values", () => {
    const storage = new StorageShim();
    const values = {};
    const config = new Config(storage, values);
    assert.equal(config.debug, true);
    assert.equal(config.demo, false);
    assert.equal(config.apiUri, "/api/v1");
  });

  it("should store values", () => {
    const storage = new StorageShim();
    const values = { siteTitle: "Foo", country: "Germany", city: "Hamburg" };
    const config = new Config(storage, values);
    assert.equal(config.storage["config"], undefined);
    config.storeValues();
    const expected = '{"siteTitle":"Foo","country":"Germany","city":"Hamburg"}';
    assert.equal(config.storage["config"], expected);
  });

  it("should return the develop feature flag value", () => {
    assert.equal(defaultConfig.featDevelop(), true);
  });

  it("should return the experimental feature flag value", () => {
    assert.equal(defaultConfig.featExperimental(), true);
  });

  it("should return the preview feature flag value", () => {
    assert.equal(defaultConfig.featPreview(), true);
  });

  it("should set and get single config value", () => {
    const storage = new StorageShim();
    const values = { siteTitle: "Foo", country: "Germany", city: "Hamburg" };

    const config = new Config(storage, values);
    config.set("city", "Berlin");
    const result = config.get("city");
    assert.equal(result, "Berlin");
  });

  it("should return app about", () => {
    assert.equal(defaultConfig.getAbout(), "PhotoPrism® CE");
  });

  it("should return app edition", () => {
    assert.equal(defaultConfig.getEdition(), "ce");
  });

  it("should return settings", () => {
    const result = defaultConfig.getSettings();
    assert.equal(result.ui.theme, "default");
    assert.equal(result.ui.language, "en");
  });

  it("should return feature", () => {
    assert.equal(defaultConfig.feature("places"), true);
    assert.equal(defaultConfig.feature("download"), true);
  });

  it("should test get name", () => {
    const result = defaultConfig.getPerson("a");
    assert.equal(result, null);

    const result2 = defaultConfig.getPerson("Andrea Sander");
    assert.equal(result2.UID, "jr0jgyx2viicdnf7");

    const result3 = defaultConfig.getPerson("Otto Sander");
    assert.equal(result3.UID, "jr0jgyx2viicdn88");
  });

  it("should create, update and delete people", () => {
    const storage = new StorageShim();
    const values = { Debug: true, siteTitle: "Foo", country: "Germany", city: "Hamburg" };

    const cfg = new Config(storage, values);
    cfg.onPeople("people.created", { entities: {} });
    assert.empty(cfg.values.people);
    cfg.onPeople("people.created", {
      entities: [
        {
          UID: "abc123",
          Name: "Test Name",
          Keywords: ["Test", "Name"],
        },
      ],
    });
    assert.equal(cfg.values.people[0].Name, "Test Name");
    cfg.onPeople("people.updated", {
      entities: [
        {
          UID: "abc123",
          Name: "New Name",
          Keywords: ["New", "Name"],
        },
      ],
    });
    assert.equal(cfg.values.people[0].Name, "New Name");
    cfg.onPeople("people.deleted", {
      entities: ["abc123"],
    });
    assert.empty(cfg.values.people);
  });

  it("should return if language is rtl", () => {
    const cfg = new Config(new StorageShim(), Object.assign({}, window.__CONFIG__));
    const result = cfg.isRtl();
    assert.equal(result, false);
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
    assert.equal(result2, true);
    const values2 = { siteTitle: "Foo" };
    const storage = new StorageShim();
    const config3 = new Config(storage, values2);
    const result3 = config3.isRtl();
    assert.equal(result3, false);
    cfg.setLanguage("en");
  });

  it("should return album categories", () => {
    const cfg = new Config(new StorageShim(), Object.assign({}, window.__CONFIG__));
    const result = cfg.albumCategories();
    assert.equal(result[0], "Animal");
    const newValues = {
      albumCategories: ["Mouse"],
    };
    cfg.setValues(newValues);
    const result2 = cfg.albumCategories();
    assert.equal(result2[0], "Mouse");
  });

  it("should update counts", () => {
    const cfg = new Config(new StorageShim(), Object.assign({}, window.__CONFIG__));
    assert.equal(cfg.values.count.all, 133);
    assert.equal(cfg.values.count.photos, 132);
    cfg.onCount("add.photos", {
      count: 2,
    });
    assert.equal(cfg.values.count.all, 135);
    assert.equal(cfg.values.count.photos, 134);
    assert.equal(cfg.values.count.videos, 1);
    cfg.onCount("add.videos", {
      count: 1,
    });
    assert.equal(cfg.values.count.all, 136);
    assert.equal(cfg.values.count.videos, 2);
    assert.equal(cfg.values.count.cameras, 6);
    cfg.onCount("add.cameras", {
      count: 3,
    });
    assert.equal(cfg.values.count.all, 136);
    assert.equal(cfg.values.count.cameras, 9);
    assert.equal(cfg.values.count.lenses, 5);
    cfg.onCount("add.lenses", {
      count: 1,
    });
    assert.equal(cfg.values.count.lenses, 6);
    assert.equal(cfg.values.count.countries, 6);
    cfg.onCount("add.countries", {
      count: 2,
    });
    assert.equal(cfg.values.count.countries, 8);
    assert.equal(cfg.values.count.states, 8);
    cfg.onCount("add.states", {
      count: 1,
    });
    assert.equal(cfg.values.count.states, 9);
    assert.equal(cfg.values.count.people, 5);
    cfg.onCount("add.people", {
      count: 4,
    });
    assert.equal(cfg.values.count.people, 9);
    assert.equal(cfg.values.count.places, 17);
    cfg.onCount("add.places", {
      count: 1,
    });
    assert.equal(cfg.values.count.places, 18);
    assert.equal(cfg.values.count.labels, 22);
    cfg.onCount("add.labels", {
      count: 2,
    });
    assert.equal(cfg.values.count.labels, 24);
    assert.equal(cfg.values.count.albums, 2);
    cfg.onCount("add.albums", {
      count: 3,
    });
    assert.equal(cfg.values.count.albums, 5);
    assert.equal(cfg.values.count.moments, 4);
    cfg.onCount("add.moments", {
      count: 1,
    });
    assert.equal(cfg.values.count.moments, 5);
    assert.equal(cfg.values.count.months, 27);
    cfg.onCount("add.months", {
      count: 4,
    });
    assert.equal(cfg.values.count.months, 31);
    assert.equal(cfg.values.count.folders, 23);
    cfg.onCount("add.folders", {
      count: 2,
    });
    assert.equal(cfg.values.count.folders, 25);
    assert.equal(cfg.values.count.files, 136);
    cfg.onCount("add.files", {
      count: 14,
    });
    assert.equal(cfg.values.count.files, 150);
    assert.equal(cfg.values.count.favorites, 1);
    cfg.onCount("add.favorites", {
      count: 4,
    });
    assert.equal(cfg.values.count.favorites, 5);
    assert.equal(cfg.values.count.review, 22);
    cfg.onCount("add.review", {
      count: 1,
    });
    assert.equal(cfg.values.count.all, 135);
    assert.equal(cfg.values.count.review, 23);
    assert.equal(cfg.values.count.private, 0);
    cfg.onCount("add.private", {
      count: 3,
    });
    assert.equal(cfg.values.count.private, 3);
    assert.equal(cfg.values.count.all, 135);
    cfg.onCount("add.photos", {
      count: 4,
    });
    assert.equal(cfg.values.count.all, 139);
  });

  it("should return user interface direction string", async () => {
    const cfg = new Config(new StorageShim(), Object.assign({}, window.__CONFIG__));
    await cfg.setLanguage("en", true);
    assert.equal(document.dir, "ltr", "document.dir should be ltr");
    assert.equal(cfg.dir(), "ltr");
    assert.equal(cfg.dir(true), "rtl");
    assert.equal(cfg.dir(false), "ltr");
    await cfg.setLanguage("he", false);
    assert.equal(document.dir, "ltr", "document.dir should still be ltr");
    await cfg.setLanguage("he", true);
    assert.equal(cfg.dir(), "rtl");
    assert.equal(document.dir, "rtl", "document.dir should now be rtl");
    assert.equal(cfg.dir(), "rtl");
    assert.equal(cfg.dir(true), "rtl");
    assert.equal(cfg.dir(false), "ltr");
    await cfg.setLanguage("en", true);
    assert.equal(document.dir, "ltr", "document.dir should be ltr again");
    assert.equal(cfg.dir(), "ltr");
  });
});
