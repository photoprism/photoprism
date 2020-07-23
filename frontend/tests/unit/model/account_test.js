import Account from "model/account";
import Photo from "model/photo";
import MockAdapter from "axios-mock-adapter";
import Api from "common/api";

let chai = require("chai/chai");
let assert = chai.assert;

const mock = new MockAdapter(Api);

mock
    .onGet("api/v1/accounts/123/folders").reply(200, "folders success")
    .onPost("api/v1/accounts/123/share").reply(200, "share success");


describe("model/account", () => {

    it("should get account defaults",  () => {
        const values = {ID: 5};
        const account = new Account(values);
        const result = account.getDefaults();
        assert.equal(result.ID, 0);
        assert.equal(result.AccShare, true);
        assert.equal(result.AccName, "");
    });

    it("should get account entity name",  () => {
        const values = {ID: 5, AccName: "Test Name"};
        const account = new Account(values);
        const result = account.getEntityName();
        assert.equal(result, "Test Name");
    });

    it("should get account id",  () => {
        const values = {ID: 5, AccName: "Test Name"};
        const account = new Account(values);
        const result = account.getId();
        assert.equal(result, 5);
    });

    it("should get folders",  (done) => {
        const values = {ID: 123, AccName: "Test Name"};
        const account = new Account(values);
        account.Folders().then(
            (response) => {
                assert.equal(response, "folders success");
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });

    it("should get share photos",  (done) => {
        const values = {ID: 123, AccName: "Test Name"};
        const account = new Account(values);
        const values1 = {ID: 5, Title: "Crazy Cat", UID: 789};
        const photo = new Photo(values1);
        const values2 = {ID: 6, Title: "Crazy Cat 2", UID: 783};
        const photo2 = new Photo(values2);
        const Photos = [photo, photo2];
        account.Share(Photos, "destinaton").then(
            (response) => {
                assert.equal(response, "share success");
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });

    it("should get collection resource",  () => {
        const result = Account.getCollectionResource();
        assert.equal(result, "accounts");
    });

    it("should get model name",  () => {
        const result = Account.getModelName();
        assert.equal(result, "Account");
    });

});