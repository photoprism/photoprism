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
        VCheckbox: { props: ["modelValue", "label"], template: "<input type='checkbox' :data-label='label' />" },
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

describe("component/people/edit/dialog verified flag", () => {
  // The flag decides whether a face reset keeps the person, so it has to be reachable where a
  // person is edited rather than only through the API.
  it("offers a verified checkbox between favorite and hidden", () => {
    const wrapper = makeWrapper();

    const labels = wrapper.findAll("input[type='checkbox']").map((c) => c.attributes("data-label"));

    expect(labels).toEqual(["Favorite", "Verified", "Hidden"]);

    wrapper.unmount();
  });

  it("round-trips the flag through the model", () => {
    const wrapper = makeWrapper();

    expect(wrapper.vm.model.Verified).toBe(false);

    wrapper.vm.model.Verified = true;

    expect(wrapper.vm.model.Verified).toBe(true);
    // Sent to the server on save, so a default of undefined would drop it from the payload.
    expect(Object.keys(wrapper.vm.model.getValues(false))).toContain("Verified");

    wrapper.unmount();
  });
});
