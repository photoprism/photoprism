import Notify from "common/notify";
let sinon = require("sinon");

let chai = require("chai/chai");
let assert = chai.assert;

describe("common/alert", () => {

    let spywarn = sinon.spy(Notify, "warn");
    let spyerror = sinon.spy(Notify, "error");

    it("should call alert.info",  () => {
        let spy = sinon.spy(Notify, "info");
        Notify.info("message");
        sinon.assert.calledOnce(spy);
        spy.resetHistory();
    });

    it("should call alert.warning",  () => {
        Notify.warn("message");
        sinon.assert.calledOnce(spywarn);
        spywarn.resetHistory();
    });

    it("should call alert.error",  () => {
        Notify.error("message");
        sinon.assert.calledOnce(spyerror);
        spyerror.resetHistory();
    });

    it("should call alert.success",  () => {
        let spy = sinon.spy(Notify, "success");
        Notify.success("message");
        sinon.assert.calledOnce(spy);
        spy.resetHistory();
    });

    it("should call alert.logout",  () => {
        let spy = sinon.spy(Notify, "logout");
        Notify.logout("message");
        sinon.assert.calledOnce(spy);
        spy.resetHistory();
    });

    it("should call wait",  () => {
        Notify.wait();
        sinon.assert.calledOnce(spywarn);
        spywarn.resetHistory();
    });
});
