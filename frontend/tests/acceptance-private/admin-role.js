import { Selector } from "testcafe";
import { ClientFunction } from "testcafe";
import testcafeconfig from "../acceptance/testcafeconfig";
import Menu from "../page-model/menu";
import Photo from "../page-model/photo";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import PhotoViews from "../page-model/photo-views";
import Label from "../page-model/label";
import Album from "../page-model/album";
import Subject from "../page-model/subject";
import Page from "../page-model/page";

fixture`Test admin role`.page`${testcafeconfig.url}`;

const menu = new Menu();
const photo = new Photo();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const photoviews = new PhotoViews();
const label = new Label();
const album = new Album();
const subject = new Subject();
const page = new Page();

const getLocation = ClientFunction(() => document.location.href);

test.meta("testID", "authentication-000")(
  "Time to start instance (will be marked as unstable)",
  async (t) => {
    await t.wait(5000);
  }
);

test.meta("testID", "admin-role-001")("Access to settings", async (t) => {
  await page.login("admin", "photoprism");
  await menu.checkMenuItemAvailability("settings", true);
  await t.navigateTo("/settings");
  await t
    .wait(5000)
    .expect(Selector(".input-language input", { timeout: 8000 }).visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk()
    .navigateTo("/settings/library")
    .expect(Selector("form.p-form-settings").visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk()
    .navigateTo("/settings/advanced")
    .expect(Selector("label").withText("Read-Only Mode").visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk()
    .navigateTo("/settings/sync")
    .expect(Selector("div.p-accounts-list").visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk();
});

test.meta("testID", "admin-role-002")("Access to archive", async (t) => {
  await page.login("admin", "photoprism");
  const PhotoCountBrowse = await photo.getPhotoCount("all");
  await menu.checkMenuItemAvailability("archive", true);
  await t.navigateTo("/archive");
  await photo.checkPhotoVisibility("pqnahct2mvee8sr4", true);
  const PhotoCountArchive = await photo.getPhotoCount("all");
  await t.expect(PhotoCountBrowse).gte(PhotoCountArchive);
});

test.meta("testID", "admin-role-003")("Access to review", async (t) => {
  await page.login("admin", "photoprism");
  const PhotoCountBrowse = await photo.getPhotoCount("all");
  await menu.checkMenuItemAvailability("review", true);
  await t.navigateTo("/review");
  await photo.checkPhotoVisibility("pqzuein2pdcg1kc7", true);
  const PhotoCountReview = await photo.getPhotoCount("all");
  await t.expect(PhotoCountBrowse).gte(PhotoCountReview);
});

test.meta("testID", "admin-role-004")("Access to private", async (t) => {
  await page.login("admin", "photoprism");
  const PhotoCountBrowse = await photo.getPhotoCount("all");
  await menu.checkMenuItemAvailability("private", true);
  await t.navigateTo("/private");
  await photo.checkPhotoVisibility("pqmxlquf9tbc8mk2", true);
  const PhotoCountPrivate = await photo.getPhotoCount("all");
  await t.expect(PhotoCountBrowse).gte(PhotoCountPrivate);
});

test.meta("testID", "admin-role-005")("Access to library", async (t) => {
  await page.login("admin", "photoprism");
  const PhotoCountBrowse = await photo.getPhotoCount("all");
  await menu.checkMenuItemAvailability("library", true);
  await t.navigateTo("/library");
  await t
    .expect(Selector(".input-index-folder input").visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk()
    .navigateTo("/library/import")
    .expect(Selector(".input-import-folder input").visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk()
    .navigateTo("/library/logs")
    .expect(Selector("div.terminal").visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk();
  await menu.checkMenuItemAvailability("originals", true);
  await t
    .navigateTo("/library/files")
    .expect(Selector("div.p-page-files").visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk();
  await menu.checkMenuItemAvailability("hidden", true);
  await t.navigateTo("/library/hidden");
  const PhotoCountHidden = await photo.getPhotoCount("all");
  await t.expect(PhotoCountBrowse).gte(PhotoCountHidden);
  await menu.checkMenuItemAvailability("errors", true);
  await t
    .navigateTo("/library/errors")
    .expect(Selector("div.p-page-errors").visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk();
});

test.meta("testID", "admin-role-006")("private/archived photos in search results", async (t) => {
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
});

test.meta("testID", "admin-role-007")("Upload functionality", async (t) => {
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

test.meta("testID", "admin-role-008")(
  "Admin can private, archive, share, add/remove to album",
  async (t) => {
    await page.login("admin", "photoprism");
    const FirstPhoto = await photo.getNthPhotoUid("image", 0);
    await photo.selectPhotoFromUID(FirstPhoto);
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
    await toolbar.triggerToolbarAction("list-view", "");
    await photo.checkListViewActionAvailability("private", false);
  }
);

test.meta("testID", "admin-role-009")("Admin can approve low quality photos", async (t) => {
  await page.login("admin", "photoprism");
  await toolbar.search('quality:0 name:"photos-013_1"');
  await photo.toggleSelectNthPhoto(0, "all");
  await contextmenu.triggerContextMenuAction("edit", "", "");
  await t.expect(Selector("button.action-approve").visible).ok();
});

test.meta("testID", "admin-role-010")("Edit dialog is not read only for admin", async (t) => {
  await page.login("admin", "photoprism");
  await toolbar.search("faces:new");
  //details
  const FirstPhoto = await photo.getNthPhotoUid("image", 0);
  await photo.selectPhotoFromUID(FirstPhoto);
  await contextmenu.triggerContextMenuAction("edit", "", "");
  await t
    .expect(Selector(".input-title input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-local-time input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-day input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-month input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-year input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-timezone input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-latitude input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-longitude input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-altitude input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-country input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-camera input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-iso input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-exposure input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-lens input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-fnumber input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-focal-length input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-subject textarea").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-artist input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-copyright input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-license textarea").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-description textarea").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-keywords textarea").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-notes textarea").hasAttribute("disabled"))
    .notOk()
    .expect(Selector("button.action-apply").visible)
    .ok();
  if (t.browser.platform !== "mobile") {
    await t.expect(Selector("button.action-done").visible).ok();
  }
  //labels
  await t
    .click(Selector("#tab-labels"))
    .expect(Selector("button.action-remove").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-label input").exists)
    .ok()
    .expect(Selector("button.p-photo-label-add").exists)
    .ok()
    .click(Selector("div.p-inline-edit"))
    .expect(Selector(".input-rename input").exists)
    .ok();
  //people
  await t
    .click(Selector("#tab-people"))
    .expect(Selector(".input-name input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector("button.input-reject").exists)
    .ok();
  //info
  await t
    .click(Selector("#tab-info"))
    .expect(Selector(".input-favorite input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-private input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-scan input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-panorama input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-stackable input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-type input").hasAttribute("disabled"))
    .notOk();
});

test.meta("testID", "admin-role-011")("Edit labels functionality", async (t) => {
  await page.login("admin", "photoprism");
  await menu.openPage("labels");
  const FirstLabel = await label.getNthLabeltUid(0);
  await label.checkHoverActionState("uid", FirstLabel, "favorite", false);
  await label.triggerHoverAction("uid", FirstLabel, "favorite");
  await label.checkHoverActionState("uid", FirstLabel, "favorite", true);
  await label.triggerHoverAction("uid", FirstLabel, "favorite");
  await label.checkHoverActionState("uid", FirstLabel, "favorite", false);
  await t
    .click(Selector(`a.uid-${FirstLabel} div.inline-edit`))
    .expect(Selector(".input-rename input").visible)
    .ok();
  await label.selectLabelFromUID(FirstLabel);
  await contextmenu.checkContextMenuActionAvailability("delete", true);
  await contextmenu.checkContextMenuActionAvailability("album", true);
});

test.meta("testID", "admin-role-012")("Edit album functionality", async (t) => {
  await page.login("admin", "photoprism");
  await menu.openPage("albums");
  await toolbar.checkToolbarActionAvailability("add", true);
  await t.expect(Selector("a.is-album button.action-share").visible).ok();
  const FirstAlbum = await album.getNthAlbumUid("all", 0);
  await album.selectAlbumFromUID(FirstAlbum);
  await contextmenu.checkContextMenuActionAvailability("edit", true);
  await contextmenu.checkContextMenuActionAvailability("share", true);
  await contextmenu.checkContextMenuActionAvailability("clone", true);
  await contextmenu.checkContextMenuActionAvailability("download", true);
  await contextmenu.checkContextMenuActionAvailability("delete", true);
  await contextmenu.clearSelection();
  await t
    .click(page.cardTitle)
    .expect(Selector(".input-description textarea").visible)
    .ok()
    .click(Selector("button.action-cancel"));
  if (await Selector(`a.uid-${FirstAlbum}`).hasClass("is-favorite")) {
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", true);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", false);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", true);
  } else {
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", false);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", true);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", false);
  }
  await album.openNthAlbum(0);
  await toolbar.checkToolbarActionAvailability("share", true);
  await toolbar.checkToolbarActionAvailability("edit", true);

  await photo.toggleSelectNthPhoto(0, "all");
  await contextmenu.checkContextMenuActionAvailability("album", true);
  await contextmenu.checkContextMenuActionAvailability("private", true);
  await contextmenu.checkContextMenuActionAvailability("share", true);
  await contextmenu.checkContextMenuActionAvailability("remove", true);
});

test.meta("testID", "admin-role-013")("Edit moment functionality", async (t) => {
  await page.login("admin", "photoprism");
  await menu.openPage("moments");
  await t.expect(Selector("a.is-album button.action-share").visible).ok();
  const FirstAlbum = await album.getNthAlbumUid("moment", 0);
  await album.selectAlbumFromUID(FirstAlbum);
  await contextmenu.checkContextMenuActionAvailability("edit", true);
  await contextmenu.checkContextMenuActionAvailability("share", true);
  await contextmenu.checkContextMenuActionAvailability("clone", true);
  await contextmenu.checkContextMenuActionAvailability("download", true);
  await contextmenu.checkContextMenuActionAvailability("delete", true);

  await contextmenu.clearSelection();
  await t
    .click(page.cardTitle)
    .expect(Selector(".input-description textarea").visible)
    .ok()
    .click(Selector("button.action-cancel"));
  if (await Selector(`a.uid-${FirstAlbum}`).hasClass("is-favorite")) {
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", true);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", false);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", true);
  } else {
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", false);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", true);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", false);
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

test.meta("testID", "admin-role-014")("Edit state functionality", async (t) => {
  await page.login("admin", "photoprism");
  await menu.openPage("states");
  await t.expect(Selector("a.is-album button.action-share").visible).ok();
  const FirstAlbum = await album.getNthAlbumUid("state", 0);
  await album.selectAlbumFromUID(FirstAlbum);
  await contextmenu.checkContextMenuActionAvailability("edit", true);
  await contextmenu.checkContextMenuActionAvailability("share", true);
  await contextmenu.checkContextMenuActionAvailability("clone", true);
  await contextmenu.checkContextMenuActionAvailability("download", true);
  await contextmenu.checkContextMenuActionAvailability("delete", true);
  await contextmenu.clearSelection();
  await t
    .click(page.cardTitle)
    .expect(Selector(".input-description textarea").visible)
    .ok()
    .click(Selector("button.action-cancel"));
  if (await Selector(`a.uid-${FirstAlbum}`).hasClass("is-favorite")) {
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", true);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", false);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", true);
  } else {
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", false);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", true);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", false);
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

test.meta("testID", "admin-role-015")("Edit calendar functionality", async (t) => {
  await page.login("admin", "photoprism");
  await menu.openPage("calendar");
  await t.expect(Selector("a.is-album button.action-share").visible).ok();
  const FirstAlbum = await album.getNthAlbumUid("month", 0);
  await album.selectAlbumFromUID(FirstAlbum);
  await contextmenu.checkContextMenuActionAvailability("edit", true);
  await contextmenu.checkContextMenuActionAvailability("share", true);
  await contextmenu.checkContextMenuActionAvailability("clone", true);
  await contextmenu.checkContextMenuActionAvailability("download", true);
  await contextmenu.checkContextMenuActionAvailability("delete", false);
  await contextmenu.clearSelection();
  await t
    .click(page.cardTitle)
    .expect(Selector(".input-description textarea").visible)
    .ok()
    .click(Selector("button.action-cancel"));
  if (await Selector(`a.uid-${FirstAlbum}`).hasClass("is-favorite")) {
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", true);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", false);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", true);
  } else {
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", false);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", true);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", false);
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
  await t.expect(Selector("a.is-album button.action-share").visible).ok();
  const FirstAlbum = await album.getNthAlbumUid("folder", 0);
  await album.selectAlbumFromUID(FirstAlbum);
  await contextmenu.checkContextMenuActionAvailability("edit", true);
  await contextmenu.checkContextMenuActionAvailability("share", true);
  await contextmenu.checkContextMenuActionAvailability("clone", true);
  await contextmenu.checkContextMenuActionAvailability("download", true);
  await contextmenu.checkContextMenuActionAvailability("delete", false);
  await contextmenu.clearSelection();
  await t
    .click(page.cardTitle)
    .expect(Selector(".input-description textarea").visible)
    .ok()
    .click(Selector("button.action-cancel"));
  if (await Selector(`a.uid-${FirstAlbum}`).hasClass("is-favorite")) {
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", true);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", false);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", true);
  } else {
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", false);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", true);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", false);
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

test.meta("testID", "admin-role-017")("Edit people functionality", async (t) => {
  await page.login("admin", "photoprism");
  await menu.openPage("people");
  await toolbar.checkToolbarActionAvailability("show-hidden", true);
  await t.expect(Selector("#tab-people_faces > a").exists).ok();
  await subject.checkSubjectVisibility("name", "Otto Visible", true);
  await subject.checkSubjectVisibility("name", "Monika Hide", false);
  await toolbar.triggerToolbarAction("show-hidden", "");
  await subject.checkSubjectVisibility("name", "Otto Visible", true);
  await subject.checkSubjectVisibility("name", "Monika Hide", true);
  await t
    .click(Selector("a div.v-card__title").nth(0))
    .expect(Selector("div.input-rename input").visible)
    .ok();
  await subject.checkHoverActionAvailability("nth", 0, "hidden", true);
  await subject.toggleSelectNthSubject(0);
  await contextmenu.checkContextMenuActionAvailability("album", "true");
  await contextmenu.clearSelection();
  const FirstSubject = await subject.getNthSubjectUid(0);
  if (await Selector(`a.uid-${FirstSubject}`).hasClass("is-favorite")) {
    await subject.checkHoverActionState("uid", FirstSubject, "favorite", true);
    await subject.triggerHoverAction("uid", FirstSubject, "favorite");
    await subject.checkHoverActionState("uid", FirstSubject, "favorite", false);
    await subject.triggerHoverAction("uid", FirstSubject, "favorite");
    await subject.checkHoverActionState("uid", FirstSubject, "favorite", true);
  } else {
    await subject.checkHoverActionState("uid", FirstSubject, "favorite", false);
    await subject.triggerHoverAction("uid", FirstSubject, "favorite");
    await subject.checkHoverActionState("uid", FirstSubject, "favorite", true);
    await subject.triggerHoverAction("uid", FirstSubject, "favorite");
    await subject.checkHoverActionState("uid", FirstSubject, "favorite", false);
  }
  await subject.openNthSubject(0);
  await photo.toggleSelectNthPhoto(0, "all");
  await contextmenu.triggerContextMenuAction("edit", "", "");
  await t
    .click(Selector("#tab-people"))
    .expect(Selector(".input-name input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector("div.v-input__icon--clear > i").hasClass("v-icon--disabled"))
    .notOk()
    .navigateTo("/people/new")
    .expect(Selector("div.is-face").visible)
    .ok();
});
