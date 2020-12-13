import Config from "common/config";
import MockAdapter from "axios-mock-adapter";
import Api from "common/api";
import StorageShim from "node-storage-shim";

window.__CONFIG__ = {"name":"PhotoPrism","version":"201213-283748ca-Linux-x86_64-DEBUG","copyright":"(c) 2018-2020 Michael Mayer \u003chello@photoprism.org\u003e","flags":"public debug experimental settings","siteUrl":"http://localhost:2342/","sitePreview":"http://localhost:2342/static/img/preview.jpg","siteTitle":"PhotoPrism","siteCaption":"Browse Your Life","siteDescription":"Open-Source Personal Photo Management","siteAuthor":"@browseyourlife","debug":true,"readonly":false,"uploadNSFW":false,"public":true,"experimental":true,"disableSettings":false,"albumCategories":null,"albums":[],"cameras":[{"ID":6,"Slug":"apple-iphone-5s","Name":"Apple iPhone 5s","Make":"Apple","Model":"iPhone 5s"},{"ID":7,"Slug":"apple-iphone-6","Name":"Apple iPhone 6","Make":"Apple","Model":"iPhone 6"},{"ID":8,"Slug":"apple-iphone-se","Name":"Apple iPhone SE","Make":"Apple","Model":"iPhone SE"},{"ID":2,"Slug":"canon-eos-5d","Name":"Canon EOS 5D","Make":"Canon","Model":"EOS 5D"},{"ID":4,"Slug":"canon-eos-6d","Name":"Canon EOS 6D","Make":"Canon","Model":"EOS 6D"},{"ID":5,"Slug":"canon-eos-7d","Name":"Canon EOS 7D","Make":"Canon","Model":"EOS 7D"},{"ID":9,"Slug":"huawei-p30","Name":"HUAWEI P30","Make":"HUAWEI","Model":"P30"},{"ID":3,"Slug":"olympus-c2500l","Name":"Olympus C2500L","Make":"Olympus","Model":"C2500L"},{"ID":1,"Slug":"zz","Name":"Unknown","Make":"","Model":"Unknown"}],"lenses":[{"ID":2,"Slug":"24-0-105-0-mm","Name":"24.0 - 105.0 mm","Make":"","Model":"24.0 - 105.0 mm","Type":""},{"ID":7,"Slug":"apple-iphone-5s-back-camera-4-12mm-f-2-2","Name":"Apple iPhone 5s back camera 4.12mm f/2.2","Make":"Apple","Model":"iPhone 5s back camera 4.12mm f/2.2","Type":""},{"ID":9,"Slug":"apple-iphone-6-back-camera-4-15mm-f-2-2","Name":"Apple iPhone 6 back camera 4.15mm f/2.2","Make":"Apple","Model":"iPhone 6 back camera 4.15mm f/2.2","Type":""},{"ID":10,"Slug":"apple-iphone-se-back-camera-4-15mm-f-2-2","Name":"Apple iPhone SE back camera 4.15mm f/2.2","Make":"Apple","Model":"iPhone SE back camera 4.15mm f/2.2","Type":""},{"ID":3,"Slug":"ef100mm-f-2-8l-macro-is-usm","Name":"EF100mm f/2.8L Macro IS USM","Make":"","Model":"EF100mm f/2.8L Macro IS USM","Type":""},{"ID":6,"Slug":"ef16-35mm-f-2-8l-ii-usm","Name":"EF16-35mm f/2.8L II USM","Make":"","Model":"EF16-35mm f/2.8L II USM","Type":""},{"ID":4,"Slug":"ef24-105mm-f-4l-is-usm","Name":"EF24-105mm f/4L IS USM","Make":"","Model":"EF24-105mm f/4L IS USM","Type":""},{"ID":8,"Slug":"ef35mm-f-2-is-usm","Name":"EF35mm f/2 IS USM","Make":"","Model":"EF35mm f/2 IS USM","Type":""},{"ID":5,"Slug":"ef70-200mm-f-4l-is-usm","Name":"EF70-200mm f/4L IS USM","Make":"","Model":"EF70-200mm f/4L IS USM","Type":""},{"ID":1,"Slug":"zz","Name":"Unknown","Make":"","Model":"Unknown","Type":""}],"countries":[{"ID":"at","Slug":"austria","Name":"Austria"},{"ID":"bw","Slug":"botswana","Name":"Botswana"},{"ID":"ca","Slug":"canada","Name":"Canada"},{"ID":"cu","Slug":"cuba","Name":"Cuba"},{"ID":"fr","Slug":"france","Name":"France"},{"ID":"de","Slug":"germany","Name":"Germany"},{"ID":"gr","Slug":"greece","Name":"Greece"},{"ID":"it","Slug":"italy","Name":"Italy"},{"ID":"za","Slug":"south-africa","Name":"South Africa"},{"ID":"ch","Slug":"switzerland","Name":"Switzerland"},{"ID":"gb","Slug":"united-kingdom","Name":"United Kingdom"},{"ID":"us","Slug":"usa","Name":"USA"},{"ID":"zz","Slug":"zz","Name":"Unknown"}],"thumbs":[{"size":"fit_720","use":"Mobile, TV","w":720,"h":720},{"size":"fit_1280","use":"Mobile, HD Ready TV","w":1280,"h":1024},{"size":"fit_1920","use":"Mobile, Full HD TV","w":1920,"h":1200},{"size":"fit_2048","use":"Tablets, Cinema 2K","w":2048,"h":2048},{"size":"fit_2560","use":"Quad HD, Retina Display","w":2560,"h":1600},{"size":"fit_4096","use":"Ultra HD, Retina 4K","w":4096,"h":4096},{"size":"fit_7680","use":"8K Ultra HD 2, Retina 6K","w":7680,"h":4320}],"status":"unregistered","mapKey":"jOTd5JGKYQV1fiAW4UZO","downloadToken":"2lbh9x09","previewToken":"public","jsHash":"6d752756","cssHash":"c5bb9de2","settings":{"ui":{"scrollbar":true,"theme":"default","language":"en"},"templates":{"default":"index.tmpl"},"maps":{"animate":0,"style":"streets"},"features":{"upload":true,"download":true,"private":true,"review":true,"files":true,"moments":true,"labels":true,"places":true,"edit":true,"archive":true,"delete":false,"share":true,"library":true,"import":true,"logs":true},"import":{"path":"/","move":false},"index":{"path":"/","convert":true,"rescan":false},"stack":{"uuid":true,"meta":true,"name":true},"share":{"title":""}},"count":{"cameras":8,"lenses":9,"countries":12,"photos":105,"videos":1,"hidden":0,"favorites":12,"private":1,"review":2,"stories":0,"albums":0,"moments":5,"months":38,"folders":3,"files":211,"places":44,"states":20,"labels":36,"labelMaxPhotos":12},"pos":{"uid":"pql8ug81ssr670tu","cid":"s2:47a85a624184","utc":"2020-08-31T16:03:10Z","lat":52.4525,"lng":13.3092},"years":[2020,2019,2018,2017,2016,2015,2014,2013,2012,2011,2010,2002],"colors":[{"Example":"#AB47BC","Name":"Purple","Slug":"purple"},{"Example":"#FF00FF","Name":"Magenta","Slug":"magenta"},{"Example":"#EC407A","Name":"Pink","Slug":"pink"},{"Example":"#EF5350","Name":"Red","Slug":"red"},{"Example":"#FFA726","Name":"Orange","Slug":"orange"},{"Example":"#D4AF37","Name":"Gold","Slug":"gold"},{"Example":"#FDD835","Name":"Yellow","Slug":"yellow"},{"Example":"#CDDC39","Name":"Lime","Slug":"lime"},{"Example":"#66BB6A","Name":"Green","Slug":"green"},{"Example":"#009688","Name":"Teal","Slug":"teal"},{"Example":"#00BCD4","Name":"Cyan","Slug":"cyan"},{"Example":"#2196F3","Name":"Blue","Slug":"blue"},{"Example":"#A1887F","Name":"Brown","Slug":"brown"},{"Example":"#F5F5F5","Name":"White","Slug":"white"},{"Example":"#9E9E9E","Name":"Grey","Slug":"grey"},{"Example":"#212121","Name":"Black","Slug":"black"}],"categories":[{"UID":"lql8ufjaz5laqysi","Slug":"animal","Name":"Animal"},{"UID":"lql8ufy1doog8nj1","Slug":"architecture","Name":"Architecture"},{"UID":"lql8ug744yjdmak5","Slug":"beach","Name":"Beach"},{"UID":"lql8ufo2s9bjihwi","Slug":"beetle","Name":"Beetle"},{"UID":"lql8ufk3uptwjvg3","Slug":"bird","Name":"Bird"},{"UID":"lql8ufj3j6r3w5nm","Slug":"building","Name":"Building"},{"UID":"lql8ufi2gjxy6361","Slug":"car","Name":"Car"},{"UID":"lql8ufj1sf295op8","Slug":"cat","Name":"Cat"},{"UID":"lql8ufs36d538t93","Slug":"farm","Name":"Farm"},{"UID":"lql8ufj1lfp81sdt","Slug":"insect","Name":"Insect"},{"UID":"lql8ufi1th1xe3te","Slug":"landscape","Name":"Landscape"},{"UID":"lql8uft3u38k3yzm","Slug":"monkey","Name":"Monkey"},{"UID":"lql8ufi1g2tjncrj","Slug":"mountains","Name":"Mountains"},{"UID":"lql8ufi3fn64dtpn","Slug":"nature","Name":"Nature"},{"UID":"lql8ufi204pq0lr1","Slug":"plant","Name":"Plant"},{"UID":"lql8ug118d50qdln","Slug":"reptile","Name":"Reptile"},{"UID":"lql8ufi1ucnpvw4w","Slug":"shop","Name":"Shop"},{"UID":"lql8ufv1zz33ie3d","Slug":"snow","Name":"Snow"},{"UID":"lql8ufy1bpizwiah","Slug":"tower","Name":"Tower"},{"UID":"lql8ufiv8ks762y1","Slug":"vehicle","Name":"Vehicle"},{"UID":"lql8ug02bt57syc8","Slug":"water","Name":"Water"},{"UID":"lql8ufo19zyipy8i","Slug":"wildlife","Name":"Wildlife"}],"clip":160,"server":{"cores":6,"routines":16,"memory":{"used":358461896,"reserved":562501688,"info":"Used 358 MB / Reserved 562 MB"}}};

let chai = require("chai/chai");
let assert = chai.assert;

const config2 = new Config(new StorageShim(), window.__CONFIG__);

describe("common/config", () => {

    const mock = new MockAdapter(Api);

    it("should get all config values",  () => {
        const storage = new StorageShim();
        const values = {siteTitle: "Foo", name: "testConfig", year: "2300"};

        const config = new Config(storage, values);
        const result = config.getValues();
        assert.equal(result.name, "testConfig");
    });

    it("should set multiple config values",  () => {
        const storage = new StorageShim();
        const values = {siteTitle: "Foo", country: "Germany", city: "Hamburg", settings: {"ui": {language:"de", theme: "default"}}};
        const newValues = {siteTitle: "Foo", new: "xxx", city: "Berlin", debug: true, settings: {"ui": {language:"en", theme: "lavender"}}};
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

    it("should store values",  () => {
        const storage = new StorageShim();
        const values = {siteTitle: "Foo", country: "Germany", city: "Hamburg"};
        const config = new Config(storage, values);
        assert.equal(config.storage["config"], undefined);
        config.storeValues();
        const expected = '{"siteTitle":"Foo","country":"Germany","city":"Hamburg"}';
        assert.equal(config.storage["config"], expected);
    });

    it("should set and get single config value",  () => {
        const storage = new StorageShim();
        const values = {};

        const config = new Config(storage, values);
        config.set("city", "Berlin");
        const result = config.get("city");
        assert.equal(result, "Berlin");
    });

    it("should return settings",  () => {
        const result = config2.settings();
        assert.equal(result.ui.theme, "default");
        assert.equal(result.ui.language, "en");
    });

    it("should return feature",  () => {
        assert.equal(config2.feature("places"), true);
        assert.equal(config2.feature("download"), true);
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
