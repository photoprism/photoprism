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
    <https://www.photoprism.app/trademark/>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>

*/

import { reactive } from "vue";
import { FaceMarkerDisplay, FaceMarkerEdit, isFaceMarkerMode } from "options/face-marker";

// FaceMarkers carries the reactive UI state for the face-marker overlay and
// its controls. Per-photo marker arrays live on the Photo model; this
// singleton holds only state global to "is the overlay active right now".
// The lightbox owns policy and writes here; the sidebar reads and emits
// events for state transitions it wants to request.
export class FaceMarkers {
  constructor() {
    this.mode = null;
    this.busy = false;
    this.pendingNameMarkerUid = "";
    this.hoveredMarkerUid = "";
  }

  // active reports whether any face-marker mode (display or draw) is on.
  get active() {
    return !!this.mode;
  }

  // display reports whether the overlay is in read-only display mode.
  get isDisplay() {
    return this.mode === FaceMarkerDisplay;
  }

  // isEdit reports whether the overlay is in edit mode (drag-to-create + click-to-remove).
  get isEdit() {
    return this.mode === FaceMarkerEdit;
  }

  // setMode flips the state machine to the given mode (or null to clear); invalid values are ignored.
  setMode(mode) {
    if (mode === null || isFaceMarkerMode(mode)) {
      this.mode = mode;
    }
  }

  // display enters read-only display mode (eye toggle on).
  display() {
    this.mode = FaceMarkerDisplay;
  }

  // edit enters edit mode (pencil toggle on — drag-to-create + click-to-remove).
  edit() {
    this.mode = FaceMarkerEdit;
  }

  // exit clears the mode so the lightbox watcher tears down the overlay;
  // paused playback is left paused (not resumed on exit, by design).
  exit() {
    this.mode = null;
  }

  // setBusy toggles the in-flight lock; sidebar buttons gate on this.
  setBusy(b) {
    this.busy = !!b;
  }

  // setPendingNameMarkerUid records the UID of a marker whose name
  // input should auto-focus. Pass "" to clear it.
  setPendingNameMarkerUid(uid) {
    this.pendingNameMarkerUid = typeof uid === "string" ? uid : "";
  }

  // setHoveredMarkerUid records the hovered marker so the matching overlay
  // rect picks up the `--hovered` modifier; sidebar rows drive this on hover.
  setHoveredMarkerUid(uid) {
    this.hoveredMarkerUid = typeof uid === "string" ? uid : "";
  }

  // reset returns every field to its default so a new slide never inherits stale state.
  reset() {
    this.mode = null;
    this.busy = false;
    this.pendingNameMarkerUid = "";
    this.hoveredMarkerUid = "";
  }
}

// $faceMarkers is the shared singleton consumed by PLightbox and PLightboxSidebar.
export const $faceMarkers = reactive(new FaceMarkers());

export default $faceMarkers;
