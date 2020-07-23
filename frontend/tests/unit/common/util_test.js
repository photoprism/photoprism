import Util from "common/util";
import MockAdapter from "axios-mock-adapter";

window.__CONFIG__ = {"name":"PhotoPrism","version":"200531-4684f66-Linux-x86_64-DEBUG","copyright":"(c) 2018-2020 PhotoPrism.org \u003chello@photoprism.org\u003e","flags":"public debug experimental settings","siteUrl":"http://localhost:2342/","siteTitle":"PhotoPrism","siteCaption":"Browse your life","siteDescription":"Personal Photo Management powered by Go and Google TensorFlow. Free and open-source.","siteAuthor":"Anonymous","debug":true,"readonly":false,"uploadNSFW":false,"public":true,"experimental":true,"disableSettings":false,"albumCategories":null,"albums":[],"cameras":[{"ID":2,"Slug":"olympus-c2500l","Name":"Olympus C2500L","Make":"Olympus","Model":"C2500L"},{"ID":1,"Slug":"zz","Name":"Unknown","Make":"","Model":"Unknown"}],"lenses":[{"ID":1,"Slug":"zz","Name":"Unknown","Make":"","Model":"Unknown","Type":""}],"countries":[{"ID":"de","Slug":"germany","Name":"Germany"},{"ID":"is","Slug":"iceland","Name":"Iceland"},{"ID":"zz","Slug":"zz","Name":"Unknown"}],"thumbs":[{"Name":"fit_720","Width":720,"Height":720},{"Name":"fit_2048","Width":2048,"Height":2048},{"Name":"fit_1280","Width":1280,"Height":1024},{"Name":"fit_1920","Width":1920,"Height":1200},{"Name":"fit_2560","Width":2560,"Height":1600},{"Name":"fit_3840","Width":3840,"Height":2400}],"downloadToken":"1uhovi0e","previewToken":"static","jsHash":"0fd34136","cssHash":"2b327230","settings":{"theme":"default","language":"en","templates":{"default":"index.tmpl"},"maps":{"animate":0,"style":"streets"},"features":{"archive":true,"private":true,"review":true,"upload":true,"import":true,"files":true,"moments":true,"labels":true,"places":true,"download":true,"edit":true,"share":true,"logs":true},"import":{"path":"/","move":false},"index":{"path":"/","convert":true,"rescan":false,"group":true}},"count":{"cameras":1,"lenses":0,"countries":2,"photos":126,"videos":0,"hidden":3,"favorites":1,"private":0,"review":0,"stories":0,"albums":0,"moments":0,"months":0,"folders":0,"files":255,"places":0,"labels":13,"labelMaxPhotos":1},"pos":{"uid":"","loc":"","utc":"0001-01-01T00:00:00Z","lat":0,"lng":0},"years":[2003,2002],"colors":[{"Example":"#AB47BC","Name":"Purple","Slug":"purple"},{"Example":"#FF00FF","Name":"Magenta","Slug":"magenta"},{"Example":"#EC407A","Name":"Pink","Slug":"pink"},{"Example":"#EF5350","Name":"Red","Slug":"red"},{"Example":"#FFA726","Name":"Orange","Slug":"orange"},{"Example":"#D4AF37","Name":"Gold","Slug":"gold"},{"Example":"#FDD835","Name":"Yellow","Slug":"yellow"},{"Example":"#CDDC39","Name":"Lime","Slug":"lime"},{"Example":"#66BB6A","Name":"Green","Slug":"green"},{"Example":"#009688","Name":"Teal","Slug":"teal"},{"Example":"#00BCD4","Name":"Cyan","Slug":"cyan"},{"Example":"#2196F3","Name":"Blue","Slug":"blue"},{"Example":"#A1887F","Name":"Brown","Slug":"brown"},{"Example":"#F5F5F5","Name":"White","Slug":"white"},{"Example":"#9E9E9E","Name":"Grey","Slug":"grey"},{"Example":"#212121","Name":"Black","Slug":"black"}],"categories":[{"UID":"lqb6y631re96cper","Slug":"animal","Name":"Animal"},{"UID":"lqb6y5gvo9avdfx5","Slug":"architecture","Name":"Architecture"},{"UID":"lqb6y633nhfj1uzt","Slug":"bird","Name":"Bird"},{"UID":"lqb6y633g3hxg1aq","Slug":"farm","Name":"Farm"},{"UID":"lqb6y4i1ez9cw5bi","Slug":"nature","Name":"Nature"},{"UID":"lqb6y4f2v7dw8irs","Slug":"plant","Name":"Plant"},{"UID":"lqb6y6s2ohhmu0fn","Slug":"reptile","Name":"Reptile"},{"UID":"lqb6y6ctgsq2g2np","Slug":"water","Name":"Water"}],"clip":160,"server":{"cores":2,"routines":23,"memory":{"used":1224531272,"reserved":1416904088,"info":"Used 1.2 GB / Reserved 1.4 GB"}}};

let chai = require("chai/chai");
let assert = chai.assert;

describe("common/util", () => {
    it("should return duration 3ns",  () => {
        const duration = Util.duration(-3);
        assert.equal(duration, "3ns");
    });
    it("should return duration 0s",  () => {
        const duration = Util.duration(0);
        assert.equal(duration, "0s");
    });
    it("should return duration 2µs",  () => {
        const duration = Util.duration(2000);
        assert.equal(duration, "2µs");
    });
    it("should return duration 4ms",  () => {
        const duration = Util.duration(4000000);
        assert.equal(duration, "4ms");
    });
    it("should return duration 6s",  () => {
        const duration = Util.duration(6000000000);
        assert.equal(duration, "00:00:06");
    });
    it("should return duration 10min",  () => {
        const duration = Util.duration(600000000000);
        assert.equal(duration, "00:10:00");
    });
    it("should convert -1 to roman",  () => {
        const roman = Util.arabicToRoman(-1);
        assert.equal(roman, "");
    });
    it("should convert 2500 to roman",  () => {
        const roman = Util.arabicToRoman(2500);
        assert.equal(roman, "MMD");
    });
    it("should convert 112 to roman",  () => {
        const roman = Util.arabicToRoman(112);
        assert.equal(roman, "CXII");
    });
    it("should convert 9 to roman",  () => {
        const roman = Util.arabicToRoman(9);
        assert.equal(roman, "IX");
    });
    it("should truncate xxx",  () => {
        const result = Util.truncate("teststring");
        assert.equal(result, "teststring");
    });
    it("should truncate xxx",  () => {
        const result = Util.truncate("teststring for mocha", 5, "ng");
        assert.equal(result, "tesng");
    });
});
