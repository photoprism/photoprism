import "../fixtures";
import $notify from "common/notify";

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

  it("should call wait", () => {
    $notify.wait();
  });
});
