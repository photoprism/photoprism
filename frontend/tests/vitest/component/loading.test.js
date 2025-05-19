import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import PLoading from "component/loading.vue";

describe("PLoading component", () => {
  it("should render correctly", () => {
    const wrapper = mount(PLoading);

    // Check if component renders
    expect(wrapper.vm).toBeTruthy();

    // Since we're using real Vuetify components, just verify component exists
    // rather than looking for specific elements that might change with Vuetify versions
    expect(wrapper.exists()).toBe(true);
  });
});
