import { Selector } from "testcafe";
import testcafeconfig from "../testcafeconfig";
import Menu from "../../page-model/menu";
import Toolbar from "../../page-model/toolbar";
import ContextMenu from "../../page-model/context-menu";
import PhotoViewer from "../../page-model/photoviewer";
import Page from "../../page-model/page";
import PhotoEdit from "../../page-model/photo-edit";
import Album from "../../page-model/album";
import Settings from "../../page-model/settings";

fixture`Test settings`.page`${testcafeconfig.url}`;

const menu = new Menu();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const photoviewer = new PhotoViewer();
const page = new Page();
const photoedit = new PhotoEdit();
const album = new Album();
const settings = new Settings();

test.meta("testID", "settings-general-001")("General Settings", async (t) => {
  await toolbar.checkToolbarActionAvailability("upload", true);
  await t.expect(Selector(".nav-browse").innerText).contains("Search").navigateTo("/browse");
  await toolbar.search("photo:true stack:true");
  await photo.triggerHoverAction("nth", 0, "select");
  await contextmenu.checkContextMenuActionAvailability("download", true);
  await contextmenu.checkContextMenuActionAvailability("share", true);
  await contextmenu.checkContextMenuActionAvailability("edit", true);
  await contextmenu.checkContextMenuActionAvailability("private", true);
  await contextmenu.triggerContextMenuAction("edit", "", "");
  await t
    .click(Selector("#tab-files"))
    .expect(photoedit.downloadFile.nth(0).visible)
    .ok()
    .click(photoedit.toggleExpandFile.nth(1))
    .expect(photoedit.downloadFile.nth(1).visible)
    .ok()
    .expect(photoedit.deleteFile.visible)
    .ok()
    .click(photoedit.dialogClose);
  await contextmenu.clearSelection();
  await photoviewer.openPhotoViewer("nth", 0);
  await photoviewer.checkPhotoViewerActionAvailability("download", true);
  await photoviewer.triggerPhotoViewerAction("close");
  await t
    .expect(page.cardLocation.visible)
    .ok()
    .click(page.cardTitle.nth(0))
    .expect(Selector(".input-title input", { timeout: 8000 }).hasAttribute("disabled"))
    .notOk()
    .click(Selector("#tab-labels"))
    .expect(photoedit.addLabel.visible)
    .ok()
    .click(Selector("#tab-details"))
    .click(photoedit.dialogClose);
  await menu.openPage("library");
  await t
    .expect(Selector("#tab-library-import a").visible)
    .ok()
    .expect(Selector("#tab-library-logs a").visible)
    .ok();
  await menu.openPage("archive");
  await photo.triggerHoverAction("nth", 0, "select");
  await contextmenu.checkContextMenuActionAvailability("delete", true);
  await contextmenu.clearSelection();
  await menu.checkMenuItemAvailability("archive", true);
  await menu.checkMenuItemAvailability("review", true);
  await menu.checkMenuItemAvailability("originals", true);
  await menu.checkMenuItemAvailability("folders", true);
  await menu.checkMenuItemAvailability("moments", true);
  await menu.checkMenuItemAvailability("people", true);
  await menu.checkMenuItemAvailability("labels", true);
  await menu.checkMenuItemAvailability("places", true);
  await menu.checkMenuItemAvailability("private", true);
  await menu.openPage("settings");

  await t
    .click(settings.languageInput)
    .hover(Selector("div").withText("Deutsch").parent('div[role="listitem"]'))
    .click(Selector("div").withText("Deutsch").parent('div[role="listitem"]'))
    .click(settings.uploadCheckbox)
    .click(settings.downloadCheckbox)
    .click(settings.importCheckbox)
    .click(settings.archiveCheckbox)
    .click(settings.editCheckbox)
    .click(settings.filesCheckbox)
    .click(settings.peopleCheckbox)
    .click(settings.momentsCheckbox)
    .click(settings.labelsCheckbox)
    .click(settings.logsCheckbox)
    .click(settings.shareCheckbox)
    .click(settings.placesCheckbox)
    .click(settings.deleteCheckbox)
    .click(settings.privateCheckbox)
    .click(Selector("#tab-settings-library"))
    .click(settings.reviewCheckbox);
  await t.eval(() => location.reload());
  await t.navigateTo("/calendar");
  await toolbar.checkToolbarActionAvailability("upload", false);
  await album.checkHoverActionAvailability("nth", 2, "share", false);
  await album.triggerHoverAction("nth", 0, "select");
  await contextmenu.checkContextMenuActionAvailability("edit", true);
  await contextmenu.checkContextMenuActionAvailability("download", false);
  await contextmenu.checkContextMenuActionAvailability("share", false);
  await contextmenu.clearSelection();
  await album.openNthAlbum(0);
  await toolbar.checkToolbarActionAvailability("upload", false);
  await toolbar.checkToolbarActionAvailability("download", false);
  await toolbar.checkToolbarActionAvailability("share", false);
  await toolbar.checkToolbarActionAvailability("edit", true);
  await t.navigateTo("/folders");
  await toolbar.checkToolbarActionAvailability("upload", false);
  await album.checkHoverActionAvailability("nth", 0, "share", false);
  await album.triggerHoverAction("nth", 0, "select");
  await contextmenu.checkContextMenuActionAvailability("edit", true);
  await contextmenu.checkContextMenuActionAvailability("download", false);
  await contextmenu.checkContextMenuActionAvailability("share", false);
  await contextmenu.clearSelection();
  await album.openNthAlbum(0);
  await toolbar.checkToolbarActionAvailability("upload", false);
  await toolbar.checkToolbarActionAvailability("download", false);
  await toolbar.checkToolbarActionAvailability("share", false);
  await toolbar.checkToolbarActionAvailability("edit", true);
  await t.navigateTo("/albums");
  await toolbar.checkToolbarActionAvailability("upload", false);
  await album.checkHoverActionAvailability("nth", 0, "share", false);
  await album.triggerHoverAction("nth", 0, "select");
  await contextmenu.checkContextMenuActionAvailability("edit", true);
  await contextmenu.checkContextMenuActionAvailability("delete", true);
  await contextmenu.checkContextMenuActionAvailability("download", false);
  await contextmenu.checkContextMenuActionAvailability("share", false);
  await contextmenu.clearSelection();
  await album.openNthAlbum(0);
  await toolbar.checkToolbarActionAvailability("upload", false);
  await toolbar.checkToolbarActionAvailability("download", false);
  await toolbar.checkToolbarActionAvailability("share", false);
  await toolbar.checkToolbarActionAvailability("edit", true);
  await t.navigateTo("/browse");
  await toolbar.checkToolbarActionAvailability("upload", false);
  await t.expect(Selector(".nav-browse").innerText).contains("Suche");
  await toolbar.search("photo:true stack:true");
  await photo.triggerHoverAction("nth", 0, "select");
  await contextmenu.checkContextMenuActionAvailability("download", false);
  await contextmenu.checkContextMenuActionAvailability("share", false);
  await contextmenu.checkContextMenuActionAvailability("edit", false);
  await contextmenu.checkContextMenuActionAvailability("private", false);
  await contextmenu.checkContextMenuActionAvailability("archive", false);
  await contextmenu.clearSelection();
  await t
    .click(page.cardTitle.nth(0))
    .click(Selector("#tab-files"))
    .expect(photoedit.downloadFile.nth(0).visible)
    .notOk()
    .click(photoedit.toggleExpandFile.nth(1))
    .expect(photoedit.downloadFile.nth(1).visible)
    .notOk()
    .expect(photoedit.deleteFile.visible)
    .notOk()
    .click(photoedit.dialogClose);
  await toolbar.search("photo:true");
  await photoviewer.openPhotoViewer("nth", 0);
  await photoviewer.checkPhotoViewerActionAvailability("download", false);
  await photoviewer.checkPhotoViewerActionAvailability("edit", true);
  await photoviewer.triggerPhotoViewerAction("close");
  await t
    .expect(page.cardLocation.exists)
    .notOk()
    .click(page.cardTitle.nth(0))
    .expect(Selector(".input-title input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-latitude input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-timezone input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-country input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-description textarea").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-keywords textarea").hasAttribute("disabled"))
    .ok()
    .click(Selector("#tab-labels"))
    .expect(photoedit.addLabel.exists)
    .notOk()
    .click(photoedit.dialogClose);
  await menu.openPage("library");
  await t
    .expect(Selector("#tab-library-import a").exists)
    .notOk()
    .expect(Selector("#tab-library-logs a").exists)
    .notOk();
  await menu.checkMenuItemAvailability("archive", false);
  await menu.checkMenuItemAvailability("review", false);
  await menu.checkMenuItemAvailability("originals", false);
  await menu.checkMenuItemAvailability("folders", true);
  await menu.checkMenuItemAvailability("moments", false);
  await menu.checkMenuItemAvailability("people", false);
  await menu.checkMenuItemAvailability("labels", false);
  await menu.checkMenuItemAvailability("places", false);
  await menu.checkMenuItemAvailability("private", false);
  await menu.openPage("settings");
  await t
    .click(settings.languageInput)
    .hover(Selector("div").withText("English").parent('div[role="listitem"]'))
    .click(Selector("div").withText("English").parent('div[role="listitem"]'))
    .click(settings.uploadCheckbox)
    .click(settings.downloadCheckbox)
    .click(settings.importCheckbox)
    .click(settings.archiveCheckbox)
    .click(settings.editCheckbox)
    .click(settings.filesCheckbox)
    .click(settings.peopleCheckbox)
    .click(settings.momentsCheckbox)
    .click(settings.labelsCheckbox)
    .click(settings.logsCheckbox)
    .click(settings.shareCheckbox)
    .click(settings.placesCheckbox)
    .click(settings.privateCheckbox)
    .click(Selector("#tab-settings-library"))
    .click(settings.reviewCheckbox);
  await menu.openPage("archive");
  await photo.triggerHoverAction("nth", 0, "select");
  await contextmenu.checkContextMenuActionAvailability("restore", true);
  await contextmenu.checkContextMenuActionAvailability("delete", false);
  await contextmenu.clearSelection();
  await menu.openPage("settings");
  await t.click(settings.deleteCheckbox);
});
