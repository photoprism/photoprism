import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { flushPromises } from "@vue/test-utils";
import { Mock } from "../../fixtures";
import $notify from "common/notify";
import PPageImport from "page/library/import.vue";

// importContext returns a stub component context for startImport.
const importContext = () => ({ settings: { import: { path: "/", move: false } } });

describe("page/library/import.vue", () => {
  let notifyError;

  beforeEach(() => {
    Mock.resetHistory();
    vi.spyOn($notify, "blockUI").mockImplementation(() => {});
    vi.spyOn($notify, "unblockUI").mockImplementation(() => {});
    notifyError = vi.spyOn($notify, "error").mockImplementation(() => {});
  });

  afterEach(() => {
    Mock.resetHandlers();
    vi.restoreAllMocks();
  });

  describe("startImport", () => {
    it("shows only the server message when a migration is in progress", async () => {
      Mock.onPost("api/v1/import").reply(409, { code: 409, error: "A migration is in progress" });
      const ctx = importContext();

      PPageImport.methods.startImport.call(ctx);
      await flushPromises();

      expect(notifyError).toHaveBeenCalledWith("A migration is in progress");
      expect(notifyError).not.toHaveBeenCalledWith("Import failed");
      expect(ctx.busy).toBe(false);
    });
    it("reports other failures as a failed import", async () => {
      Mock.onPost("api/v1/import").reply(500, { code: 500, error: "Something went wrong" });

      PPageImport.methods.startImport.call(importContext());
      await flushPromises();

      expect(notifyError).toHaveBeenCalledWith("Import failed");
    });
  });
});
