import { describe, it, expect, vi } from "vitest";
import $util from "common/util";
import PNotify from "component/notify.vue";

// notifyStub returns the surface of `this` that onNotify uses.
const notifyStub = () => ({ $config: { test: false }, $util, addSuccessMessage: vi.fn(), isolateParams: PNotify.methods.isolateParams });

describe("component/notify", () => {
  it("capitalizes the first letter of a message", () => {
    const stub = notifyStub();
    PNotify.methods.onNotify.call(stub, "notify.success", { message: "dog added" });
    expect(stub.addSuccessMessage).toHaveBeenCalledWith("Dog added");
  });
  it("isolates string parameters of translated messages", () => {
    const stub = notifyStub();
    PNotify.methods.onNotify.call(stub, "notify.success", { messageId: "One entry added to %s", messageParams: ["رحلة 2026"] });
    expect(stub.addSuccessMessage).toHaveBeenCalledWith("One entry added to \u2068رحلة 2026\u2069");
    expect(PNotify.methods.isolateParams.call(stub, [3, "a"])).toEqual([3, "\u2068a\u2069"]);
    expect(PNotify.methods.isolateParams.call(stub, undefined)).toBeUndefined();
  });
  it("capitalizes the first letter of an isolated name", () => {
    const stub = notifyStub();
    PNotify.methods.onNotify.call(stub, "notify.success", { message: "\u2068dog\u2069 wurde hinzugefügt" });
    expect(stub.addSuccessMessage).toHaveBeenCalledWith("\u2068Dog\u2069 wurde hinzugefügt");
  });
});
