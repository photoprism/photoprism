import { afterEach, beforeEach, describe, expect, it, onTestFinished, vi } from "vitest";
import { flushPromises, shallowMount } from "@vue/test-utils";
import { isProxy } from "vue";
import Places from "page/places.vue";
import * as maps from "common/map";
import api from "common/api";

vi.mock("common/map", () => ({ supportsWebGL2: vi.fn(), load: vi.fn(), groupGeoFeatures: vi.fn(), locale: vi.fn(() => ({})) }));
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

  // MapLibre memoizes values like color channels as read-only, non-configurable
  // properties, which a reactive proxy cannot return without throwing.
  it("keeps the renderer, options, and markers raw so memoized read-only values stay readable", async () => {
    const color = {};
    Object.defineProperty(color, "rgb", { value: [0, 0, 0, 1] });
    const renderer = { style: { color }, setLanguage: vi.fn(), on: vi.fn(), addControl: vi.fn(), off: vi.fn(), remove: vi.fn() };
    const marker = { setLngLat: vi.fn().mockReturnThis(), addTo: vi.fn().mockReturnThis(), getElement: () => document.createElement("div"), remove: vi.fn() };
    const control = vi.fn();
    maps.load.mockResolvedValueOnce({
      Map: vi.fn(function () {
        return renderer;
      }),
      Marker: vi.fn(function () {
        return marker;
      }),
      NavigationControl: control, TerrainControl: control, GlobeControl: control, FullscreenControl: control, GeolocateControl: control, ScaleControl: control,
    });
    wrapper = mountPlaces();
    await flushPromises();
    expect(wrapper.vm.mapError).toBe(false);
    expect(isProxy(wrapper.vm.map)).toBe(false);
    expect(isProxy(wrapper.vm.options)).toBe(false);
    expect(wrapper.vm.map.style.color.rgb).toBe(color.rgb);
    const feature = { id: 1, properties: { UID: "ps6sg6be2lvl0y12", Title: "Lake", Hash: "2cad9168fa6acc5c5c2965ddf6ec465ca42fd818" }, geometry: { coordinates: [9.5, 42.5] } };
    wrapper.vm.renderPhotoMarker(feature, "public", {}, wrapper.vm.map);
    expect(isProxy(wrapper.vm.markers[1])).toBe(false);
    expect(marker.addTo).toHaveBeenCalledWith(renderer);
  });

  it("always releases the busy state when a style cannot be rendered", () => {
    const context = {
      loading: false, $notify: { blockUI: vi.fn(), unblockUI: vi.fn() },
      $refs: { map: document.createElement("div") }, removeMap: vi.fn(), configureMap: vi.fn(),
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

  // mapLoadContext returns a page context whose map adds the photo source only when onMapLoad runs.
  function mapLoadContext(q) {
    const source = { setData: vi.fn(() => Promise.resolve()) };
    let added = false;
    const map = {
      getSource: () => (added ? source : undefined), addSource: vi.fn(() => {
        added = true;
      }), addLayer: vi.fn(), on: vi.fn(), off: vi.fn(), fitBounds: vi.fn(),
    };
    const context = {
      loading: false, initialized: false, loaded: false, filter: { q }, lastFilter: {}, map, animate: 0,
      closeCluster: vi.fn(), updateQuery: vi.fn(), searchParams: () => ({}), updateMarkers: vi.fn(), reset: vi.fn(),
      minimizeAttribCtrl: vi.fn(), getProjection: vi.fn(),
    };
    context.search = (force) => Places.methods.search.call(context, force);
    return { context, source, map };
  }

  it.each(["penta", ""])("shows and fits the results of a search that finished before the map loaded (q=%j)", async (q) => {
    const { context, source, map } = mapLoadContext(q);
    api.get.mockResolvedValue({ data: { features: [{}], bbox: [9, 42, 10, 43] } });
    await context.search();
    expect(source.setData).not.toHaveBeenCalled();
    expect(context.initialized).toBe(false);
    Places.methods.onMapLoad.call(context);
    expect(context.loaded).toBe(false);
    await flushPromises();
    expect(api.get).toHaveBeenCalledTimes(2);
    expect(source.setData).toHaveBeenCalledOnce();
    expect(map.fitBounds).toHaveBeenCalledOnce();
    expect(context.loaded).toBe(true);
  });

  it("waits for a search started before the map loaded instead of starting another one", async () => {
    const { context, source, map } = mapLoadContext("");
    let respond;
    api.get.mockReturnValue(new Promise((resolve) => {
      respond = resolve;
    }));
    const done = context.search();
    expect(() => Places.methods.onMapLoad.call(context)).not.toThrow();
    await flushPromises();
    expect(context.loaded).toBe(false);
    respond({ data: { features: [{}], bbox: [9, 42, 10, 43] } });
    await done;
    await flushPromises();
    expect(api.get).toHaveBeenCalledOnce();
    expect(source.setData).toHaveBeenCalledOnce();
    expect(map.fitBounds).toHaveBeenCalledOnce();
    expect(context.loaded).toBe(true);
  });

  it("sets the loaded state after the first search when the map loads", async () => {
    const { context, source } = mapLoadContext("");
    let respond;
    api.get.mockReturnValue(new Promise((resolve) => {
      respond = resolve;
    }));
    Places.methods.onMapLoad.call(context);
    await flushPromises();
    expect(context.loaded).toBe(false);
    respond({ data: { features: [{}], bbox: [9, 42, 10, 43] } });
    await flushPromises();
    expect(api.get).toHaveBeenCalledOnce();
    expect(source.setData).toHaveBeenCalledOnce();
    expect(context.loaded).toBe(true);
  });

  it("draws the atmosphere in globe view and restores the style sky otherwise", () => {
    const map = { getSky: vi.fn(() => undefined), setSky: vi.fn() };
    let projection = "globe";
    const context = { map, _styleSky: null, getProjection: () => projection };
    Places.methods.updateAtmosphere.call(context);
    Places.methods.updateAtmosphere.call(context);
    expect(map.setSky).toHaveBeenCalledOnce();
    expect(map.setSky.mock.calls[0][0]["atmosphere-blend"]).toBeDefined();
    projection = "mercator";
    Places.methods.updateAtmosphere.call(context);
    expect(map.setSky).toHaveBeenLastCalledWith(undefined);
    expect(context._styleSky).toBeNull();
  });

  it("restores a style sky with the defaults of the properties the atmosphere set", () => {
    const map = { getSky: () => ({ "sky-color": "#000000" }), setSky: vi.fn() };
    let projection = "globe";
    const context = { map, _styleSky: null, getProjection: () => projection };
    Places.methods.updateAtmosphere.call(context);
    projection = "mercator";
    Places.methods.updateAtmosphere.call(context);
    const restored = map.setSky.mock.calls[1][0];
    expect(restored["sky-color"]).toBe("#000000");
    expect(restored["atmosphere-blend"]).toBe(0.8);
  });

  it("retries the atmosphere after the style was still loading", () => {
    const map = { getSky: () => undefined, setSky: vi.fn().mockImplementationOnce(() => {
      throw new Error("Style is not done loading.");
    }) };
    const context = { map, _styleSky: null, getProjection: () => "globe" };
    expect(() => Places.methods.updateAtmosphere.call(context)).not.toThrow();
    expect(context._styleSky).toBeNull();
    Places.methods.updateAtmosphere.call(context);
    expect(map.setSky).toHaveBeenCalledTimes(2);
    expect(context._styleSky.map).toBe(map);
  });

  it("renders the sky once and follows the current map", () => {
    vi.stubGlobal("Image", class {
      decode() {
        return new Promise(() => {});
      }
    });
    onTestFinished(() => vi.unstubAllGlobals());
    const first = { on: vi.fn(), off: vi.fn() };
    const second = { on: vi.fn(), off: vi.fn() };
    const context = { $refs: { background: document.createElement("div") }, $config: { staticUri: "/static" }, map: first, _sky: null };
    Places.methods.renderSky.call(context);
    const renderer = context._sky;
    expect(first.on).toHaveBeenCalledWith("render", renderer.onRender);
    context.map = second;
    Places.methods.renderSky.call(context);
    expect(context._sky).toBe(renderer);
    expect(first.off).toHaveBeenCalledWith("render", renderer.onRender);
    expect(second.on).toHaveBeenCalledWith("render", renderer.onRender);
    renderer.destroy();
  });

  it("releases the sky renderer when the map is torn down", () => {
    const destroy = vi.fn();
    const remove = vi.fn();
    const context = { map: { off: vi.fn(), remove }, _sky: { destroy, detach: vi.fn() }, _styleSky: {}, markersOnScreen: {}, clearAllMarkerClicks: vi.fn(), removeMap: Places.methods.removeMap };
    Places.methods.teardownMap.call(context);
    expect(remove).toHaveBeenCalledOnce();
    expect(destroy).toHaveBeenCalledOnce();
    expect(context._sky).toBeNull();
    expect(context._styleSky).toBeNull();
  });

  it("removes the map and its markers but keeps the sky renderer for the next map", () => {
    const map = { off: vi.fn(), remove: vi.fn() };
    const marker = { remove: vi.fn() };
    const handler = vi.fn();
    const sky = { destroy: vi.fn(), detach: vi.fn() };
    const context = { map, _sky: sky, _styleSky: { map }, _mapDataHandler: handler, markers: { 1: marker }, markersOnScreen: { 1: marker }, markerHandlers: { 1: {} }, clearAllMarkerClicks: vi.fn(), updateMarkers: vi.fn() };
    Places.methods.removeMap.call(context);
    expect(map.off).toHaveBeenCalledWith("data", handler);
    ["move", "moveend", "resize", "idle"].forEach((type) => expect(map.off).toHaveBeenCalledWith(type, context.updateMarkers));
    expect(map.remove).toHaveBeenCalledOnce();
    expect(context._mapDataHandler).toBeNull();
    expect(marker.remove).toHaveBeenCalledOnce();
    expect(sky.detach).toHaveBeenCalledOnce();
    expect(sky.destroy).not.toHaveBeenCalled();
    expect(context._sky).toBe(sky);
    expect(context.map).toBeNull();
    expect(context._styleSky).toBeNull();
    expect(context.markers).toEqual({});
    expect(context.markersOnScreen).toEqual({});
  });

  it("removes the previous map before rendering a new style", () => {
    const calls = [];
    const context = {
      loading: false, $notify: { blockUI: vi.fn(), unblockUI: vi.fn() }, $refs: { map: document.createElement("div") },
      removeMap: vi.fn(() => calls.push("removeMap")), configureMap: vi.fn(), renderMap: vi.fn(() => calls.push("renderMap")), showMapError: vi.fn(),
    };
    expect(Places.methods.setStyle.call(context, "default")).toBe(true);
    expect(calls).toEqual(["removeMap", "renderMap"]);
  });

  it("shows CSS stars when the star map cannot be loaded, unless unmounted", async () => {
    vi.stubGlobal("Image", class {
      decode() {
        return Promise.reject(new Error("EncodingError"));
      }
    });
    try {
      for (const unmounted of [false, true]) {
        const el = document.createElement("div");
        const map = { on: vi.fn(), off: vi.fn() };
        const context = { $refs: { background: el }, $config: { staticUri: "/static", debug: false }, map, _sky: null, _mapUnmounted: unmounted };
        Places.methods.renderSky.call(context);
        await flushPromises();
        expect(el.querySelectorAll(".star").length > 0).toBe(!unmounted);
        expect(context._sky.destroyed).toBe(true);
        expect(map.off).toHaveBeenCalledWith("render", context._sky.onRender);
      }
    } finally {
      vi.unstubAllGlobals();
    }
  });

  it("releases the busy state and shows the error when the previous map cannot be removed", () => {
    const context = {
      loading: false, $notify: { blockUI: vi.fn(), unblockUI: vi.fn() }, $refs: { map: document.createElement("div") },
      removeMap: vi.fn(() => {
        throw new Error("remove failed");
      }), configureMap: vi.fn(), renderMap: vi.fn(), showMapError: vi.fn(),
    };
    expect(Places.methods.setStyle.call(context, "default")).toBe(false);
    expect(context.renderMap).not.toHaveBeenCalled();
    expect(context.showMapError).toHaveBeenCalledOnce();
    expect(context.$notify.unblockUI).toHaveBeenCalledOnce();
  });

  it("ignores map shortcuts while no map exists", () => {
    const context = { map: null, initialized: true, loading: false, $view: { isActive: () => true }, setProjection: vi.fn() };
    expect(() => Places.methods.onKeyDown.call(context, new KeyboardEvent("keydown", { code: "ArrowUp" }))).not.toThrow();
    Places.methods.toggleProjection.call(context);
    expect(context.setProjection).not.toHaveBeenCalled();
  });
});
