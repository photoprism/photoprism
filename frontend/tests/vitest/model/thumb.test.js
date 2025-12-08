import { describe, it, expect } from "vitest";
import "../fixtures";
import Thumb from "model/thumb";
import Photo from "model/photo";
import File from "model/file";

describe("model/thumb", () => {
  it("should get thumb defaults", () => {
    const values = {
      UID: "55",
      Title: "",
      TakenAtLocal: "",
      Caption: "",
      Favorite: false,
      Playable: false,
      Width: 0,
      Height: 0,
      DownloadUrl: "",
    };
    const thumb = new Thumb(values);
    const result = thumb.getDefaults();
    expect(result.UID).toBe("");
  });

  it("should get id", () => {
    const values = {
      UID: "55",
    };
    const thumb = new Thumb(values);
    expect(thumb.getId()).toBe("55");
  });

  it("should return hasId", () => {
    const values = {
      UID: "55",
    };
    const thumb = new Thumb(values);
    expect(thumb.hasId()).toBe(true);

    const values2 = {
      Title: "",
    };
    const thumb2 = new Thumb(values2);
    expect(thumb2.hasId()).toBe(false);
  });

  it("should toggle like", () => {
    const values = {
      UID: "55",
      Title: "",
      TakenAtLocal: "",
      Caption: "",
      Favorite: true,
      Playable: false,
      Width: 0,
      Height: 0,
      DownloadUrl: "",
    };
    const thumb = new Thumb(values);
    expect(thumb.Favorite).toBe(true);
    thumb.toggleLike();
    expect(thumb.Favorite).toBe(false);
    thumb.toggleLike();
    expect(thumb.Favorite).toBe(true);
  });

  it("should return thumb not found", () => {
    const result = Thumb.notFound();
    expect(result.UID).toBe("");
    expect(result.Favorite).toBe(false);
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
      Caption: "Nice description",
      Favorite: true,
    };
    const photo = new Photo(values2);
    const result = Thumb.fromFile(photo, file);
    expect(result.UID).toBe("5");
    expect(result.Caption).toBe("Nice description");
    expect(result.Width).toBe(500);
    const result2 = Thumb.fromFile();
    expect(result2.UID).toBe("");
  });

  it("should test from files", () => {
    const values2 = {
      UID: "5",
      Title: "Crazy Cat",
      TakenAt: "2012-07-08T14:45:39Z",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      Caption: "Nice description",
      Favorite: true,
    };
    const photo = new Photo(values2);

    const values3 = {
      UID: "5",
      Title: "Crazy Cat",
      TakenAt: "2012-07-08T14:45:39Z",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      Caption: "Nice description",
      Favorite: true,
    };
    const photo2 = new Photo(values3);
    const Photos = [photo, photo2];
    const result = Thumb.fromFiles(Photos);
    expect(result.length).toBe(0);
    const values4 = {
      ID: 8,
      UID: "ABC123",
      Caption: "Nice description 2",
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
    expect(result2[0].UID).toBe("ABC123");
    expect(result2[0].Caption).toBe("Nice description 2");
    expect(result2[0].Width).toBe(500);
    expect(result2.length).toBe(1);
    const values5 = {
      ID: 8,
      UID: "ABC123",
      Caption: "Nice description 2",
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
    expect(result3.length).toBe(1);
    expect(result3[0].UID).toBe("ABC123");
    expect(result3[0].Caption).toBe("Nice description 2");
    expect(result3[0].Width).toBe(500);
  });

  it("should test from files", () => {
    const Photos = [];
    const result = Thumb.fromFiles(Photos);
    expect(result).toEqual([]);
  });

  it("should test from photo", () => {
    const values = {
      ID: 8,
      UID: "ABC123",
      Caption: "Nice description 3",
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
    expect(result.UID).toBe("ABC123");
    expect(result.Caption).toBe("Nice description 3");
    expect(result.Width).toBe(500);
    const values3 = {
      ID: 8,
      UID: "ABC124",
      Caption: "Nice description 3",
    };
    const photo3 = new Photo(values3);
    const result2 = Thumb.fromPhoto(photo3);
    expect(result2.UID).toBe("");
    const values2 = {
      ID: 8,
      UID: "ABC123",
      Title: "Crazy Cat",
      TakenAt: "2012-07-08T14:45:39Z",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      Caption: "Nice description",
      Favorite: true,
      Hash: "xdf45m",
    };
    const photo2 = new Photo(values2);
    const result3 = Thumb.fromPhoto(photo2);
    expect(result3.UID).toBe("ABC123");
    expect(result3.Title).toBe("Crazy Cat");
    expect(result3.Caption).toBe("Nice description");
  });

  it("should test from photos", () => {
    const values = {
      ID: 8,
      UID: "ABC123",
      Caption: "Nice description 3",
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
    expect(result[0].UID).toBe("ABC123");
    expect(result[0].Caption).toBe("Nice description 3");
    expect(result[0].Width).toBe(500);
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
    expect(Thumb.downloadUrl(file)).toBe("/api/v1/dl/54ghtfd?t=2lbh9x09");
    const values2 = {
      InstanceID: 5,
      UID: "ABC123",
      FileType: "jpg",
      Name: "1/2/IMG123.jpg",
    };
    const file2 = new File(values2);
    expect(Thumb.downloadUrl(file2)).toBe("");
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
    expect(Thumb.thumbnailUrl(file, "abc")).toBe("/api/v1/t/54ghtfd/public/abc");
    const values2 = {
      InstanceID: 5,
      UID: "ABC123",
      Name: "1/2/IMG123.jpg",
    };
    const file2 = new File(values2);
    expect(Thumb.thumbnailUrl(file2, "bcd")).toBe("/static/img/404.jpg");
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
    const result = Thumb.calculateSize(file, 600, 800);
    expect(result.width).toBe(600);
    expect(result.height).toBe(567);
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
    expect(result2.width).toBe(398);
    expect(result2.height).toBe(450);
    const result4 = Thumb.calculateSize(file3, 900, 950);
    expect(result4.width).toBe(750);
    expect(result4.height).toBe(850);
  });
});
