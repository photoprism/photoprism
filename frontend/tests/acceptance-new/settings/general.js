import { Selector } from "testcafe";
import testcafeconfig from "../testcafeconfig";
import Menu from "../../page-model/menu";
import Toolbar from "../../page-model/toolbar";
import ContextMenu from "../../page-model/context-menu";
import PhotoViewer from "../../page-model/photoviewer";
import Page from "../../page-model/page";
import Photo from "../../page-model/photo";
import PhotoEdit from "../../page-model/photo-edit";
import Album from "../../page-model/album";
import Settings from "../../page-model/settings";
import Library from "../../page-model/library";

fixture`Test settings`.page`${testcafeconfig.url}`;

const menu = new Menu();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const photoviewer = new PhotoViewer();
const page = new Page();
const photo = new Photo();
const photoedit = new PhotoEdit();
const album = new Album();
const settings = new Settings();
const library = new Library();

test.meta("testID", "settings-general-001").meta({ type: "smoke" })("Disable delete", async (t) => {
  await menu.openPage("archive");
  await photo.triggerHoverAction("nth", 0, "select");
  await contextmenu.checkContextMenuActionAvailability("delete", true);
  await contextmenu.clearSelection();
  await menu.openPage("settings");
  await t.click(settings.deleteCheckbox);
  await menu.openPage("archive");
  await photo.triggerHoverAction("nth", 0, "select");

  await contextmenu.checkContextMenuActionAvailability("restore", true);
  await contextmenu.checkContextMenuActionAvailability("delete", false);
  await contextmenu.clearSelection();

  await menu.openPage("browse");
  await toolbar.search("stack:true");
  await photo.triggerHoverAction("nth", 0, "select");
  await contextmenu.triggerContextMenuAction("edit", "");
  await t.click(photoedit.filesTab);
  await t.click(photoedit.toggleExpandFile.nth(1));

  await t.expect(photoedit.deleteFile.visible).notOk();

  await t.click(photoedit.dialogClose);
  await contextmenu.clearSelection();
  await menu.openPage("settings");
  await t.click(settings.deleteCheckbox);
});

test.meta("testID", "settings-general-002").meta({ type: "smoke" })(
  "Change language",
  async (t) => {
    await t.expect(Selector(".nav-browse").innerText).contains("Search");

    await menu.openPage("settings");
    await t
      .click(settings.languageInput)
      .hover(Selector("div").withText("Deutsch").parent('div[role="listitem"]'))
      .click(Selector("div").withText("Deutsch").parent('div[role="listitem"]'));
    await t.eval(() => location.reload());

    await t.expect(Selector(".nav-browse").innerText).contains("Suche");

    await t
      .click(settings.languageInput)
      .hover(Selector("div").withText("English").parent('div[role="listitem"]'))
      .click(Selector("div").withText("English").parent('div[role="listitem"]'));

    await t.expect(Selector(".nav-browse").innerText).contains("Search");
  }
);

test.meta("testID", "settings-general-003").meta({ type: "smoke" })(
  "Disable pages: import, originals, logs, moments, places, library",
  async (t) => {
    await t.expect(page.cardLocation.exists).ok();

    await menu.openPage("library");

    await t
      .expect(library.importTab.visible)
      .ok()
      .expect(library.logsTab.visible)
      .ok()
      .expect(library.indexTab.visible)
      .ok();
    await menu.checkMenuItemAvailability("originals", true);
    await menu.checkMenuItemAvailability("folders", true);
    await menu.checkMenuItemAvailability("moments", true);
    await menu.checkMenuItemAvailability("places", true);
    await menu.checkMenuItemAvailability("library", true);

    await menu.openPage("settings");
    await t
      .click(settings.importCheckbox)
      .click(settings.filesCheckbox)
      .click(settings.momentsCheckbox)
      .click(settings.logsCheckbox)
      .click(settings.placesCheckbox);
    await t.eval(() => location.reload());

    await menu.openPage("browse");

    await t.expect(page.cardLocation.exists).notOk();

    await menu.openPage("library");

    await t
      .expect(library.importTab.visible)
      .notOk()
      .expect(library.logsTab.visible)
      .notOk()
      .expect(library.indexTab.visible)
      .ok();
    await menu.checkMenuItemAvailability("originals", false);
    await menu.checkMenuItemAvailability("folders", true);
    await menu.checkMenuItemAvailability("moments", false);
    await menu.checkMenuItemAvailability("places", false);
    await menu.checkMenuItemAvailability("library", true);

    await menu.openPage("settings");
    await t
      .click(settings.importCheckbox)
      .click(settings.filesCheckbox)
      .click(settings.momentsCheckbox)
      .click(settings.logsCheckbox)
      .click(settings.placesCheckbox)
      .click(settings.libraryCheckbox);

    await menu.checkMenuItemAvailability("originals", false);
    await menu.checkMenuItemAvailability("folders", true);
    await menu.checkMenuItemAvailability("moments", true);
    await menu.checkMenuItemAvailability("places", true);
    await menu.checkMenuItemAvailability("library", false);

    await t.click(settings.libraryCheckbox);

    await menu.checkMenuItemAvailability("originals", true);
    await menu.checkMenuItemAvailability("library", true);
  }
);

test.meta("testID", "settings-general-004").meta({ type: "smoke" })(
  "Disable people and labels",
  async (t) => {
    await t.click(page.cardTitle.nth(0));
    await t.click(photoedit.labelsTab);

    await t.expect(photoedit.addLabel.visible).ok();

    await t.click(photoedit.peopleTab);

    await t.expect(Selector("div.p-faces").visible).ok();

    await t.click(photoedit.dialogClose);
    await menu.checkMenuItemAvailability("people", true);
    await menu.checkMenuItemAvailability("labels", true);
    await menu.openPage("settings");
    await t.click(settings.peopleCheckbox).click(settings.labelsCheckbox);
    await t.eval(() => location.reload());
    await menu.openPage("browse");
    await t.click(page.cardTitle.nth(0));
    await t.click(photoedit.labelsTab);

    await t.expect(photoedit.addLabel.exists).notOk();

    await t.click(photoedit.peopleTab);

    await t.expect(Selector("div.p-faces ").exists).notOk();

    await t.click(photoedit.dialogClose);

    await menu.checkMenuItemAvailability("people", false);
    await menu.checkMenuItemAvailability("labels", false);

    await menu.openPage("settings");
    await t.click(settings.peopleCheckbox).click(settings.labelsCheckbox);

    await menu.checkMenuItemAvailability("people", true);
    await menu.checkMenuItemAvailability("labels", true);
  }
);

test.meta("testID", "settings-general-005").meta({ type: "smoke" })(
  "Disable private, archive and quality filter",
  async (t) => {
    await menu.checkMenuItemAvailability("private", true);
    await menu.checkMenuItemAvailability("archive", true);
    await menu.checkMenuItemAvailability("review", true);

    await toolbar.search("photo:true stack:true");
    await photo.triggerHoverAction("nth", 0, "select");

    await contextmenu.checkContextMenuActionAvailability("archive", true);
    await contextmenu.checkContextMenuActionAvailability("private", true);

    await contextmenu.triggerContextMenuAction("edit", "");
    await t.click(photoedit.infoTab);

    await photoedit.checkFieldDisabled(photoedit.privateInput, false);

    await t.click(photoedit.dialogClose);
    await contextmenu.clearSelection();
    await toolbar.search("Viewpoint / Mexico / 2017");

    await photo.checkPhotoVisibility("pqmxlr7188hz4bih", false);

    await toolbar.search("Truck / Vancouver / 2019");

    await photo.checkPhotoVisibility("pqmxlr0kg161o9ek", false);

    await toolbar.search("Archive / 2020");

    await photo.checkPhotoVisibility("pqnah1k2frui6p63", false);

    await menu.openPage("settings");
    await t
      .click(settings.archiveCheckbox)
      .click(settings.privateCheckbox)
      .click(Selector(settings.libraryTab))
      .click(settings.reviewCheckbox);
    await t.eval(() => location.reload());

    await menu.checkMenuItemAvailability("private", false);
    await menu.checkMenuItemAvailability("archive", false);
    await menu.checkMenuItemAvailability("review", false);

    await menu.openPage("browse");

    await toolbar.search("photo:true");
    await photo.triggerHoverAction("nth", 0, "select");

    await contextmenu.checkContextMenuActionAvailability("archive", false);
    await contextmenu.checkContextMenuActionAvailability("private", false);

    await contextmenu.triggerContextMenuAction("edit", "");
    await t.click(photoedit.infoTab);

    //await photoedit.checkFieldDisabled(photoedit.privateInput, true);

    await t.click(photoedit.dialogClose);
    await contextmenu.clearSelection();
    await toolbar.search("Viewpoint / Mexico / 2017");

    await photo.checkPhotoVisibility("pqmxlr7188hz4bih", true);

    await toolbar.search("Truck / Vancouver / 2019");

    await photo.checkPhotoVisibility("pqmxlr0kg161o9ek", false);

    await toolbar.search("Archive / 2020");

    await photo.checkPhotoVisibility("pqnah1k2frui6p63", true);

    await menu.openPage("settings");
    await t
      .click(settings.privateCheckbox)
      .click(settings.archiveCheckbox)
      .click(Selector(settings.libraryTab))
      .click(settings.reviewCheckbox);

    await menu.checkMenuItemAvailability("archive", true);
    await menu.checkMenuItemAvailability("private", true);
    await menu.checkMenuItemAvailability("review", true);
  }
);

test.meta("testID", "settings-general-006").meta({ type: "smoke" })(
  "Disable upload, download, edit and share",
  async (t) => {
    await toolbar.checkToolbarActionAvailability("upload", true);

    await toolbar.search("photo:true stack:true");
    await photo.triggerHoverAction("nth", 0, "select");

    await contextmenu.checkContextMenuActionAvailability("download", true);
    await contextmenu.checkContextMenuActionAvailability("share", true);
    await contextmenu.checkContextMenuActionAvailability("edit", true);

    await contextmenu.triggerContextMenuAction("edit", "");

    await photoedit.checkAllDetailsFieldsDisabled(false);
    await t.expect(photoedit.infoTab.visible).ok();
    await t.click(photoedit.filesTab);

    await t
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
    await menu.openPage("settings");

    await t
      .click(settings.uploadCheckbox)
      .click(settings.downloadCheckbox)
      .click(settings.editCheckbox)
      .click(settings.shareCheckbox);
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

    await toolbar.search("photo:true stack:true");
    await photo.triggerHoverAction("nth", 0, "select");

    await contextmenu.checkContextMenuActionAvailability("download", false);
    await contextmenu.checkContextMenuActionAvailability("share", false);
    await contextmenu.checkContextMenuActionAvailability("edit", false);

    await contextmenu.clearSelection();
    await t.click(page.cardTitle.nth(0));

    await photoedit.checkAllDetailsFieldsDisabled(true);
    await t.expect(photoedit.infoTab.visible).notOk();

    await t.click(photoedit.filesTab);

    await t
      .expect(photoedit.downloadFile.nth(0).visible)
      .notOk()
      .click(photoedit.toggleExpandFile.nth(1))
      .expect(photoedit.downloadFile.nth(1).visible)
      .notOk()
      .expect(photoedit.deleteFile.visible)
      .ok();

    await t.click(photoedit.dialogClose);
    await toolbar.search("photo:true");
    await photoviewer.openPhotoViewer("nth", 0);
    await photoviewer.checkPhotoViewerActionAvailability("download", false);
    await photoviewer.checkPhotoViewerActionAvailability("edit", true);
    await photoviewer.triggerPhotoViewerAction("close");

    await menu.openPage("settings");
    await t
      .click(settings.uploadCheckbox)
      .click(settings.downloadCheckbox)
      .click(settings.editCheckbox)
      .click(settings.shareCheckbox);
  }
);
