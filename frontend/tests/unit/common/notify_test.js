import Notify from "common/notify";
let sinon = require("sinon");

describe("common/alert", () => {
    it("should call alert.info",  () => {
        let spy = sinon.spy(Notify, "info");
        Notify.info("message");
        sinon.assert.calledOnce(spy);
        spy.resetHistory();
    });

    it("should call alert.warning",  () => {
        let spy = sinon.spy(Notify, "warning");
        Notify.warning("message");
        sinon.assert.calledOnce(spy);
        spy.resetHistory();
    });

    it("should call alert.error",  () => {
        let spy = sinon.spy(Notify, "error");
        Notify.error("message");
        sinon.assert.calledOnce(spy);
        spy.resetHistory();
    });

    it("should call alert.success",  () => {
        let spy = sinon.spy(Notify, "success");
        Notify.success("message");
        sinon.assert.calledOnce(spy);
        spy.resetHistory();
    });
});
