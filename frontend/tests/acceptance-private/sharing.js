import { Selector } from "testcafe";
import { Role } from "testcafe";
import testcafeconfig from "../acceptance/testcafeconfig";
import NewPage from "../page-model/page";
import Menu from "../page-model/menu";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import PhotoViews from "../page-model/photo-views";
import Album from "../page-model/album";
import PhotoViewer from "../page-model/photoviewer";
import ShareDialog from "../page-model/dialog-share";

fixture`Test link sharing`.page`${testcafeconfig.url}`.skip("Urls are not working anymore");

const newpage = new NewPage();
const menu = new Menu();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const photoviews = new PhotoViews();
const album = new Album();
const photoviewer = new PhotoViewer();
const sharedialog = new ShareDialog();

//TODO merge with other sharing test
test.skip.meta("testID", "authentication-000")(
  "Time to start instance (will be marked as unstable)",
  async (t) => {
    await t.wait(5000);
  }
);

test.skip.meta("testID", "sharing-001")("View shared albums", async (t) => {
  await newpage.login("admin", "photoprism");
  await menu.openPage("albums");
  const FirstAlbum = await album.getNthAlbumUid("all", 0);
  await album.triggerHoverAction("uid", FirstAlbum, "select");
  await contextmenu.checkContextMenuCount("1");
  await contextmenu.triggerContextMenuAction("share", "", "");
  await t.click(sharedialog.expandLink.nth(0));
  await t
    .typeText(sharedialog.linkSecretInput, "secretForTesting", { replace: true })
    .click(sharedialog.linkExpireInput)
    .click(Selector("div").withText("After 1 day").parent('div[role="listitem"]'))
    .click(sharedialog.dialogSave);
  const Url = await sharedialog.linkUrl.value;
  const Expire = await Selector("div.v-select__selections").innerText;
  await t.expect(Url).contains("secretfortesting").expect(Expire).contains("After 1 day");
  let url = Url.replace("2342", "2343");
  await t.click(sharedialog.dialogClose);
  await contextmenu.clearSelection();
  await t.click(Selector(".nav-folders"));
  const FirstFolder = await album.getNthAlbumUid("all", 0);
  await album.triggerHoverAction("uid", FirstFolder, "select");
  await contextmenu.checkContextMenuCount("1");
  await contextmenu.triggerContextMenuAction("share", "", "");
  await t.click(sharedialog.expandLink.nth(0));
  await t
    .typeText(sharedialog.linkSecretInput, "secretForTesting", { replace: true })
    .click(sharedialog.linkExpireInput)
    .click(Selector("div").withText("After 1 day").parent('div[role="listitem"]'))
    .click(sharedialog.dialogSave)
    .click(sharedialog.dialogSave);
  await contextmenu.clearSelection();
  await t.navigateTo(url);
  await t
    .expect(toolbar.toolbarTitle.withText("Christmas").visible)
    .ok()
    .click(Selector("button").withText("@photoprism_app"))
    .expect(toolbar.toolbarTitle.withText("Albums").visible)
    .ok();
  const countAlbums = await album.getAlbumCount("all");
  await t.expect(countAlbums).gte(40).useRole(Role.anonymous());
  await t.navigateTo(url);
  await t
    .expect(toolbar.toolbarTitle.withText("Christmas").visible)
    .ok()
    .click(Selector("button").withText("@photoprism_app"))
    .expect(toolbar.toolbarTitle.withText("Albums").visible)
    .ok();
  const countAlbumsAnonymous = await Selector("a.is-album").count;
  await t.expect(countAlbumsAnonymous).eql(2);
  await t.navigateTo("http://localhost:2343/browse");
  await newpage.login("admin", "photoprism");
  await menu.openPage("albums");
  await album.openAlbumWithUid(FirstAlbum);
  await toolbar.triggerToolbarAction("share", "");
  await t
    .click(sharedialog.expandLink.nth(0))
    .click(sharedialog.deleteLink)
    .useRole(Role.anonymous())
    .expect(Selector(".input-name input").visible)
    .ok();
  await t.navigateTo("http://localhost:2343/s/secretfortesting");
  await t.expect(toolbar.toolbarTitle.withText("Albums").visible).ok();
  const countAlbumsAnonymousAfterDelete = await album.getAlbumCount("all");
  await t.expect(countAlbumsAnonymousAfterDelete).eql(1);
  await t.navigateTo("http://localhost:2343/browse");
  await newpage.login("admin", "photoprism");
  await menu.openPage("folders");
  await album.openAlbumWithUid(FirstFolder);
  await toolbar.triggerToolbarAction("share", "");
  await t
    .click(sharedialog.expandLink.nth(0))
    .click(sharedialog.deleteLink)
    .useRole(Role.anonymous())
    .expect(Selector(".input-name input").visible)
    .ok();
  await t.navigateTo("http://localhost:2343/s/secretfortesting");
  await t
    .expect(toolbar.toolbarTitle.withText("Christmas").visible)
    .notOk()
    .expect(toolbar.toolbarTitle.withText("Albums").visible)
    .notOk()
    .expect(Selector(".input-name input").visible)
    .ok();
});

test.skip.meta("testID", "sharing-002")("Verify anonymous user has limited options", async (t) => {
  await t.navigateTo("http://localhost:2343/s/jxoux5ub1e/british-columbia-canada");
  await t.expect(toolbar.toolbarTitle.withText("British Columbia").visible).ok();
  await toolbar.checkToolbarActionAvailability("edit", false);
  await toolbar.checkToolbarActionAvailability("share", false);
  await toolbar.checkToolbarActionAvailability("upload", false);
  await toolbar.checkToolbarActionAvailability("reload", true);
  await toolbar.checkToolbarActionAvailability("download", true);
  await photoviews.triggerHoverAction("nth", 0, "select");
  await contextmenu.checkContextMenuActionAvailability("download", true);
  await contextmenu.checkContextMenuActionAvailability("archive", false);
  await contextmenu.checkContextMenuActionAvailability("private", false);
  await contextmenu.checkContextMenuActionAvailability("edit", false);
  await contextmenu.checkContextMenuActionAvailability("share", false);
  await contextmenu.checkContextMenuActionAvailability("album", false);
  await contextmenu.clearSelection();
  await t.expect(newpage.cardTitle.visible).notOk();
  await photoviewer.openPhotoViewer("nth", 0);
  await photoviewer.checkPhotoViewerActionAvailability("download", true);
  await photoviewer.checkPhotoViewerActionAvailability("select", true);
  await photoviewer.checkPhotoViewerActionAvailability("fullscreen", true);
  await photoviewer.checkPhotoViewerActionAvailability("slideshow", true);
  await photoviewer.checkPhotoViewerActionAvailability("like", false);
  await photoviewer.checkPhotoViewerActionAvailability("edit", false);
  await photoviewer.triggerPhotoViewerAction("close");
  await photoviews.checkHoverActionAvailability("nth", 0, "favorite", false);
  await photoviews.checkHoverActionAvailability("nth", 0, "select", true);
  await toolbar.triggerToolbarAction("view-list", "");
  await t
    .expect(Selector(`td button.input-private`).visible)
    .notOk()
    .expect(Selector(`td button.input-favorite`).visible)
    .notOk()
    .click(Selector("button").withText("@photoprism_app"))
    .expect(toolbar.toolbarTitle.withText("Albums").visible)
    .ok();
  const AlbumUid = await album.getNthAlbumUid("all", 0);
  await album.triggerHoverAction("uid", AlbumUid, "select");
  await contextmenu.checkContextMenuActionAvailability("download", true);
  await contextmenu.checkContextMenuActionAvailability("delete", false);
  await contextmenu.checkContextMenuActionAvailability("album", false);
  await contextmenu.checkContextMenuActionAvailability("edit", false);
  await contextmenu.checkContextMenuActionAvailability("share", false);
  await contextmenu.clearSelection();
});
