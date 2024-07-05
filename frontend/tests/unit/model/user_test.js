import "../fixtures";
import User from "model/user";
import File from "model/file";

let chai = require("chai/chai");
let assert = chai.assert;

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
    assert.equal(result, "max");

    const values2 = {
      ID: 6,
      Name: "",
      DisplayName: "",
      Email: "test@test.com",
      Role: "admin",
    };

    const user2 = new User(values2);
    const result2 = user2.getHandle();
    assert.equal(result2, "");
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
    assert.equal(result, "users/max");

    const values2 = {
      ID: 6,
      Name: "",
      DisplayName: "",
      Email: "test@test.com",
      Role: "admin",
    };

    const user2 = new User(values2);
    const result2 = user2.defaultBasePath();
    assert.equal(result2, "");
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
    assert.equal(result, "Max Last");

    const values2 = {
      ID: 6,
      Name: "",
      DisplayName: "",
      Email: "test@test.com",
      Role: "admin",
    };

    const user2 = new User(values2);
    const result2 = user2.getDisplayName();
    assert.equal(result2, "Unknown");

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
    assert.equal(result3, "maxi");

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
    assert.equal(result4, "Maximilian");
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
    assert.equal(result, "max");

    const values2 = {
      ID: 6,
      Name: "",
      DisplayName: "",
      Email: "test@test.com",
      Role: "admin",
    };

    const user2 = new User(values2);
    const result2 = user2.getAccountInfo();
    assert.equal(result2, "test@test.com");

    const values3 = {
      ID: 7,
      Name: "",
      DisplayName: "",
      Email: "",
      Role: "admin",
    };

    const user3 = new User(values3);
    const result3 = user3.getAccountInfo();
    assert.equal(result3, "Admin");

    const values4 = {
      ID: 8,
      Name: "",
      DisplayName: "",
      Email: "",
      Role: "",
    };

    const user4 = new User(values4);
    const result4 = user4.getAccountInfo();
    assert.equal(result4, "Account");

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
    assert.equal(result5, "Developer");
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

  it("should get avatar url", async () => {
    const values = { ID: 52, Name: "max", DisplayName: "Max Last" };
    const user = new User(values);
    const result = await user.getAvatarURL();
    assert.equal(result, "/static/img/avatar/tile_500.jpg");

    const values2 = {
      ID: 53,
      Name: "max",
      DisplayName: "Max Last",
      Thumb: "91e6c374afb78b28a52d7b4fd4fd2ea861b87123",
    };
    const user2 = new User(values2);
    const result2 = await user2.getAvatarURL();
    assert.equal(result2, "/api/v1/t/91e6c374afb78b28a52d7b4fd4fd2ea861b87123/public/tile_500");
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

    return user
      .uploadAvatar(Files)
      .then((response) => {
        assert.equal("abc", response.Thumb);
        assert.equal("manual", response.ThumbSrc);

        return Promise.resolve();
      })
      .catch((error) => {
        return Promise.reject(error);
      });
  });

  it("should get profile form", async () => {
    const values = { ID: 53, Name: "max", DisplayName: "Max Last" };
    const user = new User(values);
    const result = await user.getProfileForm();
    assert.equal(result.definition.foo, "profile");
  });

  it("should return whether user is remote", async () => {
    const values = { ID: 52, Name: "max", DisplayName: "Max Last", AuthProvider: "local" };
    const user = new User(values);
    const result = await user.isRemote();
    assert.equal(result, false);

    const values2 = { ID: 51, Name: "max", DisplayName: "Max Last", AuthProvider: "ldap" };
    const user2 = new User(values2);
    const result2 = await user2.isRemote();
    assert.equal(result2, true);
  });

  it("should return auth info", async () => {
    const values = { ID: 50, Name: "max", DisplayName: "Max Last", AuthProvider: "oidc"};
    const user = new User(values);
    const result = await user.authInfo();
    assert.equal(result, "OIDC");

    const values2 = { ID: 52, Name: "max", DisplayName: "Max Last", AuthProvider: "oidc", AuthMethod: "session" };
    const user2 = new User(values2);
    const result2 = await user2.authInfo();
    assert.equal(result2, "OIDC (Session)");
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
