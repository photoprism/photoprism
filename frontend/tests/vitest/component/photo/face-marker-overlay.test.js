import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { mount } from "@vue/test-utils";
import PFaceMarkerOverlay from "component/photo/face-marker-overlay.vue";

describe("PFaceMarkerOverlay", () => {
  let wrapper;

  const mountOverlay = (props = {}) =>
    mount(PFaceMarkerOverlay, {
      props: {
        markers: [],
        mode: "draw",
        ...props,
      },
    });

  beforeEach(() => {
    wrapper = mountOverlay();
    wrapper.vm.bounds = { left: 0, top: 0, width: 100, height: 80 };
    wrapper.vm.$refs.root = {
      getBoundingClientRect: () => ({ left: 0, top: 0 }),
    };
  });

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount();
    }
  });

  it("emits normalized face marker coordinates", () => {
    wrapper.vm.pending = { x: 10, y: 20, w: 30, h: 30 };

    wrapper.vm.onConfirmPending();

    expect(wrapper.emitted("create")).toEqual([
      [
        {
          X: 0.1,
          Y: 0.25,
          W: 0.3,
          H: 0.375,
        },
      ],
    ]);
  });

  it("creates a square draft from pointer drag", () => {
    wrapper.vm.onPointerDown({
      button: 0,
      pointerId: 1,
      clientX: 10,
      clientY: 20,
      stopPropagation() {},
      preventDefault() {},
    });

    wrapper.vm.onPointerMove({
      pointerId: 1,
      clientX: 50,
      clientY: 45,
    });

    wrapper.vm.onPointerUp({
      pointerId: 1,
    });

    expect(wrapper.vm.pending).toEqual({
      x: 10,
      y: 20,
      w: 40,
      h: 40,
    });
  });

  it("clears active drafts when draw mode is disabled", async () => {
    wrapper.vm.pending = { x: 10, y: 20, w: 30, h: 30 };
    wrapper.vm.draft = { x: 10, y: 20, w: 30, h: 30 };
    wrapper.vm.interaction = "move";

    await wrapper.setProps({ mode: "display" });

    expect(wrapper.vm.pending).toBeNull();
    expect(wrapper.vm.draft).toBeNull();
    expect(wrapper.vm.interaction).toBeNull();
  });
});
