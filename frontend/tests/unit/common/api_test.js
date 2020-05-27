import Api from "common/api";
import MockAdapter from "axios-mock-adapter";

window.__CONFIG__ = {"flags":"public debug experimental settings","name":"PhotoPrism","url":"http://localhost:2342/","title":"PhotoPrism","subtitle":"Browse your life","description":"Personal Photo Management","author":"PhotoPrism.org","version":"200527-5453cf2-Linux-x86_64-DEBUG","copyright":"(c) 2018-2020 PhotoPrism.org \u003chello@photoprism.org\u003e","debug":true,"readonly":false,"uploadNSFW":false,"public":true,"experimental":true,"disableSettings":false,"albums":[],"cameras":[{"ID":30003,"Slug":"apple-iphone-6","Model":"iPhone 6","Make":"Apple"},{"ID":30001,"Slug":"apple-iphone-se","Model":"iPhone SE","Make":"Apple"},{"ID":30004,"Slug":"canon-eos-6d","Model":"EOS 6D","Make":"Canon"},{"ID":30002,"Slug":"canon-eos-m6","Model":"EOS M6","Make":"Canon"},{"ID":30006,"Slug":"huawei-ele-l29","Model":"ELE-L29","Make":"HUAWEI"},{"ID":30005,"Slug":"motorola-moto-g-4","Model":"Moto G (4)","Make":"Motorola"},{"ID":1,"Slug":"zz","Model":"Unknown","Make":""}],"lenses":[{"ID":30003,"Slug":"22-0-mm","Model":"22.0 mm","Make":"","Type":""},{"ID":30005,"Slug":"ef16-35mm-f-2-8l-ii-usm","Model":"EF16-35mm f/2.8L II USM","Make":"","Type":""},{"ID":30004,"Slug":"iphone-6-back-camera-4-15mm-f-2-2","Model":"iPhone 6 back camera 4.15mm f/2.2","Make":"Apple","Type":""},{"ID":30001,"Slug":"iphone-se-back-camera-4-15mm-f-2-2","Model":"iPhone SE back camera 4.15mm f/2.2","Make":"Apple","Type":""},{"ID":30002,"Slug":"iphone-se-front-camera-2-15mm-f-2-4","Model":"iPhone SE front camera 2.15mm f/2.4","Make":"Apple","Type":""},{"ID":1,"Slug":"zz","Model":"Unknown","Make":"","Type":""}],"countries":[{"ID":"at","Slug":"austria","Name":"Austria"},{"ID":"bw","Slug":"botswana","Name":"Botswana"},{"ID":"ca","Slug":"canada","Name":"Canada"},{"ID":"fr","Slug":"france","Name":"France"},{"ID":"de","Slug":"germany","Name":"Germany"},{"ID":"gr","Slug":"greece","Name":"Greece"},{"ID":"za","Slug":"south-africa","Name":"South Africa"},{"ID":"us","Slug":"usa","Name":"USA"},{"ID":"zz","Slug":"zz","Name":"Unknown"}],"thumbnails":[{"Name":"fit_720","Width":720,"Height":720},{"Name":"fit_1280","Width":1280,"Height":1024},{"Name":"fit_2560","Width":2560,"Height":1600},{"Name":"fit_3840","Width":3840,"Height":2400},{"Name":"fit_1920","Width":1920,"Height":1200},{"Name":"fit_2048","Width":2048,"Height":2048}],"downloadToken":"2y71e0sr","thumbToken":"static","jsHash":"14ba2de4","cssHash":"2b327230","settings":{"theme":"default","language":"en","templates":{"default":"index.tmpl"},"maps":{"animate":0,"style":"streets"},"features":{"archive":true,"private":true,"review":true,"upload":true,"import":true,"files":true,"moments":true,"labels":true,"places":true,"download":true,"edit":true,"share":true,"logs":true},"import":{"path":"/","move":false},"index":{"path":"/","convert":true,"rescan":false,"group":true}},"count":{"photos":385,"videos":1,"hidden":0,"favorites":1,"private":2,"review":4,"stories":0,"albums":0,"folders":14,"files":394,"moments":0,"countries":8,"places":0,"labels":46,"labelMaxPhotos":54},"pos":{"uid":"pqazcltc1x8d12lo","loc":"4777dc437584","utc":"2020-02-14T12:44:19Z","lat":47.207123,"lng":11.823489},"years":[2020,2019,2018,2017,2016],"colors":[{"Example":"#AB47BC","Name":"Purple","Slug":"purple"},{"Example":"#FF00FF","Name":"Magenta","Slug":"magenta"},{"Example":"#EC407A","Name":"Pink","Slug":"pink"},{"Example":"#EF5350","Name":"Red","Slug":"red"},{"Example":"#FFA726","Name":"Orange","Slug":"orange"},{"Example":"#D4AF37","Name":"Gold","Slug":"gold"},{"Example":"#FDD835","Name":"Yellow","Slug":"yellow"},{"Example":"#CDDC39","Name":"Lime","Slug":"lime"},{"Example":"#66BB6A","Name":"Green","Slug":"green"},{"Example":"#009688","Name":"Teal","Slug":"teal"},{"Example":"#00BCD4","Name":"Cyan","Slug":"cyan"},{"Example":"#2196F3","Name":"Blue","Slug":"blue"},{"Example":"#A1887F","Name":"Brown","Slug":"brown"},{"Example":"#F5F5F5","Name":"White","Slug":"white"},{"Example":"#9E9E9E","Name":"Grey","Slug":"grey"},{"Example":"#212121","Name":"Black","Slug":"black"}],"categories":[{"UID":"lqazz283gqjo05j9","Slug":"aircraft","Name":"Aircraft"},{"UID":"lqazyyc2xos6k0op","Slug":"airport","Name":"Airport"},{"UID":"lqazyya2wbw3045h","Slug":"animal","Name":"Animal"},{"UID":"lqazyz22d0y1ham3","Slug":"architecture","Name":"Architecture"},{"UID":"lqazz7537y79uhef","Slug":"beach","Name":"Beach"},{"UID":"lqazz1v2gbroth1y","Slug":"beverage","Name":"Beverage"},{"UID":"lqazyyf213ls8byk","Slug":"building","Name":"Building"},{"UID":"lqazyzhjuf6ud8pd","Slug":"car","Name":"Car"},{"UID":"lqazzcnzc5ejq4xx","Slug":"dining","Name":"Dining"},{"UID":"lqazz1v3t6kuuid7","Slug":"drinks","Name":"Drinks"},{"UID":"lqazz4j3rrxrh9el","Slug":"event","Name":"Event"},{"UID":"lqazz422nbeeedv5","Slug":"farm","Name":"Farm"},{"UID":"lqazz5f3leym5l14","Slug":"food","Name":"Food"},{"UID":"lqazz252fhe2ibx1","Slug":"landscape","Name":"Landscape"},{"UID":"lqazyzs20lrgueeb","Slug":"nature","Name":"Nature"},{"UID":"lqazyyh3f6phuq04","Slug":"outdoor","Name":"Outdoor"},{"UID":"lqazzef1n086vptr","Slug":"people","Name":"People"},{"UID":"lqazzm019mojdmcp","Slug":"plant","Name":"Plant"},{"UID":"lqazyy72zbezq9zt","Slug":"portrait","Name":"Portrait"},{"UID":"lqazzbqrfzbq0l9l","Slug":"sand","Name":"Sand"},{"UID":"lqazz8q1q39ksabl","Slug":"shop","Name":"Shop"},{"UID":"lqazyyt1ua4i8jpy","Slug":"train","Name":"Train"},{"UID":"lqazzdxvp4d59tx4","Slug":"vegetables","Name":"Vegetables"},{"UID":"lqazyytyvhnb6fvk","Slug":"vehicle","Name":"Vehicle"},{"UID":"lqazz4b7rukovpmc","Slug":"water","Name":"Water"},{"UID":"lqazz2t3n882fps6","Slug":"wildlife","Name":"Wildlife"}],"clip":160,"server":{"cores":2,"routines":38,"memory":{"used":56042952,"reserved":144132360,"info":"Used 56 MB / Reserved 144 MB"}}};

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
