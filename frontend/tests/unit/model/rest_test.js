import "../fixtures";
import Rest from "model/rest";
import Album from "model/album";
import Label from "model/label";
import Link from "model/link";

let chai = require("chai/chai");
let assert = chai.assert;

describe("model/abstract", () => {
  it("should set values", () => {
    const values = { ID: 5, Name: "Black Cat", Slug: "black-cat" };
    const label = new Label(values);
    assert.equal(label.Name, "Black Cat");
    assert.equal(label.Slug, "black-cat");
    label.setValues();
    assert.equal(label.Name, "Black Cat");
    assert.equal(label.Slug, "black-cat");
    const values2 = { ID: 6, Name: "White Cat", Slug: "white-cat" };
    label.setValues(values2);
    assert.equal(label.Name, "White Cat");
    assert.equal(label.Slug, "white-cat");
  });

  it("should get values", () => {
    const values = { ID: 5, Name: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values);
    const result = album.getValues();
    assert.equal(result.Name, "Christmas 2019");
    assert.equal(result.UID, 66);
  });

  it("should get id", () => {
    const values = { ID: 5, Name: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values);
    const result = album.getId();
    assert.equal(result, 66);
  });

  it("should test if id exists", () => {
    const values = { ID: 5, Name: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values);
    const result = album.hasId();
    assert.equal(result, true);
  });

  it("should get model name", () => {
    const result = Rest.getModelName();
    assert.equal(result, "Item");
  });

  it("should update album", async () => {
    const values = { ID: 5, Name: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values);
    assert.equal(album.Description, undefined);
    album.Name = "Christmas 2020";
    await album.update();
    assert.equal(album.Description, "Test description");
  });

  it("should save album", async () => {
    const values = { UID: "abc", Name: "Christmas 2019", Slug: "christmas-2019" };
    const album = new Album(values);
    album.Name = "Christmas 2020";
    assert.equal(album.Description, undefined);
    await album.save();
    assert.equal(album.Description, "Test description");

    const values2 = { Name: "Christmas 2019", Slug: "christmas-2019" };
    const album2 = new Album(values2);
    album.Name = "Christmas 2020";
    assert.equal(album2.Description, undefined);
    await album2.save().then((response) => {
      assert.equal(response.success, "ok");
    });
    assert.equal(album2.Description, undefined);
  });

  it("should remove album", async () => {
    const values = { UID: "abc", Name: "Christmas 2019", Slug: "christmas-2019" };
    const album = new Album(values);
    assert.equal(album.Name, "Christmas 2019");
    await album.remove();
  });

  it("should get edit form", async () => {
    const values = { UID: "abc", Name: "Christmas 2019", Slug: "christmas-2019" };
    const album = new Album(values);
    const result = await album.getEditForm();
    assert.equal(result.definition.foo, "edit");
  });

  it("should get create form", async () => {
    const result = await Album.getCreateForm();
    assert.equal(result.definition.foo, "bar");
  });

  it("should get search form", async () => {
    const result = await Album.getSearchForm();
    assert.equal(result.definition.foo, "bar");
  });

  it("should search label", async () => {
    const result = await Album.search();
    assert.equal(result.data.ID, 51);
    assert.equal(result.data.Name, "tabby cat");
  });

  it("should get collection resource", () => {
    assert.equal(Rest.getCollectionResource(), "");
  });

  it("should get slug", () => {
    const values = { ID: 5, Name: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values);
    const result = album.getSlug();
    assert.equal(result, "christmas-2019");
  });

  it("should get slug", () => {
    const values = { ID: 5, Name: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values);
    const result = album.clone();
    assert.equal(result.Slug, "christmas-2019");
    assert.equal(result.Name, "Christmas 2019");
    assert.equal(result.ID, 5);
  });

  it("should find album", async () => {
    const values = { Name: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values);
    return album
      .find(5)
      .then((response) => {
        assert.equal(response.UID, "5");
        return Promise.resolve();
      })
      .catch((error) => {
        return Promise.reject(error);
      });
  });

  it("should get entity name", () => {
    const values = { ID: 5, Name: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values);
    const result = album.getEntityName();
    assert.equal(result, "christmas-2019");
  });

  it("should return model name", () => {
    const values = { ID: 5, Name: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values);
    const result = album.modelName();
    assert.equal(result, "Album");
  });

  it("should return limit", () => {
    const values = { ID: 5, Name: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values);
    const result = Rest.limit();
    assert.equal(result, 10000);
    assert.equal(album.constructor.limit(), 10000);
  });

  it("should create link", async () => {
    const values = { ID: 5, Name: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values);
    album
      .createLink("passwd", 8000)
      .then((response) => {
        assert.equal(response.Slug, "christmas-2019");
        return Promise.resolve();
      })
      .catch((error) => {
        return Promise.reject(error);
      });
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
    return album
      .updateLink(link)
      .then((response) => {
        assert.equal(response.Slug, "friends");
        return Promise.resolve();
      })
      .catch((error) => {
        return Promise.reject(error);
      });
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
    return album
      .removeLink(link)
      .then((response) => {
        assert.equal(response.Success, "ok");
        return Promise.resolve();
      })
      .catch((error) => {
        return Promise.reject(error);
      });
  });

  it("should return links", async () => {
    const values2 = { ID: 5, Name: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values2);
    return album
      .links()
      .then((response) => {
        assert.equal(response.count, 2);
        assert.equal(response.models.length, 2);
        return Promise.resolve();
      })
      .catch((error) => {
        return Promise.reject(error);
      });
  });
});
