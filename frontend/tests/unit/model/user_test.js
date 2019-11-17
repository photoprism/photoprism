import User from "model/user";
import MockAdapter from "axios-mock-adapter";
import Api from "common/api";

let chai = require('../../../node_modules/chai/chai');
let assert = chai.assert;

describe("model/user", () => {
    const mock = new MockAdapter(Api);

    it("should get entity name",  () => {
        const values = {ID: 5, FirstName: "Max", LastName: "Last", Email: "test@test.com", Role: "admin"};
        const user = new User(values);
        const result = user.getEntityName();
        assert.equal(result, "Max Last");
    });

    it("should get id",  () => {
        const values = {ID: 5, FirstName: "Max", LastName: "Last", Email: "test@test.com", Role: "admin"};
        const user = new User(values);
        const result = user.getId();
        assert.equal(result, 5);
    });

    it("should get model name",  () => {
        const result = User.getModelName();
        assert.equal(result, "User");
    });

    it("should get collection resource",  () => {
        const result = User.getCollectionResource();
        assert.equal(result, "users");
    });

   it("should get register form",  async() => {
       mock.onAny("users/52/register").reply(200, "registerForm");
        const values = {ID: 52, FirstName: "Max"};
        const user = new User(values);
        const result = await user.getRegisterForm();
        assert.equal(result.definition, "registerForm");
       mock.reset();
   });

    it("should get profile form",  async() => {
        mock.onAny("users/53/profile").reply(200, "profileForm");
        const values = {ID: 53, FirstName: "Max"};
        const user = new User(values);
        const result = await user.getProfileForm();
        assert.equal(result.definition, "profileForm");
        mock.reset();
    });

    it("should get change password",  async() => {
        mock.onPut("users/54/password").reply(200,  {password: "old", new_password: "new"});
        const values = {ID: 54, FirstName: "Max", LastName: "Last", Email: "test@test.com", Role: "admin"};
        const user = new User(values);
        const result = await user.changePassword("old", "new");
        assert.equal(result.new_password, "new");
    });

    it("should save profile",  async() => {
        mock.onPost("users/55/profile").reply(200,  {FirstName: "MaxNew", LastName: "LastNew"});
        const values = {ID: 55, FirstName: "Max", LastName: "Last", Email: "test@test.com", Role: "admin"};
        const user = new User(values);
        assert.equal(user.FirstName, "Max");
        assert.equal(user.LastName, "Last");
        await user.saveProfile();
        assert.equal(user.FirstName, "MaxNew");
        assert.equal(user.LastName, "LastNew");
    });
});
