import { describe, it, expect } from "vitest";
import "../fixtures";
import * as media from "common/media";
import { Photo, BatchSize } from "model/photo";

describe("model/photo", () => {
  it("should get photo entity name", () => {
    const values = { UID: 5, Title: "Crazy Cat" };
    const photo = new Photo(values);
    const result = photo.getEntityName();
    expect(result).toBe("Crazy Cat");
  });

  it("should get photo uuid", () => {
    const values = { ID: 5, Title: "Crazy Cat", UID: 789 };
    const photo = new Photo(values);
    const result = photo.getId();
    expect(result).toBe(789);
  });

  it("should get photo title", () => {
    const values = { ID: 5, Title: "Crazy Cat", UID: 789 };
    const photo = new Photo(values);
    const result = photo.getTitle();
    expect(result).toBe("Crazy Cat");
  });

  it("should get photo maps link", () => {
    const values = { ID: 5, Title: "Crazy Cat", Lat: 36.442881666666665, Lng: 28.229493333333334 };
    const photo = new Photo(values);
    const result = photo.getGoogleMapsLink();
    expect(result).toBe("https://www.google.com/maps/place/36.442881666666665,28.229493333333334");
  });

  it("should get photo thumbnail url", () => {
    const values = { ID: 5, Title: "Crazy Cat", Hash: "97b8cf7b3710bec95f6609487bbdd62489b95fb2" };
    const photo = new Photo(values);
    const result = photo.thumbnailUrl("tile500");
    expect(result).toBe("/api/v1/t/97b8cf7b3710bec95f6609487bbdd62489b95fb2/public/tile500");
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
          Hash: "be651a4fffd699196cfd5dd14b6ec9cb10a8531a",
        },
      ],
    };
    const photo2 = new Photo(values2);
    const result2 = photo2.thumbnailUrl("tile500");
    expect(result2).toBe("/api/v1/t/be651a4fffd699196cfd5dd14b6ec9cb10a8531a/public/tile500");
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
    expect(result3).toBe("/static/img/404.jpg");
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
    expect(result2).toContain("is-photo");
    expect(result2).toContain("uid-ABC127");
    expect(result2).toContain("type-video");
    expect(result2).toContain("is-portrait");
    expect(result2).toContain("is-favorite");
    expect(result2).toContain("is-private");
    expect(result2).not.toContain("is-stack");
    expect(result2).toContain("is-playable");
  });

  it("should get photo download url", () => {
    const values = { ID: 5, Title: "Crazy Cat", Hash: "97b8cf7b3710bec95f6609487bbdd62489b95fb2" };
    const photo = new Photo(values);
    const result = photo.getDownloadUrl();
    expect(result).toBe("/api/v1/dl/97b8cf7b3710bec95f6609487bbdd62489b95fb2?t=2lbh9x09");
  });

  it("should calculate photo size", () => {
    const values = { ID: 5, Title: "Crazy Cat", Width: 500, Height: 200 };
    const photo = new Photo(values);
    const result = photo.calculateSize(500, 200);
    expect(result.width).toBe(500);
    expect(result.height).toBe(200);
  });

  it("should calculate photo size with srcAspectRatio < maxAspectRatio", () => {
    const values = { ID: 5, Title: "Crazy Cat", Width: 500, Height: 200 };
    const photo = new Photo(values);
    const result = photo.calculateSize(300, 50);
    expect(result.width).toBe(125);
    expect(result.height).toBe(50);
  });

  it("should calculate photo size with srcAspectRatio > maxAspectRatio", () => {
    const values = { ID: 5, Title: "Crazy Cat", Width: 500, Height: 200 };
    const photo = new Photo(values);
    const result = photo.calculateSize(400, 300);
    expect(result.width).toBe(400);
    expect(result.height).toBe(160);
  });

  it("should get local day string", () => {
    const values = { ID: 5, Title: "Crazy Cat", TakenAt: "2012-07-08T14:45:39Z", TimeZone: "UTC" };
    const photo = new Photo(values);
    const result = photo.localDayString();
    expect(result.length).toBe(2);
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
    expect(result2).toBe("08");
  });

  it("should get local month string", () => {
    const values = { ID: 5, Title: "Crazy Cat", TakenAt: "2012-07-08T14:45:39Z", TimeZone: "UTC" };
    const photo = new Photo(values);
    const result = photo.localMonthString();
    expect(result).toBe((new Date().getMonth() + 1).toString().padStart(2, "0"));
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
    expect(result2).toBe("08");
  });

  it("should get local year string", () => {
    const values = { ID: 5, Title: "Crazy Cat", TakenAt: "2012-07-08T14:45:39Z", TimeZone: "UTC" };
    const photo = new Photo(values);
    const result = photo.localYearString();
    expect(result).toBe(new Date().getFullYear().toString().padStart(4, "0"));
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
    expect(result2).toBe("2010");
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
    expect(result).toBe("2012-07-08T14:45:39Z");
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
    expect(String(result)).toBe("2012-07-08T14:45:39.000Z");
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
    expect(String(photo.localDateString("10:00:00"))).toBe("2012-07-08T10:00:00Z");
    const result = photo.localDate();
    expect(String(result)).toBe("2012-07-08T14:45:39.000Z");
  });

  it("should get date string", () => {
    const values = {
      ID: 5,
      Title: "Crazy Cat",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      TakenAt: "2012-07-08T12:45:39Z",
      TimeZone: "Europe/Berlin",
    };
    const photo = new Photo(values);
    const result = photo.getDateString(false).replaceAll("\u202f", " ");
    expect(result.startsWith("Sunday, July 8, 2012")).toBe(true);
    expect(result.endsWith("2:45 PM")).toBe(true);
    const values2 = { ID: 5, Title: "Crazy Cat", TakenAtLocal: "", TakenAt: "", TimeZone: "UTC" };
    const photo2 = new Photo(values2);
    const result2 = photo2.getDateString();
    expect(result2).toBe("Unknown");
    const values3 = {
      ID: 5,
      Title: "Crazy Cat",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      TakenAt: "2012-07-08T14:45:39Z",
    };
    const photo3 = new Photo(values3);
    const result3 = photo3.getDateString();
    expect(result3).toBe("Sunday, July 8, 2012");
    const values4 = {
      ID: 5,
      Title: "Crazy Cat",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      TakenAt: "2012-07-08T14:45:39Z",
      Month: -1,
    };
    const photo4 = new Photo(values4);
    const result4 = photo4.getDateString();
    expect(result4).toBe("2012");
    const values5 = {
      ID: 5,
      Title: "Crazy Cat",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      TakenAt: "2012-07-08T14:45:39Z",
      Day: -1,
    };
    const photo5 = new Photo(values5);
    const result5 = photo5.getDateString();
    expect(result5).toBe("July 2012");
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
    expect(result).toBe("7/8/2012");
    const values2 = { ID: 5, Title: "Crazy Cat", TakenAtLocal: "", TakenAt: "", TimeZone: "UTC" };
    const photo2 = new Photo(values2);
    const result2 = photo2.shortDateString();
    expect(result2).toBe("Unknown");
    const values3 = {
      ID: 5,
      Title: "Crazy Cat",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      TakenAt: "2012-07-08T14:45:39Z",
    };
    const photo3 = new Photo(values3);
    const result3 = photo3.shortDateString();
    expect(result3).toBe("7/8/2012");
    const values4 = {
      ID: 5,
      Title: "Crazy Cat",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      TakenAt: "2012-07-08T14:45:39Z",
      Month: -1,
    };
    const photo4 = new Photo(values4);
    const result4 = photo4.shortDateString();
    expect(result4).toBe("2012");
    const values5 = {
      ID: 5,
      Title: "Crazy Cat",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      TakenAt: "2012-07-08T14:45:39Z",
      Day: -1,
    };
    const photo5 = new Photo(values5);
    const result5 = photo5.shortDateString();
    expect(result5).toBe("July 2012");
  });

  it("should test whether photo has location", () => {
    const values = { ID: 5, Title: "Crazy Cat", Lat: 36.442881666666665, Lng: 28.229493333333334 };
    const photo = new Photo(values);
    const result = photo.hasLocation();
    expect(result).toBe(true);
  });

  it("should test whether photo has location", () => {
    const values = { ID: 5, Title: "Crazy Cat", Lat: 0, Lng: 0 };
    const photo = new Photo(values);
    const result = photo.hasLocation();
    expect(result).toBe(false);
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
    expect(result).toBe("Cape Point, South Africa");
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
    expect(result).toBe("Cape Point, State, South Africa");
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
    expect(result).toBe("Unknown");
  });

  it("should get location", () => {
    const values = { ID: 5, Title: "Crazy Cat", CountryName: "Africa", PlaceCity: "Cape Town" };
    const photo = new Photo(values);
    const result = photo.locationInfo();
    expect(result).toBe("Unknown");
  });

  it("should get camera", () => {
    const values = { ID: 5, Title: "Crazy Cat", CameraModel: "EOSD10", CameraMake: "Canon" };
    const photo = new Photo(values);
    const result = photo.getCamera();
    expect(result).toBe("Canon EOSD10");
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
    expect(photo2.getCamera()).toBe("Canon abc");
  });

  it("should get camera", () => {
    const values = { ID: 5, Title: "Crazy Cat" };
    const photo = new Photo(values);
    const result = photo.getCamera();
    expect(result).toBe("Unknown");
  });

  it("should get collection resource", () => {
    const result = Photo.getCollectionResource();
    expect(result).toBe("photos");
  });

  it("should return batch size", () => {
    expect(Photo.batchSize()).toBe(BatchSize);
  });

  it("should get model name", () => {
    const result = Photo.getModelName();
    expect(result).toBe("Photo");
  });

  it("should like photo", () => {
    const values = { ID: 5, Title: "Crazy Cat", CountryName: "Africa", Favorite: false };
    const photo = new Photo(values);
    expect(photo.Favorite).toBe(false);
    photo.like();
    expect(photo.Favorite).toBe(true);
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
    expect(photo.Favorite).toBe(true);
    photo.unlike();
    expect(photo.Favorite).toBe(false);
  });

  it("should get photo defaults", () => {
    const values = { ID: 5, UID: "ABC123" };
    const photo = new Photo(values);
    const result = photo.getDefaults();
    expect(result.UID).toBe("");
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
          Hash: "ca3e60b9825bd61ee6369fcefe22f4eb92631bb5",
        },
      ],
    };
    const photo = new Photo(values);
    const result = photo.baseName();
    expect(result).toBe("superCuteKitten.jpg");
    const result2 = photo.baseName(5);
    expect(result2).toBe("supe…");
  });

  it("should refresh file attributes", () => {
    const values2 = { ID: 5, UID: "ABC123" };
    const photo2 = new Photo(values2);
    photo2.refreshFileAttr();
    expect(photo2.Width).toBeUndefined();
    expect(photo2.Height).toBeUndefined();
    expect(photo2.Hash).toBeUndefined();
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
          Hash: "ca3e60b9825bd61ee6369fcefe22f4eb92631bb5",
        },
      ],
    };
    const photo = new Photo(values);
    expect(photo.Width).toBeUndefined();
    expect(photo.Height).toBeUndefined();
    expect(photo.Hash).toBeUndefined();
    photo.refreshFileAttr();
    expect(photo.Width).toBe(500);
    expect(photo.Height).toBe(600);
    expect(photo.Hash).toBe("ca3e60b9825bd61ee6369fcefe22f4eb92631bb5");
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
          Hash: "ca3e60b9825bd61ee6369fcefe22f4eb92631bb5",
        },
      ],
    };
    const photo = new Photo(values);
    expect(photo.isPlayable()).toBe(false);
    const values2 = { ID: 9, UID: "ABC163" };
    const photo2 = new Photo(values2);
    expect(photo2.isPlayable()).toBe(false);
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
          Hash: "c1e30d265eab968155082c8e86d85815a8389479",
        },
      ],
    };
    const photo3 = new Photo(values3);
    expect(photo3.isPlayable()).toBe(true);
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
          Hash: "ca3e60b9825bd61ee6369fcefe22f4eb92631bb5",
          Codec: "avc1",
        },
      ],
    };
    const photo4 = new Photo(values4);
    expect(photo4.isPlayable()).toBe(true);
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
          Hash: "c1e30d265eab968155082c8e86d85815a8389479",
        },
      ],
    };
    const photo3 = new Photo(values3);
    const result = photo3.videoParams();
    expect(result.height).toBeGreaterThan(340);
    expect(result.width).toBeGreaterThan(510);
    expect(result.loop).toBe(false);
    expect(result.uri).toBe("/api/v1/videos/c1e30d265eab968155082c8e86d85815a8389479/public/avc");
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
          Hash: "c1e30d265eab968155082c8e86d85815a8389479",
        },
        {
          UID: "123fpp",
          Name: "1980/01/superCuteKitten.jpg",
          Primary: true,
          Width: 5000,
          Height: 5000,
          Hash: "ca3e60b9825bd61ee6369fcefe22f4eb92631bb5",
        },
      ],
    };
    const photo = new Photo(values);
    const result2 = photo.videoParams();
    expect(result2.height).toBeGreaterThan(340);
    expect(result2.width).toBeGreaterThan(340);
    expect(result2.loop).toBe(false);
    expect(result2.uri).toBe("/api/v1/videos/c1e30d265eab968155082c8e86d85815a8389479/public/avc");
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
          Hash: "c1e30d265eab968155082c8e86d85815a8389479",
        },
      ],
    };
    const photo = new Photo(values);
    expect(photo.videoFile()).toBeUndefined();
    const values2 = { ID: 9, UID: "ABC163" };
    const photo2 = new Photo(values2);
    expect(photo2.videoFile()).toBe(false);
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
          Hash: "c1e30d265eab968155082c8e86d85815a8389479",
        },
      ],
    };
    const photo3 = new Photo(values3);
    const file = photo3.videoFile();
    expect(file.Name).toBe("1980/01/superCuteKitten.mp4");
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
    expect(photo4.videoFile().Name).toBe("1980/01/superCuteKitten.jpg");
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
    expect(photo.videoContentType()).toBe(media.ContentTypeMp4AvcMain);
    expect(photo.videoUrl()).toBe("/api/v1/videos/703cf8f274fbb265d49c6262825780e1/public/avc");
    const values2 = { ID: 9, UID: "ABC163", Hash: "2305e512e3b183ec982d60a8b608a8ca501973ba" };
    const photo2 = new Photo(values2);
    expect(photo2.videoUrl()).toBe("/api/v1/videos/2305e512e3b183ec982d60a8b608a8ca501973ba/public/avc");
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
    expect(photo3.videoUrl()).toBe("/api/v1/videos/1xxbgdt55/public/avc");
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
    expect(photo4.videoUrl()).toBe("/api/v1/videos/1xxbgdt53/public/avc");
  });

  it("should return main file", () => {
    const values = { ID: 9, UID: "ABC163", Width: 111, Height: 222 };
    const photo = new Photo(values);
    expect(photo.primaryFile()).toBe(photo);
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
    const file = photo2.primaryFile();
    expect(file.Name).toBe("1980/01/superCuteKitten.jpg");
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
    const file2 = photo3.primaryFile();
    expect(file2.Name).toBe("1980/01/MainKitten.jpg");
  });

  it("should return jpeg files", () => {
    const values = { ID: 8, UID: "ABC123", Filename: "1980/01/superCuteKitten.jpg" };
    const photo = new Photo(values);
    const result = photo.jpegFiles();
    expect(result[0].Filename).toBe("1980/01/superCuteKitten.jpg");
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
          FileType: media.FormatJpeg,
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
    expect(file[0].Name).toBe("1980/01/superCuteKitten.jpg");
  });

  it("should return file hash", () => {
    const values = { ID: 9, UID: "ABC163" };
    const photo = new Photo(values);
    expect(photo.fileHash()).toBe("");
    photo.Hash = "123693d2c2b9afdba19f97d1c92963953e1d2cfe";
    expect(photo.fileHash()).toBe("123693d2c2b9afdba19f97d1c92963953e1d2cfe");
    const values2 = {
      ID: 10,
      UID: "ABC127",
      Type: "video",
      Hash: "123693d2c2b9afdba19f97d1c92963953e1d2cfe",
      Files: [
        {
          UID: "fsr3uh0u30trle4l",
          Name: "1980/01/superCuteKitten.mp4",
          Primary: false,
          Root: "/",
          MediaType: "video",
          FileType: "mp4",
          Width: 500,
          Height: 600,
          Hash: "617693d2c2b9afdba19f97d1c92963953e1d2cfe",
        },
        {
          UID: "fsr3uh0g2us6cwg4",
          Name: "1980/01/superCuteKitten.jpg",
          Primary: false,
          Root: "/",
          MediaType: "image",
          FileType: "jpg",
          Width: 500,
          Height: 600,
          Hash: "9249cee32bc8adc6ba996a6b78dd84c03b5a0153",
        },
      ],
    };
    const photo2 = new Photo(values2);
    expect(photo2.fileHash()).toBe("9249cee32bc8adc6ba996a6b78dd84c03b5a0153");
    photo2.Files = [
      {
        UID: "fsr3uh0u30trle4l",
        Name: "1980/01/superCuteKitten.mp4",
        Primary: false,
        Root: "/",
        MediaType: "video",
        FileType: "mp4",
        Width: 500,
        Height: 600,
        Hash: "617693d2c2b9afdba19f97d1c92963953e1d2cfe",
      },
      {
        UID: "fsr3uh0g2us6cwg4",
        Name: "1980/01/superCuteKitten.jpg",
        Primary: false,
        Root: "/",
        MediaType: "image",
        FileType: "invalid",
        Width: 500,
        Height: 600,
        Hash: "9249cee32bc8adc6ba996a6b78dd84c03b5a0153",
      },
    ];
    expect(photo2.fileHash()).toBe("617693d2c2b9afdba19f97d1c92963953e1d2cfe");
  });

  it("should return file models", () => {
    const values = { ID: 9, UID: "ABC163" };
    const photo = new Photo(values);
    expect(photo.fileModels()).toEqual([]);
    const values2 = {
      ID: 10,
      UID: "ABC127",
      Type: "video",
      Files: [
        {
          UID: "fsr3uh0u30trle4l",
          Name: "1980/01/cat.jpg",
          Primary: false,
          Root: "/",
          FileType: "jpg",
          MediaType: "image",
          Width: 500,
          Height: 600,
          Hash: "35c905d21486b400814bd2d8479ed2e780440b1a",
        },
        {
          UID: "fsr3uh0g2us6cwg4",
          Name: "1999/01/dog.jpg",
          Primary: true,
          Root: "/",
          FileType: "jpg",
          MediaType: "image",
          Width: 500,
          Height: 600,
          Hash: "617693d2c2b9afdba19f97d1c92963953e1d2cfe",
        },
        {
          UID: "fsr3uh10nrgs63a2",
          Name: "1999/01/dog.mov",
          Video: true,
          Root: "/",
          FileType: "mov",
          MediaType: "video",
          Width: 500,
          Height: 600,
          Hash: "9249cee32bc8adc6ba996a6b78dd84c03b5a0153",
        },
      ],
    };
    const photo2 = new Photo(values2);
    expect(photo2.fileModels()[0].Name).toBe("1999/01/dog.mov");
    const values3 = {
      ID: 10,
      UID: "ABC127",
      Files: [
        {
          UID: "fsr3uh0u30trle4l",
          Name: "1980/01/cat.jpg",
          Primary: true,
          FileType: "jpg",
          Width: 500,
          Height: 600,
          Hash: "35c905d21486b400814bd2d8479ed2e780440b1a",
        },
        {
          UID: "fsr3uh0g2us6cwg4",
          Name: "1999/01/dog.jpg",
          Primary: false,
          FileType: "jpg",
          Width: 500,
          Height: 600,
          Hash: "617693d2c2b9afdba19f97d1c92963953e1d2cfe",
        },
      ],
    };
    const photo3 = new Photo(values3);
    expect(photo3.fileModels()[0].Name).toBe("1980/01/cat.jpg");
    const values4 = {
      ID: 10,
      UID: "ABC127",
      Files: [
        {
          UID: "fsr3uh0u30trle4l",
          Name: "1980/01/cat.jpg",
          Primary: true,
          Root: "/",
          FileType: "jpg",
          MediaType: "image",
          Width: 500,
          Height: 600,
          Hash: "35c905d21486b400814bd2d8479ed2e780440b1a",
        },
      ],
    };
    const photo4 = new Photo(values4);
    expect(photo4.fileModels()[0].Name).toBe("1980/01/cat.jpg");
  });

  it("should get country name", () => {
    const values = { ID: 5, UID: "ABC123", Country: "zz" };
    const photo = new Photo(values);
    expect(photo.countryName()).toBe("Unknown");
    const values2 = { ID: 5, UID: "ABC123", Country: "es" };
    const photo2 = new Photo(values2);
    expect(photo2.countryName()).toBe("Spain");
  });

  it("should get location info", () => {
    const values = { ID: 5, UID: "ABC123", Country: "zz", PlaceID: "zz", PlaceLabel: "Nice beach" };
    const photo = new Photo(values);
    expect(photo.locationInfo()).toBe("Nice beach");
    const values2 = { ID: 5, UID: "ABC123", Country: "es", PlaceID: "zz" };
    const photo2 = new Photo(values2);
    expect(photo2.locationInfo()).toBe("Spain");
  });

  it("should return video info", () => {
    const values = {
      ID: 9,
      UID: "ABC163",
    };
    const photo = new Photo(values);
    expect(photo.getVideoInfo()).toBe("Video");
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
    expect(photo2.getVideoInfo()).toBe("MP4");
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
    expect(photo3.getVideoInfo()).toBe("6µs, AVC, 500 × 600, 218 KB");
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
    expect(photo4.getVideoInfo()).toBe("6µs, AVC, 300 × 500, 10 KB");
    expect(photo4.getDurationInfo()).toBe("6µs");
  });

  it("should return photo info", () => {
    const values = {
      ID: 9,
      UID: "ABC163",
    };
    const photo = new Photo(values);
    expect(photo.getCameraInfo()).toBe("Unknown");
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
    expect(photo2.getCameraInfo()).toBe("Canon abc");
    const values3 = {
      ID: 10,
      UID: "ABC127",
      CameraMake: "Canon",
      CameraModel: "EOS 6D",
      Iso: 100,
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
    expect(photo3.getCameraInfo()).toBe("Canon EOS 6D, ISO 100");
    const values4 = {
      ID: 10,
      UID: "ABC127",
      CameraID: 2,
      CameraMake: "Canon",
      Iso: 200,
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
    expect(photo4.getCameraInfo()).toBe("Canon, ISO 200");
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
      TimeZone: "Local",
      Path: "raw images/Canon EOS 700 D",
      Name: "_MG_9509",
      OriginalName: "",
      Title: "Unknown / 2018",
      Caption: "",
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
    expect(photo.getLensInfo()).toBe("EF50mm ƒ/1.8 II, 50mm, ƒ/2.8");
  });

  it("should archive photo", async () => {
    const values = { ID: 5, Title: "Crazy Cat", CountryName: "Africa", Favorite: false };
    const photo = new Photo(values);
    const response = await photo.archive();
    expect(response.status).toBe(200);
    expect(response.data).toEqual({ photos: [1, 3] });
  });

  it("should approve photo", async () => {
    const values = {
      ID: 5,
      UID: "pqbemz8276mhtobh",
      Title: "Crazy Cat",
      CountryName: "Africa",
      Favorite: false,
    };
    const photo = new Photo(values);
    const response = await photo.approve();
    expect(response.status).toBe(200);
  });

  it("should toggle private", () => {
    const values = { ID: 5, Title: "Crazy Cat", CountryName: "Africa", Private: true };
    const photo = new Photo(values);
    expect(photo.Private).toBe(true);
    photo.togglePrivate();
    expect(photo.Private).toBe(false);
    photo.togglePrivate();
    expect(photo.Private).toBe(true);
  });

  it("should mark photo as primary", async () => {
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
    const response = await photo.setPrimaryFile("fqbfk181n4ca5sud");
    expect(response.Files[0].Primary).toBe(true);
  });

  it("should unstack", async () => {
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
    const response = await photo.unstackFile("fqbfk181n4ca5sud");
    expect(response.success).toBe("ok");
  });

  it("should delete file", async () => {
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
    const response = await photo.deleteFile("fqbfk181n4ca5sud");
    expect(response.success).toBe("successfully deleted");
  });

  it("should add label", async () => {
    const values = {
      ID: 10,
      UID: "pqbemz8276mhtobh",
    };
    const photo = new Photo(values);
    const response = await photo.addLabel("Cat");
    expect(response.success).toBe("ok");
  });

  it("should activate label", async () => {
    const values = {
      ID: 10,
      UID: "pqbemz8276mhtobh",
    };
    const photo = new Photo(values);
    const response = await photo.activateLabel(12345);
    expect(response.success).toBe("ok");
  });

  it("should rename label", async () => {
    const values = {
      ID: 10,
      UID: "pqbemz8276mhtobh",
    };
    const photo = new Photo(values);
    const response = await photo.renameLabel(12345, "Sommer");
    expect(response.success).toBe("ok");
  });

  it("should remove label", async () => {
    const values = {
      ID: 10,
      UID: "pqbemz8276mhtobh",
    };
    const photo = new Photo(values);
    const response = await photo.removeLabel(12345);
    expect(response.success).toBe("ok");
  });

  it("should test update", async () => {
    const values = {
      ID: 10,
      UID: "pqbemz8276mhtobh",
      Lat: 1.1,
      Lng: 3.3,
      CameraID: 123,
      Title: "Test Titel",
      Caption: "Super nice video",
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
    photo.Caption = "New description";
    photo.Day = 21;
    photo.Country = "de";
    photo.CameraID = "newcameraid";
    photo.Details.Keywords = "newkeyword";
    photo.Details.Notes = "New Notes";
    photo.Details.Subject = "New Photo Subject";
    photo.Details.Artist = "New Artist";
    photo.Details.Copyright = "New Copyright";
    photo.Details.License = "New License";
    const response = await photo.update();
    expect(response.TitleSrc).toBe("manual");
    expect(photo.Title).toBe("New Title");
    expect(photo.Type).toBe("newtype");
    expect(photo.Caption).toBe("New description");
    expect(photo.Day).toBe(21);
    expect(photo.Country).toBe("de");
    expect(photo.CameraID).toBe("newcameraid");
    expect(photo.Details.Keywords).toBe("newkeyword");
    expect(photo.Details.Notes).toBe("New Notes");
    expect(photo.Details.Subject).toBe("New Photo Subject");
    expect(photo.Details.Artist).toBe("New Artist");
    expect(photo.Details.Copyright).toBe("New Copyright");
    expect(photo.Details.License).toBe("New License");
  });

  it("should test get Markers", () => {
    const values = {
      ID: 10,
      UID: "pqbemz8276mhtobh",
      Lat: 1.1,
      Lng: 3.3,
      CameraID: 123,
      Title: "Test Titel",
      Caption: "Super nice video",
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
    expect(result).toEqual([]);
    const values2 = {
      ID: 10,
      UID: "pqbemz8276mhtobh",
      Lat: 1.1,
      Lng: 3.3,
      CameraID: 123,
      Title: "Test Titel",
      Caption: "Super nice video",
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
    expect(result2.length).toBe(1);
    const result3 = photo2.getMarkers(false);
    expect(result3.length).toBe(2);
  });

  it("should determine if photo is a stack", () => {
    const values1 = {
      UID: "stack1",
      Type: "video",
      Files: [{ FileType: media.FormatJpeg }, { FileType: media.FormatJpeg }],
    };
    const photo1 = new Photo(values1);
    expect(photo1.isStack()).toBe(false);

    const values2 = { UID: "stack2", Type: media.Image, Files: [] };
    const photo2 = new Photo(values2);
    expect(photo2.isStack()).toBe(false);

    const values3 = { UID: "stack3", Type: media.Image, Files: [{ FileType: media.FormatJpeg }] };
    const photo3 = new Photo(values3);
    expect(photo3.isStack()).toBe(false);

    const values4 = { UID: "stack4", Type: media.Image, Files: [{ FileType: media.FormatJpeg }, { FileType: "raw" }] };
    const photo4 = new Photo(values4);
    expect(photo4.isStack()).toBe(false);

    const values5 = {
      UID: "stack5",
      Type: media.Image,
      Files: [{ FileType: media.FormatJpeg }, { FileType: media.FormatJpeg }],
    };
    const photo5 = new Photo(values5);
    expect(photo5.isStack()).toBe(true);

    const values6 = {
      UID: "stack6",
      Type: media.Image,
      Files: [{ FileType: media.FormatJpeg }, { FileType: "raw" }, { FileType: media.FormatJpeg }],
    };
    const photo6 = new Photo(values6);
    expect(photo6.isStack()).toBe(true);
  });

  it("should return the original file based on type", () => {
    const liveFiles = [
      { UID: "live_jpg", Name: "live.jpg", FileType: media.FormatJpeg, Root: "/", Primary: true },
      { UID: "live_mov", Name: "live.mov", FileType: "mov", MediaType: media.Video, Root: "/", Video: true },
      { UID: "live_sidecar", Name: "live.xmp", FileType: "xmp", Root: "sidecar", Sidecar: true },
    ];
    const photoLive = new Photo({ UID: "livePhoto", Type: media.Live, Files: liveFiles });
    expect(photoLive.originalFile().UID).toBe("live_mov");

    const videoFiles = [
      { UID: "video_jpg", Name: "video.jpg", FileType: media.FormatJpeg, Root: "/", Primary: true },
      {
        UID: "video_mp4",
        Name: "video.mp4",
        FileType: media.FormatMp4,
        MediaType: media.Video,
        Root: "/",
        Video: true,
      },
    ];
    const photoVideo = new Photo({ UID: "videoPhoto", Type: media.Video, Files: videoFiles });
    expect(photoVideo.originalFile().UID).toBe("video_mp4");

    const rawFiles = [
      { UID: "raw_jpg", Name: "raw.jpg", FileType: media.FormatJpeg, Root: "/", Primary: true },
      { UID: "raw_cr2", Name: "raw.cr2", FileType: "cr2", MediaType: media.Raw, Root: "/" },
    ];
    const photoRaw = new Photo({ UID: "rawPhoto", Type: media.Raw, Files: rawFiles });
    expect(photoRaw.originalFile().UID).toBe("raw_cr2");

    const animatedFiles = [
      {
        UID: "anim_gif",
        Name: "anim.gif",
        FileType: media.FormatGif,
        MediaType: media.Image,
        Root: "/",
        Primary: true,
      },
      { UID: "anim_sidecar", Name: "anim.xmp", FileType: "xmp", Root: "sidecar", Sidecar: true },
    ];
    const photoAnimated = new Photo({ UID: "animatedPhoto", Type: media.Animated, Files: animatedFiles });
    expect(photoAnimated.originalFile().UID).toBe("anim_gif");

    const otherFiles = [
      { UID: "other_jpg", Name: "other.jpg", FileType: media.FormatJpeg, Root: "/", Primary: true },
      { UID: "other_png", Name: "other.png", FileType: media.FormatPng, Root: "/" },
    ];
    const photoOther = new Photo({ UID: "otherPhoto", Type: media.Image, Files: otherFiles });
    expect(photoOther.originalFile().UID).toBe("other_png");

    const jpegFiles = [
      { UID: "jpeg_1", Name: "jpeg1.jpg", FileType: media.FormatJpeg, Root: "/" },
      { UID: "jpeg_2", Name: "jpeg2.jpg", FileType: media.FormatJpeg, Root: "/", Primary: true },
    ];
    const photoJpeg = new Photo({ UID: "jpegPhoto", Type: media.Image, Files: jpegFiles });
    expect(photoJpeg.originalFile().UID).toBe("jpeg_2");

    const singleFile = [
      { UID: "single_jpg", Name: "single.jpg", FileType: media.FormatJpeg, Root: "/", Primary: true },
    ];
    const photoSingle = new Photo({ UID: "singlePhoto", Type: media.Image, Files: singleFile });
    expect(photoSingle.originalFile().UID).toBe("single_jpg");

    const noFilesPhoto = new Photo({
      UID: "noFiles",
      Type: media.Image,
      OriginalName: "no_files_original.jpg",
      Name: "no_files_name.jpg",
      FileName: "no_files_filename.jpg",
    });
    expect(noFilesPhoto.originalFile()).toBe(noFilesPhoto);
  });

  it("should return the correct original name", () => {
    const files1 = [
      { UID: "f1_orig", Name: "file1.raw", OriginalName: "original_raw_name.raw", Root: "/", MediaType: media.Raw },
      {
        UID: "f1_jpg",
        Name: "file1.jpg",
        OriginalName: "original_jpg_name.jpg",
        Root: "/",
        FileType: media.FormatJpeg,
        Primary: true,
      },
    ];
    const photo1 = new Photo({ UID: "origName1", Type: media.Raw, Files: files1 });
    expect(photo1.getOriginalName()).toBe("original_raw_name.raw");

    const files2 = [
      { UID: "f2_orig", Name: "file2_actual.raw", Root: "/", MediaType: media.Raw },
      { UID: "f2_jpg", Name: "file2_actual.jpg", Root: "/", FileType: media.FormatJpeg, Primary: true },
    ];
    const photo2 = new Photo({ UID: "origName2", Type: media.Raw, Files: files2 });
    expect(photo2.getOriginalName()).toBe("file2_actual.raw");

    const photo3 = new Photo({
      UID: "origName3",
      Type: media.Image,
      OriginalName: "photo_original.jpg",
      Name: "photo_name.jpg",
      FileName: "photo_filename.jpg",
    });
    expect(photo3.getOriginalName()).toBe("photo_original.jpg");

    const photo4 = new Photo({
      UID: "origName4",
      Type: media.Image,
      Name: "photo_name.jpg",
      FileName: "photo_filename.jpg",
    });
    expect(photo4.getOriginalName()).toBe("photo_name.jpg");

    const photo5 = new Photo({ UID: "origName5", Type: media.Image, Name: "photo_name.jpg" });
    expect(photo5.getOriginalName()).toBe("photo_name.jpg");

    const photo6 = new Photo({ UID: "origName6", Type: media.Image });
    expect(photo6.getOriginalName()).toBe("Unknown");

    const files7 = [
      { UID: "f7_orig", Root: "/", MediaType: media.Raw },
      { UID: "f7_jpg", Name: "file7.jpg", Root: "/", FileType: media.FormatJpeg, Primary: true },
    ];
    const photo7 = new Photo({
      UID: "origName7",
      Type: media.Raw,
      Files: files7,
      OriginalName: "photo7_original.jpg",
      FileName: "photo7_filename.jpg",
      Name: "photo7_name.jpg",
    });
    expect(photo7.getOriginalName()).toBe("photo7_original.jpg");

    const files8 = [
      {
        UID: "f8_orig",
        Name: "some/path/file8.raw",
        OriginalName: "another/path/original_raw_name8.raw",
        Root: "/",
        MediaType: media.Raw,
      },
      { UID: "f8_jpg", Name: "file8.jpg", Root: "/", FileType: media.FormatJpeg, Primary: true },
    ];
    const photo8 = new Photo({ UID: "origName8", Type: media.Raw, Files: files8 });
    expect(photo8.getOriginalName()).toBe("original_raw_name8.raw");
  });
});
