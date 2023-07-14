import "../fixtures";
import { config } from "app/session";
import Session from "common/session";
import StorageShim from "node-storage-shim";

let chai = require("chai/chai");
let assert = chai.assert;

describe("common/session", () => {
  beforeEach(() => {
    window.onbeforeunload = () => "Oh no!";
  });

  it("should construct session", () => {
    const storage = new StorageShim();
    const session = new Session(storage, config);
    assert.equal(session.session_id, null);
  });

  it("should set, get and delete token", () => {
    const storage = new StorageShim();
    const session = new Session(storage, config);
    assert.equal(session.hasToken("2lbh9x09"), false);
    session.setId("999900000000000000000000000000000000000000000000");
    assert.equal(session.session_id, "999900000000000000000000000000000000000000000000");
    const result = session.getId();
    assert.equal(result, "999900000000000000000000000000000000000000000000");
    session.reset();
    assert.equal(session.session_id, null);
  });

  it("should set, get and delete user", () => {
    const storage = new StorageShim();
    const session = new Session(storage, config);
    assert.isFalse(session.user.hasId());

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

    session.setData();
    assert.equal(session.user.DisplayName, "");
    session.setData(data);
    assert.equal(session.user.DisplayName, "Max Example");
    assert.equal(session.user.SuperAdmin, true);
    assert.equal(session.user.Role, "admin");
    session.reset();
    assert.equal(session.user.DisplayName, "");
    assert.equal(session.user.SuperAdmin, false);
    assert.equal(session.user.Role, "");
    session.setUser(user);
    assert.equal(session.user.DisplayName, "Max Example");
    assert.equal(session.user.SuperAdmin, true);
    assert.equal(session.user.Role, "admin");

    const result = session.getUser();

    assert.equal(result.DisplayName, "Max Example");
    assert.equal(result.SuperAdmin, true);
    assert.equal(result.Role, "admin");
    assert.equal(result.Email, "test@test.com");
    assert.equal(result.ID, 5);
    session.deleteData();
    assert.isTrue(session.user.hasId());
    session.deleteUser();
    assert.isFalse(session.user.hasId());
  });

  it("should get user email", () => {
    const storage = new StorageShim();
    const session = new Session(storage, config);
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
    assert.equal(result, "test@test.com");
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
    assert.equal(result2, "");
    session.deleteData();
  });

  it("should get user display name", () => {
    const storage = new StorageShim();
    const session = new Session(storage, config);
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
    assert.equal(result, "Max Last");
    const values2 = {
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
    assert.equal(result2, "Bar");
    session.deleteData();
  });

  it("should get user full name", () => {
    const storage = new StorageShim();
    const session = new Session(storage, config);
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
    assert.equal(result, "Max Last");
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
    assert.equal(result2, "");
    session.deleteData();
  });

  it("should test whether user is set", () => {
    const storage = new StorageShim();
    const session = new Session(storage, config);
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
    assert.equal(result, true);
    session.deleteData();
  });

  it("should test whether user is admin", () => {
    const storage = new StorageShim();
    const session = new Session(storage, config);
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
    assert.equal(result, true);
    session.deleteData();
  });

  it("should test whether user is anonymous", () => {
    const storage = new StorageShim();
    const session = new Session(storage, config);
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
    assert.equal(result, false);
    session.deleteData();
  });

  it("should use session storage", () => {
    const storage = new StorageShim();
    const session = new Session(storage, config);
    assert.equal(storage.getItem("session_storage"), null);
    session.useSessionStorage();
    assert.equal(storage.getItem("session_storage"), "true");
    session.deleteData();
  });

  it("should use local storage", () => {
    const storage = new StorageShim();
    const session = new Session(storage, config);
    assert.equal(storage.getItem("session_storage"), null);
    session.useLocalStorage();
    assert.equal(storage.getItem("session_storage"), "false");
    session.deleteData();
  });

  it("should test redeem token", async () => {
    const storage = new StorageShim();
    const session = new Session(storage, config);
    assert.equal(session.data, null);
    await session.redeemToken("token123");
    assert.equal(session.data.token, "123token");
    session.deleteData();
  });
});
