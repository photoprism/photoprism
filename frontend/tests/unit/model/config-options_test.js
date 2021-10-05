import "../fixtures";
import ConfigOptions from "model/config-options";

let chai = require("chai/chai");
let assert = chai.assert;

describe("model/config-options", () => {
  it("should get options defaults", () => {
    const values = {};
    const options = new ConfigOptions(values);
    const result = options.getDefaults();
    assert.equal(result.Debug, false);
    assert.equal(result.ReadOnly, false);
    assert.equal(result.ThumbSize, 0);
  });

  it("should test changed", () => {
    const values = {};
    const options = new ConfigOptions(values);
    assert.equal(options.changed(), false);
  });

  it("should load options", (done) => {
    const values = {};
    const options = new ConfigOptions(values);
    options
      .load()
      .then((response) => {
        assert.equal(response.success, "ok");
        done();
      })
      .catch((error) => {
        done(error);
      });
    assert.equal(options.changed(), false);
  });

  it("should save options", (done) => {
    const values = { Debug: true };
    const options = new ConfigOptions(values);
    options
      .save()
      .then((response) => {
        assert.equal(response.success, "ok");
        done();
      })
      .catch((error) => {
        done(error);
      });
    assert.equal(options.changed(), false);
  });
});
