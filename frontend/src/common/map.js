let loading = false;
let maplibregl = null;

// groupGeoFeatures groups GeoJSON point features by their exact coordinates so that
// pictures sharing the same location render as a single stack marker instead of
// overlapping duplicates that hide each other once the map is zoomed past the
// clustering threshold. Features are deduplicated by UID because querySourceFeatures
// may return the same point more than once across tile boundaries.
export function groupGeoFeatures(features) {
  const groups = [];
  const byKey = {};

  if (!Array.isArray(features)) {
    return groups;
  }

  for (let i = 0; i < features.length; i++) {
    const feature = features[i];
    const coords = feature?.geometry?.coordinates;

    if (!Array.isArray(coords) || coords.length < 2) {
      continue;
    }

    const key = `${coords[0]},${coords[1]}`;
    let group = byKey[key];

    if (!group) {
      group = byKey[key] = { key, coords, features: [], uids: {} };
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
