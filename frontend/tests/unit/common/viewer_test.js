import "../fixtures";
import Viewer from "common/viewer";

let chai = require("chai/chai");
let assert = chai.assert;

describe("common/viewer", () => {
  it("should construct viewer", () => {
    const viewer = new Viewer();
    assert.equal(viewer.el, null);
  });
});
