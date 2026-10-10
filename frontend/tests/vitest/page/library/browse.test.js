import { describe, it, expect } from "vitest";
import { createRouter, createMemoryHistory } from "vue-router";

import PPageBrowse from "page/library/browse.vue";

const router = createRouter({
  history: createMemoryHistory(),
  routes: [{ name: "files", path: "/index/files/:pathMatch(.*)*", component: {} }],
});

// at returns a stub component context for the route that the router resolves for path.
const at = (path) => ({ $route: router.resolve(path) });

describe("page/library/browse.vue", () => {
  describe("routePath", () => {
    it("returns no segments for the Originals root", () => {
      expect(PPageBrowse.methods.routePath.call(at("/index/files"))).toEqual([]);
      expect(PPageBrowse.methods.routePath.call(at("/index/files/"))).toEqual([]);
    });
    it("returns the folder segments for a subfolder", () => {
      expect(PPageBrowse.methods.routePath.call(at("/index/files/2024/04"))).toEqual(["2024", "04"]);
    });
  });
  describe("getBreadcrumbs", () => {
    it("returns no breadcrumbs for the Originals root", () => {
      const ctx = at("/index/files");
      ctx.path = PPageBrowse.methods.routePath.call(ctx);
      expect(PPageBrowse.methods.getBreadcrumbs.call(ctx)).toEqual([]);
    });
    it("links each parent folder of a subfolder", () => {
      const ctx = at("/index/files/2024/04");
      ctx.path = PPageBrowse.methods.routePath.call(ctx);
      expect(PPageBrowse.methods.getBreadcrumbs.call(ctx)).toEqual([
        { key: "B_2024", uri: "/index/files/2024", name: "2024" },
        { key: "B_2024_04", uri: "/index/files/2024/04", name: "04" },
      ]);
    });
  });
});
