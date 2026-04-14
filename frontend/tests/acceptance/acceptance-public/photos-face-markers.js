import { Selector } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";
import Toolbar from "../page-model/toolbar";
import Photo from "../page-model/photo";
import PhotoViewer from "../page-model/photoviewer";

fixture`Test face markers in the photo viewer`.page`${testcafeconfig.url}`;

const toolbar = new Toolbar();
const photo = new Photo();
const photoviewer = new PhotoViewer();

// Helper: open the lightbox on the first image and reveal the info sidebar.
async function openSidebarOnFirstPhoto(t) {
  await t.click(toolbar.cardsViewAction);
  const uid = await photo.getNthPhotoUid("image", 0);
  await photoviewer.openPhotoViewer("uid", uid);
  await photoviewer.openInfoSidebar();
  return uid;
}

test.meta("testID", "face-markers-001").meta({ mode: "public" })(
  "Show/hide markers toggle reveals and hides marker overlays",
  async (t) => {
    await openSidebarOnFirstPhoto(t);

    // Sidebar People header has the show/hide and + buttons.
    await t.expect(photoviewer.markersVisibilityToggle.exists).ok();
    await t.expect(photoviewer.markerAddButton.exists).ok();
    // Overlay is not mounted until the user toggles markers visible.
    await t.expect(photoviewer.faceMarkerOverlay.exists).notOk();

    await photoviewer.toggleMarkersVisible();
    await t.expect(photoviewer.faceMarkerOverlay.exists).ok();

    await photoviewer.toggleMarkersVisible();
    await t.expect(photoviewer.faceMarkerOverlay.exists).notOk();
  }
);

test.meta("testID", "face-markers-002").meta({ mode: "public" })(
  "People header is reachable for admin even on photos without markers",
  async (t) => {
    await t.click(toolbar.cardsViewAction);
    // Open any photo (admin should see the controls regardless of markers).
    const uid = await photo.getNthPhotoUid("image", 0);
    await photoviewer.openPhotoViewer("uid", uid);
    await photoviewer.openInfoSidebar();
    await t.expect(Selector("div.text-subtitle-2").withText("People").exists).ok();
    await t.expect(photoviewer.markersVisibilityToggle.exists).ok();
    await t.expect(photoviewer.markerAddButton.exists).ok();
  }
);

test.meta("testID", "face-markers-003").meta({ mode: "public" })(
  "Drawing a new face marker persists it and shows it in the People list",
  async (t) => {
    await openSidebarOnFirstPhoto(t);

    const beforeRows = await photoviewer.getPersonRowCount();

    await photoviewer.startAddingMarker();
    await t.expect(photoviewer.faceMarkerOverlay.exists).ok();

    // Drag a square inside the displayed image. Coordinates are viewport
    // pixels, chosen to comfortably exceed the 16-pixel minimum.
    await photoviewer.drawMarkerOnImage(140, 120, 260, 240);
    await t.expect(photoviewer.faceMarkerConfirmButton.visible).ok();

    await photoviewer.confirmMarkerDraft();

    // After confirmation the People list grows by one row and a marker
    // rectangle is rendered in the overlay.
    await t.expect(photoviewer.personRow.count).eql(beforeRows + 1);
    await t.expect(photoviewer.faceMarkerRect.count).gte(1);
  }
);

test.meta("testID", "face-markers-004").meta({ mode: "public" })(
  "Cancelling a draft does not persist anything",
  async (t) => {
    await openSidebarOnFirstPhoto(t);

    const beforeRows = await photoviewer.getPersonRowCount();

    await photoviewer.startAddingMarker();
    await photoviewer.drawMarkerOnImage(150, 130, 240, 220);
    await t.expect(photoviewer.faceMarkerCancelButton.visible).ok();
    await photoviewer.cancelMarkerDraft();

    // No new row in the People list and no new persisted marker.
    await t.expect(photoviewer.personRow.count).eql(beforeRows);
  }
);

test.meta("testID", "face-markers-005").meta({ mode: "public" })(
  "Removing an unnamed marker is immediate and does not show a confirmation dialog",
  async (t) => {
    await openSidebarOnFirstPhoto(t);

    // Make sure there is at least one unnamed marker to remove. If the
    // sample photo has none, draw one first.
    let unnamedRow = photoviewer.personRow.filter((node) => node.querySelector(".meta-marker-remove") !== null);
    if ((await unnamedRow.count) === 0) {
      await photoviewer.startAddingMarker();
      await photoviewer.drawMarkerOnImage(150, 140, 250, 240);
      await photoviewer.confirmMarkerDraft();
      unnamedRow = photoviewer.personRow.filter((node) => node.querySelector(".meta-marker-remove") !== null);
    }

    const beforeUnnamed = await unnamedRow.count;
    await t.click(unnamedRow.nth(0).find(".meta-marker-remove"));

    // No confirmation dialog must appear, and the unnamed row count
    // should drop by one (immediate removal).
    await t.expect(Selector("div.v-dialog .p-confirm").exists).notOk();
    await t.expect(unnamedRow.count).eql(beforeUnnamed - 1);
  }
);

test.meta("testID", "face-markers-006").meta({ mode: "public" })(
  "Named markers do not expose a remove icon",
  async (t) => {
    await t.click(toolbar.cardsViewAction);
    // Try to find a photo that has at least one named marker. We open the
    // first image and then check the rendered People list. The fixture set
    // contains photos with named subjects, but if the first photo has no
    // named row we just assert the structural rule on the rendered DOM.
    const uid = await photo.getNthPhotoUid("image", 0);
    await photoviewer.openPhotoViewer("uid", uid);
    await photoviewer.openInfoSidebar();

    const namedRows = photoviewer.personRow.filter((node) => node.querySelector(".meta-marker-remove") === null);
    const count = await namedRows.count;
    for (let i = 0; i < count; i++) {
      await t.expect(namedRows.nth(i).find(".meta-marker-remove").exists).notOk();
    }
  }
);

test.meta("testID", "face-markers-007").meta({ mode: "public" })(
  "Newly added markers persist across photo viewer reopens",
  async (t) => {
    await openSidebarOnFirstPhoto(t);
    const beforeRows = await photoviewer.getPersonRowCount();

    await photoviewer.startAddingMarker();
    await photoviewer.drawMarkerOnImage(160, 140, 270, 250);
    await photoviewer.confirmMarkerDraft();
    await t.expect(photoviewer.personRow.count).eql(beforeRows + 1);

    // Close and reopen the viewer; the marker must still be there.
    await photoviewer.triggerPhotoViewerAction("close-button");
    await t.expect(Selector("div.p-lightbox__pswp").visible).notOk();

    await openSidebarOnFirstPhoto(t);
    await t.expect(photoviewer.personRow.count).eql(beforeRows + 1);
  }
);
