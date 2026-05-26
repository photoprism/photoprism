import { Selector } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";
import PhotoViewer from "../page-model/photoviewer";
import Menu from "../page-model/menu";
import Label from "../page-model/label";
import Album from "../page-model/album";

fixture`Test lightbox sidebar inline editing`.page`${testcafeconfig.url}`;

const photoviewer = new PhotoViewer();
const menu = new Menu();
const label = new Label();
const album = new Album();

test.meta("testID", "sidebar-edit-001").meta({ mode: "public" })(
  "Common: Edits title, caption, keywords, notes, and plain-text inline fields from the sidebar",
  async (t) => {
    await photoviewer.openSidebarOnFirstPhoto();

    const titleInput = Selector(".p-lightbox-sidebar .meta-inline-title input", { timeout: 15000 });
    await photoviewer.startInlineEditOrAdd("meta-title", "Title");
    await t.expect(titleInput.visible).ok();
    await t.typeText(titleInput, "Sidebar Edit Title", { replace: true }).pressKey("enter");
    await t.expect(Selector(".p-lightbox-sidebar .meta-title").withText("Sidebar Edit Title").exists).ok();

    // Caption commits on blur (Enter inserts a newline).
    const captionTextarea = Selector(".p-lightbox-sidebar .meta-inline-caption textarea", { timeout: 15000 });
    await photoviewer.startInlineEditOrAdd("meta-caption", "Caption");
    await t.expect(captionTextarea.visible).ok();
    await t.typeText(captionTextarea, "Caption added in sidebar edit test", { replace: true }).pressKey("tab");
    await t.expect(Selector(".p-lightbox-sidebar .meta-caption").withText("Caption added in sidebar edit test").exists).ok();

    // Keyword value is a single lowercase token to survive the backend's lowercase + dedup.
    const plainTextFields = [
      { key: "subject",   value: "Testing sidebar edits", commitKey: "enter" },
      { key: "artist",    value: "Test Artist",           commitKey: "enter" },
      { key: "copyright", value: "2024 Test",             commitKey: "enter" },
      { key: "license",   value: "Test-License-1.0",      commitKey: "enter" },
      { key: "keywords",  value: "sidebareditkw",         commitKey: "enter" },
      { key: "notes",     value: "SidebarNoteFromTest",   commitKey: "tab" },
    ];
    for (const field of plainTextFields) {
      await photoviewer.editTextFieldByKey(field.key, field.value, field.commitKey);
    }
  }
);

test.meta("testID", "sidebar-edit-002").meta({ mode: "public" })("Common: Adds a label and an album inline and persists them to the photo", async (t) => {
  const uid = await photoviewer.openSidebarOnFirstPhoto();

  // Date-stamp names so reruns don't collide with leftovers from previous failed runs.
  const stamp = Date.now();
  const labelTitle = `SidebarEditLabel-${stamp}`;
  const albumTitle = `SidebarEditAlbum-${stamp}`;

  await photoviewer.typeAndConfirmInlineChip("Labels", labelTitle);
  await photoviewer.typeAndConfirmInlineChip("Albums", albumTitle);

  await photoviewer.triggerPhotoViewerAction("close-button");

  await menu.openPage("labels");
  await label.openByTitle(labelTitle);
  await t.expect(Selector("div.is-photo").withAttribute("data-uid", uid).exists).ok();

  await menu.openPage("albums");
  await album.openByTitle(albumTitle);
  await t.expect(Selector("div.is-photo").withAttribute("data-uid", uid).exists).ok();
});

test.meta("testID", "sidebar-edit-004").meta({ mode: "public" })(
  "Common: Removes a label and an album inline, undoes before save, and keeps them on the photo",
  async (t) => {
    const uid = await photoviewer.openSidebarOnFirstPhoto();

    const stamp = Date.now();
    const labelTitle = `SidebarEditLabelUndo-${stamp}`;
    const albumTitle = `SidebarEditAlbumUndo-${stamp}`;

    await photoviewer.typeAndConfirmInlineChip("Labels", labelTitle);
    await photoviewer.typeAndConfirmInlineChip("Albums", albumTitle);

    const labelChip = photoviewer.chipByTitle("Labels", labelTitle);
    const albumChip = photoviewer.chipByTitle("Albums", albumTitle);

    await photoviewer.removeInlineChip("Labels", labelTitle);
    await t.expect(labelChip.exists).notOk();
    await photoviewer.removeInlineChip("Albums", albumTitle);
    await t.expect(albumChip.exists).notOk();

    await photoviewer.undoChipRemovals("Labels");
    await t.expect(labelChip.exists).ok();
    await photoviewer.undoChipRemovals("Albums");
    await t.expect(albumChip.exists).ok();

    // Close + reopen rehydrates the sidebar from the API; both chips must
    // still be there because Undo never reached the backend.
    await photoviewer.triggerPhotoViewerAction("close-button");
    await photoviewer.openSidebarOnPhoto(uid);

    await t.expect(photoviewer.chipByTitle("Labels", labelTitle).exists).ok();
    await t.expect(photoviewer.chipByTitle("Albums", albumTitle).exists).ok();
  }
);

test.meta("testID", "sidebar-edit-005").meta({ mode: "public" })(
  "Common: Removes a label and an album inline, saves the removal, and detaches them from the photo",
  async (t) => {
    const uid = await photoviewer.openSidebarOnFirstPhoto();

    const stamp = Date.now();
    const labelTitle = `SidebarEditLabelSave-${stamp}`;
    const albumTitle = `SidebarEditAlbumSave-${stamp}`;

    await photoviewer.typeAndConfirmInlineChip("Labels", labelTitle);
    await photoviewer.typeAndConfirmInlineChip("Albums", albumTitle);

    await photoviewer.removeInlineChip("Labels", labelTitle);
    await photoviewer.removeInlineChip("Albums", albumTitle);

    await photoviewer.confirmChipRemovals("Labels");
    await photoviewer.confirmChipRemovals("Albums");

    await t.expect(photoviewer.chipByTitle("Labels", labelTitle).exists).notOk();
    await t.expect(photoviewer.chipByTitle("Albums", albumTitle).exists).notOk();

    // Close + reopen confirms the removal persisted past the round-trip;
    // the saved photo no longer references either chip.
    await photoviewer.triggerPhotoViewerAction("close-button");
    await photoviewer.openSidebarOnPhoto(uid);

    await t.expect(photoviewer.chipByTitle("Labels", labelTitle).exists).notOk();
    await t.expect(photoviewer.chipByTitle("Albums", albumTitle).exists).notOk();
  }
);

test.meta("testID", "sidebar-edit-003").meta({ mode: "public" })(
  "Common: Edits every taken-at, camera, and location field and confirms persistence",
  async (t) => {
    await photoviewer.openSidebarOnFirstPhoto();

    const dateTimeDialog = photoviewer.dateTimeDialog;
    const cameraDialog = photoviewer.cameraDialog;
    const locationDialog = photoviewer.locationDialog;
    const optionWith = (text) => Selector('div[role="option"]').withText(text);

    // Keyboard nav is more stable than click on autocomplete options
    // (the list re-mounts between visibility and click on fast systems).
    const pickAutocomplete = async (input, value) => {
      await t.typeText(input, value, { replace: true }).pressKey("down enter");
    };
    const pickFromSelect = async (field, value) => {
      await t.click(field).click(optionWith(value));
    };

    // Snapshot the initial values so the test can restore them on exit and other
    // tests in this suite see a clean photo.
    await photoviewer.openSidebarDialog("takenAt");
    const initialYear = await dateTimeDialog.yearValue.innerText;
    const initialMonth = await dateTimeDialog.monthValue.innerText;
    const initialDay = await dateTimeDialog.dayValue.innerText;
    const initialLocalTime = await dateTimeDialog.localTime.value;
    const initialTimezone = await dateTimeDialog.timezoneValue.innerText;
    await t.click(dateTimeDialog.cancel);
    await t.expect(dateTimeDialog.root.visible).notOk();

    await photoviewer.openSidebarDialog("camera");
    const initialCamera = await cameraDialog.cameraValue.innerText;
    const initialLens = await cameraDialog.lensValue.innerText;
    const initialIso = await cameraDialog.iso.value;
    const initialExposure = await cameraDialog.exposure.value;
    const initialFnumber = await cameraDialog.fnumber.value;
    const initialFocalLength = await cameraDialog.focalLength.value;
    await t.click(cameraDialog.cancel);
    await t.expect(cameraDialog.root.visible).notOk();

    await photoviewer.openSidebarDialog("location");
    const initialCoordinates = await locationDialog.coordinates.value;
    await t.click(locationDialog.cancel);
    await t.expect(locationDialog.root.visible).notOk();

    await photoviewer.openSidebarDialog("takenAt");
    await pickAutocomplete(dateTimeDialog.year, "2022");
    await pickAutocomplete(dateTimeDialog.month, "07");
    await pickAutocomplete(dateTimeDialog.day, "15");
    await t.typeText(dateTimeDialog.localTime, "13:45:30", { replace: true }).pressKey("tab");
    await pickAutocomplete(dateTimeDialog.timezone, "UTC");
    await t.click(dateTimeDialog.confirm);
    await t.expect(dateTimeDialog.root.visible).notOk();

    // formatTime() drops the zone abbreviation on UTC photos, so "UTC" never appears
    // in the sidebar text — it's only checked via the dialog below.
    const calendarRow = photoviewer.sidebarRow("mdi-calendar");
    await t.expect(calendarRow.withText("2022").exists).ok();
    await t.expect(calendarRow.withText("Jul").exists).ok();
    await t.expect(calendarRow.withText("15").exists).ok();

    await photoviewer.openSidebarDialog("takenAt");
    await t.expect(dateTimeDialog.yearValue.innerText).eql("2022");
    await t.expect(dateTimeDialog.monthValue.innerText).eql("07");
    await t.expect(dateTimeDialog.dayValue.innerText).eql("15");
    await t.expect(dateTimeDialog.localTime.value).eql("13:45:30");
    await t.expect(dateTimeDialog.timezoneValue.innerText).eql("UTC");
    await t.click(dateTimeDialog.cancel);
    await t.expect(dateTimeDialog.root.visible).notOk();

    const cameraName = "Canon EOS M10";
    const lensName = "EF-M15-45mm f/3.5-6.3 IS STM";
    await photoviewer.openSidebarDialog("camera");
    await pickFromSelect(cameraDialog.camera, cameraName);
    await pickFromSelect(cameraDialog.lens, lensName);
    await t.typeText(cameraDialog.iso, "6400", { replace: true });
    await t.typeText(cameraDialog.exposure, "1/250", { replace: true });
    await t.typeText(cameraDialog.fnumber, "1.8", { replace: true });
    await t.typeText(cameraDialog.focalLength, "35", { replace: true });
    await t.click(cameraDialog.confirm);
    await t.expect(cameraDialog.root.visible).notOk();

    const cameraRow = photoviewer.sidebarRow("mdi-camera");
    await t.expect(cameraRow.withText(cameraName).exists).ok();
    await t.expect(cameraRow.withText("ISO 6400").exists).ok();
    await t.expect(cameraRow.withText("1/250").exists).ok();

    await photoviewer.openSidebarDialog("camera");
    await t.expect(cameraDialog.cameraValue.innerText).eql(cameraName);
    await t.expect(cameraDialog.lensValue.innerText).eql(lensName);
    await t.expect(cameraDialog.iso.value).eql("6400");
    await t.expect(cameraDialog.exposure.value).eql("1/250");
    await t.expect(cameraDialog.fnumber.value).eql("1.8");
    await t.expect(cameraDialog.focalLength.value).eql("35");
    await t.click(cameraDialog.cancel);
    await t.expect(cameraDialog.root.visible).notOk();

    // Raw coordinates avoid hitting the external reverse-geocoder.
    await photoviewer.openSidebarDialog("location");
    // (search field skipped on purpose to keep the test offline)
    await t.expect(locationDialog.coordinates.visible).ok();
    await t.typeText(locationDialog.coordinates, "52.5200, 13.4050", { replace: true }).pressKey("enter");
    await t.click(locationDialog.confirm);
    await t.expect(locationDialog.root.visible).notOk();
    await t.expect(Selector(".p-lightbox-sidebar .p-map").exists).ok();

    // Thumb.getLatLngShort() formats as 4-digit decimals with °N / °E suffix —
    // assert the suffix too so a regression that drops it doesn't silently pass.
    const locationRow = photoviewer.sidebarRow("mdi-map-marker");
    await t.expect(locationRow.withText("52.5200°N").visible).ok();
    await t.expect(locationRow.withText("13.4050°E").visible).ok();

    // Restore the snapshotted initial values. Skip empty ones — typeText("") is
    // a no-op and v-select has no clear.
    await photoviewer.openSidebarDialog("takenAt");
    if (initialYear) {
      await pickAutocomplete(dateTimeDialog.year, initialYear);
    }
    if (initialMonth) {
      await pickAutocomplete(dateTimeDialog.month, initialMonth);
    }
    if (initialDay) {
      await pickAutocomplete(dateTimeDialog.day, initialDay);
    }
    if (initialLocalTime) {
      await t.typeText(dateTimeDialog.localTime, initialLocalTime, { replace: true }).pressKey("tab");
    }
    if (initialTimezone) {
      await pickAutocomplete(dateTimeDialog.timezone, initialTimezone);
    }
    await t.click(dateTimeDialog.confirm);
    await t.expect(dateTimeDialog.root.visible).notOk();

    await photoviewer.openSidebarDialog("camera");
    if (initialCamera) {
      await pickFromSelect(cameraDialog.camera, initialCamera);
    }
    if (initialLens) {
      await pickFromSelect(cameraDialog.lens, initialLens);
    }
    if (initialIso) {
      await t.typeText(cameraDialog.iso, initialIso, { replace: true });
    }
    if (initialExposure) {
      await t.typeText(cameraDialog.exposure, initialExposure, { replace: true });
    }
    if (initialFnumber) {
      await t.typeText(cameraDialog.fnumber, initialFnumber, { replace: true });
    }
    if (initialFocalLength) {
      await t.typeText(cameraDialog.focalLength, initialFocalLength, { replace: true });
    }
    await t.click(cameraDialog.confirm);
    await t.expect(cameraDialog.root.visible).notOk();

    if (initialCoordinates) {
      await photoviewer.openSidebarDialog("location");
      await t.typeText(locationDialog.coordinates, initialCoordinates, { replace: true }).pressKey("enter");
      await t.click(locationDialog.confirm);
      await t.expect(locationDialog.root.visible).notOk();
    }
  }
);
