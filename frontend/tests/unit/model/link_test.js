import "../fixtures";
import Link from "model/link";

let chai = require("chai/chai");
let assert = chai.assert;

describe("model/link", () => {
  it("should get link defaults", () => {
    const values = { UID: 5 };
    const link = new Link(values);
    const result = link.getDefaults();
    assert.equal(result.UID, 0);
    assert.equal(result.Perm, 0);
    assert.equal(result.Comment, "");
    assert.equal(result.ShareUID, "");
  });

  it("should get link url", () => {
    const values = { UID: 5, Token: "1234hhtbbt", Slug: "friends", ShareUID: "family" };
    const link = new Link(values);
    const result = link.url();
    assert.equal(result, "http://localhost:2342/s/1234hhtbbt/friends");
    const values2 = { UID: 5, Token: "", ShareUID: "family" };
    const link2 = new Link(values2);
    const result2 = link2.url();
    assert.equal(result2, "http://localhost:2342/s/…/family");
  });

  it("should get link caption", () => {
    const values = { UID: 5, Token: "AcfgbTTh", Slug: "friends", ShareUID: "family" };
    const link = new Link(values);
    const result = link.caption();
    assert.equal(result, "/s/acfgbtth");
  });

  it("should get link id", () => {
    const values = { UID: 5 };
    const link = new Link(values);
    const result = link.getId();
    assert.equal(result, 5);
    const values2 = {};
    const link2 = new Link(values2);
    const result2 = link2.getId();
    assert.equal(result2, false);
  });

  it("should test has id", () => {
    const values = { UID: 5 };
    const link = new Link(values);
    const result = link.hasId();
    assert.equal(result, true);
  });

  it("should get link slug", () => {
    const values = { UID: 5, Token: "AcfgbTTh", Slug: "friends", ShareUID: "family" };
    const link = new Link(values);
    const result = link.getSlug();
    assert.equal(result, "friends");
  });

  it("should test has slug", () => {
    const values = { UID: 5, Token: "AcfgbTTh", Slug: "friends", ShareUID: "family" };
    const link = new Link(values);
    const result = link.hasSlug();
    assert.equal(result, true);
    const values2 = { UID: 5, Token: "AcfgbTTh", ShareUID: "family" };
    const link2 = new Link(values2);
    const result2 = link2.hasSlug();
    assert.equal(result2, false);
  });

  it("should clone link", () => {
    const values = { UID: 5, Token: "AcfgbTTh", Slug: "friends", ShareUID: "family" };
    const link = new Link(values);
    const result = link.clone();
    assert.equal(result.Slug, "friends");
    assert.equal(result.Token, "AcfgbTTh");
  });

  it("should test expire", () => {
    const values = {
      UID: 5,
      Token: "AcfgbTTh",
      Slug: "friends",
      ShareUID: "family",
      Expires: 80000,
      ModifiedAt: "2012-07-08T14:45:39Z",
    };
    const link = new Link(values);
    const result = link.expires();
    assert.equal(result, "Jul 9, 2012");
  });

  it("should get collection resource", () => {
    const result = Link.getCollectionResource();
    assert.equal(result, "links");
  });

  it("should get model name", () => {
    const result = Link.getModelName();
    assert.equal(result, "Link");
  });
});
