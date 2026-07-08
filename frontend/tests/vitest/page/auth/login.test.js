import { describe, it, expect, beforeEach, vi } from "vitest";
import { shallowMount, flushPromises, config as VTUConfig } from "@vue/test-utils";
import PPageLogin from "page/auth/login.vue";
import { buildNamespace } from "common/storage";
import clientConfig from "../../config";

const storagePrefix = buildNamespace(clientConfig.storageNamespace);

function mountLogin({ oidcEnabled = false, oidcRedirect = false, sessionOverrides = {}, configOverrides = {} } = {}) {
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
    isAuthenticated: vi.fn(() => false),
    consumeLogoutSignal: vi.fn(() => false),
    markLoginRedirectAttempt: vi.fn(),
    markOidcAttempt: vi.fn(),
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
          redirect: oidcRedirect,
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
    // The manual OIDC button arms the per-tab OIDC attempt guard, like the
    // auto-redirect in the /login route guard does. It must NOT arm the
    // post-login redirect-loop guard, or an authenticated return within the
    // loop window would discard the stored deep link / Portal return_to.
    expect(session.markOidcAttempt).toHaveBeenCalledTimes(1);
    expect(session.markLoginRedirectAttempt).not.toHaveBeenCalled();
  });

  it("reset clears the inputs and code-entry state", () => {
    const { wrapper } = mountLogin();

    Object.assign(wrapper.vm, {
      username: "x",
      password: "y",
      showPassword: true,
      useRecoveryCode: true,
      code: "123456",
      enterCode: true,
    });
    wrapper.vm.reset();

    expect(wrapper.vm.username).toBe("");
    expect(wrapper.vm.password).toBe("");
    expect(wrapper.vm.showPassword).toBe(false);
    expect(wrapper.vm.useRecoveryCode).toBe(false);
    expect(wrapper.vm.code).toBe("");
    expect(wrapper.vm.enterCode).toBe(false);
  });

  it("surfaces a stored OIDC sign-in error via the notification toast and consumes it", () => {
    window.localStorage.setItem(`${storagePrefix}session.error`, "You do not have access to this instance.");

    const { wrapper } = mountLogin({ oidcEnabled: true });

    // Consumed on mount so it does not reappear on the next reload. With no
    // messageId, the server-rendered fallback string is surfaced as-is.
    expect(window.localStorage.getItem(`${storagePrefix}session.error`)).toBeNull();
    expect(wrapper.vm.$notify.error).toHaveBeenCalledWith("You do not have access to this instance.", null, []);
  });

  it("forwards a stored OIDC error by message key so it renders in the current UI locale", () => {
    window.localStorage.setItem(`${storagePrefix}session.error`, "Registration disabled");
    window.localStorage.setItem(`${storagePrefix}session.messageId`, "Registration disabled");
    window.localStorage.setItem(`${storagePrefix}session.messageParams`, JSON.stringify([]));

    const { wrapper } = mountLogin({ oidcEnabled: true });

    // All three keys are consumed on mount so the toast does not reappear.
    expect(window.localStorage.getItem(`${storagePrefix}session.error`)).toBeNull();
    expect(window.localStorage.getItem(`${storagePrefix}session.messageId`)).toBeNull();
    expect(window.localStorage.getItem(`${storagePrefix}session.messageParams`)).toBeNull();
    // The message key is forwarded so notify.vue translates it via Tp.
    expect(wrapper.vm.$notify.error).toHaveBeenCalledWith("Registration disabled", "Registration disabled", []);
  });

  // Auto-OIDC redirect for unauthenticated visitors now lives in the
  // /login route guard (see frontend/tests/vitest/app/routes.test.js).
  // Mounting the component must NEVER bounce the user — manual visits
  // to /library/login have to keep working when PHOTOPRISM_OIDC_REDIRECT
  // is on, so users can still authenticate with local or LDAP/AD
  // credentials.
  describe("does not auto-redirect on mount", () => {
    it("never calls followRedirect when oidc.redirect is enabled and the user is unauthenticated", () => {
      const { session } = mountLogin({ oidcEnabled: true, oidcRedirect: true });

      expect(session.followRedirect).not.toHaveBeenCalled();
    });

    it("never calls followRedirect when oidc.redirect is disabled", () => {
      const { session } = mountLogin({ oidcEnabled: true, oidcRedirect: false });

      expect(session.followRedirect).not.toHaveBeenCalled();
    });

    it("never reads the logout signal during mount (route guard owns it)", () => {
      const { session } = mountLogin({ oidcEnabled: true, oidcRedirect: true });

      expect(session.consumeLogoutSignal).not.toHaveBeenCalled();
    });
  });
});
