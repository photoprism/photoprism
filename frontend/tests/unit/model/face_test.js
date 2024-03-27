import "../fixtures";
import Face from "model/face";

let chai = require("chai/chai");
let assert = chai.assert;

describe("model/face", () => {
  it("should get face defaults", () => {
    const values = {};
    const face = new Face(values);
    const result = face.getDefaults();
    assert.equal(result.ID, "");
    assert.equal(result.SampleRadius, 0.0);
  });

  it("should get route view", () => {
    const values = { ID: "f123ghytrfggd", Samples: 5 };
    const face = new Face(values);
    const result = face.route("test");
    assert.equal(result.name, "test");
    assert.equal(result.query.q, "face:f123ghytrfggd");
  });

  it("should return classes", () => {
    const values = { ID: "f123ghytrfggd", Samples: 5 };
    const face = new Face(values);
    const result = face.classes(true);
    assert.include(result, "is-face");
    assert.include(result, "uid-f123ghytrfggd");
    assert.include(result, "is-selected");
    assert.notInclude(result, "is-hidden");
    const result2 = face.classes(false);
    assert.include(result2, "is-face");
    assert.include(result2, "uid-f123ghytrfggd");
    assert.notInclude(result2, "is-selected");
    assert.notInclude(result2, "is-hidden");
    const values2 = { ID: "f123ghytrfggd", Samples: 5, Hidden: true };
    const face2 = new Face(values2);
    const result3 = face2.classes(true);
    assert.include(result3, "is-face");
    assert.include(result3, "uid-f123ghytrfggd");
    assert.include(result3, "is-selected");
    assert.include(result3, "is-hidden");
  });

  it("should get face entity name", () => {
    const values = { ID: "f123ghytrfggd", Samples: 5 };
    const face = new Face(values);
    const result = face.getEntityName();
    assert.equal(result, "f123ghytrfggd");
  });

  it("should get face title", () => {
    const values = { ID: "f123ghytrfggd", Samples: 5 };
    const face = new Face(values);
    const result = face.getTitle();
    assert.equal(result, undefined);
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

    assert.equal(result, "/api/v1/t/7ca759a2b788cc5bcc08dbbce9854ff94a2f94d1/public/xyz");

    const values2 = {
      ID: "f123ghytrfggd",
      Samples: 5,
      Thumb: "7ca759a2b788cc5bcc08dbbce9854ff94a2f94d1",
    };
    const face2 = new Face(values2);
    const result2 = face2.thumbnailUrl();

    assert.equal(result2, "/api/v1/t/7ca759a2b788cc5bcc08dbbce9854ff94a2f94d1/public/tile_160");

    const values3 = {
      ID: "f123ghytrfggd",
      Samples: 5,
      Thumb: "",
    };
    const face3 = new Face(values3);
    const result3 = face3.thumbnailUrl("tile_240");

    assert.equal(result3, "/api/v1/svg/portrait");
  });

  it("should get date string", () => {
    const values = {
      ID: "f123ghytrfggd",
      Samples: 5,
      CreatedAt: "2012-07-08T14:45:39Z",
    };
    const face = new Face(values);
    const result = face.getDateString();
    assert.equal(result.replaceAll("\u202f", " "), "Jul 8, 2012, 2:45 PM");
  });

  it("show and hide face", () => {
    const values = {
      ID: "f123ghytrfggd",
      Samples: 5,
      CreatedAt: "2012-07-08T14:45:39Z",
      Hidden: true,
    };
    const face = new Face(values);
    assert.equal(face.Hidden, true);
    face.show();
    assert.equal(face.Hidden, false);
    face.hide();
    assert.equal(face.Hidden, true);
  });

  it("should toggle hidden", () => {
    const values = {
      ID: "f123ghytrfggd",
      Samples: 5,
      CreatedAt: "2012-07-08T14:45:39Z",
      Hidden: true,
    };
    const face = new Face(values);
    assert.equal(face.Hidden, true);
    face.toggleHidden();
    assert.equal(face.Hidden, false);
    face.toggleHidden();
    assert.equal(face.Hidden, true);
  });

  it("should set name", (done) => {
    const values = { ID: "f123ghytrfggd", Samples: 5, MarkerUID: "mDC123ghytr", Name: "Jane" };
    const face = new Face(values);
    face
      .setName("testname")
      .then((response) => {
        assert.equal(response.Name, "testname");
        done();
      })
      .catch((error) => {
        done(error);
      });

    face
      .setName("")
      .then((response) => {
        assert.equal(response.Name, "Jane");
        done();
      })
      .catch((error) => {
        done(error);
      });
  });

  it("should return batch size", () => {
    assert.equal(Face.batchSize(), 24);
    Face.setBatchSize(30);
    assert.equal(Face.batchSize(), 30);
    Face.setBatchSize(24);
  });

  it("should get collection resource", () => {
    const result = Face.getCollectionResource();
    assert.equal(result, "faces");
  });

  it("should get model name", () => {
    const result = Face.getModelName();
    assert.equal(result, "Face");
  });
});
