import { Selector } from "testcafe";
import testcafeconfig from "./testcafeconfig";
import Menu from "../page-model/menu";
import Album from "../page-model/album";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import Photo from "../page-model/photo";
import PhotoViewer from "../page-model/photoviewer";
import Page from "../page-model/page";

fixture`Test albums`.page`${testcafeconfig.url}`;

const menu = new Menu();
const album = new Album();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const photo = new Photo();
const photoviewer = new PhotoViewer();
const page = new Page();

test.meta("testID", "authentication-000")(
  "Time to start instance (will be marked as unstable)",
  async (t) => {
    await t.wait(5000);
  }
);

test.meta("testID", "albums-001").meta({ type: "smoke" })(
  "Create/delete album on /albums",
  async (t) => {
    await menu.openPage("albums");
    const countAlbums = await album.getAlbumCount("all");
    await toolbar.triggerToolbarAction("add");
    const countAlbumsAfterCreate = await album.getAlbumCount("all");
    const NewAlbumUid = await album.getNthAlbumUid("all", 0);

    await t.expect(countAlbumsAfterCreate).eql(countAlbums + 1);

    await album.selectAlbumFromUID(NewAlbumUid);
    await contextmenu.triggerContextMenuAction("delete", "");
    const countAlbumsAfterDelete = await album.getAlbumCount("all");

    await t.expect(countAlbumsAfterDelete).eql(countAlbumsAfterCreate - 1);
  }
);

test.meta("testID", "albums-002").meta({ type: "smoke" })(
  "Create/delete album during add to album",
  async (t) => {
    await menu.openPage("albums");
    const countAlbums = await album.getAlbumCount("all");
    await menu.openPage("browse");
    await toolbar.search("photo:true");
    const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
    const SecondPhotoUid = await photo.getNthPhotoUid("image", 1);
    await photo.selectPhotoFromUID(SecondPhotoUid);
    await photo.selectPhotoFromUID(FirstPhotoUid);
    await contextmenu.triggerContextMenuAction("album", "NotYetExistingAlbum");
    await menu.openPage("albums");
    const countAlbumsAfterCreation = await album.getAlbumCount("all");

    await t.expect(countAlbumsAfterCreation).eql(countAlbums + 1);

    await toolbar.search("NotYetExistingAlbum");
    const AlbumUid = await album.getNthAlbumUid("all", 0);
    await album.selectAlbumFromUID(AlbumUid);
    await contextmenu.triggerContextMenuAction("delete", "");
    await menu.openPage("albums");
    const countAlbumsAfterDelete = await album.getAlbumCount("all");

    await t.expect(countAlbumsAfterDelete).eql(countAlbums);
  }
);

test.meta("testID", "albums-003").meta({ type: "smoke" })("Update album details", async (t) => {
  await menu.openPage("albums");
  await toolbar.search("Holiday");
  const AlbumUid = await album.getNthAlbumUid("all", 0);

  await t.expect(page.cardTitle.nth(0).innerText).contains("Holiday");

  await t
    .click(page.cardTitle.nth(0))
    .typeText(Selector(".input-title input"), "Animals", { replace: true })
    .expect(Selector(".input-description textarea").value)
    .eql("")
    .expect(Selector(".input-category input").value)
    .eql("")
    .typeText(Selector(".input-description textarea"), "All my animals")
    .typeText(Selector(".input-category input"), "Pets")
    .pressKey("enter")
    .click(".action-confirm");

  await t.expect(page.cardTitle.nth(0).innerText).contains("Animals");

  await album.openAlbumWithUid(AlbumUid);
  await toolbar.triggerToolbarAction("edit");
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

test.meta("testID", "albums-004").meta({ type: "smoke" })(
  "Add/Remove Photos to/from album",
  async (t) => {
    await menu.openPage("albums");
    await toolbar.search("Holiday");
    const AlbumUid = await album.getNthAlbumUid("all", 0);
    await album.openAlbumWithUid(AlbumUid);
    const PhotoCount = await photo.getPhotoCount("all");
    await menu.openPage("browse");
    await toolbar.search("photo:true");
    const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
    const SecondPhotoUid = await photo.getNthPhotoUid("image", 1);
    await photo.selectPhotoFromUID(SecondPhotoUid);
    await photoviewer.openPhotoViewer("uid", FirstPhotoUid);
    await photoviewer.triggerPhotoViewerAction("select");
    await photoviewer.triggerPhotoViewerAction("close");
    await contextmenu.triggerContextMenuAction("album", "Holiday");
    await menu.openPage("albums");
    await album.openAlbumWithUid(AlbumUid);
    const PhotoCountAfterAdd = await photo.getPhotoCount("all");

    await t.expect(PhotoCountAfterAdd).eql(PhotoCount + 2);

    await photo.selectPhotoFromUID(FirstPhotoUid);
    await photo.selectPhotoFromUID(SecondPhotoUid);
    await contextmenu.triggerContextMenuAction("remove", "");
    const PhotoCountAfterDelete = await photo.getPhotoCount("all");

    await t.expect(PhotoCountAfterDelete).eql(PhotoCountAfterAdd - 2);
  }
);

test.meta("testID", "albums-005")("Use album search and filters", async (t) => {
  await menu.openPage("albums");
  if (t.browser.platform === "mobile") {
    await toolbar.search("category:Family");
  } else {
    await toolbar.setFilter("category", "Family");
  }

  await t.expect(page.cardTitle.nth(0).innerText).contains("Christmas");
  const AlbumCount = await album.getAlbumCount("all");
  await t.expect(AlbumCount).eql(1);

  if (t.browser.platform === "mobile") {
  } else {
    await toolbar.setFilter("category", "All Categories");
  }

  await toolbar.search("Holiday");

  await t.expect(page.cardTitle.nth(0).innerText).contains("Holiday");
  const AlbumCount2 = await album.getAlbumCount("all");
  await t.expect(AlbumCount2).eql(1);
});

test.meta("testID", "albums-006")("Test album autocomplete", async (t) => {
  await toolbar.search("photo:true");
  const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
  await photo.selectPhotoFromUID(FirstPhotoUid);
  await contextmenu.openContextMenu();
  await t.click(Selector("button.action-album")).click(Selector(".input-album input"));

  await t
    .expect(page.selectOption.withText("Holiday").visible)
    .ok()
    .expect(page.selectOption.withText("Christmas").visible)
    .ok();

  await t.typeText(Selector(".input-album input"), "C", { replace: true });

  await t
    .expect(page.selectOption.withText("Holiday").visible)
    .notOk()
    .expect(page.selectOption.withText("Christmas").visible)
    .ok()
    .expect(page.selectOption.withText("C").visible)
    .ok();
});

test.meta("testID", "albums-007").meta({ type: "smoke" })(
  "Create, Edit, delete sharing link",
  async (t) => {
    await page.testCreateEditDeleteSharingLink("albums");
  }
);
