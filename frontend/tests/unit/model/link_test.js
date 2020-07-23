import Link from "model/link";
import MockAdapter from "axios-mock-adapter";
import Api from "common/api";

let chai = require("chai/chai");
let assert = chai.assert;

const mock = new MockAdapter(Api);

mock
    .onGet("api/v1/link/5").reply(200, "get success")
    .onPut("api/v1/link/5").reply(200, "put success")
    .onDelete("api/v1/link/5").reply(200, "delete success");

describe("model/link", () => {

    it("should get link defaults",  () => {
        const values = {UID: 5};
        const link = new Link(values);
        const result = link.getDefaults();
        assert.equal(result.UID, 0);
        assert.equal(result.CanEdit, false);
        assert.equal(result.Share, "");
    });

    it("should get link url",  () => {
        const values = {UID: 5, Token: "1234hhtbbt", Slug: "friends", Share: "family"};
        const link = new Link(values);
        const result = link.url();
        assert.equal(result, "http://localhost:9876/s/1234hhtbbt/friends");
        const values2 = {UID: 5, Token: "", Share: "family"};
        const link2 = new Link(values2);
        const result2 = link2.url();
        assert.equal(result2, "http://localhost:9876/s/…/family");
    });

    it("should get link caption",  () => {
        const values = {UID: 5, Token: "AcfgbTTh", Slug: "friends", Share: "family"};
        const link = new Link(values);
        const result = link.caption();
        assert.equal(result, "/s/acfgbtth");
    });

    it("should get link id",  () => {
        const values = {UID: 5};
        const link = new Link(values);
        const result = link.getId();
        assert.equal(result, 5);
    });

    it("should test has id",  () => {
        const values = {UID: 5};
        const link = new Link(values);
        const result = link.hasId();
        assert.equal(result, true);
    });

    it("should get link slug",  () => {
        const values = {UID: 5, Token: "AcfgbTTh", Slug: "friends", Share: "family"};
        const link = new Link(values);
        const result = link.getSlug();
        assert.equal(result, "friends");
    });

    it("should test has slug",  () => {
        const values = {UID: 5, Token: "AcfgbTTh", Slug: "friends", Share: "family"};
        const link = new Link(values);
        const result = link.hasSlug();
        assert.equal(result, true);
        const values2 = {UID: 5, Token: "AcfgbTTh", Share: "family"};
        const link2 = new Link(values2);
        const result2 = link2.hasSlug();
        assert.equal(result2, false);
    });

    it("should clone link",  () => {
        const values = {UID: 5, Token: "AcfgbTTh", Slug: "friends", Share: "family"};
        const link = new Link(values);
        const result = link.clone();
        assert.equal(result.Slug, "friends");
        assert.equal(result.Token, "AcfgbTTh");
    });

    it("should test expire",  () => {
        const values = {UID: 5, Token: "AcfgbTTh", Slug: "friends", Share: "family", Expires: 80000, UpdatedAt: "2012-07-08T14:45:39Z"};
        const link = new Link(values);
        const result = link.expires();
        assert.equal(result, "7/9/2012");
    });

    it("should get collection resource",  () => {
        const result = Link.getCollectionResource();
        assert.equal(result, "links");
    });

    it("should get model name",  () => {
        const result = Link.getModelName();
        assert.equal(result, "Link");
    });

});
