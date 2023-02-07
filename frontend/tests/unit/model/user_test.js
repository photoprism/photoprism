import "../fixtures";
import User from "model/user";

let chai = require("chai/chai");
let assert = chai.assert;

describe("model/user", () => {
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
    assert.equal(result, "Max Last");
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
    assert.equal(result, 5);
  });

  it("should get model name", () => {
    const result = User.getModelName();
    assert.equal(result, "User");
  });

  it("should get collection resource", () => {
    const result = User.getCollectionResource();
    assert.equal(result, "users");
  });

  it("should get register form", async () => {
    const values = { ID: 52, Name: "max", DisplayName: "Max Last" };
    const user = new User(values);
    const result = await user.getRegisterForm();
    assert.equal(result.definition.foo, "register");
  });

  it("should get profile form", async () => {
    const values = { ID: 53, Name: "max", DisplayName: "Max Last" };
    const user = new User(values);
    const result = await user.getProfileForm();
    assert.equal(result.definition.foo, "profile");
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
    assert.equal(result.new_password, "new");
  });
});
