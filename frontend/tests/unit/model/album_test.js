import assert from "assert";
import Album from "model/album";
import Api from "common/api";
import MockAdapter from "axios-mock-adapter";

const mock = new MockAdapter(Api);

const postLikeEntity = [
    {id: 5, AlbumFavorite: true},
];
mock.onPost("foo").reply(201, postLikeEntity);

describe("model/album", () => {
    it("should get album entity name",  () => {
        const values = {id: 5, AlbumName: "Christmas 2019", AlbumSlug: "christmas-2019"};
        const album = new Album(values);
        const result = album.getEntityName();
        assert.equal(result, "christmas-2019");
    });

    it("should get album id",  () => {
        const values = {id: 5, AlbumName: "Christmas 2019", AlbumSlug: "christmas-2019", AlbumUUID: 66};
        const album = new Album(values);
        const result = album.getId();
        assert.equal(result, "66");
    });

    it("should get album title",  () => {
        const values = {id: 5, AlbumName: "Christmas 2019", AlbumSlug: "christmas-2019"};
        const album = new Album(values);
        const result = album.getTitle();
        assert.equal(result, "Christmas 2019");
    });

    it("should get thumbnail url",  () => {
        const values = {id: 5, AlbumName: "Christmas 2019", AlbumSlug: "christmas-2019", AlbumUUID: 66};
        const album = new Album(values);
        const result = album.getThumbnailUrl("xyz");
        assert.equal(result, "/api/v1/albums/66/thumbnail/xyz");
    });

    it("should get thumbnail src set",  () => {
        const values = {id: 5, AlbumName: "Christmas 2019", AlbumSlug: "christmas-2019", AlbumUUID: 66};
        const album = new Album(values);
        const result = album.getThumbnailSrcset("");
        assert.equal(result, "/api/v1/albums/66/thumbnail/fit_720 720w, /api/v1/albums/66/thumbnail/fit_1280 1280w, /api/v1/albums/66/thumbnail/fit_1920 1920w, /api/v1/albums/66/thumbnail/fit_2560 2560w, /api/v1/albums/66/thumbnail/fit_3840 3840w");
    });

    it("should get thumbnail sizes",  () => {
        const values = {id: 5, AlbumName: "Christmas 2019", AlbumSlug: "christmas-2019"};
        const album = new Album(values);
        const result = album.getThumbnailSizes();
        assert.equal(result, "(min-width: 2560px) 3840px, (min-width: 1920px) 2560px, (min-width: 1280px) 1920px, (min-width: 720px) 1280px, 720px");
    });

    it("should get model name",  () => {
        const result = Album.getModelName();
        assert.equal(result, "Album");
    });

    it("should get collection resource",  () => {
        const result = Album.getCollectionResource();
        assert.equal(result, "albums");
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