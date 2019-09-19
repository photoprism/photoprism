import Label from "model/label";
import MockAdapter from "axios-mock-adapter";
import Api from "common/api";

let chai = require('../../../node_modules/chai/chai');
let assert = chai.assert;

const mock = new MockAdapter(Api);

mock
    .onPost().reply(200)
    .onDelete().reply(200);

describe("model/label", () => {
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
        assert.equal(result, "/api/v1/labels/black-cat/thumbnail/fit_720 720w, /api/v1/labels/black-cat/thumbnail/fit_1280 1280w, /api/v1/labels/black-cat/thumbnail/fit_1920 1920w, /api/v1/labels/black-cat/thumbnail/fit_2560 2560w, /api/v1/labels/black-cat/thumbnail/fit_3840 3840w");
    });

    it("should get thumbnail sizes",  () => {
        const values = {id: 5, LabelName: "Black Cat", LabelSlug: "black-cat"};
        const label = new Label(values);
        const result = label.getThumbnailSizes();
        assert.equal(result, "(min-width: 2560px) 3840px, (min-width: 1920px) 2560px, (min-width: 1280px) 1920px, (min-width: 720px) 1280px, 720px");
    });

    it("should get date string",  () => {
        const values = {ID: 5, LabelName: "Black Cat", LabelSlug: "black-cat", CreatedAt: "2012-07-08T14:45:39Z"};
        const label = new Label(values);
        const result = label.getDateString();
        assert.equal(result, "Jul 8, 2012, 2:45 PM");
    });

    it("should get model name",  () => {
        const result = Label.getModelName();
        assert.equal(result, "Label");
    });

    it("should get collection resource",  () => {
        const result = Label.getCollectionResource();
        assert.equal(result, "labels");
    });

    it("should like label",  () => {
        const values = {ID: 5, LabelName: "Black Cat", LabelSlug: "black-cat", LabelFavorite: false};
        const label = new Label(values);
        assert.equal(label.LabelFavorite, false);
        label.like();
        assert.equal(label.LabelFavorite, true);
    });

    it("should unlike label",  () => {
        const values = {ID: 5, LabelName: "Black Cat", LabelSlug: "black-cat", LabelFavorite: true};
        const label = new Label(values);
        assert.equal(label.LabelFavorite, true);
        label.unlike();
        assert.equal(label.LabelFavorite, false);
    });

    it("should toggle like",  () => {
        const values = {ID: 5, LabelName: "Black Cat", LabelSlug: "black-cat", LabelFavorite: true};
        const label = new Label(values);
        assert.equal(label.LabelFavorite, true);
        label.toggleLike();
        assert.equal(label.LabelFavorite, false);
        label.toggleLike();
        assert.equal(label.LabelFavorite, true);
    });
});
