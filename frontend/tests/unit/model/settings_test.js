import Settings from "model/settings";
import MockAdapter from "axios-mock-adapter";
import Api from "common/api";

let chai = require("chai/chai");
let assert = chai.assert;

const mock = new MockAdapter(Api);

mock
    .onGet("api/v1/settings").reply(200, {"download": true, "language": "de"})
    .onPost("api/v1/settings").reply(200, {"download": true, "language": "en"});


describe("model/settings", () => {

    it("should return if key was changed",  () => {
        const model = new Settings({"language": "de", "download": false});
        assert.equal(model.changed("download"), false);
        assert.equal(model.changed("language"), false);
    });

    it("should load settings",  (done) => {
        const model = new Settings();
        model.load().then(
            (response) => {
                assert.equal(response.download, true);
                assert.equal(response.language, "de");
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });

    it("should save settings",  (done) => {
        const model = new Settings({"language": "en"});
        model.save().then(
            (response) => {
                assert.equal(response.download, true);
                assert.equal(response.language, "en");
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });
});