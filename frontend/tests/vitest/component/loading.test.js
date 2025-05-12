import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import PLoading from "component/loading.vue";

describe("PLoading component", () => {
  it("should render correctly", () => {
    const wrapper = mount(PLoading);

    // Check if component renders
    expect(wrapper.vm).toBeTruthy();

    // Check if the progress circular element exists
    const progressCircular = wrapper.find(".vprogresscircular-stub");
    expect(progressCircular.exists()).toBe(true);
  });
});
