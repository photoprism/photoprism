import Photo from "model/photo";
import MockAdapter from "axios-mock-adapter";
import Api from "common/api";

window.__CONFIG__ = {"name":"PhotoPrism","version":"200531-4684f66-Linux-x86_64-DEBUG","copyright":"(c) 2018-2020 PhotoPrism.org \u003chello@photoprism.org\u003e","flags":"public debug experimental settings","siteUrl":"http://localhost:2342/","siteTitle":"PhotoPrism","siteCaption":"Browse your life","siteDescription":"Personal Photo Management powered by Go and Google TensorFlow. Free and open-source.","siteAuthor":"Anonymous","debug":true,"readonly":false,"uploadNSFW":false,"public":true,"experimental":true,"disableSettings":false,"albumCategories":null,"albums":[],"cameras":[{"ID":2,"Slug":"olympus-c2500l","Name":"Olympus C2500L","Make":"Olympus","Model":"C2500L"},{"ID":1,"Slug":"zz","Name":"Unknown","Make":"","Model":"Unknown"}],"lenses":[{"ID":1,"Slug":"zz","Name":"Unknown","Make":"","Model":"Unknown","Type":""}],"countries":[{"ID":"de","Slug":"germany","Name":"Germany"},{"ID":"is","Slug":"iceland","Name":"Iceland"},{"ID":"zz","Slug":"zz","Name":"Unknown"}],"thumbnails":[{"Name":"fit_720","Width":720,"Height":720},{"Name":"fit_2048","Width":2048,"Height":2048},{"Name":"fit_1280","Width":1280,"Height":1024},{"Name":"fit_1920","Width":1920,"Height":1200},{"Name":"fit_2560","Width":2560,"Height":1600},{"Name":"fit_3840","Width":3840,"Height":2400}],"downloadToken":"1uhovi0e","previewToken":"static","jsHash":"0fd34136","cssHash":"2b327230","settings":{"theme":"default","language":"en","templates":{"default":"index.tmpl"},"maps":{"animate":0,"style":"streets"},"features":{"archive":true,"private":true,"review":true,"upload":true,"import":true,"files":true,"moments":true,"labels":true,"places":true,"download":true,"edit":true,"share":true,"logs":true},"import":{"path":"/","move":false},"index":{"path":"/","convert":true,"rescan":false,"group":true}},"count":{"cameras":1,"lenses":0,"countries":2,"photos":126,"videos":0,"hidden":3,"favorites":1,"private":0,"review":0,"stories":0,"albums":0,"moments":0,"months":0,"folders":0,"files":255,"places":0,"labels":13,"labelMaxPhotos":1},"pos":{"uid":"","loc":"","utc":"0001-01-01T00:00:00Z","lat":0,"lng":0},"years":[2003,2002],"colors":[{"Example":"#AB47BC","Name":"Purple","Slug":"purple"},{"Example":"#FF00FF","Name":"Magenta","Slug":"magenta"},{"Example":"#EC407A","Name":"Pink","Slug":"pink"},{"Example":"#EF5350","Name":"Red","Slug":"red"},{"Example":"#FFA726","Name":"Orange","Slug":"orange"},{"Example":"#D4AF37","Name":"Gold","Slug":"gold"},{"Example":"#FDD835","Name":"Yellow","Slug":"yellow"},{"Example":"#CDDC39","Name":"Lime","Slug":"lime"},{"Example":"#66BB6A","Name":"Green","Slug":"green"},{"Example":"#009688","Name":"Teal","Slug":"teal"},{"Example":"#00BCD4","Name":"Cyan","Slug":"cyan"},{"Example":"#2196F3","Name":"Blue","Slug":"blue"},{"Example":"#A1887F","Name":"Brown","Slug":"brown"},{"Example":"#F5F5F5","Name":"White","Slug":"white"},{"Example":"#9E9E9E","Name":"Grey","Slug":"grey"},{"Example":"#212121","Name":"Black","Slug":"black"}],"categories":[{"UID":"lqb6y631re96cper","Slug":"animal","Name":"Animal"},{"UID":"lqb6y5gvo9avdfx5","Slug":"architecture","Name":"Architecture"},{"UID":"lqb6y633nhfj1uzt","Slug":"bird","Name":"Bird"},{"UID":"lqb6y633g3hxg1aq","Slug":"farm","Name":"Farm"},{"UID":"lqb6y4i1ez9cw5bi","Slug":"nature","Name":"Nature"},{"UID":"lqb6y4f2v7dw8irs","Slug":"plant","Name":"Plant"},{"UID":"lqb6y6s2ohhmu0fn","Slug":"reptile","Name":"Reptile"},{"UID":"lqb6y6ctgsq2g2np","Slug":"water","Name":"Water"}],"clip":160,"server":{"cores":2,"routines":23,"memory":{"used":1224531272,"reserved":1416904088,"info":"Used 1.2 GB / Reserved 1.4 GB"}}};

let chai = require('../../../node_modules/chai/chai');
let assert = chai.assert;

const mock = new MockAdapter(Api);

mock
    .onPost().reply(200)
    .onDelete().reply(200);

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
        assert.equal(result, "/api/v1/t/345982/static/tile500");
        const values2 = {ID: 10, UID: "ABC127",
            Files: [{UID: "123fgb", Name: "1980/01/superCuteKitten.mp4", Primary: false, Type: "mp4", Width: 500, Height: 600, Hash: "1xxbgdt55"}]};
        const photo2 = new Photo(values2);
        const result2 = photo2.thumbnailUrl("tile500");
        assert.equal(result2, "/api/v1/t/1xxbgdt55/static/tile500");
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
        assert.equal(result, "/api/v1/dl/345982?t=1uhovi0e");
    });

    it("should get photo thumbnail src set",  () => {
        const values = {ID: 5, Title: "Crazy Cat", Hash: 345982};
        const photo = new Photo(values);
        const result = photo.thumbnailSrcset();
        assert.equal(result, "/api/v1/t/345982/static/fit_720 720w, /api/v1/t/345982/static/fit_1280 1280w, /api/v1/t/345982/static/fit_1920 1920w, /api/v1/t/345982/static/fit_2560 2560w, /api/v1/t/345982/static/fit_3840 3840w");
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

    it("should get thumbnail sizes",  () => {
        const values = {ID: 5, Title: "Crazy Cat", Width: 500, Height: 200};
        const photo = new Photo(values);
        const result = photo.thumbnailSizes();
        assert.equal(result, "(min-width: 2560px) 3840px, (min-width: 1920px) 2560px, (min-width: 1280px) 1920px, (min-width: 720px) 1280px, 720px");
    });

    it("should get date string",  () => {
        const values = {ID: 5, Title: "Crazy Cat", TakenAt: "2012-07-08T14:45:39Z", TimeZone: "UTC"};
        const photo = new Photo(values);
        const result = photo.getDateString();
        assert.equal(result, "July 8, 2012, 2:45 PM UTC");
        const values2 = {ID: 5, Title: "Crazy Cat", TakenAt: "", TimeZone: "UTC"};
        const photo2 = new Photo(values2);
        const result2 = photo2.getDateString();
        assert.equal(result2, "Unknown");
        const values3 = {ID: 5, Title: "Crazy Cat", TakenAt: "2012-07-08T14:45:39Z"};
        const photo3 = new Photo(values3);
        const result3 = photo3.getDateString();
        assert.equal(result3, "Sunday, July 8, 2012");
    });

    it("should get short date string",  () => {
        const values = {ID: 5, Title: "Crazy Cat", TakenAt: "2012-07-08T14:45:39Z", TimeZone: "UTC"};
        const photo = new Photo(values);
        const result = photo.shortDateString();
        assert.equal(result, "Jul 8, 2012");
        const values2 = {ID: 5, Title: "Crazy Cat", TakenAt: "", TimeZone: "UTC"};
        const photo2 = new Photo(values2);
        const result2 = photo2.shortDateString();
        assert.equal(result2, "Unknown");
        const values3 = {ID: 5, Title: "Crazy Cat", TakenAt: "2012-07-08T14:45:39Z"};
        const photo3 = new Photo(values3);
        const result3 = photo3.shortDateString();
        assert.equal(result3, "Jul 8, 2012");
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
        const values = {ID: 5, Title: "Crazy Cat", LocUID: 6, LocType: "viewpoint", LocLabel: "Cape Point, South Africa", LocCountry: "South Africa"};
        const photo = new Photo(values);
        const result = photo.locationInfo();
        assert.equal(result, "Cape Point, South Africa");
    });

    it("should get location",  () => {
        const values = {ID: 5, Title: "Crazy Cat", LocUID: 6, LocType: "viewpoint", LocLabel: "Cape Point, State, South Africa", LocCountry: "South Africa", LocCity: "Cape Town", LocCounty: "County", LocState: "State"};
        const photo = new Photo(values);
        const result = photo.locationInfo();
        assert.equal(result, "Cape Point, State, South Africa");
    });

    it("should get location",  () => {
        const values = {ID: 5, Title: "Crazy Cat", LocType: "viewpoint", LocName: "Cape Point", LocCountry: "Africa", LocCity: "Cape Town", LocCounty: "County", LocState: "State"};
        const photo = new Photo(values);
        const result = photo.locationInfo();
        assert.equal(result, "Unknown");
    });

    it("should get location",  () => {
        const values = {ID: 5, Title: "Crazy Cat", CountryName: "Africa", LocCity: "Cape Town"};
        const photo = new Photo(values);
        const result = photo.locationInfo();
        assert.equal(result, "Unknown");
    });

    it("should get camera",  () => {
        const values = {ID: 5, Title: "Crazy Cat", CameraModel: "EOSD10", CameraMake: "Canon"};
        const photo = new Photo(values);
        const result = photo.getCamera();
        assert.equal(result, "Canon EOSD10");
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
        const values = {ID: 5, Title: "Crazy Cat", CountryName: "Africa", Favorite: true};
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
        assert.equal(result2, "su...");
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
        const values3 = {ID: 10, UID: "ABC127", Filename: "1980/01/superCuteKitten.mp4", FileUID: "123fgb", Files: [{UID: "123fgb", Name: "1980/01/superCuteKitten.mp4", Primary: false, Type: "mp4", Width: 500, Height: 600, Hash: "1xxbgdt55"}]};
        const photo3 = new Photo(values3);
        assert.equal(photo3.isPlayable(), true);
        const values4 = {ID: 1, UID: "ABC128", Filename: "1980/01/superCuteKitten.jpg", FileUID: "123fgb", Files: [{UID: "123fgb", Name: "1980/01/superCuteKitten.jpg", Primary: false, Type: "jpg", Width: 500, Height: 600, Hash: "1xxbgdt53", Codec: "avc1"}]};
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
        console.log(file);
        assert.equal(photo3.videoFile().Name, "1980/01/superCuteKitten.mp4");
        const values4 = {ID: 1, UID: "ABC128", Filename: "1980/01/superCuteKitten.jpg", FileUID: "123fgb", Files: [{UID: "123fgb", Name: "1980/01/superCuteKitten.jpg", Primary: false, Type: "jpg", Width: 500, Height: 600, Hash: "1xxbgdt53", Codec: "avc1"}]};
        const photo4 = new Photo(values4);
        assert.equal(photo4.videoFile().Name, "1980/01/superCuteKitten.jpg");
        const file2 = photo4.videoFile();
        console.log(file2);
    });

    it("should return video url",  () => {
        const values = {ID: 8, UID: "ABC123", Filename: "1980/01/superCuteKitten.jpg", FileUID: "123fgb", Files: [{UID: "123fgb", Name: "1980/01/superCuteKitten.jpg", Primary: true, Type: "jpg", Width: 500, Height: 600, Hash: "1xxbgdt53"}]};
        const photo = new Photo(values);
        assert.equal(photo.videoUrl(), "");
        const values2 = {ID: 9, UID: "ABC163"};
        const photo2 = new Photo(values2);
        assert.equal(photo2.videoUrl(), false);
        const values3 = {ID: 10, UID: "ABC127", Filename: "1980/01/superCuteKitten.mp4", FileUID: "123fgb", Files: [{UID: "123fgb", Name: "1980/01/superCuteKitten.mp4", Primary: false, Type: "mp4", Width: 500, Height: 600, Hash: "1xxbgdt55"}]};
        const photo3 = new Photo(values3);
        assert.equal(photo3.videoUrl(), "/api/v1/videos/1xxbgdt55/static/mp4");
        const values4 = {ID: 1, UID: "ABC128", Filename: "1980/01/superCuteKitten.jpg", FileUID: "123fgb", Files: [{UID: "123fgb", Name: "1980/01/superCuteKitten.jpg", Primary: false, Type: "jpg", Width: 500, Height: 600, Hash: "1xxbgdt53", Codec: "avc1"}]};
        const photo4 = new Photo(values4);
        assert.equal(photo4.videoUrl(), "/api/v1/videos/1xxbgdt53/static/mp4");
    });

    it("should return main file",  () => {
        const values = {ID: 9, UID: "ABC163"};
        const photo = new Photo(values);
        assert.equal(photo.mainFile(), false);
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
    });
});
