import Config from "common/config";
import MockAdapter from "axios-mock-adapter";
import Api from "common/api";

window.__CONFIG__ = {
    "name": "PhotoPrism",
    "version": "200531-4684f66-Linux-x86_64-DEBUG",
    "copyright": "(c) 2018-2020 PhotoPrism.org \u003chello@photoprism.org\u003e",
    "flags": "public debug experimental settings",
    "siteUrl": "http://localhost:2342/",
    "siteTitle": "PhotoPrism",
    "siteCaption": "Browse your life",
    "siteDescription": "Personal Photo Management powered by Go and Google TensorFlow. Free and open-source.",
    "siteAuthor": "Anonymous",
    "debug": true,
    "readonly": false,
    "uploadNSFW": false,
    "public": true,
    "experimental": true,
    "disableSettings": false,
    "albumCategories": null,
    "albums": [],
    "cameras": [{
        "ID": 2,
        "Slug": "olympus-c2500l",
        "Name": "Olympus C2500L",
        "Make": "Olympus",
        "Model": "C2500L"
    }, {"ID": 1, "Slug": "zz", "Name": "Unknown", "Make": "", "Model": "Unknown"}],
    "lenses": [{"ID": 1, "Slug": "zz", "Name": "Unknown", "Make": "", "Model": "Unknown", "Type": ""}],
    "countries": [{"ID": "de", "Slug": "germany", "Name": "Germany"}, {
        "ID": "is",
        "Slug": "iceland",
        "Name": "Iceland"
    }, {"ID": "zz", "Slug": "zz", "Name": "Unknown"}],
    "thumbs": [{"Name": "fit_720", "Width": 720, "Height": 720}, {
        "Name": "fit_2048",
        "Width": 2048,
        "Height": 2048
    }, {"Name": "fit_1280", "Width": 1280, "Height": 1024}, {
        "Name": "fit_1920",
        "Width": 1920,
        "Height": 1200
    }, {"Name": "fit_2560", "Width": 2560, "Height": 1600}, {"Name": "fit_3840", "Width": 3840, "Height": 2400}],
    "downloadToken": "1uhovi0e",
    "previewToken": "static",
    "jsHash": "0fd34136",
    "cssHash": "2b327230",
    "settings": {
        "theme": "default",
        "language": "en",
        "templates": {"default": "index.tmpl"},
        "maps": {"animate": 0, "style": "streets"},
        "features": {
            "archive": true,
            "private": true,
            "review": true,
            "upload": true,
            "import": true,
            "files": true,
            "moments": true,
            "labels": true,
            "places": true,
            "download": false,
            "edit": true,
            "share": true,
            "logs": true
        },
        "import": {"path": "/", "move": false},
        "index": {"path": "/", "convert": true, "rescan": false, "group": true}
    },
    "count": {
        "cameras": 1,
        "lenses": 0,
        "countries": 2,
        "photos": 126,
        "videos": 0,
        "hidden": 3,
        "favorites": 1,
        "private": 0,
        "review": 0,
        "states": 2,
        "albums": 0,
        "moments": 0,
        "months": 0,
        "folders": 0,
        "files": 255,
        "places": 0,
        "labels": 13,
        "labelMaxPhotos": 1
    },
    "pos": {"uid": "", "loc": "", "utc": "0001-01-01T00:00:00Z", "lat": 0, "lng": 0},
    "years": [2003, 2002],
    "colors": [{"Example": "#AB47BC", "Name": "Purple", "Slug": "purple"}, {
        "Example": "#FF00FF",
        "Name": "Magenta",
        "Slug": "magenta"
    }, {"Example": "#EC407A", "Name": "Pink", "Slug": "pink"}, {
        "Example": "#EF5350",
        "Name": "Red",
        "Slug": "red"
    }, {"Example": "#FFA726", "Name": "Orange", "Slug": "orange"}, {
        "Example": "#D4AF37",
        "Name": "Gold",
        "Slug": "gold"
    }, {"Example": "#FDD835", "Name": "Yellow", "Slug": "yellow"}, {
        "Example": "#CDDC39",
        "Name": "Lime",
        "Slug": "lime"
    }, {"Example": "#66BB6A", "Name": "Green", "Slug": "green"}, {
        "Example": "#009688",
        "Name": "Teal",
        "Slug": "teal"
    }, {"Example": "#00BCD4", "Name": "Cyan", "Slug": "cyan"}, {
        "Example": "#2196F3",
        "Name": "Blue",
        "Slug": "blue"
    }, {"Example": "#A1887F", "Name": "Brown", "Slug": "brown"}, {
        "Example": "#F5F5F5",
        "Name": "White",
        "Slug": "white"
    }, {"Example": "#9E9E9E", "Name": "Grey", "Slug": "grey"}, {
        "Example": "#212121",
        "Name": "Black",
        "Slug": "black"
    }],
    "categories": [{"UID": "lqb6y631re96cper", "Slug": "animal", "Name": "Animal"}, {
        "UID": "lqb6y5gvo9avdfx5",
        "Slug": "architecture",
        "Name": "Architecture"
    }, {"UID": "lqb6y633nhfj1uzt", "Slug": "bird", "Name": "Bird"}, {
        "UID": "lqb6y633g3hxg1aq",
        "Slug": "farm",
        "Name": "Farm"
    }, {"UID": "lqb6y4i1ez9cw5bi", "Slug": "nature", "Name": "Nature"}, {
        "UID": "lqb6y4f2v7dw8irs",
        "Slug": "plant",
        "Name": "Plant"
    }, {"UID": "lqb6y6s2ohhmu0fn", "Slug": "reptile", "Name": "Reptile"}, {
        "UID": "lqb6y6ctgsq2g2np",
        "Slug": "water",
        "Name": "Water"
    }],
    "clip": 160,
    "server": {
        "cores": 2,
        "routines": 23,
        "memory": {"used": 1224531272, "reserved": 1416904088, "info": "Used 1.2 GB / Reserved 1.4 GB"}
    }
};

let chai = require("chai/chai");
let assert = chai.assert;

const config2 = new Config(window.localStorage, window.__CONFIG__);

describe("common/config", () => {

    const mock = new MockAdapter(Api);

    it("should get all config values",  () => {
        const storage = window.localStorage;
        const values = {siteTitle: "Foo", name: "testConfig", year: "2300"};

        const config = new Config(storage, values);
        const result = config.getValues();
        assert.equal(result.name, "testConfig");
    });

    it("should set multiple config values",  () => {
        const storage = window.localStorage;
        const values = {siteTitle: "Foo", country: "Germany", city: "Hamburg"};
        const newValues = {siteTitle: "Foo", new: "xxx", city: "Berlin", debug: true, settings: {theme: "lavender"}};
        const config = new Config(storage, values);
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
        assert.equal(config.values.settings.theme, "lavender");
    });

    it("should store values",  () => {
        const storage = window.localStorage;
        const values = {siteTitle: "Foo", country: "Germany", city: "Hamburg"};
        const config = new Config(storage, values);
        assert.equal(config.storage["config"], undefined);
        config.storeValues();
        const expected = '{"siteTitle":"Foo","country":"Germany","city":"Hamburg"}';
        assert.equal(config.storage["config"], expected);
    });

    it("should set and get single config value",  () => {
        const storage = window.localStorage;
        const values = {};

        const config = new Config(storage, values);
        config.set("city", "Berlin");
        const result = config.get("city");
        assert.equal(result, "Berlin");
    });

    it("should return settings",  () => {
        const result = config2.settings();
        assert.equal(result.theme, "default");
        assert.equal(result.language, "en");
    });

    it("should return feature",  () => {
        assert.equal(config2.feature("places"), true);
        assert.equal(config2.feature("download"), false);
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
