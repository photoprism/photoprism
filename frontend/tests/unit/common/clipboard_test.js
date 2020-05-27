import Clipboard from "common/clipboard";
import Photo from "model/photo";
import Album from "model/album";

window.__CONFIG__ = {"flags":"public debug experimental settings","name":"PhotoPrism","url":"http://localhost:2342/","title":"PhotoPrism","subtitle":"Browse your life","description":"Personal Photo Management","author":"PhotoPrism.org","version":"200527-5453cf2-Linux-x86_64-DEBUG","copyright":"(c) 2018-2020 PhotoPrism.org \u003chello@photoprism.org\u003e","debug":true,"readonly":false,"uploadNSFW":false,"public":true,"experimental":true,"disableSettings":false,"albums":[],"cameras":[{"ID":30003,"Slug":"apple-iphone-6","Model":"iPhone 6","Make":"Apple"},{"ID":30001,"Slug":"apple-iphone-se","Model":"iPhone SE","Make":"Apple"},{"ID":30004,"Slug":"canon-eos-6d","Model":"EOS 6D","Make":"Canon"},{"ID":30002,"Slug":"canon-eos-m6","Model":"EOS M6","Make":"Canon"},{"ID":30006,"Slug":"huawei-ele-l29","Model":"ELE-L29","Make":"HUAWEI"},{"ID":30005,"Slug":"motorola-moto-g-4","Model":"Moto G (4)","Make":"Motorola"},{"ID":1,"Slug":"zz","Model":"Unknown","Make":""}],"lenses":[{"ID":30003,"Slug":"22-0-mm","Model":"22.0 mm","Make":"","Type":""},{"ID":30005,"Slug":"ef16-35mm-f-2-8l-ii-usm","Model":"EF16-35mm f/2.8L II USM","Make":"","Type":""},{"ID":30004,"Slug":"iphone-6-back-camera-4-15mm-f-2-2","Model":"iPhone 6 back camera 4.15mm f/2.2","Make":"Apple","Type":""},{"ID":30001,"Slug":"iphone-se-back-camera-4-15mm-f-2-2","Model":"iPhone SE back camera 4.15mm f/2.2","Make":"Apple","Type":""},{"ID":30002,"Slug":"iphone-se-front-camera-2-15mm-f-2-4","Model":"iPhone SE front camera 2.15mm f/2.4","Make":"Apple","Type":""},{"ID":1,"Slug":"zz","Model":"Unknown","Make":"","Type":""}],"countries":[{"ID":"at","Slug":"austria","Name":"Austria"},{"ID":"bw","Slug":"botswana","Name":"Botswana"},{"ID":"ca","Slug":"canada","Name":"Canada"},{"ID":"fr","Slug":"france","Name":"France"},{"ID":"de","Slug":"germany","Name":"Germany"},{"ID":"gr","Slug":"greece","Name":"Greece"},{"ID":"za","Slug":"south-africa","Name":"South Africa"},{"ID":"us","Slug":"usa","Name":"USA"},{"ID":"zz","Slug":"zz","Name":"Unknown"}],"thumbnails":[{"Name":"fit_720","Width":720,"Height":720},{"Name":"fit_1280","Width":1280,"Height":1024},{"Name":"fit_2560","Width":2560,"Height":1600},{"Name":"fit_3840","Width":3840,"Height":2400},{"Name":"fit_1920","Width":1920,"Height":1200},{"Name":"fit_2048","Width":2048,"Height":2048}],"downloadToken":"2y71e0sr","thumbToken":"static","jsHash":"14ba2de4","cssHash":"2b327230","settings":{"theme":"default","language":"en","templates":{"default":"index.tmpl"},"maps":{"animate":0,"style":"streets"},"features":{"archive":true,"private":true,"review":true,"upload":true,"import":true,"files":true,"moments":true,"labels":true,"places":true,"download":true,"edit":true,"share":true,"logs":true},"import":{"path":"/","move":false},"index":{"path":"/","convert":true,"rescan":false,"group":true}},"count":{"photos":385,"videos":1,"hidden":0,"favorites":1,"private":2,"review":4,"stories":0,"albums":0,"folders":14,"files":394,"moments":0,"countries":8,"places":0,"labels":46,"labelMaxPhotos":54},"pos":{"uid":"pqazcltc1x8d12lo","loc":"4777dc437584","utc":"2020-02-14T12:44:19Z","lat":47.207123,"lng":11.823489},"years":[2020,2019,2018,2017,2016],"colors":[{"Example":"#AB47BC","Name":"Purple","Slug":"purple"},{"Example":"#FF00FF","Name":"Magenta","Slug":"magenta"},{"Example":"#EC407A","Name":"Pink","Slug":"pink"},{"Example":"#EF5350","Name":"Red","Slug":"red"},{"Example":"#FFA726","Name":"Orange","Slug":"orange"},{"Example":"#D4AF37","Name":"Gold","Slug":"gold"},{"Example":"#FDD835","Name":"Yellow","Slug":"yellow"},{"Example":"#CDDC39","Name":"Lime","Slug":"lime"},{"Example":"#66BB6A","Name":"Green","Slug":"green"},{"Example":"#009688","Name":"Teal","Slug":"teal"},{"Example":"#00BCD4","Name":"Cyan","Slug":"cyan"},{"Example":"#2196F3","Name":"Blue","Slug":"blue"},{"Example":"#A1887F","Name":"Brown","Slug":"brown"},{"Example":"#F5F5F5","Name":"White","Slug":"white"},{"Example":"#9E9E9E","Name":"Grey","Slug":"grey"},{"Example":"#212121","Name":"Black","Slug":"black"}],"categories":[{"UID":"lqazz283gqjo05j9","Slug":"aircraft","Name":"Aircraft"},{"UID":"lqazyyc2xos6k0op","Slug":"airport","Name":"Airport"},{"UID":"lqazyya2wbw3045h","Slug":"animal","Name":"Animal"},{"UID":"lqazyz22d0y1ham3","Slug":"architecture","Name":"Architecture"},{"UID":"lqazz7537y79uhef","Slug":"beach","Name":"Beach"},{"UID":"lqazz1v2gbroth1y","Slug":"beverage","Name":"Beverage"},{"UID":"lqazyyf213ls8byk","Slug":"building","Name":"Building"},{"UID":"lqazyzhjuf6ud8pd","Slug":"car","Name":"Car"},{"UID":"lqazzcnzc5ejq4xx","Slug":"dining","Name":"Dining"},{"UID":"lqazz1v3t6kuuid7","Slug":"drinks","Name":"Drinks"},{"UID":"lqazz4j3rrxrh9el","Slug":"event","Name":"Event"},{"UID":"lqazz422nbeeedv5","Slug":"farm","Name":"Farm"},{"UID":"lqazz5f3leym5l14","Slug":"food","Name":"Food"},{"UID":"lqazz252fhe2ibx1","Slug":"landscape","Name":"Landscape"},{"UID":"lqazyzs20lrgueeb","Slug":"nature","Name":"Nature"},{"UID":"lqazyyh3f6phuq04","Slug":"outdoor","Name":"Outdoor"},{"UID":"lqazzef1n086vptr","Slug":"people","Name":"People"},{"UID":"lqazzm019mojdmcp","Slug":"plant","Name":"Plant"},{"UID":"lqazyy72zbezq9zt","Slug":"portrait","Name":"Portrait"},{"UID":"lqazzbqrfzbq0l9l","Slug":"sand","Name":"Sand"},{"UID":"lqazz8q1q39ksabl","Slug":"shop","Name":"Shop"},{"UID":"lqazyyt1ua4i8jpy","Slug":"train","Name":"Train"},{"UID":"lqazzdxvp4d59tx4","Slug":"vegetables","Name":"Vegetables"},{"UID":"lqazyytyvhnb6fvk","Slug":"vehicle","Name":"Vehicle"},{"UID":"lqazz4b7rukovpmc","Slug":"water","Name":"Water"},{"UID":"lqazz2t3n882fps6","Slug":"wildlife","Name":"Wildlife"}],"clip":160,"server":{"cores":2,"routines":38,"memory":{"used":56042952,"reserved":144132360,"info":"Used 56 MB / Reserved 144 MB"}}};

let chai = require('../../../node_modules/chai/chai');
let assert = chai.assert;
let sinon = require("sinon");

describe("common/clipboard", () => {
    it("should construct clipboard",  () => {
        const storage = window.localStorage;
        const key = "clipboard";

        const clipboard = new Clipboard(storage, key);
        assert.equal(clipboard.storageKey, "clipboard");
        assert.equal(clipboard.selection, "");
    });

    it("should toggle model",  () => {
        let spy = sinon.spy(console, "log");
        const storage = window.localStorage;
        const key = "clipboard";

        const clipboard = new Clipboard(storage, key);
        clipboard.clear();
        clipboard.toggle();
        assert(spy.calledWith("Clipboard::toggle() - not a model:"));
        assert.equal(clipboard.storageKey, "clipboard");
        assert.equal(clipboard.selection, "");

        const values = {ID: 5, UID: "ABC123", Title: "Crazy Cat"};
        const photo = new Photo(values);
        clipboard.toggle(photo);
        assert.equal(clipboard.selection[0], "ABC123");
        const values2 = {ID: 8, UID: "ABC124", Title: "Crazy Cat"};
        const photo2 = new Photo(values2);
        clipboard.toggle(photo2);
        assert.equal(clipboard.selection[0], "ABC123");
        assert.equal(clipboard.selection[1], "ABC124");
        clipboard.toggle(photo);
        assert.equal(clipboard.selection[0], "ABC124");
        console.log.restore();
    });

    it("should toggle id",  () => {
        const storage = window.localStorage;
        const key = "clipboard";

        const clipboard = new Clipboard(storage, key);
        clipboard.clear();
        clipboard.toggleId(3);
        assert.equal(clipboard.selection[0], 3);
        clipboard.toggleId(3);
        assert.equal(clipboard.selection, "");
    });

    it("should add model",  () => {
        let spy = sinon.spy(console, "log");
        const storage = window.localStorage;
        const key = "clipboard";

        const clipboard = new Clipboard(storage, key);
        clipboard.clear();
        clipboard.add();
        assert.equal(clipboard.storageKey, "clipboard");
        assert.equal(clipboard.selection, "");
        assert(spy.calledWith("Clipboard::add() - not a model:"));

        const values = {ID: 5, UID: "ABC124", Title: "Crazy Cat"};
        const photo = new Photo(values);
        clipboard.add(photo);
        assert.equal(clipboard.selection[0], "ABC124");
        clipboard.add(photo);
        assert.equal(clipboard.selection[0], "ABC124");
        console.log.restore();
    });

    it("should add id",  () => {
        const storage = window.localStorage;
        const key = "clipboard";

        const clipboard = new Clipboard(storage, key);
        clipboard.clear();
        clipboard.addId(99);
        assert.equal(clipboard.selection[0], 99);
    });

    it("should test whether clipboard has model",  () => {
        let spy = sinon.spy(console, "log");
        const storage = window.localStorage;
        const key = "clipboard";

        const clipboard = new Clipboard(storage, key);
        clipboard.clear();
        clipboard.has();
        assert.equal(clipboard.storageKey, "clipboard");
        assert.equal(clipboard.selection, "");
        assert(spy.calledWith("Clipboard::has() - not a model:"));

        const values = {ID: 5, UID: "ABC124", Title: "Crazy Cat"};
        const photo = new Photo(values);
        clipboard.add(photo);
        assert.equal(clipboard.selection[0], "ABC124");
        const result = clipboard.has(photo);
        assert.equal(result, true);
        const values2 = {ID: 5, Title: "Christmas 2019", Slug: "christmas-2019", UID: 66};
        const album = new Album(values2);
        const result2 = clipboard.has(album);
        assert.equal(result2, false);
        console.log.restore();
    });

    it("should test whether clipboard has id",  () => {
        const storage = window.localStorage;
        const key = "clipboard";

        const clipboard = new Clipboard(storage, key);
        clipboard.clear();
        clipboard.addId(77);
        assert.equal(clipboard.hasId(77), true);
        assert.equal(clipboard.hasId(78), false);
    });

    it("should remove model",  () => {
        let spy = sinon.spy(console, "log");
        const storage = window.localStorage;
        const key = "clipboard";

        const clipboard = new Clipboard(storage, key);
        clipboard.clear();
        clipboard.remove();
        assert.equal(clipboard.storageKey, "clipboard");
        assert.equal(clipboard.selection, "");
        assert(spy.calledWith("Clipboard::remove() - not a model:"));

        const values = {ID: 5, UID: "ABC123", Title: "Crazy Cat"};
        const photo = new Photo(values);
        clipboard.add(photo);
        assert.equal(clipboard.selection[0], "ABC123");

        clipboard.remove(photo);
        assert.equal(clipboard.selection, "");
        const values2 = {ID: 5, Name: "Christmas 2019", Slug: "christmas-2019", UID: 66};
        const album = new Album(values2);
        clipboard.remove(album);
        assert.equal(clipboard.selection, "");
        console.log.restore();
    });

    it("should set and get ids",  () => {
        const storage = window.localStorage;
        const key = "clipboard";

        const clipboard = new Clipboard(storage, key);
        clipboard.clear();
        clipboard.setIds(8);
        assert.equal(clipboard.selection, "");
        clipboard.setIds([5, 6, 9]);
        assert.equal(clipboard.selection[0], 5);
        assert.equal(clipboard.selection[2], 9);
        const result = clipboard.getIds();
        assert.equal(result[1], 6);
        assert.equal(result.length, 3);
    });

    it("should clear",  () => {
        const storage = window.localStorage;
        const key = "clipboard";

        const clipboard = new Clipboard(storage, key);
        clipboard.clear();
        clipboard.setIds([5, 6, 9]);
        assert.equal(clipboard.selection[0], 5);
        clipboard.clear();
        assert.equal(clipboard.selection, "");
    });

});
