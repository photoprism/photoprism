import "../fixtures";
import {
  canUseAv1,
  canUseAvc,
  canUseHevc,
  canUseOGV,
  canUseVideo,
  canUseVP8,
  canUseVP9,
  canUseWebM,
} from "common/caniuse";

let chai = require("chai/chai");
let assert = chai.assert;

describe("common/caniuse", () => {
  it("canUseVideo", () => {
    assert.equal(canUseVideo, true);
  });

  it("canUseAvc", () => {
    assert.equal(canUseAvc, true);
  });

  it("canUseOGV", () => {
    assert.equal(canUseOGV, true);
  });

  it("canUseVP8", () => {
    assert.equal(canUseVP8, true);
  });

  it("canUseVP9", () => {
    assert.equal(canUseVP9, true);
  });

  it("canUseAv1", () => {
    assert.equal(canUseAv1, true);
  });

  it("canUseWebM", () => {
    assert.equal(canUseWebM, true);
  });

  it("canUseHevc", () => {
    assert.equal(canUseHevc, false);
  });
});
