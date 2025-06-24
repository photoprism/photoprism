import { describe, it, expect } from "vitest";
import "../fixtures";
import { Label, BatchSize } from "model/label";

describe("model/label", () => {
  it("should get route view", () => {
    const values = { ID: 5, UID: "ABC123", Name: "Black Cat", Slug: "black-cat" };
    const label = new Label(values);
    const result = label.route("test");
    expect(result.name).toBe("test");
    expect(result.query.q).toBe("label:black-cat");
  });

  it("should return batch size", () => {
    expect(Label.batchSize()).toBe(BatchSize);
    Label.setBatchSize(30);
    expect(Label.batchSize()).toBe(30);
    Label.setBatchSize(BatchSize);
  });

  it("should return classes", () => {
    const values = { ID: 5, UID: "ABC123", Name: "Black Cat", Slug: "black-cat", Favorite: true };
    const label = new Label(values);
    const result = label.classes(true);
    expect(result).toContain("is-label");
    expect(result).toContain("uid-ABC123");
    expect(result).toContain("is-selected");
    expect(result).toContain("is-favorite");
  });

  it("should get label entity name", () => {
    const values = { ID: 5, UID: "ABC123", Name: "Black Cat", Slug: "black-cat" };
    const label = new Label(values);
    const result = label.getEntityName();
    expect(result).toBe("black-cat");
  });

  it("should get label id", () => {
    const values = { ID: 5, UID: "ABC123", Name: "Black Cat", Slug: "black-cat" };
    const label = new Label(values);
    const result = label.getId();
    expect(result).toBe("ABC123");
  });

  it("should get label title", () => {
    const values = { ID: 5, UID: "ABC123", Name: "Black Cat", Slug: "black-cat" };
    const label = new Label(values);
    const result = label.getTitle();
    expect(result).toBe("Black Cat");
  });

  it("should get thumbnail url", () => {
    const values = {
      ID: 5,
      UID: "ABC123",
      Thumb: "c6b24d688564f7ddc7b245a414f003a8d8ff5a67",
      Name: "Black Cat",
      Slug: "black-cat",
    };
    const label = new Label(values);
    const result = label.thumbnailUrl("xyz");
    expect(result).toBe("/api/v1/t/c6b24d688564f7ddc7b245a414f003a8d8ff5a67/public/xyz");

    const values2 = {
      ID: 5,
      UID: "ABC123",
      Name: "Black Cat",
      Slug: "black-cat",
    };
    const label2 = new Label(values2);
    const result2 = label2.thumbnailUrl("xyz");
    expect(result2).toBe("/api/v1/labels/ABC123/t/public/xyz");

    const values3 = {
      ID: 5,
      Name: "Black Cat",
      Slug: "black-cat",
    };
    const label3 = new Label(values3);
    const result3 = label3.thumbnailUrl("xyz");
    expect(result3).toBe("/api/v1/svg/label");
  });

  it("should get date string", () => {
    const values = {
      ID: 5,
      UID: "ABC123",
      Name: "Black Cat",
      Slug: "black-cat",
      CreatedAt: "2012-07-08T14:45:39Z",
    };
    const label = new Label(values);
    const result = label.getDateString();
    expect(result.replaceAll("\u202f", " ")).toBe("Jul 8, 2012, 2:45 PM");
  });

  it("should get model name", () => {
    const result = Label.getModelName();
    expect(result).toBe("Label");
  });

  it("should get collection resource", () => {
    const result = Label.getCollectionResource();
    expect(result).toBe("labels");
  });

  it("should like label", () => {
    const values = { ID: 5, UID: "ABC123", Name: "Black Cat", Slug: "black-cat", Favorite: false };
    const label = new Label(values);
    expect(label.Favorite).toBe(false);
    label.like();
    expect(label.Favorite).toBe(true);
  });

  it("should unlike label", () => {
    const values = { ID: 5, UID: "ABC123", Name: "Black Cat", Slug: "black-cat", Favorite: true };
    const label = new Label(values);
    expect(label.Favorite).toBe(true);
    label.unlike();
    expect(label.Favorite).toBe(false);
  });

  it("should toggle like", () => {
    const values = { ID: 5, UID: "ABC123", Name: "Black Cat", Slug: "black-cat", Favorite: true };
    const label = new Label(values);
    expect(label.Favorite).toBe(true);
    label.toggleLike();
    expect(label.Favorite).toBe(false);
    label.toggleLike();
    expect(label.Favorite).toBe(true);
  });

  it("should get label defaults", () => {
    const values = { ID: 5, UID: "ABC123" };
    const label = new Label(values);
    const result = label.getDefaults();
    expect(result.ID).toBe(0);
  });
});
