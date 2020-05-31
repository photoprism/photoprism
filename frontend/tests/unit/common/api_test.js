import Api from "common/api";
import MockAdapter from "axios-mock-adapter";

window.__CONFIG__ = {"name":"PhotoPrism","version":"200531-d019959-Linux-x86_64-DEBUG","copyright":"(c) 2018-2020 PhotoPrism.org \u003chello@photoprism.org\u003e","flags":"public debug experimental settings","siteUrl":"http://localhost:2342/","siteTitle":"PhotoPrism","siteCaption":"Browse your life","siteDescription":"Personal Photo Management powered by Go and Google TensorFlow. Free and open-source.","siteAuthor":"Anonymous","debug":true,"readonly":false,"uploadNSFW":false,"public":true,"experimental":true,"disableSettings":false,"albumCategories":null,"albums":[],"cameras":[{"ID":1,"Slug":"zz","Name":"Unknown","Make":"","Model":"Unknown"}],"lenses":[{"ID":1,"Slug":"zz","Name":"Unknown","Make":"","Model":"Unknown","Type":""}],"countries":[{"ID":"zz","Slug":"zz","Name":"Unknown"}],"thumbnails":[{"Name":"fit_1920","Width":1920,"Height":1200},{"Name":"fit_2048","Width":2048,"Height":2048},{"Name":"fit_720","Width":720,"Height":720},{"Name":"fit_1280","Width":1280,"Height":1024},{"Name":"fit_2560","Width":2560,"Height":1600},{"Name":"fit_3840","Width":3840,"Height":2400}],"downloadToken":"21kaatus","previewToken":"static","jsHash":"c5acefae","cssHash":"2b327230","settings":{"theme":"default","language":"en","templates":{"default":"index.tmpl"},"maps":{"animate":0,"style":"streets"},"features":{"archive":true,"private":true,"review":true,"upload":true,"import":true,"files":true,"moments":true,"labels":true,"places":true,"download":true,"edit":true,"share":true,"logs":true},"import":{"path":"/","move":false},"index":{"path":"/","convert":true,"rescan":false,"group":true}},"count":{"cameras":0,"lenses":0,"countries":0,"photos":0,"videos":0,"hidden":0,"favorites":0,"private":0,"review":0,"stories":0,"albums":0,"moments":0,"months":0,"folders":0,"files":0,"places":0,"labels":0,"labelMaxPhotos":0},"pos":{"uid":"","loc":"","utc":"0001-01-01T00:00:00Z","lat":0,"lng":0},"years":null,"colors":[{"Example":"#AB47BC","Name":"Purple","Slug":"purple"},{"Example":"#FF00FF","Name":"Magenta","Slug":"magenta"},{"Example":"#EC407A","Name":"Pink","Slug":"pink"},{"Example":"#EF5350","Name":"Red","Slug":"red"},{"Example":"#FFA726","Name":"Orange","Slug":"orange"},{"Example":"#D4AF37","Name":"Gold","Slug":"gold"},{"Example":"#FDD835","Name":"Yellow","Slug":"yellow"},{"Example":"#CDDC39","Name":"Lime","Slug":"lime"},{"Example":"#66BB6A","Name":"Green","Slug":"green"},{"Example":"#009688","Name":"Teal","Slug":"teal"},{"Example":"#00BCD4","Name":"Cyan","Slug":"cyan"},{"Example":"#2196F3","Name":"Blue","Slug":"blue"},{"Example":"#A1887F","Name":"Brown","Slug":"brown"},{"Example":"#F5F5F5","Name":"White","Slug":"white"},{"Example":"#9E9E9E","Name":"Grey","Slug":"grey"},{"Example":"#212121","Name":"Black","Slug":"black"}],"categories":[],"clip":160,"server":{"cores":2,"routines":14,"memory":{"used":355636448,"reserved":490369328,"info":"Used 356 MB / Reserved 490 MB"}}};

let chai = require("../../../node_modules/chai/chai");
let assert = chai.assert;

describe("common/api", () => {

    const mock = new MockAdapter(Api);

    const getCollectionResponse = [
        {id: 1, name: "John Smith"},
        {id: 1, name: "John Smith"}
    ];

    const getEntityResponse = {
        id: 1, name: "John Smith"
    };

    const postEntityResponse = {
        users: [
            {id: 1, name: "John Smith"}
        ]
    };

    const putEntityResponse = {
        users: [
            {id: 2, name: "John Foo"}
        ]
    };

    const deleteEntityResponse = null;

    mock.onGet("foo").reply(200, getCollectionResponse);
    mock.onGet("foo/123").reply(200, getEntityResponse);
    mock.onPost("foo").reply(201, postEntityResponse);
    mock.onPut("foo/2").reply(200, putEntityResponse);
    mock.onDelete("foo/2").reply(204, deleteEntityResponse);
    mock.onGet("error").reply(401, "custom error cat");

    it("get should return a list of results and return with HTTP code 200", (done) => {
        Api.get("foo").then(
            (response) => {
                assert.equal(200, response.status);
                assert.deepEqual(getCollectionResponse, response.data);
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });

    it("get should return one item and return with HTTP code 200", (done) => {
        Api.get("foo/123").then(
            (response) => {
                assert.equal(200, response.status);
                assert.deepEqual(getEntityResponse, response.data);
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });

    it("post should create one item and return with HTTP code 201", (done) => {
        Api.post("foo", postEntityResponse).then(
            (response) => {
                assert.equal(201, response.status);
                assert.deepEqual(postEntityResponse, response.data);
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });

    it("put should update one item and return with HTTP code 200", (done) => {
        Api.put("foo/2", putEntityResponse).then(
            (response) => {
                assert.equal(200, response.status);
                assert.deepEqual(putEntityResponse, response.data);
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });

    it("delete should delete one item and return with HTTP code 204", (done) => {
        Api.delete("foo/2", deleteEntityResponse).then(
            (response) => {
                assert.equal(204, response.status);
                assert.deepEqual(deleteEntityResponse, response.data);
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });

    it("get error", function() {
        return Api.get("error")
            .then(function(m) { throw new Error("was not supposed to succeed"); })
            .catch(function(m) { assert.equal(m.message, "Request failed with status code 401")});
    });
});
