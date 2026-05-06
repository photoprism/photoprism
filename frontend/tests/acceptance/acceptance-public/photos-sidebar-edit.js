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

    // The backend lower-cases and dedupes keywords; assert on a single
    // lowercase token that survives that pass.
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

    const dateTimeDialog = photoviewer.dateTimeDialog;
    const cameraDialog = photoviewer.cameraDialog;
    const locationDialog = photoviewer.locationDialog;
    const optionWith = (text) => Selector('div[role="option"]').withText(text);

    // Clicking a rendered option is racy on fast systems where the list
    // re-mounts between visibility and click; keyboard nav is stable.
    const pickAutocomplete = async (input, value) => {
      await t.typeText(input, value, { replace: true }).pressKey("down enter");
    };
    const pickFromSelect = async (field, value) => {
      await t.click(field).click(optionWith(value));
    };

    // Other tests in this suite edit the same photo; snapshot now and
    // restore at the end so leftover timezone/values don't leak.
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

    // For UTC photos formatTime() drops the zone abbreviation, so "UTC"
    // never appears in the sidebar text — verified via the dialog below.
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
    await t.expect(locationDialog.coordinates.visible).ok();
    await t.typeText(locationDialog.coordinates, "52.5200, 13.4050", { replace: true }).pressKey("enter");
    await t.click(locationDialog.confirm);
    await t.expect(locationDialog.root.visible).notOk();
    await t.expect(Selector(".p-sidebar-info .p-map").exists).ok();

    // Substring match — the optional " · <altitude> m" suffix can follow.
    const coordinatesRow = photoviewer.sidebarRow("mdi-map-marker").nextSibling(".v-list-item");
    await t.expect(coordinatesRow.withText("52.52000").exists).ok();
    await t.expect(coordinatesRow.withText("13.40500").exists).ok();

    // Skip empty fields: typeText("") is a no-op and v-select has no clear.
    await photoviewer.openSidebarDialog("takenAt");
    if (initialYear) await pickAutocomplete(dateTimeDialog.year, initialYear);
    if (initialMonth) await pickAutocomplete(dateTimeDialog.month, initialMonth);
    if (initialDay) await pickAutocomplete(dateTimeDialog.day, initialDay);
    if (initialLocalTime) {
      await t.typeText(dateTimeDialog.localTime, initialLocalTime, { replace: true }).pressKey("tab");
    }
    if (initialTimezone) await pickAutocomplete(dateTimeDialog.timezone, initialTimezone);
    await t.click(dateTimeDialog.confirm);
    await t.expect(dateTimeDialog.root.visible).notOk();

    await photoviewer.openSidebarDialog("camera");
    if (initialCamera) await pickFromSelect(cameraDialog.camera, initialCamera);
    if (initialLens) await pickFromSelect(cameraDialog.lens, initialLens);
    if (initialIso) await t.typeText(cameraDialog.iso, initialIso, { replace: true });
    if (initialExposure) await t.typeText(cameraDialog.exposure, initialExposure, { replace: true });
    if (initialFnumber) await t.typeText(cameraDialog.fnumber, initialFnumber, { replace: true });
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
