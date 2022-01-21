import { Selector } from "testcafe";
import testcafeconfig from "./testcafeconfig";
import Menu from "../page-model/menu";
import Album from "../page-model/album";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import Photo from "../page-model/photo";
import PhotoViewer from "../page-model/photoviewer";
import NewPage from "../page-model/page";

fixture`Test albums`.page`${testcafeconfig.url}`;

const menu = new Menu();
const album = new Album();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const photo = new Photo();
const photoviewer = new PhotoViewer();
const newpage = new NewPage();

test.meta("testID", "authentication-000")(
  "Time to start instance (will be marked as unstable)",
  async (t) => {
    await t.wait(5000);
  }
);

test.meta("testID", "albums-001")("Create/delete album on /albums", async (t) => {
  await menu.openPage("albums");
  const countAlbums = await album.getAlbumCount("all");
  await toolbar.triggerToolbarAction("add", "");
  const countAlbumsAfterCreate = await album.getAlbumCount("all");
  const NewAlbum = await album.getNthAlbumUid("all", 0);
  await t.expect(countAlbumsAfterCreate).eql(countAlbums + 1);
  await album.selectAlbumFromUID(NewAlbum);
  await contextmenu.triggerContextMenuAction("delete", "", "");
  const countAlbumsAfterDelete = await album.getAlbumCount("all");
  await t.expect(countAlbumsAfterDelete).eql(countAlbumsAfterCreate - 1);
});

test.meta("testID", "albums-002")("Update album", async (t) => {
  await menu.openPage("albums");
  await toolbar.search("Holiday");
  const AlbumUid = await album.getNthAlbumUid("all", 0);
  await t
    .expect(newpage.cardTitle.nth(0).innerText)
    .contains("Holiday")
    .click(newpage.cardTitle.nth(0))
    .typeText(Selector(".input-title input"), "Animals", { replace: true })
    .expect(Selector(".input-description textarea").value)
    .eql("")
    .expect(Selector(".input-category input").value)
    .eql("")
    .typeText(Selector(".input-description textarea"), "All my animals")
    .typeText(Selector(".input-category input"), "Pets")
    .pressKey("enter")
    .click(".action-confirm");
  await album.openNthAlbum(0);
  const PhotoCount = await photo.getPhotoCount("all");
  await t.expect(toolbar.toolbarTitle.innerText).contains("Animals");
  await t.expect(toolbar.toolbarDescription.innerText).contains("All my animals");
  await menu.openPage("browse");
  await toolbar.search("photo:true");
  const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
  const SecondPhotoUid = await photo.getNthPhotoUid("image", 1);
  await photo.selectPhotoFromUID(SecondPhotoUid);
  await photoviewer.openPhotoViewer("uid", FirstPhotoUid);
  await photoviewer.triggerPhotoViewerAction("select");
  await photoviewer.triggerPhotoViewerAction("close");
  await contextmenu.triggerContextMenuAction("album", "Animals", "album");
  await menu.openPage("albums");
  if (t.browser.platform === "mobile") {
    await toolbar.search("category:Family");
  } else {
    await toolbar.setFilter("category", "Family");
  }
  await t.expect(newpage.cardTitle.nth(0).innerText).contains("Christmas");
  await menu.openPage("albums");
  await toolbar.triggerToolbarAction("reload", "");
  if (t.browser.platform === "mobile") {
  } else {
    await toolbar.setFilter("category", "All Categories");
  }
  await album.openAlbumWithUid(AlbumUid);
  const PhotoCountAfterAdd = await photo.getPhotoCount("all");
  await t.expect(PhotoCountAfterAdd).eql(PhotoCount + 2);
  await photo.selectPhotoFromUID(FirstPhotoUid);
  await photo.selectPhotoFromUID(SecondPhotoUid);
  await contextmenu.triggerContextMenuAction("remove", "", "");
  const PhotoCountAfterDelete = await photo.getPhotoCount("all");
  await t.expect(PhotoCountAfterDelete).eql(PhotoCountAfterAdd - 2);
  await toolbar.triggerToolbarAction("edit", "");
  await t
    .typeText(Selector(".input-title input"), "Holiday", { replace: true })
    .expect(Selector(".input-description textarea").value)
    .eql("All my animals")
    .expect(Selector(".input-category input").value)
    .eql("Pets")
    .click(Selector(".input-description textarea"))
    .pressKey("ctrl+a delete")
    .pressKey("enter")
    .click(Selector(".input-category input"))
    .pressKey("ctrl+a delete")
    .pressKey("enter")
    .click(".action-confirm");
  await menu.openPage("albums");
  await t
    .expect(Selector("div").withText("Holiday").visible)
    .ok()
    .expect(Selector("div").withText("Animals").exists)
    .notOk();
});

//TODO test that sharing link works as expected --> move to sharing.js
test.meta("testID", "albums-006")("Create, Edit, delete sharing link", async (t) => {
  await newpage.testCreateEditDeleteSharingLink("albums");
});

test.meta("testID", "albums-007")("Create/delete album during add to album", async (t) => {
  await menu.openPage("albums");
  const countAlbums = await album.getAlbumCount("all");
  await menu.openPage("browse");
  await toolbar.search("photo:true");
  const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
  const SecondPhotoUid = await photo.getNthPhotoUid("image", 1);
  await photo.selectPhotoFromUID(SecondPhotoUid);
  await photo.selectPhotoFromUID(FirstPhotoUid);
  await contextmenu.triggerContextMenuAction("album", "NotYetExistingAlbum", "album");
  await menu.openPage("albums");
  const countAlbumsAfterCreation = await album.getAlbumCount("all");
  await t.expect(countAlbumsAfterCreation).eql(countAlbums + 1);
  await toolbar.search("NotYetExistingAlbum");
  const AlbumUid = await album.getNthAlbumUid("all", 0);
  await album.selectAlbumFromUID(AlbumUid);
  await contextmenu.triggerContextMenuAction("delete", "", "");
  await menu.openPage("albums");
  const countAlbumsAfterDelete = await album.getAlbumCount("all");
  await t.expect(countAlbumsAfterDelete).eql(countAlbums);
});

test.meta("testID", "albums-008")("Test album autocomplete", async (t) => {
  await toolbar.search("photo:true");
  const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
  await photo.selectPhotoFromUID(FirstPhotoUid);
  await contextmenu.openContextMenu();
  await t
    .click(Selector("button.action-album"))
    .click(Selector(".input-album input"))
    .expect(newpage.selectOption.withText("Holiday").visible)
    .ok()
    .expect(newpage.selectOption.withText("Christmas").visible)
    .ok()
    .typeText(Selector(".input-album input"), "C", { replace: true })
    .expect(newpage.selectOption.withText("Holiday").visible)
    .notOk()
    .expect(newpage.selectOption.withText("Christmas").visible)
    .ok()
    .expect(newpage.selectOption.withText("C").visible)
    .ok();
});
