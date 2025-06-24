import { describe, it, expect } from "vitest";
import "../fixtures";
import { Subject, BatchSize } from "model/subject";

describe("model/subject", () => {
  it("should get face defaults", () => {
    const values = {};
    const subject = new Subject(values);
    const result = subject.getDefaults();
    expect(result.UID).toBe("");
    expect(result.Favorite).toBe(false);
  });

  it("should get route view", () => {
    const values = { UID: "s123ghytrfggd", Type: "person", Src: "manual" };
    const subject = new Subject(values);
    const result = subject.route("test");
    expect(result.name).toBe("test");
    expect(result.query.q).toBe("subject:s123ghytrfggd");
    const values2 = {
      UID: "s123ghytrfggd",
      Type: "person",
      Src: "manual",
      Name: "Jane Doe",
      Slug: "jane-doe",
    };
    const subject2 = new Subject(values2);
    const result2 = subject2.route("test");
    expect(result2.name).toBe("test");
    expect(result2.query.q).toBe("person:jane-doe");
  });

  it("should return classes", () => {
    const values = {
      UID: "s123ghytrfggd",
      Type: "person",
      Src: "manual",
      Name: "Jane Doe",
      Slug: "jane-doe",
      Favorite: false,
      Excluded: true,
      Private: true,
      Hidden: true,
    };
    const subject = new Subject(values);
    const result = subject.classes(true);
    expect(result).toContain("is-subject");
    expect(result).toContain("uid-s123ghytrfggd");
    expect(result).toContain("is-selected");
    expect(result).not.toContain("is-favorite");
    expect(result).toContain("is-private");
    expect(result).toContain("is-excluded");
    expect(result).toContain("is-hidden");
    const values2 = {
      UID: "s123ghytrfggd",
      Type: "person",
      Src: "manual",
      Name: "Jane Doe",
      Slug: "jane-doe",
      Favorite: true,
      Excluded: false,
      Private: false,
    };
    const subject2 = new Subject(values2);
    const result2 = subject2.classes(false);
    expect(result2).toContain("is-subject");
    expect(result2).toContain("uid-s123ghytrfggd");
    expect(result2).not.toContain("is-selected");
    expect(result2).toContain("is-favorite");
    expect(result2).not.toContain("is-private");
    expect(result2).not.toContain("is-excluded");
  });

  it("should get subject entity name", () => {
    const values = {
      UID: "s123ghytrfggd",
      Type: "person",
      Src: "manual",
      Name: "Jane Doe",
      Slug: "jane-doe",
      Favorite: false,
      Excluded: true,
      Private: true,
    };
    const subject = new Subject(values);
    const result = subject.getEntityName();
    expect(result).toBe("jane-doe");
  });

  it("should get subject title", () => {
    const values = {
      UID: "s123ghytrfggd",
      Type: "person",
      Src: "manual",
      Name: "Jane Doe",
      Slug: "jane-doe",
      Favorite: false,
      Excluded: true,
      Private: true,
    };
    const subject = new Subject(values);
    const result = subject.getTitle();
    expect(result).toBe("Jane Doe");
  });

  it("should get thumbnail url", () => {
    const values = {
      UID: "s123ghytrfggd",
      Type: "person",
      Src: "manual",
      Name: "Jane Doe",
      Slug: "jane-doe",
      Favorite: false,
      Excluded: true,
      Private: true,
      Thumb: "nicethumb",
    };
    const subject = new Subject(values);
    const result = subject.thumbnailUrl("xyz");
    expect(result).toBe("/api/v1/t/nicethumb/public/xyz");
    const result2 = subject.thumbnailUrl();
    expect(result2).toBe("/api/v1/t/nicethumb/public/tile_160");
    const values2 = {
      UID: "s123ghytrfggd",
      Type: "person",
      Src: "manual",
      Name: "Jane Doe",
      Slug: "jane-doe",
      Favorite: false,
      Excluded: true,
      Private: true,
    };
    const subject2 = new Subject(values2);
    const result3 = subject2.thumbnailUrl("xyz");
    expect(result3).toBe("/api/v1/svg/portrait");
  });

  it("should get date string", () => {
    const values = {
      UID: "s123ghytrfggd",
      Type: "person",
      Src: "manual",
      Name: "Jane Doe",
      Slug: "jane-doe",
      Favorite: false,
      Excluded: true,
      Private: true,
      Thumb: "nicethumb",
      CreatedAt: "2012-07-08T14:45:39Z",
    };
    const subject = new Subject(values);
    const result = subject.getDateString();
    expect(result.replaceAll("\u202f", " ")).toBe("Jul 8, 2012, 2:45 PM");
  });

  it("should like subject", () => {
    const values = {
      UID: "s123ghytrfggd",
      Type: "person",
      Src: "manual",
      Name: "Jane Doe",
      Slug: "jane-doe",
      Favorite: false,
    };
    const subject = new Subject(values);
    expect(subject.Favorite).toBe(false);
    subject.like();
    expect(subject.Favorite).toBe(true);
  });

  it("should unlike subject", () => {
    const values = {
      UID: "s123ghytrfggd",
      Type: "person",
      Src: "manual",
      Name: "Jane Doe",
      Slug: "jane-doe",
      Favorite: true,
    };
    const subject = new Subject(values);
    expect(subject.Favorite).toBe(true);
    subject.unlike();
    expect(subject.Favorite).toBe(false);
  });

  it("should toggle like", () => {
    const values = {
      UID: "s123ghytrfggd",
      Type: "person",
      Src: "manual",
      Name: "Jane Doe",
      Slug: "jane-doe",
      Favorite: true,
    };
    const subject = new Subject(values);
    expect(subject.Favorite).toBe(true);
    subject.toggleLike();
    expect(subject.Favorite).toBe(false);
    subject.toggleLike();
    expect(subject.Favorite).toBe(true);
  });

  it("show and hide subject", () => {
    const values = {
      UID: "s123ghytrfggd",
      Type: "person",
      Src: "manual",
      Name: "Jane Doe",
      Slug: "jane-doe",
      Hidden: true,
    };
    const subject = new Subject(values);
    expect(subject.Hidden).toBe(true);
    subject.show();
    expect(subject.Hidden).toBe(false);
    subject.hide();
    expect(subject.Hidden).toBe(true);
  });

  it("should toggle hidden", () => {
    const values = {
      UID: "s123ghytrfggd",
      Type: "person",
      Src: "manual",
      Name: "Jane Doe",
      Slug: "jane-doe",
      Hidden: true,
    };
    const subject = new Subject(values);
    expect(subject.Hidden).toBe(true);
    subject.toggleHidden();
    expect(subject.Hidden).toBe(false);
    subject.toggleHidden();
    expect(subject.Hidden).toBe(true);
  });

  it("should return batch size", () => {
    expect(Subject.batchSize()).toBe(BatchSize);
    Subject.setBatchSize(30);
    expect(Subject.batchSize()).toBe(30);
    Subject.setBatchSize(BatchSize);
  });

  it("should get collection resource", () => {
    const result = Subject.getCollectionResource();
    expect(result).toBe("subjects");
  });

  it("should get model name", () => {
    const result = Subject.getModelName();
    expect(result).toBe("Person");
  });
});
