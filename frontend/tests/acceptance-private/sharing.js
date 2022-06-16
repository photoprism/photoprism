import { Selector } from "testcafe";
import { Role } from "testcafe";
import testcafeconfig from "../acceptance/testcafeconfig";
import Page from "../page-model/page";
import Menu from "../page-model/menu";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import Album from "../page-model/album";
import PhotoViewer from "../page-model/photoviewer";
import ShareDialog from "../page-model/dialog-share";
import Photo from "../page-model/photo";

fixture`Test link sharing`.page`${testcafeconfig.url}`;

const page = new Page();
const menu = new Menu();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const album = new Album();
const photoviewer = new PhotoViewer();
const sharedialog = new ShareDialog();
const photo = new Photo();

test.skip.meta("testID", "authentication-000")(
  "Time to start instance (will be marked as unstable)",
  async (t) => {
    await t.wait(5000);
  }
);

test.meta("testID", "sharing-001")("Create, view, delete shared albums", async (t) => {
  await page.login("admin", "photoprism");
  await menu.openPage("albums");
  const FirstAlbumUid = await album.getNthAlbumUid("all", 0);
  await album.triggerHoverAction("uid", FirstAlbumUid, "select");
  await contextmenu.checkContextMenuCount("1");
  await contextmenu.triggerContextMenuAction("share", "");
  await t.click(sharedialog.expandLink.nth(0));
  await t
    .typeText(sharedialog.linkSecretInput, "secretForTesting", { replace: true })
    .click(sharedialog.linkExpireInput)
    .click(Selector("div").withText("After 1 day").parent('div[role="listitem"]'))
    .click(sharedialog.dialogSave);
  const Url = await sharedialog.linkUrl.innerText;
  const Expire = await Selector("div.v-select__selections").innerText;

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
  await t.click(sharedialog.expandLink.nth(0));
  await t
    .typeText(sharedialog.linkSecretInput, "secretForTesting", { replace: true })
    .click(sharedialog.linkExpireInput)
    .click(Selector("div").withText("After 1 day").parent('div[role="listitem"]'))
    .click(sharedialog.dialogSave)
    .click(sharedialog.dialogSave);
  await contextmenu.clearSelection();
  await t.navigateTo(url);

  await t.expect(toolbar.toolbarSecondTitle.withText("Christmas").visible).ok();

  await t.click(Selector("button").withText("@photoprism_app"));

  await t.expect(toolbar.toolbarSecondTitle.withText("Albums").visible).ok();

  const AlbumCount = await album.getAlbumCount("all");

  await t.expect(AlbumCount).gte(40);

  await t.useRole(Role.anonymous());
  await t.navigateTo(url);

  await t.expect(toolbar.toolbarSecondTitle.withText("Christmas").visible).ok();
  const photoCountShared = await photo.getPhotoCount("all");
  //don't show private photo
  await t.expect(photoCountShared).eql(1);

  await t.click(Selector("button").withText("@photoprism_app"));

  await t.expect(toolbar.toolbarSecondTitle.withText("Albums").visible).ok();

  const AlbumCountAnonymous = await Selector("a.is-album").count;

  await t.expect(AlbumCountAnonymous).eql(2);

  await t.navigateTo("http://localhost:2343/browse");
  await page.login("admin", "photoprism");
  await menu.openPage("albums");
  await album.openAlbumWithUid(FirstAlbumUid);
  await toolbar.triggerToolbarAction("share");
  await t
    .click(sharedialog.expandLink.nth(0))
    .click(sharedialog.deleteLink)
    .useRole(Role.anonymous());

  await t.expect(Selector(".input-name input").visible).ok();

  await t.navigateTo("http://localhost:2343/s/secretfortesting");

  await t.expect(toolbar.toolbarSecondTitle.withText("Albums").visible).ok();

  const AlbumCountAnonymousAfterDelete = await album.getAlbumCount("all");

  await t.expect(AlbumCountAnonymousAfterDelete).eql(1);

  await t.navigateTo("http://localhost:2343/browse");
  await page.login("admin", "photoprism");
  await menu.openPage("folders");
  await album.openAlbumWithUid(FirstFolderUid);
  await toolbar.triggerToolbarAction("share");
  await t
    .click(sharedialog.expandLink.nth(0))
    .click(sharedialog.deleteLink)
    .useRole(Role.anonymous());

  await t.expect(Selector(".input-name input").visible).ok();

  await t.navigateTo("http://localhost:2343/s/secretfortesting");

  await t
    .expect(toolbar.toolbarSecondTitle.withText("Christmas").visible)
    .notOk()
    .expect(toolbar.toolbarSecondTitle.withText("Albums").visible)
    .notOk()
    .expect(Selector(".input-name input").visible)
    .ok();
});

test.meta("testID", "sharing-002").meta({ type: "smoke" })(
  "Verify anonymous user has limited options",
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

    await t.expect(page.cardTitle.visible).notOk();

    await photoviewer.openPhotoViewer("nth", 0);

    await photoviewer.checkPhotoViewerActionAvailability("download", true);
    await photoviewer.checkPhotoViewerActionAvailability("select", true);
    await photoviewer.checkPhotoViewerActionAvailability("toggle-fullscreen", true);
    await photoviewer.checkPhotoViewerActionAvailability("slideshow", true);
    await photoviewer.checkPhotoViewerActionAvailability("like", false);
    await photoviewer.checkPhotoViewerActionAvailability("edit", false);

    await photoviewer.triggerPhotoViewerAction("close");

    await photo.checkHoverActionAvailability("nth", 0, "favorite", false);
    await photo.checkHoverActionAvailability("nth", 0, "select", true);

    await toolbar.triggerToolbarAction("view-list");

    await t
      .expect(Selector(`td button.input-private`).visible)
      .notOk()
      .expect(Selector(`td button.input-favorite`).visible)
      .notOk()
      .click(Selector("button").withText("@photoprism_app"))
      .expect(toolbar.toolbarSecondTitle.withText("Albums").visible)
      .ok();

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
