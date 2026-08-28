// Covers the Enter handling of the shared confirmation dialog.
//
// A dialog raised by an Enter keypress receives that keypress's keyup once it mounts, so confirming
// on Enter unconditionally lets one keystroke both ask the question and answer it. Callers that open
// the dialog from a key handler opt out with confirm-on-enter.
import { describe, it, expect, vi } from "vitest";
import { mount } from "@vue/test-utils";

import PConfirmDialog from "component/confirm/dialog.vue";

const mountDialog = (props) =>
  mount(PConfirmDialog, {
    props: Object.assign({ visible: true }, props),
    global: {
      mocks: {
        $gettext: (msg) => msg,
        $view: { enter: vi.fn(), leave: vi.fn() },
      },
    },
  });

describe("PConfirmDialog", () => {
  it("confirms on Enter by default", () => {
    const wrapper = mountDialog();

    wrapper.vm.onEnter();

    expect(wrapper.emitted("confirm")).toHaveLength(1);
  });

  it("ignores Enter when the caller opted out", () => {
    const wrapper = mountDialog({ confirmOnEnter: false });

    wrapper.vm.onEnter();

    expect(wrapper.emitted("confirm")).toBeUndefined();
  });

  it("still confirms on click when Enter is disabled", () => {
    const wrapper = mountDialog({ confirmOnEnter: false });

    wrapper.vm.confirm();

    expect(wrapper.emitted("confirm")).toHaveLength(1);
  });

  it("still closes on cancel when Enter is disabled", () => {
    const wrapper = mountDialog({ confirmOnEnter: false });

    wrapper.vm.close();

    expect(wrapper.emitted("close")).toHaveLength(1);
  });
});
