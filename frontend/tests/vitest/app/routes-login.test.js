import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import "../fixtures";
import routes from "app/routes";
import { $config, $session } from "app/session";

// Find the /login route's beforeEnter so it can be invoked directly with
// mocked to/from/next args. Driving navigation through a real router would
// require booting every page component the file imports — too heavy for a
// guard-focused unit test.
const loginRoute = routes.find((r) => r.name === "login");
const loginGuard = loginRoute.beforeEnter;

describe("app/routes /login guard", () => {
  let restore;

  beforeEach(() => {
    window.localStorage.clear();
    window.sessionStorage.clear();
    // Snapshot the singleton session/config fields we mutate so each test
    // starts from a clean slate.
    const sessionState = {
      authToken: $session.authToken,
      id: $session.id,
      user: $session.user,
      auth: $session.auth,
      loginRedirect: $session.loginRedirect,
    };
    const configValues = $config.values;
    restore = () => {
      $session.authToken = sessionState.authToken;
      $session.id = sessionState.id;
      $session.user = sessionState.user;
      $session.auth = sessionState.auth;
      $session.loginRedirect = sessionState.loginRedirect;
      $config.values = configValues;
    };

    // Force an unauthenticated session by default.
    $session.authToken = null;
    $session.id = null;
    $session.user = { hasId: () => false };
    $session.auth = false;
    $session.loginRedirect = false;
    if ($session.localStorage?.removeItem) {
      $session.localStorage.removeItem("login.next");
      $session.localStorage.removeItem("login.logout");
    }
  });

  afterEach(() => {
    restore?.();
    vi.restoreAllMocks();
  });

  // The legacy bug (#5506 follow-up): a direct visit to /library/login was
  // auto-bouncing through the IdP, so admins could not sign in with local
  // or LDAP/AD credentials when PHOTOPRISM_OIDC_REDIRECT was on. The guard
  // must show the form unless an explicit deep-link target was recorded.
  it("shows the form on a direct visit when PHOTOPRISM_OIDC_REDIRECT is on but no deep-link target is recorded", () => {
    $config.values = {
      ...$config.values,
      ext: { oidc: { enabled: true, redirect: true, loginUri: "/api/v1/oidc/login" } },
    };
    const followRedirect = vi.spyOn($session, "followRedirect").mockImplementation(() => {});
    const next = vi.fn();

    loginGuard({}, {}, next);

    expect(followRedirect).not.toHaveBeenCalled();
    expect(next).toHaveBeenCalledWith();
  });

  it("auto-redirects to the configured OIDC login URI when a deep-link target was recorded", () => {
    $config.values = {
      ...$config.values,
      ext: { oidc: { enabled: true, redirect: true, loginUri: "/api/v1/oidc/login" } },
    };
    $session.setLoginRedirectUrl("/library/albums/at1sqs7gr75pl5r7/view");
    const followRedirect = vi.spyOn($session, "followRedirect").mockImplementation(() => {});
    const next = vi.fn();

    loginGuard({}, {}, next);

    expect(followRedirect).toHaveBeenCalledWith("/api/v1/oidc/login");
    expect(next).toHaveBeenCalledWith(false);
  });

  it("respects the one-shot logout signal and shows the form even with a deep-link target", () => {
    $config.values = {
      ...$config.values,
      ext: { oidc: { enabled: true, redirect: true, loginUri: "/api/v1/oidc/login" } },
    };
    $session.setLoginRedirectUrl("/library/people");
    $session.localStorage.setItem("login.logout", "1");
    const followRedirect = vi.spyOn($session, "followRedirect").mockImplementation(() => {});
    const next = vi.fn();

    loginGuard({}, {}, next);

    expect(followRedirect).not.toHaveBeenCalled();
    expect(next).toHaveBeenCalledWith();
    // The signal is consumed by the guard so a subsequent navigation
    // (e.g. user retries the deep link) can auto-bounce again.
    expect($session.localStorage.getItem("login.logout")).toBeNull();
  });

  it("does not auto-redirect when PHOTOPRISM_OIDC_REDIRECT is off, even on deep-link arrivals", () => {
    $config.values = {
      ...$config.values,
      ext: { oidc: { enabled: true, redirect: false, loginUri: "/api/v1/oidc/login" } },
    };
    $session.setLoginRedirectUrl("/library/photos");
    const followRedirect = vi.spyOn($session, "followRedirect").mockImplementation(() => {});
    const next = vi.fn();

    loginGuard({}, {}, next);

    expect(followRedirect).not.toHaveBeenCalled();
    expect(next).toHaveBeenCalledWith();
  });

  it("does not auto-redirect when OIDC is disabled entirely", () => {
    $config.values = {
      ...$config.values,
      ext: { oidc: { enabled: false, redirect: true, loginUri: "" } },
    };
    $session.setLoginRedirectUrl("/library/folders");
    const followRedirect = vi.spyOn($session, "followRedirect").mockImplementation(() => {});
    const next = vi.fn();

    loginGuard({}, {}, next);

    expect(followRedirect).not.toHaveBeenCalled();
    expect(next).toHaveBeenCalledWith();
  });

  // Post-OIDC roundtrip: auth.gohtml writes the new session and reloads
  // /library/login. The guard must consume the stored deep-link URL and
  // hard-navigate back to it instead of falling through to the default
  // route (which would lose the user's intended destination).
  it("follows the recorded deep-link URL when the user is already authenticated", () => {
    $session.user = { hasId: () => true };
    $session.authToken = "999900000000000000000000000000000000000000000000";
    $session.id = "a9b8ff820bf40ab451910f8bbfe401b2432446693aa539538fbd2399560a722f";
    $session.auth = true;
    $session.setLoginRedirectUrl("/library/albums/xyz/view");
    const followLogin = vi.spyOn($session, "followLoginRedirectUrl").mockImplementation(() => {});
    const next = vi.fn();

    loginGuard({}, {}, next);

    expect(followLogin).toHaveBeenCalled();
    expect(next).toHaveBeenCalledWith(false);
  });

  it("falls back to the default route when the user is authenticated and no deep link is recorded", () => {
    $session.user = { hasId: () => true };
    $session.authToken = "999900000000000000000000000000000000000000000000";
    $session.id = "a9b8ff820bf40ab451910f8bbfe401b2432446693aa539538fbd2399560a722f";
    $session.auth = true;
    const getDefault = vi.spyOn($session, "getDefaultRoute").mockReturnValue("browse");
    const next = vi.fn();

    loginGuard({}, {}, next);

    expect(getDefault).toHaveBeenCalled();
    expect(next).toHaveBeenCalledWith({ name: "browse" });
  });
});
