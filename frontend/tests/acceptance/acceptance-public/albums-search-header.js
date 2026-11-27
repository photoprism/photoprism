import testcafeconfig from "../../testcafeconfig.json";
import Album from "../page-model/album";
import AlbumDialog from "../page-model/dialog-album";
import Menu from "../page-model/menu";
import Notifies from "../page-model/notifications";
import Photo from "../page-model/photo";
import PhotoViewer from "../page-model/photoviewer";
import Toolbar from "../page-model/toolbar";
import { helperBeforeFixture, helperBeforeEach, helperAfterEach } from "../page-model/helpers";

fixture`Test album search and header`
  .page`${testcafeconfig.url}`
  .before(helperBeforeFixture)
  .beforeEach(helperBeforeEach)
  .afterEach(helperAfterEach);

const album = new Album();
const albumDialog = new AlbumDialog();
const menu = new Menu();
const notifies = new Notifies();
const photo = new Photo();
const viewer = new PhotoViewer();
const toolbar = new Toolbar();

const albumCat = "pqmxlr23mv73aj2f";
const outsideCat = "pqmxlr6ze5u9jjig";
const tower = "pqmxlr12pi8arvey";

// openCollection opens a fixture collection by title and waits for its results.
async function openCollection(page, title) {
  await menu.openPage(page);
  await toolbar.search(title);
  await album.openAlbumWithUid(await album.getAlbumUidByTitle(title));
  await notifies.waitForPhotosToLoad(8000);
}

test.meta("testID", "albums-search-001").meta({ type: "short", mode: "public" })("Common: Keep album search scoped across reload and clear", async (t) => {
  await menu.openPage("browse");
  await toolbar.search("label:cat");
  await photo.checkPhotoVisibility(outsideCat, true);
  await openCollection("albums", "Holiday");
  await photo.checkPhotoVisibility(tower, true);
  await toolbar.searchAlbum("label:cat");
  await t.expect(photo.results.count).eql(1);
  await photo.checkPhotoVisibility(albumCat, true);
  await photo.checkPhotoVisibility(outsideCat, false);
  await photo.checkPhotoVisibility(tower, false);
  await t.eval(() => window.location.reload());
  await notifies.waitForPhotosToLoad(8000);
  await t.expect(toolbar.albumSearchField.value).eql("label:cat").expect(photo.results.count).eql(1);
  await photo.checkPhotoVisibility(albumCat, true);
  await toolbar.searchAlbum("no-matching-picture-zzqqxx");
  await t.expect(album.emptyResults.visible).ok();
  await t.click(album.clearSearch);
  await notifies.waitForPhotosToLoad(8000);
  await photo.checkPhotoVisibility(tower, true);
});

test.meta("testID", "albums-search-002").meta({ type: "short", mode: "public" })("Common: Sort by title and date within an album", async (t) => {
  await openCollection("albums", "Holiday");
  await toolbar.sortAlbum("title");
  await t.expect(photo.results.nth(0).getAttribute("data-uid")).eql(albumCat);
  await toolbar.sortAlbum("title");
  await t.expect(photo.results.nth(0).getAttribute("data-uid")).eql(tower);
  await toolbar.sortAlbum("newest");
  await t.expect(photo.results.nth(0).getAttribute("data-uid")).eql(albumCat);
  await toolbar.sortAlbum("newest");
  await t.expect(photo.results.nth(0).getAttribute("data-uid")).eql(tower);
  await t.click(toolbar.albumSortAction);
  await t.expect(toolbar.albumSortOption("newest").find(".v-list-item__append .v-icon").exists).ok();
  await t.expect(toolbar.albumSortOption("added").find(".v-list-item__append .v-icon").exists).notOk();
});

test.meta("testID", "albums-search-003").meta({ type: "short", mode: "public" })("Common: Expand a description and keep the compact album navigation usable", async (t) => {
  await openCollection("albums", "Holiday");
  await toolbar.triggerToolbarAction("edit");
  await t.typeText(albumDialog.description, "A long album description for the expanded header. ".repeat(45), { replace: true, paste: true });
  await t.click(albumDialog.dialogSave);
  await t.expect(album.headerDescription.clipped).ok();
  await t.click(album.headerMore).expect(album.headerDescription.clipped).notOk();
  await t.click(album.headerMore).expect(album.headerDescription.clipped).ok();
  await t.expect(toolbar.albumCoverToolbar.visible).ok();
  await viewer.openPhotoViewer("uid", tower);
  const oldCover = await album.headerCover.getAttribute("src");
  await viewer.triggerPhotoViewerAction("cover");
  await notifies.waitForSpecficEvent("Changes successfully saved", 8000);
  await viewer.triggerPhotoViewerAction("close-button");
  await t.expect(album.headerCover.getAttribute("src")).notEql(oldCover);
  await openCollection("folders", "Images");
  await t.expect(toolbar.albumCoverToolbar.visible).ok();
  await t.scrollBy(0, 700);
  await t.expect(toolbar.albumCoverToolbar.exists).notOk();
  await t.expect(toolbar.albumBackLink.visible).ok().expect(toolbar.compactAlbumTitle.withText("Images").visible).ok();
});

test.meta("testID", "albums-search-004").meta({ type: "short", mode: "public" })("Common: Search folders and moments by a plain word", async () => {
  await openCollection("folders", "Images");
  await toolbar.searchAlbum("Tower");
  await photo.checkPhotoVisibility(tower, true);
  await photo.checkPhotoVisibility(albumCat, false);
  await openCollection("moments", "France 2017");
  await toolbar.searchAlbum("Bonifacio");
  await photo.checkPhotoVisibility(outsideCat, true);
  await photo.checkPhotoVisibility(tower, false);
});
