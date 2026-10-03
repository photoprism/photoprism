import { describe, it, expect, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { h } from "vue";
import { VTextField } from "vuetify/components";

import PInputAction from "component/input/action.vue";

// mountAction mounts the component with a fixed icon and label.
const mountAction = () => mount(PInputAction, { props: { icon: "mdi-tune", label: "Search Filters" } });

describe("component/input/action.vue", () => {
  it("renders a focusable button named by its label", () => {
    const icon = mountAction().find("i");
    expect(icon.classes()).toContain("mdi-tune");
    expect(icon.attributes("role")).toBe("button");
    expect(icon.attributes("tabindex")).toBe("0");
    expect(icon.attributes("aria-label")).toBe("Search Filters");
  });
  it("emits click when clicked", async () => {
    const wrapper = mountAction();
    await wrapper.find("i").trigger("click");
    expect(wrapper.emitted("click")).toHaveLength(1);
  });
  it("emits click for Enter and Space but not for other keys", async () => {
    const wrapper = mountAction();
    const icon = wrapper.find("i");
    await icon.trigger("keydown", { key: "Enter" });
    await icon.trigger("keydown", { key: " " });
    await icon.trigger("keydown", { key: "a" });
    expect(wrapper.emitted("click")).toHaveLength(2);
  });
  it("keeps Enter and Space from reaching the field's key handlers", async () => {
    const onKeyup = vi.fn();
    const wrapper = mount({ render: () => h("div", { onKeyup }, [h(PInputAction, { icon: "mdi-tune", label: "Search Filters" })]) });
    const icon = wrapper.find("i");
    await icon.trigger("keyup", { key: "Enter" });
    await icon.trigger("keyup", { key: " " });
    expect(onKeyup).not.toHaveBeenCalled();
    await icon.trigger("keyup", { key: "a" });
    expect(onKeyup).toHaveBeenCalledTimes(1);
  });
  it("keeps the focus where it is when pressed with a pointer", () => {
    const ev = new MouseEvent("mousedown", { bubbles: true, cancelable: true });
    mountAction().find("i").element.dispatchEvent(ev);
    expect(ev.defaultPrevented).toBe(true);
  });
  it("keeps its name in a field slot without a field label", () => {
    const wrapper = mount({
      render: () =>
        h(VTextField, { placeholder: "Search" }, { "prepend-inner": () => h(PInputAction, { icon: "mdi-tune", label: "Search Filters" }) }),
    });
    const icon = wrapper.find(".v-field__prepend-inner i");
    expect(icon.attributes("aria-label")).toBe("Search Filters");
  });
});
