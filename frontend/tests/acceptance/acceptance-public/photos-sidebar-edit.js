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

test.meta("testID", "sidebar-edit-003").meta({ mode: "public" })("Common: Edits taken-at, camera, and location via the sidebar dialogs", async (t) => {
  await photoviewer.openSidebarOnFirstPhoto();

  // Asserting on the year alone is enough to catch the "dialog confirms
  // but nothing persists" regression that vitest missed.
  await photoviewer.openSidebarDialog("takenAt");
  const yearInput = photoviewer.dateTimeDialog.find(".input-year input");
  await t.typeText(yearInput, "2022", { replace: true }).pressKey("enter");
  await t.click(photoviewer.dateTimeDialog.find(".action-confirm"));
  await t.expect(photoviewer.dateTimeDialog.visible).notOk();
  await t.expect(photoviewer.sidebarRow("mdi-calendar").withText("2022").exists).ok();

  await photoviewer.openSidebarDialog("camera");
  const isoInput = photoviewer.cameraDialog.find(".input-iso input");
  await t.typeText(isoInput, "6400", { replace: true });
  await t.click(photoviewer.cameraDialog.find(".action-confirm"));
  await t.expect(photoviewer.cameraDialog.visible).notOk();

  // A raw coordinate string avoids hitting an external reverse-geocoder,
  // which keeps the test deterministic in offline acceptance environments.
  await photoviewer.openSidebarDialog("location");
  const coordsInput = photoviewer.locationDialog.find(".input-coordinates input");
  await t.expect(coordsInput.visible).ok();
  await t.typeText(coordsInput, "52.5200, 13.4050", { replace: true }).pressKey("enter");
  await t.click(photoviewer.locationDialog.find(".action-confirm"));
  await t.expect(photoviewer.locationDialog.visible).notOk();
  await t.expect(Selector(".p-sidebar-info .p-map").exists).ok();
});
