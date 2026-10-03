import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import "../fixtures";
import routes from "app/routes";
import { $config, $session } from "app/session";

describe("app/routes Discover guards", () => {
  let restore;

  beforeEach(() => {
    const sessionState = {
      authToken: $session.authToken,
      id: $session.id,
      user: $session.user,
      auth: $session.auth,
    };
    const configValues = $config.values;
    restore = () => {
      $session.authToken = sessionState.authToken;
      $session.id = sessionState.id;
      $session.user = sessionState.user;
      $session.auth = sessionState.auth;
      $config.values = configValues;
    };
    $session.authToken = "token";
    $session.auth = true;
  });

  afterEach(() => {
    restore();
  });

  it("maps /discover to This Day and /discover/month to This Month", () => {
    expect(routes.find((r) => r.name === "discover").props.tab).toBe(0);
    expect(routes.find((r) => r.name === "discover_month").path).toBe("/discover/month");
    expect(routes.find((r) => r.name === "discover_month").props.tab).toBe(1);
    expect(routes.find((r) => r.name === "discover_random").props.tab).toBe(2);
    expect(routes.some((r) => r.path === "/discover/colors" && r.component)).toBe(false);
  });

  it("redirects the old Similar, Season, and Colors paths to This Day", () => {
    expect(routes.find((r) => r.name === "discover_similar").redirect).toEqual({ name: "discover" });
    expect(routes.find((r) => r.name === "discover_season").redirect).toEqual({ name: "discover" });
    expect(routes.find((r) => r.path === "/discover/colors").redirect).toEqual({ name: "discover" });
  });

  it("sends users away when Discover is disabled", () => {
    const guard = routes.find((r) => r.name === "discover").beforeEnter;
    const next = vi.fn();
    vi.spyOn($config, "feature").mockReturnValue(false);
    vi.spyOn($config, "deny").mockReturnValue(false);
    vi.spyOn($session, "loginRequired").mockReturnValue(false);
    vi.spyOn($session, "getDefaultRoute").mockReturnValue("browse");

    guard({}, {}, next);

    expect(next).toHaveBeenCalledWith({ name: "browse" });
  });
});
