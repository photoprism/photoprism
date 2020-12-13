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

    it("should return if key was changed", () => {
        const model = new Settings({"ui": {"language": "de", "scrollbar": false}});
        assert.equal(model.changed("ui", "scrollbar"), false);
        assert.equal(model.changed("ui", "language"), false);
    });

    it("should load settings", (done) => {
        const model = new Settings({"ui": {"language": "de", "scrollbar": false}});
        model.load().then(
            (response) => {
                assert.equal(response["ui"]["scrollbar"], false);
                assert.equal(response["ui"]["language"], "de");
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });

    it("should save settings", (done) => {
        const model = new Settings({"ui": {"language": "de", "scrollbar": false}});
        model.save().then(
            (response) => {
                assert.equal(response["ui"]["scrollbar"], false);
                assert.equal(response["ui"]["language"], "de");
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });
});