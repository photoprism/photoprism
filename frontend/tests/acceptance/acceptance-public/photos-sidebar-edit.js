import { Selector } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";
import PhotoViewer from "../page-model/photoviewer";
import Menu from "../page-model/menu";
import Label from "../page-model/label";
import Album from "../page-model/album";

// Drives inline editing of every editable sidebar field against a real
// backend. The companion Vitest matrix covers per-role visibility; these
// tests pin the DOM wiring and persistence path end-to-end.
fixture`Test lightbox sidebar inline editing`.page`${testcafeconfig.url}`;

const photoviewer = new PhotoViewer();
const menu = new Menu();
const label = new Label();
const album = new Album();

test.meta("testID", "sidebar-edit-001").meta({ mode: "public" })(
  "Common: Edits title, caption, keywords, notes, and plain-text inline fields from the sidebar",
  async (t) => {
    await photoviewer.openSidebarOnFirstPhoto();

    const titleInput = Selector(".p-sidebar-info .meta-inline-title input", { timeout: 15000 });
    await photoviewer.startInlineEditOrAdd("meta-title", "Title");
    await t.expect(titleInput.visible).ok();
    await t.typeText(titleInput, "Sidebar Edit Title", { replace: true }).pressKey("enter");
    await t.expect(Selector(".p-sidebar-info .meta-title").withText("Sidebar Edit Title").exists).ok();

    const captionTextarea = Selector(".p-sidebar-info .meta-inline-caption textarea", { timeout: 15000 });
    await photoviewer.startInlineEditOrAdd("meta-caption", "Caption");
    await t.expect(captionTextarea.visible).ok();
    await t.typeText(captionTextarea, "Caption added in sidebar edit test", { replace: true });
    const captionConfirm = captionTextarea.parent(".p-sidebar-info .v-list-item").find(".meta-inline-confirm");
    await t.click(captionConfirm);
    await t.expect(Selector(".p-sidebar-info .meta-caption").withText("Caption added in sidebar edit test").exists).ok();

    const plainTextFields = [
      { icon: "mdi-text-box-outline", value: "Testing sidebar edits" }, // Subject
      { icon: "mdi-palette", value: "Test Artist" },
      { icon: "mdi-copyright", value: "2024 Test" },
      { icon: "mdi-license", value: "Test-License-1.0" },
    ];
    for (const field of plainTextFields) {
      const input = await photoviewer.startInlineEditByIcon(field.icon);
      await t.typeText(input, field.value, { replace: true });
      await photoviewer.confirmInlineEditByIcon(field.icon);
      await t.expect(photoviewer.sidebarRow(field.icon).withText(field.value).exists).ok();
    }

    // The backend normalizes keywords (split on whitespace/commas,
    // lowercased, deduped), so we assert on a unique single-word token
    // that survives that pass.
    const keywordsInput = await photoviewer.startInlineEditBySection("Keywords");
    await t.typeText(keywordsInput, "sidebareditkw", { replace: true });
    await photoviewer.confirmInlineEditBySection("Keywords");
    await t.expect(Selector(".p-sidebar-info .meta-keywords").withText("sidebareditkw").exists).ok();

    const notesInput = await photoviewer.startInlineEditBySection("Notes");
    await t.typeText(notesInput, "SidebarNoteFromTest", { replace: true });
    await photoviewer.confirmInlineEditBySection("Notes");
    await t.expect(Selector(".p-sidebar-info .meta-notes").withText("SidebarNoteFromTest").exists).ok();
  }
);

test.meta("testID", "sidebar-edit-002").meta({ mode: "public" })("Common: Adds a label and an album inline and persists them to the photo", async (t) => {
  const uid = await photoviewer.openSidebarOnFirstPhoto();

  // Unique names per run avoid collisions with leftover fixtures from
  // earlier executions that didn't tear down cleanly.
  const stamp = Date.now();
  const labelTitle = `SidebarEditLabel-${stamp}`;
  const albumTitle = `SidebarEditAlbum-${stamp}`;

  await photoviewer.typeAndConfirmInlineChip("Labels", labelTitle);
  await photoviewer.typeAndConfirmInlineChip("Albums", albumTitle);

  await photoviewer.closePhotoViewer();

  await menu.openPage("labels");
  await label.openByTitle(labelTitle);
  await t.expect(Selector("div.is-photo").withAttribute("data-uid", uid).exists).ok();

  await menu.openPage("albums");
  await album.openByTitle(albumTitle);
  await t.expect(Selector("div.is-photo").withAttribute("data-uid", uid).exists).ok();
});

test.meta("testID", "sidebar-edit-003").meta({ mode: "public" })(
  "Common: Edits every taken-at, camera, and location field and confirms persistence",
  async (t) => {
    await photoviewer.openSidebarOnFirstPhoto();

    // Date / time dialog: change all five user-editable fields. Re-opening
    // the dialog after save is the most direct check that each field round-
    // tripped through the API; the sidebar only formats a subset of these
    // (year is always rendered, timezone only when set).
    await photoviewer.openSidebarDialog("takenAt");
    const yearInput = photoviewer.dateTimeDialog.find(".input-year input");
    const monthInput = photoviewer.dateTimeDialog.find(".input-month input");
    const dayInput = photoviewer.dateTimeDialog.find(".input-day input");
    const timeInput = photoviewer.dateTimeDialog.find(".input-local-time input");
    const timeZoneInput = photoviewer.dateTimeDialog.find(".input-timezone input");
    await t.typeText(yearInput, "2022", { replace: true }).pressKey("tab");
    await t.typeText(monthInput, "7", { replace: true }).pressKey("tab");
    await t.typeText(dayInput, "15", { replace: true }).pressKey("tab");
    await t.typeText(timeInput, "13:45:30", { replace: true }).pressKey("tab");
    await t.typeText(timeZoneInput, "UTC", { replace: true }).pressKey("enter");
    await t.click(photoviewer.dateTimeDialog.find(".action-confirm"));
    await t.expect(photoviewer.dateTimeDialog.visible).notOk();
    await t.expect(photoviewer.sidebarRow("mdi-calendar").withText("2022").exists).ok();

    // Re-open and verify each input was persisted. The autocompletes render
    // their selected value as the input's `value` attribute, so we read it
    // back rather than relying on locale-formatted display strings.
    await photoviewer.openSidebarDialog("takenAt");
    await t.expect(yearInput.value).eql("2022");
    await t.expect(monthInput.value).eql("7");
    await t.expect(dayInput.value).eql("15");
    await t.expect(timeInput.value).eql("13:45:30");
    await t.expect(timeZoneInput.value).eql("UTC");
    await t.click(photoviewer.dateTimeDialog.find(".action-cancel"));
    await t.expect(photoviewer.dateTimeDialog.visible).notOk();

    // Camera dialog: change every numeric field. Camera/Lens autocompletes
    // are intentionally skipped — their item lists depend on backend fixture
    // data, which varies across acceptance environments. The four numeric
    // fields cover the dialog's persistence path end-to-end.
    await photoviewer.openSidebarDialog("camera");
    const isoInput = photoviewer.cameraDialog.find(".input-iso input");
    const exposureInput = photoviewer.cameraDialog.find(".input-exposure input");
    const fNumberInput = photoviewer.cameraDialog.find(".input-fnumber input");
    const focalLengthInput = photoviewer.cameraDialog.find(".input-focal-length input");
    await t.typeText(isoInput, "6400", { replace: true });
    await t.typeText(exposureInput, "1/250", { replace: true });
    await t.typeText(fNumberInput, "1.8", { replace: true });
    await t.typeText(focalLengthInput, "35", { replace: true });
    await t.click(photoviewer.cameraDialog.find(".action-confirm"));
    await t.expect(photoviewer.cameraDialog.visible).notOk();

    // Re-open and verify every numeric field came back from the backend.
    await photoviewer.openSidebarDialog("camera");
    await t.expect(isoInput.value).eql("6400");
    await t.expect(exposureInput.value).eql("1/250");
    await t.expect(fNumberInput.value).eql("1.8");
    await t.expect(focalLengthInput.value).eql("35");
    await t.click(photoviewer.cameraDialog.find(".action-cancel"));
    await t.expect(photoviewer.cameraDialog.visible).notOk();

    // Location dialog: a raw coordinate string avoids hitting an external
    // reverse-geocoder, which keeps the test deterministic in offline
    // acceptance environments.
    await photoviewer.openSidebarDialog("location");
    const coordsInput = photoviewer.locationDialog.find(".input-coordinates input");
    await t.expect(coordsInput.visible).ok();
    await t.typeText(coordsInput, "52.5200, 13.4050", { replace: true }).pressKey("enter");
    await t.click(photoviewer.locationDialog.find(".action-confirm"));
    await t.expect(photoviewer.locationDialog.visible).notOk();
    await t.expect(Selector(".p-sidebar-info .p-map").exists).ok();

    // The coordinates row title is composed by Photo.getLatLng(); we assert
    // both halves so the test fails noisily if either axis fails to persist.
    const coordinatesRow = photoviewer.sidebarRow("mdi-map-marker").nextSibling(".v-list-item");
    await t.expect(coordinatesRow.getAttribute("title")).contains("52.52");
    await t.expect(coordinatesRow.getAttribute("title")).contains("13.405");
  }
);
