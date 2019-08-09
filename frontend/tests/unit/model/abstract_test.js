import assert from "assert";
import Abstract from "model/abstract";
import Album from "model/album";
import Label from "model/label";

describe("model/abstract", () => {
    it("should set values",  () => {
        const values = {id: 5, LabelName: "Black Cat", LabelSlug: "black-cat"};
        const label = new Label(values);
        assert.equal(label.LabelName, "Black Cat");
        assert.equal(label.LabelSlug, "black-cat");
        label.setValues();
        assert.equal(label.LabelName, "Black Cat");
        assert.equal(label.LabelSlug, "black-cat");
        const values2 = {id: 6, LabelName: "White Cat", LabelSlug: "white-cat"};
        label.setValues(values2);
        assert.equal(label.LabelName, "White Cat");
        assert.equal(label.LabelSlug, "white-cat");
    });

    it("should get values",  () => {
        const values = {id: 5, AlbumName: "Christmas 2019", AlbumSlug: "christmas-2019", AlbumUUID: 66};
        const album = new Album(values);
        const result = album.getValues();
        assert.equal(result.AlbumName, "Christmas 2019");
        assert.equal(result.AlbumUUID, 66);
    });

    it("should get id",  () => {
        const values = {id: 5, AlbumName: "Christmas 2019", AlbumSlug: "christmas-2019", AlbumUUID: 66};
        const album = new Album(values);
        const result = album.getId();
        assert.equal(result, 66);
    });

    it("should test if id exists",  () => {
        const values = {id: 5, AlbumName: "Christmas 2019", AlbumSlug: "christmas-2019", AlbumUUID: 66};
        const album = new Album(values);
        const result = album.hasId();
        assert.equal(result, true);
    });

    it("should get model name",  () => {
        const result = Abstract.getModelName();
        assert.equal(result, "Item");
    });

});