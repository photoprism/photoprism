//TODO Add tokens to config data and test hasToken
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
            "download": true,
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
        "stories": 0,
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

import {config} from "session";
import Session from "common/session";
import MockAdapter from "axios-mock-adapter";
import Api from "common/api";

let chai = require("chai/chai");
let assert = chai.assert;

describe('common/session', () => {

    const mock = new MockAdapter(Api);

    beforeEach(() => {
        window.onbeforeunload = () => 'Oh no!';
    });

    it('should construct session', () => {
        const storage = window.localStorage;
        const session = new Session(storage, config);
        assert.equal(session.session_id, null);
    });

    it('should set, get and delete token', () => {
        const storage = window.localStorage;
        const session = new Session(storage, config);
        assert.equal(session.hasToken("1uhovi0e"), false);
        assert.equal(session.session_id, null);
        session.setId(123421);
        assert.equal(session.session_id, 123421);
        const result = session.getId();
        assert.equal(result, 123421);
        session.deleteId();
        assert.equal(session.session_id, null);
    });

    it('should set, get and delete user', () => {
        const storage = window.localStorage;
        const session = new Session(storage, config);
        assert.isFalse(session.user.hasId());
        const values = {"user": {ID: 5, FirstName: "Max", LastName: "Last", Email: "test@test.com", Admin: true}};
        session.setData();
        assert.equal(session.user.FirstName, "");
        session.setData(values);
        assert.equal(session.user.FirstName, "Max");
        assert.equal(session.user.Admin, true);
        const result = session.getUser();
        assert.equal(result.ID, 5);
        assert.equal(result.Email, "test@test.com");
        session.deleteData();
        assert.isFalse(session.user.hasId());
    });

    it('should get user email', () => {
        const storage = window.localStorage;
        const session = new Session(storage, config);
        const values = {"user": {ID: 5, FirstName: "Max", LastName: "Last", Email: "test@test.com", Admin: true}};
        session.setData(values);
        const result = session.getEmail();
        assert.equal(result, "test@test.com");
        const values2 = {"user": {FirstName: "Max", LastName: "Last", Email: "test@test.com", Admin: true}};
        session.setData(values2);
        const result2 = session.getEmail();
        assert.equal(result2, "");
        session.deleteData();
    });

    it('should get user firstname', () => {
        const storage = window.localStorage;
        const session = new Session(storage, config);
        const values = {"user": {ID: 5, FirstName: "Max", LastName: "Last", Email: "test@test.com", Admin: true}};
        session.setData(values);
        const result = session.getFirstName();
        assert.equal(result, "Max");
        const values2 = {"user": {FirstName: "Max", LastName: "Last", Email: "test@test.com", Admin: true}};
        session.setData(values2);
        const result2 = session.getFirstName();
        assert.equal(result2, "");
        session.deleteData();
    });

    it('should get user full name', () => {
        const storage = window.localStorage;
        const session = new Session(storage, config);
        const values = {"user": {ID: 5, FirstName: "Max", LastName: "Last", Email: "test@test.com", Admin: true}};
        session.setData(values);
        const result = session.getFullName();
        assert.equal(result, "Max Last");
        const values2 = {"user": {FirstName: "Max", LastName: "Last", Email: "test@test.com", Admin: true}};
        session.setData(values2);
        const result2 = session.getFullName();
        assert.equal(result2, "");
        session.deleteData();
    });

    it('should test whether user is set', () => {
        const storage = window.localStorage;
        const session = new Session(storage, config);
        const values = {"user": {ID: 5, FirstName: "Max", LastName: "Last", Email: "test@test.com", Admin: true}};
        session.setData(values);
        const result = session.isUser();
        assert.equal(result, true);
        session.deleteData();
    });

    it('should test whether user is admin', () => {
        const storage = window.localStorage;
        const session = new Session(storage, config);
        const values = {"user": {ID: 5, FirstName: "Max", LastName: "Last", Email: "test@test.com", Admin: true}};
        session.setData(values);
        const result = session.isAdmin();
        assert.equal(result, true);
        session.deleteData();
    });

    it('should test whether user is anonymous', () => {
        const storage = window.localStorage;
        const session = new Session(storage, config);
        const values = {"user": {ID: 5, FirstName: "Max", LastName: "Last", Email: "test@test.com", Admin: true}};
        session.setData(values);
        const result = session.isAnonymous();
        assert.equal(result, false);
        session.deleteData();
    });

    it('should test login and logout', async () => {
        mock
            .onPost("session").reply(200, {id: "8877", data: {user: {ID: 1, Email: "test@test.com"}}})
            .onDelete("session/8877").reply(200);
        const storage = window.localStorage;
        const session = new Session(storage, config);
        assert.equal(session.session_id, null);
        assert.equal(session.storage.data, undefined);
        await session.login("test@test.com", "passwd");
        assert.equal(session.session_id, 8877);
        assert.equal(session.storage.data, '{"user":{"ID":1,"Email":"test@test.com"}}');
        await session.logout();
        assert.equal(session.session_id, null);
        mock.reset();
    });

    //TODO Why does it make other tests fail?
    /*it('should test onLogout', async () => {
        mock
            .onPost("session").reply(200, {id: "8877", data: {user: {ID: 1, Email: "test@test.com"}}})
            .onDelete("session/8877").reply(200);
        const storage = window.localStorage;
        const session = new Session(storage, config);
        //assert.equal(session.session_id, null);
        //assert.equal(session.storage.data, undefined);
        await session.login("test@test.com", "passwd");
        assert.equal(session.session_id, 8877);
        assert.equal(session.storage.data, '{"user":{"ID":1,"Email":"test@test.com"}}');
        await session.onLogout();
        assert.equal(session.session_id, null);
        mock.reset();
        //session.deleteData();
    });*/

    it('should use session storage', () => {
        const storage = window.sessionStorage;
        const session = new Session(storage, config);
        assert.equal(storage.getItem("session_storage"), null);
        session.useSessionStorage();
        assert.equal(storage.getItem("session_storage"), "true");
        session.deleteData();
    });

    it('should use local storage', () => {
        const storage = window.localStorage;
        const session = new Session(storage, config);
        assert.equal(storage.getItem("session_storage"), null);
        session.useLocalStorage();
        assert.equal(storage.getItem("session_storage"), "false");
        session.deleteData();
    });

    it('should test redeem token', async () => {
        mock
            .onPost("session").reply(200, {id: "123", data: {token: "123token"}});
        const storage = window.localStorage;
        const session = new Session(storage, config);
        assert.equal(session.data, null);
        await session.redeemToken("token123");
        assert.equal(session.data.token, "123token");
        mock.reset();
        session.deleteData();
    });
});
