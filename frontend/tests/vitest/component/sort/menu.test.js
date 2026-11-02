import { describe, it, expect, afterEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import PSortMenu from "component/sort/menu.vue";
import { photoOptions, albumOptions } from "common/sort";

const wrappers = [];

// mountMenu mounts the actual Vuetify menu and tracks its cleanup.
function mountMenu(props = {}) {
  const wrapper = mount(PSortMenu, {
    attachTo: document.body,
    props: { options: photoOptions("browse"), modelValue: { order: "newest", reverse: false }, ...props },
  });
  wrappers.push(wrapper);
  return wrapper;
}

afterEach(() => {
  wrappers.splice(0).forEach((wrapper) => wrapper.unmount());
  document.body.innerHTML = "";
});

describe("sort menu", () => {
  it("exposes the full current direction in the button label", () => {
    const wrapper = mountMenu();
    expect(wrapper.get(".action-sort").attributes("aria-label")).toBe("Sorted by Date Taken, newest first");
    expect(wrapper.get(".action-sort").attributes("title")).toBe("Sorted by Date Taken, newest first");
  });
  it("opens the real menu and toggles the active date row", async () => {
    const wrapper = mountMenu();
    await wrapper.get(".action-sort").trigger("click");
    await flushPromises();
    const row = document.querySelector('.v-list-item[data-sort="newest"]');
    expect(row).not.toBeNull();
    row.click();
    await flushPromises();
    expect(wrapper.emitted("update:modelValue")).toEqual([[{ order: "oldest", reverse: false }]]);
    expect(wrapper.vm.open).toBe(false);
  });
  it("uses arrow buttons without also toggling the containing row", async () => {
    const wrapper = mountMenu();
    await wrapper.get(".action-sort").trigger("click");
    await flushPromises();
    const arrow = document.querySelector('button[data-sort="size"][data-direction="ascending"]');
    expect(arrow.getAttribute("aria-label")).toBe("Sorted by File Size, smallest first");
    expect(arrow.closest(".v-list-item")).toBeNull();
    arrow.click();
    await flushPromises();
    expect(wrapper.emitted("update:modelValue")).toEqual([[{ order: "size", reverse: true }]]);
  });
  it("does not emit when the active direction or one-way option is selected", () => {
    const wrapper = mountMenu({ modelValue: { order: "added", reverse: false } });
    wrapper.vm.choose(wrapper.vm.options.find((item) => item.value === "added"));
    expect(wrapper.emitted("update:modelValue")).toBeUndefined();
    wrapper.vm.choose(wrapper.vm.options[0], false);
    expect(wrapper.emitted("update:modelValue")).toEqual([[{ order: "newest", reverse: false }]]);
  });
  it("does not show direction arrows for location or other one-way options", async () => {
    const wrapper = mountMenu({ options: albumOptions("album"), modelValue: { order: "place", reverse: false } });
    await wrapper.get(".action-sort").trigger("click");
    await flushPromises();
    expect(document.querySelector('button[data-sort="place"]')).toBeNull();
    expect(wrapper.get(".action-sort").findAll(".v-icon")).toHaveLength(1);
  });
  it("reacts to route-driven direction changes", async () => {
    const wrapper = mountMenu();
    await wrapper.setProps({ modelValue: { order: "oldest", reverse: false } });
    expect(wrapper.vm.ascending).toBe(true);
    expect(wrapper.get(".action-sort").attributes("aria-label")).toBe("Sorted by Date Taken, oldest first");
  });
  it("exposes row actions and their selected state to assistive technology", async () => {
    const wrapper = mountMenu({ modelValue: { order: "added", reverse: false } });
    await wrapper.get(".action-sort").trigger("click");
    await flushPromises();
    const row = document.querySelector('.v-list-item[data-sort="added"]');
    expect(row.getAttribute("role")).toBe("button");
    expect(row.getAttribute("aria-pressed")).toBe("true");
    expect(document.querySelector('.v-list-item[data-sort="title"]').getAttribute("aria-pressed")).toBe("false");
  });
  it("restores focus after loading and respects a deliberately focused input", async () => {
    const wrapper = mountMenu();
    wrapper.vm.restoreFocus = true;
    await wrapper.setProps({ disabled: true });
    document.body.focus();
    await wrapper.setProps({ disabled: false });
    await flushPromises();
    expect(document.activeElement).toBe(wrapper.get(".action-sort").element);
    const input = document.createElement("input");
    document.body.appendChild(input);
    input.focus();
    wrapper.vm.restoreFocus = true;
    await wrapper.setProps({ disabled: true });
    await wrapper.setProps({ disabled: false });
    await flushPromises();
    expect(document.activeElement).toBe(input);
  });
  it("keeps focus restoration pending across asynchronous route navigation", async () => {
    const wrapper = mountMenu();
    wrapper.vm.choose(wrapper.vm.options[1]);
    await flushPromises();
    expect(wrapper.vm.restoreFocus).toBe(true);
    await wrapper.setProps({ disabled: true });
    wrapper.get(".action-sort").element.blur();
    await wrapper.setProps({ disabled: false });
    await flushPromises();
    expect(document.activeElement).toBe(wrapper.get(".action-sort").element);
    expect(wrapper.vm.restoreFocus).toBe(false);
  });
  it.each(["Enter", " "])("activates an option using the %s key", async (key) => {
    const wrapper = mountMenu();
    await wrapper.get(".action-sort").trigger("click");
    await flushPromises();
    const row = document.querySelector('.v-list-item[data-sort="title"]');
    row.focus();
    row.dispatchEvent(new KeyboardEvent("keydown", { key, bubbles: true }));
    await flushPromises();
    expect(wrapper.emitted("update:modelValue")).toEqual([[{ order: "title", reverse: false }]]);
  });
  it("closes on Escape without changing the sort", async () => {
    const wrapper = mountMenu();
    await wrapper.get(".action-sort").trigger("click");
    await flushPromises();
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }));
    await flushPromises();
    expect(wrapper.vm.open).toBe(false);
    expect(wrapper.emitted("update:modelValue")).toBeUndefined();
  });
  it("disables selection while results are loading", () => {
    const wrapper = mountMenu({ disabled: true });
    expect(wrapper.get(".action-sort").attributes("disabled")).toBeDefined();
    wrapper.vm.choose(wrapper.vm.options[1]);
    expect(wrapper.emitted("update:modelValue")).toBeUndefined();
  });
});
