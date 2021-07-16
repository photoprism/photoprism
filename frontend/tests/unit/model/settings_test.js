import "../fixtures";
import Settings from "model/settings";

let chai = require("chai/chai");
let assert = chai.assert;

describe("model/settings", () => {
  it("should return if key was changed", () => {
    const model = new Settings({ ui: { language: "de", scrollbar: false } });
    assert.equal(model.changed("ui", "scrollbar"), false);
    assert.equal(model.changed("ui", "language"), false);
  });

  it("should load settings", (done) => {
    const model = new Settings({ ui: { language: "de", scrollbar: false } });
    model
      .load()
      .then((response) => {
        assert.equal(response["ui"]["scrollbar"], false);
        assert.equal(response["ui"]["language"], "de");
        done();
      })
      .catch((error) => {
        done(error);
      });
  });

  it("should save settings", (done) => {
    const model = new Settings({ ui: { language: "de", scrollbar: false } });
    model
      .save()
      .then((response) => {
        assert.equal(response["ui"]["scrollbar"], false);
        assert.equal(response["ui"]["language"], "de");
        done();
      })
      .catch((error) => {
        done(error);
      });
  });
});
