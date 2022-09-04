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
    assert.equal(session.session_id, null);
    session.setId(123421);
    assert.equal(session.session_id, 123421);
    const result = session.getId();
    assert.equal(result, 123421);
    session.deleteId();
    assert.equal(session.session_id, null);
  });

  it("should set, get and delete user", () => {
    const storage = new StorageShim();
    const session = new Session(storage, config);
    assert.isFalse(session.user.hasId());
    const values = {
      user: {
        ID: 5,
        NickName: "Foo",
        FullName: "Max Last",
        PrimaryEmail: "test@test.com",
        RoleAdmin: true,
      },
    };
    session.setData();
    assert.equal(session.user.FullName, "");
    session.setData(values);
    assert.equal(session.user.FullName, "Max Last");
    assert.equal(session.user.RoleAdmin, true);
    const result = session.getUser();
    assert.equal(result.ID, 5);
    assert.equal(result.PrimaryEmail, "test@test.com");
    session.deleteData();
    assert.isFalse(session.user.hasId());
  });

  it("should get user email", () => {
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
    const result = session.getEmail();
    assert.equal(result, "test@test.com");
    const values2 = {
      user: {
        DisplayName: "Foo",
        FullName: "Max Last",
        Email: "test@test.com",
        Role: "admin",
      },
    };
    session.setData(values2);
    const result2 = session.getEmail();
    assert.equal(result2, "");
    session.deleteData();
  });

  it("should get user nick name", () => {
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
    const result = session.getDisplayName();
    assert.equal(result, "Foo");
    const values2 = {
      user: {
        DisplayName: "Bar",
        FullName: "Max Last",
        Email: "test@test.com",
        Role: "admin",
      },
    };
    session.setData(values2);
    const result2 = session.getDisplayName();
    assert.equal(result2, "");
    session.deleteData();
  });

  it("should get user full name", () => {
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
    const result = session.getDisplayName();
    assert.equal(result, "Foo");
    const values2 = {
      user: {
        DisplayName: "Bar",
        FullName: "Max New",
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
        DisplayName: "Foo",
        FullName: "Max Last",
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
        DisplayName: "Foo",
        FullName: "Max Last",
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
