import { describe, it, expect } from "vitest";
import "../fixtures";
import Link from "model/link";

describe("model/link", () => {
  it("should get link defaults", () => {
    const values = { UID: 5 };
    const link = new Link(values);
    const result = link.getDefaults();
    expect(result.UID).toBe("");
    expect(result.Perm).toBe(0);
    expect(result.Comment).toBe("");
    expect(result.ShareUID).toBe("");
  });

  it("should get link url", () => {
    const values = { UID: 5, Token: "1234hhtbbt", Slug: "friends", ShareUID: "family" };
    const link = new Link(values);
    const result = link.url();
    expect(result).toBe("http://localhost:2342/s/1234hhtbbt/friends");
    const values2 = { UID: 5, Token: "", ShareUID: "family" };
    const link2 = new Link(values2);
    const result2 = link2.url();
    expect(result2).toBe("http://localhost:2342/s/…/family");
  });

  it("should get link caption", () => {
    const values = { UID: 5, Token: "AcfgbTTh", Slug: "friends", ShareUID: "family" };
    const link = new Link(values);
    const result = link.caption();
    expect(result).toBe("/s/acfgbtth");
  });

  it("should get link id", () => {
    const values = { UID: 5 };
    const link = new Link(values);
    const result = link.getId();
    expect(result).toBe(5);
    const values2 = {};
    const link2 = new Link(values2);
    const result2 = link2.getId();
    expect(result2).toBe(false);
  });

  it("should test has id", () => {
    const values = { UID: 5 };
    const link = new Link(values);
    const result = link.hasId();
    expect(result).toBe(true);
  });

  it("should get link slug", () => {
    const values = { UID: 5, Token: "AcfgbTTh", Slug: "friends", ShareUID: "family" };
    const link = new Link(values);
    const result = link.getSlug();
    expect(result).toBe("friends");
  });

  it("should test has slug", () => {
    const values = { UID: 5, Token: "AcfgbTTh", Slug: "friends", ShareUID: "family" };
    const link = new Link(values);
    const result = link.hasSlug();
    expect(result).toBe(true);
    const values2 = { UID: 5, Token: "AcfgbTTh", ShareUID: "family" };
    const link2 = new Link(values2);
    const result2 = link2.hasSlug();
    expect(result2).toBe(false);
  });

  it("should clone link", () => {
    const values = { UID: 5, Token: "AcfgbTTh", Slug: "friends", ShareUID: "family" };
    const link = new Link(values);
    const result = link.clone();
    expect(result.Slug).toBe("friends");
    expect(result.Token).toBe("AcfgbTTh");
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
    expect(result).toBe("Jul 9, 2012");
  });

  it("should get collection resource", () => {
    const result = Link.getCollectionResource();
    expect(result).toBe("links");
  });

  it("should get model name", () => {
    const result = Link.getModelName();
    expect(result).toBe("Link");
  });
});
