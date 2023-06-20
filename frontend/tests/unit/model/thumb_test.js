import "../fixtures";
import Thumb from "model/thumb";
import Photo from "model/photo";
import File from "model/file";

let chai = require("chai/chai");
let assert = chai.assert;

describe("model/thumb", () => {
  it("should get thumb defaults", () => {
    const values = {
      UID: "55",
      Title: "",
      TakenAtLocal: "",
      Description: "",
      Favorite: false,
      Playable: false,
      Width: 0,
      Height: 0,
      DownloadUrl: "",
    };
    const thumb = new Thumb(values);
    const result = thumb.getDefaults();
    assert.equal(result.UID, "");
  });

  it("should get id", () => {
    const values = {
      UID: "55",
    };
    const thumb = new Thumb(values);
    assert.equal(thumb.getId(), "55");
  });

  it("should return hasId", () => {
    const values = {
      UID: "55",
    };
    const thumb = new Thumb(values);
    assert.equal(thumb.hasId(), true);

    const values2 = {
      Title: "",
    };
    const thumb2 = new Thumb(values2);
    assert.equal(thumb2.hasId(), false);
  });

  it("should toggle like", () => {
    const values = {
      UID: "55",
      Title: "",
      TakenAtLocal: "",
      Description: "",
      Favorite: true,
      Playable: false,
      Width: 0,
      Height: 0,
      DownloadUrl: "",
    };
    const thumb = new Thumb(values);
    assert.equal(thumb.Favorite, true);
    thumb.toggleLike();
    assert.equal(thumb.Favorite, false);
    thumb.toggleLike();
    assert.equal(thumb.Favorite, true);
  });

  it("should return thumb not found", () => {
    const result = Thumb.notFound();
    assert.equal(result.UID, "");
    assert.equal(result.Favorite, false);
  });

  it("should test from file", () => {
    const values = {
      InstanceID: 5,
      UID: "ABC123",
      Name: "1/2/IMG123.jpg",
      Hash: "abc123",
      Width: 500,
      Height: 900,
    };
    const file = new File(values);

    const values2 = {
      UID: "5",
      Title: "Crazy Cat",
      TakenAt: "2012-07-08T14:45:39Z",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      Description: "Nice description",
      Favorite: true,
    };
    const photo = new Photo(values2);
    const result = Thumb.fromFile(photo, file);
    assert.equal(result.UID, "5");
    assert.equal(result.Description, "Nice description");
    assert.equal(result.Width, 500);
    const result2 = Thumb.fromFile();
    assert.equal(result2.UID, "");
  });

  it("should test from files", () => {
    const values2 = {
      UID: "5",
      Title: "Crazy Cat",
      TakenAt: "2012-07-08T14:45:39Z",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      Description: "Nice description",
      Favorite: true,
    };
    const photo = new Photo(values2);

    const values3 = {
      UID: "5",
      Title: "Crazy Cat",
      TakenAt: "2012-07-08T14:45:39Z",
      TakenAtLocal: "2012-07-08T14:45:39Z",
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
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.jpg",
          Primary: true,
          FileType: "jpg",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt53",
        },
      ],
    };
    const photo3 = new Photo(values4);
    const Photos2 = [photo, photo2, photo3];
    const result2 = Thumb.fromFiles(Photos2);
    assert.equal(result2[0].UID, "ABC123");
    assert.equal(result2[0].Description, "Nice description 2");
    assert.equal(result2[0].Width, 500);
    assert.equal(result2.length, 1);
    const values5 = {
      ID: 8,
      UID: "ABC123",
      Description: "Nice description 2",
      Hash: "abc345",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.jpg",
          Primary: true,
          FileType: "mov",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt53",
        },
      ],
    };
    const photo4 = new Photo(values5);
    const Photos3 = [photo3, photo2, photo4];
    const result3 = Thumb.fromFiles(Photos3);
    assert.equal(result3.length, 1);
    assert.equal(result3[0].UID, "ABC123");
    assert.equal(result3[0].Description, "Nice description 2");
    assert.equal(result3[0].Width, 500);
  });

  it("should test from files", () => {
    const Photos = [];
    const result = Thumb.fromFiles(Photos);
    assert.equal(result, "");
  });

  it("should test from photo", () => {
    const values = {
      ID: 8,
      UID: "ABC123",
      Description: "Nice description 3",
      Hash: "345ggh",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.jpg",
          Primary: true,
          FileType: "jpg",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt53",
        },
      ],
    };
    const photo = new Photo(values);
    const result = Thumb.fromPhoto(photo);
    assert.equal(result.UID, "ABC123");
    assert.equal(result.Description, "Nice description 3");
    assert.equal(result.Width, 500);
    const values3 = {
      ID: 8,
      UID: "ABC124",
      Description: "Nice description 3",
    };
    const photo3 = new Photo(values3);
    const result2 = Thumb.fromPhoto(photo3);
    assert.equal(result2.UID, "");
    const values2 = {
      ID: 8,
      UID: "ABC123",
      Title: "Crazy Cat",
      TakenAt: "2012-07-08T14:45:39Z",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      Description: "Nice description",
      Favorite: true,
      Hash: "xdf45m",
    };
    const photo2 = new Photo(values2);
    const result3 = Thumb.fromPhoto(photo2);
    assert.equal(result3.UID, "ABC123");
    assert.equal(result3.Title, "Crazy Cat");
    assert.equal(result3.Description, "Nice description");
  });

  it("should test from photos", () => {
    const values = {
      ID: 8,
      UID: "ABC123",
      Description: "Nice description 3",
      Hash: "345ggh",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.jpg",
          Primary: true,
          FileType: "jpg",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt53",
        },
      ],
    };
    const photo = new Photo(values);
    const Photos = [photo];
    const result = Thumb.fromPhotos(Photos);
    assert.equal(result[0].UID, "ABC123");
    assert.equal(result[0].Description, "Nice description 3");
    assert.equal(result[0].Width, 500);
  });

  it("should return download url", () => {
    const values = {
      InstanceID: 5,
      UID: "ABC123",
      Hash: "54ghtfd",
      FileType: "jpg",
      Name: "1/2/IMG123.jpg",
    };
    const file = new File(values);
    assert.equal(Thumb.downloadUrl(file), "/api/v1/dl/54ghtfd?t=2lbh9x09");
    const values2 = {
      InstanceID: 5,
      UID: "ABC123",
      FileType: "jpg",
      Name: "1/2/IMG123.jpg",
    };
    const file2 = new File(values2);
    assert.equal(Thumb.downloadUrl(file2), "");
  });

  it("should return thumbnail url", () => {
    const values = {
      InstanceID: 5,
      UID: "ABC123",
      Hash: "54ghtfd",
      FileType: "jpg",
      Name: "1/2/IMG123.jpg",
    };
    const file = new File(values);
    assert.equal(Thumb.thumbnailUrl(file, "abc"), "/api/v1/t/54ghtfd/public/abc");
    const values2 = {
      InstanceID: 5,
      UID: "ABC123",
      Name: "1/2/IMG123.jpg",
    };
    const file2 = new File(values2);
    assert.equal(Thumb.thumbnailUrl(file2, "bcd"), "/static/img/404.jpg");
  });

  it("should calculate size", () => {
    const values = {
      InstanceID: 5,
      UID: "ABC123",
      Hash: "54ghtfd",
      FileType: "jpg",
      Width: 900,
      Height: 850,
      Name: "1/2/IMG123.jpg",
    };
    const file = new File(values);
    const result = Thumb.calculateSize(file, 600, 800); //max 0,75
    assert.equal(result.width, 600);
    assert.equal(result.height, 567);
    const values3 = {
      InstanceID: 5,
      UID: "ABC123",
      Hash: "54ghtfd",
      FileType: "jpg",
      Width: 750,
      Height: 850,
      Name: "1/2/IMG123.jpg",
    };
    const file3 = new File(values3);
    const result2 = Thumb.calculateSize(file3, 900, 450);
    assert.equal(result2.width, 397);
    assert.equal(result2.height, 450);
    const result4 = Thumb.calculateSize(file3, 900, 950);
    assert.equal(result4.width, 750);
    assert.equal(result4.height, 850);
  });
});
