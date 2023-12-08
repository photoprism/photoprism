import "../fixtures";
import Subject from "model/subject";

let chai = require("chai/chai");
let assert = chai.assert;

describe("model/subject", () => {
  it("should get face defaults", () => {
    const values = {};
    const subject = new Subject(values);
    const result = subject.getDefaults();
    assert.equal(result.UID, "");
    assert.equal(result.Favorite, false);
  });

  it("should get route view", () => {
    const values = { UID: "s123ghytrfggd", Type: "person", Src: "manual" };
    const subject = new Subject(values);
    const result = subject.route("test");
    assert.equal(result.name, "test");
    assert.equal(result.query.q, "subject:s123ghytrfggd");
    const values2 = {
      UID: "s123ghytrfggd",
      Type: "person",
      Src: "manual",
      Name: "Jane Doe",
      Slug: "jane-doe",
    };
    const subject2 = new Subject(values2);
    const result2 = subject2.route("test");
    assert.equal(result2.name, "test");
    assert.equal(result2.query.q, "person:jane-doe");
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
    assert.include(result, "is-subject");
    assert.include(result, "uid-s123ghytrfggd");
    assert.include(result, "is-selected");
    assert.notInclude(result, "is-favorite");
    assert.include(result, "is-private");
    assert.include(result, "is-excluded");
    assert.include(result, "is-hidden");
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
    assert.include(result2, "is-subject");
    assert.include(result2, "uid-s123ghytrfggd");
    assert.notInclude(result2, "is-selected");
    assert.include(result2, "is-favorite");
    assert.notInclude(result2, "is-private");
    assert.notInclude(result2, "is-excluded");
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
    assert.equal(result, "jane-doe");
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
    assert.equal(result, "Jane Doe");
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
    assert.equal(result, "/api/v1/t/nicethumb/public/xyz");
    const result2 = subject.thumbnailUrl();
    assert.equal(result2, "/api/v1/t/nicethumb/public/tile_160");
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
    assert.equal(result3, "/api/v1/svg/portrait");
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
    assert.equal(result.replaceAll("\u202f", " "), "Jul 8, 2012, 2:45 PM");
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
    assert.equal(subject.Favorite, false);
    subject.like();
    assert.equal(subject.Favorite, true);
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
    assert.equal(subject.Favorite, true);
    subject.unlike();
    assert.equal(subject.Favorite, false);
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
    assert.equal(subject.Favorite, true);
    subject.toggleLike();
    assert.equal(subject.Favorite, false);
    subject.toggleLike();
    assert.equal(subject.Favorite, true);
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
    assert.equal(subject.Hidden, true);
    subject.show();
    assert.equal(subject.Hidden, false);
    subject.hide();
    assert.equal(subject.Hidden, true);
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
    assert.equal(subject.Hidden, true);
    subject.toggleHidden();
    assert.equal(subject.Hidden, false);
    subject.toggleHidden();
    assert.equal(subject.Hidden, true);
  });

  it("should return batch size", () => {
    assert.equal(Subject.batchSize(), 60);
    Subject.setBatchSize(30);
    assert.equal(Subject.batchSize(), 30);
    Subject.setBatchSize(60);
  });

  it("should get collection resource", () => {
    const result = Subject.getCollectionResource();
    assert.equal(result, "subjects");
  });

  it("should get model name", () => {
    const result = Subject.getModelName();
    assert.equal(result, "Subject");
  });
});
