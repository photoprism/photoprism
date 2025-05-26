import { describe, it, expect } from "vitest";
import "../fixtures";
import { Album, BatchSize } from "model/album";

describe("model/album", () => {
  it("should get route view", () => {
    const values = { ID: 5, Title: "Christmas 2019", Slug: "christmas-2019" };
    const album = new Album(values);
    const result = album.route("test");
    expect(result.name).toBe("test");
    expect(result.params.slug).toBe("view");
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
    expect(result).toContain("is-album");
    expect(result).toContain("uid-5");
    expect(result).toContain("type-moment");
    expect(result).toContain("is-selected");
    expect(result).toContain("is-favorite");
    expect(result).toContain("is-private");
  });

  it("should get album entity name", () => {
    const values = { ID: 5, Title: "Christmas 2019", Slug: "christmas-2019" };
    const album = new Album(values);
    const result = album.getEntityName();
    expect(result).toBe("christmas-2019");
  });

  it("should get album id", () => {
    const values = { ID: 5, Title: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values);
    const result = album.getId();
    expect(result).toBe(66);
  });

  it("should get album title", () => {
    const values = { ID: 5, Title: "Christmas 2019", Slug: "christmas-2019" };
    const album = new Album(values);
    const result = album.getTitle();
    expect(result).toBe("Christmas 2019");
  });

  it("should get album country", () => {
    const values = { ID: 5, Title: "Christmas 2019", Slug: "christmas-2019", Country: "at" };
    const album = new Album(values);
    const result = album.getCountry();
    expect(result).toBe("Austria");

    const values2 = { ID: 5, Title: "Christmas 2019", Slug: "christmas-2019", Country: "zz" };
    const album2 = new Album(values2);
    const result2 = album2.getCountry();
    expect(result2).toBe("");

    const values3 = { ID: 5, Title: "Christmas 2019", Slug: "christmas-2019", Country: "xx" };
    const album3 = new Album(values3);
    const result3 = album3.getCountry();
    expect(result3).toBe("");
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
    expect(result).toBe(false);

    const values2 = { ID: 5, Title: "Christmas 2019", Slug: "christmas-2019", Country: "at" };
    const album2 = new Album(values2);
    const result2 = album2.hasLocation();
    expect(result2).toBe(true);
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
    expect(result).toBe("Salzburg, Austria");

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
    expect(result2).toBe("");

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
    expect(result3).toBe("Austria");

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
    expect(result5).toBe("Austria");

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
    expect(result6).toBe("Salzburg");
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
    expect(result).toBe("/api/v1/t/d6b24d688564f7ddc7b245a414f003a8d8ff5a67/public/xyz");

    const values2 = {
      ID: 5,
      Title: "Christmas 2019",
      Slug: "christmas-2019",
      UID: 66,
    };
    const album2 = new Album(values2);
    const result2 = album2.thumbnailUrl("xyz");
    expect(result2).toBe("/api/v1/albums/66/t/public/xyz");

    const values3 = {
      ID: 5,
      Title: "Christmas 2019",
      Slug: "christmas-2019",
    };
    const album3 = new Album(values3);
    const result3 = album3.thumbnailUrl("xyz");
    expect(result3).toBe("/api/v1/svg/album");
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
    expect(result.replaceAll("\u202f", " ")).toBe("Jul 8, 2012, 2:45 PM");
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
    expect(result).toBe("May 2019");
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
    expect(result).toBe("2000");
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
    expect(result).toBe("Unknown");
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
    expect(result).toBe("Monday, May 1, 2000");
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
    expect(result).toBe("08");
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
    expect(result).toBe("01");
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
    expect(result).toBe(new Date().getFullYear().toString().padStart(4, "0"));
  });

  it("should get model name", () => {
    const result = Album.getModelName();
    expect(result).toBe("Album");
  });

  it("should get collection resource", () => {
    const result = Album.getCollectionResource();
    expect(result).toBe("albums");
  });

  it("should return batch size", () => {
    expect(Album.batchSize()).toBe(BatchSize);
    Album.setBatchSize(30);
    expect(Album.batchSize()).toBe(30);
    Album.setBatchSize(BatchSize);
  });

  it("should like album", () => {
    const values = { ID: 5, Title: "Christmas 2019", Slug: "christmas-2019", Favorite: false };
    const album = new Album(values);
    expect(album.Favorite).toBe(false);
    album.like();
    expect(album.Favorite).toBe(true);
  });

  it("should unlike album", () => {
    const values = { ID: 5, Title: "Christmas 2019", Slug: "christmas-2019", Favorite: true };
    const album = new Album(values);
    expect(album.Favorite).toBe(true);
    album.unlike();
    expect(album.Favorite).toBe(false);
  });

  it("should toggle like", () => {
    const values = { ID: 5, Title: "Christmas 2019", Slug: "christmas-2019", Favorite: true };
    const album = new Album(values);
    expect(album.Favorite).toBe(true);
    album.toggleLike();
    expect(album.Favorite).toBe(false);
    album.toggleLike();
    expect(album.Favorite).toBe(true);
  });
});
