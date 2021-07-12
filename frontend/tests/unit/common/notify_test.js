import "../fixtures";
import Notify from "common/notify";

describe("common/alert", () => {
  it("should call alert.info", () => {
    Notify.info("message");
  });

  it("should call alert.warning", () => {
    Notify.warn("message");
  });

  it("should call alert.error", () => {
    Notify.error("message");
  });

  it("should call alert.success", () => {
    Notify.success("message");
  });

  it("should call wait", () => {
    Notify.wait();
  });
});
