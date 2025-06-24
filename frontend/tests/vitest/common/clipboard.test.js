import { describe, it, expect } from "vitest";
import "../fixtures";
import { Clipboard } from "common/clipboard";
import Photo from "model/photo";
import Album from "model/album";
import StorageShim from "node-storage-shim";

describe("common/clipboard", () => {
  it("should construct clipboard", () => {
    const storage = new StorageShim();
    const key = "clipboard";

    const clipboard = new Clipboard(storage, key);
    expect(clipboard.storageKey).toBe("clipboard");
    expect(clipboard.selection).toEqual([]);
  });

  it("should toggle model", () => {
    const storage = new StorageShim();
    const key = "clipboard";

    const clipboard = new Clipboard(storage, key);
    clipboard.clear();
    clipboard.toggle();
    expect(clipboard.storageKey).toBe("clipboard");
    expect(clipboard.selection).toEqual([]);

    const values = { ID: 5, UID: "ABC123", Title: "Crazy Cat" };
    const photo = new Photo(values);
    clipboard.toggle(photo);
    expect(clipboard.selection[0]).toBe("ABC123");
    const values2 = { ID: 8, UID: "ABC124", Title: "Crazy Cat" };
    const photo2 = new Photo(values2);
    clipboard.toggle(photo2);
    expect(clipboard.selection[0]).toBe("ABC123");
    expect(clipboard.selection[1]).toBe("ABC124");
    clipboard.toggle(photo);
    expect(clipboard.selection[0]).toBe("ABC124");
  });

  it("should toggle id", () => {
    const storage = new StorageShim();
    const key = "clipboard";

    const clipboard = new Clipboard(storage, key);
    clipboard.clear();
    clipboard.toggleId(3);
    expect(clipboard.selection[0]).toBe(3);
    clipboard.toggleId(3);
    expect(clipboard.selection).toEqual([]);
  });

  it("should add model", () => {
    const storage = new StorageShim();
    const key = "clipboard";

    const clipboard = new Clipboard(storage, key);
    clipboard.clear();
    clipboard.add();
    expect(clipboard.storageKey).toBe("clipboard");
    expect(clipboard.selection).toEqual([]);

    const values = { ID: 5, UID: "ABC124", Title: "Crazy Cat" };
    const photo = new Photo(values);
    clipboard.add(photo);
    expect(clipboard.selection[0]).toBe("ABC124");
    clipboard.add(photo);
    expect(clipboard.selection[0]).toBe("ABC124");
  });

  it("should add id", () => {
    const storage = new StorageShim();
    const key = "clipboard";

    const clipboard = new Clipboard(storage, key);
    clipboard.clear();
    clipboard.addId(99);
    expect(clipboard.selection[0]).toBe(99);
  });

  it("should test whether clipboard has model", () => {
    const storage = new StorageShim();
    const key = "clipboard";

    const clipboard = new Clipboard(storage, key);
    clipboard.clear();
    clipboard.has();
    expect(clipboard.storageKey).toBe("clipboard");
    expect(clipboard.selection).toEqual([]);

    const values = { ID: 5, UID: "ABC124", Title: "Crazy Cat" };
    const photo = new Photo(values);
    clipboard.add(photo);
    expect(clipboard.selection[0]).toBe("ABC124");
    const result = clipboard.has(photo);
    expect(result).toBe(true);
    const values2 = { ID: 5, Title: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values2);
    const result2 = clipboard.has(album);
    expect(result2).toBe(false);
  });

  it("should test whether clipboard has id", () => {
    const storage = new StorageShim();
    const key = "clipboard";

    const clipboard = new Clipboard(storage, key);
    clipboard.clear();
    clipboard.addId(77);
    expect(clipboard.hasId(77)).toBe(true);
    expect(clipboard.hasId(78)).toBe(false);
  });

  it("should remove model", () => {
    const storage = new StorageShim();
    const key = "clipboard";

    const clipboard = new Clipboard(storage, key);
    clipboard.clear();
    clipboard.remove();
    expect(clipboard.storageKey).toBe("clipboard");
    expect(clipboard.selection).toEqual([]);

    const values = { ID: 5, UID: "ABC123", Title: "Crazy Cat" };
    const photo = new Photo(values);
    clipboard.add(photo);
    expect(clipboard.selection[0]).toBe("ABC123");

    clipboard.remove(photo);
    expect(clipboard.selection).toEqual([]);
    const values2 = { ID: 5, Name: "Christmas 2019", Slug: "christmas-2019", UID: 66 };
    const album = new Album(values2);
    clipboard.remove(album);
    expect(clipboard.selection).toEqual([]);
  });

  it("should set and get ids", () => {
    const storage = new StorageShim();
    const key = "clipboard";

    const clipboard = new Clipboard(storage, key);
    clipboard.clear();
    clipboard.setIds(8);
    expect(clipboard.selection).toEqual([]);
    clipboard.setIds([5, 6, 9]);
    expect(clipboard.selection[0]).toBe(5);
    expect(clipboard.selection[2]).toBe(9);
    const result = clipboard.getIds();
    expect(result[1]).toBe(6);
    expect(result.length).toBe(3);
  });

  it("should clear", () => {
    const storage = new StorageShim();
    const key = "clipboard";

    const clipboard = new Clipboard(storage, key);
    clipboard.clear();
    clipboard.setIds([5, 6, 9]);
    expect(clipboard.selection[0]).toBe(5);
    clipboard.clear();
    expect(clipboard.selection).toEqual([]);
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
    expect(clipboard.selection.length).toBe(0);
    clipboard.clear();
    clipboard.addRange(2, Photos);
    expect(clipboard.selection[0]).toBe("ABC128");
    expect(clipboard.selection.length).toBe(1);
    clipboard.addRange(1, Photos);
    expect(clipboard.selection.length).toBe(2);
    expect(clipboard.selection[0]).toBe("ABC128");
    expect(clipboard.selection[1]).toBe("ABC125");
    clipboard.clear();
    clipboard.add(photo);
    expect(clipboard.selection.length).toBe(1);
    clipboard.addRange(3, Photos);
    expect(clipboard.selection.length).toBe(4);
  });
});
