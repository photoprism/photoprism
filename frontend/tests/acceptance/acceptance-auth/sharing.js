import { Selector } from "testcafe";
import { Role } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";
import Page from "../page-model/page";
import Menu from "../page-model/menu";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import Album from "../page-model/album";
import PhotoViewer from "../page-model/photoviewer";
import ShareDialog from "../page-model/dialog-share";
import Photo from "../page-model/photo";
import Places from "../page-model/places";

fixture`Test link sharing`.page`${testcafeconfig.url}`;

const page = new Page();
const menu = new Menu();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const album = new Album();
const photoviewer = new PhotoViewer();
const sharedialog = new ShareDialog();
const photo = new Photo();
const places = new Places();

test.meta("testID", "sharing-001").meta({ mode: "auth" })("Common: Create, view, delete shared albums", async (t) => {
  await page.login("admin", "photoprism");
  await menu.openPage("albums");
  const FirstAlbumUid = await album.getNthAlbumUid("all", 0);
  await album.triggerHoverAction("uid", FirstAlbumUid, "select");
  await contextmenu.checkContextMenuCount("1");
  await contextmenu.triggerContextMenuAction("share", "");
  await t
    .typeText(sharedialog.linkSecretInput, "secretForTesting", { replace: true })
    .click(sharedialog.linkExpireInput)
    .click(Selector("div").withText("After 1 day").parent('div[role="option"]'))
    .click(sharedialog.dialogSave);
  const Url = await sharedialog.linkUrl.value;
  const Expire = await Selector(".input-expires .v-select__selection-text").innerText;

  await t.expect(Url).contains("secretfortesting").expect(Expire).contains("After 1 day");
  let url = "http://localhost:2343/s/secretfortesting/christmas";
  await t.click(sharedialog.dialogClose);
  await contextmenu.clearSelection();
  await album.openAlbumWithUid(FirstAlbumUid);
  const photoCount = await photo.getPhotoCount("all");
  await t.expect(photoCount).eql(2);
  await menu.openPage("folders");
  const FirstFolderUid = await album.getNthAlbumUid("all", 0);
  await album.triggerHoverAction("uid", FirstFolderUid, "select");
  await contextmenu.checkContextMenuCount("1");
  await contextmenu.triggerContextMenuAction("share", "");
  await t
    .typeText(sharedialog.linkSecretInput, "secretForTesting", { replace: true })
    .click(sharedialog.linkExpireInput)
    .click(Selector("div").withText("After 1 day").parent('div[role="option"]'))
    .click(sharedialog.dialogSave)
    .click(sharedialog.dialogSave);
  await contextmenu.clearSelection();
  await t.navigateTo(url);

  await t.expect(toolbar.toolbarSecondTitle.withText("Christmas").visible).ok();

  await t.click(Selector("div.v-toolbar-title a").withText("Albums"));
  const AlbumCount = await album.getAlbumCount("all");

  await t.expect(AlbumCount).eql(3);

  await menu.openPage("folders");
  const FolderCount = await album.getAlbumCount("all");

  await t.expect(FolderCount).gte(1);

  await t.useRole(Role.anonymous());
  await t.navigateTo(url);

  await t.expect(toolbar.toolbarSecondTitle.withText("Christmas").visible).ok();

  const photoCountShared = await photo.getPhotoCount("all");
  //don't show private photo
  await t.expect(photoCountShared).eql(1);

  await t.click(Selector("div.v-toolbar-title a").withText("Albums"));
  const AlbumCountAnonymous = await Selector("div.result.is-album").count;

  await t.expect(AlbumCountAnonymous).eql(1);

  await menu.openPage("calendar");
  const CalendarCountAnonymous = await Selector("div.result.is-album").count;

  await t.expect(CalendarCountAnonymous).eql(0);

  await menu.openPage("folders");
  const FolderCountAnonymous = await Selector("div.result.is-album").count;

  await t.expect(FolderCountAnonymous).eql(1);

  await t.navigateTo("http://localhost:2343/library/browse");
  await album.checkAlbumVisibility("aqmxlts2b2rx38wl", true);
  await album.checkAlbumVisibility("aqmxlt22ilujuxux", false);

  await page.logout();
  await page.login("admin", "photoprism");
  await menu.openPage("albums");
  await album.openAlbumWithUid(FirstAlbumUid);
  await toolbar.triggerToolbarAction("share");

  await t.click(sharedialog.deleteLink).useRole(Role.anonymous());

  await t.navigateTo("http://localhost:2343/s/secretfortesting");

  const AlbumCountAnonymousAfterDelete = await album.getAlbumCount("all");

  await t.expect(AlbumCountAnonymousAfterDelete).eql(0);

  await menu.openPage("folders");
  const FolderCountAnonymousAfterDelete = await album.getAlbumCount("all");

  await t.expect(FolderCountAnonymousAfterDelete).eql(1);

  await page.logout();
  await page.login("admin", "photoprism");
  await menu.openPage("folders");
  await album.openAlbumWithUid(FirstFolderUid);
  await toolbar.triggerToolbarAction("share");
  await t.click(sharedialog.deleteLink).useRole(Role.anonymous());

  await t.navigateTo("http://localhost:2343/s/secretfortesting");

  await t
    .expect(toolbar.toolbarSecondTitle.withText("Christmas").visible)
    .notOk()
    .expect(toolbar.toolbarSecondTitle.withText("Albums").visible)
    .notOk()
    .expect(Selector(".input-username input").visible)
    .ok();
});

test.meta("testID", "sharing-002").meta({ type: "short", mode: "auth" })(
  "Multi-Window: Verify visitor role has limited permissions",
  async (t) => {
    await t.navigateTo("http://localhost:2343/s/jxoux5ub1e/british-columbia-canada");
    await t.expect(toolbar.toolbarSecondTitle.withText("British Columbia").visible).ok();

    await toolbar.checkToolbarActionAvailability("edit", false);
    await toolbar.checkToolbarActionAvailability("share", false);
    await toolbar.checkToolbarActionAvailability("upload", false);
    await toolbar.checkToolbarActionAvailability("reload", true);
    await toolbar.checkToolbarActionAvailability("download", true);

    await photo.triggerHoverAction("nth", 0, "select");

    await contextmenu.checkContextMenuActionAvailability("download", true);
    await contextmenu.checkContextMenuActionAvailability("archive", false);
    await contextmenu.checkContextMenuActionAvailability("private", false);
    await contextmenu.checkContextMenuActionAvailability("edit", false);
    await contextmenu.checkContextMenuActionAvailability("share", false);
    await contextmenu.checkContextMenuActionAvailability("album", false);

    await contextmenu.clearSelection();

    await photoviewer.openPhotoViewer("nth", 0);

    await photoviewer.checkPhotoViewerActionAvailability("download-button", true);
    await photoviewer.checkPhotoViewerActionAvailability("select-toggle", true);
    await photoviewer.checkPhotoViewerActionAvailability("fullscreen-toggle", true);
    await photoviewer.checkPhotoViewerActionAvailability("slideshow-toggle", true);
    await photoviewer.checkPhotoViewerActionAvailability("favorite-toggle", false);
    await photoviewer.checkPhotoViewerActionAvailability("edit-button", false);

    await photoviewer.triggerPhotoViewerAction("close");
    await t.expect(Selector("div.p-lightbox__pswp").visible).notOk();

    await photo.checkHoverActionAvailability("nth", 0, "favorite", false);
    await photo.checkHoverActionAvailability("nth", 0, "select", true);

    await toolbar.triggerToolbarAction("view-list");

    await t
      .expect(Selector(`td button.input-private`).visible)
      .notOk()
      .expect(Selector(`td button.input-favorite`).visible)
      .notOk();
    await toolbar.triggerToolbarAction("view-mosaic");
    await toolbar.triggerToolbarAction("view-cards");
    await t.click(page.cardLocation.nth(0));
    await t.expect(places.placesSearch.visible).notOk();
    await t.expect(Selector('div[title="Cape / Bowen Island / 2019"]').visible).ok();
    await t.click(places.zoomOut).click(places.zoomOut).click(places.zoomOut).click(places.zoomOut);
    await t.click(Selector("div.cluster-marker"));
    await t.expect(places.openClusterInSearch.visible).notOk();
    await t.expect(places.closeCluster.visible).ok();

    await t.navigateTo("/library/states");

    const AlbumUid = await album.getNthAlbumUid("all", 0);
    await album.triggerHoverAction("uid", AlbumUid, "select");

    await contextmenu.checkContextMenuActionAvailability("download", true);
    await contextmenu.checkContextMenuActionAvailability("delete", false);
    await contextmenu.checkContextMenuActionAvailability("album", false);
    await contextmenu.checkContextMenuActionAvailability("edit", false);
    await contextmenu.checkContextMenuActionAvailability("share", false);
    await contextmenu.clearSelection();
  }
);
