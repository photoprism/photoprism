import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { shallowMount } from "@vue/test-utils";
import "../../../fixtures";
import { Label } from "model/label";
import PLabelEditDialog from "component/label/edit/dialog.vue";

const makeWrapper = () => {
  const label = new Label({ UID: "lbl1", Name: "Cat", Favorite: false });

  const wrapper = shallowMount(PLabelEditDialog, {
    props: { visible: true, label },
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

  const update = vi.fn().mockResolvedValue({});
  wrapper.vm.model.update = update;
  return { wrapper, update };
};

const overrideFormRef = (vm, validate) => {
  vm.$.refs.form = { validate };
};

describe("component/label/edit/dialog", () => {
  let wrapper;
  let update;

  beforeEach(() => {
    ({ wrapper, update } = makeWrapper());
  });

  afterEach(() => {
    if (wrapper) wrapper.unmount();
  });

  it("blocks save and notifies when form validation fails", async () => {
    const validate = vi.fn().mockResolvedValue({ valid: false });
    overrideFormRef(wrapper.vm, validate);

    await wrapper.vm.confirm();

    expect(validate).toHaveBeenCalled();
    expect(update).not.toHaveBeenCalled();
    expect(wrapper.vm.$notify.error).toHaveBeenCalledWith("Changes could not be saved");
    expect(wrapper.emitted("close")).toBeFalsy();
  });

  it("proceeds with save when form validation passes", async () => {
    const validate = vi.fn().mockResolvedValue({ valid: true });
    overrideFormRef(wrapper.vm, validate);

    await wrapper.vm.confirm();

    expect(validate).toHaveBeenCalled();
    expect(update).toHaveBeenCalled();
    expect(wrapper.emitted("close")).toBeTruthy();
  });
});
