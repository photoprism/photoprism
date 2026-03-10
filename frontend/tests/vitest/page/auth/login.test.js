import { describe, it, expect, beforeEach, vi } from "vitest";
import { shallowMount, flushPromises, config as VTUConfig } from "@vue/test-utils";
import PPageLogin from "page/auth/login.vue";
import { buildNamespace } from "common/storage";
import clientConfig from "../../config";

const storagePrefix = buildNamespace(clientConfig.storageNamespace);

function mountLogin({ oidcEnabled = false, sessionOverrides = {}, configOverrides = {} } = {}) {
  const baseConfig = VTUConfig.global.mocks.$config || {};
  const baseSession = VTUConfig.global.mocks.$session || {};
  const baseNotify = VTUConfig.global.mocks.$notify || {};
  const baseView = VTUConfig.global.mocks.$view || {};

  const session = {
    login: vi.fn(() => Promise.resolve(false)),
    followRedirect: vi.fn(),
    followLoginRedirectUrl: vi.fn(),
    useLocalStorage: vi.fn(),
    useSessionStorage: vi.fn(),
    usesSessionStorage: vi.fn(() => false),
    getDefaultRoute: vi.fn(() => "browse"),
    ...baseSession,
    ...sessionOverrides,
  };

  const configMock = {
    ...baseConfig,
    isSponsor: vi.fn(() => false),
    getSiteDescription: vi.fn(() => "Open-Source Photo Management"),
    getSettings: vi.fn(() => ({ ui: { language: "en" }, features: {} })),
    isRtl: vi.fn(() => false),
    values: {
      ...clientConfig,
      registerUri: "",
      passwordResetUri: "",
      wallpaperUri: "",
      ext: {
        oidc: {
          enabled: oidcEnabled,
          loginUri: oidcEnabled ? "/api/v1/oidc/login" : "",
          provider: "OIDC",
          icon: "/oidc.svg",
        },
      },
      ...configOverrides,
    },
  };

  return {
    wrapper: shallowMount(PPageLogin, {
      global: {
        mocks: {
          $config: configMock,
          $session: session,
          $router: { resolve: vi.fn(() => ({ href: "/library/browse" })) },
          $view: {
            ...baseView,
            enter: vi.fn(),
            leave: vi.fn(),
            focus: vi.fn(),
            redirect: vi.fn(),
          },
          $notify: {
            ...baseNotify,
            warn: vi.fn(),
            error: vi.fn(),
          },
        },
        stubs: {
          PAuthHeader: true,
          PAuthFooter: true,
        },
      },
    }),
    session,
  };
}

describe("page/auth/login", () => {
  beforeEach(() => {
    window.localStorage.clear();
    window.sessionStorage.clear();
  });

  it("defaults to staying signed in when no session preference is stored", () => {
    const { wrapper } = mountLogin();
    expect(wrapper.vm.staySignedIn).toBe(true);
  });

  it("initializes from the current session state when sessionStorage is active", () => {
    const { wrapper } = mountLogin({
      sessionOverrides: {
        usesSessionStorage: vi.fn(() => true),
      },
    });

    expect(wrapper.vm.staySignedIn).toBe(false);
  });

  it("falls back to namespaced localStorage when session state is unavailable", () => {
    window.localStorage.setItem(`${storagePrefix}session`, "true");
    const { wrapper } = mountLogin({
      sessionOverrides: {
        usesSessionStorage: undefined,
      },
    });

    expect(wrapper.vm.staySignedIn).toBe(false);
  });

  it("uses localStorage-backed sessions for password login when stay signed in is enabled", async () => {
    const { wrapper, session } = mountLogin();

    wrapper.vm.username = " admin ";
    wrapper.vm.password = " photoprism ";
    wrapper.vm.code = " 123456 ";
    wrapper.vm.staySignedIn = true;
    wrapper.vm.onLogin();

    await flushPromises();

    expect(session.useLocalStorage).toHaveBeenCalledTimes(1);
    expect(session.useSessionStorage).not.toHaveBeenCalled();
    expect(session.login).toHaveBeenCalledWith("admin", "photoprism", "123456");
  });

  it("uses sessionStorage-backed sessions for OIDC login when stay signed in is disabled", () => {
    const { wrapper, session } = mountLogin({ oidcEnabled: true });

    wrapper.vm.staySignedIn = false;
    wrapper.vm.onOidcLogin();

    expect(session.useSessionStorage).toHaveBeenCalledTimes(1);
    expect(session.useLocalStorage).not.toHaveBeenCalled();
    expect(session.followRedirect).toHaveBeenCalledWith("/api/v1/oidc/login");
  });
});
