import Api from "common/api";
import MockAdapter from "axios-mock-adapter";

let chai = require("../../../node_modules/chai/chai");
let assert = chai.assert;

describe("common/api", () => {

    const mock = new MockAdapter(Api);

    const getCollectionResponse = [
        {id: 1, name: "John Smith"},
        {id: 1, name: "John Smith"}
    ];

    const getEntityResponse = {
        id: 1, name: "John Smith"
    };

    const postEntityResponse = {
        users: [
            {id: 1, name: "John Smith"}
        ]
    };

    const putEntityResponse = {
        users: [
            {id: 2, name: "John Foo"}
        ]
    };

    const deleteEntityResponse = null;

    mock.onGet("foo").reply(200, getCollectionResponse);
    mock.onGet("foo/123").reply(200, getEntityResponse);
    mock.onPost("foo").reply(201, postEntityResponse);
    mock.onPut("foo/2").reply(200, putEntityResponse);
    mock.onDelete("foo/2").reply(204, deleteEntityResponse);
    mock.onGet("error").reply(401, "custom error cat");

    it("get should return a list of results and return with HTTP code 200", (done) => {
        Api.get("foo").then(
            (response) => {
                assert.equal(200, response.status);
                assert.deepEqual(getCollectionResponse, response.data);
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });

    it("get should return one item and return with HTTP code 200", (done) => {
        Api.get("foo/123").then(
            (response) => {
                assert.equal(200, response.status);
                assert.deepEqual(getEntityResponse, response.data);
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });

    it("post should create one item and return with HTTP code 201", (done) => {
        Api.post("foo", postEntityResponse).then(
            (response) => {
                assert.equal(201, response.status);
                assert.deepEqual(postEntityResponse, response.data);
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });

    it("put should update one item and return with HTTP code 200", (done) => {
        Api.put("foo/2", putEntityResponse).then(
            (response) => {
                assert.equal(200, response.status);
                assert.deepEqual(putEntityResponse, response.data);
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });

    it("delete should delete one item and return with HTTP code 204", (done) => {
        Api.delete("foo/2", deleteEntityResponse).then(
            (response) => {
                assert.equal(204, response.status);
                assert.deepEqual(deleteEntityResponse, response.data);
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });

    it("get error", function() {
        return Api.get("error")
            .then(function(m) { throw new Error("was not supposed to succeed"); })
            .catch(function(m) { assert.equal(m.message, "Request failed with status code 401")});
    });
});
