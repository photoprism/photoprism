import "../fixtures";
import User from "model/user";

let chai = require("chai/chai");
let assert = chai.assert;

describe("model/user", () => {
  it("should get entity name", () => {
    const values = { ID: 5, FullName: "Max Last", PrimaryEmail: "test@test.com", RoleAdmin: true };
    const user = new User(values);
    const result = user.getEntityName();
    assert.equal(result, "Max Last");
  });

  it("should get id", () => {
    const values = { ID: 5, FullName: "Max Last", PrimaryEmail: "test@test.com", RoleAdmin: true };
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
    const values = { ID: 52, FullName: "Max Last" };
    const user = new User(values);
    const result = await user.getRegisterForm();
    assert.equal(result.definition.foo, "register");
  });

  it("should get profile form", async () => {
    const values = { ID: 53, FullName: "Max Last" };
    const user = new User(values);
    const result = await user.getProfileForm();
    assert.equal(result.definition.foo, "profile");
  });

  it("should get change password", async () => {
    const values = { ID: 54, FullName: "Max Last", PrimaryEmail: "test@test.com", RoleAdmin: true };
    const user = new User(values);
    const result = await user.changePassword("old", "new");
    assert.equal(result.new_password, "new");
  });

  it("should save profile", async () => {
    const values = { ID: 55, FullName: "Max Last", PrimaryEmail: "test@test.com", RoleAdmin: true };
    const user = new User(values);
    assert.equal(user.FullName, "Max Last");
    await user.saveProfile();
    assert.equal(user.FullName, "Max New");
  });
});
