import { describe, it, expect, vi } from "vitest";
import { $api } from "../fixtures";
import clientConfig from "../config";
import { buildNamespace } from "common/storage";
import $notify from "common/notify";

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

  it("summarizes a failed request by its method, path and message", async () => {
    const { requestErrorSummary } = await import("common/api");
    const error = {
      message: "Request failed with status code 400",
      config: { method: "post", url: "albums?count=10", data: '{"Title":"Holiday 2026"}', headers: { "X-Client-Uri": "/static/build/app.js" } },
    };

    const summary = requestErrorSummary(error);
    expect(summary).toBe("api: POST albums › Request failed with status code 400");
    expect(summary).not.toContain("Holiday");
    expect(summary).not.toContain("count=10");
    expect(requestErrorSummary(undefined)).toBe("api: › Request failed");
  });

  it("logs the details of a failed request only in debug mode", async () => {
    const { Mock } = await import("../fixtures");
    const log = vi.spyOn(console, "log").mockImplementation(() => {});

    // logged returns everything the interceptor logged for a failed request, serialized.
    const logged = async () => {
      log.mockClear();
      Mock.onPost("log-test").replyOnce(400, { code: 400, error: "Bad request" });
      await expect($api.post("log-test", { Title: "Holiday 2026" }, { headers: { "X-Custom": "custom-value" } })).rejects.toBeDefined();
      return log.mock.calls.map((args) => (typeof args[0] === "string" ? args[0] : JSON.stringify(args[0]?.config ?? args[0])));
    };

    try {
      clientConfig.debug = false;
      const summary = await logged();
      expect(summary.join(" ")).not.toContain("Holiday");
      expect(summary.join(" ")).not.toContain("custom-value");
      expect(summary.some((l) => l.startsWith("api: POST log-test"))).toBe(true);

      clientConfig.debug = true;
      const details = await logged();
      expect(details.join(" ")).toContain("Holiday");
    } finally {
      clientConfig.debug = false;
      log.mockRestore();
    }
  });

  it("removes the request body and auth headers from failed requests outside debug mode", async () => {
    const { Mock } = await import("../fixtures");
    const log = vi.spyOn(console, "log").mockImplementation(() => {});

    // rejected returns the error of a failed request.
    const rejected = async () => {
      Mock.onPut("redact-test").replyOnce(400, { code: 400, error: "Bad request" });
      try {
        await $api.put("redact-test", { Title: "Holiday 2026" }, { headers: { "X-Auth-Token": "a1b2c3", "X-Custom": "kept" } });
      } catch (err) {
        return err;
      }
    };

    const defaultToken = $api.defaults.headers.common["X-Auth-Token"];
    $api.defaults.headers.common["X-Auth-Token"] = "d4e5f6";

    try {
      clientConfig.debug = false;
      const err = await rejected();
      expect($api.defaults.headers.common["X-Auth-Token"]).toBe("d4e5f6");
      expect(err.config.data).toBeUndefined();
      expect(err.config.headers["X-Auth-Token"]).toBeUndefined();
      expect(err.config.headers["X-Custom"]).toBe("kept");
      expect(err.config.url).toBe("redact-test");
      expect(JSON.stringify(err.response?.config ?? {})).not.toContain("Holiday");

      clientConfig.debug = true;
      const debugErr = await rejected();
      expect(debugErr.config.data).toContain("Holiday");
      expect(debugErr.config.headers["X-Auth-Token"]).toBe("a1b2c3");
    } finally {
      clientConfig.debug = false;
      log.mockRestore();

      if (defaultToken === undefined) {
        delete $api.defaults.headers.common["X-Auth-Token"];
      } else {
        $api.defaults.headers.common["X-Auth-Token"] = defaultToken;
      }
    }
  });

  // signsOut sends a request that the server answers with the specified status and body, and reports whether it signed out.
  const signsOut = async (status, body) => {
    const { Mock } = await import("../fixtures");
    const log = vi.spyOn(console, "log").mockImplementation(() => {});
    const logout = vi.spyOn($notify, "logout").mockImplementation(() => {});

    try {
      Mock.onPost("signout-test").replyOnce(status, body);
      await expect($api.post("signout-test", {})).rejects.toBeDefined();
      return logout.mock.calls.length > 0;
    } finally {
      logout.mockRestore();
      log.mockRestore();
    }
  };

  it("keeps the session after invalid credentials were entered", async () => {
    expect(await signsOut(401, { error: "invalid credentials", code: 44, message: "Invalid credentials" })).toBe(false);
  });

  it("signs out when the session is no longer valid", async () => {
    expect(await signsOut(401, { error: "Unauthorized", code: 401 })).toBe(true);
  });

  it("returns the error itself for logging only in debug mode", async () => {
    const { requestErrorDetails } = await import("common/api");
    const error = { message: "Network Error", config: { method: 1, url: "albums" } };

    try {
      expect(requestErrorDetails(error)).toBe("api: 1 albums › Network Error");
      const other = new Error("unexpected file data");
      expect(requestErrorDetails(other)).toBe(other);
      clientConfig.trace = true;
      expect(requestErrorDetails(error)).toBe(error);
    } finally {
      clientConfig.trace = false;
    }
  });

  it("should bootstrap auth header from namespaced storage", async () => {
    const baseUri = "/i/pro-1";
    const namespaceKey = "ns-pro-1";
    const token = "999900000000000000000000000000000000000000000000";
    const prevConfig = window.__CONFIG__;
    window.localStorage.clear();
    window.__CONFIG__ = { ...clientConfig, baseUri, storageNamespace: namespaceKey };
    window.localStorage.setItem(`${buildNamespace(namespaceKey)}session.token`, token);

    vi.resetModules();
    const { default: api } = await import("common/api");

    expect(api.defaults.headers.common["X-Auth-Token"]).toBe(token);

    window.__CONFIG__ = prevConfig;
    window.localStorage.clear();
  });
});
