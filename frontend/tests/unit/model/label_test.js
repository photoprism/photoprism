import assert from "assert";
import Label from "model/label";

describe.only("model/label", () => {
    it("should get label entity name",  () => {
        const values = {id: 5, LabelName: "Black Cat", LabelSlug: "black-cat"};
        const label = new Label(values);
        const result = label.getEntityName();
        assert.equal(result, "black-cat");
    });

    it("should get label id",  () => {
        const values = {id: 5, LabelName: "Black Cat", LabelSlug: "black-cat"};
        const label = new Label(values);
        const result = label.getId();
        assert.equal(result, "black-cat");
    });

    it("should get label title",  () => {
        const values = {id: 5, LabelName: "Black Cat", LabelSlug: "black-cat"};
        const label = new Label(values);
        const result = label.getTitle();
        assert.equal(result, "Black Cat");
    });

    it("should get thumbnail url",  () => {
        const values = {id: 5, LabelName: "Black Cat", LabelSlug: "black-cat"};
        const label = new Label(values);
        const result = label.getThumbnailUrl("xyz");
        assert.equal(result, "/api/v1/labels/black-cat/thumbnail/xyz");
    });

    it("should get thumbnail src set",  () => {
        const values = {id: 5, LabelName: "Black Cat", LabelSlug: "black-cat"};
        const label = new Label(values);
        const result = label.getThumbnailSrcset("");
        console.log(result);
        assert.equal(result, "/api/v1/labels/black-cat/thumbnail/fit_720 720w, /api/v1/labels/black-cat/thumbnail/fit_1280 1280w, /api/v1/labels/black-cat/thumbnail/fit_1920 1920w, /api/v1/labels/black-cat/thumbnail/fit_2560 2560w, /api/v1/labels/black-cat/thumbnail/fit_3840 3840w");
    });

});