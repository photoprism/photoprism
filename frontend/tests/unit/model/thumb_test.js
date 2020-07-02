import Thumb from "model/thumb";
import MockAdapter from "axios-mock-adapter";
import Api from "common/api";
import Photo from "model/photo";
import File from "model/file";

let chai = require("chai/chai");
let assert = chai.assert;

const mock = new MockAdapter(Api);

mock
    .onGet("api/v1/settings").reply(200, {"download": true, "language": "de"})
    .onPost("api/v1/settings").reply(200, {"download": true, "language": "en"});


describe("model/thumb", () => {

    it("should get thumb defaults",  () => {
        const values = {
            uid: "55",
            title: "",
            taken: "",
            description: "",
            favorite: false,
            playable: false,
            original_w: 0,
            original_h: 0,
            download_url: "",
        };
        const thumb = new Thumb(values);
        const result = thumb.getDefaults();
        assert.equal(result.uid, "");
    });

    it("should toggle like",  () => {
        const values = {
            uid: "55",
            title: "",
            taken: "",
            description: "",
            favorite: true,
            playable: false,
            original_w: 0,
            original_h: 0,
            download_url: "",
        };
        const thumb = new Thumb(values);
        assert.equal(thumb.favorite, true);
        thumb.toggleLike();
        assert.equal(thumb.favorite, false);
        thumb.toggleLike();
        assert.equal(thumb.favorite, true);
    });

    it("should return thumb not found",  () => {
        const result = Thumb.thumbNotFound();
        assert.equal(result.uid, "");
        assert.equal(result.favorite, false);

    });

    it("should test from file",  () => {
        const values = {
            InstanceID: 5,
            UID: "ABC123",
            Name: "1/2/IMG123.jpg",
            Hash: "abc123",
            Width: 500,
            Height: 900};
        const file = new File(values);

        const values2 = {
            UID: "5",
            Title: "Crazy Cat",
            TakenAt: "2012-07-08T14:45:39Z",
            Description: "Nice description",
            Favorite: true,
        };
        const photo = new Photo(values2);
        const result = Thumb.fromFile(photo, file);
        assert.equal(result.uid, "5");
        assert.equal(result.description, "Nice description");
        assert.equal(result.original_w, 500);
        const result2 = Thumb.fromFile();
        assert.equal(result2.uid, "");
    });

    it("should test from files",  () => {
        const values2 = {
            UID: "5",
            Title: "Crazy Cat",
            TakenAt: "2012-07-08T14:45:39Z",
            Description: "Nice description",
            Favorite: true,
        };
        const photo = new Photo(values2);

        const values3 = {
            UID: "5",
            Title: "Crazy Cat",
            TakenAt: "2012-07-08T14:45:39Z",
            Description: "Nice description",
            Favorite: true,
        };
        const photo2 = new Photo(values3);
        const Photos = [photo, photo2];
        const result = Thumb.fromFiles(Photos);
        assert.equal(result.length, 0);
        const values4 = {
            ID: 8,
            UID: "ABC123",
            Description: "Nice description 2",
            Hash: "abc345",
            Files: [{
                UID: "123fgb",
                Name: "1980/01/superCuteKitten.jpg",
                Primary: true,
                Type: "jpg",
                Width: 500,
                Height: 600,
                Hash: "1xxbgdt53"}]};
        const photo3 = new Photo(values4);
        const Photos2 = [photo, photo2, photo3];
        const result2 = Thumb.fromFiles(Photos2);
        assert.equal(result2[0].uid, "ABC123");
        assert.equal(result2[0].description, "Nice description 2");
        assert.equal(result2[0].original_w, 500);
        assert.equal(result2.length, 1);

    });

    it("should test from photo",  () => {
        const values = {
            ID: 8,
            UID: "ABC123",
            Description: "Nice description 3",
            Hash: "345ggh",
            Files: [{
                UID: "123fgb",
                Name: "1980/01/superCuteKitten.jpg",
                Primary: true,
                Type: "jpg",
                Width: 500,
                Height: 600,
                Hash: "1xxbgdt53"}]};
        const photo = new Photo(values);
        const result = Thumb.fromPhoto(photo);
        assert.equal(result.uid, "ABC123");
        assert.equal(result.description, "Nice description 3");
        assert.equal(result.original_w, 500);
        const values3 = {
            ID: 8,
            UID: "ABC124",
            Description: "Nice description 3",
        };
        const photo3 = new Photo(values3);
        const result2 = Thumb.fromPhoto(photo3);
        assert.equal(result2.uid, "");
        const values2 = {
            ID: 8,
            UID: "ABC123",
            Title: "Crazy Cat",
            TakenAt: "2012-07-08T14:45:39Z",
            Description: "Nice description",
            Favorite: true,
            Hash: "xdf45m",

        };
        const photo2 = new Photo(values2);
        const result3 = Thumb.fromPhoto(photo2);
        assert.equal(result3.uid, "ABC123");
        assert.equal(result3.title, "Crazy Cat");
        assert.equal(result3.description, "Nice description");
    });

    it("should test from photos",  () => {
        const values = {
            ID: 8,
            UID: "ABC123",
            Description: "Nice description 3",
            Hash: "345ggh",
            Files: [{
                UID: "123fgb",
                Name: "1980/01/superCuteKitten.jpg",
                Primary: true,
                Type: "jpg",
                Width: 500,
                Height: 600,
                Hash: "1xxbgdt53"}]};
        const photo = new Photo(values);
        const Photos = [photo];
        const result = Thumb.fromPhotos(Photos);
        assert.equal(result[0].uid, "ABC123");
        assert.equal(result[0].description, "Nice description 3");
        assert.equal(result[0].original_w, 500);
    });
});