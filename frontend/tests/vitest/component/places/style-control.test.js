import { describe, it, expect } from "vitest";

import MapStyleControl from "component/places/style-control";

describe("component/places/style-control", () => {
  it("names the style switcher button", () => {
    const control = new MapStyleControl(null, null, () => {});
    control.onAdd({});
    expect(control.styleButton.title).toBe("Map Style");
    expect(control.styleButton.getAttribute("aria-label")).toBe("Map Style");
  });
});
