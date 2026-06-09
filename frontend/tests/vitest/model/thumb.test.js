import { describe, it, expect, vi, beforeEach } from "vitest";
import { Mock } from "../fixtures";
import Thumb from "model/thumb";
import Photo from "model/photo";
import File from "model/file";

describe("model/thumb", () => {
  it("should get thumb defaults", () => {
    const values = {
      UID: "55",
      Title: "",
      TakenAtLocal: "",
      Caption: "",
      Favorite: false,
      Playable: false,
      Width: 0,
      Height: 0,
      DownloadUrl: "",
    };
    const thumb = new Thumb(values);
    const result = thumb.getDefaults();
    expect(result.UID).toBe("");
  });

  it("should get id", () => {
    const values = {
      UID: "55",
    };
    const thumb = new Thumb(values);
    expect(thumb.getId()).toBe("55");
  });

  it("should return hasId", () => {
    const values = {
      UID: "55",
    };
    const thumb = new Thumb(values);
    expect(thumb.hasId()).toBe(true);

    const values2 = {
      Title: "",
    };
    const thumb2 = new Thumb(values2);
    expect(thumb2.hasId()).toBe(false);
  });

  it("should toggle like", () => {
    const values = {
      UID: "55",
      Title: "",
      TakenAtLocal: "",
      Caption: "",
      Favorite: true,
      Playable: false,
      Width: 0,
      Height: 0,
      DownloadUrl: "",
    };
    const thumb = new Thumb(values);
    expect(thumb.Favorite).toBe(true);
    thumb.toggleLike();
    expect(thumb.Favorite).toBe(false);
    thumb.toggleLike();
    expect(thumb.Favorite).toBe(true);
  });

  it("should return thumb not found", () => {
    const result = Thumb.notFound();
    expect(result.UID).toBe("");
    expect(result.Favorite).toBe(false);
  });

  it("should test from file", () => {
    const values = {
      InstanceID: 5,
      UID: "ABC123",
      Name: "1/2/IMG123.jpg",
      Hash: "abc123",
      Width: 500,
      Height: 900,
    };
    const file = new File(values);

    const values2 = {
      UID: "5",
      Title: "Crazy Cat",
      TakenAt: "2012-07-08T14:45:39Z",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      Caption: "Nice description",
      Favorite: true,
    };
    const photo = new Photo(values2);
    const result = Thumb.fromFile(photo, file);
    expect(result.UID).toBe("5");
    expect(result.Caption).toBe("Nice description");
    expect(result.Width).toBe(500);
    const result2 = Thumb.fromFile();
    expect(result2.UID).toBe("");
  });

  it("should test from files", () => {
    const values2 = {
      UID: "5",
      Title: "Crazy Cat",
      TakenAt: "2012-07-08T14:45:39Z",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      Caption: "Nice description",
      Favorite: true,
    };
    const photo = new Photo(values2);

    const values3 = {
      UID: "5",
      Title: "Crazy Cat",
      TakenAt: "2012-07-08T14:45:39Z",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      Caption: "Nice description",
      Favorite: true,
    };
    const photo2 = new Photo(values3);
    const Photos = [photo, photo2];
    const result = Thumb.fromFiles(Photos);
    expect(result.length).toBe(0);
    const values4 = {
      ID: 8,
      UID: "ABC123",
      Caption: "Nice description 2",
      Hash: "abc345",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.jpg",
          Primary: true,
          FileType: "jpg",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt53",
        },
      ],
    };
    const photo3 = new Photo(values4);
    const Photos2 = [photo, photo2, photo3];
    const result2 = Thumb.fromFiles(Photos2);
    expect(result2[0].UID).toBe("ABC123");
    expect(result2[0].Caption).toBe("Nice description 2");
    expect(result2[0].Width).toBe(500);
    expect(result2.length).toBe(1);
    const values5 = {
      ID: 8,
      UID: "ABC123",
      Caption: "Nice description 2",
      Hash: "abc345",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.jpg",
          Primary: true,
          FileType: "mov",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt53",
        },
      ],
    };
    const photo4 = new Photo(values5);
    const Photos3 = [photo3, photo2, photo4];
    const result3 = Thumb.fromFiles(Photos3);
    expect(result3.length).toBe(1);
    expect(result3[0].UID).toBe("ABC123");
    expect(result3[0].Caption).toBe("Nice description 2");
    expect(result3[0].Width).toBe(500);
  });

  it("should test from files", () => {
    const Photos = [];
    const result = Thumb.fromFiles(Photos);
    expect(result).toEqual([]);
  });

  it("should test from photo", () => {
    const values = {
      ID: 8,
      UID: "ABC123",
      Caption: "Nice description 3",
      Hash: "345ggh",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.jpg",
          Primary: true,
          FileType: "jpg",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt53",
        },
      ],
    };
    const photo = new Photo(values);
    const result = Thumb.fromPhoto(photo);
    expect(result.UID).toBe("ABC123");
    expect(result.Caption).toBe("Nice description 3");
    expect(result.Width).toBe(500);
    const values3 = {
      ID: 8,
      UID: "ABC124",
      Caption: "Nice description 3",
    };
    const photo3 = new Photo(values3);
    const result2 = Thumb.fromPhoto(photo3);
    expect(result2.UID).toBe("");
    const values2 = {
      ID: 8,
      UID: "ABC123",
      Title: "Crazy Cat",
      TakenAt: "2012-07-08T14:45:39Z",
      TakenAtLocal: "2012-07-08T14:45:39Z",
      Caption: "Nice description",
      Favorite: true,
      Hash: "xdf45m",
    };
    const photo2 = new Photo(values2);
    const result3 = Thumb.fromPhoto(photo2);
    expect(result3.UID).toBe("ABC123");
    expect(result3.Title).toBe("Crazy Cat");
    expect(result3.Caption).toBe("Nice description");
  });

  it("should test from photos", () => {
    const values = {
      ID: 8,
      UID: "ABC123",
      Caption: "Nice description 3",
      Hash: "345ggh",
      Files: [
        {
          UID: "123fgb",
          Name: "1980/01/superCuteKitten.jpg",
          Primary: true,
          FileType: "jpg",
          Width: 500,
          Height: 600,
          Hash: "1xxbgdt53",
        },
      ],
    };
    const photo = new Photo(values);
    const Photos = [photo];
    const result = Thumb.fromPhotos(Photos);
    expect(result[0].UID).toBe("ABC123");
    expect(result[0].Caption).toBe("Nice description 3");
    expect(result[0].Width).toBe(500);
  });

  it("should return download url", () => {
    const values = {
      InstanceID: 5,
      UID: "ABC123",
      Hash: "54ghtfd",
      FileType: "jpg",
      Name: "1/2/IMG123.jpg",
    };
    const file = new File(values);
    expect(Thumb.downloadUrl(file)).toBe("/api/v1/dl/54ghtfd?t=2lbh9x09");
    const values2 = {
      InstanceID: 5,
      UID: "ABC123",
      FileType: "jpg",
      Name: "1/2/IMG123.jpg",
    };
    const file2 = new File(values2);
    expect(Thumb.downloadUrl(file2)).toBe("");
  });

  it("should return thumbnail url", () => {
    const values = {
      InstanceID: 5,
      UID: "ABC123",
      Hash: "54ghtfd",
      FileType: "jpg",
      Name: "1/2/IMG123.jpg",
    };
    const file = new File(values);
    expect(Thumb.thumbnailUrl(file, "abc")).toBe("/api/v1/t/54ghtfd/public/abc");
    const values2 = {
      InstanceID: 5,
      UID: "ABC123",
      Name: "1/2/IMG123.jpg",
    };
    const file2 = new File(values2);
    expect(Thumb.thumbnailUrl(file2, "bcd")).toBe("/static/img/404.jpg");
  });

  it("should calculate size", () => {
    const values = {
      InstanceID: 5,
      UID: "ABC123",
      Hash: "54ghtfd",
      FileType: "jpg",
      Width: 900,
      Height: 850,
      Name: "1/2/IMG123.jpg",
    };
    const file = new File(values);
    const result = Thumb.calculateSize(file, 600, 800);
    expect(result.width).toBe(600);
    expect(result.height).toBe(567);
    const values3 = {
      InstanceID: 5,
      UID: "ABC123",
      Hash: "54ghtfd",
      FileType: "jpg",
      Width: 750,
      Height: 850,
      Name: "1/2/IMG123.jpg",
    };
    const file3 = new File(values3);
    const result2 = Thumb.calculateSize(file3, 900, 450);
    expect(result2.width).toBe(398);
    expect(result2.height).toBe(450);
    const result4 = Thumb.calculateSize(file3, 900, 950);
    expect(result4.width).toBe(750);
    expect(result4.height).toBe(850);
  });

  describe("loadPhoto", () => {
    it("delegates to Photo.findCached for thumbs with a UID", async () => {
      const result = new Photo({ UID: "abc123", Title: "Loaded" });
      const spy = vi.spyOn(Photo, "findCached").mockResolvedValue(result);
      const thumb = new Thumb({ UID: "abc123" });
      const photo = await thumb.loadPhoto();
      expect(spy).toHaveBeenCalledWith("abc123");
      expect(photo).toBe(result);
      spy.mockRestore();
    });

    it("resolves to an empty Photo placeholder when the thumb has no UID", async () => {
      const spy = vi.spyOn(Photo, "findCached");
      const thumb = new Thumb({ UID: "" });
      const photo = await thumb.loadPhoto();
      // Returns a Photo (not null/undefined) so consumers can read
      // .X without nullable chains; UID stays empty as a "not loaded"
      // signal. findCached must NOT be called for an empty UID.
      expect(photo).toBeInstanceOf(Photo);
      expect(photo.UID).toBe("");
      expect(spy).not.toHaveBeenCalled();
      spy.mockRestore();
    });

    it("propagates ModelCacheStaleFetchError so callers' .catch handlers fire", async () => {
      const err = new Error("ModelCache: discarded stale fetch after clear()");
      err.name = "ModelCacheStaleFetchError";
      const spy = vi.spyOn(Photo, "findCached").mockRejectedValue(err);
      const thumb = new Thumb({ UID: "abc123" });
      await expect(thumb.loadPhoto()).rejects.toBe(err);
      spy.mockRestore();
    });
  });

  describe("evictPhoto", () => {
    it("calls Photo.evictCache with the thumb UID", () => {
      const spy = vi.spyOn(Photo, "evictCache");
      const thumb = new Thumb({ UID: "abc123" });
      thumb.evictPhoto();
      expect(spy).toHaveBeenCalledWith("abc123");
      spy.mockRestore();
    });

    it("is a no-op when the thumb has no UID", () => {
      const spy = vi.spyOn(Photo, "evictCache");
      const thumb = new Thumb({ UID: "" });
      thumb.evictPhoto();
      expect(spy).not.toHaveBeenCalled();
      spy.mockRestore();
    });
  });

  // archive / restore / removeFromAlbum drive the lightbox menu
  // visibility (this.model?.Archived / this.model?.Removed checks
  // around lightbox.vue:1437). The tests below pin two contracts:
  // 1. Right URL + payload shape — verified through the global Mock
  //    in fixtures.js plus Mock.history for the request log.
  // 2. Optimistic-flip + restore-previous-value rollback — a
  //    literal `false`/`true` rollback would silently promote an
  //    undefined tri-state field, and an already-archived re-archive
  //    must not flip back to "not archived" on failure.
  //
  // Rapid-fire prevention is intentionally a UI concern (see
  // $notify.blockUI("busy") in lightbox.vue onArchive / onRestore /
  // onRemoveFromAlbum) — the model stays stateless.
  //
  // Rejection tests use `vi.spyOn($api, ...).mockRejectedValueOnce`
  // because axios-mock-adapter matches handlers in registration
  // order (see node_modules/axios-mock-adapter/src/utils.js#find);
  // a `replyOnce(500)` registered after the persistent `reply(200)`
  // in fixtures.js never wins, so it can't simulate a one-off
  // failure for an endpoint that has a global success mock.
  describe("archive", () => {
    beforeEach(() => {
      Mock.history.post.length = 0;
    });

    it("flips Archived to true, posts to batch/photos/archive, and resolves on success", async () => {
      const thumb = new Thumb({ UID: "abc123" });
      // Pre-condition: Archived is undefined (not in defaults — see
      // the explicit-tri-state checks at lightbox.vue:1437).
      expect(thumb.Archived).toBeUndefined();
      const response = await thumb.archive();
      expect(response.status).toBe(200);
      expect(thumb.Archived).toBe(true);
      const calls = Mock.history.post.filter((r) => r.url === "batch/photos/archive");
      expect(calls).toHaveLength(1);
      expect(JSON.parse(calls[0].data)).toEqual({ photos: ["abc123"] });
    });

    it("restores the pre-call Archived value on rejection (was undefined)", async () => {
      const $api = (await import("common/api")).default;
      const err = new Error("offline");
      const spy = vi.spyOn($api, "post").mockRejectedValueOnce(err);
      const thumb = new Thumb({ UID: "abc123" });
      // Pre-state is undefined (default, since Archived isn't in
      // getDefaults()). A literal `false` rollback would silently
      // promote the field to a boolean — capturing prev preserves
      // the tri-state semantics the menu logic depends on.
      await expect(thumb.archive()).rejects.toBe(err);
      expect(thumb.Archived).toBeUndefined();
      spy.mockRestore();
    });

    it("preserves Archived on rejection when called on an already-archived thumb", async () => {
      // No-op archive of an already-archived photo: backend may
      // succeed or 4xx, but a literal `false` rollback would flip a
      // truthful "archived" UI to "not archived" on failure.
      const $api = (await import("common/api")).default;
      const err = new Error("offline");
      const spy = vi.spyOn($api, "post").mockRejectedValueOnce(err);
      const thumb = new Thumb({ UID: "abc123", Archived: true });
      await expect(thumb.archive()).rejects.toBe(err);
      expect(thumb.Archived).toBe(true);
      spy.mockRestore();
    });
  });

  describe("restore", () => {
    beforeEach(() => {
      Mock.history.post.length = 0;
    });

    it("flips Archived to false, posts to batch/photos/restore, and resolves on success", async () => {
      const thumb = new Thumb({ UID: "abc123", Archived: true });
      const response = await thumb.restore();
      expect(response.status).toBe(200);
      expect(thumb.Archived).toBe(false);
      const calls = Mock.history.post.filter((r) => r.url === "batch/photos/restore");
      expect(calls).toHaveLength(1);
      expect(JSON.parse(calls[0].data)).toEqual({ photos: ["abc123"] });
    });

    it("restores the pre-call Archived value on rejection (was true)", async () => {
      const $api = (await import("common/api")).default;
      const err = new Error("offline");
      const spy = vi.spyOn($api, "post").mockRejectedValueOnce(err);
      const thumb = new Thumb({ UID: "abc123", Archived: true });
      await expect(thumb.restore()).rejects.toBe(err);
      expect(thumb.Archived).toBe(true);
      spy.mockRestore();
    });

    it("preserves Archived on rejection when called on an already-restored thumb", async () => {
      // No-op restore of a non-archived photo: capturing prev
      // ensures we don't silently flip undefined → true on failure.
      const $api = (await import("common/api")).default;
      const err = new Error("offline");
      const spy = vi.spyOn($api, "post").mockRejectedValueOnce(err);
      const thumb = new Thumb({ UID: "abc123" });
      await expect(thumb.restore()).rejects.toBe(err);
      expect(thumb.Archived).toBeUndefined();
      spy.mockRestore();
    });
  });

  describe("removeFromAlbum", () => {
    beforeEach(() => {
      Mock.history.delete.length = 0;
    });

    it("flips Removed to true, DELETEs albums/:albumUID/photos, and resolves on success", async () => {
      const thumb = new Thumb({ UID: "abc123" });
      expect(thumb.Removed).toBeUndefined();
      const response = await thumb.removeFromAlbum("album-1");
      expect(response.status).toBe(200);
      expect(thumb.Removed).toBe(true);
      const calls = Mock.history.delete.filter((r) => r.url === "albums/album-1/photos");
      expect(calls).toHaveLength(1);
      expect(JSON.parse(calls[0].data)).toEqual({ photos: ["abc123"] });
    });

    it("restores the pre-call Removed value on rejection (was undefined)", async () => {
      const $api = (await import("common/api")).default;
      const err = new Error("offline");
      const spy = vi.spyOn($api, "delete").mockRejectedValueOnce(err);
      const thumb = new Thumb({ UID: "abc123" });
      await expect(thumb.removeFromAlbum("album-1")).rejects.toBe(err);
      expect(thumb.Removed).toBeUndefined();
      spy.mockRestore();
    });
  });
});
