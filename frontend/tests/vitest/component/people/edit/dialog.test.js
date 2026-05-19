import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { shallowMount } from "@vue/test-utils";
import "../../../fixtures";
import { Subject } from "model/subject";
import PPeopleEditDialog from "component/people/edit/dialog.vue";

const makeWrapper = () => {
  const person = new Subject({ UID: "sbj1", Name: "Alice", Favorite: false, Hidden: false });

  const wrapper = shallowMount(PPeopleEditDialog, {
    props: { visible: true, person },
    global: {
      mocks: {
        $gettext: (s) => s,
        $notify: { error: vi.fn(), success: vi.fn() },
        $view: { enter: vi.fn(), leave: vi.fn() },
        $config: { allow: () => true },
      },
      stubs: {
        VDialog: { template: "<div><slot /></div>" },
        VForm: { template: "<form><slot /></form>" },
        VCard: { template: "<div><slot /></div>" },
        VCardText: { template: "<div><slot /></div>" },
        VCardActions: { template: "<div><slot /></div>" },
        VToolbar: { template: "<div><slot /></div>" },
        VToolbarTitle: { template: "<div><slot /></div>" },
        VRow: { template: "<div><slot /></div>" },
        VCol: { template: "<div><slot /></div>" },
        VTextField: { template: "<input />" },
        VCheckbox: { template: "<input type='checkbox' />" },
        VBtn: { template: "<button><slot /></button>" },
        VIcon: { template: "<i><slot /></i>" },
      },
    },
  });

  return wrapper;
};

const overrideFormRef = (vm, validate) => {
  vm.$.refs.form = { validate };
};

describe("component/people/edit/dialog", () => {
  let wrapper;

  beforeEach(() => {
    wrapper = makeWrapper();
  });

  afterEach(() => {
    if (wrapper) wrapper.unmount();
  });

  it("blocks confirm and notifies when form validation fails", async () => {
    const validate = vi.fn().mockResolvedValue({ valid: false });
    overrideFormRef(wrapper.vm, validate);

    await wrapper.vm.confirm();

    expect(validate).toHaveBeenCalled();
    expect(wrapper.emitted("confirm")).toBeFalsy();
    expect(wrapper.vm.$notify.error).toHaveBeenCalledWith("Changes could not be saved");
  });

  it("emits confirm with the model when form validation passes", async () => {
    const validate = vi.fn().mockResolvedValue({ valid: true });
    overrideFormRef(wrapper.vm, validate);

    await wrapper.vm.confirm();

    expect(validate).toHaveBeenCalled();
    expect(wrapper.emitted("confirm")).toBeTruthy();
    expect(wrapper.emitted("confirm")[0][0]).toBe(wrapper.vm.model);
  });
});
