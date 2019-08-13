import Alert from "common/alert";
let sinon = require("sinon");

describe("common/alert", () => {
    it("should call alert.info",  () => {
        let spy = sinon.spy(Alert, "info");
        Alert.info("message");
        sinon.assert.calledOnce(spy);
        spy.resetHistory();
    });

    it("should call alert.warning",  () => {
        let spy = sinon.spy(Alert, "warning");
        Alert.warning("message");
        sinon.assert.calledOnce(spy);
        spy.resetHistory();
    });

    it("should call alert.error",  () => {
        let spy = sinon.spy(Alert, "error");
        Alert.error("message");
        sinon.assert.calledOnce(spy);
        spy.resetHistory();
    });

    it("should call alert.success",  () => {
        let spy = sinon.spy(Alert, "success");
        Alert.success("message");
        sinon.assert.calledOnce(spy);
        spy.resetHistory();
    });
});