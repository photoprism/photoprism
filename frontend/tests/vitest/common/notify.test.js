import { describe, it, expect, vi } from "vitest";
import "../fixtures";
import $notify from "common/notify";
import $event from "common/event";

describe("common/alert", () => {
  it("should call alert.info", () => {
    $notify.info("message");
  });

  it("should call alert.warning", () => {
    $notify.warn("message");
  });

  it("should call alert.error", () => {
    $notify.error("message");
  });

  it("should call alert.success", () => {
    $notify.success("message");
  });

  it("error forwards an optional message id and params for UI-locale rendering", async () => {
    const handler = vi.fn();
    const token = $event.subscribe("notify.error", handler);
    $notify.error("fallback", "Registration disabled", ["x"]);
    // pubsub.js delivers asynchronously, so wait a tick before asserting.
    await new Promise((resolve) => setTimeout(resolve, 0));
    $event.unsubscribe(token);
    expect(handler).toHaveBeenCalledWith("notify.error", {
      message: "fallback",
      messageId: "Registration disabled",
      messageParams: ["x"],
    });
  });

  it("should call wait", () => {
    $notify.wait();
  });
});
