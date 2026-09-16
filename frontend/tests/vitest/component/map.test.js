import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, mount } from "@vue/test-utils";
import PMap from "component/map.vue";

const mocks = vi.hoisted(() => ({ load: vi.fn(), supportsWebGL2: vi.fn() }));
vi.mock("common/map", () => mocks);

let wrapper;
let renderer;
let marker;
let events;
let markerEvents;
let namespace;

// mountMap mounts a map with isolated renderer and configuration dependencies.
const mountMap = (props = {}) => mount(PMap, {
  props: { latlng: [47.5, 16.5], ...props },
  global: { mocks: { $util: { mapAnimateDuration: () => 0 }, $config: { getSettings: () => ({}) } } },
});

beforeEach(() => {
  events = {};
  markerEvents = {};
  renderer = {
    on: vi.fn((event, handler) => {
      events[event] = handler;
    }),
    addControl: vi.fn(), addImage: vi.fn(), resize: vi.fn(), remove: vi.fn(),
    setMissingStyleImageResolver: vi.fn(), setCenter: vi.fn(), jumpTo: vi.fn(), flyTo: vi.fn(),
  };
  marker = {
    setLngLat: vi.fn().mockReturnThis(), addTo: vi.fn().mockReturnThis(), remove: vi.fn(),
    getLngLat: () => ({ lat: 48, lng: 17 }), getElement: () => document.createElement("button"),
    on: vi.fn((event, handler) => {
      markerEvents[event] = handler;
    }),
  };
  namespace = {
    Map: vi.fn(function () {
      return renderer;
    }),
    Marker: vi.fn(function () {
      return marker;
    }),
    NavigationControl: vi.fn(), ScaleControl: vi.fn(), GeolocateControl: vi.fn(),
  };
  mocks.supportsWebGL2.mockReturnValue(true);
  mocks.load.mockResolvedValue(namespace);
});
afterEach(() => {
  wrapper?.unmount();
  wrapper = null;
  vi.restoreAllMocks();
});

describe("PMap", () => {
  it("shows an accessible fallback without loading a renderer on unsupported clients", async () => {
    mocks.supportsWebGL2.mockReturnValue(false);
    wrapper = mountMap();
    await flushPromises();
    expect(wrapper.get('[role="status"]').text()).toContain("Maps are unavailable");
    expect(mocks.load).not.toHaveBeenCalled();
    expect(wrapper.emitted("update:latlng")).toBeUndefined();
  });

  it("handles module and GPU initialization failures", async () => {
    mocks.load.mockRejectedValueOnce(new Error("worker load"));
    wrapper = mountMap();
    await flushPromises();
    expect(wrapper.get('[role="status"]').text()).toContain("Maps are unavailable");
    wrapper.unmount();
    vi.spyOn(console, "error").mockImplementation(() => {});
    namespace.Map.mockImplementationOnce(() => {
      throw new Error("GPU unavailable");
    });
    wrapper = mountMap();
    await flushPromises();
    expect(wrapper.get('[role="status"]').text()).toContain("Maps are unavailable");
  });

  it("does not create a map after unmounting during module loading", async () => {
    let resolve;
    mocks.load.mockReturnValue(new Promise((done) => {
      resolve = done;
    }));
    wrapper = mountMap();
    wrapper.unmount();
    wrapper = null;
    resolve(namespace);
    await flushPromises();
    expect(namespace.Map).not.toHaveBeenCalled();
  });

  it("preserves controls, coordinate ordering, click and drag events, and cleanup", async () => {
    wrapper = mountMap({ interactive: true, clickable: true, draggable: true, showControls: true });
    await flushPromises();
    expect(namespace.Map).toHaveBeenCalledWith(expect.objectContaining({ center: [16.5, 47.5] }));
    expect(renderer.addControl).toHaveBeenCalledTimes(3);
    expect(renderer.setMissingStyleImageResolver).toHaveBeenCalledOnce();
    expect(events.styleimagemissing).toBeUndefined();
    events.load();
    expect(marker.setLngLat).toHaveBeenCalledWith([16.5, 47.5]);
    events.click({ lngLat: { lat: 48, lng: 17 } });
    markerEvents.dragend();
    expect(wrapper.emitted("update:latlng")).toEqual([[[48, 17]], [[48, 17]]]);
    expect(wrapper.emitted("marker-moved")[0]).toEqual([{ lat: 48, lng: 17 }]);
    wrapper.unmount();
    wrapper = null;
    expect(renderer.remove).toHaveBeenCalledOnce();
  });
});
