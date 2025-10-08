import { describe, it, expect } from "vitest";
import "../fixtures";
import Collection from "model/collection";
import RestModel from "model/rest";
import { Album } from "model/album";
import { Label } from "model/label";
import { Photo } from "model/photo";

describe("model/collection", () => {
  it("extends RestModel", () => {
    const collection = new Collection({ UID: "C001", Title: "Sample" });
    expect(collection).toBeInstanceOf(RestModel);
    expect(collection.UID).toBe("C001");
    expect(collection.Title).toBe("Sample");
  });

  it("treats albums as collections", () => {
    const album = new Album({ UID: "A001" });
    expect(album).toBeInstanceOf(Collection);
  });

  it("treats labels as collections", () => {
    const label = new Label({ UID: "L001" });
    expect(label).toBeInstanceOf(Collection);
  });

  it("does not treat non-collection models as collections", () => {
    const photo = new Photo({ UID: "P001" });
    expect(photo).not.toBeInstanceOf(Collection);
  });

  it("sets album cover through collection helper", async () => {
    const hash = "0123456789abcdef0123456789abcdef01234567";
    const album = new Album({ UID: "ACVR1" });
    await album.setCover(hash);
    expect(album.Thumb).toBe(hash);
    expect(album.ThumbSrc).toBe("manual");
  });

  it("sets label cover through collection helper", async () => {
    const hash = "89abcdef012345670123456789abcdef01234567";
    const label = new Label({ UID: "LCVR1" });
    await label.setCover(hash);
    expect(label.Thumb).toBe(hash);
    expect(label.ThumbSrc).toBe("manual");
  });

  it("ignores invalid cover hashes", () => {
    const album = new Album({ UID: "A0002" });
    const result = album.setCover("short");
    expect(result).toBeUndefined();
  });
});
