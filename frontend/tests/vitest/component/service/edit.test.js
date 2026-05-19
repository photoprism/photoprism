import { describe, it, expect, vi } from "vitest";
import "../../fixtures";
import { shallowMount } from "@vue/test-utils";
import PServiceEdit from "component/service/edit.vue";

// Builds a stand-in for a Service model. The dialog mutates fields via
// v-model bindings (template) and via this.model.X = … (script) on a
// caller-supplied clone, then calls model.update() in save(). Tests
// verify the prop-rename + computed-alias contract: writes through
// the `model` computed land on the underlying `service` prop, and
// reassigning the prop reactively re-targets the alias.
function buildService(overrides = {}) {
  return {
    UID: "svc-1",
    AccName: "Original Account",
    AccType: "webdav",
    AccShare: false,
    AccSync: false,
    SharePath: "/",
    ShareSize: "fit_2048",
    ShareExpires: 0,
    SyncPath: "/",
    SyncStatus: "",
    update: vi.fn().mockResolvedValue(undefined),
    ...overrides,
  };
}

function mountEdit(props = {}) {
  return shallowMount(PServiceEdit, {
    props: {
      visible: false,
      scope: "",
      service: buildService(),
      ...props,
    },
    global: {
      mocks: {
        $config: {
          get: () => false,
          values: {},
        },
        $view: { enter: vi.fn(), leave: vi.fn() },
        $notify: { success: vi.fn(), error: vi.fn(), busy: vi.fn() },
      },
    },
  });
}

describe("component/service/edit", () => {
  describe("model computed alias", () => {
    it("aliases the service prop so v-model bindings write through to the same object", () => {
      const service = buildService();
      const wrapper = mountEdit({ service });

      // Vue 3 wraps prop objects in a reactive Proxy, so wrapper.vm.model
      // is not strictly === the literal `service` reference. The contract
      // we care about is "writes through this.model land on the underlying
      // service object" — assert it via mutation flow rather than identity.
      wrapper.vm.model.AccShare = true;
      wrapper.vm.model.SharePath = "/uploads";
      expect(service.AccShare).toBe(true);
      expect(service.SharePath).toBe("/uploads");
    });

    it("reactively re-targets when the parent reassigns the service prop", async () => {
      const first = buildService({ UID: "svc-1", AccName: "First" });
      const wrapper = mountEdit({ service: first });
      expect(wrapper.vm.model.AccName).toBe("First");

      // services.vue does `this.model = service.clone()` per dialog open;
      // the dialog stays mounted, so it must pick up the new prop value.
      const second = buildService({ UID: "svc-2", AccName: "Second" });
      await wrapper.setProps({ service: second });

      expect(wrapper.vm.model.UID).toBe("svc-2");
      expect(wrapper.vm.model.AccName).toBe("Second");

      // Writes after the swap must now flow to `second`, not `first`.
      wrapper.vm.model.AccShare = true;
      expect(second.AccShare).toBe(true);
      expect(first.AccShare).toBe(false);
    });
  });

  describe("save / confirm / disable / enable", () => {
    it("save() calls service.update() and clears loading on success", async () => {
      const service = buildService();
      const wrapper = mountEdit({ service });

      await wrapper.vm.save();

      expect(service.update).toHaveBeenCalledTimes(1);
      expect(wrapper.vm.loading).toBe(false);
      expect(wrapper.vm.$notify.success).toHaveBeenCalledTimes(1);
    });

    it("save() short-circuits with $notify.busy when already loading", async () => {
      const service = buildService();
      const wrapper = mountEdit({ service });

      wrapper.vm.loading = true;
      wrapper.vm.save();

      expect(service.update).not.toHaveBeenCalled();
      expect(wrapper.vm.$notify.busy).toHaveBeenCalledTimes(1);
    });

    it("confirm() flips AccShare on the underlying service and calls update", async () => {
      const service = buildService({ AccShare: false });
      const wrapper = mountEdit({ service });

      await wrapper.vm.confirm();

      expect(service.AccShare).toBe(true);
      expect(service.update).toHaveBeenCalledTimes(1);
    });

    it("disable(prop) sets the property to false and saves", async () => {
      const service = buildService({ AccShare: true });
      const wrapper = mountEdit({ service });

      await wrapper.vm.disable("AccShare");

      expect(service.AccShare).toBe(false);
      expect(service.update).toHaveBeenCalledTimes(1);
    });

    it("enable(prop) sets the property to true without auto-saving", () => {
      const service = buildService({ AccShare: false });
      const wrapper = mountEdit({ service });

      wrapper.vm.enable("AccShare");

      expect(service.AccShare).toBe(true);
      expect(service.update).not.toHaveBeenCalled();
    });
  });

  describe("emits declaration", () => {
    // Pin the emits surface so removing a declared event is a visible diff.
    // The actual emit-call assertions live in the parent's integration tests
    // (settings/services); this codebase's vitest setup does not capture
    // component-level $emit() calls in wrapper.emitted() reliably.
    it("declares close, remove, and confirm as the dialog's outbound events", () => {
      const declared = PServiceEdit.emits || [];
      expect(declared).toEqual(expect.arrayContaining(["close", "remove", "confirm"]));
    });
  });
});
