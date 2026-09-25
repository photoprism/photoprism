import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, shallowMount } from "@vue/test-utils";
import Places from "page/places.vue";
import * as maps from "common/map";
import api from "common/api";

vi.mock("common/map", () => ({ supportsWebGL2: vi.fn(), load: vi.fn(), groupGeoFeatures: vi.fn() }));
vi.mock("common/api", () => ({ default: { get: vi.fn() } }));
vi.mock("page/photos.vue", () => ({ default: { template: "<div />" } }));

const message = "Maps are unavailable. Try another browser or device.";
let wrapper;
beforeEach(() => {
  maps.supportsWebGL2.mockReturnValue(true);
});
afterEach(() => {
  wrapper?.unmount();
  wrapper = null;
});

// mountPlaces isolates the renderer while preserving the page's mounted error handling.
function mountPlaces() {
  return shallowMount(Places, {
    global: {
      renderStubDefaultSlot: true,
      mocks: {
        $util: { mapAnimateDuration: () => 0 },
        $route: { query: {}, params: {} },
        $config: {
          getSettings: () => ({ features: {}, maps: { style: "default" } }),
          load: () => Promise.resolve(),
          values: { settings: { maps: { style: "default" }, ui: { language: "en" } } },
          isRtl: () => false, featExperimental: () => false, has: () => false, feature: () => true, allow: () => true, aclClasses: () => "",
        },
      },
    },
  });
}

describe("Places map compatibility", () => {
  it("renders the map-only fallback on WebGL1-only clients", async () => {
    maps.supportsWebGL2.mockReturnValue(false);
    wrapper = mountPlaces();
    await flushPromises();
    expect(wrapper.vm.mapError).toBe(message);
    expect(wrapper.text()).toContain(message);
    expect(maps.load).not.toHaveBeenCalled();
  });

  it("handles renderer loading failures without an unhandled rejection", async () => {
    maps.load.mockRejectedValueOnce(new Error("load failed"));
    wrapper = mountPlaces();
    await flushPromises();
    expect(wrapper.vm.mapError).toBe(message);
  });

  it("always releases the busy state when a style cannot be rendered", () => {
    const context = {
      loading: false, $notify: { blockUI: vi.fn(), unblockUI: vi.fn() },
      $refs: { map: document.createElement("div") }, configureMap: vi.fn(),
      renderMap: vi.fn(() => {
        throw new Error("GPU unavailable");
      }), showMapError: vi.fn(),
    };
    expect(Places.methods.setStyle.call(context, "default")).toBe(false);
    expect(context.showMapError).toHaveBeenCalledOnce();
    expect(context.$notify.unblockUI).toHaveBeenCalledOnce();
  });

  it("waits for clearing the source before reconciling markers", async () => {
    let finish;
    const source = { setData: vi.fn(() => new Promise((resolve) => {
      finish = resolve;
    })) };
    const context = { result: { features: [{}] }, map: { getSource: () => source }, updateMarkers: vi.fn() };
    const done = Places.methods.reset.call(context);
    expect(context.updateMarkers).not.toHaveBeenCalled();
    finish();
    await done;
    expect(context.updateMarkers).toHaveBeenCalledOnce();
    expect(source.setData).toHaveBeenCalledWith({ features: [] });
  });

  it("contains rejected worker updates when clearing a removed source", async () => {
    const source = { setData: vi.fn().mockRejectedValue(new Error("removed")) };
    const context = { result: {}, map: { getSource: () => source }, updateMarkers: vi.fn() };
    await expect(Places.methods.reset.call(context)).resolves.toBeUndefined();
    expect(context.updateMarkers).not.toHaveBeenCalled();
  });

  it("waits for photo data before fitting the map and updating markers", async () => {
    let finish;
    const source = { setData: vi.fn(() => new Promise((resolve) => {
      finish = resolve;
    })) };
    const map = { getSource: () => source, fitBounds: vi.fn() };
    const context = {
      loading: false, initialized: false, filter: { q: "Berlin" }, lastFilter: {}, map,
      closeCluster: vi.fn(), updateQuery: vi.fn(), searchParams: () => ({}),
      updateMarkers: vi.fn(), reset: vi.fn(), animate: 0,
    };
    api.get.mockResolvedValue({ data: { features: [{}], bbox: [13, 52, 14, 53] } });
    const done = Places.methods.search.call(context);
    await flushPromises();
    expect(map.fitBounds).not.toHaveBeenCalled();
    finish();
    await done;
    expect(map.fitBounds).toHaveBeenCalledOnce();
    expect(context.updateMarkers).toHaveBeenCalledOnce();
    expect(context.loading).toBe(false);
  });
});
