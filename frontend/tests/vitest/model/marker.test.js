import { describe, it, expect, beforeEach } from "vitest";
import { Marker, BatchSize } from "model/marker";
import { setupMarkerMocks } from "../fixtures";

beforeEach(() => {
  setupMarkerMocks();
});

describe("model/marker", () => {
  it("should get marker defaults", () => {
    const values = { FileUID: "fghjojp" };
    const marker = new Marker(values);
    const result = marker.getDefaults();
    expect(result.UID).toBe("");
    expect(result.FileUID).toBe("");
  });

  it("should get route view", () => {
    const values = { UID: "ABC123ghytr", FileUID: "fhjouohnnmnd", Type: "face", Src: "image" };
    const marker = new Marker(values);
    const result = marker.route("test");
    expect(result.name).toBe("test");
    expect(result.query.q).toBe("marker:ABC123ghytr");
  });

  it("should return classes", () => {
    const values = { UID: "ABC123ghytr", FileUID: "fhjouohnnmnd", Type: "face", Src: "image" };
    const marker = new Marker(values);
    const result = marker.classes(true);
    expect(result).toContain("is-marker");
    expect(result).toContain("uid-ABC123ghytr");
    expect(result).toContain("is-selected");
    expect(result).not.toContain("is-review");
    expect(result).not.toContain("is-invalid");

    const result2 = marker.classes(false);
    expect(result2).toContain("is-marker");
    expect(result2).toContain("uid-ABC123ghytr");
    expect(result2).not.toContain("is-selected");
    expect(result2).not.toContain("is-review");
    expect(result2).not.toContain("is-invalid");

    const values2 = {
      UID: "mBC123ghytr",
      FileUID: "fhjouohnnmnd",
      Type: "face",
      Src: "image",
      Invalid: true,
      Review: true,
    };
    const marker2 = new Marker(values2);
    const result3 = marker2.classes(true);
    expect(result3).toContain("is-marker");
    expect(result3).toContain("uid-mBC123ghytr");
    expect(result3).toContain("is-selected");
    expect(result3).toContain("is-review");
    expect(result3).toContain("is-invalid");
  });

  it("should get marker entity name", () => {
    const values = {
      UID: "ABC123ghytr",
      FileUID: "fhjouohnnmnd",
      Type: "face",
      Src: "image",
      Name: "test",
    };
    const marker = new Marker(values);
    const result = marker.getEntityName();
    expect(result).toBe("test");
  });

  it("should get marker title", () => {
    const values = {
      UID: "ABC123ghytr",
      FileUID: "fhjouohnnmnd",
      Type: "face",
      Src: "image",
      Name: "test",
    };
    const marker = new Marker(values);
    const result = marker.getTitle();
    expect(result).toBe("test");
  });

  it("should get thumbnail url", () => {
    const values = { UID: "ABC123ghytr", FileUID: "fhjouohnnmnd", Type: "face", Src: "image" };
    const marker = new Marker(values);
    const result = marker.thumbnailUrl("xyz");
    expect(result).toBe("/api/v1/svg/portrait");

    const values2 = {
      UID: "ABC123ghytr",
      FileUID: "fhjouohnnmnd",
      Type: "face",
      Src: "image",
      Thumb: "nicethumbuid",
    };
    const marker2 = new Marker(values2);
    const result2 = marker2.thumbnailUrl();
    expect(result2).toBe("/api/v1/t/nicethumbuid/public/tile_160");
  });

  it("should get date string", () => {
    const values = {
      UID: "ABC123ghytr",
      FileUID: "fhjouohnnmnd",
      Type: "face",
      Src: "image",
      CreatedAt: "2012-07-08T14:45:39Z",
    };
    const marker = new Marker(values);
    const result = marker.getDateString();
    expect(result).toBe("2023-10-01 10:00:00");
  });

  it("should approve marker", () => {
    const values = {
      UID: "mBC123ghytr",
      FileUID: "fhjouohnnmnd",
      Type: "face",
      Src: "image",
      Invalid: true,
      Review: true,
    };
    const marker = new Marker(values);
    expect(marker.Review).toBe(true);
    expect(marker.Invalid).toBe(true);
    marker.approve();
    expect(marker.Review).toBe(false);
    expect(marker.Invalid).toBe(false);
  });

  it("should reject marker", () => {
    const values = {
      UID: "mCC123ghytr",
      FileUID: "fhjouohnnmnd",
      Type: "face",
      Src: "image",
      Invalid: false,
      Review: true,
    };
    const marker = new Marker(values);
    expect(marker.Review).toBe(true);
    expect(marker.Invalid).toBe(false);
    marker.reject();
    expect(marker.Review).toBe(false);
    expect(marker.Invalid).toBe(true);
  });

  it("should rename marker", async () => {
    const values = {
      UID: "mDC123ghytr",
      FileUID: "fhjouohnnmnd",
      Type: "face",
      Src: "image",
      Subject: "skhljkpigh",
      Name: "",
      SubjSrc: "manual",
    };
    const marker = new Marker(values);
    expect(marker.Name).toBe("");
    marker.setName();
    expect(marker.Name).toBe("");

    const values2 = {
      UID: "mDC123ghytr",
      FileUID: "fhjouohnnmnd",
      Type: "face",
      Src: "image",
      Subject: "skhljkpigh",
      Name: "testname",
      SubjSrc: "manual",
    };
    const marker2 = new Marker(values2);
    expect(marker2.Name).toBe("testname");

    const response = await marker2.setName();
    expect(response.success).toBe("ok");
  });

  it("should clear subject", async () => {
    const values = {
      UID: "mEC123ghytr",
      FileUID: "fhjouohnnmnd",
      Type: "face",
      Src: "image",
      Subject: "skhljkpigh",
      Name: "testname",
      SubjSrc: "manual",
    };
    const marker = new Marker(values);
    const response = await marker.clearSubject();
    expect(response.success).toBe("ok");
  });

  it("should return batch size", () => {
    expect(Marker.batchSize()).toBe(BatchSize);
    Marker.setBatchSize(30);
    expect(Marker.batchSize()).toBe(30);
    Marker.setBatchSize(BatchSize);
  });

  it("should get collection resource", () => {
    const result = Marker.getCollectionResource();
    expect(result).toBe("markers");
  });

  it("should get model name", () => {
    const result = Marker.getModelName();
    expect(result).toBe("Marker");
  });
});
