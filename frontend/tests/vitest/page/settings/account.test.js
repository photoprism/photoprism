import { describe, it, expect } from "vitest";

import PSettingsAccount from "page/settings/account.vue";
import User from "model/user";

// canEditEmail evaluates the computed property for a session user with the specified permissions on accounts.
const canEditEmail = (user, perms) =>
  PSettingsAccount.computed.canEditEmail.call({
    $session: { getUser: () => user },
    $config: { allowAny: (resource, wanted) => resource === "users" && wanted.some((perm) => (perms || []).includes(perm)) },
  });

describe("page/settings/account", () => {
  describe("canEditEmail", () => {
    it("allows roles that manage accounts", () => {
      expect(canEditEmail(new User({ UID: "uqxetse3cy5eo9z2", Role: "admin" }), ["manage"])).toBe(true);
      expect(canEditEmail(new User({ UID: "uqxetse3cy5eo9z2", Role: "admin" }), ["manage_own"])).toBe(true);
    });
    it("allows super admins", () => {
      expect(canEditEmail(new User({ UID: "uqxetse3cy5eo9z2", Role: "user", SuperAdmin: true }), [])).toBe(true);
    });
    it("refuses other accounts", () => {
      expect(canEditEmail(new User({ UID: "uqxc08w3d0ej2283", Role: "admin" }), ["view", "update_own"])).toBe(false);
      expect(canEditEmail(new User({ UID: "uqxc08w3d0ej2283", Role: "user" }), [])).toBe(false);
    });
    it("refuses a missing user", () => {
      expect(canEditEmail(null, ["manage"])).toBe(false);
    });
  });
});
