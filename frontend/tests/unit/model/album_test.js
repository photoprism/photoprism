import "../fixtures";
import Album from "model/album";

let chai = require("chai/chai");
let assert = chai.assert;

describe("model/album", () => {
  it("should get route view", () => {
    const values = { ID: 5, Title: "Christmas 2019", Slug: "christmas-2019" };
    const album = new Album(values);
    const result = album.route("test");
    assert.equal(result.name, "test");
    assert.equal(result.params.slug, "view");
  });

  it("should return classes", () => {
    const values = {
      UID: 5,
      Title: "Christmas 2019",
      Slug: "christmas-2019",
      Type: "moment",
      Favorite: true,
      Private: true,
    };
    const album = new Album(values);
    const result = album.classes(true);
    assert.include(result, "is-album");
    assert.include(result, "uid-5");
    assert.include(result, "type-moment");
    assert.include(result, "is-selected");
    assert.include(result, "is-favorite");
    assert.include(result, "is-private");
  });

  it("should get album entity name", () => {
    const values = { ID: 5, Title: "Christmas 2019", Slug: "christmas-2019" };
    const album = new Album(values);
    const result = album.getEntityName();
    assert.equal(result, "christmas-2019");
  });

  it("should get album id", () => {
    const values = { ID: 5, Title: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values);
    const result = album.getId();
    assert.equal(result, "66");
  });

  it("should get album title", () => {
    const values = { ID: 5, Title: "Christmas 2019", Slug: "christmas-2019" };
    const album = new Album(values);
    const result = album.getTitle();
    assert.equal(result, "Christmas 2019");
  });

  it("should get album country", () => {
    const values = { ID: 5, Title: "Christmas 2019", Slug: "christmas-2019", Country: "at" };
    const album = new Album(values);
    const result = album.getCountry();
    assert.equal(result, "Austria");

    const values2 = { ID: 5, Title: "Christmas 2019", Slug: "christmas-2019", Country: "zz" };
    const album2 = new Album(values2);
    const result2 = album2.getCountry();
    assert.equal(result2, "");

    const values3 = { ID: 5, Title: "Christmas 2019", Slug: "christmas-2019", Country: "xx" };
    const album3 = new Album(values3);
    const result3 = album3.getCountry();
    assert.equal(result3, "");
  });

  it("should check if album has location", () => {
    const values = {
      ID: 5,
      Title: "Christmas 2019",
      Slug: "christmas-2019",
      Country: "zz",
      State: "",
      Location: "",
    };
    const album = new Album(values);
    const result = album.hasLocation();
    assert.equal(result, false);

    const values2 = { ID: 5, Title: "Christmas 2019", Slug: "christmas-2019", Country: "at" };
    const album2 = new Album(values2);
    const result2 = album2.hasLocation();
    assert.equal(result2, true);
  });

  it("should get album location", () => {
    const values = {
      ID: 5,
      Title: "Christmas 2019",
      Slug: "christmas-2019",
      Country: "at",
      State: "Salzburg",
      Location: "",
    };
    const album = new Album(values);
    const result = album.getLocation();
    assert.equal(result, "Salzburg, Austria");

    const values2 = {
      ID: 5,
      Title: "Christmas 2019",
      Slug: "christmas-2019",
      Country: "zz",
      State: "",
      Location: "",
    };
    const album2 = new Album(values2);
    const result2 = album2.getLocation();
    assert.equal(result2, "");

    const values3 = {
      ID: 5,
      Title: "Christmas 2019",
      Slug: "christmas-2019",
      Country: "zz",
      State: "",
      Location: "Austria",
    };
    const album3 = new Album(values3);
    const result3 = album3.getLocation();
    assert.equal(result3, "Austria");

    const values5 = {
      ID: 5,
      Title: "Salzburg",
      Slug: "salzburg",
      Country: "at",
      State: "Salzburg",
      Location: "",
    };
    const album5 = new Album(values5);
    const result5 = album5.getLocation();
    assert.equal(result5, "Austria");

    const values6 = {
      ID: 5,
      Title: "Austria",
      Slug: "austria",
      Country: "at",
      State: "Salzburg",
      Location: "",
    };
    const album6 = new Album(values6);
    const result6 = album6.getLocation();
    assert.equal(result6, "Salzburg");
  });

  it("should get thumbnail url", () => {
    const values = {
      ID: 5,
      Thumb: "d6b24d688564f7ddc7b245a414f003a8d8ff5a67",
      Title: "Christmas 2019",
      Slug: "christmas-2019",
      UID: 66,
    };
    const album = new Album(values);
    const result = album.thumbnailUrl("xyz");
    assert.equal(result, "/api/v1/t/d6b24d688564f7ddc7b245a414f003a8d8ff5a67/public/xyz");

    const values2 = {
      ID: 5,
      Title: "Christmas 2019",
      Slug: "christmas-2019",
      UID: 66,
    };
    const album2 = new Album(values2);
    const result2 = album2.thumbnailUrl("xyz");
    assert.equal(result2, "/api/v1/albums/66/t/public/xyz");

    const values3 = {
      ID: 5,
      Title: "Christmas 2019",
      Slug: "christmas-2019",
    };
    const album3 = new Album(values3);
    const result3 = album3.thumbnailUrl("xyz");
    assert.equal(result3, "/api/v1/svg/album");
  });

  it("should get created date string", () => {
    const values = {
      ID: 5,
      Title: "Christmas 2019",
      Slug: "christmas-2019",
      CreatedAt: "2012-07-08T14:45:39Z",
    };
    const album = new Album(values);
    const result = album.getCreatedString();
    assert.equal(result.replaceAll("\u202f", " "), "Jul 8, 2012, 2:45 PM");
  });

  it("should get album date string with invalid day", () => {
    const values = {
      ID: 5,
      Title: "Christmas 2019",
      Slug: "christmas-2019",
      CreatedAt: "2012-07-08T14:45:39Z",
      Day: -1,
      Month: 5,
      Year: 2019,
    };
    const album = new Album(values);
    const result = album.getDateString();
    assert.equal(result, "May 2019");
  });

  it("should get album date string with invalid month", () => {
    const values = {
      ID: 5,
      Title: "Christmas 2019",
      Slug: "christmas-2019",
      CreatedAt: "2012-07-08T14:45:39Z",
      Day: 1,
      Month: -5,
      Year: 2000,
    };
    const album = new Album(values);
    const result = album.getDateString();
    assert.equal(result, "2000");
  });

  it("should get album date string with invalid year", () => {
    const values = {
      ID: 5,
      Title: "Christmas 2019",
      Slug: "christmas-2019",
      CreatedAt: "2012-07-08T14:45:39Z",
      Day: 1,
      Month: 5,
      Year: 800,
    };
    const album = new Album(values);
    const result = album.getDateString();
    assert.equal(result, "Unknown");
  });

  it("should get album date string", () => {
    const values = {
      ID: 5,
      Title: "Christmas 2019",
      Slug: "christmas-2019",
      CreatedAt: "2012-07-08T14:45:39Z",
      Day: 1,
      Month: 5,
      Year: 2000,
    };
    const album = new Album(values);
    const result = album.getDateString();
    assert.equal(result, "Monday, May 1, 2000");
  });

  it("should get day string", () => {
    const values = {
      ID: 5,
      Title: "Christmas 2019",
      Slug: "christmas-2019",
      CreatedAt: "2012-07-08T14:45:39Z",
      Day: 8,
      Month: 5,
      Year: 2019,
    };
    const album = new Album(values);
    const result = album.dayString();
    assert.equal(result, "08");
  });

  it("should get month string", () => {
    const values = {
      ID: 5,
      Title: "Christmas 2019",
      Slug: "christmas-2019",
      CreatedAt: "2012-07-08T14:45:39Z",
      Day: 8,
      Month: -5,
      Year: 2019,
    };
    const album = new Album(values);
    const result = album.monthString();
    assert.equal(result, "01");
  });

  it("should get year string", () => {
    const values = {
      ID: 5,
      Title: "Christmas 2019",
      Slug: "christmas-2019",
      CreatedAt: "2012-07-08T14:45:39Z",
      Day: 8,
      Month: -5,
      Year: 800,
    };
    const album = new Album(values);
    const result = album.yearString();
    assert.equal(result, new Date().getFullYear().toString().padStart(4, "0"));
  });

  it("should get model name", () => {
    const result = Album.getModelName();
    assert.equal(result, "Album");
  });

  it("should get collection resource", () => {
    const result = Album.getCollectionResource();
    assert.equal(result, "albums");
  });

  it("should return batch size", () => {
    assert.equal(Album.batchSize(), 24);
    Album.setBatchSize(30);
    assert.equal(Album.batchSize(), 30);
    Album.setBatchSize(24);
  });

  it("should like album", () => {
    const values = { ID: 5, Title: "Christmas 2019", Slug: "christmas-2019", Favorite: false };
    const album = new Album(values);
    assert.equal(album.Favorite, false);
    album.like();
    assert.equal(album.Favorite, true);
  });

  it("should unlike album", () => {
    const values = { ID: 5, Title: "Christmas 2019", Slug: "christmas-2019", Favorite: true };
    const album = new Album(values);
    assert.equal(album.Favorite, true);
    album.unlike();
    assert.equal(album.Favorite, false);
  });

  it("should toggle like", () => {
    const values = { ID: 5, Title: "Christmas 2019", Slug: "christmas-2019", Favorite: true };
    const album = new Album(values);
    assert.equal(album.Favorite, true);
    album.toggleLike();
    assert.equal(album.Favorite, false);
    album.toggleLike();
    assert.equal(album.Favorite, true);
  });
});
