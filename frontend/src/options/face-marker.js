// Face-marker UI state-machine values. Shared between the lightbox
// (`$faceMarkers.mode`), the face-marker overlay (`mode` prop), and
// the sidebar info panel (which derives `markersVisible` / `addingMarker`
// booleans from these). `null` represents the inactive state; we don't
// give it a name because `null` is meaningful in itself.

// FaceMarkerDisplay shows read-only face boxes over the slide.
export const FaceMarkerDisplay = "display";

// FaceMarkerEdit enables drag-to-create new face regions AND
// click-to-remove existing unnamed markers. Replaces the older
// "draw" naming so the mode label matches the user-facing toggle
// ("Edit Faces"); the mode covers both add and remove gestures.
export const FaceMarkerEdit = "edit";

// FaceMarkerModes enumerates the non-null states the lightbox can be in.
// Useful for prop validators and exhaustive switches.
export const FaceMarkerModes = [FaceMarkerDisplay, FaceMarkerEdit];

// isFaceMarkerMode reports whether the given value is a valid non-null
// face-marker mode. Treats null/undefined as "inactive", not invalid.
export function isFaceMarkerMode(value) {
  return value === FaceMarkerDisplay || value === FaceMarkerEdit;
}
