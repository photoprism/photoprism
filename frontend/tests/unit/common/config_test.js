import "../fixtures";
import Config from "common/config";
import StorageShim from "node-storage-shim";

let chai = require("chai/chai");
let assert = chai.assert;

const config2 = new Config(new StorageShim(), window.__CONFIG__);

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

  it("should return settings", () => {
    const result = config2.settings();
    assert.equal(result.ui.theme, "default");
    assert.equal(result.ui.language, "en");
  });

  it("should return feature", () => {
    assert.equal(config2.feature("places"), true);
    assert.equal(config2.feature("download"), true);
  });

  it("should test get name", () => {
    const result = config2.getPerson("a");
    assert.equal(result, null);

    const result2 = config2.getPerson("Andrea Sander");
    assert.equal(result2.UID, "jr0jgyx2viicdnf7");

    const result3 = config2.getPerson("Otto Sander");
    assert.equal(result3.UID, "jr0jgyx2viicdn88");
  });

  it("should create, update and delete people", () => {
    const storage = new StorageShim();
    const values = { Debug: true, siteTitle: "Foo", country: "Germany", city: "Hamburg" };

    const config = new Config(storage, values);
    config.onPeople(".created");
    assert.empty(config.values.people);
    config.onPeople(".created", {
      entities: [
        {
          UID: "abc123",
          Name: "Test Name",
          Keywords: ["Test", "Name"],
        },
      ],
    });
    assert.equal(config.values.people[0].Name, "Test Name");
    config.onPeople(".updated", {
      entities: [
        {
          UID: "abc123",
          Name: "New Name",
          Keywords: ["New", "Name"],
        },
      ],
    });
    assert.equal(config.values.people[0].Name, "New Name");
    config.onPeople(".deleted", {
      entities: [
        {
          UID: "abc123",
          Name: "New Name",
          Keywords: ["New", "Name"],
        },
      ],
    });
    assert.equal(config.values.people[0].Name, "New Name");
  });

  it("should return if language is rtl", () => {
    const result = config2.rtl();
    assert.equal(result, false);
    const storage = new StorageShim();
    const values = {
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
    const config = new Config(storage, values);
    const result2 = config.rtl();
    assert.equal(result2, true);
    const values2 = { siteTitle: "Foo", country: "Germany", city: "Hamburg" };
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

  //TODO
  /*it.only("should test onCount",  () => {
        const items = [{}, {}, {}];
        assert.equal(config2.values.count.cameras, 1);
        config2.onCount("add.camera", items);
        assert.equal(config2.values.count.cameras, 1);
        console.log(config2.values.count);
        config2.onCount("add.cameras", items);
        console.log(config2.values.count);
        assert.equal(config2.values.count.cameras, 4);
        config2.onCount("add.lenses", items);
        assert.equal(config2.values.count.lenses, 3);
        config2.onCount("add.countries", items);
        assert.equal(config2.values.count.countries, 5);
        config2.onCount("add.photos", items);
        assert.equal(config2.values.count.photos, 129);
        config2.onCount("add.videos", items);
        assert.equal(config2.values.count.videos, 3);
        config2.onCount("add.hidden", items);
        assert.equal(config2.values.count.hidden, 6);
        config2.onCount("add.favorites", items);
        assert.equal(config2.values.count.favorites, 4);
        config2.onCount("add.private", items);
        assert.equal(config2.values.count.private, 3);
        config2.onCount("add.review", items);
        assert.equal(config2.values.count.review, 3);
        config2.onCount("add.states", items);
        assert.equal(config2.values.count.states, 5);
        config2.onCount("add.albums", items);
        assert.equal(config2.values.count.albums, 3);
        config2.onCount("add.moments", items);
        assert.equal(config2.values.count.moments, 3);
        config2.onCount("add.months", items);
        assert.equal(config2.values.count.months, 3);
        config2.onCount("add.folders", items);
        assert.equal(config2.values.count.folders, 3);
        config2.onCount("add.files", items);
        assert.equal(config2.values.count.files, 258);
        config2.onCount("add.places", items);
        assert.equal(config2.values.count.places, 3);
        config2.onCount("add.labels", items);
        assert.equal(config2.values.count.labels, 12);
    });*/
});
