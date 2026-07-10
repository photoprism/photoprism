import { describe, expect, it } from "vitest";
import MockAdapter from "axios-mock-adapter";
import $api from "common/api";
import { Photo } from "model/photo";

// Echo the request body back so update() resolves with the values it sent.
const Mock = new MockAdapter($api);
Mock.onPut(/photos\//).reply((config) => [200, JSON.parse(config.data)]);

describe("model/photo rating", () => {
  it("declares unrated defaults", () => {
    const photo = new Photo();
    expect(photo.Rating).toBe(0);
    expect(photo.RatingSrc).toBe("");
  });

  it("keeps assigned rating values", () => {
    const photo = new Photo({ ID: 2, UID: "pqbcf5j446s0futz", Rating: 4, RatingSrc: "meta" });
    expect(photo.Rating).toBe(4);
    expect(photo.RatingSrc).toBe("meta");
  });

  it("round-trips rating through getValues", () => {
    const photo = new Photo({ ID: 3, UID: "pqbcf5j446s0fut0", Rating: 5, RatingSrc: "manual" });
    const values = photo.getValues(false);
    expect(values.Rating).toBe(5);
    expect(values.RatingSrc).toBe("manual");
  });

  it("stamps the manual source when stars are picked", async () => {
    const photo = new Photo({ ID: 4, UID: "pqbcf5j446s0fut1", Rating: 0, RatingSrc: "" });
    photo.Rating = 4;
    await photo.update();
    const sent = JSON.parse(Mock.history.put.at(-1).data);
    expect(sent.Rating).toBe(4);
    expect(sent.RatingSrc).toBe("manual");
  });

  it("clearing the stars reverts to never rated", async () => {
    const photo = new Photo({ ID: 5, UID: "pqbcf5j446s0fut2", Rating: 3, RatingSrc: "manual" });
    photo.Rating = 0;
    await photo.update();
    const sent = JSON.parse(Mock.history.put.at(-1).data);
    expect(sent.Rating).toBe(0);
    expect(sent.RatingSrc).toBe("");
  });

  it("treats a null rating from the input as a cleared rating", async () => {
    const photo = new Photo({ ID: 6, UID: "pqbcf5j446s0fut3", Rating: 2, RatingSrc: "manual" });
    photo.Rating = null;
    await photo.update();
    const sent = JSON.parse(Mock.history.put.at(-1).data);
    expect(sent.Rating).toBe(0);
    expect(sent.RatingSrc).toBe("");
  });
});
