import Photo from "model/photo";
import MockAdapter from "axios-mock-adapter";
import Api from "common/api";

window.__CONFIG__ = {"name":"PhotoPrism","version":"201213-283748ca-Linux-x86_64-DEBUG","copyright":"(c) 2018-2020 Michael Mayer \u003chello@photoprism.org\u003e","flags":"public debug experimental settings","siteUrl":"http://localhost:2342/","sitePreview":"http://localhost:2342/static/img/preview.jpg","siteTitle":"PhotoPrism","siteCaption":"Browse Your Life","siteDescription":"Open-Source Personal Photo Management","siteAuthor":"@browseyourlife","debug":true,"readonly":false,"uploadNSFW":false,"public":true,"experimental":true,"disableSettings":false,"albumCategories":null,"albums":[],"cameras":[{"ID":6,"Slug":"apple-iphone-5s","Name":"Apple iPhone 5s","Make":"Apple","Model":"iPhone 5s"},{"ID":7,"Slug":"apple-iphone-6","Name":"Apple iPhone 6","Make":"Apple","Model":"iPhone 6"},{"ID":8,"Slug":"apple-iphone-se","Name":"Apple iPhone SE","Make":"Apple","Model":"iPhone SE"},{"ID":2,"Slug":"canon-eos-5d","Name":"Canon EOS 5D","Make":"Canon","Model":"EOS 5D"},{"ID":4,"Slug":"canon-eos-6d","Name":"Canon EOS 6D","Make":"Canon","Model":"EOS 6D"},{"ID":5,"Slug":"canon-eos-7d","Name":"Canon EOS 7D","Make":"Canon","Model":"EOS 7D"},{"ID":9,"Slug":"huawei-p30","Name":"HUAWEI P30","Make":"HUAWEI","Model":"P30"},{"ID":3,"Slug":"olympus-c2500l","Name":"Olympus C2500L","Make":"Olympus","Model":"C2500L"},{"ID":1,"Slug":"zz","Name":"Unknown","Make":"","Model":"Unknown"}],"lenses":[{"ID":2,"Slug":"24-0-105-0-mm","Name":"24.0 - 105.0 mm","Make":"","Model":"24.0 - 105.0 mm","Type":""},{"ID":7,"Slug":"apple-iphone-5s-back-camera-4-12mm-f-2-2","Name":"Apple iPhone 5s back camera 4.12mm f/2.2","Make":"Apple","Model":"iPhone 5s back camera 4.12mm f/2.2","Type":""},{"ID":9,"Slug":"apple-iphone-6-back-camera-4-15mm-f-2-2","Name":"Apple iPhone 6 back camera 4.15mm f/2.2","Make":"Apple","Model":"iPhone 6 back camera 4.15mm f/2.2","Type":""},{"ID":10,"Slug":"apple-iphone-se-back-camera-4-15mm-f-2-2","Name":"Apple iPhone SE back camera 4.15mm f/2.2","Make":"Apple","Model":"iPhone SE back camera 4.15mm f/2.2","Type":""},{"ID":3,"Slug":"ef100mm-f-2-8l-macro-is-usm","Name":"EF100mm f/2.8L Macro IS USM","Make":"","Model":"EF100mm f/2.8L Macro IS USM","Type":""},{"ID":6,"Slug":"ef16-35mm-f-2-8l-ii-usm","Name":"EF16-35mm f/2.8L II USM","Make":"","Model":"EF16-35mm f/2.8L II USM","Type":""},{"ID":4,"Slug":"ef24-105mm-f-4l-is-usm","Name":"EF24-105mm f/4L IS USM","Make":"","Model":"EF24-105mm f/4L IS USM","Type":""},{"ID":8,"Slug":"ef35mm-f-2-is-usm","Name":"EF35mm f/2 IS USM","Make":"","Model":"EF35mm f/2 IS USM","Type":""},{"ID":5,"Slug":"ef70-200mm-f-4l-is-usm","Name":"EF70-200mm f/4L IS USM","Make":"","Model":"EF70-200mm f/4L IS USM","Type":""},{"ID":1,"Slug":"zz","Name":"Unknown","Make":"","Model":"Unknown","Type":""}],"countries":[{"ID":"at","Slug":"austria","Name":"Austria"},{"ID":"bw","Slug":"botswana","Name":"Botswana"},{"ID":"ca","Slug":"canada","Name":"Canada"},{"ID":"cu","Slug":"cuba","Name":"Cuba"},{"ID":"fr","Slug":"france","Name":"France"},{"ID":"de","Slug":"germany","Name":"Germany"},{"ID":"gr","Slug":"greece","Name":"Greece"},{"ID":"it","Slug":"italy","Name":"Italy"},{"ID":"za","Slug":"south-africa","Name":"South Africa"},{"ID":"ch","Slug":"switzerland","Name":"Switzerland"},{"ID":"gb","Slug":"united-kingdom","Name":"United Kingdom"},{"ID":"us","Slug":"usa","Name":"USA"},{"ID":"zz","Slug":"zz","Name":"Unknown"}],"thumbs":[{"size":"fit_720","use":"Mobile, TV","w":720,"h":720},{"size":"fit_1280","use":"Mobile, HD Ready TV","w":1280,"h":1024},{"size":"fit_1920","use":"Mobile, Full HD TV","w":1920,"h":1200},{"size":"fit_2048","use":"Tablets, Cinema 2K","w":2048,"h":2048},{"size":"fit_2560","use":"Quad HD, Retina Display","w":2560,"h":1600},{"size":"fit_4096","use":"Ultra HD, Retina 4K","w":4096,"h":4096},{"size":"fit_7680","use":"8K Ultra HD 2, Retina 6K","w":7680,"h":4320}],"status":"unregistered","mapKey":"jOTd5JGKYQV1fiAW4UZO","downloadToken":"2lbh9x09","previewToken":"public","jsHash":"6d752756","cssHash":"c5bb9de2","settings":{"ui":{"scrollbar":true,"theme":"default","language":"en"},"templates":{"default":"index.tmpl"},"maps":{"animate":0,"style":"streets"},"features":{"upload":true,"download":true,"private":true,"review":true,"files":true,"moments":true,"labels":true,"places":true,"edit":true,"archive":true,"delete":false,"share":true,"library":true,"import":true,"logs":true},"import":{"path":"/","move":false},"index":{"path":"/","convert":true,"rescan":false},"stack":{"uuid":true,"meta":true,"name":true},"share":{"title":""}},"count":{"cameras":8,"lenses":9,"countries":12,"photos":105,"videos":1,"hidden":0,"favorites":12,"private":1,"review":2,"stories":0,"albums":0,"moments":5,"months":38,"folders":3,"files":211,"places":44,"states":20,"labels":36,"labelMaxPhotos":12},"pos":{"uid":"pql8ug81ssr670tu","cid":"s2:47a85a624184","utc":"2020-08-31T16:03:10Z","lat":52.4525,"lng":13.3092},"years":[2020,2019,2018,2017,2016,2015,2014,2013,2012,2011,2010,2002],"colors":[{"Example":"#AB47BC","Name":"Purple","Slug":"purple"},{"Example":"#FF00FF","Name":"Magenta","Slug":"magenta"},{"Example":"#EC407A","Name":"Pink","Slug":"pink"},{"Example":"#EF5350","Name":"Red","Slug":"red"},{"Example":"#FFA726","Name":"Orange","Slug":"orange"},{"Example":"#D4AF37","Name":"Gold","Slug":"gold"},{"Example":"#FDD835","Name":"Yellow","Slug":"yellow"},{"Example":"#CDDC39","Name":"Lime","Slug":"lime"},{"Example":"#66BB6A","Name":"Green","Slug":"green"},{"Example":"#009688","Name":"Teal","Slug":"teal"},{"Example":"#00BCD4","Name":"Cyan","Slug":"cyan"},{"Example":"#2196F3","Name":"Blue","Slug":"blue"},{"Example":"#A1887F","Name":"Brown","Slug":"brown"},{"Example":"#F5F5F5","Name":"White","Slug":"white"},{"Example":"#9E9E9E","Name":"Grey","Slug":"grey"},{"Example":"#212121","Name":"Black","Slug":"black"}],"categories":[{"UID":"lql8ufjaz5laqysi","Slug":"animal","Name":"Animal"},{"UID":"lql8ufy1doog8nj1","Slug":"architecture","Name":"Architecture"},{"UID":"lql8ug744yjdmak5","Slug":"beach","Name":"Beach"},{"UID":"lql8ufo2s9bjihwi","Slug":"beetle","Name":"Beetle"},{"UID":"lql8ufk3uptwjvg3","Slug":"bird","Name":"Bird"},{"UID":"lql8ufj3j6r3w5nm","Slug":"building","Name":"Building"},{"UID":"lql8ufi2gjxy6361","Slug":"car","Name":"Car"},{"UID":"lql8ufj1sf295op8","Slug":"cat","Name":"Cat"},{"UID":"lql8ufs36d538t93","Slug":"farm","Name":"Farm"},{"UID":"lql8ufj1lfp81sdt","Slug":"insect","Name":"Insect"},{"UID":"lql8ufi1th1xe3te","Slug":"landscape","Name":"Landscape"},{"UID":"lql8uft3u38k3yzm","Slug":"monkey","Name":"Monkey"},{"UID":"lql8ufi1g2tjncrj","Slug":"mountains","Name":"Mountains"},{"UID":"lql8ufi3fn64dtpn","Slug":"nature","Name":"Nature"},{"UID":"lql8ufi204pq0lr1","Slug":"plant","Name":"Plant"},{"UID":"lql8ug118d50qdln","Slug":"reptile","Name":"Reptile"},{"UID":"lql8ufi1ucnpvw4w","Slug":"shop","Name":"Shop"},{"UID":"lql8ufv1zz33ie3d","Slug":"snow","Name":"Snow"},{"UID":"lql8ufy1bpizwiah","Slug":"tower","Name":"Tower"},{"UID":"lql8ufiv8ks762y1","Slug":"vehicle","Name":"Vehicle"},{"UID":"lql8ug02bt57syc8","Slug":"water","Name":"Water"},{"UID":"lql8ufo19zyipy8i","Slug":"wildlife","Name":"Wildlife"}],"clip":160,"server":{"cores":6,"routines":16,"memory":{"used":358461896,"reserved":562501688,"info":"Used 358 MB / Reserved 562 MB"}}};

let chai = require("chai/chai");
let assert = chai.assert;

const mock = new MockAdapter(Api);

mock
    .onPost("batch/photos/archive").reply(200, {"photos": [1, 3]})
    .onPost("api/v1/photos/pqbemz8276mhtobh/approve").reply(200, {})
    .onPost("api/v1/photos/pqbemz8276mhtobh/files/fqbfk181n4ca5sud/primary").reply(200,
    {
        ID: 10,
        UID: "pqbemz8276mhtobh",
        Files: [{
        UID: "fqbfk181n4ca5sud",
        Name: "1980/01/superCuteKitten.mp4",
        Primary: true,
        Type: "mp4",
        Hash: "1xxbgdt55"}]})
    .onPut("api/v1/photos/pqbemz8276mhtobh").reply(200,
    {
        ID: 10,
        UID: "pqbemz8276mhtobh",
        TitleSrc: "manual",
        Files: [{
            UID: "fqbfk181n4ca5sud",
            Name: "1980/01/superCuteKitten.mp4",
            Primary: false,
            Type: "mp4",
            Hash: "1xxbgdt55"}]})
    .onDelete("api/v1/photos/abc123/unlike").reply(200)
    .onDelete("api/v1/photos/pqbemz8276mhtobh/files/fqbfk181n4ca5sud").reply(200, {"success": "successfully deleted"})
    .onPost("api/v1/photos/pqbemz8276mhtobh/files/fqbfk181n4ca5sud/unstack").reply(200, {"success": "ok"})
    .onPost("api/v1/photos/pqbemz8276mhtobh/label", {Name: "Cat", Priority: 10}).reply(200, {"success": "ok"})
    .onPut("api/v1/photos/pqbemz8276mhtobh/label/12345", {Uncertainty: 0}).reply(200, {"success": "ok"})
    .onPut("api/v1/photos/pqbemz8276mhtobh/label/12345", {Label: {Name: "Sommer"}}).reply(200, {"success": "ok"})
    .onDelete("api/v1/photos/pqbemz8276mhtobh/label/12345").reply(200, {"success": "ok"});



describe("model/photo", () => {
    it("should get photo entity name",  () => {
        const values = {UID: 5, Title: "Crazy Cat"};
        const photo = new Photo(values);
        const result = photo.getEntityName();
        assert.equal(result, "Crazy Cat");
    });

    it("should get photo uuid",  () => {
        const values = {ID: 5, Title: "Crazy Cat", UID: 789};
        const photo = new Photo(values);
        const result = photo.getId();
        assert.equal(result, 789);
    });

    it("should get photo title",  () => {
        const values = {ID: 5, Title: "Crazy Cat", UID: 789};
        const photo = new Photo(values);
        const result = photo.getTitle();
        assert.equal(result, "Crazy Cat");
    });

    it("should get photo maps link",  () => {
        const values = {ID: 5, Title: "Crazy Cat", Lat: 36.442881666666665, Lng: 28.229493333333334};
        const photo = new Photo(values);
        const result = photo.getGoogleMapsLink();
        assert.equal(result, "https://www.google.com/maps/place/36.442881666666665,28.229493333333334");
    });

    it("should get photo thumbnail url",  () => {
        const values = {ID: 5, Title: "Crazy Cat", Hash: 345982};
        const photo = new Photo(values);
        const result = photo.thumbnailUrl("tile500");
        assert.equal(result, "/api/v1/t/345982/public/tile500");
        const values2 = {ID: 10, UID: "ABC127",
            Files: [{UID: "123fgb", Name: "1980/01/superCuteKitten.mp4", Primary: false, Type: "mp4", Width: 500, Height: 600, Hash: "1xxbgdt55"}]};
        const photo2 = new Photo(values2);
        const result2 = photo2.thumbnailUrl("tile500");
        assert.equal(result2, "/api/v1/t/1xxbgdt55/public/tile500");
        const values3 = {ID: 10, UID: "ABC127",
            Files: [{UID: "123fgb", Name: "1980/01/superCuteKitten.mp4", Primary: false, Type: "mp4", Width: 500, Height: 600}]};
        const photo3 = new Photo(values3);
        const result3 = photo3.thumbnailUrl("tile500");
        assert.equal(result3, "/api/v1/svg/photo");
    });

    it("should get photo download url",  () => {
        const values = {ID: 5, Title: "Crazy Cat", Hash: 345982};
        const photo = new Photo(values);
        const result = photo.getDownloadUrl();
        assert.equal(result, "/api/v1/dl/345982?t=2lbh9x09");
    });

    it("should calculate photo size",  () => {
        const values = {ID: 5, Title: "Crazy Cat", Width: 500,Height: 200};
        const photo = new Photo(values);
        const result = photo.calculateSize(500, 200);
        assert.equal(result.width, 500);
        assert.equal(result.height, 200);
    });

    it("should calculate photo size with srcAspectRatio < maxAspectRatio",  () => {
        const values = {ID: 5, Title: "Crazy Cat", Width: 500, Height: 200};
        const photo = new Photo(values);
        const result = photo.calculateSize(300, 50);
        assert.equal(result.width, 125);
        assert.equal(result.height, 50);
    });

    it("should calculate photo size with srcAspectRatio > maxAspectRatio",  () => {
        const values = {ID: 5, Title: "Crazy Cat", Width: 500, Height: 200};
        const photo = new Photo(values);
        const result = photo.calculateSize(400, 300);
        assert.equal(result.width, 400);
        assert.equal(result.height, 160);
    });

    it("should get local day string",  () => {
        const values = {ID: 5, Title: "Crazy Cat", TakenAt: "2012-07-08T14:45:39Z", TimeZone: "UTC"};
        const photo = new Photo(values);
        const result = photo.localDayString();
        // Current day of the month (changes):
        assert.equal(result.length, 2);
        const values2 = {ID: 5, Title: "Crazy Cat", TakenAtLocal: "2012-07-08T14:45:39Z", TakenAt: "2012-07-08T14:45:39Z", TimeZone: "UTC", Day: 8};
        const photo2 = new Photo(values2);
        const result2 = photo2.localDayString();
        assert.equal(result2, "08");
    });

    it("should get local month string",  () => {
        const values = {ID: 5, Title: "Crazy Cat", TakenAt: "2012-07-08T14:45:39Z", TimeZone: "UTC"};
        const photo = new Photo(values);
        const result = photo.localMonthString();
        assert.equal(result, (new Date().getMonth() + 1).toString().padStart(2, "0"));
        const values2 = {ID: 5, Title: "Crazy Cat", TakenAtLocal: "2012-07-08T14:45:39Z", TakenAt: "2012-07-08T14:45:39Z", TimeZone: "UTC", Month: 8};
        const photo2 = new Photo(values2);
        const result2 = photo2.localMonthString();
        assert.equal(result2, "08");
    });

    it("should get local year string",  () => {
        const values = {ID: 5, Title: "Crazy Cat", TakenAt: "2012-07-08T14:45:39Z", TimeZone: "UTC"};
        const photo = new Photo(values);
        const result = photo.localYearString();
        assert.equal(result, new Date().getFullYear().toString().padStart(4, "0"));
        const values2 = {ID: 5, Title: "Crazy Cat", TakenAtLocal: "2012-07-08T14:45:39Z", TakenAt: "2012-07-08T14:45:39Z", TimeZone: "UTC", Year: 2010};
        const photo2 = new Photo(values2);
        const result2 = photo2.localYearString();
        assert.equal(result2, "2010");
    });

    it("should get local date string",  () => {
        const values = {ID: 5, Title: "Crazy Cat", TakenAtLocal: "2012-07-08T14:45:39Z", TakenAt: "2012-07-08T14:45:39Z", TimeZone: "UTC"};
        const photo = new Photo(values);
        const result = photo.localDateString();
        assert.equal(result, "2012-07-08T14:45:39");
    });

    it("should get local date",  () => {
        const values = {ID: 5, Title: "Crazy Cat", TakenAt: "2012-07-08T14:45:39Z", TimeZone: "Indian/Reunion"};
        const photo = new Photo(values);
        const result = photo.localDate();
        assert.equal(String(result), "2012-07-08T14:45:39.000Z");
    });

    it("should get date string",  () => {
        const values = {ID: 5, Title: "Crazy Cat", TakenAtLocal: "2012-07-08T14:45:39Z", TakenAt: "2012-07-08T14:45:39Z", TimeZone: "UTC"};
        const photo = new Photo(values);
        const result = photo.getDateString();
        assert.equal(result, "July 8, 2012, 2:45 PM UTC");
        const values2 = {ID: 5, Title: "Crazy Cat", TakenAtLocal: "", TakenAt: "", TimeZone: "UTC"};
        const photo2 = new Photo(values2);
        const result2 = photo2.getDateString();
        assert.equal(result2, "Unknown");
        const values3 = {ID: 5, Title: "Crazy Cat", TakenAtLocal: "2012-07-08T14:45:39Z", TakenAt: "2012-07-08T14:45:39Z"};
        const photo3 = new Photo(values3);
        const result3 = photo3.getDateString();
        assert.equal(result3, "Sunday, July 8, 2012");
        const values4 = {ID: 5, Title: "Crazy Cat", TakenAtLocal: "2012-07-08T14:45:39Z", TakenAt: "2012-07-08T14:45:39Z", Month: -1};
        const photo4 = new Photo(values4);
        const result4 = photo4.getDateString();
        assert.equal(result4, "2012");
        const values5 = {ID: 5, Title: "Crazy Cat", TakenAtLocal: "2012-07-08T14:45:39Z", TakenAt: "2012-07-08T14:45:39Z", Day: -1};
        const photo5 = new Photo(values5);
        const result5 = photo5.getDateString();
        assert.equal(result5, "July 2012");
    });

    it("should get short date string",  () => {
        const values = {ID: 5, Title: "Crazy Cat", TakenAtLocal: "2012-07-08T14:45:39Z", TakenAt: "2012-07-08T14:45:39Z", TimeZone: "UTC"};
        const photo = new Photo(values);
        const result = photo.shortDateString();
        assert.equal(result, "Jul 8, 2012");
        const values2 = {ID: 5, Title: "Crazy Cat", TakenAtLocal: "", TakenAt: "", TimeZone: "UTC"};
        const photo2 = new Photo(values2);
        const result2 = photo2.shortDateString();
        assert.equal(result2, "Unknown");
        const values3 = {ID: 5, Title: "Crazy Cat", TakenAtLocal: "2012-07-08T14:45:39Z", TakenAt: "2012-07-08T14:45:39Z"};
        const photo3 = new Photo(values3);
        const result3 = photo3.shortDateString();
        assert.equal(result3, "Jul 8, 2012");
        const values4 = {ID: 5, Title: "Crazy Cat", TakenAtLocal: "2012-07-08T14:45:39Z", TakenAt: "2012-07-08T14:45:39Z", Month: -1};
        const photo4 = new Photo(values4);
        const result4 = photo4.shortDateString();
        assert.equal(result4, "2012");
        const values5 = {ID: 5, Title: "Crazy Cat", TakenAtLocal: "2012-07-08T14:45:39Z", TakenAt: "2012-07-08T14:45:39Z", Day: -1};
        const photo5 = new Photo(values5);
        const result5 = photo5.shortDateString();
        assert.equal(result5, "July 2012");
    });

    it("should test whether photo has location",  () => {
        const values = {ID: 5, Title: "Crazy Cat", Lat: 36.442881666666665, Lng: 28.229493333333334};
        const photo = new Photo(values);
        const result = photo.hasLocation();
        assert.equal(result, true);
    });

    it("should test whether photo has location",  () => {
        const values = {ID: 5, Title: "Crazy Cat", Lat: 0, Lng: 0};
        const photo = new Photo(values);
        const result = photo.hasLocation();
        assert.equal(result, false);
    });

    it("should get location",  () => {
        const values = {ID: 5, Title: "Crazy Cat", CellID: 6, CellCategory: "viewpoint", PlaceLabel: "Cape Point, South Africa", PlaceCountry: "South Africa"};
        const photo = new Photo(values);
        const result = photo.locationInfo();
        assert.equal(result, "Cape Point, South Africa");
    });

    it("should get location",  () => {
        const values = {ID: 5, Title: "Crazy Cat", CellID: 6, CellCategory: "viewpoint", PlaceLabel: "Cape Point, State, South Africa", PlaceCountry: "South Africa", PlaceCity: "Cape Town", PlaceCounty: "County", PlaceState: "State"};
        const photo = new Photo(values);
        const result = photo.locationInfo();
        assert.equal(result, "Cape Point, State, South Africa");
    });

    it("should get location",  () => {
        const values = {ID: 5, Title: "Crazy Cat", CellCategory: "viewpoint", CellName: "Cape Point", PlaceCountry: "Africa", PlaceCity: "Cape Town", PlaceCounty: "County", PlaceState: "State"};
        const photo = new Photo(values);
        const result = photo.locationInfo();
        assert.equal(result, "Unknown");
    });

    it("should get location",  () => {
        const values = {ID: 5, Title: "Crazy Cat", CountryName: "Africa", PlaceCity: "Cape Town"};
        const photo = new Photo(values);
        const result = photo.locationInfo();
        assert.equal(result, "Unknown");
    });

    it("should get camera",  () => {
        const values = {ID: 5, Title: "Crazy Cat", CameraModel: "EOSD10", CameraMake: "Canon"};
        const photo = new Photo(values);
        const result = photo.getCamera();
        assert.equal(result, "Canon EOSD10");
        const values2 = {
            ID: 10,
            UID: "ABC127",
            Files: [{
                UID: "123fgb",
                Name: "1980/01/superCuteKitten.jpg",
                Primary: true,
                Type: "jpg",
                Hash: "1xxbgdt55"}],
            Camera: {
                Make: "Canon",
                Model: "abc",
            },
        };
        const photo2 = new Photo(values2);
        assert.equal(photo2.getCamera(), "Canon abc");

    });

    it("should get camera",  () => {
        const values = {ID: 5, Title: "Crazy Cat"};
        const photo = new Photo(values);
        const result = photo.getCamera();
        assert.equal(result, "Unknown");
    });

    it("should get collection resource",  () => {
        const result = Photo.getCollectionResource();
        assert.equal(result, "photos");
    });

    it("should get model name",  () => {
        const result = Photo.getModelName();
        assert.equal(result, "Photo");
    });

    it("should like photo",  () => {
        const values = {ID: 5, Title: "Crazy Cat", CountryName: "Africa", Favorite: false};
        const photo = new Photo(values);
        assert.equal(photo.Favorite, false);
        photo.like();
        assert.equal(photo.Favorite, true);
    });

    it("should unlike photo",  () => {
        const values = {ID: 5, UID: "abc123", Title: "Crazy Cat", CountryName: "Africa", Favorite: true};
        const photo = new Photo(values);
        assert.equal(photo.Favorite, true);
        photo.unlike();
        assert.equal(photo.Favorite, false);
    });

    it("should toggle like",  () => {
        const values = {ID: 5, Title: "Crazy Cat", CountryName: "Africa", Favorite: true};
        const photo = new Photo(values);
        assert.equal(photo.Favorite, true);
        photo.toggleLike();
        assert.equal(photo.Favorite, false);
        photo.toggleLike();
        assert.equal(photo.Favorite, true);
    });

    it("should get photo defaults",  () => {
        const values = {ID: 5, UID: "ABC123"};
        const photo = new Photo(values);
        const result = photo.getDefaults();
        assert.equal(result.UID, "");
    });

    it("should get photos base name",  () => {
        const values = {ID: 8, UID: "ABC123", Filename: "1980/01/superCuteKitten.jpg", FileUID: "123fgb", Files: [{UID: "123fgb", Name: "1980/01/superCuteKitten.jpg", Primary: true, Type: "TypeJpeg", Width: 500, Height: 600, Hash: "1xxbgdt53"}]};
        const photo = new Photo(values);
        const result = photo.baseName();
        assert.equal(result, "superCuteKitten.jpg");
        const result2 = photo.baseName(5);
        assert.equal(result2, "supe…");
    });

    it("should refresh file attributes",  () => {
        const values2 = {ID: 5, UID: "ABC123"};
        const photo2 = new Photo(values2);
        photo2.refreshFileAttr();
        assert.equal(photo2.Width, undefined);
        assert.equal(photo2.Height, undefined);
        assert.equal(photo2.Hash, undefined);
        const values = {ID: 8, UID: "ABC123", Filename: "1980/01/superCuteKitten.jpg", FileUID: "123fgb", Files: [{UID: "123fgb", Name: "1980/01/superCuteKitten.jpg", Primary: true, Type: "TypeJpeg", Width: 500, Height: 600, Hash: "1xxbgdt53"}]};
        const photo = new Photo(values);
        assert.equal(photo.Width, undefined);
        assert.equal(photo.Height, undefined);
        assert.equal(photo.Hash, undefined);
        photo.refreshFileAttr();
        assert.equal(photo.Width, 500);
        assert.equal(photo.Height, 600);
        assert.equal(photo.Hash, "1xxbgdt53");
    });

    it("should return is playable",  () => {
        const values = {ID: 8, UID: "ABC123", Filename: "1980/01/superCuteKitten.jpg", FileUID: "123fgb", Files: [{UID: "123fgb", Name: "1980/01/superCuteKitten.jpg", Primary: true, Type: "TypeJpeg", Width: 500, Height: 600, Hash: "1xxbgdt53"}]};
        const photo = new Photo(values);
        assert.equal(photo.isPlayable(), false);
        const values2 = {ID: 9, UID: "ABC163"};
        const photo2 = new Photo(values2);
        assert.equal(photo2.isPlayable(), false);
        const values3 = {ID: 10, UID: "ABC127", Filename: "1980/01/superCuteKitten.mp4", FileUID: "123fgb", Files: [{UID: "123fgb", Name: "1980/01/superCuteKitten.mp4", Primary: false, Video: true, Type: "mp4", Width: 500, Height: 600, Hash: "1xxbgdt55"}]};
        const photo3 = new Photo(values3);
        assert.equal(photo3.isPlayable(), true);
        const values4 = {ID: 1, UID: "ABC128", Filename: "1980/01/superCuteKitten.jpg", FileUID: "123fgb", Files: [{UID: "123fgb", Name: "1980/01/superCuteKitten.jpg", Primary: false, Video: true, Type: "jpg", Width: 500, Height: 600, Hash: "1xxbgdt53", Codec: "avc1"}]};
        const photo4 = new Photo(values4);
        assert.equal(photo4.isPlayable(), true);
    });

    it("should return videofile",  () => {
        const values = {ID: 8, UID: "ABC123", Filename: "1980/01/superCuteKitten.jpg", FileUID: "123fgb", Files: [{UID: "123fgb", Name: "1980/01/superCuteKitten.jpg", Primary: true, Type: "jpg", Width: 500, Height: 600, Hash: "1xxbgdt53"}]};
        const photo = new Photo(values);
        assert.equal(photo.videoFile(), undefined);
        const values2 = {ID: 9, UID: "ABC163"};
        const photo2 = new Photo(values2);
        assert.equal(photo2.videoFile(), false);
        const values3 = {ID: 10, UID: "ABC127", Filename: "1980/01/superCuteKitten.mp4", FileUID: "123fgb", Files: [{UID: "123fgb", Name: "1980/01/superCuteKitten.mp4", Primary: false, Type: "mp4", Width: 500, Height: 600, Hash: "1xxbgdt55"}]};
        const photo3 = new Photo(values3);
        const file = photo3.videoFile();
        assert.equal(photo3.videoFile().Name, "1980/01/superCuteKitten.mp4");
        const values4 = {ID: 1, UID: "ABC128", Filename: "1980/01/superCuteKitten.jpg", FileUID: "123fgb", Files: [{UID: "123fgb", Name: "1980/01/superCuteKitten.jpg", Primary: false, Type: "jpg", Width: 500, Height: 600, Hash: "1xxbgdt53", Codec: "avc1"}]};
        const photo4 = new Photo(values4);
        assert.equal(photo4.videoFile().Name, "1980/01/superCuteKitten.jpg");
        const file2 = photo4.videoFile();
    });

    it("should return video url",  () => {
        const values = {ID: 8, UID: "ABC123", Filename: "1980/01/superCuteKitten.jpg", Hash: "703cf8f274fbb265d49c6262825780e1", FileUID: "123fgb", Files: [{UID: "123fgb", Name: "1980/01/superCuteKitten.jpg", Primary: true, Type: "jpg", Width: 500, Height: 600, Hash: "1xxbgdt53"}]};
        const photo = new Photo(values);
        assert.equal(photo.videoUrl(), "/api/v1/videos/703cf8f274fbb265d49c6262825780e1/public/avc");
        const values2 = {ID: 9, UID: "ABC163", Hash: "2305e512e3b183ec982d60a8b608a8ca501973ba"};
        const photo2 = new Photo(values2);
        assert.equal(photo2.videoUrl(), "/api/v1/videos/2305e512e3b183ec982d60a8b608a8ca501973ba/public/avc");
        const values3 = {ID: 10, UID: "ABC127", Filename: "1980/01/superCuteKitten.mp4", FileUID: "123fgb", Files: [{UID: "123fgb", Name: "1980/01/superCuteKitten.mp4", Primary: false, Type: "mp4", Width: 500, Height: 600, Hash: "1xxbgdt55"}]};
        const photo3 = new Photo(values3);
        assert.equal(photo3.videoUrl(), "/api/v1/videos/1xxbgdt55/public/avc");
        const values4 = {ID: 1, UID: "ABC128", Filename: "1980/01/superCuteKitten.jpg", FileUID: "123fgb", Files: [{UID: "123fgb", Name: "1980/01/superCuteKitten.jpg", Primary: false, Type: "jpg", Width: 500, Height: 600, Hash: "1xxbgdt53", Codec: "avc1"}]};
        const photo4 = new Photo(values4);
        assert.equal(photo4.videoUrl(), "/api/v1/videos/1xxbgdt53/public/avc");
    });

    it("should return main file",  () => {
        const values = {ID: 9, UID: "ABC163", Width: 111, Height: 222};
        const photo = new Photo(values);
        assert.equal(photo.mainFile(), photo);
        const values2 = {ID: 10,
            UID: "ABC127",
            Files: [
                {UID: "123fgb", Name: "1980/01/superCuteKitten.mp4", Primary: false, Type: "mp4", Width: 500, Height: 600, Hash: "1xxbgdt55"}
                ,{UID: "123fgb", Name: "1980/01/superCuteKitten.jpg", Primary: false, Type: "jpg", Width: 500, Height: 600, Hash: "1xxbgdt56"}]};
        const photo2 = new Photo(values2);
        const file = photo2.mainFile();
        assert.equal(file.Name, "1980/01/superCuteKitten.jpg");
        const values3 = {ID: 1,
            UID: "ABC128",
            Files: [
                {UID: "123fgb", Name: "1980/01/NotMainKitten.jpg", Primary: false, Type: "jpg", Width: 500, Height: 600, Hash: "1xxbgdt53"},
                {UID: "123fgb", Name: "1980/01/MainKitten.jpg", Primary: true, Type: "jpg", Width: 500, Height: 600, Hash: "1xxbgdt54"}
            ]};
        const photo3 = new Photo(values3);
        const file2 = photo3.mainFile();
        assert.equal(file2.Name, "1980/01/MainKitten.jpg");
    });

    it("should return main hash",  () => {
        const values = {ID: 9, UID: "ABC163"};
        const photo = new Photo(values);
        assert.equal(photo.mainFileHash(), "");
        const values2 = {ID: 10,
            UID: "ABC127",
            Files: [
                {UID: "123fgb", Name: "1980/01/superCuteKitten.mp4", Primary: false, Type: "mp4", Width: 500, Height: 600, Hash: "1xxbgdt55"}
                ,{UID: "123fgb", Name: "1980/01/superCuteKitten.jpg", Primary: false, Type: "jpg", Width: 500, Height: 600, Hash: "1xxbgdt56"}]};
        const photo2 = new Photo(values2);
        assert.equal(photo2.mainFileHash(), "1xxbgdt56");
    });

    it("should test filemodels",  () => {
        const values = {ID: 9, UID: "ABC163"};
        const photo = new Photo(values);
        assert.empty(photo.fileModels());
        const values2 = {ID: 10,
            UID: "ABC127",
            Files: [
                {UID: "123fgb", Name: "1980/01/cat.jpg", Primary: false, Type: "jpg", Width: 500, Height: 600, Hash: "1xxbgdt55"}
                ,{UID: "123fgb", Name: "1999/01/dog.jpg", Primary: true, Type: "jpg", Width: 500, Height: 600, Hash: "1xxbgdt56"}]};
        const photo2 = new Photo(values2);
        assert.equal(photo2.fileModels()[0].Name, "1999/01/dog.jpg");
        const values3 = {ID: 10,
            UID: "ABC127",
            Files: [
                {UID: "123fgb", Name: "1980/01/cat.jpg", Primary: true, Type: "jpg", Width: 500, Height: 600, Hash: "1xxbgdt55"}
                ,{UID: "123fgb", Name: "1999/01/dog.jpg", Primary: false, Type: "jpg", Width: 500, Height: 600, Hash: "1xxbgdt56"}]};
        const photo3 = new Photo(values3);
        assert.equal(photo3.fileModels()[0].Name, "1980/01/cat.jpg");
        const values4 = {ID: 10,
            UID: "ABC127",
            Files: [
                {UID: "123fgb", Name: "1980/01/cat.jpg", Primary: true, Type: "jpg", Width: 500, Height: 600, Hash: "1xxbgdt55"}]};
        const photo4 = new Photo(values4);
        assert.equal(photo4.fileModels()[0].Name, "1980/01/cat.jpg");
    });

    it("should get country name",  () => {
        const values = {ID: 5, UID: "ABC123", Country: "zz"};
        const photo = new Photo(values);
        assert.equal(photo.countryName(), "Unknown");
        const values2 = {ID: 5, UID: "ABC123", Country: "es"};
        const photo2 = new Photo(values2);
        assert.equal(photo2.countryName(), "Spain");
    });

    it("should get location info",  () => {
        const values = {ID: 5, UID: "ABC123", Country: "zz", PlaceID: "zz", PlaceLabel: "Nice beach"};
        const photo = new Photo(values);
        assert.equal(photo.locationInfo(), "Nice beach");
        const values2 = {ID: 5, UID: "ABC123", Country: "es", PlaceID: "zz"};
        const photo2 = new Photo(values2);
        assert.equal(photo2.locationInfo(), "Spain");
    });

    it("should return video info",  () => {
        const values = {
            ID: 9,
            UID: "ABC163"};
        const photo = new Photo(values);
        assert.equal(photo.getVideoInfo(), "Video");
        const values2 = {
            ID: 10,
            UID: "ABC127",
            Files: [{
                UID: "123fgb",
                Name: "1980/01/superCuteKitten.mp4",
                Primary: false,
                Type: "mp4",
                Hash: "1xxbgdt55"}]};
        const photo2 = new Photo(values2);
        assert.equal(photo2.getVideoInfo(), "Video");
        const values3 = {
            ID: 10,
            UID: "ABC127",
            Files: [{
                UID: "123fgb",
                Name: "1980/01/superCuteKitten.mp4",
                Primary: false,
                Type: "mp4",
                Width: 500,
                Height: 600,
                Hash: "1xxbgdt55",
                Duration: 6000,
                Size: 222897,
                Codec: "avc1"}]};
        const photo3 = new Photo(values3);
        assert.equal(photo3.getVideoInfo(), "6µs, AVC1, 500 × 600, 0.2 MB");
        const values4 = {
            ID: 10,
            UID: "ABC127",
            Files: [{
                UID: "123fgb",
                Name: "1980/01/superCuteKitten.mp4",
                Primary: false,
                Type: "mp4",
                Hash: "1xxbgdt55",
                Duration: 6000,
                Size: 10240,
                Codec: "avc1"},
                {
                UID: "345fgb",
                Name: "1980/01/superCuteKitten.jpg",
                Primary: true,
                Type: "jpg",
                Hash: "1xxbgjhu5",
                Width: 300,
                Height: 500}]};
        const photo4 = new Photo(values4);
        assert.equal(photo4.getVideoInfo(), "6µs, AVC1, 300 × 500, 10.0 KB");
    });

    it("should return photo info",  () => {
        const values = {
            ID: 9,
            UID: "ABC163"};
        const photo = new Photo(values);
        assert.equal(photo.getPhotoInfo(), "Unknown");
        const values2 = {
            ID: 10,
            UID: "ABC127",
            Files: [{
                UID: "123fgb",
                Name: "1980/01/superCuteKitten.jpg",
                Primary: true,
                Type: "jpg",
                Hash: "1xxbgdt55"}],
                Size: "300",
            Camera: {
                Make: "Canon",
                Model: "abc",
            },
        };
        const photo2 = new Photo(values2);
        assert.equal(photo2.getPhotoInfo(), "Canon abc");
        const values3 = {
            ID: 10,
            UID: "ABC127",
            CameraMake: "Canon",
            CameraModel: "abcde",
            Files: [{
                UID: "123fgb",
                Name: "1980/01/superCuteKitten.mp4",
                Primary: false,
                Type: "mp4",
                Width: 500,
                Height: 600,
                Hash: "1xxbgdt55",
                Duration: 6000,
                Codec: "avc1"}]};
        const photo3 = new Photo(values3);
        assert.equal(photo3.getPhotoInfo(), "Canon abcde, AVC1, 500 × 600");
        const values4 = {
            ID: 10,
            UID: "ABC127",
            Files: [{
                UID: "123fgb",
                Name: "1980/01/superCuteKitten.mp4",
                Primary: false,
                Type: "mp4",
                Width: 500,
                Height: 600,
                Hash: "1xxbgdt55",
                Duration: 6000,
                Size: 300,
                Codec: "avc1"},
                {
                UID: "123fgx",
                Name: "1980/01/superCuteKitten.jpg",
                Primary: true,
                Type: "jpg",
                Width: 800,
                Height: 600,
                Hash: "1xxbgdt55",
                Duration: 6000,
                Size: 200,
                Codec: "avc1"},
            ]};
        const photo4 = new Photo(values4);
        assert.equal(photo3.getPhotoInfo(), "Canon abcde, AVC1, 500 × 600");
    });

    it("should archive photo",  (done) => {
        const values = {ID: 5, Title: "Crazy Cat", CountryName: "Africa", Favorite: false};
        const photo = new Photo(values);
        photo.archive().then(
            (response) => {
                assert.equal(200, response.status);
                assert.deepEqual({ photos: [ 1, 3 ] }, response.data);
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });

    it("should approve photo",  (done) => {
        const values = {ID: 5, UID: "pqbemz8276mhtobh", Title: "Crazy Cat", CountryName: "Africa", Favorite: false};
        const photo = new Photo(values);
        photo.approve().then(
            (response) => {
                assert.equal(200, response.status);
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });

    it("should toggle private",  () => {
        const values = {ID: 5, Title: "Crazy Cat", CountryName: "Africa", Private: true};
        const photo = new Photo(values);
        assert.equal(photo.Private, true);
        photo.togglePrivate();
        assert.equal(photo.Private, false);
        photo.togglePrivate();
        assert.equal(photo.Private, true);
    });

    it("should mark photo as primary",  (done) => {
        const values = {
            ID: 10,
            UID: "pqbemz8276mhtobh",
            Files: [{
                UID: "fqbfk181n4ca5sud",
                Name: "1980/01/superCuteKitten.mp4",
                Primary: false,
                Type: "mp4",
                Hash: "1xxbgdt55"}]};
        const photo = new Photo(values);
        photo.primaryFile("fqbfk181n4ca5sud").then(
            (response) => {
                assert.equal(response.Files[0].Primary, true);
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });

    it("should unstack",  (done) => {
        const values = {
            ID: 10,
            UID: "pqbemz8276mhtobh",
            Files: [{
                UID: "fqbfk181n4ca5sud",
                Name: "1980/01/superCuteKitten.mp4",
                Primary: false,
                Type: "mp4",
                Hash: "1xxbgdt55"}]};
        const photo = new Photo(values);
        photo.unstackFile("fqbfk181n4ca5sud").then(
            (response) => {
                assert.equal(response.success, "ok");
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });

    it("should delete file",  (done) => {
        const values = {
            ID: 10,
            UID: "pqbemz8276mhtobh",
            Files: [{
                UID: "fqbfk181n4ca5sud",
                Name: "1980/01/superCuteKitten.mp4",
                Primary: false,
                Type: "mp4",
                Hash: "1xxbgdt55"},
                {
                UID: "fqbfk181n4ca5abc",
                Name: "1980/01/superCuteKitten.mp4",
                Primary: true,
                Type: "mp4",
                Hash: "1xxbgdt89"}]};
        const photo = new Photo(values);
        photo.deleteFile("fqbfk181n4ca5sud").then(
            (response) => {
                assert.equal(response.success, "successfully deleted");
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });

    it("should add label",  (done) => {
        const values = {
            ID: 10,
            UID: "pqbemz8276mhtobh"};
        const photo = new Photo(values);
        photo.addLabel("Cat").then(
            (response) => {
                assert.equal(response.success, "ok");
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });

    it("should activate label",  (done) => {
        const values = {
            ID: 10,
            UID: "pqbemz8276mhtobh"};
        const photo = new Photo(values);
        photo.activateLabel(12345).then(
            (response) => {
                assert.equal(response.success, "ok");
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });

    it("should rename label",  (done) => {
        const values = {
            ID: 10,
            UID: "pqbemz8276mhtobh"};
        const photo = new Photo(values);
        photo.renameLabel(12345, "Sommer").then(
            (response) => {
                assert.equal(response.success, "ok");
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });

    it("should remove label",  (done) => {
        const values = {
            ID: 10,
            UID: "pqbemz8276mhtobh"};
        const photo = new Photo(values);
        photo.removeLabel(12345).then(
            (response) => {
                assert.equal(response.success, "ok");
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });

    it("should test update",  (done) => {
        const values = {
            ID: 10,
            UID: "pqbemz8276mhtobh",
            Lat: 1.1,
            Lng: 3.3,
            CameraID: 123,
            Title: "Test Titel",
            Description: "Super nice video",
            Files: [{
                UID: "fqbfk181n4ca5sud",
                Name: "1980/01/superCuteKitten.mp4",
                Primary: false,
                Type: "mp4",
                Hash: "1xxbgdt55"}]};
        const photo = new Photo(values);
        photo.update().then(
            (response) => {
                assert.equal(response.TitleSrc, "manual");
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });
});
