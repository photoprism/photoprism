import testcafeconfig from "../../testcafeconfig.json";
import PhotoViewer from "../page-model/photoviewer";
import Menu from "../page-model/menu";
import { helperBeforeFixture, helperBeforeEach, helperAfterEach } from "../page-model/helpers";

fixture`Test face markers in the photo viewer`
.page`${testcafeconfig.url}`
.beforeEach(async t => {
  await helperBeforeEach(t);
})
.afterEach(async t => {
  await helperAfterEach(t);
})
.before(async ctx => {
  await helperBeforeFixture(ctx);
});

const photoviewer = new PhotoViewer();
const menu = new Menu();

test.meta("testID", "face-markers-001").meta({ mode: "public" })("Common: Show/hide markers toggle reveals and hides marker overlays", async (t) => {
  // `faces:1` guarantees the photo carries at least one face marker so the rect assertions are non-vacuous.
  await photoviewer.openSidebarOnFirstPhoto("faces:1");

  // Public mode = editable, so only the pencil renders.
  await t.expect(photoviewer.markersEditToggle.visible).ok();
  await t.expect(photoviewer.faceMarkerOverlay.exists).notOk();

  await t.click(photoviewer.markersEditToggle);
  await t.expect(photoviewer.faceMarkerOverlay.visible).ok();
  await t.expect(photoviewer.faceMarkerRect.count).gte(1);

  await t.click(photoviewer.markersEditToggle);
  await t.expect(photoviewer.faceMarkerOverlay.exists).notOk();
});

test.meta("testID", "face-markers-002").meta({ mode: "public" })(
  "Common: People header and marker controls render for editable users on a photo without face markers",
  async (t) => {
    // `faces:no` excludes photos with detected faces
    await photoviewer.openSidebarOnFirstPhoto("faces:no");
    await t.expect(photoviewer.peopleHeader.visible).ok();
    await t.expect(photoviewer.markersEditToggle.visible).ok();
  }
);

test.meta("testID", "face-markers-003").meta({ mode: "public" })("Common: Drawing a new face marker persists it and shows it in the People list", async (t) => {
  // `faces:no` ensures the photo has no pre-existing markers, so the cleanup at the end targets the marker we drew.
  await photoviewer.openSidebarOnFirstPhoto("faces:no");

  const beforeRows = await photoviewer.getPersonRowCount();

  await photoviewer.startMarkersEdit();
  await t.expect(photoviewer.faceMarkerOverlay.visible).ok();
  await photoviewer.drawMarkerInCenter();
  await t.expect(photoviewer.faceMarkerConfirmButton.visible).ok();
  await photoviewer.confirmMarkerDraft();

  await t.expect(photoviewer.personRow.count).eql(beforeRows + 1);
  await t.expect(photoviewer.faceMarkerRect.count).gte(1);

  await photoviewer.removeLastUnnamedMarker();
  await t.expect(photoviewer.personRow.count).eql(beforeRows);
});

test.meta("testID", "face-markers-004").meta({ mode: "public" })("Common: Cancelling a draft does not persist anything", async (t) => {
  await photoviewer.openSidebarOnFirstPhoto("faces:no");

  const beforeRows = await photoviewer.getPersonRowCount();

  await photoviewer.startMarkersEdit();
  await t.expect(photoviewer.faceMarkerOverlay.visible).ok();
  await photoviewer.drawMarkerInCenter();
  await t.expect(photoviewer.faceMarkerCancelButton.visible).ok();
  await photoviewer.cancelMarkerDraft();

  await t.expect(photoviewer.personRow.count).eql(beforeRows);
});

test.meta("testID", "face-markers-005").meta({ mode: "public" })(
  "Common: Removing an unnamed marker requires confirmation via the overlay pill",
  async (t) => {
    await photoviewer.openSidebarOnFirstPhoto("faces:no");

    // Add an unnamed marker so the removal flow is deterministic and self-undoing.
    await photoviewer.startMarkersEdit();
    await t.expect(photoviewer.faceMarkerOverlay.visible).ok();
    await photoviewer.drawMarkerInCenter();
    await t.expect(photoviewer.faceMarkerConfirmButton.visible).ok();
    await photoviewer.confirmMarkerDraft();
    await t.expect(photoviewer.personRow.count).eql(1);

    // Dismiss path: clicking ✕ on the pill keeps the marker.
    await t.click(photoviewer.faceMarkerUnnamedRect.nth(-1));
    await t.expect(photoviewer.faceMarkerRemoveConfirm.visible).ok();
    await t.click(photoviewer.faceMarkerCancelButton);
    await t.expect(photoviewer.faceMarkerRemoveConfirm.exists).notOk();
    await t.expect(photoviewer.personRow.count).eql(1);

    // Confirm path: clicking ✓ on the pill removes the marker.
    await t.click(photoviewer.faceMarkerUnnamedRect.nth(-1));
    await t.expect(photoviewer.faceMarkerRemoveConfirm.visible).ok();
    await t.click(photoviewer.faceMarkerRemoveButton);
    await t.expect(photoviewer.personRow.count).eql(0);
  }
);

test.meta("testID", "face-markers-006").meta({ mode: "public" })("Common: Named markers expose only the Unassign icon in edit mode", async (t) => {
  await photoviewer.openSidebarOnFirstPhoto("faces:no");

  // Create a named marker we control so the assertion is non-vacuous.
  await photoviewer.startMarkersEdit();
  await t.expect(photoviewer.faceMarkerOverlay.visible).ok();
  await photoviewer.drawMarkerInCenter();
  await t.expect(photoviewer.faceMarkerConfirmButton.visible).ok();
  await photoviewer.confirmMarkerDraft();
  const unnamed = photoviewer.unnamedPersonRows.nth(-1);
  await t.typeText(unnamed.find(".meta-inline-marker input"), "UnassignIconTest").pressKey("enter");

  const named = photoviewer.namedPersonRows.withText("UnassignIconTest");
  await t.expect(named.visible).ok();
  await t.expect(named.find(".meta-marker-clear-subject").exists).ok();

  // Undo: clearing the subject unlinks the name, then remove the now-unnamed marker.
  await photoviewer.clearMarkerSubject(named);
  await photoviewer.removeLastUnnamedMarker();
  await t.expect(photoviewer.personRow.count).eql(0);
});

test.meta("testID", "face-markers-007").meta({ mode: "public" })("Common: Newly added markers persist across photo viewer reopens", async (t) => {
  const uid = await photoviewer.openSidebarOnFirstPhoto("faces:no");
  const beforeRows = await photoviewer.getPersonRowCount();

  await photoviewer.startMarkersEdit();
  await t.expect(photoviewer.faceMarkerOverlay.visible).ok();
  await photoviewer.drawMarkerInCenter();
  await t.expect(photoviewer.faceMarkerConfirmButton.visible).ok();
  await photoviewer.confirmMarkerDraft();
  await t.expect(photoviewer.personRow.count).eql(beforeRows + 1);

  // Toggle add-mode off before closing — its full-viewer hit area otherwise blocks the close button.
  await photoviewer.startMarkersEdit();
  await photoviewer.triggerPhotoViewerAction("close-button");
  await t
    .expect(menu.navDrawer.visible).ok() // confirm that the base page is loaded
    .expect(photoviewer.viewer.exists).notOk(); // visible takes 15 seconds.

  // Clear the `faces:no` filter (the photo now has 1 face and would be excluded) and reopen by UID.
  await photoviewer.openSidebarOnPhoto(uid);
  await t.expect(photoviewer.personRow.count).eql(beforeRows + 1);

  await photoviewer.removeLastUnnamedMarker();
  await t.expect(photoviewer.personRow.count).eql(beforeRows);
});

test.meta("testID", "face-markers-008").meta({ mode: "public" })("Common: Naming an unnamed marker via the inline combobox persists the subject", async (t) => {
  await photoviewer.openSidebarOnFirstPhoto("faces:no");
  const beforeRows = await photoviewer.getPersonRowCount();

  await photoviewer.startMarkersEdit();
  await t.expect(photoviewer.faceMarkerOverlay.visible).ok();
  // Center-relative size keeps the draft inside the rendered photo across viewports.
  await photoviewer.drawMarkerInCenter();
  await t.expect(photoviewer.faceMarkerConfirmButton.visible).ok();
  await photoviewer.confirmMarkerDraft();

  const newRow = photoviewer.unnamedPersonRows.nth(-1);
  await t.expect(newRow.visible).ok();
  const nameInput = newRow.find(".meta-inline-marker input");
  await t.expect(nameInput.visible).ok();
  await t.typeText(nameInput, "SidebarFaceTestPerson").pressKey("enter");

  const namedRow = photoviewer.namedPersonRows.withText("SidebarFaceTestPerson");
  await t.expect(namedRow.visible).ok();

  // Undo: clearing the subject unlinks the name, then remove the now-unnamed marker.
  await photoviewer.clearMarkerSubject(namedRow);
  await photoviewer.removeLastUnnamedMarker();
  await t.expect(photoviewer.personRow.count).eql(beforeRows);
});

test.meta("testID", "face-markers-010").meta({ mode: "public" })("Common: Blurring an unnamed marker with a novel name opens the Add-name dialog", async (t) => {
  await photoviewer.openSidebarOnFirstPhoto("faces:no");
  const beforeRows = await photoviewer.getPersonRowCount();

  await photoviewer.startMarkersEdit();
  await t.expect(photoviewer.faceMarkerOverlay.visible).ok();
  await photoviewer.drawMarkerInCenter();
  await t.expect(photoviewer.faceMarkerConfirmButton.visible).ok();
  await photoviewer.confirmMarkerDraft();

  const newRow = photoviewer.unnamedPersonRows.nth(-1);
  await t.expect(newRow.visible).ok();
  const nameInput = newRow.find(".meta-inline-marker input");
  await t.expect(nameInput.visible).ok();

  // Blur via header click (not Enter) — the typed novel name must trigger the Add-name dialog.
  await t.typeText(nameInput, "SidebarBlurDialogPerson");
  await t.click(photoviewer.peopleHeader);
  await t.expect(photoviewer.addNameDialog.withText("SidebarBlurDialogPerson").visible).ok();

  await t.click(photoviewer.addNameDialogCancel);
  await t.expect(photoviewer.addNameDialog.withText("SidebarBlurDialogPerson").exists).notOk();
  await t.expect(photoviewer.personRow.withText("SidebarBlurDialogPerson").exists).notOk();

  // The cancel only discards the typed name — the unnamed marker the draft confirm persisted is still there.
  await photoviewer.removeLastUnnamedMarker();
  await t.expect(photoviewer.personRow.count).eql(beforeRows);
});

test.meta("testID", "face-markers-009").meta({ mode: "public" })("Common: Unassigning a named marker removes the subject link but keeps the marker", async (t) => {
  await photoviewer.openSidebarOnFirstPhoto("faces:no");
  const beforeRows = await photoviewer.getPersonRowCount();

  // Draw + name a marker we control so the test is deterministic.
  await photoviewer.startMarkersEdit();
  await t.expect(photoviewer.faceMarkerOverlay.visible).ok();
  await photoviewer.drawMarkerInCenter();
  await t.expect(photoviewer.faceMarkerConfirmButton.visible).ok();
  await photoviewer.confirmMarkerDraft();
  const unnamed = photoviewer.unnamedPersonRows.nth(-1);
  await t.typeText(unnamed.find(".meta-inline-marker input"), "SidebarClearSubjectTest").pressKey("enter");

  const namedRow = photoviewer.namedPersonRows.withText("SidebarClearSubjectTest");
  await t.expect(namedRow.visible).ok();
  const totalAfterName = await photoviewer.personRow.count;

  await photoviewer.clearMarkerSubject(namedRow);

  // Clearing the subject unlinks the name without removing the marker — total rows stay, the named row goes away.
  await t.expect(photoviewer.personRow.count).eql(totalAfterName);
  await t.expect(namedRow.exists).notOk();

  await photoviewer.removeLastUnnamedMarker();
  await t.expect(photoviewer.personRow.count).eql(beforeRows);
});
