import assert from "assert";
import Photo from "model/photo";

describe.only("model/photo", () => {
    it("should get photo entity name",  () => {
        const values = {id: 5, PhotoTitle: "Crazy Cat"};
        const photo = new Photo(values);
        const result = photo.getEntityName();
        assert.equal(result, "Crazy Cat");
    });

    it("should get photo id",  () => {
        const values = {ID: 5, PhotoTitle: "Crazy Cat", PhotoUUID: 789};
        const photo = new Photo(values);
        const result = photo.getId();
        assert.equal(result, 5);
    });

    it("should get photo title",  () => {
        const values = {ID: 5, PhotoTitle: "Crazy Cat", PhotoUUID: 789};
        const photo = new Photo(values);
        const result = photo.getTitle();
        assert.equal(result, "Crazy Cat");
    });

    it("should get photo color brown",  () => {
        const values = {ID: 5, PhotoTitle: "Crazy Cat", PhotoColor: "brown"};
        const photo = new Photo(values);
        const result = photo.getColor();
        assert.equal(result, "grey lighten-2");
    });

    it("should get photo color grey",  () => {
        const values = {ID: 5, PhotoTitle: "Crazy Cat", PhotoColor: "grey"};
        const photo = new Photo(values);
        const result = photo.getColor();
        assert.equal(result, "grey lighten-2");
    });

    it("should get photo color pink",  () => {
        const values = {ID: 5, PhotoTitle: "Crazy Cat", PhotoColor: "pink"};
        const photo = new Photo(values);
        const result = photo.getColor();
        assert.equal(result, "pink lighten-4");
    });

    it("should get photo maps link",  () => {
        const values = {ID: 5, PhotoTitle: "Crazy Cat", PhotoLat: 36.442881666666665, PhotoLong: 28.229493333333334};
        const photo = new Photo(values);
        const result = photo.getGoogleMapsLink();
        assert.equal(result, "https://www.google.com/maps/place/36.442881666666665,28.229493333333334");
    });


});