import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { shallowMount, flushPromises, config as VTUConfig } from "@vue/test-utils";
import { Mock } from "../../fixtures";
import PPageConsent, { redirectAllowed } from "page/auth/consent.vue";

const consentResponse = {
  client: { id: "cs5gfen1bgxz7s9i", name: "Claude", origin: "https://claude.ai" },
  account: { name: "admin", displayName: "Admin" },
  siteUrl: "https://photos.example.com/",
  scopes: [
    { name: "mcp", available: true },
    { name: "photos", available: true },
  ],
  available: true,
  expiresAt: "2026-10-18T12:00:00Z",
};

function mountConsent(request = "abc123") {
  const baseView = VTUConfig.global.mocks.$view || {};
  const baseConfig = VTUConfig.global.mocks.$config || {};
  const view = { ...baseView, enter: vi.fn(), leave: vi.fn(), redirect: vi.fn() };
  const notify = { error: vi.fn() };

  const wrapper = shallowMount(PPageConsent, {
    global: {
      mocks: {
        $route: { query: { request } },
        $view: view,
        $notify: notify,
        $config: { ...baseConfig, values: { ...(baseConfig.values || {}), wallpaperUri: "" } },
      },
      stubs: { PAuthHeader: true, PAuthFooter: true },
      renderStubDefaultSlot: true,
    },
  });

  return { wrapper, view, notify };
}

describe("page/auth/consent", () => {
  beforeEach(() => {
    Mock.resetHistory();
  });

  afterEach(() => {
    Mock.resetHandlers();
  });

  it("loads the request and preselects usable scopes", async () => {
    Mock.onPost("api/v1/oauth/consent", { request: "abc123" }).reply(200, consentResponse);
    const { wrapper } = mountConsent();
    await flushPromises();

    expect(wrapper.vm.request.client.name).toBe("Claude");
    expect(wrapper.vm.selected).toEqual(["mcp", "photos"]);
    expect(wrapper.vm.siteHost).toBe("photos.example.com");
    expect(wrapper.vm.canApprove).toBe(true);
  });

  it("approves the selected scopes and follows the redirect", async () => {
    Mock.onPost("api/v1/oauth/consent").reply(200, consentResponse);
    Mock.onPost("api/v1/oauth/consent/approve", { request: "abc123", scopes: ["mcp"] }).reply(200, {
      redirectUri: "https://claude.ai/api/mcp/auth_callback?code=xyz&iss=https%3A%2F%2Fphotos.example.com%2F",
    });
    const { wrapper, view } = mountConsent();
    await flushPromises();

    wrapper.vm.selected = ["mcp"];
    wrapper.vm.onApprove();
    await flushPromises();

    expect(Mock.history.post.length).toBe(2);
    expect(view.redirect).toHaveBeenCalledWith("https://claude.ai/api/mcp/auth_callback?code=xyz&iss=https%3A%2F%2Fphotos.example.com%2F", 0, false);
  });

  it("cannot approve without the mcp scope", async () => {
    Mock.onPost("api/v1/oauth/consent").reply(200, consentResponse);
    const { wrapper } = mountConsent();
    await flushPromises();

    wrapper.vm.selected = ["photos"];
    expect(wrapper.vm.canApprove).toBe(false);
    wrapper.vm.onApprove();
    await flushPromises();

    expect(Mock.history.post.length).toBe(1);
  });

  it("offers only the link back when the account cannot use the app", async () => {
    Mock.onPost("api/v1/oauth/consent").reply(200, {
      ...consentResponse,
      available: false,
      scopes: [{ name: "mcp", available: false }],
    });
    Mock.onPost("api/v1/oauth/consent/deny", { request: "abc123" }).reply(200, {
      redirectUri: "https://claude.ai/api/mcp/auth_callback?error=access_denied",
    });
    const { wrapper, view } = mountConsent();
    await flushPromises();

    expect(wrapper.vm.selected).toEqual([]);
    expect(wrapper.find(".action-approve").exists()).toBe(false);
    expect(wrapper.find(".consent-unavailable").exists()).toBe(true);

    wrapper.vm.onDeny();
    await flushPromises();

    expect(view.redirect).toHaveBeenCalledWith("https://claude.ai/api/mcp/auth_callback?error=access_denied", 0, false);
  });

  it("keeps the request when approval is refused", async () => {
    Mock.onPost("api/v1/oauth/consent").reply(200, consentResponse);
    Mock.onPost("api/v1/oauth/consent/approve").reply(403, { error: "forbidden" });
    const { wrapper, view, notify } = mountConsent();
    await flushPromises();

    wrapper.vm.onApprove();
    await flushPromises();

    expect(wrapper.vm.request).not.toBe(null);
    expect(wrapper.vm.busy).toBe(false);
    expect(notify.error).toHaveBeenCalled();
    expect(view.redirect).not.toHaveBeenCalled();
  });

  it("shows that the request expired when it is no longer found", async () => {
    Mock.onPost("api/v1/oauth/consent").reply(200, consentResponse);
    Mock.onPost("api/v1/oauth/consent/deny").reply(404, { error: "not found" });
    const { wrapper } = mountConsent();
    await flushPromises();

    wrapper.vm.onDeny();
    await flushPromises();

    expect(wrapper.vm.request).toBe(null);
  });

  it("names the client, account and site in one sentence", async () => {
    Mock.onPost("api/v1/oauth/consent").reply(200, consentResponse);
    const { wrapper } = mountConsent();
    await flushPromises();

    expect(wrapper.vm.consentParams).toEqual({ client: "Claude", account: "Admin", site: "photos.example.com" });
    expect(wrapper.find(".consent-origin").text()).toBe("https://claude.ai");
  });

  it("refuses to follow a script URL", async () => {
    Mock.onPost("api/v1/oauth/consent").reply(200, consentResponse);
    Mock.onPost("api/v1/oauth/consent/deny").reply(200, { redirectUri: "javascript:alert(1)" });
    const { wrapper, view, notify } = mountConsent();
    await flushPromises();

    wrapper.vm.onDeny();
    await flushPromises();

    expect(view.redirect).not.toHaveBeenCalled();
    expect(notify.error).toHaveBeenCalled();
  });

  it("shows an error when the request is missing", async () => {
    const { wrapper } = mountConsent("");
    await flushPromises();

    expect(wrapper.vm.request).toBe(null);
    expect(wrapper.find(".consent-not-found").exists()).toBe(true);
    expect(Mock.history.post.length).toBe(0);
  });
});

describe("redirectAllowed", () => {
  it("allows web URLs and private-use schemes", () => {
    expect(redirectAllowed("https://claude.ai/cb?code=1")).toBe(true);
    expect(redirectAllowed("http://127.0.0.1:8080/cb")).toBe(true);
    expect(redirectAllowed("app.photoprism.mobile:/oauth?code=1")).toBe(true);
  });

  it("refuses script, data and invalid URLs", () => {
    expect(redirectAllowed("javascript:alert(1)")).toBe(false);
    expect(redirectAllowed("JavaScript:alert(1)")).toBe(false);
    expect(redirectAllowed("data:text/html,hi")).toBe(false);
    expect(redirectAllowed("vbscript:msgbox")).toBe(false);
    expect(redirectAllowed("/relative")).toBe(false);
    expect(redirectAllowed("")).toBe(false);
    expect(redirectAllowed(null)).toBe(false);
  });
});
