import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import "../fixtures";
import routes, { safeReturnTo } from "app/routes";
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

  // One-shot guard: the first deep-link arrival auto-redirects to OIDC and
  // marks the attempt; a second arrival within the same tab (typical of a
  // failed/abandoned IdP roundtrip) consumes the flag and shows the form
  // instead of looping back to OIDC.
  it("auto-redirects only once per tab so a failed OIDC roundtrip falls back to the form", () => {
    $config.values = {
      ...$config.values,
      ext: { oidc: { enabled: true, redirect: true, loginUri: "/api/v1/oidc/login" } },
    };
    $session.setLoginRedirectUrl("/library/people");
    const followRedirect = vi.spyOn($session, "followRedirect").mockImplementation(() => {});
    const firstNext = vi.fn();
    const secondNext = vi.fn();

    loginGuard({}, {}, firstNext);
    expect(followRedirect).toHaveBeenCalledTimes(1);
    expect(firstNext).toHaveBeenCalledWith(false);

    // Simulate the user dismissing the IdP and arriving back at /login. The
    // deep-link target is still in localStorage; the attempt flag in
    // sessionStorage breaks the loop.
    loginGuard({}, {}, secondNext);
    expect(followRedirect).toHaveBeenCalledTimes(1);
    expect(secondNext).toHaveBeenCalledWith();
  });

  it("re-arms the OIDC auto-redirect after a successful login or explicit logout", () => {
    $config.values = {
      ...$config.values,
      ext: { oidc: { enabled: true, redirect: true, loginUri: "/api/v1/oidc/login" } },
    };
    $session.setLoginRedirectUrl("/library/people");
    const followRedirect = vi.spyOn($session, "followRedirect").mockImplementation(() => {});

    loginGuard({}, {}, vi.fn());
    expect(followRedirect).toHaveBeenCalledTimes(1);

    // Successful login (or onLogout) clears both the deep-link target and
    // the attempt flag via clearLoginRedirectUrl. A fresh deep-link arrival
    // should be allowed to auto-redirect again.
    $session.clearLoginRedirectUrl();
    $session.setLoginRedirectUrl("/library/albums/new/view");

    loginGuard({}, {}, vi.fn());
    expect(followRedirect).toHaveBeenCalledTimes(2);
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

  // Defense in depth: if the recorded deep link keeps bouncing the authenticated
  // user straight back to /login (e.g. the OIDC OP can't authenticate the
  // navigation), the guard must drop the target and fall through to the default
  // route instead of looping forever.
  it("breaks the redirect loop when the recorded deep link keeps bouncing back", () => {
    $session.user = { hasId: () => true };
    $session.authToken = "999900000000000000000000000000000000000000000000";
    $session.id = "a9b8ff820bf40ab451910f8bbfe401b2432446693aa539538fbd2399560a722f";
    $session.auth = true;
    $session.setLoginRedirectUrl("/api/v1/oauth/authorize?client_id=abc&state=x");
    $session.markLoginRedirectAttempt(); // simulate having just followed it
    const followLogin = vi.spyOn($session, "followLoginRedirectUrl").mockImplementation(() => {});
    const getDefault = vi.spyOn($session, "getDefaultRoute").mockReturnValue("browse");
    const next = vi.fn();

    loginGuard({}, {}, next);

    expect(followLogin).not.toHaveBeenCalled();
    expect(getDefault).toHaveBeenCalled();
    expect(next).toHaveBeenCalledWith({ name: "browse" });
    expect($session.hasLoginRedirectUrl()).toBe(false);
  });

  // The Portal OIDC OP redirects unauthenticated users to
  // /portal/login?return_to=<authorize URL> via a top-level browser
  // navigation, so the global router guard never gets to record the deep
  // link via setLoginRedirectUrl(). The login route reads the inbound
  // `return_to` query parameter directly to bridge the hand-off.
  it("records a safe return_to query param as the post-login deep link", () => {
    const next = vi.fn();

    loginGuard({ query: { return_to: "/api/v1/oauth/authorize?client_id=abc&state=x" } }, {}, next);

    expect($session.hasLoginRedirectUrl()).toBe(true);
    expect($session.getLoginRedirectUrl()).toBe("/api/v1/oauth/authorize?client_id=abc&state=x");
    expect(next).toHaveBeenCalledWith();
  });

  it("ignores an unsafe return_to that escapes the current origin", () => {
    const next = vi.fn();

    loginGuard({ query: { return_to: "https://attacker.example/steal" } }, {}, next);

    expect($session.hasLoginRedirectUrl()).toBe(false);
    expect(next).toHaveBeenCalledWith();
  });
});

describe("app/routes /logout guard", () => {
  const logoutGuard = routes.find((r) => r.name === "logout").beforeEnter;
  let restore;

  beforeEach(() => {
    const configValues = $config.values;
    const provider = $session.provider;
    restore = () => {
      $config.values = configValues;
      $session.provider = provider;
    };
  });

  afterEach(() => {
    restore?.();
    vi.restoreAllMocks();
  });

  // A cluster-OIDC user who hits /logout directly is bounced to the Portal login
  // page so they re-auth with their cluster account and pick an instance there.
  // The redirect must wait for the cluster-wide sign-out (which clears the Portal OP
  // cookie) so the Portal shows its login form instead of silently re-issuing a session.
  it("bounces a cluster-OIDC sign-out to the Portal login after clearing the OP cookie", async () => {
    $config.values = {
      ...$config.values,
      ext: {
        oidc: {
          enabled: true,
          redirect: true,
          cluster: true,
          loginUri: "/api/v1/oidc/login",
          portalLoginUri: "https://app.example.com/portal/login",
        },
      },
    };
    $session.provider = "oidc";
    // logoutEverywhere resolves to the landing URL (here the Portal login, since RP-logout
    // is off so no provider logout URL is returned).
    const logoutEverywhere = vi.spyOn($session, "logoutEverywhere").mockResolvedValue("https://app.example.com/portal/login");
    const followRedirect = vi.spyOn($session, "followRedirect").mockImplementation(() => {});
    const next = vi.fn();

    logoutGuard({}, {}, next);

    // The SPA navigation is cancelled immediately, but the redirect is deferred
    // until the cluster-wide sign-out resolves.
    expect(next).toHaveBeenCalledWith(false);
    expect(logoutEverywhere).toHaveBeenCalledWith(true);
    expect(followRedirect).not.toHaveBeenCalled();

    await new Promise((resolve) => setTimeout(resolve, 0));
    expect(followRedirect).toHaveBeenCalledWith("https://app.example.com/portal/login");
  });

  // With RP-initiated logout enabled, logoutEverywhere resolves to the provider logout URL
  // (the Portal end-session endpoint); the guard must follow it so the upstream session ends.
  it("follows the provider logout URL on a cluster-OIDC sign-out with RP-initiated logout", async () => {
    $config.values = {
      ...$config.values,
      ext: { oidc: { enabled: true, redirect: true, cluster: true, logout: true, loginUri: "/api/v1/oidc/login", portalLoginUri: "https://app.example.com/portal/login" } },
    };
    $session.provider = "oidc";
    const providerLogoutUri = "https://app.example.com/api/v1/oauth/logout?id_token_hint=abc";
    const logoutEverywhere = vi.spyOn($session, "logoutEverywhere").mockResolvedValue(providerLogoutUri);
    const followRedirect = vi.spyOn($session, "followRedirect").mockImplementation(() => {});
    const next = vi.fn();

    logoutGuard({}, {}, next);

    expect(next).toHaveBeenCalledWith(false);
    expect(logoutEverywhere).toHaveBeenCalledWith(true);

    await new Promise((resolve) => setTimeout(resolve, 0));
    expect(followRedirect).toHaveBeenCalledWith(providerLogoutUri);
  });

  // A node that has not (re-)registered against a current Portal has no Portal
  // login URL yet — the cluster-wide sign-out and OP-cookie clearing must still
  // run, with the local login form as the fallback landing.
  it("still signs out cluster-wide when the Portal login URL is unknown", async () => {
    $config.values = {
      ...$config.values,
      ext: { oidc: { enabled: true, redirect: true, cluster: true, loginUri: "/api/v1/oidc/login" } },
    };
    $session.provider = "oidc";
    const logoutEverywhere = vi.spyOn($session, "logoutEverywhere").mockResolvedValue($config.loginUri);
    const followRedirect = vi.spyOn($session, "followRedirect").mockImplementation(() => {});
    const next = vi.fn();

    logoutGuard({}, {}, next);

    expect(next).toHaveBeenCalledWith(false);
    expect(logoutEverywhere).toHaveBeenCalledWith(true);

    await new Promise((resolve) => setTimeout(resolve, 0));
    expect(followRedirect).toHaveBeenCalledWith($config.loginUri);
  });

  // A non-cluster OIDC node reached via direct /logout entry must still chain RP-initiated
  // logout: the guard runs the cluster-wide sign-out (peer fan-out is a no-op when there are
  // no peers) and follows the provider logout URL it resolves, matching the nav-menu Sign-Out.
  it("follows the provider logout URL on a standalone OIDC sign-out with RP-initiated logout", async () => {
    $config.values = {
      ...$config.values,
      ext: { oidc: { enabled: true, redirect: false, cluster: false, logout: true, loginUri: "/api/v1/oidc/login" } },
    };
    $session.provider = "oidc";
    const providerLogoutUri = "https://keycloak.example.com/realms/master/protocol/openid-connect/logout?id_token_hint=abc";
    const logoutEverywhere = vi.spyOn($session, "logoutEverywhere").mockResolvedValue(providerLogoutUri);
    const followRedirect = vi.spyOn($session, "followRedirect").mockImplementation(() => {});
    const next = vi.fn();

    logoutGuard({}, {}, next);

    expect(next).toHaveBeenCalledWith(false);
    expect(logoutEverywhere).toHaveBeenCalledWith(true);

    await new Promise((resolve) => setTimeout(resolve, 0));
    expect(followRedirect).toHaveBeenCalledWith(providerLogoutUri);
  });

  it("sends a local sign-out to the login route", () => {
    $config.values = {
      ...$config.values,
      ext: { oidc: { enabled: false, redirect: false, cluster: false, loginUri: "" } },
    };
    $session.provider = "local";
    vi.spyOn($session, "signOut").mockReturnValue($session);
    const followRedirect = vi.spyOn($session, "followRedirect").mockImplementation(() => {});
    const next = vi.fn();

    logoutGuard({}, {}, next);

    expect(followRedirect).not.toHaveBeenCalled();
    expect(next).toHaveBeenCalledWith({ name: "login" });
  });
});

describe("app/routes safeReturnTo", () => {
  it("accepts root-relative paths", () => {
    expect(safeReturnTo("/library/photos")).toBe("/library/photos");
    expect(safeReturnTo("/api/v1/oauth/authorize?client_id=x&state=y")).toBe("/api/v1/oauth/authorize?client_id=x&state=y");
  });
  it("accepts absolute URLs on the same origin and returns the path+query+hash", () => {
    const here = window.location?.origin;
    if (!here) {
      return;
    }
    expect(safeReturnTo(here + "/portal/cluster")).toBe("/portal/cluster");
    expect(safeReturnTo(here + "/api/v1/oauth/authorize?client_id=x#frag")).toBe("/api/v1/oauth/authorize?client_id=x#frag");
  });
  it("rejects cross-origin absolutes", () => {
    expect(safeReturnTo("https://attacker.example/steal")).toBe("");
    expect(safeReturnTo("http://attacker.example/steal")).toBe("");
  });
  it("rejects protocol-relative and backslash-prefixed values that some browsers misparse", () => {
    expect(safeReturnTo("//attacker.example/")).toBe("");
    expect(safeReturnTo("\\\\attacker.example\\path")).toBe("");
  });
  it("rejects empty, whitespace, or non-string inputs", () => {
    expect(safeReturnTo("")).toBe("");
    expect(safeReturnTo("   ")).toBe("");
    expect(safeReturnTo(undefined)).toBe("");
    expect(safeReturnTo(null)).toBe("");
    expect(safeReturnTo(42)).toBe("");
  });
});
