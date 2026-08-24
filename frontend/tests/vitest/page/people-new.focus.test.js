// Targets onFocusName in page/people/new.vue. Calling the Options API method
// directly with a stub `this` pins the contract the template relies on: exactly
// one tile id is marked at a time, so only that tile's combobox is handed the
// suggestion list.
import { describe, it, expect, vi } from "vitest";

import PPageFaces from "page/people/new.vue";

const onFocusName = PPageFaces.methods.onFocusName;

// Captures the surface of `this` that the handler touches.
function newStub() {
  return {
    focused: "",
    people: [{ UID: "ps-1", Name: "Alpha" }],
    noPeople: [],
    loadPeople: vi.fn(() => Promise.resolve()),
  };
}

describe("page/people/new.vue onFocusName", () => {
  it("Success", () => {
    const stub = newStub();
    onFocusName.call(stub, { ID: "FACE1" });
    expect(stub.focused).toBe("FACE1");
    expect(stub.loadPeople).toHaveBeenCalledTimes(1);
  });
  it("MovesToTheNextTile", () => {
    const stub = newStub();
    onFocusName.call(stub, { ID: "FACE1" });
    onFocusName.call(stub, { ID: "FACE2" });
    expect(stub.focused).toBe("FACE2");
    expect(stub.loadPeople).toHaveBeenCalledTimes(2);
  });
  it("InvalidRequest", () => {
    const stub = newStub();
    stub.focused = "FACE1";
    onFocusName.call(stub, undefined);
    expect(stub.focused).toBe("");
    onFocusName.call(stub, { ID: "" });
    expect(stub.focused).toBe("");
    expect(stub.loadPeople).toHaveBeenCalledTimes(2);
  });
  it("ReturnsTheSuggestionRequest", async () => {
    const stub = newStub();
    await expect(onFocusName.call(stub, { ID: "FACE1" })).resolves.toBeUndefined();
  });
});

describe("page/people/new.vue data", () => {
  it("StartsWithNoFocusedTile", () => {
    const data = PPageFaces.data.call({
      $route: { query: {}, name: "people_faces" },
      $config: { values: {} },
      sortOrder: PPageFaces.methods.sortOrder,
    });
    expect(data.focused).toBe("");
    expect(data.people).toEqual([]);
    expect(data.noPeople).toEqual([]);
  });
});
