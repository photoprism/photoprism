import "../fixtures";
import Label from "model/label";

let chai = require("chai/chai");
let assert = chai.assert;

describe("model/label", () => {
  it("should get route view", () => {
    const values = { ID: 5, UID: "ABC123", Name: "Black Cat", Slug: "black-cat" };
    const label = new Label(values);
    const result = label.route("test");
    assert.equal(result.name, "test");
    assert.equal(result.query.q, "label:black-cat");
  });

  it("should return batch size", () => {
    assert.equal(Label.batchSize(), 24);
    Label.setBatchSize(30);
    assert.equal(Label.batchSize(), 30);
    Label.setBatchSize(24);
  });

  it("should return classes", () => {
    const values = { ID: 5, UID: "ABC123", Name: "Black Cat", Slug: "black-cat", Favorite: true };
    const label = new Label(values);
    const result = label.classes(true);
    assert.include(result, "is-label");
    assert.include(result, "uid-ABC123");
    assert.include(result, "is-selected");
    assert.include(result, "is-favorite");
  });

  it("should get label entity name", () => {
    const values = { ID: 5, UID: "ABC123", Name: "Black Cat", Slug: "black-cat" };
    const label = new Label(values);
    const result = label.getEntityName();
    assert.equal(result, "black-cat");
  });

  it("should get label id", () => {
    const values = { ID: 5, UID: "ABC123", Name: "Black Cat", Slug: "black-cat" };
    const label = new Label(values);
    const result = label.getId();
    assert.equal(result, "ABC123");
  });

  it("should get label title", () => {
    const values = { ID: 5, UID: "ABC123", Name: "Black Cat", Slug: "black-cat" };
    const label = new Label(values);
    const result = label.getTitle();
    assert.equal(result, "Black Cat");
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
    assert.equal(result, "/api/v1/t/c6b24d688564f7ddc7b245a414f003a8d8ff5a67/public/xyz");

    const values2 = {
      ID: 5,
      UID: "ABC123",
      Name: "Black Cat",
      Slug: "black-cat",
    };
    const label2 = new Label(values2);
    const result2 = label2.thumbnailUrl("xyz");
    assert.equal(result2, "/api/v1/labels/ABC123/t/public/xyz");

    const values3 = {
      ID: 5,
      Name: "Black Cat",
      Slug: "black-cat",
    };
    const label3 = new Label(values3);
    const result3 = label3.thumbnailUrl("xyz");
    assert.equal(result3, "/api/v1/svg/label");
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
    assert.equal(result.replaceAll("\u202f", " "), "Jul 8, 2012, 2:45 PM");
  });

  it("should get model name", () => {
    const result = Label.getModelName();
    assert.equal(result, "Label");
  });

  it("should get collection resource", () => {
    const result = Label.getCollectionResource();
    assert.equal(result, "labels");
  });

  it("should like label", () => {
    const values = { ID: 5, UID: "ABC123", Name: "Black Cat", Slug: "black-cat", Favorite: false };
    const label = new Label(values);
    assert.equal(label.Favorite, false);
    label.like();
    assert.equal(label.Favorite, true);
  });

  it("should unlike label", () => {
    const values = { ID: 5, UID: "ABC123", Name: "Black Cat", Slug: "black-cat", Favorite: true };
    const label = new Label(values);
    assert.equal(label.Favorite, true);
    label.unlike();
    assert.equal(label.Favorite, false);
  });

  it("should toggle like", () => {
    const values = { ID: 5, UID: "ABC123", Name: "Black Cat", Slug: "black-cat", Favorite: true };
    const label = new Label(values);
    assert.equal(label.Favorite, true);
    label.toggleLike();
    assert.equal(label.Favorite, false);
    label.toggleLike();
    assert.equal(label.Favorite, true);
  });

  it("should get label defaults", () => {
    const values = { ID: 5, UID: "ABC123" };
    const label = new Label(values);
    const result = label.getDefaults();
    assert.equal(result.ID, 0);
  });
});
