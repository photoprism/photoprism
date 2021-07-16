import "../fixtures";
import { Clipboard } from "common/clipboard";
import Photo from "model/photo";
import Album from "model/album";
import StorageShim from "node-storage-shim";

let chai = require("chai/chai");
let assert = chai.assert;

describe("common/clipboard", () => {
  it("should construct clipboard", () => {
    const storage = new StorageShim();
    const key = "clipboard";

    const clipboard = new Clipboard(storage, key);
    assert.equal(clipboard.storageKey, "clipboard");
    assert.equal(clipboard.selection, "");
  });

  it("should toggle model", () => {
    const storage = new StorageShim();
    const key = "clipboard";

    const clipboard = new Clipboard(storage, key);
    clipboard.clear();
    clipboard.toggle();
    assert.equal(clipboard.storageKey, "clipboard");
    assert.equal(clipboard.selection, "");

    const values = { ID: 5, UID: "ABC123", Title: "Crazy Cat" };
    const photo = new Photo(values);
    clipboard.toggle(photo);
    assert.equal(clipboard.selection[0], "ABC123");
    const values2 = { ID: 8, UID: "ABC124", Title: "Crazy Cat" };
    const photo2 = new Photo(values2);
    clipboard.toggle(photo2);
    assert.equal(clipboard.selection[0], "ABC123");
    assert.equal(clipboard.selection[1], "ABC124");
    clipboard.toggle(photo);
    assert.equal(clipboard.selection[0], "ABC124");
  });

  it("should toggle id", () => {
    const storage = new StorageShim();
    const key = "clipboard";

    const clipboard = new Clipboard(storage, key);
    clipboard.clear();
    clipboard.toggleId(3);
    assert.equal(clipboard.selection[0], 3);
    clipboard.toggleId(3);
    assert.equal(clipboard.selection, "");
  });

  it("should add model", () => {
    const storage = new StorageShim();
    const key = "clipboard";

    const clipboard = new Clipboard(storage, key);
    clipboard.clear();
    clipboard.add();
    assert.equal(clipboard.storageKey, "clipboard");
    assert.equal(clipboard.selection, "");

    const values = { ID: 5, UID: "ABC124", Title: "Crazy Cat" };
    const photo = new Photo(values);
    clipboard.add(photo);
    assert.equal(clipboard.selection[0], "ABC124");
    clipboard.add(photo);
    assert.equal(clipboard.selection[0], "ABC124");
  });

  it("should add id", () => {
    const storage = new StorageShim();
    const key = "clipboard";

    const clipboard = new Clipboard(storage, key);
    clipboard.clear();
    clipboard.addId(99);
    assert.equal(clipboard.selection[0], 99);
  });

  it("should test whether clipboard has model", () => {
    const storage = new StorageShim();
    const key = "clipboard";

    const clipboard = new Clipboard(storage, key);
    clipboard.clear();
    clipboard.has();
    assert.equal(clipboard.storageKey, "clipboard");
    assert.equal(clipboard.selection, "");

    const values = { ID: 5, UID: "ABC124", Title: "Crazy Cat" };
    const photo = new Photo(values);
    clipboard.add(photo);
    assert.equal(clipboard.selection[0], "ABC124");
    const result = clipboard.has(photo);
    assert.equal(result, true);
    const values2 = { ID: 5, Title: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values2);
    const result2 = clipboard.has(album);
    assert.equal(result2, false);
  });

  it("should test whether clipboard has id", () => {
    const storage = new StorageShim();
    const key = "clipboard";

    const clipboard = new Clipboard(storage, key);
    clipboard.clear();
    clipboard.addId(77);
    assert.equal(clipboard.hasId(77), true);
    assert.equal(clipboard.hasId(78), false);
  });

  it("should remove model", () => {
    const storage = new StorageShim();
    const key = "clipboard";

    const clipboard = new Clipboard(storage, key);
    clipboard.clear();
    clipboard.remove();
    assert.equal(clipboard.storageKey, "clipboard");
    assert.equal(clipboard.selection, "");

    const values = { ID: 5, UID: "ABC123", Title: "Crazy Cat" };
    const photo = new Photo(values);
    clipboard.add(photo);
    assert.equal(clipboard.selection[0], "ABC123");

    clipboard.remove(photo);
    assert.equal(clipboard.selection, "");
    const values2 = { ID: 5, Name: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values2);
    clipboard.remove(album);
    assert.equal(clipboard.selection, "");
  });

  it("should set and get ids", () => {
    const storage = new StorageShim();
    const key = "clipboard";

    const clipboard = new Clipboard(storage, key);
    clipboard.clear();
    clipboard.setIds(8);
    assert.equal(clipboard.selection, "");
    clipboard.setIds([5, 6, 9]);
    assert.equal(clipboard.selection[0], 5);
    assert.equal(clipboard.selection[2], 9);
    const result = clipboard.getIds();
    assert.equal(result[1], 6);
    assert.equal(result.length, 3);
  });

  it("should clear", () => {
    const storage = new StorageShim();
    const key = "clipboard";

    const clipboard = new Clipboard(storage, key);
    clipboard.clear();
    clipboard.setIds([5, 6, 9]);
    assert.equal(clipboard.selection[0], 5);
    clipboard.clear();
    assert.equal(clipboard.selection, "");
  });

  it("should add range", () => {
    const storage = new StorageShim();
    const key = "clipboard";
    const clipboard = new Clipboard(storage, key);
    clipboard.clear();
    const values = { ID: 5, UID: "ABC124", Title: "Crazy Cat" };
    const photo = new Photo(values);
    const values2 = { ID: 6, UID: "ABC125", Title: "Crazy Dog" };
    const photo2 = new Photo(values2);
    const values3 = { ID: 7, UID: "ABC128", Title: "Cute Dog" };
    const photo3 = new Photo(values3);
    const values4 = { ID: 8, UID: "ABC129", Title: "Turtle" };
    const photo4 = new Photo(values4);
    const Photos = [photo, photo2, photo3, photo4];
    clipboard.addRange(2);
    assert.equal(clipboard.selection.length, 0);
    clipboard.clear();
    clipboard.addRange(2, Photos);
    assert.equal(clipboard.selection[0], "ABC128");
    assert.equal(clipboard.selection.length, 1);
    clipboard.addRange(1, Photos);
    assert.equal(clipboard.selection.length, 2);
    assert.equal(clipboard.selection[0], "ABC128");
    assert.equal(clipboard.selection[1], "ABC125");
    clipboard.clear();
    clipboard.add(photo);
    assert.equal(clipboard.selection.length, 1);
    clipboard.addRange(3, Photos);
    assert.equal(clipboard.selection.length, 4);
  });
});
