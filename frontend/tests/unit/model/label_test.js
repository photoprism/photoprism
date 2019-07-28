import assert from "assert";
import Label from "model/label";
import Api from "common/api";
import MockAdapter from "axios-mock-adapter";

const mock = new MockAdapter(Api);

const postLikeEntity = [
    {id: 5, LabelFavorite: true},
];
mock.onPost("foo").reply(201, postLikeEntity);

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

    it("should get model name",  () => {
        const result = Label.getModelName();
        assert.equal(result, "Label");
    });

    it("should get collection resource",  () => {
        const result = Label.getCollectionResource();
        assert.equal(result, "labels");
    });

    it("should toggle like",  () => {
        Api.post('foo', postLikeEntity).then(
            (response) => {
                console.log(response);
                assert.equal(201, response.status);
                assert.deepEqual(postLikeEntity, response.data);
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });


});