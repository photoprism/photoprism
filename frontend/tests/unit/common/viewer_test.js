import assert from "assert";
import Viewer from "common/viewer";

describe("common/viewer", () => {
    it("should construct viewer",  () => {
        const viewer = new Viewer();
        assert.equal(viewer.photos, "");
        assert.equal(viewer.el, null);
    });
});