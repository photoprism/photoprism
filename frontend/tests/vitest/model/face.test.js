import { describe, it, expect } from "vitest";
import "../fixtures";
import { Face, BatchSize } from "model/face";

describe("model/face", () => {
  it("should get face defaults", () => {
    const values = {};
    const face = new Face(values);
    const result = face.getDefaults();
    expect(result.ID).toBe("");
    expect(result.SampleRadius).toBe(0.0);
  });

  it("should get route view", () => {
    const values = { ID: "f123ghytrfggd", Samples: 5 };
    const face = new Face(values);
    const result = face.route("test");
    expect(result.name).toBe("test");
    expect(result.query.q).toBe("face:f123ghytrfggd");
  });

  it("should return classes", () => {
    const values = { ID: "f123ghytrfggd", Samples: 5 };
    const face = new Face(values);
    const result = face.classes(true);
    expect(result).toContain("is-face");
    expect(result).toContain("uid-f123ghytrfggd");
    expect(result).toContain("is-selected");
    expect(result).not.toContain("is-hidden");

    const result2 = face.classes(false);
    expect(result2).toContain("is-face");
    expect(result2).toContain("uid-f123ghytrfggd");
    expect(result2).not.toContain("is-selected");
    expect(result2).not.toContain("is-hidden");

    const values2 = { ID: "f123ghytrfggd", Samples: 5, Hidden: true };
    const face2 = new Face(values2);
    const result3 = face2.classes(true);
    expect(result3).toContain("is-face");
    expect(result3).toContain("uid-f123ghytrfggd");
    expect(result3).toContain("is-selected");
    expect(result3).toContain("is-hidden");
  });

  it("should get face entity name", () => {
    const values = { ID: "f123ghytrfggd", Samples: 5 };
    const face = new Face(values);
    const result = face.getEntityName();
    expect(result).toBe("f123ghytrfggd");
  });

  it("should get face title", () => {
    const values = { ID: "f123ghytrfggd", Samples: 5 };
    const face = new Face(values);
    const result = face.getTitle();
    expect(result).toBeUndefined();
  });

  it("should get thumbnail url", () => {
    const values = {
      ID: "f123ghytrfggd",
      Samples: 5,
      MarkerUID: "ABC123ghytr",
      FileUID: "fhjouohnnmnd",
      Name: "",
      Thumb: "7ca759a2b788cc5bcc08dbbce9854ff94a2f94d1",
    };

    const face = new Face(values);
    const result = face.thumbnailUrl("xyz");
    expect(result).toBe("/api/v1/t/7ca759a2b788cc5bcc08dbbce9854ff94a2f94d1/public/xyz");

    const values2 = {
      ID: "f123ghytrfggd",
      Samples: 5,
      Thumb: "7ca759a2b788cc5bcc08dbbce9854ff94a2f94d1",
    };
    const face2 = new Face(values2);
    const result2 = face2.thumbnailUrl();
    expect(result2).toBe("/api/v1/t/7ca759a2b788cc5bcc08dbbce9854ff94a2f94d1/public/tile_160");

    const values3 = {
      ID: "f123ghytrfggd",
      Samples: 5,
      Thumb: "",
    };
    const face3 = new Face(values3);
    const result3 = face3.thumbnailUrl("tile_240");
    expect(result3).toBe("/api/v1/svg/portrait");
  });

  it("should get date string", () => {
    const values = {
      ID: "f123ghytrfggd",
      Samples: 5,
      CreatedAt: "2012-07-08T14:45:39Z",
    };
    const face = new Face(values);
    const result = face.getDateString();
    expect(result.replaceAll("\u202f", " ")).toBe("Jul 8, 2012, 2:45 PM");
  });

  it("show and hide face", () => {
    const values = {
      ID: "f123ghytrfggd",
      Samples: 5,
      CreatedAt: "2012-07-08T14:45:39Z",
      Hidden: true,
    };
    const face = new Face(values);
    expect(face.Hidden).toBe(true);
    face.show();
    expect(face.Hidden).toBe(false);
    face.hide();
    expect(face.Hidden).toBe(true);
  });

  it("should toggle hidden", () => {
    const values = {
      ID: "f123ghytrfggd",
      Samples: 5,
      CreatedAt: "2012-07-08T14:45:39Z",
      Hidden: true,
    };
    const face = new Face(values);
    expect(face.Hidden).toBe(true);
    face.toggleHidden();
    expect(face.Hidden).toBe(false);
    face.toggleHidden();
    expect(face.Hidden).toBe(true);
  });

  it("should set name", async () => {
    const values = { ID: "f123ghytrfggd", Samples: 5, MarkerUID: "mDC123ghytr", Name: "Jane" };
    const face = new Face(values);

    const response1 = await face.setName("testname");
    expect(response1.Name).toBe("testname");

    const response2 = await face.setName("");
    expect(response2.Name).toBe("testname");
  });

  it("should return batch size", () => {
    expect(Face.batchSize()).toBe(BatchSize);
    Face.setBatchSize(30);
    expect(Face.batchSize()).toBe(30);
    Face.setBatchSize(BatchSize);
  });

  it("should get collection resource", () => {
    const result = Face.getCollectionResource();
    expect(result).toBe("faces");
  });

  it("should get model name", () => {
    const result = Face.getModelName();
    expect(result).toBe("Face");
  });
});
