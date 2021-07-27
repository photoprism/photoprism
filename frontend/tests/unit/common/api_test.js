import { Api } from "../fixtures";

let chai = require("chai/chai");
let assert = chai.assert;

describe("common/api", () => {
  const getCollectionResponse = [
    { id: 1, name: "John Smith" },
    { id: 1, name: "John Smith" },
  ];

  const getEntityResponse = {
    id: 1,
    name: "John Smith",
  };

  const postEntityResponse = {
    users: [{ id: 1, name: "John Smith" }],
  };

  const putEntityResponse = {
    users: [{ id: 2, name: "John Foo" }],
  };

  const deleteEntityResponse = null;

  it("get should return a list of results and return with HTTP code 200", (done) => {
    Api.get("foo")
      .then((response) => {
        assert.equal(200, response.status);
        assert.deepEqual(getCollectionResponse, response.data);
        done();
      })
      .catch((error) => {
        done(error);
      });
  });

  it("get should return one item and return with HTTP code 200", (done) => {
    Api.get("foo/123")
      .then((response) => {
        assert.equal(200, response.status);
        assert.deepEqual(getEntityResponse, response.data);
        done();
      })
      .catch((error) => {
        done(error);
      });
  });

  it("post should create one item and return with HTTP code 201", (done) => {
    Api.post("foo", postEntityResponse)
      .then((response) => {
        assert.equal(201, response.status);
        assert.deepEqual(postEntityResponse, response.data);
        done();
      })
      .catch((error) => {
        done(error);
      });
  });

  it("put should update one item and return with HTTP code 200", (done) => {
    Api.put("foo/2", putEntityResponse)
      .then((response) => {
        assert.equal(200, response.status);
        assert.deepEqual(putEntityResponse, response.data);
        done();
      })
      .catch((error) => {
        done(error);
      });
  });

  it("delete should delete one item and return with HTTP code 204", (done) => {
    Api.delete("foo/2", deleteEntityResponse)
      .then((response) => {
        assert.equal(204, response.status);
        assert.deepEqual(deleteEntityResponse, response.data);
        done();
      })
      .catch((error) => {
        done(error);
      });
  });

  /*
  it("get error", function () {
    return Api.get("error")
      .then(function (m) {
        return Promise.reject("error expected");
      })
      .catch(function (m) {
        assert.equal(m.message, "Request failed with status code 401");
        return Promise.resolve();
      });
  });
   */
});
