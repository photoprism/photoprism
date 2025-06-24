import { describe, it, expect } from "vitest";
import "../fixtures";
import File from "model/file";

describe("model/file", () => {
  it("should return classes", () => {
    const values = {
      InstanceID: 5,
      UID: "ABC123",
      Name: "1/2/IMG123.jpg",
      Primary: true,
      Sidecar: true,
      Video: true,
    };
    const file = new File(values);
    const result = file.classes(true);
    expect(result).toContain("is-file");
    expect(result).toContain("uid-ABC123");
    expect(result).toContain("is-primary");
    expect(result).toContain("is-sidecar");
    expect(result).toContain("is-video");
    expect(result).toContain("is-selected");
  });

  it("should get file defaults", () => {
    const values = {
      InstanceID: 5,
      UID: "ABC123",
    };
    const file = new File(values);
    const result = file.getDefaults();
    expect(result.UID).toBe("");
    expect(result.Size).toBe(0);
  });

  it("should get file base name", () => {
    const values = {
      InstanceID: 5,
      UID: "ABC123",
      Name: "1/2/IMG123.jpg",
    };
    const file = new File(values);
    const result = file.baseName();
    expect(result).toBe("IMG123.jpg");
    const result2 = file.baseName(8);
    expect(result2).toBe("IMG123.…");
  });

  it("should return true", () => {
    const values = {
      InstanceID: 5,
      UID: "ABC123",
      Name: "1/2/IMG123.jpg",
    };
    const file = new File(values);
    expect(file.isFile()).toBe(true);
  });

  it("should return entity name", () => {
    const values = {
      InstanceID: 5,
      UID: "ABC123",
      Root: "",
      Name: "1/2/IMG123.jpg",
    };
    const file = new File(values);
    expect(file.getEntityName()).toBe("/1/2/IMG123.jpg");
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
    expect(file.thumbnailUrl("tile_224")).toBe("/api/v1/t/54ghtfd/public/tile_224");

    const values2 = {
      InstanceID: 5,
      UID: "ABC123",
      Name: "1/2/IMG123.jpg",
      Error: true,
    };
    const file2 = new File(values2);
    expect(file2.thumbnailUrl("tile_224")).toBe("/api/v1/svg/broken");

    const values3 = {
      InstanceID: 5,
      UID: "ABC123",
      Hash: "bd66bd2c304f45f6c160df375f34b49eb7aef321",
      Name: "1/2/IMG123.jpg",
      FileType: "raw",
    };
    const file3 = new File(values3);
    expect(file3.thumbnailUrl("tile_224")).toBe("/api/v1/t/bd66bd2c304f45f6c160df375f34b49eb7aef321/public/tile_224");

    const values4 = {
      InstanceID: 5,
      UID: "ABC123",
      Hash: "0e437256ec20da874318b64027750b320548378c",
      FileType: "jpg",
      Name: "1/2/IMG123.jpg",
      Sidecar: true,
    };
    const file4 = new File(values4);
    expect(file4.thumbnailUrl("tile_224")).toBe("/api/v1/svg/file");
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
    expect(file.getDownloadUrl("abc")).toBe("/api/v1/dl/54ghtfd?t=2lbh9x09");
  });

  it("should not download as hash is missing", () => {
    const values = {
      InstanceID: 5,
      UID: "ABC123",
      FileType: "jpg",
      Name: "1/2/IMG123.jpg",
    };
    const file = new File(values);
    expect(file.download()).toBeUndefined();
  });

  it("should calculate size", () => {
    const values = {
      InstanceID: 5,
      UID: "ABC123",
      Hash: "54ghtfd",
      FileType: "jpg",
      Width: 500,
      Height: 700,
      Name: "1/2/IMG123.jpg",
    };
    const file = new File(values);
    expect(file.calculateSize(600, 800).width).toBe(500);
    expect(file.calculateSize(600, 800).height).toBe(700);

    const values2 = {
      InstanceID: 5,
      UID: "ABC123",
      Hash: "54ghtfd",
      FileType: "jpg",
      Width: 900,
      Height: 850,
      Name: "1/2/IMG123.jpg",
    };
    const file2 = new File(values2);
    expect(file2.calculateSize(600, 800).width).toBe(600);
    expect(file2.calculateSize(600, 800).height).toBe(567);

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
    expect(file3.calculateSize(900, 450).width).toBe(398);
    expect(file3.calculateSize(900, 450).height).toBe(450);
  });

  it("should get date string", () => {
    const values = {
      InstanceID: 5,
      UID: "ABC123",
      Hash: "54ghtfd",
      FileType: "jpg",
      Name: "1/2/IMG123.jpg",
      CreatedAt: "2012-07-08T14:45:39Z",
      UpdatedAt: "2012-07-08T14:45:39Z",
    };
    const file = new File(values);
    expect(file.getDateString().replaceAll("\u202f", " ")).toBe("Jul 8, 2012, 2:45 PM");
  });

  it("should get info", () => {
    const values = {
      InstanceID: 5,
      UID: "ABC123",
      Hash: "54ghtfd",
      FileType: "jpg",
      Name: "1/2/IMG123.jpg",
      CreatedAt: "2012-07-08T14:45:39Z",
      UpdatedAt: "2012-07-08T14:45:39Z",
    };
    const file = new File(values);
    expect(file.getInfo()).toBe("JPG");

    const values2 = {
      InstanceID: 6,
      UID: "ABC124",
      Hash: "54ghtfd",
      FileType: "mp4",
      Duration: 8009,
      FPS: 60,
      Name: "1/2/IMG123.mp4",
      CreatedAt: "2012-07-08T14:45:39Z",
      UpdatedAt: "2012-07-08T14:45:39Z",
    };
    const file2 = new File(values2);
    expect(file2.getInfo()).toBe("MP4, 8µs, 60.0 FPS");
  });

  it("should return storage location", () => {
    const values = {
      InstanceID: 5,
      UID: "ABC123",
      Hash: "54ghtfd",
      FileType: "jpg",
      Name: "1/2/IMG123.jpg",
      Root: "sidecar",
      CreatedAt: "2012-07-08T14:45:39Z",
      UpdatedAt: "2012-07-08T14:45:39Z",
    };
    const file = new File(values);
    expect(file.storageInfo()).toBe("Sidecar");

    const values2 = {
      InstanceID: 6,
      UID: "ABC124",
      Hash: "54ghtfd",
      FileType: "mp4",
      Duration: 8009,
      FPS: 60,
      Root: "/",
      Name: "1/2/IMG123.mp4",
      CreatedAt: "2012-07-08T14:45:39Z",
      UpdatedAt: "2012-07-08T14:45:39Z",
    };
    const file2 = new File(values2);
    expect(file2.storageInfo()).toBe("Originals");

    const values3 = {
      InstanceID: 6,
      UID: "ABC124",
      Hash: "54ghtfd",
      FileType: "mp4",
      Duration: 8009,
      FPS: 60,
      Root: "",
      Name: "1/2/IMG123.mp4",
      CreatedAt: "2012-07-08T14:45:39Z",
      UpdatedAt: "2012-07-08T14:45:39Z",
    };
    const file3 = new File(values3);
    expect(file3.storageInfo()).toBe("");
  });

  it("should return whether file is animated", () => {
    const values = {
      InstanceID: 5,
      UID: "ABC123",
      MediaType: "image",
      Duration: 500,
    };
    const file = new File(values);
    expect(file.isAnimated()).toBe(true);
  });

  it("should get type info", () => {
    const values = {
      InstanceID: 5,
      UID: "ABC123",
      Hash: "54ghtfd",
      FileType: "jpg",
      Primary: true,
      Name: "1/2/IMG123.jpg",
      CreatedAt: "2012-07-08T14:45:39Z",
      UpdatedAt: "2012-07-08T14:45:39Z",
    };
    const file = new File(values);
    expect(file.typeInfo()).toBe("Image");

    const values2 = {
      InstanceID: 5,
      UID: "ABC123",
      Hash: "54ghtfd",
      FileType: "mp4",
      Duration: 8009,
      FPS: 60,
      Name: "1/2/IMG123.mp4",
      Video: true,
      CreatedAt: "2012-07-08T14:45:39Z",
      UpdatedAt: "2012-07-08T14:45:39Z",
    };
    const file2 = new File(values2);
    expect(file2.typeInfo()).toBe("Video");

    const values3 = {
      InstanceID: 5,
      UID: "ABC123",
      Hash: "54ghtfd",
      FileType: "jpg",
      Name: "1/2/IMG123.jpg",
      Sidecar: true,
      CreatedAt: "2012-07-08T14:45:39Z",
      UpdatedAt: "2012-07-08T14:45:39Z",
    };
    const file3 = new File(values3);
    expect(file3.typeInfo()).toBe("Sidecar JPEG");

    const values4 = {
      InstanceID: 5,
      UID: "ABC123",
      Hash: "54ghtfd",
      FileType: "gif",
      MediaType: "image",
      Duration: 8009,
      Name: "1/2/IMG123.jpg",
      Sidecar: true,
      CreatedAt: "2012-07-08T14:45:39Z",
      UpdatedAt: "2012-07-08T14:45:39Z",
    };
    const file4 = new File(values4);
    expect(file4.typeInfo()).toBe("Sidecar GIF Image");

    const values5 = {
      InstanceID: 5,
      UID: "ABC123",
      Hash: "54ghtfd",
      FileType: "svg",
      MediaType: "vector",
      Name: "1/2/IMG123.svg",
      CreatedAt: "2012-07-08T14:45:39Z",
      UpdatedAt: "2012-07-08T14:45:39Z",
    };
    const file5 = new File(values5);
    expect(file5.typeInfo()).toBe("SVG");
  });

  it("should get size info", () => {
    const values = {
      InstanceID: 5,
      UID: "ABC123",
      Hash: "54ghtfd",
      FileType: "jpg",
      Size: 8009,
      Name: "1/2/IMG123.jpg",
      CreatedAt: "2012-07-08T14:45:39Z",
      UpdatedAt: "2012-07-08T14:45:39Z",
    };
    const file = new File(values);
    expect(file.sizeInfo()).toBe("8 KB");

    const values2 = {
      InstanceID: 5,
      UID: "ABC123",
      Hash: "54ghtfd",
      FileType: "jpg",
      Size: 8009999987,
      Name: "1/2/IMG123.jpg",
      CreatedAt: "2012-07-08T14:45:39Z",
      UpdatedAt: "2012-07-08T14:45:39Z",
    };
    const file2 = new File(values2);
    expect(file2.sizeInfo()).toBe("7.5 GB");

    const values3 = {
      InstanceID: 5,
      UID: "ABC123",
      Hash: "54ghtfd",
      FileType: "jpg",
      Size: 8009999987,
      Name: "1/2/IMG123.jpg",
      Width: 500,
      Height: 800,
      CreatedAt: "2012-07-08T14:45:39Z",
      UpdatedAt: "2012-07-08T14:45:39Z",
    };
    const file3 = new File(values3);
    expect(file3.sizeInfo()).toBe("500 × 800, 7.5 GB");
  });

  it("should like file", () => {
    const values = {
      InstanceID: 5,
      UID: "ABC123",
      Hash: "54ghtfd",
      FileType: "jpg",
      Duration: 8009,
      Favorite: false,
      Name: "1/2/IMG123.jpg",
      CreatedAt: "2012-07-08T14:45:39Z",
      UpdatedAt: "2012-07-08T14:45:39Z",
    };
    const file = new File(values);
    expect(file.Favorite).toBe(false);
    file.like();
    expect(file.Favorite).toBe(true);
  });

  it("should unlike file", () => {
    const values = {
      InstanceID: 5,
      UID: "ABC123",
      Hash: "54ghtfd",
      FileType: "jpg",
      Duration: 8009,
      Favorite: true,
      Name: "1/2/IMG123.jpg",
      CreatedAt: "2012-07-08T14:45:39Z",
      UpdatedAt: "2012-07-08T14:45:39Z",
    };
    const file = new File(values);
    expect(file.Favorite).toBe(true);
    file.unlike();
    expect(file.Favorite).toBe(false);
  });

  it("should toggle like", () => {
    const values = {
      InstanceID: 5,
      UID: "ABC123",
      Hash: "54ghtfd",
      FileType: "jpg",
      Duration: 8009,
      Favorite: true,
      Name: "1/2/IMG123.jpg",
      CreatedAt: "2012-07-08T14:45:39Z",
      UpdatedAt: "2012-07-08T14:45:39Z",
    };
    const file = new File(values);
    expect(file.Favorite).toBe(true);
    file.toggleLike();
    expect(file.Favorite).toBe(false);
    file.toggleLike();
    expect(file.Favorite).toBe(true);
  });

  it("should get photo resource", () => {
    const values = {
      InstanceID: 5,
      PhotoUID: "bgad457",
      UID: "ABC123",
      Hash: "54ghtfd",
      FileType: "jpg",
      Duration: 8009,
      Favorite: true,
      Name: "1/2/IMG123.jpg",
      CreatedAt: "2012-07-08T14:45:39Z",
      UpdatedAt: "2012-07-08T14:45:39Z",
    };
    const file = new File(values);
    expect(file.getPhotoResource()).toBe("photos/bgad457");
  });

  it("should get collection resource", () => {
    const result = File.getCollectionResource();
    expect(result).toBe("files");
  });

  it("should get model name", () => {
    const result = File.getModelName();
    expect(result).toBe("File");
  });
});
