import { Selector } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";
import { ClientFunction } from "testcafe";
import Menu from "../page-model/menu";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import Photo from "../page-model/photo";
import PhotoViewer from "../page-model/photoviewer";
import Page from "../page-model/page";
import PhotoEdit from "../page-model/photo-edit";
import Notifies from "../page-model/notifications";

fixture`Test batch edit`.page`${testcafeconfig.url}`;

const menu = new Menu();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const photo = new Photo();
const photoviewer = new PhotoViewer();
const page = new Page();
const photoedit = new PhotoEdit();
const notifies = new Notifies();

test.meta("testID", "batch-001").meta({ mode: "public" })("Common: Test batch dialog selection and lightbox", async (t) => {
  await menu.openPage("browse");
  await toolbar.search("cat");
  await photo.toggleSelectNthPhoto(0, "image");
  await toolbar.search("canada");
  await photo.toggleSelectNthPhoto(0, "image");
  await toolbar.search("album:holiday");
  await photo.toggleSelectNthPhoto(0, "image");
  await toolbar.search("archived:true");
  await photo.toggleSelectNthPhoto(0, "image");
  await contextmenu.checkContextMenuCount("4");
  await contextmenu.triggerContextMenuAction("edit");
  // verify that archived photo is excluded
  await t.expect(photoedit.batchDialog.visible).ok().expect(photoedit.batchDialogTitle.innerText).contains("(3)");
  await t.click(photoedit.batchDialogPreview.nth(0));
  // verify that lightbox actions are limited
  await t.expect(Selector("div.p-lightbox__pswp").visible).ok();
  await photoviewer.checkPhotoViewerActionAvailability("edit-button", false);
  await photoviewer.checkPhotoViewerActionAvailability("select-toggle", false);
  await photoviewer.checkPhotoViewerActionAvailability("fullscreen-toggle", true);
  await photoviewer.checkPhotoViewerActionAvailability("favorite-toggle", true);
  await photoviewer.checkPhotoViewerActionAvailability("info-button", true);
  await photoviewer.triggerPhotoViewerAction("close-button");
  await t.expect(photoedit.batchDialog.visible).ok();
  await t.expect(Selector("div.p-lightbox__pswp").visible).notOk();
  // verify count is updated
  await t.click(photoedit.batchToggleAllCheckbox);
  await t.expect(photoedit.batchDialog.visible).ok().expect(photoedit.batchDialogTitle.innerText).contains("(0)");
  await t.click(photoedit.batchToggleSelectCheckbox.nth(0));
  await t.expect(photoedit.batchDialog.visible).ok().expect(photoedit.batchDialogTitle.innerText).contains("(1)");
  await t.click(photoedit.batchToggleAllCheckbox);
  await t.expect(photoedit.batchDialog.visible).ok().expect(photoedit.batchDialogTitle.innerText).contains("(3)");
  await t.click(photoedit.batchDialogToolbarCloseAction);
  await t.expect(photoedit.batchDialog.visible).notOk();
  await contextmenu.clearSelection();
});

test.meta("testID", "batch-002").meta({ mode: "public" })("Common: Test batch dialog cannot be closed with unsaved changes", async (t) => {
  await menu.openPage("browse");
  await toolbar.search("cat");
  await photo.toggleSelectNthPhoto(0, "image");
  await toolbar.search("canada");
  await photo.toggleSelectNthPhoto(0, "image");
  await toolbar.search("album:holiday");
  await photo.toggleSelectNthPhoto(0, "image");
  await toolbar.search("archived:true");
  await photo.toggleSelectNthPhoto(0, "image");
  await contextmenu.checkContextMenuCount("4");
  await contextmenu.triggerContextMenuAction("edit");
  await t.expect(photoedit.batchDialogCloseAction.innerText).contains("CLOSE");
  await t.expect(photoedit.batchDialogApplyAction.hasAttribute("disabled")).ok();
  await t.expect(photoedit.batchDialogCloseAction.hasAttribute("disabled")).notOk();
  await t.expect(photoedit.title.getAttribute("placeholder")).eql("<mixed>");
  await t.click(photoedit.batchDialogToolbarCloseAction);
  await t.expect(photoedit.batchDialog.visible).notOk();
  await contextmenu.triggerContextMenuAction("edit");
  await t.expect(photoedit.batchDialog.visible).ok();
  await t.typeText(photoedit.title, "Batch Title");
  await t.expect(photoedit.batchDialogCloseAction.innerText).contains("DISCARD");
  await t.expect(photoedit.batchDialogApplyAction.hasAttribute("disabled")).notOk();
  await t.expect(photoedit.batchDialogCloseAction.hasAttribute("disabled")).notOk();
  await t.click(photoedit.batchDialogToolbarCloseAction);
  await t.expect(photoedit.batchDialog.visible).ok();
  await t.click(photoedit.batchDialogCloseAction);
  await t.expect(photoedit.batchDialog.visible).notOk();
  await contextmenu.clearSelection();
});

test.meta("testID", "batch-003").meta({ mode: "public" })("Common: Test batch dialog form functionality", async (t) => {
  await menu.openPage("browse");
  await toolbar.search("cat");
  await photo.toggleSelectNthPhoto(0, "image");
  await toolbar.search("canada");
  await photo.toggleSelectNthPhoto(0, "image");
  await toolbar.search("album:holiday");
  await photo.toggleSelectNthPhoto(0, "image");
  await toolbar.search("archived:true");
  await photo.toggleSelectNthPhoto(0, "image");
  await contextmenu.checkContextMenuCount("4");
  await contextmenu.triggerContextMenuAction("edit");
  await t.expect(photoedit.dayValue.innerText).eql("<mixed>");
  await t.expect(photoedit.countryValue.innerText).eql("<mixed>");
  await t.expect(photoedit.title.getAttribute("placeholder")).eql("<mixed>");
  await t.typeText(photoedit.title, "Batch Title");
  await t.expect(photoedit.title.value).eql("Batch Title");
  await t.click(Selector(".input-title i.mdi-undo"));
  await t.expect(photoedit.title.getAttribute("placeholder")).eql("<mixed>");
  await t.click(Selector(".input-title i.mdi-close-circle"));
  await t.expect(photoedit.title.value).eql("");
  await t.expect(photoedit.title.getAttribute("placeholder")).eql("<deleted>");
  await t.expect(Selector("div.chip").withText("Cat").visible).ok();
  await t.expect(Selector("div.chip").withText("Holiday").visible).ok();
  await t.expect(photoedit.country.hasAttribute("readonly")).notOk();
  await t.click(photoedit.locationAction);
  await t.typeText(photoedit.locationSearch, "Brandenburger Tor Berlin").wait(5000).pressKey("enter");
  await t.click(photoedit.locationConfirm);
  await t.expect(photoedit.country.hasAttribute("readonly")).ok();
  await t.expect(photoedit.countryValue.innerText).eql("Germany");
  await t.click(Selector(".input-labels input"));

  await t.expect(page.selectOption.withText("People").visible).ok().expect(page.selectOption.withText("Cat").visible).ok();

  await t.typeText(Selector(".input-labels input"), "P", { replace: true });

  await t
    .expect(page.selectOption.withText("Cat").visible)
    .notOk()
    .expect(page.selectOption.withText("Portrait").visible)
    .ok()
    .expect(page.selectOption.withText("P").visible)
    .ok();

  await t.click(Selector(".input-albums input"));

  await t.expect(page.selectOption.withText("Holiday").visible).ok().expect(page.selectOption.withText("Christmas").visible).ok();

  await t.typeText(Selector(".input-albums input"), "C", { replace: true });

  await t
    .expect(page.selectOption.withText("Holiday").visible)
    .notOk()
    .expect(page.selectOption.withText("Christmas").visible)
    .ok()
    .expect(page.selectOption.withText("C").visible)
    .ok();
  await t.click(photoedit.batchDialogCloseAction);
  await t.expect(photoedit.batchDialog.visible).notOk();
  await contextmenu.clearSelection();
});
