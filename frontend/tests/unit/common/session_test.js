import Session from 'common/session';
import Config from 'common/config'
import User from 'model/user';
import MockAdapter from "axios-mock-adapter";
import Api from "common/api";

window.clientConfig = {"albums":[{"ID":1,"CoverUUID":"","AlbumUUID":"aq4gs3w2x9r0m7ppx","AlbumSlug":"cat-power","AlbumName":"Cat Power","AlbumDescription":"","AlbumNotes":"","AlbumViews":0,"AlbumFavorite":true,"AlbumPublic":false,"AlbumLat":0,"AlbumLng":0,"AlbumRadius":0,"AlbumOrder":"","AlbumTemplate":"","CreatedAt":"2020-01-21T15:52:45Z","UpdatedAt":"2020-01-21T15:52:45Z","DeletedAt":null}],"author":"PhotoPrism.org","cameras":[{"ID":60009,"CameraSlug":"neingrenze-5000t","CameraModel":"5000T","CameraMake":"NEINGRENZE","CameraType":"","CameraOwner":"","CameraDescription":"","CameraNotes":"","CreatedAt":"2020-01-15T17:26:05Z","UpdatedAt":"2020-01-15T17:26:05Z","DeletedAt":null},{"ID":60007,"CameraSlug":"olympus-optical-co-ltd-c2500l","CameraModel":"C2500L      ","CameraMake":"OLYMPUS OPTICAL CO.,LTD","CameraType":"","CameraOwner":"","CameraDescription":"","CameraNotes":"","CreatedAt":"2020-01-15T17:25:33Z","UpdatedAt":"2020-01-15T17:27:51Z","DeletedAt":null},{"ID":60004,"CameraSlug":"canon-eos-10d","CameraModel":"EOS 10D","CameraMake":"Canon","CameraType":"","CameraOwner":"","CameraDescription":"","CameraNotes":"","CreatedAt":"2020-01-15T17:25:10Z","UpdatedAt":"2020-01-15T17:27:42Z","DeletedAt":null},{"ID":60006,"CameraSlug":"canon-eos-5d","CameraModel":"EOS 5D","CameraMake":"Canon","CameraType":"","CameraOwner":"","CameraDescription":"","CameraNotes":"","CreatedAt":"2020-01-15T17:25:31Z","UpdatedAt":"2020-01-22T09:33:39Z","DeletedAt":null},{"ID":60002,"CameraSlug":"canon-eos-6d","CameraModel":"EOS 6D","CameraMake":"Canon","CameraType":"","CameraOwner":"","CameraDescription":"","CameraNotes":"","CreatedAt":"2020-01-15T17:25:04Z","UpdatedAt":"2020-01-20T10:36:41Z","DeletedAt":null},{"ID":60005,"CameraSlug":"canon-eos-7d","CameraModel":"EOS 7D","CameraMake":"Canon","CameraType":"","CameraOwner":"","CameraDescription":"","CameraNotes":"","CreatedAt":"2020-01-15T17:25:27Z","UpdatedAt":"2020-01-15T17:27:55Z","DeletedAt":null},{"ID":30001,"CameraSlug":"unknown","CameraModel":"Unknown","CameraMake":"","CameraType":"","CameraOwner":"","CameraDescription":"","CameraNotes":"","CreatedAt":"2020-01-15T03:59:45Z","UpdatedAt":"2020-01-16T22:45:46Z","DeletedAt":null},{"ID":60008,"CameraSlug":"apple-iphone-4s","CameraModel":"iPhone 4S","CameraMake":"Apple","CameraType":"","CameraOwner":"","CameraDescription":"","CameraNotes":"","CreatedAt":"2020-01-15T17:25:34Z","UpdatedAt":"2020-01-15T17:25:34Z","DeletedAt":null},{"ID":60001,"CameraSlug":"apple-iphone-5s","CameraModel":"iPhone 5s","CameraMake":"Apple","CameraType":"","CameraOwner":"","CameraDescription":"","CameraNotes":"","CreatedAt":"2020-01-15T17:25:03Z","UpdatedAt":"2020-01-15T17:27:53Z","DeletedAt":null},{"ID":60003,"CameraSlug":"apple-iphone-6","CameraModel":"iPhone 6","CameraMake":"Apple","CameraType":"","CameraOwner":"","CameraDescription":"","CameraNotes":"","CreatedAt":"2020-01-15T17:25:10Z","UpdatedAt":"2020-01-16T22:45:43Z","DeletedAt":null},{"ID":1,"CameraSlug":"apple-iphone-se","CameraModel":"iPhone SE","CameraMake":"Apple","CameraType":"","CameraOwner":"","CameraDescription":"","CameraNotes":"","CreatedAt":"2020-01-14T01:16:46Z","UpdatedAt":"2020-01-15T17:27:15Z","DeletedAt":null}],"categories":[{"LabelName":"airport","Title":"Airport"},{"LabelName":"animal","Title":"Animal"},{"LabelName":"architecture","Title":"Architecture"},{"LabelName":"beach","Title":"Beach"},{"LabelName":"beetle","Title":"Beetle"},{"LabelName":"beverage","Title":"Beverage"},{"LabelName":"bird","Title":"Bird"},{"LabelName":"bridge","Title":"Bridge"},{"LabelName":"building","Title":"Building"},{"LabelName":"car","Title":"Car"},{"LabelName":"cat","Title":"Cat"},{"LabelName":"computer","Title":"Computer"},{"LabelName":"drinks","Title":"Drinks"},{"LabelName":"farm","Title":"Farm"},{"LabelName":"flower","Title":"Flower"},{"LabelName":"fruit","Title":"Fruit"},{"LabelName":"info","Title":"Info"},{"LabelName":"kitchen","Title":"Kitchen"},{"LabelName":"landscape","Title":"Landscape"},{"LabelName":"lizard","Title":"Lizard"},{"LabelName":"monkey","Title":"Monkey"},{"LabelName":"nature","Title":"Nature"},{"LabelName":"office","Title":"Office"},{"LabelName":"outdoor","Title":"Outdoor"},{"LabelName":"people","Title":"People"},{"LabelName":"plant","Title":"Plant"},{"LabelName":"portrait","Title":"Portrait"},{"LabelName":"reptile","Title":"Reptile"},{"LabelName":"rodent","Title":"Rodent"},{"LabelName":"sand","Title":"Sand"},{"LabelName":"shop","Title":"Shop"},{"LabelName":"table","Title":"Table"},{"LabelName":"tower","Title":"Tower"},{"LabelName":"train","Title":"Train"},{"LabelName":"vegetables","Title":"Vegetables"},{"LabelName":"vehicle","Title":"Vehicle"},{"LabelName":"water","Title":"Water"},{"LabelName":"wildlife","Title":"Wildlife"}],"colors":[{"example":"#AB47BC","label":"Purple","name":"purple"},{"example":"#FF00FF","label":"Magenta","name":"magenta"},{"example":"#EC407A","label":"Pink","name":"pink"},{"example":"#EF5350","label":"Red","name":"red"},{"example":"#FFA726","label":"Orange","name":"orange"},{"example":"#D4AF37","label":"Gold","name":"gold"},{"example":"#FDD835","label":"Yellow","name":"yellow"},{"example":"#CDDC39","label":"Lime","name":"lime"},{"example":"#66BB6A","label":"Green","name":"green"},{"example":"#009688","label":"Teal","name":"teal"},{"example":"#00BCD4","label":"Cyan","name":"cyan"},{"example":"#2196F3","label":"Blue","name":"blue"},{"example":"#A1887F","label":"Brown","name":"brown"},{"example":"#F5F5F5","label":"White","name":"white"},{"example":"#9E9E9E","label":"Grey","name":"grey"},{"example":"#212121","label":"Black","name":"black"}],"copyright":"(c) 2018-2020 The PhotoPrism contributors \u003chello@photoprism.org\u003e","count":{"photos":712,"favorites":9,"private":2,"stories":2,"labels":86,"albums":1,"countries":8,"places":44},"countries":[{"code":"bw","name":"Botswana"},{"code":"cu","name":"Cuba"},{"code":"fr","name":"France"},{"code":"de","name":"Germany"},{"code":"gr","name":"Greece"},{"code":"za","name":"South Africa"},{"code":"ch","name":"Switzerland"},{"code":"us","name":"USA"},{"code":"zz","name":"Unknown"}],"cssHash":"9104a800d818ae72175b68dfcbc44c3536351746","debug":true,"description":"Personal Photo Management","experimental":true,"flags":"public debug experimental","jsHash":"12f304d05e6ef4f50f6f5394e5a95fc697426204","lenses":[{"ID":60007,"LensSlug":"28-0-75-0-mm","LensModel":"28.0-75.0 mm","LensMake":"","LensType":"","LensOwner":"","LensDescription":"","LensNotes":"","CreatedAt":"2020-01-15T17:26:02Z","UpdatedAt":"2020-01-15T17:26:02Z","DeletedAt":null},{"ID":60005,"LensSlug":"ef100mm-f-2-8l-macro-is-usm","LensModel":"EF100mm f/2.8L Macro IS USM","LensMake":"","LensType":"","LensOwner":"","LensDescription":"","LensNotes":"","CreatedAt":"2020-01-15T17:25:29Z","UpdatedAt":"2020-01-15T17:27:46Z","DeletedAt":null},{"ID":60004,"LensSlug":"ef16-35mm-f-2-8l-ii-usm","LensModel":"EF16-35mm f/2.8L II USM","LensMake":"","LensType":"","LensOwner":"","LensDescription":"","LensNotes":"","CreatedAt":"2020-01-15T17:25:28Z","UpdatedAt":"2020-01-15T17:26:41Z","DeletedAt":null},{"ID":60002,"LensSlug":"ef24-105mm-f-4l-is-usm","LensModel":"EF24-105mm f/4L IS USM","LensMake":"","LensType":"","LensOwner":"","LensDescription":"","LensNotes":"","CreatedAt":"2020-01-15T17:25:04Z","UpdatedAt":"2020-01-16T22:22:51Z","DeletedAt":null},{"ID":60008,"LensSlug":"ef35mm-f-2-is-usm","LensModel":"EF35mm f/2 IS USM","LensMake":"","LensType":"","LensOwner":"","LensDescription":"","LensNotes":"","CreatedAt":"2020-01-15T17:26:16Z","UpdatedAt":"2020-01-20T10:36:41Z","DeletedAt":null},{"ID":60006,"LensSlug":"ef70-200mm-f-4l-is-usm","LensModel":"EF70-200mm f/4L IS USM","LensMake":"","LensType":"","LensOwner":"","LensDescription":"","LensNotes":"","CreatedAt":"2020-01-15T17:25:53Z","UpdatedAt":"2020-01-16T22:22:57Z","DeletedAt":null},{"ID":30001,"LensSlug":"unknown","LensModel":"Unknown","LensMake":"","LensType":"","LensOwner":"","LensDescription":"","LensNotes":"","CreatedAt":"2020-01-15T03:59:45Z","UpdatedAt":"2020-01-22T09:33:39Z","DeletedAt":null},{"ID":60001,"LensSlug":"iphone-5s-back-camera-4-12mm-f-2-2","LensModel":"iPhone 5s back camera 4.12mm f/2.2","LensMake":"Apple","LensType":"","LensOwner":"","LensDescription":"","LensNotes":"","CreatedAt":"2020-01-15T17:25:03Z","UpdatedAt":"2020-01-15T17:27:53Z","DeletedAt":null},{"ID":60003,"LensSlug":"iphone-6-back-camera-4-15mm-f-2-2","LensModel":"iPhone 6 back camera 4.15mm f/2.2","LensMake":"Apple","LensType":"","LensOwner":"","LensDescription":"","LensNotes":"","CreatedAt":"2020-01-15T17:25:10Z","UpdatedAt":"2020-01-16T22:45:43Z","DeletedAt":null},{"ID":90001,"LensSlug":"iphone-6-front-camera-2-65mm-f-2-2","LensModel":"iPhone 6 front camera 2.65mm f/2.2","LensMake":"Apple","LensType":"","LensOwner":"","LensDescription":"","LensNotes":"","CreatedAt":"2020-01-16T22:22:53Z","UpdatedAt":"2020-01-16T22:44:43Z","DeletedAt":null},{"ID":1,"LensSlug":"iphone-se-back-camera-4-15mm-f-2-2","LensModel":"iPhone SE back camera 4.15mm f/2.2","LensMake":"Apple","LensType":"","LensOwner":"","LensDescription":"","LensNotes":"","CreatedAt":"2020-01-14T01:16:46Z","UpdatedAt":"2020-01-15T17:27:15Z","DeletedAt":null}],"name":"PhotoPrism","pos":{"photo":"pq45sez3bayygvd6o","location":"479a03fda574","utc":"2019-05-31T18:15:08Z","lat":48.29912222222222,"lng":8.929649999999999},"public":true,"readonly":false,"settings":{"theme":"default","language":"en"},"subtitle":"Browse your life","thumbnails":[{"Name":"fit_720","Width":720,"Height":720},{"Name":"fit_2048","Width":2048,"Height":2048},{"Name":"fit_2560","Width":2560,"Height":1600},{"Name":"fit_3840","Width":3840,"Height":2400},{"Name":"fit_1280","Width":1280,"Height":1024},{"Name":"fit_1920","Width":1920,"Height":1200}],"title":"PhotoPrism","twitter":"@browseyourlife","uploadNSFW":false,"url":"http://localhost:2342/","version":"200122-f569c3a-Linux-x86_64-DEBUG","years":[2020,2019,2018,2017,2016,2015,2014,2013,2012,2011,2010,2004,2002]};

let chai = require('../../../node_modules/chai/chai');
let assert = chai.assert;
const config = new Config(window.localStorage, window.clientConfig);

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
