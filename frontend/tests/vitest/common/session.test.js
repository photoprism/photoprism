import { describe, it, expect, beforeEach } from "vitest";
import "../fixtures";
import { $config } from "app/session";
import Session from "common/session";
import StorageShim from "node-storage-shim";

describe("common/session", () => {
  beforeEach(() => {
    window.onbeforeunload = () => "Oh no!";
  });

  it("should construct session", () => {
    const storage = new StorageShim();
    const session = new Session(storage, $config);
    expect(session.authToken).toBe(null);
  });

  it("should set, get and delete token", () => {
    const storage = new StorageShim();
    const session = new Session(storage, $config);
    expect(session.hasToken("2lbh9x09")).toBe(false);
    session.setAuthToken("999900000000000000000000000000000000000000000000");
    expect(session.authToken).toBe("999900000000000000000000000000000000000000000000");
    const result = session.getAuthToken();
    expect(result).toBe("999900000000000000000000000000000000000000000000");
    session.reset();
    expect(session.authToken).toBe(null);
  });

  it("should set, get and delete user", () => {
    const storage = new StorageShim();
    const session = new Session(storage, $config);
    expect(session.user.hasId()).toBe(false);

    const user = {
      ID: 5,
      NickName: "Foo",
      GivenName: "Max",
      DisplayName: "Max Example",
      Email: "test@test.com",
      SuperAdmin: true,
      Role: "admin",
    };

    const data = {
      user,
    };

    expect(session.hasId()).toBe(false);
    expect(session.hasAuthToken()).toBe(false);
    expect(session.isAuthenticated()).toBe(false);
    expect(session.hasProvider()).toBe(false);
    session.setData();
    expect(session.user.DisplayName).toBe("");
    session.setData(data);
    expect(session.hasId()).toBe(false);
    expect(session.hasAuthToken()).toBe(false);
    expect(session.hasProvider()).toBe(false);
    session.setId("a9b8ff820bf40ab451910f8bbfe401b2432446693aa539538fbd2399560a722f");
    session.setAuthToken("234200000000000000000000000000000000000000000000");
    session.setProvider("public");
    expect(session.hasId()).toBe(true);
    expect(session.hasAuthToken()).toBe(true);
    expect(session.isAuthenticated()).toBe(true);
    expect(session.hasProvider()).toBe(true);
    expect(session.user.DisplayName).toBe("Max Example");
    expect(session.user.SuperAdmin).toBe(true);
    expect(session.user.Role).toBe("admin");
    session.reset();
    expect(session.user.DisplayName).toBe("");
    expect(session.user.SuperAdmin).toBe(false);
    expect(session.user.Role).toBe("");
    session.setUser(user);
    expect(session.user.DisplayName).toBe("Max Example");
    expect(session.user.SuperAdmin).toBe(true);
    expect(session.user.Role).toBe("admin");

    const result = session.getUser();

    expect(result.DisplayName).toBe("Max Example");
    expect(result.SuperAdmin).toBe(true);
    expect(result.Role).toBe("admin");
    expect(result.Email).toBe("test@test.com");
    expect(result.ID).toBe(5);
    session.deleteData();
    expect(session.user.hasId()).toBe(true);
    session.deleteUser();
    expect(session.user.hasId()).toBe(false);
  });

  it("should get user email", () => {
    const storage = new StorageShim();
    const session = new Session(storage, $config);

    session.setId("a9b8ff820bf40ab451910f8bbfe401b2432446693aa539538fbd2399560a722f");
    session.setAuthToken("234200000000000000000000000000000000000000000000");
    session.setProvider("public");

    const values = {
      user: {
        ID: 5,
        Name: "foo",
        DisplayName: "Max Last",
        Email: "test@test.com",
        Role: "admin",
      },
    };

    session.setData(values);
    const result = session.getEmail();
    expect(result).toBe("test@test.com");
    const values2 = {
      user: {
        Name: "foo",
        DisplayName: "Max Last",
        Email: "test@test.com",
        Role: "admin",
      },
    };
    session.setData(values2);
    const result2 = session.getEmail();
    expect(result2).toBe("");
    session.deleteData();
  });

  it("should get user display name", () => {
    const storage = new StorageShim();
    const session = new Session(storage, $config);
    const values = {
      user: {
        ID: 5,
        Name: "foo",
        DisplayName: "Max Last",
        Email: "test@test.com",
        Role: "admin",
      },
    };
    session.setData(values);
    const result = session.getDisplayName();
    expect(result).toBe("Max Last");
    const values2 = {
      id: "a9b8ff820bf40ab451910f8bbfe401b2432446693aa539538fbd2399560a722f",
      access_token: "234200000000000000000000000000000000000000000000",
      provider: "public",
      data: {},
      user: {
        ID: 5,
        Name: "bar",
        DisplayName: "",
        Email: "test@test.com",
        Role: "admin",
      },
    };
    session.setData(values2);
    const result2 = session.getDisplayName();
    expect(result2).toBe("Bar");
    session.deleteData();
  });

  it("should get user full name", () => {
    const storage = new StorageShim();
    const session = new Session(storage, $config);
    const values = {
      user: {
        ID: 5,
        Name: "foo",
        DisplayName: "Max Last",
        Email: "test@test.com",
        Role: "admin",
      },
    };
    session.setData(values);
    const result = session.getDisplayName();
    expect(result).toBe("Max Last");
    const values2 = {
      user: {
        Name: "bar",
        DisplayName: "Max New",
        Email: "test@test.com",
        Role: "admin",
      },
    };
    session.setData(values2);
    const result2 = session.getDisplayName();
    expect(result2).toBe("");
    session.deleteData();
  });

  it("should manage scope state", () => {
    const storage = new StorageShim();
    const session = new Session(storage, $config);

    // Default scope is unrestricted.
    expect(session.hasScope()).toBe(false);
    expect(session.getScope()).toBe("*");

    session.setId("a9b8ff820bf40ab451910f8bbfe401b2432446693aa539538fbd2399560a722f");
    session.setAuthToken("234200000000000000000000000000000000000000000000");
    session.setScope("photos:view");
    expect(session.hasScope()).toBe(true);
    expect(session.getScope()).toBe("photos:view");

    // Scope flag should survive re-instantiation with the same storage.
    const restoredSession = new Session(storage, $config);
    expect(restoredSession.hasScope()).toBe(true);
    expect(restoredSession.getScope()).toBe("photos:view");

    session.deleteAuthentication();
  });

  it("should test whether user is set", () => {
    const storage = new StorageShim();
    const session = new Session(storage, $config);
    const values = {
      user: {
        ID: 5,
        Name: "foo",
        DisplayName: "Max Last",
        Email: "test@test.com",
        Role: "admin",
      },
    };
    session.setData(values);
    const result = session.isUser();
    expect(result).toBe(true);
    session.deleteData();
  });

  it("should test whether user is admin", () => {
    const storage = new StorageShim();
    const session = new Session(storage, $config);
    const values = {
      user: {
        ID: 5,
        Name: "foo",
        DisplayName: "Max Last",
        Email: "test@test.com",
        Role: "admin",
      },
    };
    session.setData(values);
    const result = session.isAdmin();
    expect(result).toBe(true);
    session.deleteData();
  });

  it("should test whether user is anonymous", () => {
    const storage = new StorageShim();
    const session = new Session(storage, $config);
    const values = {
      user: {
        ID: 5,
        DisplayName: "Foo",
        FullName: "Max Last",
        Email: "test@test.com",
        Role: "admin",
      },
    };
    session.setData(values);
    const result = session.isAnonymous();
    expect(result).toBe(false);
    session.deleteData();
  });

  it("should use session storage", () => {
    const storage = new StorageShim();
    const session = new Session(storage, $config);
    expect(storage.getItem("session")).toBe(null);
    session.useSessionStorage();
    expect(storage.getItem("session")).toBe("true");
    session.deleteData();
  });

  it("should use local storage", () => {
    const storage = new StorageShim();
    const session = new Session(storage, $config);
    expect(storage.getItem("session")).toBe(null);
    session.useLocalStorage();
    expect(storage.getItem("session")).toBe("false");
    session.deleteData();
  });

  it("should test redeem token", async () => {
    const storage = new StorageShim();
    const session = new Session(storage, $config);
    expect(session.data).toBe(null);
    await session.redeemToken("token123");
    expect(session.data.token).toBe("123token");
    session.deleteData();
  });
});
