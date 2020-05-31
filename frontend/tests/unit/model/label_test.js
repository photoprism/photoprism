import Label from "model/label";
import MockAdapter from "axios-mock-adapter";
import Api from "common/api";

window.__CONFIG__ = {"name":"PhotoPrism","version":"200531-d019959-Linux-x86_64-DEBUG","copyright":"(c) 2018-2020 PhotoPrism.org \u003chello@photoprism.org\u003e","flags":"public debug experimental settings","siteUrl":"http://localhost:2342/","siteTitle":"PhotoPrism","siteCaption":"Browse your life","siteDescription":"Personal Photo Management powered by Go and Google TensorFlow. Free and open-source.","siteAuthor":"Anonymous","debug":true,"readonly":false,"uploadNSFW":false,"public":true,"experimental":true,"disableSettings":false,"albumCategories":null,"albums":[],"cameras":[{"ID":1,"Slug":"zz","Name":"Unknown","Make":"","Model":"Unknown"}],"lenses":[{"ID":1,"Slug":"zz","Name":"Unknown","Make":"","Model":"Unknown","Type":""}],"countries":[{"ID":"zz","Slug":"zz","Name":"Unknown"}],"thumbnails":[{"Name":"fit_1920","Width":1920,"Height":1200},{"Name":"fit_2048","Width":2048,"Height":2048},{"Name":"fit_720","Width":720,"Height":720},{"Name":"fit_1280","Width":1280,"Height":1024},{"Name":"fit_2560","Width":2560,"Height":1600},{"Name":"fit_3840","Width":3840,"Height":2400}],"downloadToken":"21kaatus","previewToken":"static","jsHash":"c5acefae","cssHash":"2b327230","settings":{"theme":"default","language":"en","templates":{"default":"index.tmpl"},"maps":{"animate":0,"style":"streets"},"features":{"archive":true,"private":true,"review":true,"upload":true,"import":true,"files":true,"moments":true,"labels":true,"places":true,"download":true,"edit":true,"share":true,"logs":true},"import":{"path":"/","move":false},"index":{"path":"/","convert":true,"rescan":false,"group":true}},"count":{"cameras":0,"lenses":0,"countries":0,"photos":0,"videos":0,"hidden":0,"favorites":0,"private":0,"review":0,"stories":0,"albums":0,"moments":0,"months":0,"folders":0,"files":0,"places":0,"labels":0,"labelMaxPhotos":0},"pos":{"uid":"","loc":"","utc":"0001-01-01T00:00:00Z","lat":0,"lng":0},"years":null,"colors":[{"Example":"#AB47BC","Name":"Purple","Slug":"purple"},{"Example":"#FF00FF","Name":"Magenta","Slug":"magenta"},{"Example":"#EC407A","Name":"Pink","Slug":"pink"},{"Example":"#EF5350","Name":"Red","Slug":"red"},{"Example":"#FFA726","Name":"Orange","Slug":"orange"},{"Example":"#D4AF37","Name":"Gold","Slug":"gold"},{"Example":"#FDD835","Name":"Yellow","Slug":"yellow"},{"Example":"#CDDC39","Name":"Lime","Slug":"lime"},{"Example":"#66BB6A","Name":"Green","Slug":"green"},{"Example":"#009688","Name":"Teal","Slug":"teal"},{"Example":"#00BCD4","Name":"Cyan","Slug":"cyan"},{"Example":"#2196F3","Name":"Blue","Slug":"blue"},{"Example":"#A1887F","Name":"Brown","Slug":"brown"},{"Example":"#F5F5F5","Name":"White","Slug":"white"},{"Example":"#9E9E9E","Name":"Grey","Slug":"grey"},{"Example":"#212121","Name":"Black","Slug":"black"}],"categories":[],"clip":160,"server":{"cores":2,"routines":14,"memory":{"used":355636448,"reserved":490369328,"info":"Used 356 MB / Reserved 490 MB"}}};

let chai = require('../../../node_modules/chai/chai');
let assert = chai.assert;

const mock = new MockAdapter(Api);

mock
    .onPost().reply(200)
    .onDelete().reply(200);

describe("model/label", () => {
    it("should get label entity name",  () => {
        const values = {ID: 5, UID: "ABC123", Name: "Black Cat", Slug: "black-cat"};
        const label = new Label(values);
        const result = label.getEntityName();
        assert.equal(result, "black-cat");
    });

    it("should get label id",  () => {
        const values = {ID: 5, UID: "ABC123", Name: "Black Cat", Slug: "black-cat"};
        const label = new Label(values);
        const result = label.getId();
        assert.equal(result, "ABC123");
    });

    it("should get label title",  () => {
        const values = {ID: 5, UID: "ABC123", Name: "Black Cat", Slug: "black-cat"};
        const label = new Label(values);
        const result = label.getTitle();
        assert.equal(result, "Black Cat");
    });

    it("should get thumbnail url",  () => {
        const values = {ID: 5, UID: "ABC123", Name: "Black Cat", Slug: "black-cat"};
        const label = new Label(values);
        const result = label.thumbnailUrl("xyz");
        assert.equal(result, "/api/v1/labels/ABC123/t/static/xyz");
    });

    it("should get thumbnail src set",  () => {
        const values = {ID: 5, UID: "ABC123", Name: "Black Cat", Slug: "black-cat"};
        const label = new Label(values);
        const result = label.thumbnailSrcset("");
        assert.equal(result, "/api/v1/labels/ABC123/t/static/fit_720 720w, /api/v1/labels/ABC123/t/static/fit_1280 1280w, /api/v1/labels/ABC123/t/static/fit_1920 1920w, /api/v1/labels/ABC123/t/static/fit_2560 2560w, /api/v1/labels/ABC123/t/static/fit_3840 3840w");
    });

    it("should get thumbnail sizes",  () => {
        const values = {ID: 5, UID: "ABC123", Name: "Black Cat", Slug: "black-cat"};
        const label = new Label(values);
        const result = label.thumbnailSizes();
        assert.equal(result, "(min-width: 2560px) 3840px, (min-width: 1920px) 2560px, (min-width: 1280px) 1920px, (min-width: 720px) 1280px, 720px");
    });

    it("should get date string",  () => {
        const values = {ID: 5, UID: "ABC123", Name: "Black Cat", Slug: "black-cat", CreatedAt: "2012-07-08T14:45:39Z"};
        const label = new Label(values);
        const result = label.getDateString();
        assert.equal(result, "Jul 8, 2012, 2:45 PM");
    });

    it("should get model name",  () => {
        const result = Label.getModelName();
        assert.equal(result, "Label");
    });

    it("should get collection resource",  () => {
        const result = Label.getCollectionResource();
        assert.equal(result, "labels");
    });

    it("should like label",  () => {
        const values = {ID: 5, UID: "ABC123", Name: "Black Cat", Slug: "black-cat", Favorite: false};
        const label = new Label(values);
        assert.equal(label.Favorite, false);
        label.like();
        assert.equal(label.Favorite, true);
    });

    it("should unlike label",  () => {
        const values = {ID: 5, UID: "ABC123",Name: "Black Cat", Slug: "black-cat", Favorite: true};
        const label = new Label(values);
        assert.equal(label.Favorite, true);
        label.unlike();
        assert.equal(label.Favorite, false);
    });

    it("should toggle like",  () => {
        const values = {ID: 5, UID: "ABC123", Name: "Black Cat", Slug: "black-cat", Favorite: true};
        const label = new Label(values);
        assert.equal(label.Favorite, true);
        label.toggleLike();
        assert.equal(label.Favorite, false);
        label.toggleLike();
        assert.equal(label.Favorite, true);
    });
});
