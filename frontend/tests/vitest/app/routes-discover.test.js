import { describe, expect, it } from "vitest";
import "../fixtures";
import routes from "app/routes";

describe("app/routes discover", () => {
  it("routes the Memories tab under Discover", () => {
    const memories = routes.find((route) => route.name === "discover_memories");
    const similar = routes.find((route) => route.name === "discover_similar");
    const season = routes.find((route) => route.name === "discover_season");
    const random = routes.find((route) => route.name === "discover_random");

    expect(memories).toMatchObject({
      path: "/discover/memories",
      props: { tab: 1 },
    });
    expect(similar.props).toEqual({ tab: 2 });
    expect(season.props).toEqual({ tab: 3 });
    expect(random.props).toEqual({ tab: 4 });
  });
});
