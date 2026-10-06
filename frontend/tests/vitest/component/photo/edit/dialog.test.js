import { describe, it, expect, vi } from "vitest";
import PPhotoEditDialog from "component/photo/edit/dialog.vue";
import Photo from "model/photo";

const uid = "ps6sg6be2lvl0yh7";

// makeContext returns a dialog instance stand-in whose open photo refetches the given server values.
const makeContext = (server) => {
  const model = new Photo({ UID: uid, Title: "Saved Title", Caption: "Saved caption", Details: { Subject: "Saved subject", Notes: "Saved notes" } });
  model.find = vi.fn(() => Promise.resolve(new Photo({ UID: uid, ...server })));

  return { model, loading: false };
};

// update delivers a photos.updated event to the dialog and waits for its refetch to settle.
const update = async (ctx, entities = [uid]) => {
  PPhotoEditDialog.methods.onUpdate.call(ctx, "photos.updated", { entities });
  await Promise.resolve();
  await Promise.resolve();
};

describe("component/photo/edit/dialog", () => {
  describe("onUpdate", () => {
    it("applies server values to fields without unsaved edits", async () => {
      const ctx = makeContext({ Title: "Server Title", Caption: "Server caption" });

      await update(ctx);

      expect(ctx.model.Title).toBe("Server Title");
      expect(ctx.model.Caption).toBe("Server caption");
      expect(ctx.model.wasChanged()).toBe(false);
    });

    it("keeps unsaved edits and updates the other fields", async () => {
      const ctx = makeContext({ Title: "Server Title", Caption: "Server caption", Details: { Subject: "Server subject", Notes: "Server notes" } });
      ctx.model.Caption = "Typed caption";
      ctx.model.Details.Subject = "Typed subject";

      await update(ctx);

      expect(ctx.model.Title).toBe("Server Title");
      expect(ctx.model.Details.Notes).toBe("Server notes");
      expect(ctx.model.Details.Subject).toBe("Typed subject");
      expect(ctx.model.Caption).toBe("Typed caption");
      expect(ctx.model.originalValue("Caption")).toBe("Saved caption");
      expect(ctx.model.wasChanged()).toBe(true);
    });

    it("ignores events for other photos", async () => {
      const ctx = makeContext({ Title: "Server Title", Caption: "Server caption" });

      await update(ctx, ["ps6sg6be2lvl0yh8"]);

      expect(ctx.model.find).not.toHaveBeenCalled();
      expect(ctx.model.Title).toBe("Saved Title");
    });

    it("applies values of photos without a title", async () => {
      const ctx = makeContext({ Title: "", Caption: "Server caption" });

      await update(ctx);

      expect(ctx.model.Title).toBe("");
      expect(ctx.model.Caption).toBe("Server caption");
      expect(ctx.model.wasChanged()).toBe(false);
    });

    it("ignores responses for another photo", async () => {
      const ctx = makeContext({ UID: "ps6sg6be2lvl0yh8", Title: "Other Title", Caption: "Other caption" });

      await update(ctx);

      expect(ctx.model.Title).toBe("Saved Title");
      expect(ctx.model.Caption).toBe("Saved caption");
    });
  });
});
