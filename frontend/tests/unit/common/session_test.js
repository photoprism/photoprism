import Session from 'common/session';
import Config from 'common/config'
import User from 'model/user';
import MockAdapter from "axios-mock-adapter";
import Api from "common/api";

window.__CONFIG__ = {"flags":"public debug experimental settings","name":"PhotoPrism","url":"http://localhost:2342/","title":"PhotoPrism","subtitle":"Browse your life","description":"Personal Photo Management","author":"PhotoPrism.org","version":"200527-5453cf2-Linux-x86_64-DEBUG","copyright":"(c) 2018-2020 PhotoPrism.org \u003chello@photoprism.org\u003e","debug":true,"readonly":false,"uploadNSFW":false,"public":true,"experimental":true,"disableSettings":false,"albums":[],"cameras":[{"ID":30003,"Slug":"apple-iphone-6","Model":"iPhone 6","Make":"Apple"},{"ID":30001,"Slug":"apple-iphone-se","Model":"iPhone SE","Make":"Apple"},{"ID":30004,"Slug":"canon-eos-6d","Model":"EOS 6D","Make":"Canon"},{"ID":30002,"Slug":"canon-eos-m6","Model":"EOS M6","Make":"Canon"},{"ID":30006,"Slug":"huawei-ele-l29","Model":"ELE-L29","Make":"HUAWEI"},{"ID":30005,"Slug":"motorola-moto-g-4","Model":"Moto G (4)","Make":"Motorola"},{"ID":1,"Slug":"zz","Model":"Unknown","Make":""}],"lenses":[{"ID":30003,"Slug":"22-0-mm","Model":"22.0 mm","Make":"","Type":""},{"ID":30005,"Slug":"ef16-35mm-f-2-8l-ii-usm","Model":"EF16-35mm f/2.8L II USM","Make":"","Type":""},{"ID":30004,"Slug":"iphone-6-back-camera-4-15mm-f-2-2","Model":"iPhone 6 back camera 4.15mm f/2.2","Make":"Apple","Type":""},{"ID":30001,"Slug":"iphone-se-back-camera-4-15mm-f-2-2","Model":"iPhone SE back camera 4.15mm f/2.2","Make":"Apple","Type":""},{"ID":30002,"Slug":"iphone-se-front-camera-2-15mm-f-2-4","Model":"iPhone SE front camera 2.15mm f/2.4","Make":"Apple","Type":""},{"ID":1,"Slug":"zz","Model":"Unknown","Make":"","Type":""}],"countries":[{"ID":"at","Slug":"austria","Name":"Austria"},{"ID":"bw","Slug":"botswana","Name":"Botswana"},{"ID":"ca","Slug":"canada","Name":"Canada"},{"ID":"fr","Slug":"france","Name":"France"},{"ID":"de","Slug":"germany","Name":"Germany"},{"ID":"gr","Slug":"greece","Name":"Greece"},{"ID":"za","Slug":"south-africa","Name":"South Africa"},{"ID":"us","Slug":"usa","Name":"USA"},{"ID":"zz","Slug":"zz","Name":"Unknown"}],"thumbnails":[{"Name":"fit_720","Width":720,"Height":720},{"Name":"fit_1280","Width":1280,"Height":1024},{"Name":"fit_2560","Width":2560,"Height":1600},{"Name":"fit_3840","Width":3840,"Height":2400},{"Name":"fit_1920","Width":1920,"Height":1200},{"Name":"fit_2048","Width":2048,"Height":2048}],"downloadToken":"2y71e0sr","thumbToken":"static","jsHash":"14ba2de4","cssHash":"2b327230","settings":{"theme":"default","language":"en","templates":{"default":"index.tmpl"},"maps":{"animate":0,"style":"streets"},"features":{"archive":true,"private":true,"review":true,"upload":true,"import":true,"files":true,"moments":true,"labels":true,"places":true,"download":true,"edit":true,"share":true,"logs":true},"import":{"path":"/","move":false},"index":{"path":"/","convert":true,"rescan":false,"group":true}},"count":{"photos":385,"videos":1,"hidden":0,"favorites":1,"private":2,"review":4,"stories":0,"albums":0,"folders":14,"files":394,"moments":0,"countries":8,"places":0,"labels":46,"labelMaxPhotos":54},"pos":{"uid":"pqazcltc1x8d12lo","loc":"4777dc437584","utc":"2020-02-14T12:44:19Z","lat":47.207123,"lng":11.823489},"years":[2020,2019,2018,2017,2016],"colors":[{"Example":"#AB47BC","Name":"Purple","Slug":"purple"},{"Example":"#FF00FF","Name":"Magenta","Slug":"magenta"},{"Example":"#EC407A","Name":"Pink","Slug":"pink"},{"Example":"#EF5350","Name":"Red","Slug":"red"},{"Example":"#FFA726","Name":"Orange","Slug":"orange"},{"Example":"#D4AF37","Name":"Gold","Slug":"gold"},{"Example":"#FDD835","Name":"Yellow","Slug":"yellow"},{"Example":"#CDDC39","Name":"Lime","Slug":"lime"},{"Example":"#66BB6A","Name":"Green","Slug":"green"},{"Example":"#009688","Name":"Teal","Slug":"teal"},{"Example":"#00BCD4","Name":"Cyan","Slug":"cyan"},{"Example":"#2196F3","Name":"Blue","Slug":"blue"},{"Example":"#A1887F","Name":"Brown","Slug":"brown"},{"Example":"#F5F5F5","Name":"White","Slug":"white"},{"Example":"#9E9E9E","Name":"Grey","Slug":"grey"},{"Example":"#212121","Name":"Black","Slug":"black"}],"categories":[{"UID":"lqazz283gqjo05j9","Slug":"aircraft","Name":"Aircraft"},{"UID":"lqazyyc2xos6k0op","Slug":"airport","Name":"Airport"},{"UID":"lqazyya2wbw3045h","Slug":"animal","Name":"Animal"},{"UID":"lqazyz22d0y1ham3","Slug":"architecture","Name":"Architecture"},{"UID":"lqazz7537y79uhef","Slug":"beach","Name":"Beach"},{"UID":"lqazz1v2gbroth1y","Slug":"beverage","Name":"Beverage"},{"UID":"lqazyyf213ls8byk","Slug":"building","Name":"Building"},{"UID":"lqazyzhjuf6ud8pd","Slug":"car","Name":"Car"},{"UID":"lqazzcnzc5ejq4xx","Slug":"dining","Name":"Dining"},{"UID":"lqazz1v3t6kuuid7","Slug":"drinks","Name":"Drinks"},{"UID":"lqazz4j3rrxrh9el","Slug":"event","Name":"Event"},{"UID":"lqazz422nbeeedv5","Slug":"farm","Name":"Farm"},{"UID":"lqazz5f3leym5l14","Slug":"food","Name":"Food"},{"UID":"lqazz252fhe2ibx1","Slug":"landscape","Name":"Landscape"},{"UID":"lqazyzs20lrgueeb","Slug":"nature","Name":"Nature"},{"UID":"lqazyyh3f6phuq04","Slug":"outdoor","Name":"Outdoor"},{"UID":"lqazzef1n086vptr","Slug":"people","Name":"People"},{"UID":"lqazzm019mojdmcp","Slug":"plant","Name":"Plant"},{"UID":"lqazyy72zbezq9zt","Slug":"portrait","Name":"Portrait"},{"UID":"lqazzbqrfzbq0l9l","Slug":"sand","Name":"Sand"},{"UID":"lqazz8q1q39ksabl","Slug":"shop","Name":"Shop"},{"UID":"lqazyyt1ua4i8jpy","Slug":"train","Name":"Train"},{"UID":"lqazzdxvp4d59tx4","Slug":"vegetables","Name":"Vegetables"},{"UID":"lqazyytyvhnb6fvk","Slug":"vehicle","Name":"Vehicle"},{"UID":"lqazz4b7rukovpmc","Slug":"water","Name":"Water"},{"UID":"lqazz2t3n882fps6","Slug":"wildlife","Name":"Wildlife"}],"clip":160,"server":{"cores":2,"routines":38,"memory":{"used":56042952,"reserved":144132360,"info":"Used 56 MB / Reserved 144 MB"}}};

let chai = require('../../../node_modules/chai/chai');
let assert = chai.assert;
const config = new Config(window.localStorage, window.__CONFIG__);

describe('common/session', () => {

    const mock = new MockAdapter(Api);

    beforeEach(() => {
        window.onbeforeunload = () => 'Oh no!';
    });

    it('should construct session',  () => {
        const storage = window.localStorage;
        const session = new Session(storage, config);
        assert.equal(session.session_token, null);
    });

    it('should set, get and delete token',  () => {
        const storage = window.localStorage;
        const session = new Session(storage, config);
        assert.equal(session.session_token, null);
        session.setToken(123421);
        assert.equal(session.session_token, 123421);
        const result = session.getToken();
        assert.equal(result, 123421);
        session.deleteToken();
        assert.equal(session.session_token, null);
    });

    it('should set, get and delete user',  () => {
        const storage = window.localStorage;
        const session = new Session(storage, config);
        assert.equal(session.user, null);
        const values = {ID: 5, FirstName: "Max", LastName: "Last", Email: "test@test.com", Role: "admin"};
        const user = new User(values);
        session.setUser(user);
        assert.equal(session.user.FirstName, "Max");
        assert.equal(session.user.Role, "admin");
        const result = session.getUser();
        assert.equal(result.ID, 5);
        assert.equal(result.Email, "test@test.com");
        session.deleteUser();
        assert.equal(session.user, null);
    });

    it('should get user email',  () => {
        const storage = window.localStorage;
        const session = new Session(storage, config);
        const values = {ID: 5, FirstName: "Max", LastName: "Last", Email: "test@test.com", Role: "admin"};
        const user = new User(values);
        session.setUser(user);
        const result = session.getEmail();
        assert.equal(result, "test@test.com");
        const values2 = { FirstName: "Max", LastName: "Last", Email: "test@test.com", Role: "admin"};
        const user2 = new User(values2);
        session.setUser(user2);
        const result2 = session.getEmail();
        assert.equal(result2, "");
        session.deleteUser();
    });

    it('should get user firstname',  () => {
        const storage = window.localStorage;
        const session = new Session(storage, config);
        const values = {ID: 5, FirstName: "Max", LastName: "Last", Email: "test@test.com", Role: "admin"};
        const user = new User(values);
        session.setUser(user);
        const result = session.getFirstName();
        assert.equal(result, "Max");
        const values2 = { FirstName: "Max", LastName: "Last", Email: "test@test.com", Role: "admin"};
        const user2 = new User(values2);
        session.setUser(user2);
        const result2 = session.getFirstName();
        assert.equal(result2, "");
        session.deleteUser();
    });

    it('should get user full name',  () => {
        const storage = window.localStorage;
        const session = new Session(storage, config);
        const values = {ID: 5, FirstName: "Max", LastName: "Last", Email: "test@test.com", Role: "admin"};
        const user = new User(values);
        session.setUser(user);
        const result = session.getFullName();
        assert.equal(result, "Max Last");
        const values2 = { FirstName: "Max", LastName: "Last", Email: "test@test.com", Role: "admin"};
        const user2 = new User(values2);
        session.setUser(user2);
        const result2 = session.getFullName();
        assert.equal(result2, "");
        session.deleteUser();
    });

    it('should test whether user is set',  () => {
        const storage = window.localStorage;
        const session = new Session(storage, config);
        const values = {ID: 5, FirstName: "Max", LastName: "Last", Email: "test@test.com", Role: "admin"};
        const user = new User(values);
        session.setUser(user);
        const result = session.isUser();
        assert.equal(result, true);
        session.deleteUser();
    });

    it('should test whether user is admin',  () => {
        const storage = window.localStorage;
        const session = new Session(storage, config);
        const values = {ID: 5, FirstName: "Max", LastName: "Last", Email: "test@test.com", Role: "admin"};
        const user = new User(values);
        session.setUser(user);
        const result = session.isAdmin();
        assert.equal(result, true);
        session.deleteUser();
    });

    it('should test whether user is anonymous',  () => {
        const storage = window.localStorage;
        const session = new Session(storage, config);
        const values = {ID: 5, FirstName: "Max", LastName: "Last", Email: "test@test.com", Role: "admin"};
        const user = new User(values);
        session.setUser(user);
        const result = session.isAnonymous();
        assert.equal(result, false);
        session.deleteUser();
    });

    it('should test login and logout',  async() => {
        mock
            .onPost("session").reply(200,  {token: "8877", user: {ID: 1, Email: "test@test.com"}})
            .onDelete("session/8877").reply(200);
        const storage = window.localStorage;
        const session = new Session(storage, config);
        assert.equal(session.session_token, null);
        assert.equal(session.storage.user, undefined);
        await session.login("test@test.com", "passwd");
        assert.equal(session.session_token, 8877);
        assert.equal(session.storage.user, '{"ID":1,"Email":"test@test.com"}');
        await session.logout();
        assert.equal(session.session_token, null);
        mock.reset();
    });

});
