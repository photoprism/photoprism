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

  describe("setValues", () => {
    it("returns the instance so calls can be chained", () => {
      const album = new Album({ Name: "First" });
      expect(album.setValues({ Name: "Second" })).toBe(album);
    });

    it("ignores a falsy values argument", () => {
      const album = new Album({ Name: "Kept" });
      album.setValues(null);
      album.setValues(undefined);
      album.setValues(false);
      expect(album.Name).toBe("Kept");
      expect(album.wasChanged()).toBe(false);
    });

    it("skips the reserved __originalValues key", () => {
      const album = new Album({ Name: "Original" });
      album.setValues({ __originalValues: { Name: "Tampered" } });
      expect(album.__originalValues.Name).toBe("Original");
    });

    it("snapshots scalars and deep-clones objects so mutations don't bleed into __originalValues", () => {
      const link = new Link({ ID: 1, UID: "abc", Token: "tkn", ShareUID: "shr" });
      // Scalar mutation never reaches the snapshot.
      link.Token = "changed";
      expect(link.__originalValues.Token).toBe("tkn");
    });

    it("skips object snapshots when scalarOnly is true", () => {
      const album = new Album({ Title: "Trip" });
      const nested = { foo: "bar" };
      album.setValues({ Title: "Trip", Notes: "x", Extras: nested }, true);
      expect(album.Extras).toBe(nested);
      // scalarOnly: object values are not tracked in __originalValues, so
      // wasChanged() can't detect mutations on them — by design.
      expect(album.__originalValues.Extras).toBeUndefined();
      expect(album.__originalValues.Title).toBe("Trip");
    });
  });

  describe("getValues", () => {
    it("with changed=true returns only fields that differ from the snapshot", () => {
      const album = new Album({ ID: 1, Title: "Trip", Slug: "trip", PhotoCount: 0 });
      album.Title = "Vacation";
      album.PhotoCount = 5;
      const diff = album.getValues(true);
      expect(diff).toEqual({ Title: "Vacation", PhotoCount: 5 });
    });

    it("coerces strings, numbers, and booleans according to getDefaults()", () => {
      const album = new Album({ Title: "Trip", PhotoCount: 0, Favorite: false });
      // String default => null becomes "".
      album.Title = null;
      // Number default => parseFloat applied to whatever is on the instance.
      album.PhotoCount = "42";
      // Boolean default => coerced via Boolean().
      album.Favorite = 1;
      const values = album.getValues();
      expect(values.Title).toBe("");
      expect(values.PhotoCount).toBe(42);
      expect(values.Favorite).toBe(true);
    });

    it("passes through fields that have no default verbatim", () => {
      // Album.getDefaults() has no `ID`, so ID is tracked in __originalValues
      // but bypasses type coercion — getValues() returns it as-is.
      const album = new Album({ ID: 7, Title: "Trip" });
      const values = album.getValues();
      expect(values.ID).toBe(7);
    });
  });

  describe("originalValue", () => {
    it("returns the snapshot value for a tracked key", () => {
      const album = new Album({ Title: "Original", PhotoCount: 3 });
      album.Title = "Mutated";
      album.PhotoCount = 99;
      expect(album.originalValue("Title")).toBe("Original");
      expect(album.originalValue("PhotoCount")).toBe(3);
    });

    it("falls back to the live value for an untracked own property", () => {
      const album = new Album({ Title: "Trip" });
      album.RuntimeOnly = "ad-hoc";
      // RuntimeOnly was assigned after construction, so it isn't in
      // __originalValues; the fallback returns whatever's currently on `this`.
      expect(album.originalValue("RuntimeOnly")).toBe("ad-hoc");
    });

    it("returns null for unknown keys", () => {
      const album = new Album({ Title: "Trip" });
      expect(album.originalValue("DoesNotExist")).toBeNull();
    });

    it("returns null when the reserved __originalValues key is requested", () => {
      const album = new Album({ Title: "Trip" });
      expect(album.originalValue("__originalValues")).toBeNull();
    });
  });

  describe("wasChanged", () => {
    it("returns false on a freshly loaded model", () => {
      const album = new Album({ ID: 1, Title: "Trip" });
      expect(album.wasChanged()).toBe(false);
    });

    it("returns true after any tracked field is mutated", () => {
      const album = new Album({ ID: 1, Title: "Trip" });
      album.Title = "Vacation";
      expect(album.wasChanged()).toBe(true);
    });

    it("flips back to false after rollback() restores the snapshot", () => {
      const album = new Album({ ID: 1, Title: "Trip" });
      album.Title = "Vacation";
      album.rollback();
      expect(album.wasChanged()).toBe(false);
    });
  });

  describe("rollback", () => {
    it("restores scalar fields to the __originalValues snapshot", () => {
      const album = new Album({ ID: 5, Name: "Christmas 2019", Slug: "christmas-2019", UID: 66 });
      album.Name = "Mutated";
      album.Slug = "mutated";
      expect(album.wasChanged()).toBe(true);
      album.rollback();
      expect(album.Name).toBe("Christmas 2019");
      expect(album.Slug).toBe("christmas-2019");
      expect(album.wasChanged()).toBe(false);
    });

    it("deep-clones object fields so post-rollback mutations don't bleed into the snapshot", () => {
      const link = new Link({ ID: 1, UID: "abc", Token: "tkn", ShareUID: "shr" });
      link.Token = "changed";
      link.rollback();
      expect(link.Token).toBe("tkn");
      // A second mutation must still be rollable — the snapshot wasn't aliased.
      link.Token = "changed-again";
      link.rollback();
      expect(link.Token).toBe("tkn");
    });

    it("returns the model so calls can be chained", () => {
      const album = new Album({ Name: "X" });
      album.Name = "Y";
      expect(album.rollback()).toBe(album);
    });

    it("is a no-op on a clean model", () => {
      const album = new Album({ ID: 1, Name: "Pristine" });
      expect(album.wasChanged()).toBe(false);
      album.rollback();
      expect(album.Name).toBe("Pristine");
      expect(album.wasChanged()).toBe(false);
    });
  });
});
