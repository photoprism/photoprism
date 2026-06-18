/*

Copyright (c) 2018 - 2026 PhotoPrism UG. All rights reserved.

    This program is free software: you can redistribute it and/or modify
    it under Version 3 of the GNU Affero General Public License (the "AGPL"):
    <https://docs.photoprism.app/license/agpl>

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    The AGPL is supplemented by our Trademark and Brand Guidelines,
    which describe how our Brand Assets may be used:
    <https://www.photoprism.app/trademark>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>

*/

import * as media from "common/media";

// EQUIRECTANGULAR_RATIO_TOLERANCE bounds how far a video frame may deviate from the
// 2:1 equirectangular aspect ratio and still be treated as 360° media.
const EQUIRECTANGULAR_RATIO_TOLERANCE = 0.15;

// is360Equirectangular reports whether a slide/thumb model is equirectangular 360°
// content that should open in the sphere viewer. An explicit "equirectangular"
// projection is authoritative, but a full sphere is ~2:1: when the frame size is known
// and clearly not 2:1 it stays in the flat viewer, since a partial/cylindrical panorama
// tagged equirectangular (or promoted from GPano markers) would otherwise render as a
// vertically distorted sphere. Any other non-empty projection (cubemap, …) is never 360°.
// Many 360° videos carry no projection tag PhotoPrism can read, so a panorama-flagged
// video with no projection falls back to the same 2:1 frame test — this keeps genuine
// spherical videos opening while cubemap (4:3, 6:1) and ultrawide (~2.35:1) clips stay flat.
export function is360Equirectangular(model) {
  const projection = (model?.Projection || "").toLowerCase();
  const w = Number(model?.Width);
  const h = Number(model?.Height);
  const known = w > 0 && h > 0;
  const is2to1 = known && Math.abs(w / h - 2) <= EQUIRECTANGULAR_RATIO_TOLERANCE;
  if (projection === "equirectangular") {
    return !known || is2to1;
  }
  if (projection) {
    return false;
  }
  const isVideo = model?.Type === media.Video || model?.Type === media.Animated;
  if (!isVideo || model?.Panorama !== true) {
    return false;
  }
  return is2to1;
}

// createSphereViewer mounts a Photo Sphere Viewer instance for an equirectangular photo or video.
// The renderer (and its ThreeJS dependency) is dynamic-imported on first call so the base bundle
// is unaffected when no 360° media is opened.
export async function createSphereViewer(container, src, opts = {}) {
  const [coreMod, videoMod, videoAdapterMod] = await Promise.all([
    import(/* webpackChunkName: "sphere-viewer" */ "@photo-sphere-viewer/core"),
    opts.isVideo
      ? import(/* webpackChunkName: "sphere-viewer" */ "@photo-sphere-viewer/video-plugin")
      : Promise.resolve(null),
    opts.isVideo
      ? import(/* webpackChunkName: "sphere-viewer" */ "@photo-sphere-viewer/equirectangular-video-adapter")
      : Promise.resolve(null),
  ]);

  await import(/* webpackChunkName: "sphere-viewer" */ "@photo-sphere-viewer/core/index.css");

  if (opts.isVideo) {
    await import(/* webpackChunkName: "sphere-viewer" */ "@photo-sphere-viewer/video-plugin/index.css");
    return new coreMod.Viewer({
      container,
      adapter: [videoAdapterMod.EquirectangularVideoAdapter, { muted: !!opts.muted, autoplay: false }],
      panorama: { source: src },
      plugins: [videoMod.VideoPlugin],
      navbar: false,
      keyboard: "always",
      defaultYaw: 0,
      defaultPitch: 0,
    });
  }

  return new coreMod.Viewer({
    container,
    panorama: src,
    navbar: false,
    keyboard: "always",
    defaultYaw: 0,
    defaultPitch: 0,
  });
}

// findSphereVideoElement returns the underlying HTMLVideoElement that PSV's
// EquirectangularVideoAdapter uses as a WebGL texture source. The adapter does
// not insert the element into the DOM, so we read it off the adapter instance
// the viewer exposes after the panorama has finished loading.
export function findSphereVideoElement(viewer) {
  return viewer?.adapter?.video || null;
}

// destroySphereViewer tears down a viewer instance, releasing its WebGL context and textures.
// Safe to call on null or an already-destroyed viewer.
export function destroySphereViewer(viewer) {
  if (viewer && typeof viewer.destroy === "function") {
    viewer.destroy();
  }
}
