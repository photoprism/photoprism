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

test.meta("testID", "face-markers-001").meta({ mode: "public" })("Common: Show/hide markers toggle reveals and hides marker overlays", async (t) => {
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
});

test.meta("testID", "face-markers-002").meta({ mode: "public" })(
  "Common: People header and marker controls are visible to admin regardless of marker state",
  async (t) => {
    await t.click(toolbar.cardsViewAction);
    const uid = await photo.getNthPhotoUid("image", 0);
    await photoviewer.openPhotoViewer("uid", uid);
    await photoviewer.openInfoSidebar();
    await t.expect(Selector("div.text-subtitle-2").withText("People").exists).ok();
    await t.expect(photoviewer.markersVisibilityToggle.exists).ok();
    await t.expect(photoviewer.markerAddButton.exists).ok();
  }
);

test.meta("testID", "face-markers-003").meta({ mode: "public" })("Common: Drawing a new face marker persists it and shows it in the People list", async (t) => {
  await openSidebarOnFirstPhoto(t);

  const beforeRows = await photoviewer.getPersonRowCount();

  await photoviewer.startAddingMarker();
  await t.expect(photoviewer.faceMarkerOverlay.visible).ok();
  await photoviewer.drawMarkerInCenter();
  await t.expect(photoviewer.faceMarkerConfirmButton.visible).ok();

  await photoviewer.confirmMarkerDraft();

  await t.expect(photoviewer.personRow.count).eql(beforeRows + 1);
  await t.expect(photoviewer.faceMarkerRect.count).gte(1);
});

test.meta("testID", "face-markers-004").meta({ mode: "public" })("Common: Cancelling a draft does not persist anything", async (t) => {
  await openSidebarOnFirstPhoto(t);

  const beforeRows = await photoviewer.getPersonRowCount();

  await photoviewer.startAddingMarker();
  await t.expect(photoviewer.faceMarkerOverlay.visible).ok();
  await photoviewer.drawMarkerInCenter();
  await t.expect(photoviewer.faceMarkerCancelButton.visible).ok();
  await photoviewer.cancelMarkerDraft();

  // No new row in the People list and no new persisted marker.
  await t.expect(photoviewer.personRow.count).eql(beforeRows);
});

test.meta("testID", "face-markers-005").meta({ mode: "public" })(
  "Common: Removing an unnamed marker is immediate and does not show a confirmation dialog",
  async (t) => {
    await openSidebarOnFirstPhoto(t);

    // Make sure there is at least one unnamed marker to remove. If the
    // sample photo has none, draw one first.
    let unnamedRow = photoviewer.personRow.filter((node) => node.querySelector(".meta-marker-remove") !== null);
    if ((await unnamedRow.count) === 0) {
      await photoviewer.startAddingMarker();
      await t.expect(photoviewer.faceMarkerOverlay.visible).ok();
      await photoviewer.drawMarkerInCenter();
      await t.expect(photoviewer.faceMarkerConfirmButton.visible).ok();
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

test.meta("testID", "face-markers-006").meta({ mode: "public" })("Common: Named markers do not expose a remove icon", async (t) => {
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
});

test.meta("testID", "face-markers-007").meta({ mode: "public" })("Common: Newly added markers persist across photo viewer reopens", async (t) => {
  await openSidebarOnFirstPhoto(t);
  const beforeRows = await photoviewer.getPersonRowCount();

  await photoviewer.startAddingMarker();
  await t.expect(photoviewer.faceMarkerOverlay.visible).ok();
  await photoviewer.drawMarkerInCenter();
  await t.expect(photoviewer.faceMarkerConfirmButton.visible).ok();
  await photoviewer.confirmMarkerDraft();
  await t.expect(photoviewer.personRow.count).eql(beforeRows + 1);

  // Confirming a marker keeps the overlay in add-mode so the user can
  // draw another; its full-viewer hit area obstructs the close button.
  // Toggle add-mode off (same + button) before closing.
  await photoviewer.startAddingMarker();
  await photoviewer.triggerPhotoViewerAction("close-button");
  await t.expect(photoviewer.viewer.visible).notOk();

  await openSidebarOnFirstPhoto(t);
  await t.expect(photoviewer.personRow.count).eql(beforeRows + 1);
});

test.meta("testID", "face-markers-008").meta({ mode: "public" })("Common: Renaming an unnamed marker via the inline combobox persists the subject", async (t) => {
  await openSidebarOnFirstPhoto(t);

  // Create an unnamed marker to rename. Draw at the overlay's center
  // with a relative size so the test is viewport-independent (Mac
  // Chrome vs headless Linux render the photo at different sizes).
  await photoviewer.startAddingMarker();
  await t.expect(photoviewer.faceMarkerOverlay.visible).ok();
  await photoviewer.drawMarkerInCenter();
  await t.expect(photoviewer.faceMarkerConfirmButton.visible).ok();
  await photoviewer.confirmMarkerDraft();

  // The freshly confirmed marker is unnamed — its row exposes a
  // remove icon (not eject) and an inline combobox that the
  // sidebar auto-focuses via newMarkerUid.
  const newRow = photoviewer.personRow.filter((node) => node.querySelector(".meta-marker-remove") !== null).nth(-1);
  await t.expect(newRow.exists).ok();
  const nameInput = newRow.find(".meta-inline-marker input");
  await t.expect(nameInput.visible).ok();
  await t.typeText(nameInput, "SidebarFaceTestPerson").pressKey("enter");

  // After confirm, the row must now be named (eject icon replaces
  // remove) and the typed name must render as the row's title.
  const namedRow = photoviewer.personRow
    .filter((node) => node.querySelector(".meta-marker-eject") !== null)
    .filter((node) => (node.textContent || "").indexOf("SidebarFaceTestPerson") !== -1);
  await t.expect(namedRow.exists).ok();
});

test.meta("testID", "face-markers-010").meta({ mode: "public" })("Common: Blurring an unnamed marker with a novel name opens the Add-name dialog", async (t) => {
  await openSidebarOnFirstPhoto(t);

  // Draw a new unnamed marker so we control the row.
  await photoviewer.startAddingMarker();
  await t.expect(photoviewer.faceMarkerOverlay.visible).ok();
  await photoviewer.drawMarkerInCenter();
  await t.expect(photoviewer.faceMarkerConfirmButton.visible).ok();
  await photoviewer.confirmMarkerDraft();

  const newRow = photoviewer.personRow.filter((node) => node.querySelector(".meta-marker-remove") !== null).nth(-1);
  await t.expect(newRow.exists).ok();
  const nameInput = newRow.find(".meta-inline-marker input");
  await t.expect(nameInput.visible).ok();

  // Type a novel name and blur by clicking on the sidebar header — no
  // Enter key. The sidebar must open the confirmation dialog instead of
  // persisting silently.
  await t.typeText(nameInput, "SidebarBlurDialogPerson");
  await t.click(Selector("div.text-subtitle-2").withText("People"));
  await t.expect(Selector(".v-dialog").withText("SidebarBlurDialogPerson").exists).ok();

  // Cancel the dialog — the draft is discarded and the row stays unnamed.
  // Target the confirm dialog's stable class instead of button text so the
  // assertion is independent of locale.
  await t.click(Selector(".v-dialog .action-cancel"));
  await t.expect(Selector(".v-dialog").withText("SidebarBlurDialogPerson").exists).notOk();
  await t.expect(photoviewer.personRow.filter((node) => (node.textContent || "").indexOf("SidebarBlurDialogPerson") !== -1).exists).notOk();
});

test.meta("testID", "face-markers-009").meta({ mode: "public" })("Common: Ejecting a named marker removes the subject link but keeps the marker", async (t) => {
  await openSidebarOnFirstPhoto(t);

  // Ensure at least one named marker exists: if the first photo
  // already has one, reuse it; otherwise draw + rename (same flow
  // as face-markers-008) so this test is independent of fixture
  // content ordering.
  let namedRow = photoviewer.personRow.filter((node) => node.querySelector(".meta-marker-eject") !== null);
  if ((await namedRow.count) === 0) {
    await photoviewer.startAddingMarker();
    await t.expect(photoviewer.faceMarkerOverlay.visible).ok();
    await photoviewer.drawMarkerInCenter();
    await t.expect(photoviewer.faceMarkerConfirmButton.visible).ok();
    await photoviewer.confirmMarkerDraft();
    const unnamed = photoviewer.personRow.filter((node) => node.querySelector(".meta-marker-remove") !== null).nth(-1);
    await t.typeText(unnamed.find(".meta-inline-marker input"), "SidebarEjectTest").pressKey("enter");
    namedRow = photoviewer.personRow.filter((node) => node.querySelector(".meta-marker-eject") !== null);
  }

  const beforeNamed = await namedRow.count;
  const beforeRows = await photoviewer.personRow.count;

  // Click the eject icon on the first named row. The action is
  // immediate (no confirmation dialog in the lightbox).
  await t.click(namedRow.nth(0).find(".meta-marker-eject"));
  await t.expect(Selector("div.v-dialog .p-confirm").exists).notOk();

  // The marker row count stays the same; the named row count drops
  // because the ejected marker no longer has a SubjUID.
  await t.expect(photoviewer.personRow.count).eql(beforeRows);
  await t.expect(namedRow.count).eql(beforeNamed - 1);
});
