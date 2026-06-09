import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import PLightboxSidebarToolbar from "component/lightbox/sidebar/toolbar.vue";

describe("PLightboxSidebarToolbar", () => {
  // Default editing=false / canUndo=false: only the pencil renders.
  it("renders the pencil icon when not editing", () => {
    const w = mount(PLightboxSidebarToolbar);
    expect(w.find(".meta-inline-pencil").exists()).toBe(true);
    expect(w.find(".meta-inline-confirm").exists()).toBe(false);
    expect(w.find(".meta-inline-undo").exists()).toBe(false);
  });

  // editing=true, canUndo=false: only the check icon renders (inline-text
  // editor pattern — pencil ↔ check toggle, no undo).
  it("renders the check icon when editing and canUndo is false", () => {
    const w = mount(PLightboxSidebarToolbar, { props: { editing: true } });
    expect(w.find(".meta-inline-confirm").exists()).toBe(true);
    expect(w.find(".meta-inline-pencil").exists()).toBe(false);
    expect(w.find(".meta-inline-undo").exists()).toBe(false);
  });

  // editing=true, canUndo=true: chip-section toolbar pattern — undo icon
  // renders alongside the check icon so users can revert pending removals.
  it("renders both undo and check icons when editing and canUndo are true", () => {
    const w = mount(PLightboxSidebarToolbar, { props: { editing: true, canUndo: true } });
    expect(w.find(".meta-inline-undo").exists()).toBe(true);
    expect(w.find(".meta-inline-confirm").exists()).toBe(true);
    expect(w.find(".meta-inline-pencil").exists()).toBe(false);
  });

  // canUndo only takes effect while editing — outside edit mode the pencil
  // wins so we never show a bare Undo icon with no companion action.
  it("suppresses the undo icon when canUndo is true but editing is false", () => {
    const w = mount(PLightboxSidebarToolbar, { props: { editing: false, canUndo: true } });
    expect(w.find(".meta-inline-pencil").exists()).toBe(true);
    expect(w.find(".meta-inline-undo").exists()).toBe(false);
    expect(w.find(".meta-inline-confirm").exists()).toBe(false);
  });

  it("emits confirm when the check icon is clicked", async () => {
    const w = mount(PLightboxSidebarToolbar, { props: { editing: true } });
    await w.find(".meta-inline-confirm").trigger("click");
    expect(w.emitted("confirm")).toBeTruthy();
    expect(w.emitted("confirm")).toHaveLength(1);
  });

  it("emits start when the pencil icon is clicked", async () => {
    const w = mount(PLightboxSidebarToolbar);
    await w.find(".meta-inline-pencil").trigger("click");
    expect(w.emitted("start")).toBeTruthy();
    expect(w.emitted("start")).toHaveLength(1);
  });

  it("emits undo when the undo icon is clicked", async () => {
    const w = mount(PLightboxSidebarToolbar, { props: { editing: true, canUndo: true } });
    await w.find(".meta-inline-undo").trigger("click");
    expect(w.emitted("undo")).toBeTruthy();
    expect(w.emitted("undo")).toHaveLength(1);
    // Clicking undo must NOT also fire confirm (event handlers carry .stop).
    expect(w.emitted("confirm")).toBeFalsy();
  });
});
