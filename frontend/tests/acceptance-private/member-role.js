import { Selector } from "testcafe";
import testcafeconfig from "../acceptance/testcafeconfig";
import { ClientFunction } from "testcafe";
import Menu from "../page-model/menu";
import Photo from "../page-model/photo";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import PhotoViews from "../page-model/photo-views";
import Label from "../page-model/label";
import Album from "../page-model/album";
import Subject from "../page-model/subject";
import PhotoViewer from "../page-model/photoviewer";
import PhotoEdit from "../page-model/photo-edit";
import NewPage from "../page-model/page";

fixture`Test member role`.page`${testcafeconfig.url}`;

const menu = new Menu();
const photo = new Photo();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const photoviews = new PhotoViews();
const label = new Label();
const album = new Album();
const subject = new Subject();
const photoviewer = new PhotoViewer();
const photoedit = new PhotoEdit();
const newpage = new NewPage();
const getLocation = ClientFunction(() => document.location.href);

test.meta("testID", "member-role-001")("No access to settings", async (t) => {
  await newpage.login("member", "passwdmember");
  await menu.checkMenuItemAvailability("settings", false);
  await t.navigateTo("/settings");
  await t
    .expect(Selector(".input-language input", { timeout: 8000 }).visible)
    .notOk()
    .expect(Selector("div.p-page-photos").visible)
    .ok()
    .navigateTo("/settings/library")
    .expect(Selector("form.p-form-settings").visible)
    .notOk()
    .expect(Selector("div.p-page-photos").visible)
    .ok()
    .navigateTo("/settings/advanced")
    .expect(Selector("label").withText("Read-Only Mode").visible)
    .notOk()
    .expect(Selector("div.p-page-photos").visible)
    .ok()
    .navigateTo("/settings/sync")
    .expect(Selector("div.p-accounts-list").visible)
    .notOk()
    .expect(Selector("div.p-page-photos").visible)
    .ok();
});

test.meta("testID", "member-role-002")("No access to archive", async (t) => {
  await newpage.login("member", "passwdmember");
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
  await newpage.login("member", "passwdmember");
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
  await newpage.login("member", "passwdmember");
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
  await newpage.login("member", "passwdmember");
  if (t.browser.platform === "mobile") {
    await t.wait(5000);
  }
  const PhotoCountBrowse = await photo.getPhotoCount("all");
  await menu.checkMenuItemAvailability("library", false);
  await t.navigateTo("/library");
  await t
    .expect(Selector(".input-index-folder input").visible)
    .notOk()
    .expect(Selector("div.p-page-photos").visible)
    .ok()
    .navigateTo("/library/import")
    .expect(Selector(".input-import-folder input").visible)
    .notOk()
    .expect(Selector("div.p-page-photos").visible)
    .ok()
    .navigateTo("/library/logs")
    .expect(Selector("p.p-log-message").visible)
    .notOk()
    .expect(Selector("div.p-page-photos").visible)
    .ok();
  await menu.checkMenuItemAvailability("originals", false);
  await t
    .navigateTo("/library/files")
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
  await t
    .navigateTo("/library/errors")
    .expect(Selector("div.p-page-errors").visible)
    .notOk()
    .expect(Selector("div.p-page-photos").visible)
    .ok();
});

test.meta("testID", "member-role-006")(
  "No private/archived photos in search results",
  async (t) => {
    await newpage.login("member", "passwdmember");
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
  await newpage.login("member", "passwdmember");
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
  await newpage.login("member", "passwdmember");
  await t.wait(5000);
  const FirstPhoto = await photo.getNthPhotoUid("image", 0);
  const SecondPhoto = await photo.getNthPhotoUid("all", 1);
  await menu.openPage("favorites");
  const FirstFavorite = await photo.getNthPhotoUid("image", 0);
  await photoviews.checkHoverActionState("uid", FirstFavorite, "favorite", true);
  await photoviews.triggerHoverAction("uid", FirstFavorite, "favorite");
  await photoviews.checkHoverActionState("uid", FirstFavorite, "favorite", true);
  await t
    .expect(Selector("div").withAttribute("data-uid", FirstPhoto).exists, { timeout: 5000 })
    .notOk()
    .expect(Selector("div").withAttribute("data-uid", SecondPhoto).exists, { timeout: 5000 })
    .notOk();
  await menu.openPage("browse");
  await photoviewer.openPhotoViewer("uid", FirstPhoto);
  await t.expect(Selector("#photo-viewer").visible).ok();
  await photoviewer.checkPhotoViewerActionAvailability("like", false);
  await photoviewer.checkPhotoViewerActionAvailability("close", true);
  await photoviewer.triggerPhotoViewerAction("close");
  await t.wait(5000).click(Selector(".p-expand-search", { timeout: 10000 }));
  await toolbar.setFilter("view", "Cards");
  await photoviews.checkHoverActionState("uid", FirstPhoto, "favorite", false);
  await photoviews.triggerHoverAction("uid", FirstPhoto, "favorite");
  await photoviews.checkHoverActionState("uid", FirstPhoto, "favorite", false);
  await photo.selectPhotoFromUID(SecondPhoto);
  await contextmenu.triggerContextMenuAction("edit", "", "");
  await t
    .click("#tab-info")
    .expect(Selector(".input-favorite input").hasAttribute("disabled"))
    .ok();
  await photoedit.turnSwitchOn("favorite");
  await t.click(Selector(".action-close"));
  await contextmenu.clearSelection();
  await photoviews.checkHoverActionState("uid", SecondPhoto, "favorite", false);
  await menu.openPage("browse");
  await toolbar.setFilter("view", "Mosaic");
  await photoviews.checkHoverActionState("uid", FirstPhoto, "favorite", false);
  await photoviews.triggerHoverAction("uid", FirstPhoto, "favorite");
  await photoviews.checkHoverActionState("uid", FirstPhoto, "favorite", false);
  await toolbar.setFilter("view", "List");
  await photoviews.checkListViewActionAvailability("like", true);
  await photoviews.triggerListViewActions("nth", 0, "like");
  await toolbar.setFilter("view", "Cards");
  await photoviews.checkHoverActionState("uid", FirstPhoto, "favorite", false);
  await menu.openPage("albums");
  await t.click(Selector("a.is-album").nth(0));
  await toolbar.triggerToolbarAction("list-view", "");
  if (t.browser.platform === "mobile") {
    await toolbar.triggerToolbarAction("list-view", "");
  }
  await photoviews.checkListViewActionAvailability("like", true);
});

test.meta("testID", "member-role-009")(
  "Member cannot private, archive, share, add/remove to album",
  async (t) => {
    await newpage.login("member", "passwdmember");
    const FirstPhoto = await photo.getNthPhotoUid("image", 0);
    await photo.selectPhotoFromUID(FirstPhoto);
    await contextmenu.checkContextMenuActionAvailability("private", false);
    await contextmenu.checkContextMenuActionAvailability("archive", false);
    await contextmenu.checkContextMenuActionAvailability("share", false);
    await contextmenu.checkContextMenuActionAvailability("album", false);
    await contextmenu.checkContextMenuActionAvailability("download", true);
    await contextmenu.checkContextMenuActionAvailability("edit", true);
    await contextmenu.clearSelection();
    await toolbar.setFilter("view", "List");
    await photoviews.checkListViewActionAvailability("private", true);
    await menu.openPage("albums");
    await album.openNthAlbum(0);
    await toolbar.checkToolbarActionAvailability("share", false);
    await photo.toggleSelectNthPhoto(0, "all");
    await contextmenu.checkContextMenuActionAvailability("private", false);
    await contextmenu.checkContextMenuActionAvailability("remove", false);
    await contextmenu.checkContextMenuActionAvailability("share", false);
    await contextmenu.checkContextMenuActionAvailability("album", false);
    await contextmenu.clearSelection();
    await toolbar.triggerToolbarAction("list-view", "");
    if (t.browser.platform === "mobile") {
      await toolbar.triggerToolbarAction("list-view", "");
    }
    await photoviews.checkListViewActionAvailability("private", true);
  }
);

test.meta("testID", "member-role-010")("Member cannot approve low quality photos", async (t) => {
  await newpage.login("member", "passwdmember");
  await toolbar.search('quality:0 name:"photos-013_1"');
  await photo.toggleSelectNthPhoto(0, "all");
  await contextmenu.triggerContextMenuAction("edit", "", "");
  await t.expect(Selector("button.action-approve").visible).notOk();
});

test.meta("testID", "member-role-011")("Edit dialog is read only for member", async (t) => {
  await newpage.login("member", "passwdmember");
  await toolbar.search("faces:new");
  //details
  const FirstPhoto = await photo.getNthPhotoUid("image", 0);
  await photo.selectPhotoFromUID(FirstPhoto);
  await contextmenu.triggerContextMenuAction("edit", "", "");
  await t
    .expect(Selector(".input-title input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-local-time input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-day input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-month input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-year input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-timezone input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-latitude input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-longitude input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-altitude input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-country input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-camera input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-iso input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-exposure input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-lens input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-fnumber input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-focal-length input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-subject textarea").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-artist input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-copyright input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-license textarea").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-description textarea").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-keywords textarea").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-notes textarea").hasAttribute("disabled"))
    .ok()
    .expect(Selector("button.action-apply").visible)
    .notOk();
  if (t.browser.platform !== "mobile") {
    await t.expect(Selector("button.action-done").visible).notOk();
  }
  //labels
  await t
    .click(Selector("#tab-labels"))
    .expect(Selector("button.action-remove").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-label input").exists)
    .notOk()
    .expect(Selector("button.p-photo-label-add").exists)
    .notOk()
    .click(Selector("div.p-inline-edit"))
    .expect(Selector(".input-rename input").exists)
    .notOk();
  //people
  await t
    .click(Selector("#tab-people"))
    .expect(Selector(".input-name input").hasAttribute("disabled"))
    .ok()
    .expect(Selector("button.input-reject").exists)
    .notOk();
  //info
  await t
    .click(Selector("#tab-info"))
    .expect(Selector(".input-favorite input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-private input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-scan input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-panorama input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-stackable input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-type input").hasAttribute("disabled"))
    .ok();
});

test.meta("testID", "member-role-012")("No edit album functionality", async (t) => {
  await newpage.login("member", "passwdmember");
  await menu.openPage("albums");
  await toolbar.checkToolbarActionAvailability("add", false);
  await t.expect(Selector("a.is-album button.action-share").visible).notOk();
  const FirstAlbum = await album.getNthAlbumUid("all", 0);
  await album.selectAlbumFromUID(FirstAlbum);
  await contextmenu.checkContextMenuActionAvailability("edit", false);
  await contextmenu.checkContextMenuActionAvailability("share", false);
  await contextmenu.checkContextMenuActionAvailability("clone", false);
  await contextmenu.checkContextMenuActionAvailability("download", true);
  await contextmenu.checkContextMenuActionAvailability("delete", false);

  await contextmenu.clearSelection();
  await t
    .click(newpage.cardTitle)
    .expect(Selector(".input-description textarea").visible)
    .notOk();
  if (await Selector(`a.uid-${FirstAlbum}`).hasClass("is-favorite")) {
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", true);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", true);
  } else {
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", false);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", false);
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
  await newpage.login("member", "passwdmember");
  await menu.openPage("moments");
  await t.expect(Selector("a.is-album button.action-share").visible).notOk();
  const FirstAlbum = await album.getNthAlbumUid("all", 0);
  await album.selectAlbumFromUID(FirstAlbum);
  await contextmenu.checkContextMenuActionAvailability("edit", false);
  await contextmenu.checkContextMenuActionAvailability("share", false);
  await contextmenu.checkContextMenuActionAvailability("clone", false);
  await contextmenu.checkContextMenuActionAvailability("download", true);
  await contextmenu.checkContextMenuActionAvailability("delete", false);

  await contextmenu.clearSelection();
  await t
    .click(newpage.cardTitle)
    .expect(Selector(".input-description textarea").visible)
    .notOk();
  if (await Selector(`a.uid-${FirstAlbum}`).hasClass("is-favorite")) {
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", true);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", true);
  } else {
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", false);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", false);
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
  await newpage.login("member", "passwdmember");
  await menu.openPage("states");
  await t.expect(Selector("a.is-album button.action-share").visible).notOk();
  const FirstAlbum = await album.getNthAlbumUid("all", 0);
  await album.selectAlbumFromUID(FirstAlbum);
  await contextmenu.checkContextMenuActionAvailability("edit", false);
  await contextmenu.checkContextMenuActionAvailability("share", false);
  await contextmenu.checkContextMenuActionAvailability("clone", false);
  await contextmenu.checkContextMenuActionAvailability("download", true);
  await contextmenu.checkContextMenuActionAvailability("delete", false);

  await contextmenu.clearSelection();
  await t
    .click(newpage.cardTitle)
    .expect(Selector(".input-description textarea").visible)
    .notOk();
  if (await Selector(`a.uid-${FirstAlbum}`).hasClass("is-favorite")) {
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", true);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", true);
  } else {
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", false);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", false);
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
  await newpage.login("member", "passwdmember");
  await menu.openPage("calendar");
  await t.expect(Selector("a.is-album button.action-share").visible).notOk();
  const FirstAlbum = await album.getNthAlbumUid("all", 0);
  await album.selectAlbumFromUID(FirstAlbum);
  await contextmenu.checkContextMenuActionAvailability("edit", false);
  await contextmenu.checkContextMenuActionAvailability("share", false);
  await contextmenu.checkContextMenuActionAvailability("clone", false);
  await contextmenu.checkContextMenuActionAvailability("download", true);
  await contextmenu.checkContextMenuActionAvailability("delete", false);

  await contextmenu.clearSelection();
  await t
    .click(newpage.cardTitle)
    .expect(Selector(".input-description textarea").visible)
    .notOk();
  if (await Selector(`a.uid-${FirstAlbum}`).hasClass("is-favorite")) {
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", true);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", true);
  } else {
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", false);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", false);
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
  await newpage.login("member", "passwdmember");
  await menu.openPage("folders");
  await t.expect(Selector("a.is-album button.action-share").visible).notOk();
  const FirstAlbum = await album.getNthAlbumUid("all", 0);
  await album.selectAlbumFromUID(FirstAlbum);
  await contextmenu.checkContextMenuActionAvailability("edit", false);
  await contextmenu.checkContextMenuActionAvailability("share", false);
  await contextmenu.checkContextMenuActionAvailability("clone", false);
  await contextmenu.checkContextMenuActionAvailability("download", true);
  await contextmenu.checkContextMenuActionAvailability("delete", false);

  await contextmenu.clearSelection();
  await t
    .click(newpage.cardTitle)
    .expect(Selector(".input-description textarea").visible)
    .notOk();
  if (await Selector(`a.uid-${FirstAlbum}`).hasClass("is-favorite")) {
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", true);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", true);
  } else {
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", false);
    await album.triggerHoverAction("uid", FirstAlbum, "favorite");
    await album.checkHoverActionState("uid", FirstAlbum, "favorite", false);
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
  await newpage.login("member", "passwdmember");
  await menu.openPage("labels");
  const FirstLabel = await label.getNthLabeltUid(0);
  await label.checkHoverActionState("uid", FirstLabel, "favorite", false);
  await label.triggerHoverAction("uid", FirstLabel, "favorite");
  await label.checkHoverActionState("uid", FirstLabel, "favorite", false);
  await t
    .click(Selector(`a.uid-${FirstLabel} div.inline-edit`))
    .expect(Selector(".input-rename input").visible)
    .notOk();
  await label.selectLabelFromUID(FirstLabel);
  await contextmenu.checkContextMenuActionAvailability("delete", false);
  await contextmenu.checkContextMenuActionAvailability("album", false);
});

test.meta("testID", "member-role-018")("No unstack, change primary actions", async (t) => {
  await newpage.login("member", "passwdmember");
  await toolbar.search("stack:true");
  //details
  const FirstPhoto = await photo.getNthPhotoUid("image", 0);
  await photo.selectPhotoFromUID(FirstPhoto);
  await contextmenu.triggerContextMenuAction("edit", "", "");
  await t
    .click(Selector("#tab-files"))
    .expect(Selector("button.action-download").visible)
    .ok()
    .expect(Selector("button.action-download").hasAttribute("disabled"))
    .notOk()
    .click(Selector("li.v-expansion-panel__container").nth(1))
    .expect(Selector("button.action-download").visible)
    .ok()
    .expect(Selector("button.action-download").hasAttribute("disabled"))
    .notOk()
    .expect(Selector("button.action-unstack").visible)
    .notOk()
    .expect(Selector("button.action-primary").visible)
    .notOk()
    .expect(Selector("button.action-delete").visible)
    .notOk();
});

test.meta("testID", "member-role-019")("No edit people functionality", async (t) => {
  await newpage.login("member", "passwdmember");
  await menu.openPage("people");
  await toolbar.checkToolbarActionAvailability("show-hidden", false);
  await toolbar.checkToolbarActionAvailability("upload", false);
  await subject.checkSubjectVisibility("name", "Otto Visible", true);
  await subject.checkSubjectVisibility("name", "Monika Hide", false);
  await t
    .expect(Selector("#tab-people_faces > a").exists)
    .notOk()
    .click(Selector("a div.v-card__title").nth(0))
    .expect(Selector("div.input-rename input").visible)
    .notOk();
  await subject.checkHoverActionAvailability("nth", 0, "hidden", false);
  await subject.toggleSelectNthSubject(0);
  await contextmenu.checkContextMenuActionAvailability("album", "false");
  await contextmenu.checkContextMenuActionAvailability("download", "true");
  await contextmenu.clearSelection();
  const FirstSubject = subject.getNthSubjectUid(0);
  if (await Selector(`a.uid-${FirstSubject}`).hasClass("is-favorite")) {
    await subject.checkHoverActionState("uid", FirstSubject, "favorite", true);
    await subject.triggerHoverAction("uid", FirstSubject, "favorite");
    await subject.checkHoverActionState("uid", FirstSubject, "favorite", true);
  } else {
    await subject.checkHoverActionState("uid", FirstSubject, "favorite", false);
    await subject.triggerHoverAction("uid", FirstSubject, "favorite");
    await subject.checkHoverActionState("uid", FirstSubject, "favorite", false);
  }
  await subject.openNthSubject(0);
  await photo.toggleSelectNthPhoto(0);
  await contextmenu.triggerContextMenuAction("edit", "", "");
  await t
    .click(Selector("#tab-people"))
    .expect(Selector(".input-name input").hasAttribute("disabled"))
    .ok()
    .expect(Selector("div.v-input__icon--clear > i").hasClass("v-icon--disabled"))
    .ok()
    .navigateTo("/people/new")
    .expect(Selector("div.is-face").visible)
    .notOk()
    .expect(Selector("#tab-people_faces > a").exists)
    .notOk()
    .navigateTo("/people?hidden=yes&order=relevance")
    .expect(Selector("a.is-subject").visible)
    .notOk();
});
