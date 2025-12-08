import { describe, it, expect } from "vitest";
import "../fixtures";
import User from "model/user";
import File from "model/file";
import Config from "common/config";
import StorageShim from "node-storage-shim";

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
});
