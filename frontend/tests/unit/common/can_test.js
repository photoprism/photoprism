import "../fixtures";
import * as can from "common/can";

let chai = require("chai/chai");
let assert = chai.assert;

describe("common/can", () => {
  it("useVideo", () => {
    assert.equal(can.useVideo, true);
  });

  it("useMp4Avc", () => {
    assert.equal(can.useMp4Avc, true);
  });

  it("useMp4Hvc", () => {
    assert.equal(can.useMp4Hvc, false);
  });

  it("useMp4Hev", () => {
    assert.equal(can.useMp4Hev, false);
  });

  it("useMp4Vvc", () => {
    assert.equal(can.useMp4Vvc, false);
  });

  it("useMp4Evc", () => {
    assert.equal(can.useMp4Evc, false);
  });

  it("useMp4Av1", () => {
    assert.equal(can.useMp4Av1, true);
  });

  it("useVP8", () => {
    assert.equal(can.useVP8, true);
  });

  it("useVP9", () => {
    assert.equal(can.useVP9, true);
  });

  it("useWebmAv1", () => {
    assert.equal(can.useWebmAv1, true);
  });

  it("useMkvAv1", () => {
    assert.equal(can.useMkvAv1, false);
  });

  it("useWebM", () => {
    assert.equal(can.useWebM, true);
  });

  it("useTheora", () => {
    assert.equal(can.useTheora, true);
  });
});
