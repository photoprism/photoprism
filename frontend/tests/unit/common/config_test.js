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

    const config = new Config(storage, values);
    const result = config.getValues();
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
    const config = new Config(storage, values);
    assert.equal(config.values.settings.ui.theme, "default");
    assert.equal(config.values.settings.ui.language, "de");
    assert.equal(config.values.new, undefined);
    assert.equal(config.values.city, "Hamburg");
    config.setValues();
    assert.equal(config.values.new, undefined);
    assert.equal(config.values.city, "Hamburg");
    config.setValues(newValues);
    const result = config.getValues();
    assert.equal(result.city, "Berlin");
    assert.equal(result.new, "xxx");
    assert.equal(result.country, "Germany");
    assert.equal(config.values.settings.ui.theme, "lavender");
    assert.equal(config.values.settings.ui.language, "en");
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
    const result = defaultConfig.settings();
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

    const config = new Config(storage, values);
    config.onPeople("people.created", { entities: {} });
    assert.empty(config.values.people);
    config.onPeople("people.created", {
      entities: [
        {
          UID: "abc123",
          Name: "Test Name",
          Keywords: ["Test", "Name"],
        },
      ],
    });
    assert.equal(config.values.people[0].Name, "Test Name");
    config.onPeople("people.updated", {
      entities: [
        {
          UID: "abc123",
          Name: "New Name",
          Keywords: ["New", "Name"],
        },
      ],
    });
    assert.equal(config.values.people[0].Name, "New Name");
    config.onPeople("people.deleted", {
      entities: ["abc123"],
    });
    assert.empty(config.values.people);
  });

  it("should return if language is rtl", () => {
    const myConfig = new Config(new StorageShim(), Object.assign({}, window.__CONFIG__));
    const result = myConfig.rtl();
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
    myConfig.setValues(newValues);
    const result2 = myConfig.rtl();
    assert.equal(result2, true);
    const values2 = { siteTitle: "Foo" };
    const storage = new StorageShim();
    const config3 = new Config(storage, values2);
    const result3 = config3.rtl();
    assert.equal(result3, false);
  });

  it("should return album categories", () => {
    const myConfig = new Config(new StorageShim(), Object.assign({}, window.__CONFIG__));
    const result = myConfig.albumCategories();
    assert.equal(result[0], "Animal");
    const newValues = {
      albumCategories: ["Mouse"],
    };
    myConfig.setValues(newValues);
    const result2 = myConfig.albumCategories();
    assert.equal(result2[0], "Mouse");
  });

  it("should update counts", () => {
    const myConfig = new Config(new StorageShim(), Object.assign({}, window.__CONFIG__));
    assert.equal(myConfig.values.count.all, 133);
    assert.equal(myConfig.values.count.photos, 132);
    myConfig.onCount("add.photos", {
      count: 2,
    });
    assert.equal(myConfig.values.count.all, 135);
    assert.equal(myConfig.values.count.photos, 134);
    assert.equal(myConfig.values.count.videos, 1);
    myConfig.onCount("add.videos", {
      count: 1,
    });
    assert.equal(myConfig.values.count.all, 136);
    assert.equal(myConfig.values.count.videos, 2);
    assert.equal(myConfig.values.count.cameras, 6);
    myConfig.onCount("add.cameras", {
      count: 3,
    });
    assert.equal(myConfig.values.count.all, 136);
    assert.equal(myConfig.values.count.cameras, 9);
    assert.equal(myConfig.values.count.lenses, 5);
    myConfig.onCount("add.lenses", {
      count: 1,
    });
    assert.equal(myConfig.values.count.lenses, 6);
    assert.equal(myConfig.values.count.countries, 6);
    myConfig.onCount("add.countries", {
      count: 2,
    });
    assert.equal(myConfig.values.count.countries, 8);
    assert.equal(myConfig.values.count.states, 8);
    myConfig.onCount("add.states", {
      count: 1,
    });
    assert.equal(myConfig.values.count.states, 9);
    assert.equal(myConfig.values.count.people, 5);
    myConfig.onCount("add.people", {
      count: 4,
    });
    assert.equal(myConfig.values.count.people, 9);
    assert.equal(myConfig.values.count.places, 17);
    myConfig.onCount("add.places", {
      count: 1,
    });
    assert.equal(myConfig.values.count.places, 18);
    assert.equal(myConfig.values.count.labels, 22);
    myConfig.onCount("add.labels", {
      count: 2,
    });
    assert.equal(myConfig.values.count.labels, 24);
    assert.equal(myConfig.values.count.albums, 2);
    myConfig.onCount("add.albums", {
      count: 3,
    });
    assert.equal(myConfig.values.count.albums, 5);
    assert.equal(myConfig.values.count.moments, 4);
    myConfig.onCount("add.moments", {
      count: 1,
    });
    assert.equal(myConfig.values.count.moments, 5);
    assert.equal(myConfig.values.count.months, 27);
    myConfig.onCount("add.months", {
      count: 4,
    });
    assert.equal(myConfig.values.count.months, 31);
    assert.equal(myConfig.values.count.folders, 23);
    myConfig.onCount("add.folders", {
      count: 2,
    });
    assert.equal(myConfig.values.count.folders, 25);
    assert.equal(myConfig.values.count.files, 136);
    myConfig.onCount("add.files", {
      count: 14,
    });
    assert.equal(myConfig.values.count.files, 150);
    assert.equal(myConfig.values.count.favorites, 1);
    myConfig.onCount("add.favorites", {
      count: 4,
    });
    assert.equal(myConfig.values.count.favorites, 5);
    assert.equal(myConfig.values.count.review, 22);
    myConfig.onCount("add.review", {
      count: 1,
    });
    assert.equal(myConfig.values.count.all, 135);
    assert.equal(myConfig.values.count.review, 23);
    assert.equal(myConfig.values.count.private, 0);
    myConfig.onCount("add.private", {
      count: 3,
    });
    assert.equal(myConfig.values.count.private, 3);
    assert.equal(myConfig.values.count.all, 135);
    myConfig.onCount("add.photos", {
      count: 4,
    });
    assert.equal(myConfig.values.count.all, 139);
  });
});
