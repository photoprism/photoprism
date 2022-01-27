import { Selector } from "testcafe";
import { ClientFunction } from "testcafe";
import testcafeconfig from "../acceptance/testcafeconfig";
import Menu from "../page-model/menu";
import Photo from "../page-model/photo";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import Label from "../page-model/label";
import Album from "../page-model/album";
import Subject from "../page-model/subject";
import Page from "../page-model/page";
import Settings from "../page-model/settings";
import Library from "../page-model/library";
import PhotoEdit from "../page-model/photo-edit";
import AlbumDialog from "../page-model/dialog-album";

fixture`Test admin role`.page`${testcafeconfig.url}`;

const menu = new Menu();
const photo = new Photo();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const label = new Label();
const album = new Album();
const subject = new Subject();
const page = new Page();
const settings = new Settings();
const library = new Library();
const photoedit = new PhotoEdit();
const albumdialog = new AlbumDialog();

const getLocation = ClientFunction(() => document.location.href);

test.meta("testID", "authentication-000")(
  "Time to start instance (will be marked as unstable)",
  async (t) => {
    await t.wait(5000);
  }
);

test.meta("testID", "admin-role-001").meta({ type: "smoke" })("Access to settings", async (t) => {
  await page.login("admin", "photoprism");

  await menu.checkMenuItemAvailability("settings", true);

  await t.navigateTo("/settings");
  await t
    .expect(settings.languageInput.visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk();

  await t.navigateTo("/settings/library");

  await t
    .expect(Selector("form.p-form-settings").visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk();

  await t.navigateTo("/settings/advanced");

  await t
    .expect(Selector("label").withText("Read-Only Mode").visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk();

  await t.navigateTo("/settings/sync");

  await t
    .expect(Selector("div.p-accounts-list").visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk();
});

test.meta("testID", "admin-role-002").meta({ type: "smoke" })("Access to archive", async (t) => {
  await page.login("admin", "photoprism");
  const PhotoCountBrowse = await photo.getPhotoCount("all");

  await menu.checkMenuItemAvailability("archive", true);

  await t.navigateTo("/archive");

  await photo.checkPhotoVisibility("pqnahct2mvee8sr4", true);

  const PhotoCountArchive = await photo.getPhotoCount("all");

  await t.expect(PhotoCountBrowse).gte(PhotoCountArchive);
});

test.meta("testID", "admin-role-003").meta({ type: "smoke" })("Access to review", async (t) => {
  await page.login("admin", "photoprism");
  const PhotoCountBrowse = await photo.getPhotoCount("all");

  await menu.checkMenuItemAvailability("review", true);

  await t.navigateTo("/review");

  await photo.checkPhotoVisibility("pqzuein2pdcg1kc7", true);

  const PhotoCountReview = await photo.getPhotoCount("all");

  await t.expect(PhotoCountBrowse).gte(PhotoCountReview);
});

test.meta("testID", "admin-role-004").meta({ type: "smoke" })("Access to private", async (t) => {
  await page.login("admin", "photoprism");
  const PhotoCountBrowse = await photo.getPhotoCount("all");

  await menu.checkMenuItemAvailability("private", true);

  await t.navigateTo("/private");

  await photo.checkPhotoVisibility("pqmxlquf9tbc8mk2", true);

  const PhotoCountPrivate = await photo.getPhotoCount("all");

  await t.expect(PhotoCountBrowse).gte(PhotoCountPrivate);
});

test.meta("testID", "admin-role-005").meta({ type: "smoke" })("Access to library", async (t) => {
  await page.login("admin", "photoprism");
  const PhotoCountBrowse = await photo.getPhotoCount("all");

  await menu.checkMenuItemAvailability("library", true);

  await t.navigateTo("/library");

  await t
    .expect(library.indexFolderSelect.visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk();

  await t.navigateTo("/library/import");

  await t
    .expect(library.openImportFolderSelect.visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk();

  await t.navigateTo("/library/logs");

  await t
    .expect(Selector("div.terminal").visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk();
  await menu.checkMenuItemAvailability("originals", true);

  await t.navigateTo("/library/files");

  await t
    .expect(Selector("div.p-page-files").visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk();
  await menu.checkMenuItemAvailability("hidden", true);

  await t.navigateTo("/library/hidden");
  const PhotoCountHidden = await photo.getPhotoCount("all");

  await t.expect(PhotoCountBrowse).gte(PhotoCountHidden);
  await menu.checkMenuItemAvailability("errors", true);

  await t.navigateTo("/library/errors");

  await t
    .expect(Selector("div.p-page-errors").visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk();
});

test.meta("testID", "admin-role-006").meta({ type: "smoke" })(
  "private/archived photos in search results",
  async (t) => {
    await page.login("admin", "photoprism");
    const PhotoCountBrowse = await photo.getPhotoCount("all");
    await toolbar.search("private:true");
    const PhotoCountPrivate = await photo.getPhotoCount("all");

    await t.expect(PhotoCountPrivate).eql(2);
    await photo.checkPhotoVisibility("pqmxlquf9tbc8mk2", true);

    await toolbar.search("archived:true");
    const PhotoCountArchive = await photo.getPhotoCount("all");

    await t.expect(PhotoCountArchive).eql(3);
    await photo.checkPhotoVisibility("pqnahct2mvee8sr4", true);

    await toolbar.search("quality:0");
    const PhotoCountReview = await photo.getPhotoCount("all");

    await t.expect(PhotoCountReview).gte(PhotoCountBrowse);
    await photo.checkPhotoVisibility("pqzuein2pdcg1kc7", true);

    await menu.openPage("places");

    await t
      .expect(Selector("#map").exists, { timeout: 15000 })
      .ok()
      .expect(Selector("div.p-map-control").visible)
      .ok()
      .wait(5000);

    await t
      .typeText(Selector('input[aria-label="Search"]'), "oaxaca", { replace: true })
      .pressKey("enter");

    await t
      .expect(Selector("div.p-map-control").visible)
      .ok()
      .expect(getLocation())
      .contains("oaxaca")
      .wait(5000)
      .expect(Selector('div[title="Viewpoint / Mexico / 2017"]').visible)
      .notOk()
      .expect(Selector('div[title="Viewpoint / Mexico / 2018"]').visible)
      .notOk();

    await t
      .typeText(Selector('input[aria-label="Search"]'), "canada", { replace: true })
      .pressKey("enter");

    await t
      .expect(Selector("div.p-map-control").visible)
      .ok()
      .expect(getLocation())
      .contains("canada")
      .wait(8000)
      .expect(Selector('div[title="Cape / Bowen Island / 2019"]').visible)
      .ok()
      .expect(Selector('div[title="Truck / Vancouver / 2019"]').visible)
      .notOk();
  }
);

test.meta("testID", "admin-role-007").meta({ type: "smoke" })("Upload functionality", async (t) => {
  await page.login("admin", "photoprism");
  await toolbar.checkToolbarActionAvailability("upload", true);
  await menu.openPage("albums");
  await toolbar.checkToolbarActionAvailability("upload", true);
  await album.openNthAlbum(0);
  await toolbar.checkToolbarActionAvailability("upload", true);
  await menu.openPage("video");
  await toolbar.checkToolbarActionAvailability("upload", true);
  await menu.openPage("favorites");
  await toolbar.checkToolbarActionAvailability("upload", true);
  await menu.openPage("moments");
  await toolbar.checkToolbarActionAvailability("upload", true);
  await album.openNthAlbum(0);
  await toolbar.checkToolbarActionAvailability("upload", true);
  await menu.openPage("calendar");
  await toolbar.checkToolbarActionAvailability("upload", true);
  await album.openNthAlbum(0);
  await toolbar.checkToolbarActionAvailability("upload", true);
  await menu.openPage("states");
  await toolbar.checkToolbarActionAvailability("upload", true);
  await album.openNthAlbum(0);
  await toolbar.checkToolbarActionAvailability("upload", true);
  await menu.openPage("folders");
  await toolbar.checkToolbarActionAvailability("upload", true);
  await album.openNthAlbum(0);
  await toolbar.checkToolbarActionAvailability("upload", true);
});

test.meta("testID", "admin-role-008").meta({ type: "smoke" })(
  "Admin can private, archive, share, add/remove to album",
  async (t) => {
    await page.login("admin", "photoprism");
    const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
    await photo.selectPhotoFromUID(FirstPhotoUid);

    await contextmenu.checkContextMenuActionAvailability("private", true);
    await contextmenu.checkContextMenuActionAvailability("archive", true);
    await contextmenu.checkContextMenuActionAvailability("share", true);
    await contextmenu.checkContextMenuActionAvailability("album", true);

    await contextmenu.clearSelection();
    await toolbar.setFilter("view", "List");

    await photo.checkListViewActionAvailability("private", false);

    await menu.openPage("albums");
    await album.openNthAlbum(0);

    await toolbar.checkToolbarActionAvailability("share", true);

    await photo.toggleSelectNthPhoto(0, "all");

    await contextmenu.checkContextMenuActionAvailability("private", true);
    await contextmenu.checkContextMenuActionAvailability("remove", true);
    await contextmenu.checkContextMenuActionAvailability("share", true);
    await contextmenu.checkContextMenuActionAvailability("album", true);

    await contextmenu.clearSelection();
    await toolbar.triggerToolbarAction("list-view");

    await photo.checkListViewActionAvailability("private", false);
  }
);

test.meta("testID", "admin-role-009").meta({ type: "smoke" })(
  "Admin can approve low quality photos",
  async (t) => {
    await page.login("admin", "photoprism");
    await toolbar.search('quality:0 name:"photos-013_1"');
    await photo.toggleSelectNthPhoto(0, "all");
    await contextmenu.triggerContextMenuAction("edit", "");

    await t.expect(photoedit.detailsApprove.visible).ok();
  }
);

test.meta("testID", "admin-role-010").meta({ type: "smoke" })(
  "Edit dialog is not read only for admin",
  async (t) => {
    await page.login("admin", "photoprism");
    await toolbar.search("faces:new");
    const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
    await photo.selectPhotoFromUID(FirstPhotoUid);
    await contextmenu.triggerContextMenuAction("edit", "");

    await photoedit.checkAllDetailsFieldsDisabled(false);
    await t.expect(photoedit.detailsApply.visible).ok();
    if (t.browser.platform !== "mobile") {
      await t.expect(photoedit.detailsDone.visible).ok();
    }

    await t.click(photoedit.labelsTab);

    await photoedit.checkFieldDisabled(photoedit.removeLabel, false);
    await t.expect(photoedit.inputLabelName.exists).ok().expect(photoedit.addLabel.exists).ok();

    await t.click(photoedit.openInlineEdit);

    await t.expect(photoedit.inputLabelRename.exists).ok();

    await t.click(photoedit.peopleTab);

    await photoedit.checkFieldDisabled(photoedit.inputName, false);
    await t.expect(photoedit.removeMarker.exists).ok();

    await t.click(photoedit.infoTab);

    await photoedit.checkAllInfoFieldsDisabled(false);
  }
);

test.meta("testID", "admin-role-011").meta({ type: "smoke" })(
  "Edit labels functionality",
  async (t) => {
    await page.login("admin", "photoprism");
    await menu.openPage("labels");
    const FirstLabelUid = await label.getNthLabeltUid(0);

    await label.checkHoverActionState("uid", FirstLabelUid, "favorite", false);

    await label.triggerHoverAction("uid", FirstLabelUid, "favorite");

    await label.checkHoverActionState("uid", FirstLabelUid, "favorite", true);

    await label.triggerHoverAction("uid", FirstLabelUid, "favorite");

    await label.checkHoverActionState("uid", FirstLabelUid, "favorite", false);

    await t.click(Selector(`a.uid-${FirstLabelUid} div.inline-edit`));

    await t.expect(photoedit.inputLabelRename.visible).ok();

    await label.selectLabelFromUID(FirstLabelUid);

    await contextmenu.checkContextMenuActionAvailability("delete", true);
    await contextmenu.checkContextMenuActionAvailability("album", true);
  }
);

test.meta("testID", "admin-role-012").meta({ type: "smoke" })(
  "Edit album functionality",
  async (t) => {
    await page.login("admin", "photoprism");
    await menu.openPage("albums");

    await toolbar.checkToolbarActionAvailability("add", true);
    await album.checkHoverActionAvailability("nth", 1, "share", true);

    const FirstAlbumUid = await album.getNthAlbumUid("all", 0);
    await album.selectAlbumFromUID(FirstAlbumUid);

    await contextmenu.checkContextMenuActionAvailability("edit", true);
    await contextmenu.checkContextMenuActionAvailability("share", true);
    await contextmenu.checkContextMenuActionAvailability("clone", true);
    await contextmenu.checkContextMenuActionAvailability("download", true);
    await contextmenu.checkContextMenuActionAvailability("delete", true);

    await contextmenu.clearSelection();
    await t.click(page.cardTitle);

    await t.expect(albumdialog.description.visible).ok();

    await t.click(albumdialog.dialogCancel);

    if (await Selector(`a.uid-${FirstAlbumUid}`).hasClass("is-favorite")) {
      await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", true);
      await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
      await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", false);
      await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
      await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", true);
    } else {
      await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", false);
      await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
      await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", true);
      await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
      await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", false);
    }
    await album.openNthAlbum(0);

    await toolbar.checkToolbarActionAvailability("share", true);
    await toolbar.checkToolbarActionAvailability("edit", true);

    await photo.toggleSelectNthPhoto(0, "all");

    await contextmenu.checkContextMenuActionAvailability("album", true);
    await contextmenu.checkContextMenuActionAvailability("private", true);
    await contextmenu.checkContextMenuActionAvailability("share", true);
    await contextmenu.checkContextMenuActionAvailability("remove", true);
  }
);

test.meta("testID", "admin-role-013")("Edit moment functionality", async (t) => {
  await page.login("admin", "photoprism");
  await menu.openPage("moments");

  await album.checkHoverActionAvailability("nth", 0, "share", true);

  const FirstAlbumUid = await album.getNthAlbumUid("moment", 0);
  await album.selectAlbumFromUID(FirstAlbumUid);

  await contextmenu.checkContextMenuActionAvailability("edit", true);
  await contextmenu.checkContextMenuActionAvailability("share", true);
  await contextmenu.checkContextMenuActionAvailability("clone", true);
  await contextmenu.checkContextMenuActionAvailability("download", true);
  await contextmenu.checkContextMenuActionAvailability("delete", true);

  await contextmenu.clearSelection();
  await t.click(page.cardTitle);

  await t.expect(albumdialog.description.visible).ok();

  await t.click(albumdialog.dialogCancel);

  if (await Selector(`a.uid-${FirstAlbumUid}`).hasClass("is-favorite")) {
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", true);
    await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", false);
    await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", true);
  } else {
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", false);
    await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", true);
    await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", false);
  }
  await album.openNthAlbum(0);

  await toolbar.checkToolbarActionAvailability("share", true);
  await toolbar.checkToolbarActionAvailability("edit", true);

  await photo.toggleSelectNthPhoto(0, "all");

  await contextmenu.checkContextMenuActionAvailability("album", true);
  await contextmenu.checkContextMenuActionAvailability("private", true);
  await contextmenu.checkContextMenuActionAvailability("share", true);
  await contextmenu.checkContextMenuActionAvailability("archive", true);
  await contextmenu.checkContextMenuActionAvailability("remove", false);
});

test.meta("testID", "admin-role-014").meta({ type: "smoke" })(
  "Edit state functionality",
  async (t) => {
    await page.login("admin", "photoprism");
    await menu.openPage("states");

    await album.checkHoverActionAvailability("nth", 0, "share", true);

    const FirstAlbumUid = await album.getNthAlbumUid("state", 0);
    await album.selectAlbumFromUID(FirstAlbumUid);

    await contextmenu.checkContextMenuActionAvailability("edit", true);
    await contextmenu.checkContextMenuActionAvailability("share", true);
    await contextmenu.checkContextMenuActionAvailability("clone", true);
    await contextmenu.checkContextMenuActionAvailability("download", true);
    await contextmenu.checkContextMenuActionAvailability("delete", true);

    await contextmenu.clearSelection();
    await t.click(page.cardTitle);

    await t.expect(albumdialog.description.visible).ok();

    await t.click(albumdialog.dialogCancel);

    if (await Selector(`a.uid-${FirstAlbumUid}`).hasClass("is-favorite")) {
      await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", true);
      await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
      await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", false);
      await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
      await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", true);
    } else {
      await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", false);
      await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
      await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", true);
      await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
      await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", false);
    }

    await album.openNthAlbum(0);

    await toolbar.checkToolbarActionAvailability("share", true);
    await toolbar.checkToolbarActionAvailability("edit", true);

    await photo.toggleSelectNthPhoto(0, "all");

    await contextmenu.checkContextMenuActionAvailability("album", true);
    await contextmenu.checkContextMenuActionAvailability("private", true);
    await contextmenu.checkContextMenuActionAvailability("share", true);
    await contextmenu.checkContextMenuActionAvailability("archive", true);
    await contextmenu.checkContextMenuActionAvailability("remove", false);
  }
);

test.meta("testID", "admin-role-015")("Edit calendar functionality", async (t) => {
  await page.login("admin", "photoprism");
  await menu.openPage("calendar");

  await album.checkHoverActionAvailability("nth", 0, "share", true);

  const FirstAlbumUid = await album.getNthAlbumUid("month", 0);
  await album.selectAlbumFromUID(FirstAlbumUid);

  await contextmenu.checkContextMenuActionAvailability("edit", true);
  await contextmenu.checkContextMenuActionAvailability("share", true);
  await contextmenu.checkContextMenuActionAvailability("clone", true);
  await contextmenu.checkContextMenuActionAvailability("download", true);
  await contextmenu.checkContextMenuActionAvailability("delete", false);

  await contextmenu.clearSelection();
  await t.click(page.cardTitle);

  await t.expect(albumdialog.description.visible).ok();

  await t.click(albumdialog.dialogCancel);

  if (await Selector(`a.uid-${FirstAlbumUid}`).hasClass("is-favorite")) {
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", true);
    await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", false);
    await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", true);
  } else {
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", false);
    await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", true);
    await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", false);
  }

  await album.openNthAlbum(0);

  await toolbar.checkToolbarActionAvailability("share", true);
  await toolbar.checkToolbarActionAvailability("edit", true);

  await photo.toggleSelectNthPhoto(0, "all");

  await contextmenu.checkContextMenuActionAvailability("album", true);
  await contextmenu.checkContextMenuActionAvailability("private", true);
  await contextmenu.checkContextMenuActionAvailability("share", true);
  await contextmenu.checkContextMenuActionAvailability("archive", true);
});

test.meta("testID", "admin-role-016")("Edit folder functionality", async (t) => {
  await page.login("admin", "photoprism");
  await menu.openPage("folders");

  await album.checkHoverActionAvailability("nth", 0, "share", true);

  const FirstAlbumUid = await album.getNthAlbumUid("folder", 0);
  await album.selectAlbumFromUID(FirstAlbumUid);

  await contextmenu.checkContextMenuActionAvailability("edit", true);
  await contextmenu.checkContextMenuActionAvailability("share", true);
  await contextmenu.checkContextMenuActionAvailability("clone", true);
  await contextmenu.checkContextMenuActionAvailability("download", true);
  await contextmenu.checkContextMenuActionAvailability("delete", false);

  await contextmenu.clearSelection();
  await t.click(page.cardTitle);

  await t.expect(albumdialog.description.visible).ok();

  await t.click(albumdialog.dialogCancel);

  if (await Selector(`a.uid-${FirstAlbumUid}`).hasClass("is-favorite")) {
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", true);
    await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", false);
    await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", true);
  } else {
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", false);
    await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", true);
    await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", false);
  }

  await album.openNthAlbum(0);

  await toolbar.checkToolbarActionAvailability("share", true);
  await toolbar.checkToolbarActionAvailability("edit", true);

  await photo.toggleSelectNthPhoto(0, "all");

  await contextmenu.checkContextMenuActionAvailability("album", true);
  await contextmenu.checkContextMenuActionAvailability("private", true);
  await contextmenu.checkContextMenuActionAvailability("share", true);
  await contextmenu.checkContextMenuActionAvailability("archive", true);
});

test.meta("testID", "admin-role-017").meta({ type: "smoke" })(
  "Edit people functionality",
  async (t) => {
    await page.login("admin", "photoprism");
    await menu.openPage("people");

    await toolbar.checkToolbarActionAvailability("show-hidden", true);
    await t.expect(subject.newTab.exists).ok();
    await subject.checkSubjectVisibility("name", "Otto Visible", true);
    await subject.checkSubjectVisibility("name", "Monika Hide", false);

    await toolbar.triggerToolbarAction("show-hidden");

    await subject.checkSubjectVisibility("name", "Otto Visible", true);
    await subject.checkSubjectVisibility("name", "Monika Hide", true);

    await t.click(Selector("a div.v-card__title").nth(0));

    await t.expect(Selector("div.input-rename input").visible).ok();
    await subject.checkHoverActionAvailability("nth", 0, "hidden", true);

    await subject.toggleSelectNthSubject(0);
    await contextmenu.checkContextMenuActionAvailability("album", "true");
    await contextmenu.clearSelection();

    const FirstSubjectUid = await subject.getNthSubjectUid(0);

    if (await Selector(`a.uid-${FirstSubjectUid}`).hasClass("is-favorite")) {
      await subject.checkHoverActionState("uid", FirstSubjectUid, "favorite", true);
      await subject.triggerHoverAction("uid", FirstSubjectUid, "favorite");
      await subject.checkHoverActionState("uid", FirstSubjectUid, "favorite", false);
      await subject.triggerHoverAction("uid", FirstSubjectUid, "favorite");
      await subject.checkHoverActionState("uid", FirstSubjectUid, "favorite", true);
    } else {
      await subject.checkHoverActionState("uid", FirstSubjectUid, "favorite", false);
      await subject.triggerHoverAction("uid", FirstSubjectUid, "favorite");
      await subject.checkHoverActionState("uid", FirstSubjectUid, "favorite", true);
      await subject.triggerHoverAction("uid", FirstSubjectUid, "favorite");
      await subject.checkHoverActionState("uid", FirstSubjectUid, "favorite", false);
    }

    await subject.openNthSubject(0);
    await photo.toggleSelectNthPhoto(0, "all");
    await contextmenu.triggerContextMenuAction("edit", "");

    await t.click(photoedit.peopleTab);

    await photoedit.checkFieldDisabled(photoedit.inputName, false);
    await t.expect(photoedit.rejectName.hasClass("v-icon--disabled")).notOk();

    await t.navigateTo("/people/new");

    await t.expect(Selector("div.is-face").visible).ok();

    await t.navigateTo("/people?hidden=yes&order=relevance");
    await t.expect(Selector("a.is-subject").visible).ok();
  }
);
