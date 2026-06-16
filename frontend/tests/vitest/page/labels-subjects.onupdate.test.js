// Targets the onUpdate switch and the refetchResults helper in
// page/labels.vue and page/people/recognized.vue. Calling the Options
// API methods directly with a stub `this` exercises the dispatch logic
// in isolation, pinning the contract that labels.updated and
// subjects.updated are UID-only signals answered by one uid-filtered
// query instead of a full result refresh.
import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";

import PPageLabels from "page/labels.vue";
import PPageSubjects from "page/people/recognized.vue";
import Label from "model/label";
import Subject from "model/subject";

// Captures the surface of `this` that the handlers touch.
function newStub() {
  return {
    listen: true,
    dirty: false,
    results: [],
    refresh: vi.fn(),
    refetchResults: vi.fn(),
    removeSelection: vi.fn(),
  };
}

describe("page/labels.vue onUpdate", () => {
  const onUpdate = PPageLabels.methods.onUpdate;

  let warnSpy;
  beforeEach(() => {
    warnSpy = vi.spyOn(console, "warn").mockImplementation(() => {});
  });

  it("delegates labels.updated to the uid-scoped refetch", () => {
    const stub = newStub();

    onUpdate.call(stub, "labels.updated", { entities: ["label-1"] });

    expect(stub.refetchResults).toHaveBeenCalledWith(["label-1"]);
    expect(stub.refresh).not.toHaveBeenCalled();
    expect(warnSpy).not.toHaveBeenCalled();
  });

  it("marks the list dirty on labels.created", () => {
    const stub = newStub();

    onUpdate.call(stub, "labels.created", { entities: ["label-2"] });

    expect(stub.dirty).toBe(true);
    expect(stub.refetchResults).not.toHaveBeenCalled();
  });

  it("removes deleted labels from results and selection", () => {
    const stub = newStub();
    stub.results = [{ UID: "label-1" }, { UID: "label-2" }];

    onUpdate.call(stub, "labels.deleted", { entities: ["label-1"] });

    expect(stub.dirty).toBe(true);
    expect(stub.results.map((m) => m.UID)).toEqual(["label-2"]);
    expect(stub.removeSelection).toHaveBeenCalledWith("label-1");
  });

  it("ignores events when listen=false", () => {
    const stub = newStub();
    stub.listen = false;

    onUpdate.call(stub, "labels.updated", { entities: ["label-1"] });

    expect(stub.refetchResults).not.toHaveBeenCalled();
  });
});

describe("page/labels.vue refetchResults", () => {
  const refetchResults = PPageLabels.methods.refetchResults;

  let searchSpy;
  afterEach(() => {
    searchSpy?.mockRestore();
  });

  it("patches affected loaded labels through one uid-scoped search", async () => {
    const stub = newStub();
    stub.results = [{ UID: "label-1", Name: "Old", PhotoCount: 1 }];
    searchSpy = vi.spyOn(Label, "search").mockResolvedValue({ models: [{ UID: "label-1", Name: "New", PhotoCount: 2 }] });

    refetchResults.call(stub, ["label-1", "label-not-loaded"]);
    await Promise.resolve();
    await Promise.resolve();

    expect(searchSpy).toHaveBeenCalledTimes(1);
    expect(searchSpy).toHaveBeenCalledWith({ uid: "label-1", count: 1 });
    expect(stub.results[0].Name).toBe("New");
    expect(stub.results[0].PhotoCount).toBe(2);
    expect(stub.dirty).toBe(false);
  });

  it("removes labels the scoped search no longer returns", async () => {
    const stub = newStub();
    stub.results = [{ UID: "label-gone" }];
    searchSpy = vi.spyOn(Label, "search").mockResolvedValue({ models: [] });

    refetchResults.call(stub, ["label-gone"]);
    await Promise.resolve();
    await Promise.resolve();

    expect(stub.results).toEqual([]);
    expect(stub.removeSelection).toHaveBeenCalledWith("label-gone");
  });

  it("does nothing when no affected label is loaded", () => {
    const stub = newStub();
    searchSpy = vi.spyOn(Label, "search").mockResolvedValue({ models: [] });

    refetchResults.call(stub, ["label-1"]);

    expect(searchSpy).not.toHaveBeenCalled();
  });

  it("falls back to the dirty flag for oversized batches", () => {
    const stub = newStub();
    const uids = Array.from({ length: 51 }, (_, i) => `label-${i}`);
    stub.results = uids.map((uid) => ({ UID: uid }));
    searchSpy = vi.spyOn(Label, "search").mockResolvedValue({ models: [] });

    refetchResults.call(stub, uids);

    expect(searchSpy).not.toHaveBeenCalled();
    expect(stub.dirty).toBe(true);
  });

  it("marks the results dirty when the refetch fails", async () => {
    const stub = newStub();
    stub.results = [{ UID: "label-1" }];
    searchSpy = vi.spyOn(Label, "search").mockRejectedValue(new Error("offline"));

    refetchResults.call(stub, ["label-1"]);
    await Promise.resolve();
    await Promise.resolve();

    expect(stub.dirty).toBe(true);
  });
});

describe("page/people/recognized.vue onUpdate", () => {
  const onUpdate = PPageSubjects.methods.onUpdate;

  it("delegates subjects.updated to the uid-scoped refetch", () => {
    const stub = newStub();

    onUpdate.call(stub, "subjects.updated", { entities: ["subj-1"] });

    expect(stub.refetchResults).toHaveBeenCalledWith(["subj-1"]);
    expect(stub.refresh).not.toHaveBeenCalled();
  });

  it("marks the list dirty on subjects.created", () => {
    const stub = newStub();

    onUpdate.call(stub, "subjects.created", { entities: ["subj-2"] });

    expect(stub.dirty).toBe(true);
    expect(stub.refetchResults).not.toHaveBeenCalled();
  });

  it("removes deleted subjects from results and selection", () => {
    const stub = newStub();
    stub.results = [{ UID: "subj-1" }, { UID: "subj-2" }];

    onUpdate.call(stub, "subjects.deleted", { entities: ["subj-1"] });

    expect(stub.dirty).toBe(true);
    expect(stub.results.map((m) => m.UID)).toEqual(["subj-2"]);
    expect(stub.removeSelection).toHaveBeenCalledWith("subj-1");
  });

  it("ignores events when listen=false", () => {
    const stub = newStub();
    stub.listen = false;

    onUpdate.call(stub, "subjects.updated", { entities: ["subj-1"] });

    expect(stub.refetchResults).not.toHaveBeenCalled();
  });
});

describe("page/people/recognized.vue refetchResults", () => {
  const refetchResults = PPageSubjects.methods.refetchResults;

  let searchSpy;
  afterEach(() => {
    searchSpy?.mockRestore();
  });

  it("patches affected loaded subjects through one uid-scoped search", async () => {
    const stub = newStub();
    stub.results = [{ UID: "subj-1", Name: "Old Name", Favorite: false }];
    searchSpy = vi.spyOn(Subject, "search").mockResolvedValue({ models: [{ UID: "subj-1", Name: "New Name", Favorite: true }] });

    refetchResults.call(stub, ["subj-1"]);
    await Promise.resolve();
    await Promise.resolve();

    expect(searchSpy).toHaveBeenCalledWith({ uid: "subj-1", count: 1 });
    expect(stub.results[0].Name).toBe("New Name");
    expect(stub.results[0].Favorite).toBe(true);
    expect(stub.dirty).toBe(false);
  });

  it("removes subjects the scoped search no longer returns", async () => {
    const stub = newStub();
    stub.results = [{ UID: "subj-gone" }];
    searchSpy = vi.spyOn(Subject, "search").mockResolvedValue({ models: [] });

    refetchResults.call(stub, ["subj-gone"]);
    await Promise.resolve();
    await Promise.resolve();

    expect(stub.results).toEqual([]);
    expect(stub.removeSelection).toHaveBeenCalledWith("subj-gone");
  });

  it("marks the results dirty when the refetch fails", async () => {
    const stub = newStub();
    stub.results = [{ UID: "subj-1" }];
    searchSpy = vi.spyOn(Subject, "search").mockRejectedValue(new Error("offline"));

    refetchResults.call(stub, ["subj-1"]);
    await Promise.resolve();
    await Promise.resolve();

    expect(stub.dirty).toBe(true);
  });
});
