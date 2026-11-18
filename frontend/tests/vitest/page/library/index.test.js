import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { flushPromises } from "@vue/test-utils";
import { Mock } from "../../fixtures";
import $notify from "common/notify";
import PPageIndex from "page/library/index.vue";

// indexContext returns a stub component context for startIndexing.
const indexContext = () => ({ settings: { index: { path: "/", rescan: false } }, cleanup: false });

describe("page/library/index.vue", () => {
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

  describe("startIndexing", () => {
    it("shows only the server message when a migration is in progress", async () => {
      Mock.onPost("api/v1/index").reply(409, { code: 409, error: "A migration is in progress" });
      const ctx = indexContext();

      PPageIndex.methods.startIndexing.call(ctx);
      await flushPromises();

      expect(notifyError).toHaveBeenCalledWith("A migration is in progress");
      expect(notifyError).not.toHaveBeenCalledWith("Indexing failed");
      expect(ctx.busy).toBe(false);
    });
    it("reports other failures as failed indexing", async () => {
      Mock.onPost("api/v1/index").reply(500, { code: 500, error: "Something went wrong" });

      PPageIndex.methods.startIndexing.call(indexContext());
      await flushPromises();

      expect(notifyError).toHaveBeenCalledWith("Indexing failed");
    });
  });
});
