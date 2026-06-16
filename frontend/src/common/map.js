let loading = false;
let maplibregl = null;

// Screen-space distance (in pixels) below which two photo markers are treated as sharing the
// same spot and merged into one stack. Roughly the marker size, so pictures that would
// otherwise overlap and hide each other are grouped until the map is zoomed in far enough to
// separate them.
const stackTolerancePx = 40;

// projectCoords resolves a feature's [lng, lat] coordinate to a screen point using the given
// projection. It accepts maplibregl Point objects ({x, y}) as well as plain [x, y] arrays and
// falls back to the raw coordinate so the grouping stays testable without a live map.
function projectCoords(coords, project) {
  if (typeof project === "function") {
    const p = project(coords);
    if (p && typeof p.x === "number" && typeof p.y === "number") {
      return [p.x, p.y];
    }
    if (Array.isArray(p) && p.length >= 2) {
      return [p[0], p[1]];
    }
  }
  // No projection: compare raw [lng, lat] degrees. The tolerance is then in degrees, not
  // pixels, and grouping is no longer zoom-aware — callers rendering on a live map MUST pass
  // a projection. This path exists so the grouping can be unit-tested without a map.
  return [coords[0], coords[1]];
}

// groupGeoFeatures groups point features that render close enough to overlap so that
// coincident or near-coincident pictures stack into a single marker with a counter instead of
// hiding each other once the map is zoomed past the clustering threshold. The optional
// `project` callback maps a feature's [lng, lat] coordinate to screen pixels, which keeps the
// grouping zoom-aware: truly coincident pictures always stack, while ones a few meters apart
// only separate once zoomed in far enough. Exact-coordinate keying is intentionally avoided
// because stored coordinates rarely match bit for bit (GPS jitter, tile quantization).
// Features are deduplicated by UID because querySourceFeatures may return the same point once
// per tile it overlaps. Each group carries the lat/lng bounding box of its members and a
// stable key (smallest UID) so a click opens exactly those pictures and the marker survives
// re-renders.
export function groupGeoFeatures(features, project, tolerance = stackTolerancePx) {
  const groups = [];

  if (!Array.isArray(features)) {
    return groups;
  }

  const limit = tolerance * tolerance;

  for (let i = 0; i < features.length; i++) {
    const feature = features[i];
    const coords = feature?.geometry?.coordinates;

    if (!Array.isArray(coords) || coords.length < 2) {
      continue;
    }

    const point = projectCoords(coords, project);

    // Join the first existing group whose anchor lies within the tolerance.
    let group = null;
    for (let g = 0; g < groups.length; g++) {
      const dx = groups[g].point[0] - point[0];
      const dy = groups[g].point[1] - point[1];
      if (dx * dx + dy * dy <= limit) {
        group = groups[g];
        break;
      }
    }

    if (!group) {
      group = {
        coords,
        point,
        features: [],
        uids: {},
        bounds: { latN: coords[1], lngE: coords[0], latS: coords[1], lngW: coords[0] },
      };
      groups.push(group);
    }

    const uid = feature.properties?.UID;

    // Skip duplicate features that reference the same picture.
    if (uid) {
      if (group.uids[uid]) {
        continue;
      }
      group.uids[uid] = true;
    }

    group.features.push(feature);

    // Expand the group's lat/lng bounding box to include this picture.
    const lng = coords[0];
    const lat = coords[1];
    if (lat > group.bounds.latN) {
      group.bounds.latN = lat;
    }
    if (lat < group.bounds.latS) {
      group.bounds.latS = lat;
    }
    if (lng > group.bounds.lngE) {
      group.bounds.lngE = lng;
    }
    if (lng < group.bounds.lngW) {
      group.bounds.lngW = lng;
    }
  }

  // Derive a stable key from the smallest UID so the marker survives re-renders, even when
  // querySourceFeatures returns the same pictures in a different order between frames. Fall
  // back to the first feature's id (never the float coordinates, which drift between frames).
  for (let g = 0; g < groups.length; g++) {
    const group = groups[g];
    const uids = Object.keys(group.uids);
    group.key = uids.length ? uids.sort()[0] : String(group.features[0]?.id ?? g);
  }

  return groups;
}

// Loads the maplibregl library.
export async function load() {
  if (maplibregl !== null || loading) {
    return Promise.resolve(maplibregl);
  }

  loading = true;

  try {
    const module = await import("./maplibregl.js");
    maplibregl = module.default;
    loading = false;
  } catch (e) {
    loading = false;
    console.error("maps: failed to load maplibregl", e);
    return Promise.reject(e);
  }

  return Promise.resolve(maplibregl);
}
