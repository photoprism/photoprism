import Viewer from "common/viewer";

let chai = require('../../../node_modules/chai/chai');
let assert = chai.assert;

describe("common/viewer", () => {
    it("should construct viewer",  () => {
        const viewer = new Viewer();
        assert.equal(viewer.photos, "");
        assert.equal(viewer.el, null);
    });
});