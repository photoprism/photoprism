import Notify from "common/notify";
let sinon = require("sinon");

let chai = require('../../../node_modules/chai/chai');
let assert = chai.assert;

describe("common/alert", () => {

    let spywarn = sinon.spy(Notify, "warn");
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

    //TODO How to access element?
    /*it("should test blocking an unblocking UI",  () => {
        const el = document.getElementById("p-busy-overlay");
        assert.equal(el.style.display, "xxx");
        Notify.blockUI();
        assert.equal(el.style.display, "xxx");
        Notify.unblockUI();
        assert.equal(el.style.display, "xxx");
    });*/
});
