import "../fixtures";
import { Photo, FormatJpeg } from "model/photo";

let chai = require("chai/chai");
let assert = chai.assert;

describe("model/photo", () => {
  it("should get photo entity name", () => {
    const values = { UID: 5, Title: "Crazy Cat" };
    const photo = new Photo(values);
    const result = photo.getEntityName();
    assert.equal(result, "Crazy Cat");
  });

  it("should get photo uuid", () => {
    const values = { ID: 5, Title: "Crazy Cat", UID: 789 };
    const photo = new Photo(values);
    const result = photo.getId();
    assert.equal(result, 789);
  });

  it("should get photo title", () => {
    const values = { ID: 5, Title: "Crazy Cat", UID: 789 };
    const photo = new Photo(values);
    const result = photo.getTitle();
    assert.equal(result, "Crazy Cat");
  });

  it("should get photo maps link", () => {
    const values = { ID: 5, Title: "Crazy Cat", Lat: 36.442881666666665, Lng: 28.229493333333334 };
    const photo = new Photo(values);
    const result = photo.getGoogleMapsLink();
    assert.equal(result, "https://www.google.com/maps/place/36.442881666666665,28.229493333333334");
  });

  it("should get photo thumbnail url", () => {
    const values = { ID: 5, Title: "Crazy Cat", Hash: 345982 };
    const photo = new Photo(values);
    const result = photo.thumbnailUrl("tile500");
    assert.equal(result, "/api/v1/t/345982/public/tile500");
    const values2 = {
      ID: 10,
      UID: "ABC127",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.mp4",
          Primary: false,
          FileType: "mp4",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt55",
        },
      ],
    };
    const photo2 = new Photo(values2);
    const result2 = photo2.thumbnailUrl("tile500");
    assert.equal(result2, "/api/v1/t/1xxbgdt55/public/tile500");
    const values3 = {
      ID: 10,
      UID: "ABC127",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.mp4",
          Primary: false,
          FileType: "mp4",
          Width: 500,
          Height: 600,
        },
      ],
    };
    const photo3 = new Photo(values3);
    const result3 = photo3.thumbnailUrl("tile500");
    assert.equal(result3, "/static/img/404.jpg");
  });

  it("should get classes", () => {
    const values2 = {
      ID: 10,
      UID: "ABC127",
      Type: "video",
      Portrait: true,
      Favorite: true,
      Private: true,
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.mp4",
          Primary: false,
          FileType: "mp4",
          Video: true,
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt55",
        },
        {
          UID: "123fde",
          Name: "1980/01/superCuteKitten.mp4",
          Primary: false,
          FileType: "mp4",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdkkk",
        },
      ],
    };
    const photo2 = new Photo(values2);
    const result2 = photo2.classes();
    assert.include(result2, "is-photo");
    assert.include(result2, "uid-ABC127");
    assert.include(result2, "type-video");
    assert.include(result2, "is-portrait");
    assert.include(result2, "is-favorite");
    assert.include(result2, "is-private");
    assert.include(result2, "is-stack");
    assert.include(result2, "is-playable");
  });

  it("should get photo download url", () => {
    const values = { ID: 5, Title: "Crazy Cat", Hash: 345982 };
    const photo = new Photo(values);
    const result = photo.getDownloadUrl();
    assert.equal(result, "/api/v1/dl/345982?t=2lbh9x09");
  });

  it("should calculate photo size", () => {
    const values = { ID: 5, Title: "Crazy Cat", Width: 500, Height: 200 };
    const photo = new Photo(values);
    const result = photo.calculateSize(500, 200);
    assert.equal(result.width, 500);
    assert.equal(result.height, 200);
  });

  it("should calculate photo size with srcAspectRatio < maxAspectRatio", () => {
    const values = { ID: 5, Title: "Crazy Cat", Width: 500, Height: 200 };
    const photo = new Photo(values);
    const result = photo.calculateSize(300, 50);
    assert.equal(result.width, 125);
    assert.equal(result.height, 50);
  });

  it("should calculate photo size with srcAspectRatio > maxAspectRatio", () => {
    const values = { ID: 5, Title: "Crazy Cat", Width: 500, Height: 200 };
    const photo = new Photo(values);
    const result = photo.calculateSize(400, 300);
    assert.equal(result.width, 400);
    assert.equal(result.height, 160);
  });

  it("should get local day string", () => {
    const values = { ID: 5, Title: "Crazy Cat", TakenAt: "2012-07-08T14:45:39Z", TimeZone: "UTC" };
    const photo = new Photo(values);
    const result = photo.localDayString();
    // Current day of the month (changes):
    assert.equal(result.length, 2);
    const values2 = {
      ID: 5,
      Title: "Crazy Cat",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      TakenAt: "2012-07-08T14:45:39Z",
      TimeZone: "UTC",
      Day: 8,
    };
    const photo2 = new Photo(values2);
    const result2 = photo2.localDayString();
    assert.equal(result2, "08");
  });

  it("should get local month string", () => {
    const values = { ID: 5, Title: "Crazy Cat", TakenAt: "2012-07-08T14:45:39Z", TimeZone: "UTC" };
    const photo = new Photo(values);
    const result = photo.localMonthString();
    assert.equal(result, (new Date().getMonth() + 1).toString().padStart(2, "0"));
    const values2 = {
      ID: 5,
      Title: "Crazy Cat",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      TakenAt: "2012-07-08T14:45:39Z",
      TimeZone: "UTC",
      Month: 8,
    };
    const photo2 = new Photo(values2);
    const result2 = photo2.localMonthString();
    assert.equal(result2, "08");
  });

  it("should get local year string", () => {
    const values = { ID: 5, Title: "Crazy Cat", TakenAt: "2012-07-08T14:45:39Z", TimeZone: "UTC" };
    const photo = new Photo(values);
    const result = photo.localYearString();
    assert.equal(result, new Date().getFullYear().toString().padStart(4, "0"));
    const values2 = {
      ID: 5,
      Title: "Crazy Cat",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      TakenAt: "2012-07-08T14:45:39Z",
      TimeZone: "UTC",
      Year: 2010,
    };
    const photo2 = new Photo(values2);
    const result2 = photo2.localYearString();
    assert.equal(result2, "2010");
  });

  it("should get local date string", () => {
    const values = {
      ID: 5,
      Title: "Crazy Cat",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      TakenAt: "2012-07-08T14:45:39Z",
      TimeZone: "UTC",
    };
    const photo = new Photo(values);
    const result = photo.localDateString();
    assert.equal(result, "2012-07-08T14:45:39Z");
  });

  it("should get local date", () => {
    const values = {
      ID: 5,
      Title: "Crazy Cat",
      TakenAt: "2012-07-08T14:45:39Z",
      TimeZone: "Indian/Reunion",
    };
    const photo = new Photo(values);
    const result = photo.localDate();
    assert.equal(String(result), "2012-07-08T14:45:39.000Z");
  });

  it("UTC", () => {
    const values = {
      ID: 9999,
      Title: "Video",
      TakenAt: "2012-07-08T14:45:39Z",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      TimeZone: "UTC",
    };
    const photo = new Photo(values);
    assert.equal(String(photo.localDateString("10:00:00")), "2012-07-08T10:00:00Z");
    const result = photo.localDate();
    assert.equal(String(result), "2012-07-08T14:45:39.000Z");
  });

  it("should get date string", () => {
    const values = {
      ID: 5,
      Title: "Crazy Cat",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      TakenAt: "2012-07-08T14:45:39Z",
      TimeZone: "UTC",
    };
    const photo = new Photo(values);
    const result = photo.getDateString().replaceAll("\u202f", " ");
    assert.isTrue(result.startsWith("Sunday, July 8, 2012"));
    assert.isTrue(result.endsWith("2:45 PM"));
    const values2 = { ID: 5, Title: "Crazy Cat", TakenAtLocal: "", TakenAt: "", TimeZone: "UTC" };
    const photo2 = new Photo(values2);
    const result2 = photo2.getDateString();
    assert.equal(result2, "Unknown");
    const values3 = {
      ID: 5,
      Title: "Crazy Cat",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      TakenAt: "2012-07-08T14:45:39Z",
    };
    const photo3 = new Photo(values3);
    const result3 = photo3.getDateString();
    assert.equal(result3, "Sunday, July 8, 2012");
    const values4 = {
      ID: 5,
      Title: "Crazy Cat",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      TakenAt: "2012-07-08T14:45:39Z",
      Month: -1,
    };
    const photo4 = new Photo(values4);
    const result4 = photo4.getDateString();
    assert.equal(result4, "2012");
    const values5 = {
      ID: 5,
      Title: "Crazy Cat",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      TakenAt: "2012-07-08T14:45:39Z",
      Day: -1,
    };
    const photo5 = new Photo(values5);
    const result5 = photo5.getDateString();
    assert.equal(result5, "July 2012");
  });

  it("should get short date string", () => {
    const values = {
      ID: 5,
      Title: "Crazy Cat",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      TakenAt: "2012-07-08T14:45:39Z",
      TimeZone: "UTC",
    };
    const photo = new Photo(values);
    const result = photo.shortDateString();
    assert.equal(result, "Jul 8, 2012");
    const values2 = { ID: 5, Title: "Crazy Cat", TakenAtLocal: "", TakenAt: "", TimeZone: "UTC" };
    const photo2 = new Photo(values2);
    const result2 = photo2.shortDateString();
    assert.equal(result2, "Unknown");
    const values3 = {
      ID: 5,
      Title: "Crazy Cat",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      TakenAt: "2012-07-08T14:45:39Z",
    };
    const photo3 = new Photo(values3);
    const result3 = photo3.shortDateString();
    assert.equal(result3, "Jul 8, 2012");
    const values4 = {
      ID: 5,
      Title: "Crazy Cat",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      TakenAt: "2012-07-08T14:45:39Z",
      Month: -1,
    };
    const photo4 = new Photo(values4);
    const result4 = photo4.shortDateString();
    assert.equal(result4, "2012");
    const values5 = {
      ID: 5,
      Title: "Crazy Cat",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      TakenAt: "2012-07-08T14:45:39Z",
      Day: -1,
    };
    const photo5 = new Photo(values5);
    const result5 = photo5.shortDateString();
    assert.equal(result5, "July 2012");
  });

  it("should test whether photo has location", () => {
    const values = { ID: 5, Title: "Crazy Cat", Lat: 36.442881666666665, Lng: 28.229493333333334 };
    const photo = new Photo(values);
    const result = photo.hasLocation();
    assert.equal(result, true);
  });

  it("should test whether photo has location", () => {
    const values = { ID: 5, Title: "Crazy Cat", Lat: 0, Lng: 0 };
    const photo = new Photo(values);
    const result = photo.hasLocation();
    assert.equal(result, false);
  });

  it("should get location", () => {
    const values = {
      ID: 5,
      Title: "Crazy Cat",
      CellID: 6,
      CellCategory: "viewpoint",
      PlaceLabel: "Cape Point, South Africa",
      PlaceCountry: "South Africa",
    };
    const photo = new Photo(values);
    const result = photo.locationInfo();
    assert.equal(result, "Cape Point, South Africa");
  });

  it("should get location", () => {
    const values = {
      ID: 5,
      Title: "Crazy Cat",
      CellID: 6,
      CellCategory: "viewpoint",
      PlaceLabel: "Cape Point, State, South Africa",
      PlaceCountry: "South Africa",
      PlaceCity: "Cape Town",
      PlaceCounty: "County",
      PlaceState: "State",
    };
    const photo = new Photo(values);
    const result = photo.locationInfo();
    assert.equal(result, "Cape Point, State, South Africa");
  });

  it("should get location", () => {
    const values = {
      ID: 5,
      Title: "Crazy Cat",
      CellCategory: "viewpoint",
      CellName: "Cape Point",
      PlaceCountry: "Africa",
      PlaceCity: "Cape Town",
      PlaceCounty: "County",
      PlaceState: "State",
    };
    const photo = new Photo(values);
    const result = photo.locationInfo();
    assert.equal(result, "Unknown");
  });

  it("should get location", () => {
    const values = { ID: 5, Title: "Crazy Cat", CountryName: "Africa", PlaceCity: "Cape Town" };
    const photo = new Photo(values);
    const result = photo.locationInfo();
    assert.equal(result, "Unknown");
  });

  it("should get camera", () => {
    const values = { ID: 5, Title: "Crazy Cat", CameraModel: "EOSD10", CameraMake: "Canon" };
    const photo = new Photo(values);
    const result = photo.getCamera();
    assert.equal(result, "Canon EOSD10");
    const values2 = {
      ID: 10,
      UID: "ABC127",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.jpg",
          Primary: true,
          FileType: "jpg",
          Hash: "1xxbgdt55",
        },
      ],
      Camera: {
        Make: "Canon",
        Model: "abc",
      },
    };
    const photo2 = new Photo(values2);
    assert.equal(photo2.getCamera(), "Canon abc");
  });

  it("should get camera", () => {
    const values = { ID: 5, Title: "Crazy Cat" };
    const photo = new Photo(values);
    const result = photo.getCamera();
    assert.equal(result, "Unknown");
  });

  it("should get collection resource", () => {
    const result = Photo.getCollectionResource();
    assert.equal(result, "photos");
  });

  it("should return batch size", () => {
    assert.equal(Photo.batchSize(), 90);
  });

  it("should get model name", () => {
    const result = Photo.getModelName();
    assert.equal(result, "Photo");
  });

  it("should like photo", () => {
    const values = { ID: 5, Title: "Crazy Cat", CountryName: "Africa", Favorite: false };
    const photo = new Photo(values);
    assert.equal(photo.Favorite, false);
    photo.like();
    assert.equal(photo.Favorite, true);
  });

  it("should unlike photo", () => {
    const values = {
      ID: 5,
      UID: "abc123",
      Title: "Crazy Cat",
      CountryName: "Africa",
      Favorite: true,
    };
    const photo = new Photo(values);
    assert.equal(photo.Favorite, true);
    photo.unlike();
    assert.equal(photo.Favorite, false);
  });

  /* TODO
    it("should toggle like",  () => {
        const values = {ID: 5, Title: "Crazy Cat", CountryName: "Africa", Favorite: true};
        const photo = new Photo(values);
        assert.equal(photo.Favorite, true);
        photo.toggleLike();
        assert.equal(photo.Favorite, false);
        photo.toggleLike();
        assert.equal(photo.Favorite, true);
    });
     */

  it("should get photo defaults", () => {
    const values = { ID: 5, UID: "ABC123" };
    const photo = new Photo(values);
    const result = photo.getDefaults();
    assert.equal(result.UID, "");
  });

  it("should get photos base name", () => {
    const values = {
      ID: 8,
      UID: "ABC123",
      Filename: "1980/01/superCuteKitten.jpg",
      FileUID: "123fgb",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.jpg",
          Primary: true,
          FileType: "TypeJpeg",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt53",
        },
      ],
    };
    const photo = new Photo(values);
    const result = photo.baseName();
    assert.equal(result, "superCuteKitten.jpg");
    const result2 = photo.baseName(5);
    assert.equal(result2, "supe…");
  });

  it("should refresh file attributes", () => {
    const values2 = { ID: 5, UID: "ABC123" };
    const photo2 = new Photo(values2);
    photo2.refreshFileAttr();
    assert.equal(photo2.Width, undefined);
    assert.equal(photo2.Height, undefined);
    assert.equal(photo2.Hash, undefined);
    const values = {
      ID: 8,
      UID: "ABC123",
      Filename: "1980/01/superCuteKitten.jpg",
      FileUID: "123fgb",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.jpg",
          Primary: true,
          FileType: "TypeJpeg",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt53",
        },
      ],
    };
    const photo = new Photo(values);
    assert.equal(photo.Width, undefined);
    assert.equal(photo.Height, undefined);
    assert.equal(photo.Hash, undefined);
    photo.refreshFileAttr();
    assert.equal(photo.Width, 500);
    assert.equal(photo.Height, 600);
    assert.equal(photo.Hash, "1xxbgdt53");
  });

  it("should return is playable", () => {
    const values = {
      ID: 8,
      UID: "ABC123",
      Filename: "1980/01/superCuteKitten.jpg",
      FileUID: "123fgb",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.jpg",
          Primary: true,
          FileType: "TypeJpeg",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt53",
        },
      ],
    };
    const photo = new Photo(values);
    assert.equal(photo.isPlayable(), false);
    const values2 = { ID: 9, UID: "ABC163" };
    const photo2 = new Photo(values2);
    assert.equal(photo2.isPlayable(), false);
    const values3 = {
      ID: 10,
      UID: "ABC127",
      Filename: "1980/01/superCuteKitten.mp4",
      FileUID: "123fgb",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.mp4",
          Primary: false,
          Video: true,
          FileType: "mp4",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt55",
        },
      ],
    };
    const photo3 = new Photo(values3);
    assert.equal(photo3.isPlayable(), true);
    const values4 = {
      ID: 1,
      UID: "ABC128",
      Filename: "1980/01/superCuteKitten.jpg",
      FileUID: "123fgb",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.jpg",
          Primary: false,
          Video: true,
          FileType: "jpg",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt53",
          Codec: "avc1",
        },
      ],
    };
    const photo4 = new Photo(values4);
    assert.equal(photo4.isPlayable(), true);
  });

  it("should return video params", () => {
    const values3 = {
      ID: 10,
      UID: "ABC127",
      Filename: "1980/01/superCuteKitten.mp4",
      FileUID: "123fgb",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.mp4",
          Primary: false,
          Video: true,
          FileType: "mp4",
          Width: 900,
          Height: 600,
          Hash: "1xxbgdt55",
        },
      ],
    };
    const photo3 = new Photo(values3);
    const result = photo3.videoParams();
    assert.equal(result.height, 457);
    assert.equal(result.width, 685);
    assert.equal(result.loop, false);
    assert.equal(result.uri, "/api/v1/videos/1xxbgdt55/public/avc");
    const values = {
      ID: 11,
      UID: "ABC127",
      Filename: "1980/01/superCuteKitten.mp4",
      FileUID: "123fgb",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.mp4",
          Primary: false,
          Video: true,
          FileType: "mp4",
          Width: 0,
          Height: 0,
          Hash: "1xxbgdt55",
        },
        {
          UID: "123fpp",
          Name: "1980/01/superCuteKitten.jpg",
          Primary: true,
          Width: 5000,
          Height: 5000,
          Hash: "1xxbgdt544",
        },
      ],
    };
    const photo = new Photo(values);
    const result2 = photo.videoParams();
    assert.equal(result2.height, 500);
    assert.equal(result2.width, 500);
    assert.equal(result2.loop, false);
    assert.equal(result2.uri, "/api/v1/videos/1xxbgdt55/public/avc");
  });

  it("should return videofile", () => {
    const values = {
      ID: 8,
      UID: "ABC123",
      Filename: "1980/01/superCuteKitten.jpg",
      FileUID: "123fgb",
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
    assert.equal(photo.videoFile(), undefined);
    const values2 = { ID: 9, UID: "ABC163" };
    const photo2 = new Photo(values2);
    assert.equal(photo2.videoFile(), false);
    const values3 = {
      ID: 10,
      UID: "ABC127",
      Filename: "1980/01/superCuteKitten.mp4",
      FileUID: "123fgb",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.mp4",
          Primary: false,
          FileType: "mp4",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt55",
        },
      ],
    };
    const photo3 = new Photo(values3);
    const file = photo3.videoFile();
    assert.equal(photo3.videoFile().Name, "1980/01/superCuteKitten.mp4");
    const values4 = {
      ID: 1,
      UID: "ABC128",
      Filename: "1980/01/superCuteKitten.jpg",
      FileUID: "123fgb",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.jpg",
          Primary: false,
          FileType: "jpg",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt53",
          Codec: "avc1",
        },
      ],
    };
    const photo4 = new Photo(values4);
    assert.equal(photo4.videoFile().Name, "1980/01/superCuteKitten.jpg");
    const file2 = photo4.videoFile();
  });

  it("should return video url", () => {
    const values = {
      ID: 8,
      UID: "ABC123",
      Filename: "1980/01/superCuteKitten.jpg",
      Hash: "703cf8f274fbb265d49c6262825780e1",
      FileUID: "123fgb",
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
    assert.equal(photo.videoUrl(), "/api/v1/videos/703cf8f274fbb265d49c6262825780e1/public/avc");
    const values2 = { ID: 9, UID: "ABC163", Hash: "2305e512e3b183ec982d60a8b608a8ca501973ba" };
    const photo2 = new Photo(values2);
    assert.equal(
      photo2.videoUrl(),
      "/api/v1/videos/2305e512e3b183ec982d60a8b608a8ca501973ba/public/avc"
    );
    const values3 = {
      ID: 10,
      UID: "ABC127",
      Filename: "1980/01/superCuteKitten.mp4",
      FileUID: "123fgb",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.mp4",
          Primary: false,
          FileType: "mp4",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt55",
        },
      ],
    };
    const photo3 = new Photo(values3);

    assert.equal(photo3.videoUrl(), "/api/v1/videos/1xxbgdt55/public/avc");
    const values4 = {
      ID: 1,
      UID: "ABC128",
      Filename: "1980/01/superCuteKitten.jpg",
      FileUID: "123fgb",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.jpg",
          Primary: false,
          FileType: "jpg",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt53",
          Codec: "avc1",
        },
      ],
    };

    const photo4 = new Photo(values4);
    assert.equal(photo4.videoUrl(), "/api/v1/videos/1xxbgdt53/public/avc");
  });

  it("should return main file", () => {
    const values = { ID: 9, UID: "ABC163", Width: 111, Height: 222 };
    const photo = new Photo(values);
    assert.equal(photo.mainFile(), photo);
    const values2 = {
      ID: 10,
      UID: "ABC127",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.mp4",
          Primary: false,
          FileType: "mp4",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt55",
        },
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.jpg",
          Primary: false,
          FileType: "jpg",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt56",
        },
      ],
    };
    const photo2 = new Photo(values2);
    const file = photo2.mainFile();
    assert.equal(file.Name, "1980/01/superCuteKitten.jpg");
    const values3 = {
      ID: 1,
      UID: "ABC128",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/NotMainKitten.jpg",
          Primary: false,
          FileType: "jpg",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt53",
        },
        {
          UID: "123fgb",
          Name: "1980/01/MainKitten.jpg",
          Primary: true,
          FileType: "jpg",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt54",
        },
      ],
    };
    const photo3 = new Photo(values3);
    const file2 = photo3.mainFile();
    assert.equal(file2.Name, "1980/01/MainKitten.jpg");
  });

  it("should return jpeg files", () => {
    const values = { ID: 8, UID: "ABC123", Filename: "1980/01/superCuteKitten.jpg" };
    const photo = new Photo(values);
    const result = photo.jpegFiles();
    assert.equal(result[0].Filename, "1980/01/superCuteKitten.jpg");
    const values3 = {
      ID: 10,
      UID: "ABC127",
      Filename: "1980/01/superCuteKitten.mp4",
      FileUID: "123fgb",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.jpg",
          Primary: false,
          FileType: FormatJpeg,
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt55",
        },
        {
          UID: "123fgz",
          Name: "1980/01/superCuteKitten.mp4",
          Primary: false,
          FileType: "mp4",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt66",
        },
      ],
    };
    const photo3 = new Photo(values3);
    const file = photo3.jpegFiles();
    assert.equal(file[0].Name, "1980/01/superCuteKitten.jpg");
  });

  it("should return main hash", () => {
    const values = { ID: 9, UID: "ABC163" };
    const photo = new Photo(values);
    assert.equal(photo.mainFileHash(), "");
    const values2 = {
      ID: 10,
      UID: "ABC127",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.mp4",
          Primary: false,
          FileType: "mp4",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt55",
        },
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.jpg",
          Primary: false,
          FileType: "jpg",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt56",
        },
      ],
    };
    const photo2 = new Photo(values2);
    assert.equal(photo2.mainFileHash(), "1xxbgdt56");
  });

  it("should test filemodels", () => {
    const values = { ID: 9, UID: "ABC163" };
    const photo = new Photo(values);
    assert.empty(photo.fileModels());
    const values2 = {
      ID: 10,
      UID: "ABC127",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/cat.jpg",
          Primary: false,
          FileType: "jpg",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt55",
        },
        {
          UID: "123fgb",
          Name: "1999/01/dog.jpg",
          Primary: true,
          FileType: "jpg",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt56",
        },
      ],
    };
    const photo2 = new Photo(values2);
    assert.equal(photo2.fileModels()[0].Name, "1999/01/dog.jpg");
    const values3 = {
      ID: 10,
      UID: "ABC127",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/cat.jpg",
          Primary: true,
          FileType: "jpg",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt55",
        },
        {
          UID: "123fgb",
          Name: "1999/01/dog.jpg",
          Primary: false,
          FileType: "jpg",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt56",
        },
      ],
    };
    const photo3 = new Photo(values3);
    assert.equal(photo3.fileModels()[0].Name, "1980/01/cat.jpg");
    const values4 = {
      ID: 10,
      UID: "ABC127",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/cat.jpg",
          Primary: true,
          FileType: "jpg",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt55",
        },
      ],
    };
    const photo4 = new Photo(values4);
    assert.equal(photo4.fileModels()[0].Name, "1980/01/cat.jpg");
  });

  it("should get country name", () => {
    const values = { ID: 5, UID: "ABC123", Country: "zz" };
    const photo = new Photo(values);
    assert.equal(photo.countryName(), "Unknown");
    const values2 = { ID: 5, UID: "ABC123", Country: "es" };
    const photo2 = new Photo(values2);
    assert.equal(photo2.countryName(), "Spain");
  });

  it("should get location info", () => {
    const values = { ID: 5, UID: "ABC123", Country: "zz", PlaceID: "zz", PlaceLabel: "Nice beach" };
    const photo = new Photo(values);
    assert.equal(photo.locationInfo(), "Nice beach");
    const values2 = { ID: 5, UID: "ABC123", Country: "es", PlaceID: "zz" };
    const photo2 = new Photo(values2);
    assert.equal(photo2.locationInfo(), "Spain");
  });

  it("should return video info", () => {
    const values = {
      ID: 9,
      UID: "ABC163",
    };
    const photo = new Photo(values);
    assert.equal(photo.getVideoInfo(), "Video");
    const values2 = {
      ID: 10,
      UID: "ABC127",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.mp4",
          Primary: false,
          FileType: "mp4",
          Hash: "1xxbgdt55",
        },
      ],
    };
    const photo2 = new Photo(values2);
    assert.equal(photo2.getVideoInfo(), "Video");
    const values3 = {
      ID: 10,
      UID: "ABC127",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.mp4",
          Primary: false,
          FileType: "mp4",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt55",
          Duration: 6000,
          Size: 222897,
          Codec: "avc1",
        },
      ],
    };
    const photo3 = new Photo(values3);
    assert.equal(photo3.getVideoInfo(), "6µs, AVC, 500 × 600, 0.2 MB");
    const values4 = {
      ID: 10,
      UID: "ABC127",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.mp4",
          Primary: false,
          FileType: "mp4",
          Hash: "1xxbgdt55",
          Duration: 6000,
          Size: 10240,
          Codec: "avc1",
        },
        {
          UID: "345fgb",
          Name: "1980/01/superCuteKitten.jpg",
          Primary: true,
          FileType: "jpg",
          Hash: "1xxbgjhu5",
          Width: 300,
          Height: 500,
        },
      ],
    };
    const photo4 = new Photo(values4);
    assert.equal(photo4.getVideoInfo(), "6µs, AVC, 300 × 500, 10.0 KB");
  });

  it("should return photo info", () => {
    const values = {
      ID: 9,
      UID: "ABC163",
    };
    const photo = new Photo(values);
    assert.equal(photo.getPhotoInfo(), "Unknown");
    const values2 = {
      ID: 10,
      UID: "ABC127",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.jpg",
          Primary: true,
          FileType: "jpg",
          Hash: "1xxbgdt55",
        },
      ],
      Size: "300",
      Camera: {
        Make: "Canon",
        Model: "abc",
      },
    };
    const photo2 = new Photo(values2);
    assert.equal(photo2.getPhotoInfo(), "Canon abc");
    const values3 = {
      ID: 10,
      UID: "ABC127",
      CameraMake: "Canon",
      CameraModel: "abcde",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.mp4",
          Primary: false,
          FileType: "mp4",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt55",
          Duration: 6000,
          Codec: "avc1",
        },
      ],
    };
    const photo3 = new Photo(values3);
    assert.equal(photo3.getPhotoInfo(), "Canon abcde, AVC, 500 × 600");
    const values4 = {
      ID: 10,
      UID: "ABC127",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.mp4",
          Primary: false,
          FileType: "mp4",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt55",
          Duration: 6000,
          Size: 300,
          Codec: "avc1",
        },
        {
          UID: "123fgx",
          Name: "1980/01/superCuteKitten.jpg",
          Primary: true,
          FileType: "jpg",
          Width: 800,
          Height: 600,
          Hash: "1xxbgdt55",
          Duration: 6000,
          Size: 200,
          Codec: "avc1",
        },
      ],
    };
    const photo4 = new Photo(values4);
    assert.equal(photo3.getPhotoInfo(), "Canon abcde, AVC, 500 × 600");
  });

  it("should return lens info", () => {
    const values = {
      ID: "674-860",
      UID: "ps22wlskqtcmu9l3",
      Type: "raw",
      TypeSrc: "",
      TakenAt: "2018-10-05T08:47:32Z",
      TakenAtLocal: "2018-10-05T08:47:32Z",
      TakenSrc: "meta",
      TimeZone: "",
      Path: "raw images/Canon EOS 700 D",
      Name: "_MG_9509",
      OriginalName: "",
      Title: "Unknown / 2018",
      Description: "",
      Year: 2018,
      Month: 10,
      Day: 5,
      Country: "zz",
      Stack: 0,
      Favorite: false,
      Private: false,
      Iso: 100,
      FocalLength: 50,
      FNumber: 2.8,
      Exposure: "1/1600",
      Quality: 3,
      Resolution: 18,
      Color: 0,
      Scan: false,
      Panorama: false,
      CameraID: 47,
      CameraSrc: "meta",
      CameraSerial: "338075021697",
      CameraModel: "EOS 700D",
      CameraMake: "Canon",
      LensID: 47,
      LensModel: "EF50mm f/1.8 II",
      CellID: "zz",
      PlaceID: "zz",
      PlaceSrc: "",
      PlaceLabel: "Unknown",
      PlaceCity: "Unknown",
      PlaceState: "Unknown",
      PlaceCountry: "zz",
      InstanceID: "",
      FileUID: "fs25jsa22w9g851o",
      FileRoot: "sidecar",
      FileName: "raw images/Canon EOS 700 D/_MG_9509.CR2.jpg",
      Hash: "7dc01e8cb588f3cfe31694ac2fece10167d88eec",
      Width: 5198,
      Height: 3462,
      Portrait: false,
      Files: [],
    };
    const photo = new Photo(values);
    assert.equal(photo.getLensInfo(), "EF50mm ƒ/1.8 II, 50mm, ƒ/2.8, ISO 100, 1/1600");
  });

  it("should archive photo", () => {
    const values = { ID: 5, Title: "Crazy Cat", CountryName: "Africa", Favorite: false };
    const photo = new Photo(values);
    return photo
      .archive()
      .then((response) => {
        assert.equal(200, response.status);
        assert.deepEqual({ photos: [1, 3] }, response.data);
        return Promise.resolve();
      })
      .catch((error) => {
        return Promise.reject(error);
      });
  });

  it("should approve photo", () => {
    const values = {
      ID: 5,
      UID: "pqbemz8276mhtobh",
      Title: "Crazy Cat",
      CountryName: "Africa",
      Favorite: false,
    };
    const photo = new Photo(values);
    return photo
      .approve()
      .then((response) => {
        assert.equal(200, response.status);
        return Promise.resolve();
      })
      .catch((error) => {
        return Promise.reject(error);
      });
  });

  it("should toggle private", () => {
    const values = { ID: 5, Title: "Crazy Cat", CountryName: "Africa", Private: true };
    const photo = new Photo(values);
    assert.equal(photo.Private, true);
    photo.togglePrivate();
    assert.equal(photo.Private, false);
    photo.togglePrivate();
    assert.equal(photo.Private, true);
  });

  it("should mark photo as primary", (done) => {
    const values = {
      ID: 10,
      UID: "pqbemz8276mhtobh",
      Files: [
        {
          UID: "fqbfk181n4ca5sud",
          Name: "1980/01/superCuteKitten.mp4",
          Primary: false,
          FileType: "mp4",
          Hash: "1xxbgdt55",
        },
      ],
    };
    const photo = new Photo(values);
    photo
      .primaryFile("fqbfk181n4ca5sud")
      .then((response) => {
        assert.equal(response.Files[0].Primary, true);
        done();
      })
      .catch((error) => {
        done(error);
      });
  });

  it("should unstack", (done) => {
    const values = {
      ID: 10,
      UID: "pqbemz8276mhtobh",
      Files: [
        {
          UID: "fqbfk181n4ca5sud",
          Name: "1980/01/superCuteKitten.mp4",
          Primary: false,
          FileType: "mp4",
          Hash: "1xxbgdt55",
        },
      ],
    };
    const photo = new Photo(values);
    photo
      .unstackFile("fqbfk181n4ca5sud")
      .then((response) => {
        assert.equal(response.success, "ok");
        done();
      })
      .catch((error) => {
        done(error);
      });
  });

  it("should delete file", (done) => {
    const values = {
      ID: 10,
      UID: "pqbemz8276mhtobh",
      Files: [
        {
          UID: "fqbfk181n4ca5sud",
          Name: "1980/01/superCuteKitten.mp4",
          Primary: false,
          FileType: "mp4",
          Hash: "1xxbgdt55",
        },
        {
          UID: "fqbfk181n4ca5abc",
          Name: "1980/01/superCuteKitten.mp4",
          Primary: true,
          FileType: "mp4",
          Hash: "1xxbgdt89",
        },
      ],
    };
    const photo = new Photo(values);
    photo
      .deleteFile("fqbfk181n4ca5sud")
      .then((response) => {
        assert.equal(response.success, "successfully deleted");
        done();
      })
      .catch((error) => {
        done(error);
      });
  });

  it("should add label", (done) => {
    const values = {
      ID: 10,
      UID: "pqbemz8276mhtobh",
    };
    const photo = new Photo(values);
    photo
      .addLabel("Cat")
      .then((response) => {
        assert.equal(response.success, "ok");
        done();
      })
      .catch((error) => {
        done(error);
      });
  });

  it("should activate label", (done) => {
    const values = {
      ID: 10,
      UID: "pqbemz8276mhtobh",
    };
    const photo = new Photo(values);
    photo
      .activateLabel(12345)
      .then((response) => {
        assert.equal(response.success, "ok");
        done();
      })
      .catch((error) => {
        done(error);
      });
  });

  it("should rename label", (done) => {
    const values = {
      ID: 10,
      UID: "pqbemz8276mhtobh",
    };
    const photo = new Photo(values);
    photo
      .renameLabel(12345, "Sommer")
      .then((response) => {
        assert.equal(response.success, "ok");
        done();
      })
      .catch((error) => {
        done(error);
      });
  });

  it("should remove label", (done) => {
    const values = {
      ID: 10,
      UID: "pqbemz8276mhtobh",
    };
    const photo = new Photo(values);
    photo
      .removeLabel(12345)
      .then((response) => {
        assert.equal(response.success, "ok");
        done();
      })
      .catch((error) => {
        done(error);
      });
  });

  it("should test update", (done) => {
    const values = {
      ID: 10,
      UID: "pqbemz8276mhtobh",
      Lat: 1.1,
      Lng: 3.3,
      CameraID: 123,
      Title: "Test Titel",
      Description: "Super nice video",
      Day: 10,
      Country: "es",
      Files: [
        {
          UID: "fqbfk181n4ca5sud",
          Name: "1980/01/superCuteKitten.mp4",
          Primary: false,
          FileType: "mp4",
          Hash: "1xxbgdt55",
        },
      ],
      Details: [
        {
          Keywords: "old",
          Notes: "old notes",
          Subject: "old subject",
          Artist: "Old Artist",
          Copyright: "ABC",
          License: "test",
        },
      ],
    };
    const photo = new Photo(values);
    photo.Title = "New Title";
    photo.Type = "newtype";
    photo.Description = "New description";
    photo.Day = 21;
    photo.Country = "de";
    photo.CameraID = "newcameraid";
    photo.Details.Keywords = "newkeyword";
    photo.Details.Notes = "New Notes";
    photo.Details.Subject = "New Photo Subject";
    photo.Details.Artist = "New Artist";
    photo.Details.Copyright = "New Copyright";
    photo.Details.License = "New License";
    photo
      .update()
      .then((response) => {
        assert.equal(response.TitleSrc, "manual");
        done();
      })
      .catch((error) => {
        done(error);
      });
    assert.equal(photo.Title, "New Title");
    assert.equal(photo.Type, "newtype");
    assert.equal(photo.Description, "New description");
    assert.equal(photo.Day, 21);
    assert.equal(photo.Country, "de");
    assert.equal(photo.CameraID, "newcameraid");
    assert.equal(photo.Details.Keywords, "newkeyword");
    assert.equal(photo.Details.Notes, "New Notes");
    assert.equal(photo.Details.Subject, "New Photo Subject");
    assert.equal(photo.Details.Artist, "New Artist");
    assert.equal(photo.Details.Copyright, "New Copyright");
    assert.equal(photo.Details.License, "New License");
  });

  it("should test get Markers", () => {
    const values = {
      ID: 10,
      UID: "pqbemz8276mhtobh",
      Lat: 1.1,
      Lng: 3.3,
      CameraID: 123,
      Title: "Test Titel",
      Description: "Super nice video",
      Files: [
        {
          UID: "fqbfk181n4ca5sud",
          Name: "1980/01/superCuteKitten.mp4",
          Primary: false,
          FileType: "mp4",
          Hash: "1xxbgdt55",
        },
      ],
    };
    const photo = new Photo(values);
    const result = photo.getMarkers(true);
    assert.empty(result);
    const values2 = {
      ID: 10,
      UID: "pqbemz8276mhtobh",
      Lat: 1.1,
      Lng: 3.3,
      CameraID: 123,
      Title: "Test Titel",
      Description: "Super nice video",
      Files: [
        {
          UID: "fqbfk181n4ca5sud",
          Name: "1980/01/superCuteKitten.mp4",
          Primary: true,
          FileType: "mp4",
          Hash: "1xxbgdt55",
          Markers: [
            {
              UID: "aaa123",
              Invalid: false,
            },
            {
              UID: "bbb123",
              Invalid: true,
            },
          ],
        },
      ],
    };
    const photo2 = new Photo(values2);
    const result2 = photo2.getMarkers(true);
    assert.equal(result2.length, 1);
    const result3 = photo2.getMarkers(false);
    assert.equal(result3.length, 2);
  });
});
