import "../fixtures";

import Service from "model/service";
import Photo from "model/photo";

let chai = require("chai/chai");
let assert = chai.assert;

describe("model/service", () => {
  it("should get service defaults", () => {
    const values = { ID: 5 };
    const service = new Service(values);
    const result = service.getDefaults();
    assert.equal(result.ID, 0);
    assert.equal(result.AccShare, true);
    assert.equal(result.AccName, "");
  });

  it("should get service entity name", () => {
    const values = { ID: 5, AccName: "Test Name" };
    const service = new Service(values);
    const result = service.getEntityName();
    assert.equal(result, "Test Name");
  });

  it("should get service id", () => {
    const values = { ID: 5, AccName: "Test Name" };
    const service = new Service(values);
    const result = service.getId();
    assert.equal(result, 5);
  });

  it("should get folders", (done) => {
    const values = { ID: 123, AccName: "Test Name" };
    const service = new Service(values);
    service
      .Folders()
      .then((response) => {
        assert.equal(response.foo, "folders");
        done();
      })
      .catch((error) => {
        done(error);
      });
  });

  it("should get share photos", (done) => {
    const values = { ID: 123, AccName: "Test Name" };
    const service = new Service(values);
    const values1 = { ID: 5, Title: "Crazy Cat", UID: 789 };
    const photo = new Photo(values1);
    const values2 = { ID: 6, Title: "Crazy Cat 2", UID: 783 };
    const photo2 = new Photo(values2);
    const Photos = [photo, photo2];
    service
      .Upload(Photos, "destination")
      .then((response) => {
        assert.equal(response.foo, "upload");
        done();
      })
      .catch((error) => {
        done(error);
      });
  });

  it("should get collection resource", () => {
    const result = Service.getCollectionResource();
    assert.equal(result, "services");
  });

  it("should get model name", () => {
    const result = Service.getModelName();
    assert.equal(result, "Account");
  });
});
