import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { shallowMount } from "@vue/test-utils";
import "../../../fixtures";
import { Album } from "model/album";
import PAlbumEditDialog from "component/album/edit/dialog.vue";

const makeWrapper = () => {
  const album = new Album({
    UID: "alb1",
    Title: "Vacation",
    Location: "Berlin",
    Description: "",
    Category: "",
    Type: "album",
    Order: "newest",
    Favorite: false,
    Private: false,
  });

  const wrapper = shallowMount(PAlbumEditDialog, {
    props: { visible: true, album },
    global: {
      mocks: {
        $gettext: (s) => s,
        $pgettext: (_c, s) => s,
        $notify: { error: vi.fn(), success: vi.fn() },
        $view: { enter: vi.fn(), leave: vi.fn() },
        $config: {
          feature: () => false,
          get: () => false,
          ce: () => true,
          allow: () => true,
          albumCategories: () => [],
        },
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
        VTextarea: { template: "<textarea></textarea>" },
        VCombobox: { template: "<input />" },
        VSelect: { template: "<select></select>" },
        VCheckbox: { template: "<input type='checkbox' />" },
        VBtn: { template: "<button><slot /></button>" },
        VIcon: { template: "<i><slot /></i>" },
      },
    },
  });

  // Stub the watched model's update() so confirm() doesn't hit the API.
  const update = vi.fn().mockResolvedValue({});
  wrapper.vm.model.update = update;
  return { wrapper, update };
};

// vm.$.refs is the underlying refs container the proxy reads from; the
// public vm.$refs proxy ignores plain assignments.
const overrideFormRef = (vm, validate) => {
  vm.$.refs.form = { validate };
};

describe("component/album/edit/dialog", () => {
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
