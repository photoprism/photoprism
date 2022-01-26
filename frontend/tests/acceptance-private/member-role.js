import { Selector } from "testcafe";
import testcafeconfig from "../acceptance/testcafeconfig";
import { ClientFunction } from "testcafe";
import Menu from "../page-model/menu";
import Photo from "../page-model/photo";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import Label from "../page-model/label";
import Album from "../page-model/album";
import Subject from "../page-model/subject";
import PhotoViewer from "../page-model/photoviewer";
import PhotoEdit from "../page-model/photo-edit";
import Page from "../page-model/page";
import Settings from "../page-model/settings";
import Library from "../page-model/library";
import AlbumDialog from "../page-model/dialog-album";

fixture`Test member role`.page`${testcafeconfig.url}`;

const menu = new Menu();
const photo = new Photo();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const label = new Label();
const album = new Album();
const subject = new Subject();
const photoviewer = new PhotoViewer();
const photoedit = new PhotoEdit();
const page = new Page();
const settings = new Settings();
const library = new Library();
const albumdialog = new AlbumDialog();

const getLocation = ClientFunction(() => document.location.href);

test.meta("testID", "member-role-001")("No access to settings", async (t) => {
  await page.login("member", "passwdmember");

  await menu.checkMenuItemAvailability("settings", false);

  await t.navigateTo("/settings");

  await t
    .expect(settings.languageInput.visible)
    .notOk()
    .expect(Selector("div.p-page-photos").visible)
    .ok();

  await t.navigateTo("/settings/library");

  await t
    .expect(Selector("form.p-form-settings").visible)
    .notOk()
    .expect(Selector("div.p-page-photos").visible)
    .ok();

  await t.navigateTo("/settings/advanced");

  await t
    .expect(Selector("label").withText("Read-Only Mode").visible)
    .notOk()
    .expect(Selector("div.p-page-photos").visible)
    .ok();

  await t.navigateTo("/settings/sync");

  await t
    .expect(Selector("div.p-accounts-list").visible)
    .notOk()
    .expect(Selector("div.p-page-photos").visible)
    .ok();
});

test.meta("testID", "member-role-002")("No access to archive", async (t) => {
  await page.login("member", "passwdmember");
  if (t.browser.platform === "mobile") {
    await t.wait(5000);
  }
  const PhotoCountBrowse = await photo.getPhotoCount("all");

  await menu.checkMenuItemAvailability("archive", false);

  await t.navigateTo("/archive");

  await photo.checkPhotoVisibility("pqnahct2mvee8sr4", false);

  if (t.browser.platform === "mobile") {
    await t.wait(5000);
  }

  const PhotoCountArchive = await photo.getPhotoCount("all");
  await t.expect(PhotoCountBrowse).eql(PhotoCountArchive);
});

test.meta("testID", "member-role-003")("No access to review", async (t) => {
  await page.login("member", "passwdmember");
  if (t.browser.platform === "mobile") {
    await t.wait(5000);
  }
  const PhotoCountBrowse = await photo.getPhotoCount("all");

  await menu.checkMenuItemAvailability("review", false);

  await t.navigateTo("/review");

  await photo.checkPhotoVisibility("pqzuein2pdcg1kc7", false);

  if (t.browser.platform === "mobile") {
    await t.wait(5000);
  }
  const PhotoCountReview = await photo.getPhotoCount("all");

  await t.expect(PhotoCountBrowse).eql(PhotoCountReview);
});

test.meta("testID", "member-role-004")("No access to private", async (t) => {
  await page.login("member", "passwdmember");
  if (t.browser.platform === "mobile") {
    await t.wait(5000);
  }
  const PhotoCountBrowse = await photo.getPhotoCount("all");

  await menu.checkMenuItemAvailability("private", false);

  await t.navigateTo("/private");

  await photo.checkPhotoVisibility("pqmxlquf9tbc8mk2", false);

  if (t.browser.platform === "mobile") {
    await t.wait(5000);
  }
  const PhotoCountPrivate = await photo.getPhotoCount("all");

  await t.expect(PhotoCountBrowse).eql(PhotoCountPrivate);
});

test.meta("testID", "member-role-005")("No access to library", async (t) => {
  await page.login("member", "passwdmember");
  if (t.browser.platform === "mobile") {
    await t.wait(5000);
  }
  const PhotoCountBrowse = await photo.getPhotoCount("all");

  await menu.checkMenuItemAvailability("library", false);

  await t.navigateTo("/library");

  await t
    .expect(library.indexFolderSelect.visible)
    .notOk()
    .expect(Selector("div.p-page-photos").visible)
    .ok();

  await t.navigateTo("/library/import");

  await t
    .expect(library.openImportFolderSelect.visible)
    .notOk()
    .expect(Selector("div.p-page-photos").visible)
    .ok();

  await t.navigateTo("/library/logs");

  await t
    .expect(Selector("p.p-log-message").visible)
    .notOk()
    .expect(Selector("div.p-page-photos").visible)
    .ok();
  await menu.checkMenuItemAvailability("originals", false);

  await t.navigateTo("/library/files");

  await t
    .expect(Selector("div.p-page-files").visible)
    .notOk()
    .expect(Selector("div.p-page-photos").visible)
    .ok();
  await menu.checkMenuItemAvailability("hidden", false);

  await t.navigateTo("/library/hidden");
  if (t.browser.platform === "mobile") {
    await t.wait(5000);
  }
  const PhotoCountHidden = await photo.getPhotoCount("all");

  await t.expect(PhotoCountBrowse).eql(PhotoCountHidden);
  await menu.checkMenuItemAvailability("errors", false);

  await t.navigateTo("/library/errors");

  await t
    .expect(Selector("div.p-page-errors").visible)
    .notOk()
    .expect(Selector("div.p-page-photos").visible)
    .ok();
});

test.meta("testID", "member-role-006")(
  "No private/archived photos in search results",
  async (t) => {
    await page.login("member", "passwdmember");
    if (t.browser.platform === "mobile") {
      await t.wait(5000);
    }
    const PhotoCountBrowse = await photo.getPhotoCount("all");
    await toolbar.search("private:true");
    const PhotoCountPrivate = await photo.getPhotoCount("all");

    await t.expect(PhotoCountPrivate).eql(0);
    await photo.checkPhotoVisibility("pqmxlquf9tbc8mk2", false);

    await toolbar.search("archived:true");
    const PhotoCountArchive = await photo.getPhotoCount("all");

    await t.expect(PhotoCountArchive).eql(0);
    await photo.checkPhotoVisibility("pqnahct2mvee8sr4", false);

    await toolbar.search("quality:0");
    if (t.browser.platform === "mobile") {
      await t.wait(5000);
    }
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

test.meta("testID", "member-role-007")("No upload functionality", async (t) => {
  await page.login("member", "passwdmember");
  await toolbar.checkToolbarActionAvailability("upload", false);
  await menu.openPage("albums");
  await toolbar.checkToolbarActionAvailability("upload", false);
  await album.openNthAlbum(0);
  await toolbar.checkToolbarActionAvailability("upload", false);
  await menu.openPage("video");
  await toolbar.checkToolbarActionAvailability("upload", false);
  await menu.openPage("people");
  await toolbar.checkToolbarActionAvailability("upload", false);
  await menu.openPage("favorites");
  await toolbar.checkToolbarActionAvailability("upload", false);
  await menu.openPage("moments");
  await toolbar.checkToolbarActionAvailability("upload", false);
  await album.openNthAlbum(0);
  await toolbar.checkToolbarActionAvailability("upload", false);
  await menu.openPage("calendar");
  await toolbar.checkToolbarActionAvailability("upload", false);
  await album.openNthAlbum(0);
  await toolbar.checkToolbarActionAvailability("upload", false);
  await menu.openPage("states");
  await toolbar.checkToolbarActionAvailability("upload", false);
  await album.openNthAlbum(0);
  await toolbar.checkToolbarActionAvailability("upload", false);
  await menu.openPage("folders");
  await toolbar.checkToolbarActionAvailability("upload", false);
  await album.openNthAlbum(0);
  await toolbar.checkToolbarActionAvailability("upload", false);
});

test.meta("testID", "member-role-008")("Member cannot like photos", async (t) => {
  await page.login("member", "passwdmember");
  const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
  const SecondPhotoUid = await photo.getNthPhotoUid("all", 1);
  await menu.openPage("favorites");
  const FirstFavoriteUid = await photo.getNthPhotoUid("image", 0);

  await photo.checkHoverActionState("uid", FirstFavoriteUid, "favorite", true);

  await photo.triggerHoverAction("uid", FirstFavoriteUid, "favorite");

  await photo.checkHoverActionState("uid", FirstFavoriteUid, "favorite", true);
  await photo.checkPhotoVisibility(FirstPhotoUid, false);
  await photo.checkPhotoVisibility(SecondPhotoUid, false);

  await menu.openPage("browse");
  await photoviewer.openPhotoViewer("uid", FirstPhotoUid);

  await photoviewer.checkPhotoViewerActionAvailability("like", false);
  await photoviewer.checkPhotoViewerActionAvailability("close", true);

  await photoviewer.triggerPhotoViewerAction("close");
  await t.wait(5000).click(Selector(".p-expand-search", { timeout: 10000 }));
  await toolbar.setFilter("view", "Cards");

  await photo.checkHoverActionState("uid", FirstPhotoUid, "favorite", false);

  await photo.triggerHoverAction("uid", FirstPhotoUid, "favorite");

  await photo.checkHoverActionState("uid", FirstPhotoUid, "favorite", false);

  await photo.selectPhotoFromUID(SecondPhotoUid);
  await contextmenu.triggerContextMenuAction("edit", "");
  await t.click(photoedit.infoTab);

  await t.expect(Selector(".input-favorite input").hasAttribute("disabled")).ok();

  await photoedit.turnSwitchOn("favorite");
  await t.click(Selector(".action-close"));
  await contextmenu.clearSelection();

  await photo.checkHoverActionState("uid", SecondPhotoUid, "favorite", false);

  await menu.openPage("browse");
  await toolbar.setFilter("view", "Mosaic");

  await photo.checkHoverActionState("uid", FirstPhotoUid, "favorite", false);

  await photo.triggerHoverAction("uid", FirstPhotoUid, "favorite");

  await photo.checkHoverActionState("uid", FirstPhotoUid, "favorite", false);

  await toolbar.setFilter("view", "List");

  await photo.checkListViewActionAvailability("like", true);

  await photo.triggerListViewActions("nth", 0, "like");
  await toolbar.setFilter("view", "Cards");

  await photo.checkHoverActionState("uid", FirstPhotoUid, "favorite", false);

  await menu.openPage("albums");
  await album.openNthAlbum(0);
  await toolbar.triggerToolbarAction("list-view");
  if (t.browser.platform === "mobile") {
    await toolbar.triggerToolbarAction("list-view");
  }

  await photo.checkListViewActionAvailability("like", true);
});

test.meta("testID", "member-role-009")(
  "Member cannot private, archive, share, add/remove to album",
  async (t) => {
    await page.login("member", "passwdmember");
    const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
    await photo.selectPhotoFromUID(FirstPhotoUid);

    await contextmenu.checkContextMenuActionAvailability("private", false);
    await contextmenu.checkContextMenuActionAvailability("archive", false);
    await contextmenu.checkContextMenuActionAvailability("share", false);
    await contextmenu.checkContextMenuActionAvailability("album", false);
    await contextmenu.checkContextMenuActionAvailability("download", true);
    await contextmenu.checkContextMenuActionAvailability("edit", true);

    await contextmenu.clearSelection();
    await toolbar.setFilter("view", "List");

    await photo.checkListViewActionAvailability("private", true);

    await menu.openPage("albums");
    await album.openNthAlbum(0);

    await toolbar.checkToolbarActionAvailability("share", false);

    await photo.toggleSelectNthPhoto(0, "all");

    await contextmenu.checkContextMenuActionAvailability("private", false);
    await contextmenu.checkContextMenuActionAvailability("remove", false);
    await contextmenu.checkContextMenuActionAvailability("share", false);
    await contextmenu.checkContextMenuActionAvailability("album", false);

    await contextmenu.clearSelection();
    await toolbar.triggerToolbarAction("list-view");
    if (t.browser.platform === "mobile") {
      await toolbar.triggerToolbarAction("list-view");
    }

    await photo.checkListViewActionAvailability("private", true);
  }
);

test.meta("testID", "member-role-010")("Member cannot approve low quality photos", async (t) => {
  await page.login("member", "passwdmember");
  await toolbar.search('quality:0 name:"photos-013_1"');
  await photo.toggleSelectNthPhoto(0, "all");
  await contextmenu.triggerContextMenuAction("edit", "");

  await t.expect(photoedit.detailsApprove.visible).notOk();
});

test.meta("testID", "member-role-011")("Edit dialog is read only for member", async (t) => {
  await page.login("member", "passwdmember");
  await toolbar.search("faces:new");
  const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
  await photo.selectPhotoFromUID(FirstPhotoUid);
  await contextmenu.triggerContextMenuAction("edit", "");

  await t
    .expect(photoedit.title.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.localTime.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.day.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.month.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.year.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.timezone.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.latitude.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.longitude.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.altitude.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.country.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.camera.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.iso.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.exposure.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.lens.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.fnumber.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.focallength.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.subject.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.artist.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.copyright.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.license.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.description.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.keywords.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.notes.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.detailsApply.visible)
    .notOk();
  if (t.browser.platform !== "mobile") {
    await t.expect(photoedit.detailsDone.visible).notOk();
  }

  await t.click(photoedit.labelsTab);

  await t
    .expect(photoedit.removeLabel.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.inputLabelName.exists)
    .notOk()
    .expect(photoedit.addLabel.exists)
    .notOk();

  await t.click(photoedit.openInlineEdit);

  await t.expect(photoedit.inputLabelRename.exists).notOk();

  await t.click(photoedit.peopleTab);

  await t
    .expect(photoedit.inputName.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.removeMarker.exists)
    .notOk();

  await t.click(photoedit.infoTab);

  await t
    .expect(photoedit.favoriteInput.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.privateInput.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.scanInput.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.panoramaInput.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.stackableInput.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.typeInput.hasAttribute("disabled"))
    .ok();
});

test.meta("testID", "member-role-012")("No edit album functionality", async (t) => {
  await page.login("member", "passwdmember");
  await menu.openPage("albums");

  await toolbar.checkToolbarActionAvailability("add", false);
  await album.checkHoverActionAvailability("nth", 1, "share", false);

  const FirstAlbumUid = await album.getNthAlbumUid("all", 0);
  await album.selectAlbumFromUID(FirstAlbumUid);

  await contextmenu.checkContextMenuActionAvailability("edit", false);
  await contextmenu.checkContextMenuActionAvailability("share", false);
  await contextmenu.checkContextMenuActionAvailability("clone", false);
  await contextmenu.checkContextMenuActionAvailability("download", true);
  await contextmenu.checkContextMenuActionAvailability("delete", false);

  await contextmenu.clearSelection();
  await t.click(page.cardTitle);

  await t.expect(albumdialog.description.visible).notOk();

  if (await Selector(`a.uid-${FirstAlbumUid}`).hasClass("is-favorite")) {
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", true);
    await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", true);
  } else {
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", false);
    await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", false);
  }
  await album.openNthAlbum(0);

  await toolbar.checkToolbarActionAvailability("share", false);
  await toolbar.checkToolbarActionAvailability("edit", false);

  await photo.toggleSelectNthPhoto(0, "all");

  await contextmenu.checkContextMenuActionAvailability("album", false);
  await contextmenu.checkContextMenuActionAvailability("private", false);
  await contextmenu.checkContextMenuActionAvailability("share", false);
  await contextmenu.checkContextMenuActionAvailability("remove", false);
});

test.meta("testID", "member-role-013")("No edit moment functionality", async (t) => {
  await page.login("member", "passwdmember");
  await menu.openPage("moments");

  await album.checkHoverActionAvailability("nth", 0, "share", false);

  const FirstAlbumUid = await album.getNthAlbumUid("all", 0);
  await album.selectAlbumFromUID(FirstAlbumUid);

  await contextmenu.checkContextMenuActionAvailability("edit", false);
  await contextmenu.checkContextMenuActionAvailability("share", false);
  await contextmenu.checkContextMenuActionAvailability("clone", false);
  await contextmenu.checkContextMenuActionAvailability("download", true);
  await contextmenu.checkContextMenuActionAvailability("delete", false);

  await contextmenu.clearSelection();
  await t.click(page.cardTitle);

  await t.expect(albumdialog.description.visible).notOk();

  if (await Selector(`a.uid-${FirstAlbumUid}`).hasClass("is-favorite")) {
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", true);
    await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", true);
  } else {
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", false);
    await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", false);
  }
  await album.openNthAlbum(0);

  await toolbar.checkToolbarActionAvailability("share", false);
  await toolbar.checkToolbarActionAvailability("edit", false);

  await photo.toggleSelectNthPhoto(0, "all");

  await contextmenu.checkContextMenuActionAvailability("album", false);
  await contextmenu.checkContextMenuActionAvailability("private", false);
  await contextmenu.checkContextMenuActionAvailability("share", false);
  await contextmenu.checkContextMenuActionAvailability("remove", false);
});

test.meta("testID", "member-role-014")("No edit state functionality", async (t) => {
  await page.login("member", "passwdmember");
  await menu.openPage("states");

  await album.checkHoverActionAvailability("nth", 0, "share", false);

  const FirstAlbumUid = await album.getNthAlbumUid("all", 0);
  await album.selectAlbumFromUID(FirstAlbumUid);

  await contextmenu.checkContextMenuActionAvailability("edit", false);
  await contextmenu.checkContextMenuActionAvailability("share", false);
  await contextmenu.checkContextMenuActionAvailability("clone", false);
  await contextmenu.checkContextMenuActionAvailability("download", true);
  await contextmenu.checkContextMenuActionAvailability("delete", false);

  await contextmenu.clearSelection();
  await t.click(page.cardTitle);

  await t.expect(albumdialog.description.visible).notOk();

  if (await Selector(`a.uid-${FirstAlbumUid}`).hasClass("is-favorite")) {
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", true);
    await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", true);
  } else {
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", false);
    await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", false);
  }
  await album.openNthAlbum(0);

  await toolbar.checkToolbarActionAvailability("share", false);
  await toolbar.checkToolbarActionAvailability("edit", false);

  await photo.toggleSelectNthPhoto(0, "all");

  await contextmenu.checkContextMenuActionAvailability("album", false);
  await contextmenu.checkContextMenuActionAvailability("private", false);
  await contextmenu.checkContextMenuActionAvailability("share", false);
  await contextmenu.checkContextMenuActionAvailability("remove", false);
});

test.meta("testID", "member-role-015")("No edit calendar functionality", async (t) => {
  await page.login("member", "passwdmember");
  await menu.openPage("calendar");

  await album.checkHoverActionAvailability("nth", 0, "share", false);

  const FirstAlbumUid = await album.getNthAlbumUid("all", 0);
  await album.selectAlbumFromUID(FirstAlbumUid);

  await contextmenu.checkContextMenuActionAvailability("edit", false);
  await contextmenu.checkContextMenuActionAvailability("share", false);
  await contextmenu.checkContextMenuActionAvailability("clone", false);
  await contextmenu.checkContextMenuActionAvailability("download", true);
  await contextmenu.checkContextMenuActionAvailability("delete", false);

  await contextmenu.clearSelection();
  await t.click(page.cardTitle);

  await t.expect(albumdialog.description.visible).notOk();

  if (await Selector(`a.uid-${FirstAlbumUid}`).hasClass("is-favorite")) {
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", true);
    await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", true);
  } else {
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", false);
    await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", false);
  }
  await album.openNthAlbum(0);

  await toolbar.checkToolbarActionAvailability("share", false);
  await toolbar.checkToolbarActionAvailability("edit", false);

  await photo.toggleSelectNthPhoto(0, "all");

  await contextmenu.checkContextMenuActionAvailability("album", false);
  await contextmenu.checkContextMenuActionAvailability("private", false);
  await contextmenu.checkContextMenuActionAvailability("share", false);
  await contextmenu.checkContextMenuActionAvailability("remove", false);
});

test.meta("testID", "member-role-016")("No edit folder functionality", async (t) => {
  await page.login("member", "passwdmember");
  await menu.openPage("folders");

  await album.checkHoverActionAvailability("nth", 0, "share", false);

  const FirstAlbumUid = await album.getNthAlbumUid("all", 0);
  await album.selectAlbumFromUID(FirstAlbumUid);

  await contextmenu.checkContextMenuActionAvailability("edit", false);
  await contextmenu.checkContextMenuActionAvailability("share", false);
  await contextmenu.checkContextMenuActionAvailability("clone", false);
  await contextmenu.checkContextMenuActionAvailability("download", true);
  await contextmenu.checkContextMenuActionAvailability("delete", false);

  await contextmenu.clearSelection();
  await t.click(page.cardTitle);

  await t.expect(albumdialog.description.visible).notOk();

  if (await Selector(`a.uid-${FirstAlbumUid}`).hasClass("is-favorite")) {
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", true);
    await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", true);
  } else {
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", false);
    await album.triggerHoverAction("uid", FirstAlbumUid, "favorite");
    await album.checkHoverActionState("uid", FirstAlbumUid, "favorite", false);
  }

  await album.openNthAlbum(0);

  await toolbar.checkToolbarActionAvailability("share", false);
  await toolbar.checkToolbarActionAvailability("edit", false);

  await photo.toggleSelectNthPhoto(0, "all");

  await contextmenu.checkContextMenuActionAvailability("album", false);
  await contextmenu.checkContextMenuActionAvailability("private", false);
  await contextmenu.checkContextMenuActionAvailability("share", false);
  await contextmenu.checkContextMenuActionAvailability("remove", false);
});

test.meta("testID", "member-role-017")("No edit labels functionality", async (t) => {
  await page.login("member", "passwdmember");
  await menu.openPage("labels");
  const FirstLabelUid = await label.getNthLabeltUid(0);

  await label.checkHoverActionState("uid", FirstLabelUid, "favorite", false);

  await label.triggerHoverAction("uid", FirstLabelUid, "favorite");

  await label.checkHoverActionState("uid", FirstLabelUid, "favorite", false);

  await t.click(Selector(`a.uid-${FirstLabelUid} div.inline-edit`));

  await t.expect(photoedit.inputLabelRename.visible).notOk();

  await label.selectLabelFromUID(FirstLabelUid);

  await contextmenu.checkContextMenuActionAvailability("delete", false);
  await contextmenu.checkContextMenuActionAvailability("album", false);
});

test.meta("testID", "member-role-018")("No unstack, change primary actions", async (t) => {
  await page.login("member", "passwdmember");
  await toolbar.search("stack:true");

  const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
  await photo.selectPhotoFromUID(FirstPhotoUid);
  await contextmenu.triggerContextMenuAction("edit", "");
  await t.click(photoedit.filesTab);

  await t
    .expect(photoedit.downloadFile.visible)
    .ok()
    .expect(photoedit.downloadFile.hasAttribute("disabled"))
    .notOk();

  await t.click(photoedit.toggleExpandFile.nth(1));

  await t
    .expect(photoedit.downloadFile.visible)
    .ok()
    .expect(photoedit.downloadFile.hasAttribute("disabled"))
    .notOk()
    .expect(photoedit.unstackFile.visible)
    .notOk()
    .expect(photoedit.makeFilePrimary.visible)
    .notOk()
    .expect(photoedit.deleteFile.visible)
    .notOk();
});

test.meta("testID", "member-role-019")("No edit people functionality", async (t) => {
  await page.login("member", "passwdmember");
  await menu.openPage("people");

  await toolbar.checkToolbarActionAvailability("show-hidden", false);
  await toolbar.checkToolbarActionAvailability("upload", false);
  await subject.checkSubjectVisibility("name", "Otto Visible", true);
  await subject.checkSubjectVisibility("name", "Monika Hide", false);
  await t.expect(subject.newTab.exists).notOk();

  await t.click(Selector("a div.v-card__title").nth(0));

  await t.expect(Selector("div.input-rename input").visible).notOk();
  await subject.checkHoverActionAvailability("nth", 0, "hidden", false);

  await subject.toggleSelectNthSubject(0);

  await contextmenu.checkContextMenuActionAvailability("album", "false");
  await contextmenu.checkContextMenuActionAvailability("download", "true");

  await contextmenu.clearSelection();
  const FirstSubjectUid = subject.getNthSubjectUid(0);

  if (await Selector(`a.uid-${FirstSubjectUid}`).hasClass("is-favorite")) {
    await subject.checkHoverActionState("uid", FirstSubjectUid, "favorite", true);
    await subject.triggerHoverAction("uid", FirstSubjectUid, "favorite");
    await subject.checkHoverActionState("uid", FirstSubjectUid, "favorite", true);
  } else {
    await subject.checkHoverActionState("uid", FirstSubjectUid, "favorite", false);
    await subject.triggerHoverAction("uid", FirstSubjectUid, "favorite");
    await subject.checkHoverActionState("uid", FirstSubjectUid, "favorite", false);
  }

  await subject.openNthSubject(0);
  await photo.toggleSelectNthPhoto(0);
  await contextmenu.triggerContextMenuAction("edit", "");
  await t.click(photoedit.peopleTab);

  await t
    .expect(photoedit.inputName.hasAttribute("disabled"))
    .ok()
    .expect(photoedit.rejectName.hasClass("v-icon--disabled"))
    .ok();

  await t.navigateTo("/people/new");

  await t.expect(Selector("div.is-face").visible).notOk().expect(subject.newTab.exists).notOk();

  await t.navigateTo("/people?hidden=yes&order=relevance");
  await t.expect(Selector("a.is-subject").visible).notOk();
});
