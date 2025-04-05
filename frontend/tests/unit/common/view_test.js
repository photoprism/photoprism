import { $view } from "common/view";

let chai = require("chai/chai");
let assert = chai.assert;

describe("common/view", () => {
  it("should return parent", () => {
    assert.equal($view.getParent(), null);
  });
  it("should return parent name", () => {
    assert.equal($view.getParentName(), "");
  });
  it("should return data", () => {
    assert.containsSubset($view.getData(), {});
  });
  it("should return number of layers", () => {
    assert.containsSubset($view.len(), 0);
  });
  it("should return if root view is active", () => {
    assert.equal($view.isRoot(), true);
  });
  it("should return if view is app", () => {
    assert.equal($view.isApp(), true);
  });
});
