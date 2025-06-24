import { describe, it, expect } from "vitest";
import { $view } from "common/view";

describe("common/view", () => {
  it("should return parent", () => {
    expect($view.getParent()).toBe(null);
  });
  it("should return parent name", () => {
    expect($view.getParentName()).toBe("");
  });
  it("should return data", () => {
    expect($view.getData()).toEqual({});
  });
  it("should return number of layers", () => {
    expect($view.len()).toBe(0);
  });
  it("should return if root view is active", () => {
    expect($view.isRoot()).toBe(true);
  });
  it("should return if view is app", () => {
    expect($view.isApp()).toBe(true);
  });
});
