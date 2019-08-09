import assert from "assert";
import User from "model/user";

describe.only("model/user", () => {
    it("should get entity name",  () => {
        const values = {ID: 5, userFirstName: "Max", userLastName: "Last", userEmail: "test@test.com", userRole: "admin"};
        const user = new User(values);
        const result = user.getEntityName();
        assert.equal(result, "Max Last");
    });

    it("should get id",  () => {
        const values = {ID: 5, userFirstName: "Max", userLastName: "Last", userEmail: "test@test.com", userRole: "admin"};
        const user = new User(values);
        const result = user.getId();
        assert.equal(result, 5);
    });

});