import "../fixtures";
import Marker from "model/marker";

let chai = require("chai/chai");
let assert = chai.assert;

describe("model/marker", () => {
  it("should get marker defaults", () => {
    const values = { FileUID: "fghjojp" };
    const marker = new Marker(values);
    const result = marker.getDefaults();
    assert.equal(result.UID, "");
    assert.equal(result.FileUID, "");
  });

  it("should get route view", () => {
    const values = { UID: "ABC123ghytr", FileUID: "fhjouohnnmnd", Type: "face", Src: "image" };
    const marker = new Marker(values);
    const result = marker.route("test");
    assert.equal(result.name, "test");
    assert.equal(result.query.q, "marker:ABC123ghytr");
  });

  it("should return classes", () => {
    const values = { UID: "ABC123ghytr", FileUID: "fhjouohnnmnd", Type: "face", Src: "image" };
    const marker = new Marker(values);
    const result = marker.classes(true);
    assert.include(result, "is-marker");
    assert.include(result, "uid-ABC123ghytr");
    assert.include(result, "is-selected");
    assert.notInclude(result, "is-review");
    assert.notInclude(result, "is-invalid");
    const result2 = marker.classes(false);
    assert.include(result2, "is-marker");
    assert.include(result2, "uid-ABC123ghytr");
    assert.notInclude(result2, "is-selected");
    assert.notInclude(result2, "is-review");
    assert.notInclude(result2, "is-invalid");
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
    assert.include(result3, "is-marker");
    assert.include(result3, "uid-mBC123ghytr");
    assert.include(result3, "is-selected");
    assert.include(result3, "is-review");
    assert.include(result3, "is-invalid");
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
    assert.equal(result, "test");
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
    assert.equal(result, "test");
  });

  it("should get thumbnail url", () => {
    const values = { UID: "ABC123ghytr", FileUID: "fhjouohnnmnd", Type: "face", Src: "image" };
    const marker = new Marker(values);
    const result = marker.thumbnailUrl("xyz");
    assert.equal(result, "/api/v1/svg/portrait");

    const values2 = {
      UID: "ABC123ghytr",
      FileUID: "fhjouohnnmnd",
      Type: "face",
      Src: "image",
      Thumb: "nicethumbuid",
    };
    const marker2 = new Marker(values2);
    const result2 = marker2.thumbnailUrl();
    assert.equal(result2, "/api/v1/t/nicethumbuid/public/tile_160");
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
    assert.equal(result.replaceAll("\u202f", " "), "Jul 8, 2012, 2:45 PM");
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
    assert.equal(marker.Review, true);
    assert.equal(marker.Invalid, true);
    marker.approve();
    assert.equal(marker.Review, false);
    assert.equal(marker.Invalid, false);
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
    assert.equal(marker.Review, true);
    assert.equal(marker.Invalid, false);
    marker.reject();
    assert.equal(marker.Review, false);
    assert.equal(marker.Invalid, true);
  });

  it("should rename marker", (done) => {
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
    assert.equal(marker.Name, "");
    marker.rename();
    assert.equal(marker.Name, "");
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
    assert.equal(marker2.Name, "testname");
    marker2
      .rename()
      .then((response) => {
        assert.equal(response.success, "ok");
        done();
      })
      .catch((error) => {
        done(error);
      });
  });

  it("should clear subject", (done) => {
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
    marker
      .clearSubject()
      .then((response) => {
        assert.equal(response.success, "ok");
        done();
      })
      .catch((error) => {
        done(error);
      });
  });

  it("should return batch size", () => {
    assert.equal(Marker.batchSize(), 48);
    Marker.setBatchSize(30);
    assert.equal(Marker.batchSize(), 30);
    Marker.setBatchSize(48);
  });

  it("should get collection resource", () => {
    const result = Marker.getCollectionResource();
    assert.equal(result, "markers");
  });

  it("should get model name", () => {
    const result = Marker.getModelName();
    assert.equal(result, "Marker");
  });
});
