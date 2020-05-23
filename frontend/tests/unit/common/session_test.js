import Session from 'common/session';
import Config from 'common/config'
import User from 'model/user';
import MockAdapter from "axios-mock-adapter";
import Api from "common/api";

window.__CONFIG__ = {"albums":[{"ID":1,"UID":"aqasltilq1bnar68","CoverUID":"","ParentUID":"","FolderUID":"","Slug":"example-album","Name":"Example Album","Type":"","Filter":"","Description":"","Notes":"","Order":"oldest","Template":"","Favorite":true,"Links":null,"CreatedAt":"2020-05-23T16:29:42Z","UpdatedAt":"2020-05-23T16:29:42Z"}],"author":"PhotoPrism.org","cameras":[{"ID":8,"Slug":"apple-iphone-5s","Model":"iPhone 5s","Make":"Apple"},{"ID":9,"Slug":"apple-iphone-6","Model":"iPhone 6","Make":"Apple"},{"ID":13,"Slug":"apple-iphone-se","Model":"iPhone SE","Make":"Apple"},{"ID":17,"Slug":"apple-iphone-x","Model":"iPhone X","Make":"Apple"},{"ID":7,"Slug":"canon-eos-60d","Model":"EOS 60D","Make":"Canon"},{"ID":5,"Slug":"canon-eos-6d","Model":"EOS 6D","Make":"Canon"},{"ID":14,"Slug":"canon-eos-700d","Model":"EOS 700D","Make":"Canon"},{"ID":3,"Slug":"canon-eos-7d","Model":"EOS 7D","Make":"Canon"},{"ID":16,"Slug":"gopro-hero4-black","Model":"HERO4 Black","Make":"GoPro"},{"ID":19,"Slug":"huawei-ele-l29","Model":"ELE-L29","Make":"HUAWEI"},{"ID":18,"Slug":"motorola-moto-g-7-power","Model":"moto g(7) power","Make":"motorola"},{"ID":12,"Slug":"nikon-corporation-nikon-d90","Model":"NIKON D90","Make":"NIKON CORPORATION"},{"ID":11,"Slug":"panasonic-dc-gh5s","Model":"DC-GH5S","Make":"Panasonic"},{"ID":4,"Slug":"panasonic-dmc-fz100","Model":"DMC-FZ100","Make":"Panasonic"},{"ID":10,"Slug":"panasonic-dmc-gx85","Model":"DMC-GX85","Make":"Panasonic"},{"ID":6,"Slug":"panasonic-dmc-lx5","Model":"DMC-LX5","Make":"Panasonic"},{"ID":2,"Slug":"samsung-gt-i9000","Model":"GT-I9000","Make":"SAMSUNG"},{"ID":30001,"Slug":"samsung-sm-g900f","Model":"SM-G900F","Make":"samsung"},{"ID":15,"Slug":"sony-ilce-5100","Model":"ILCE-5100","Make":"SONY"},{"ID":1,"Slug":"zz","Model":"Unknown","Make":""}],"categories":[{"UID":"lqaslsn3njvs7phk","Slug":"animal","Name":"Animal"},{"UID":"lqaslv92v88svntk","Slug":"architecture","Name":"Architecture"},{"UID":"lqaslssgexee300l","Slug":"beach","Name":"Beach"},{"UID":"lqaslte2t6c421m8","Slug":"beetle","Name":"Beetle"},{"UID":"lqaslsw2hb8ldc8r","Slug":"bird","Name":"Bird"},{"UID":"lqaslt931dbwp7qv","Slug":"cat","Name":"Cat"},{"UID":"lqaslu6z3bx3gw68","Slug":"farm","Name":"Farm"},{"UID":"lqaslso23vakxxjc","Slug":"insect","Name":"Insect"},{"UID":"lqasluk337xy8ig3","Slug":"monkey","Name":"Monkey"},{"UID":"lqasluq12dlf4qxp","Slug":"people","Name":"People"},{"UID":"lqaslw21ovq4fkzi","Slug":"reptile","Name":"Reptile"},{"UID":"lqasluq1v7gv0h5m","Slug":"snow","Name":"Snow"},{"UID":"lqaslv9187r4rwuz","Slug":"tower","Name":"Tower"},{"UID":"lqaslss1i0zlvdqz","Slug":"water","Name":"Water"},{"UID":"lqaslt93im1tw5b7","Slug":"wildlife","Name":"Wildlife"}],"clip":160,"colors":[{"Example":"#AB47BC","Name":"Purple","Slug":"purple"},{"Example":"#FF00FF","Name":"Magenta","Slug":"magenta"},{"Example":"#EC407A","Name":"Pink","Slug":"pink"},{"Example":"#EF5350","Name":"Red","Slug":"red"},{"Example":"#FFA726","Name":"Orange","Slug":"orange"},{"Example":"#D4AF37","Name":"Gold","Slug":"gold"},{"Example":"#FDD835","Name":"Yellow","Slug":"yellow"},{"Example":"#CDDC39","Name":"Lime","Slug":"lime"},{"Example":"#66BB6A","Name":"Green","Slug":"green"},{"Example":"#009688","Name":"Teal","Slug":"teal"},{"Example":"#00BCD4","Name":"Cyan","Slug":"cyan"},{"Example":"#2196F3","Name":"Blue","Slug":"blue"},{"Example":"#A1887F","Name":"Brown","Slug":"brown"},{"Example":"#F5F5F5","Name":"White","Slug":"white"},{"Example":"#9E9E9E","Name":"Grey","Slug":"grey"},{"Example":"#212121","Name":"Black","Slug":"black"}],"copyright":"(c) 2018-2020 PhotoPrism.org \u003chello@photoprism.org\u003e","count":{"photos":74,"videos":5,"hidden":0,"favorites":4,"private":2,"review":1,"albums":1,"folders":41,"moments":0,"countries":6,"places":17,"labels":24,"labelMaxPhotos":7},"countries":[{"ID":"fr","Slug":"france","Name":"France"},{"ID":"de","Slug":"germany","Name":"Germany"},{"ID":"gr","Slug":"greece","Name":"Greece"},{"ID":"it","Slug":"italy","Name":"Italy"},{"ID":"za","Slug":"south-africa","Name":"South Africa"},{"ID":"us","Slug":"usa","Name":"USA"},{"ID":"zz","Slug":"zz","Name":"Unknown"}],"cssHash":"4671c86b","debug":true,"description":"Personal Photo Management","disableSettings":false,"experimental":true,"flags":"public debug experimental settings","jsHash":"08ea4b1d","lenses":[{"ID":5,"Slug":"10-0-22-0-mm","Model":"10.0 - 22.0 mm","Make":"","Type":""},{"ID":15,"Slug":"16-0-35-0-mm","Model":"16.0 - 35.0 mm","Make":"","Type":""},{"ID":13,"Slug":"18-0-55-0-mm","Model":"18.0 - 55.0 mm","Make":"","Type":""},{"ID":10,"Slug":"18-300mm-f-3-5-6-3","Model":"18-300mm f/3.5-6.3","Make":"","Type":""},{"ID":7,"Slug":"24-0-105-0-mm","Model":"24.0 - 105.0 mm","Make":"","Type":""},{"ID":8,"Slug":"35-0-mm","Model":"35.0 mm","Make":"","Type":""},{"ID":12,"Slug":"50-0-mm","Model":"50.0 mm","Make":"","Type":""},{"ID":14,"Slug":"e-pz-16-50mm-f3-5-5-6-oss","Model":"E PZ 16-50mm F3.5-5.6 OSS","Make":"","Type":""},{"ID":3,"Slug":"ef100mm-f-2-8l-macro-is-usm","Model":"EF100mm f/2.8L Macro IS USM","Make":"","Type":""},{"ID":2,"Slug":"ef24-105mm-f-4l-is-usm","Model":"EF24-105mm f/4L IS USM","Make":"","Type":""},{"ID":4,"Slug":"ef70-200mm-f-4l-is-usm","Model":"EF70-200mm f/4L IS USM","Make":"","Type":""},{"ID":6,"Slug":"iphone-5s-back-camera-4-12mm-f-2-2","Model":"iPhone 5s back camera 4.12mm f/2.2","Make":"Apple","Type":""},{"ID":9,"Slug":"iphone-6-back-camera-4-15mm-f-2-2","Model":"iPhone 6 back camera 4.15mm f/2.2","Make":"Apple","Type":""},{"ID":11,"Slug":"iphone-se-back-camera-4-15mm-f-2-2","Model":"iPhone SE back camera 4.15mm f/2.2","Make":"Apple","Type":""},{"ID":1,"Slug":"zz","Model":"Unknown","Make":"","Type":""}],"name":"PhotoPrism","pos":{"photo":"pqaslyb76kzu6h3r","location":"47a85a7fbc6c","utc":"2020-05-14T09:34:41Z","lat":52.4649,"lng":13.3148},"public":true,"readonly":false,"server":{"cores":2,"routines":43,"memory":{"used":67112104,"reserved":144427272,"info":"Used 67 MB / Reserved 144 MB"}},"settings":{"theme":"default","language":"en","templates":{"default":"index.tmpl"},"maps":{"animate":0,"style":"streets"},"features":{"archive":true,"private":true,"review":true,"upload":true,"import":true,"folders":true,"moments":true,"labels":true,"places":true,"download":true,"edit":true,"share":true,"logs":true},"import":{"path":"/","move":false},"index":{"path":"/","convert":true,"rescan":true,"group":true}},"subtitle":"Browse your life","thumbnails":[{"Name":"fit_720","Width":720,"Height":720},{"Name":"fit_1920","Width":1920,"Height":1200},{"Name":"fit_2048","Width":2048,"Height":2048},{"Name":"fit_1280","Width":1280,"Height":1024},{"Name":"fit_2560","Width":2560,"Height":1600},{"Name":"fit_3840","Width":3840,"Height":2400}],"title":"PhotoPrism","uploadNSFW":false,"url":"http://localhost:2342/","version":"200523-fc62279-Linux-x86_64-DEBUG","years":[2020,2019,2018,2015,2014,2013,2012,2011,2000]};

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
