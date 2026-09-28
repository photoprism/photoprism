import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { shallowMount } from "@vue/test-utils";
import "../../../fixtures";
import { Subject, BirthYearMin } from "model/subject";
import PPeopleEditDialog from "component/people/edit/dialog.vue";

const makeWrapper = (values = {}) => {
  const person = new Subject({ UID: "sbj1", Name: "Alice", Favorite: false, Hidden: false, ...values });

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
        VDateInput: {
          name: "VDateInput",
          props: ["modelValue", "label", "min", "max", "menu"],
          emits: ["update:modelValue", "update:menu"],
          // Opens its picker on Enter keydown like VDateInput, so the dialog's listener phase matters.
          template: "<input type='date' :data-label='label' @keydown.enter=\"$emit('update:menu', true)\" />",
        },
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
  it("offers the verified and private checkboxes beside favorite and hidden", () => {
    const wrapper = makeWrapper();

    const labels = wrapper.findAll("input[type='checkbox']").map((c) => c.attributes("data-label"));

    expect(labels).toEqual(["Favorite", "Verified", "Hidden", "Private"]);

    wrapper.unmount();
  });

  it("round-trips the private flag through the model", () => {
    // The one visibility flag with no other writer: SaveForm is the only path that sets it, so the
    // checkbox is the whole feature and a default of undefined would drop it from the payload.
    const wrapper = makeWrapper();

    expect(wrapper.vm.model.Private).toBe(false);

    wrapper.vm.model.Private = true;

    expect(Object.keys(wrapper.vm.model.getValues(false))).toContain("Private");
    expect(wrapper.vm.model.getValues(true)).toEqual({ Private: true });

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

describe("component/people/edit/dialog birthday", () => {
  // Entered here and nowhere else, so the picker has to be seeded from the person and write back in
  // the shape the API stores - a local Date on either side, UTC midnight in the model.
  it("seeds the picker from the stored date", async () => {
    const wrapper = makeWrapper({ Birthday: "1990-08-01T00:00:00Z" });

    // The person is cloned when the dialog opens, so the model is only seeded across that edge.
    await wrapper.setProps({ visible: false });
    await wrapper.setProps({ visible: true });

    const picker = wrapper.findComponent({ name: "VDateInput" });

    expect(picker.props("label")).toBe("Birth Date");
    expect(picker.props("modelValue")).toEqual(new Date(1990, 7, 1));

    wrapper.unmount();
  });

  it("bounds the picker to the range the API accepts", () => {
    const wrapper = makeWrapper();

    const picker = wrapper.findComponent({ name: "VDateInput" });
    const max = picker.props("max");
    const min = picker.props("min");

    expect(max).toBeInstanceOf(Date);
    expect(max.getTime()).toBeLessThanOrEqual(Date.now());
    expect(min).toBeInstanceOf(Date);
    expect(min.getFullYear()).toBe(BirthYearMin);

    wrapper.unmount();
  });

  it("writes a picked date back to the model", async () => {
    const wrapper = makeWrapper();

    expect(wrapper.vm.model.Birthday).toBeNull();

    await wrapper.findComponent({ name: "VDateInput" }).vm.$emit("update:modelValue", new Date(1991, 8, 2));
    expect(wrapper.vm.model.Birthday).toBe("1991-09-02T00:00:00Z");

    await wrapper.findComponent({ name: "VDateInput" }).vm.$emit("update:modelValue", null);
    expect(wrapper.vm.model.Birthday).toBeNull();

    wrapper.unmount();
  });
});

describe("component/people/edit/dialog enter key", () => {
  // Enter confirms the dialog on keyup, while the date input and its picker act on keydown, so one
  // press must not do both.
  const withConfirm = () => {
    const wrapper = makeWrapper();
    const confirm = vi.spyOn(wrapper.vm, "confirm").mockResolvedValue(undefined);
    return { wrapper, confirm };
  };

  const press = (el, type, init = {}) => el.dispatchEvent(new KeyboardEvent(type, { key: "Enter", bubbles: true, ...init }));

  const enter = (el, init = {}) => {
    press(el, "keydown", init);
    press(el, "keyup", init);
  };

  const nameInput = (wrapper) => wrapper.find("input:not([type])").element;

  const dateInput = (wrapper) => wrapper.find("input[type='date']").element;

  it("confirms on Enter in the name field", () => {
    const { wrapper, confirm } = withConfirm();

    enter(nameInput(wrapper));

    expect(confirm).toHaveBeenCalledTimes(1);

    wrapper.unmount();
  });

  it("confirms on Enter in the date field while the picker is closed", () => {
    const { wrapper, confirm } = withConfirm();

    enter(dateInput(wrapper));

    expect(confirm).toHaveBeenCalledTimes(1);

    wrapper.unmount();
  });

  it("closes only the picker on Enter while it is open", async () => {
    const { wrapper, confirm } = withConfirm();
    const picker = wrapper.findComponent({ name: "VDateInput" });

    await picker.vm.$emit("update:menu", true);
    enter(dateInput(wrapper));
    await wrapper.vm.$nextTick();

    expect(confirm).not.toHaveBeenCalled();
    expect(picker.props("menu")).toBe(false);

    wrapper.unmount();
  });

  it("ignores a keyup whose keydown happened outside the dialog", () => {
    const { wrapper, confirm } = withConfirm();

    press(dateInput(wrapper), "keyup");

    expect(confirm).not.toHaveBeenCalled();

    wrapper.unmount();
  });

  it("ignores a modified Enter without leaving it pending", () => {
    const { wrapper, confirm } = withConfirm();

    enter(nameInput(wrapper), { shiftKey: true });
    press(nameInput(wrapper), "keydown", { shiftKey: true });
    press(nameInput(wrapper), "keyup");
    press(nameInput(wrapper), "keydown");
    press(nameInput(wrapper), "keyup", { ctrlKey: true });
    press(nameInput(wrapper), "keyup");

    expect(confirm).not.toHaveBeenCalled();

    wrapper.unmount();
  });

  it("forgets an unmatched keydown when the dialog opens again", async () => {
    const { wrapper, confirm } = withConfirm();

    press(nameInput(wrapper), "keydown");
    await wrapper.setProps({ visible: false });
    await wrapper.setProps({ visible: true });
    press(nameInput(wrapper), "keyup");

    expect(confirm).not.toHaveBeenCalled();

    wrapper.unmount();
  });
});
