import { describe, it, expect } from "vitest";
import "../fixtures";
import User from "model/user";
import File from "model/file";
import Config from "common/config";
import StorageShim from "node-storage-shim";
import { $session } from "app/session";

const defaultConfig = new Config(new StorageShim(), window.__CONFIG__);

describe("model/user", () => {
  it("should get handle", () => {
    const values = {
      ID: 5,
      Name: "max",
      DisplayName: "Max Last",
      Email: "test@test.com",
      Role: "admin",
    };

    const user = new User(values);
    const result = user.getHandle();
    expect(result).toBe("max");

    const values2 = {
      ID: 6,
      Name: "",
      DisplayName: "",
      Email: "test@test.com",
      Role: "admin",
    };

    const user2 = new User(values2);
    const result2 = user2.getHandle();
    expect(result2).toBe("");
  });

  it("should get default base path", () => {
    const values = {
      ID: 5,
      Name: "max",
      DisplayName: "Max Last",
      Email: "test@test.com",
      Role: "admin",
    };

    const user = new User(values);
    const result = user.defaultBasePath();
    expect(result).toBe("users/max");

    const values2 = {
      ID: 6,
      Name: "",
      DisplayName: "",
      Email: "test@test.com",
      Role: "admin",
    };

    const user2 = new User(values2);
    const result2 = user2.defaultBasePath();
    expect(result2).toBe("");
  });

  it("should get display name", () => {
    const values = {
      ID: 5,
      Name: "max",
      DisplayName: "Max Last",
      Email: "test@test.com",
      Role: "admin",
    };

    const user = new User(values);
    const result = user.getDisplayName();
    expect(result).toBe("Max Last");

    const values2 = {
      ID: 6,
      Name: "",
      DisplayName: "",
      Email: "test@test.com",
      Role: "admin",
    };

    const user2 = new User(values2);
    const result2 = user2.getDisplayName();
    expect(result2).toBe("Unknown");

    const values3 = {
      ID: 7,
      Name: "",
      DisplayName: "",
      Email: "test@test.com",
      Role: "admin",
      Details: {
        NickName: "maxi",
        GivenName: "Maximilian",
      },
    };

    const user3 = new User(values3);
    const result3 = user3.getDisplayName();
    expect(result3).toBe("maxi");

    const values4 = {
      ID: 8,
      Name: "",
      DisplayName: "",
      Email: "test@test.com",
      Role: "admin",
      Details: {
        NickName: "",
        GivenName: "Maximilian",
      },
    };

    const user4 = new User(values4);
    const result4 = user4.getDisplayName();
    expect(result4).toBe("Maximilian");
  });

  it("should get account info", () => {
    const values = {
      ID: 5,
      Name: "max",
      DisplayName: "Max Last",
      Email: "test@test.com",
      Role: "admin",
    };

    const user = new User(values);
    const result = user.getAccountInfo();
    expect(result).toBe("max");

    const values2 = {
      ID: 6,
      Name: "",
      DisplayName: "",
      Email: "test@test.com",
      Role: "admin",
    };

    const user2 = new User(values2);
    const result2 = user2.getAccountInfo();
    expect(result2).toBe("test@test.com");

    const values3 = {
      ID: 7,
      Name: "",
      DisplayName: "",
      Email: "",
      Role: "admin",
    };

    const user3 = new User(values3);
    const result3 = user3.getAccountInfo();
    expect(result3).toBe("Admin");

    const values4 = {
      ID: 8,
      Name: "",
      DisplayName: "",
      Email: "",
      Role: "",
    };

    const user4 = new User(values4);
    const result4 = user4.getAccountInfo();
    expect(result4).toBe("Account");

    const values5 = {
      ID: 9,
      Name: "",
      DisplayName: "",
      Email: "",
      Role: "admin",
      Details: {
        JobTitle: "Developer",
      },
    };

    const user5 = new User(values5);
    const result5 = user5.getAccountInfo();
    expect(result5).toBe("Developer");
  });

  it("should get entity name", () => {
    const values = {
      ID: 5,
      Name: "max",
      DisplayName: "Max Last",
      Email: "test@test.com",
      Role: "admin",
    };

    const user = new User(values);
    const result = user.getEntityName();
    expect(result).toBe("Max Last");
  });

  it("should manage scope helpers", () => {
    const unrestricted = new User({ Scope: "*" });
    expect(unrestricted.hasScope()).toBe(false);
    expect(unrestricted.getScope()).toBe("*");

    const restricted = new User({ Scope: "photos:view" });
    expect(restricted.hasScope()).toBe(true);
    expect(restricted.getScope()).toBe("photos:view");
  });

  it("should get id", () => {
    const values = {
      ID: 5,
      Name: "max",
      DisplayName: "Max Last",
      Email: "test@test.com",
      Role: "admin",
    };

    const user = new User(values);
    const result = user.getId();
    expect(result).toBe(5);
  });

  it("should get model name", () => {
    const result = User.getModelName();
    expect(result).toBe("User");
  });

  it("should get collection resource", () => {
    const result = User.getCollectionResource();
    expect(result).toBe("users");
  });

  it("should get register form", async () => {
    const values = { ID: 52, Name: "max", DisplayName: "Max Last" };
    const user = new User(values);
    const result = await user.getRegisterForm();
    expect(result.definition.foo).toBe("register");
  });

  it("should get avatar url", async () => {
    const values = { ID: 52, Name: "max", DisplayName: "Max Last" };
    const user = new User(values);
    const result = await user.getAvatarURL();
    expect(result).toBe("/static/img/avatar/tile_500.jpg");

    const values2 = {
      ID: 53,
      Name: "max",
      DisplayName: "Max Last",
      Thumb: "91e6c374afb78b28a52d7b4fd4fd2ea861b87123",
    };
    const user2 = new User(values2);
    const result2 = await user2.getAvatarURL("tile_500", defaultConfig);
    expect(result2).toBe("/api/v1/t/91e6c374afb78b28a52d7b4fd4fd2ea861b87123/public/tile_500");
  });

  it("should upload avatar", async () => {
    const values = { ID: 52, Name: "max", DisplayName: "Max Last" };
    const user = new User(values);

    const values2 = {
      InstanceID: 5,
      UID: "ABC123",
      Hash: "54ghtfd",
      FileType: "jpg",
      MediaType: "image",
      Name: "1/2/IMG123.jpg",
      CreatedAt: "2012-07-08T14:45:39Z",
      UpdatedAt: "2012-07-08T14:45:39Z",
    };
    const file = new File(values2);

    const Files = [file];

    const response = await user.uploadAvatar(Files);
    expect(response.Thumb).toBe("abc");
    expect(response.ThumbSrc).toBe("manual");
  });

  it("should get profile form", async () => {
    const values = { ID: 53, Name: "max", DisplayName: "Max Last" };
    const user = new User(values);
    const result = await user.getProfileForm();
    expect(result.definition.foo).toBe("profile");
  });

  it("should return whether user is remote", async () => {
    const values = { ID: 52, Name: "max", DisplayName: "Max Last", AuthProvider: "local" };
    const user = new User(values);
    const result = await user.isRemote();
    expect(result).toBe(false);

    const values2 = { ID: 51, Name: "max", DisplayName: "Max Last", AuthProvider: "ldap" };
    const user2 = new User(values2);
    const result2 = await user2.isRemote();
    expect(result2).toBe(true);
  });

  it("should return auth info", async () => {
    const values = { ID: 50, Name: "max", DisplayName: "Max Last", AuthProvider: "oidc" };
    const user = new User(values);
    const result = await user.authInfo();
    expect(result).toBe("OIDC");

    const values2 = { ID: 52, Name: "max", DisplayName: "Max Last", AuthProvider: "oidc", AuthMethod: "session" };
    const user2 = new User(values2);
    const result2 = await user2.authInfo();
    expect(result2).toBe("OIDC (Session)");
  });

  it("should get change password", async () => {
    const values = {
      ID: 54,
      Name: "max",
      DisplayName: "Max Last",
      Email: "test@test.com",
      Role: "admin",
    };

    const user = new User(values);
    const result = await user.changePassword("old", "new");
    expect(result.new_password).toBe("new");
  });

  // A10 contract: isRemote / hasWebDAV must always return a Boolean, so a
  // `:disabled` binding to these methods never passes undefined / "" to a
  // Vuetify Boolean prop. See specs/frontend/best-practices.md#a10.
  describe("isRemote / hasWebDAV Boolean contract", () => {
    it("isRemote returns Boolean false when AuthProvider is missing", () => {
      const user = new User({ ID: 1, Name: "max" });
      const result = user.isRemote();
      expect(typeof result).toBe("boolean");
      expect(result).toBe(false);
    });
    it("isRemote returns Boolean false when AuthProvider is empty string", () => {
      const user = new User({ ID: 1, Name: "max", AuthProvider: "" });
      expect(typeof user.isRemote()).toBe("boolean");
      expect(user.isRemote()).toBe(false);
    });
    it("hasWebDAV returns Boolean false when WebDAV is missing", () => {
      const user = new User({ ID: 1, Name: "max", Role: "admin" });
      const result = user.hasWebDAV();
      expect(typeof result).toBe("boolean");
      expect(result).toBe(false);
    });
    it("hasWebDAV returns Boolean false when WebDAV is 0", () => {
      const user = new User({ ID: 1, Name: "max", Role: "admin", WebDAV: 0 });
      const result = user.hasWebDAV();
      expect(typeof result).toBe("boolean");
      expect(result).toBe(false);
    });
    it("hasWebDAV returns Boolean true when WebDAV is true and role permits", () => {
      const user = new User({ ID: 1, Name: "max", Role: "admin", WebDAV: true });
      const result = user.hasWebDAV();
      expect(typeof result).toBe("boolean");
      expect(result).toBe(true);
    });
  });

  // canHavePassword gates the admin "change password" action: visitors, system
  // users, and accounts with authentication disabled cannot have a local password,
  // mirroring the backend, which only allows passwords for registered accounts.
  describe("canHavePassword", () => {
    it("returns true for a registered local user", () => {
      expect(new User({ ID: 1, Name: "max", Role: "user", AuthProvider: "local" }).canHavePassword()).toBe(true);
    });
    it("returns true for a registered user without an explicit provider", () => {
      expect(new User({ ID: 1, Name: "max", Role: "admin" }).canHavePassword()).toBe(true);
    });
    it("returns false for a visitor, even when named guest", () => {
      expect(new User({ ID: 1, Name: "guest", Role: "visitor", AuthProvider: "link" }).canHavePassword()).toBe(false);
    });
    it("returns false for an account without a role", () => {
      expect(new User({ ID: 1, Name: "max", Role: "", AuthProvider: "local" }).canHavePassword()).toBe(false);
    });
    it("returns false for a remote LDAP account (credentials managed by the directory)", () => {
      expect(new User({ ID: 1, Name: "max", Role: "user", AuthProvider: "ldap" }).canHavePassword()).toBe(false);
    });
    it("returns true for an OIDC account (a local password remains usable)", () => {
      expect(new User({ ID: 1, Name: "max", Role: "user", AuthProvider: "oidc" }).canHavePassword()).toBe(true);
    });
    it("returns false when authentication is disabled (provider none)", () => {
      expect(new User({ ID: 1, Name: "max", Role: "user", AuthProvider: "none" }).canHavePassword()).toBe(false);
    });
    it("returns false for a system user with ID below 1", () => {
      expect(new User({ ID: 0, Name: "max", Role: "user" }).canHavePassword()).toBe(false);
    });
    it("returns false without a username", () => {
      expect(new User({ ID: 1, Name: "", Role: "user" }).canHavePassword()).toBe(false);
    });
  });

  // isCurrentUser drives the admin-UI self-lockout guards: the table login
  // toggle and the dialog role/auth/login fields lock for the signed-in user so
  // an operator cannot lock themselves out. See specs/portal/cluster-admin-ui.md.
  describe("isCurrentUser", () => {
    it("returns true for the signed-in user and false for others", () => {
      $session.setUser({ ID: 5, UID: "us1234567890self", Name: "max", Role: "admin" });

      const me = new User({ ID: 5, UID: "us1234567890self", Name: "max", Role: "admin" });
      expect(me.isCurrentUser()).toBe(true);

      const other = new User({ ID: 6, UID: "us1234567890othr", Name: "alice", Role: "user" });
      expect(other.isCurrentUser()).toBe(false);
    });
    it("returns Boolean false when the account has no UID", () => {
      $session.setUser({ ID: 5, UID: "us1234567890self", Name: "max", Role: "admin" });

      const blank = new User({ ID: 0, Name: "" });
      expect(typeof blank.isCurrentUser()).toBe("boolean");
      expect(blank.isCurrentUser()).toBe(false);
    });
  });

  // isAdmin / isClusterAdmin replace hardcoded role-string checks in the admin
  // dialogs (e.g. the Super Admin toggle); isAdmin mirrors the backend admin-tier
  // set {admin, cluster_admin}, isClusterAdmin matches only the Portal operator role.
  describe("isAdmin / isClusterAdmin", () => {
    it("isAdmin is true for admin and cluster_admin, false otherwise", () => {
      expect(new User({ ID: 1, Name: "a", Role: "admin" }).isAdmin()).toBe(true);
      expect(new User({ ID: 2, Name: "b", Role: "cluster_admin" }).isAdmin()).toBe(true);
      expect(new User({ ID: 3, Name: "c", Role: "user" }).isAdmin()).toBe(false);
      expect(new User({ ID: 4, Name: "d", Role: "" }).isAdmin()).toBe(false);
    });
    it("isClusterAdmin is true only for cluster_admin", () => {
      expect(new User({ ID: 2, Name: "b", Role: "cluster_admin" }).isClusterAdmin()).toBe(true);
      expect(new User({ ID: 1, Name: "a", Role: "admin" }).isClusterAdmin()).toBe(false);
    });
  });
});
