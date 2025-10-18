import { describe, expect, it } from "vitest";
import ClusterNode from "../../../../pro/portal/frontend/model/cluster-node.js";

describe("pro/portal/model/cluster-node", () => {
  it("derives the node identifier", () => {
    const node = new ClusterNode({ UUID: "abc-123", Name: "portal-node" });

    expect(node.getId()).toBe("abc-123");
  });

  it("formats role labels", () => {
    const admin = new ClusterNode({ Role: "admin" });
    const unknown = new ClusterNode({ Role: "custom" });

    expect(admin.roleLabel()).toContain("Admin");
    expect(unknown.roleLabel()).toBe("Custom");
  });

  it("converts labels into sorted key/value pairs", () => {
    const node = new ClusterNode({
      Labels: {
        beta: "true",
        alpha: "42",
      },
    });

    expect(node.labelEntries()).toEqual([
      { key: "alpha", value: "42" },
      { key: "beta", value: "true" },
    ]);
  });

  it("reports database metadata availability", () => {
    const node = new ClusterNode({
      Database: {
        name: "photoprism",
        user: "photoprism",
        driver: "mysql",
      },
    });

    expect(node.hasDatabase()).toBe(true);
  });
});

