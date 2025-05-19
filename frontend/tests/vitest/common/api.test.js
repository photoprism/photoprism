import { describe, it, expect } from "vitest";
import { $api } from "../fixtures";

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

  it("get should return a list of results and return with HTTP code 200", async () => {
    const response = await $api.get("foo");
    expect(response.status).toBe(200);
    expect(response.data).toEqual(getCollectionResponse);
  });

  it("get should return one item and return with HTTP code 200", async () => {
    const response = await $api.get("foo/123");
    expect(response.status).toBe(200);
    expect(response.data).toEqual(getEntityResponse);
  });

  it("post should create one item and return with HTTP code 201", async () => {
    const response = await $api.post("foo", postEntityResponse);
    expect(response.status).toBe(201);
    expect(response.data).toEqual(postEntityResponse);
  });

  it("put should update one item and return with HTTP code 200", async () => {
    const response = await $api.put("foo/2", putEntityResponse);
    expect(response.status).toBe(200);
    expect(response.data).toEqual(putEntityResponse);
  });

  it("delete should delete one item and return with HTTP code 204", async () => {
    const response = await $api.delete("foo/2", deleteEntityResponse);
    expect(response.status).toBe(204);
    expect(response.data).toEqual(deleteEntityResponse);
  });

  /*
  it("get error", async () => {
    // Assuming Api.get should be $api.get for this conversion context
    // If Api is a different client, this test would need adjustment based on that client.
    await expect($api.get("error")).rejects.toThrow("Request failed with status code 401");
  });
   */
});
