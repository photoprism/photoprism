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
