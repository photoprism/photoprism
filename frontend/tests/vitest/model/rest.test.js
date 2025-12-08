import { describe, it, expect } from "vitest";
import "../fixtures";
import Rest from "model/rest";
import Album from "model/album";
import Label from "model/label";
import Link from "model/link";

describe("model/abstract", () => {
  it("should set values", () => {
    const values = { ID: 5, Name: "Black Cat", Slug: "black-cat" };
    const label = new Label(values);
    expect(label.Name).toBe("Black Cat");
    expect(label.Slug).toBe("black-cat");
    label.setValues();
    expect(label.Name).toBe("Black Cat");
    expect(label.Slug).toBe("black-cat");
    const values2 = { ID: 6, Name: "White Cat", Slug: "white-cat" };
    label.setValues(values2);
    expect(label.Name).toBe("White Cat");
    expect(label.Slug).toBe("white-cat");
  });

  it("should get values", () => {
    const values = { ID: 5, Name: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values);
    const result = album.getValues();
    expect(result.Name).toBe("Christmas 2019");
    expect(result.UID).toBe(66);
  });

  it("should get id", () => {
    const values = { ID: 5, Name: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values);
    const result = album.getId();
    expect(result).toBe(66);
  });

  it("should test if id exists", () => {
    const values = { ID: 5, Name: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values);
    const result = album.hasId();
    expect(result).toBe(true);
  });

  it("should get model name", () => {
    const result = Rest.getModelName();
    expect(result).toBe("Item");
  });

  it("should update album", async () => {
    const values = { ID: 5, Name: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values);
    expect(album.Description).toBeUndefined();
    album.Name = "Christmas 2020";
    await album.update();
    expect(album.Description).toBe("Test description");
  });

  it("should save album", async () => {
    const values = { UID: "abc", Name: "Christmas 2019", Slug: "christmas-2019" };
    const album = new Album(values);
    album.Name = "Christmas 2020";
    expect(album.Description).toBeUndefined();
    await album.save();
    expect(album.Description).toBe("Test description");

    const values2 = { Name: "Christmas 2019", Slug: "christmas-2019" };
    const album2 = new Album(values2);
    album2.Name = "Christmas 2020";
    expect(album2.Description).toBeUndefined();
    const response = await album2.save();
    expect(response.success).toBe("ok");
    expect(album2.Description).toBeUndefined();
  });

  it("should remove album", async () => {
    const values = { UID: "abc", Name: "Christmas 2019", Slug: "christmas-2019" };
    const album = new Album(values);
    expect(album.Name).toBe("Christmas 2019");
    await album.remove();
  });

  it("should get edit form", async () => {
    const values = { UID: "abc", Name: "Christmas 2019", Slug: "christmas-2019" };
    const album = new Album(values);
    const result = await album.getEditForm();
    expect(result.definition.foo).toBe("edit");
  });

  it("should get create form", async () => {
    const result = await Album.getCreateForm();
    expect(result.definition.foo).toBe("bar");
  });

  it("should get search form", async () => {
    const result = await Album.getSearchForm();
    expect(result.definition.foo).toBe("bar");
  });

  it("should search label", async () => {
    const result = await Album.search();
    expect(result.data.ID).toBe(51);
    expect(result.data.Name).toBe("tabby cat");
  });

  it("should get collection resource", () => {
    expect(Rest.getCollectionResource()).toBe("");
  });

  it("should get slug", () => {
    const values = { ID: 5, Name: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values);
    const result = album.getSlug();
    expect(result).toBe("christmas-2019");
  });

  it("should get slug", () => {
    const values = { ID: 5, Name: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values);
    const result = album.clone();
    expect(result.Slug).toBe("christmas-2019");
    expect(result.Name).toBe("Christmas 2019");
    expect(result.ID).toBe(5);
  });

  it("should find album", async () => {
    const values = { Name: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values);
    const response = await album.find(5);
    expect(response.UID).toBe("5");
  });

  it("should get entity name", () => {
    const values = { ID: 5, Name: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values);
    const result = album.getEntityName();
    expect(result).toBe("christmas-2019");
  });

  it("should return model name", () => {
    const values = { ID: 5, Name: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values);
    const result = album.modelName();
    expect(result).toBe("Album");
  });

  it("should return limit", () => {
    const values = { ID: 5, Name: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values);
    const result = Rest.limit();
    expect(result).toBe(100000);
    expect(album.constructor.limit()).toBe(100000);
  });

  it("should create link", async () => {
    const values = { ID: 5, Name: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values);
    const response = await album.createLink("passwd", 8000);
    expect(response.Slug).toBe("christmas-2019");
  });

  it("should update link", async () => {
    const values = {
      UID: 5,
      Password: "passwd",
      Slug: "friends",
      Expires: 80000,
      UpdatedAt: "2012-07-08T14:45:39Z",
      Token: "abchhgftryue2345",
    };
    const link = new Link(values);
    const values2 = { ID: 5, Name: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values2);
    const response = await album.updateLink(link);
    expect(response.Slug).toBe("friends");
  });

  it("should remove link", async () => {
    const values = {
      UID: 5,
      Password: "passwd",
      Slug: "friends",
      Expires: 80000,
      UpdatedAt: "2012-07-08T14:45:39Z",
    };
    const link = new Link(values);
    const values2 = { ID: 5, Name: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values2);
    const response = await album.removeLink(link);
    expect(response.Success).toBe("ok");
  });

  it("should return links", async () => {
    const values2 = { ID: 5, Name: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values2);
    const response = await album.links();
    expect(response.count).toBe(2);
    expect(response.models.length).toBe(2);
  });
});
