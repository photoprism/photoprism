import { Selector } from "testcafe";
import testcafeconfig from "./testcafeconfig";
import Page from "./page-model";
import { ClientFunction } from "testcafe";
import fs from "fs";
import Menu from "../page-model/menu";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import Photo from "../page-model/photo";
import NewPage from "../page-model/page";
import PhotoViews from "../page-model/photo-views";
import PhotoEdit from "../page-model/photo-edit";
import Originals from "../page-model/originals";
import Album from "../page-model/album";

fixture`Test photos upload and delete`.page`${testcafeconfig.url}`;

const menu = new Menu();
const album = new Album();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const photo = new Photo();
const newpage = new NewPage();
const photoviews = new PhotoViews();
const photoedit = new PhotoEdit();
const originals = new Originals();

test.meta("testID", "photos-upload-delete-001")("Upload + Delete jpg/json", async (t) => {
  await t.expect(fs.existsSync("../storage/acceptance/originals/2020/10")).notOk();
  await toolbar.search("digikam");
  const PhotoCount = await photo.getPhotoCount("all");
  await t.expect(PhotoCount).eql(0);
  await toolbar.triggerToolbarAction("upload", "");
  await t
    .setFilesToUpload(Selector(".input-upload"), [
      "./upload-files/digikam.jpg",
      "./upload-files/digikam.json",
    ])
    .wait(15000);
  const PhotoCountAfterUpload = await photo.getPhotoCount("all");
  await t.expect(PhotoCountAfterUpload).eql(1);
  const UploadedPhoto = await photo.getNthPhotoUid("all", 0);
  await t.navigateTo("/library/files/2020/10");
  const FileCount = await originals.getFileCount();
  await t.expect(FileCount).eql(2);
  await menu.openPage("browse");
  await toolbar.search("digikam");
  await photoviews.triggerHoverAction("uid", UploadedPhoto, "select");
  await contextmenu.triggerContextMenuAction("edit", "", "");
  await t
    .click("#tab-files")
    .expect(Selector("div.caption").withText(".json").visible)
    .ok()
    .expect(Selector("div.caption").withText(".jpg").visible)
    .ok()
    .click(photoedit.dialogClose);
  if (t.browser.platform !== "mobile") {
    await t.expect(fs.existsSync("../storage/acceptance/originals/2020/10")).ok();
    const originalsLength = fs.readdirSync("../storage/acceptance/originals/2020/10").length;
    await t.expect(originalsLength).eql(2);
  }
  await contextmenu.triggerContextMenuAction("archive", "", "");
  await menu.openPage("archive");
  await photoviews.triggerHoverAction("uid", UploadedPhoto, "select");
  await contextmenu.triggerContextMenuAction("delete", "", "");
  await menu.openPage("browse");
  await toolbar.search("digikam");
  await photo.checkPhotoVisibility(UploadedPhoto, false);
  await t.navigateTo("/library/files/2020/10");
  const FileCountAfterDelete = await originals.getFileCount();
  await t.expect(FileCountAfterDelete).eql(0);
  if (t.browser.platform !== "mobile") {
    const originalsLengthAfterDelete = fs.readdirSync(
      "../storage/acceptance/originals/2020/10"
    ).length;
    await t.expect(originalsLengthAfterDelete).eql(0);
  }
});

test.meta("testID", "photos-upload-delete-002")("Upload + Delete video", async (t) => {
  await t.expect(fs.existsSync("../storage/acceptance/originals/2020/06")).notOk();
  await toolbar.search("korn");
  const PhotoCount = await photo.getPhotoCount("all");
  await t.expect(PhotoCount).eql(0);
  await toolbar.triggerToolbarAction("upload", "");
  await t.setFilesToUpload(Selector(".input-upload"), ["./upload-files/korn.mp4"]).wait(15000);
  const PhotoCountAfterUpload = await photo.getPhotoCount("all");
  await t.expect(PhotoCountAfterUpload).eql(1);
  const UploadedPhoto = await photo.getNthPhotoUid("all", 0);
  await t.navigateTo("/library/files/2020/06");
  const FileCount = await originals.getFileCount();
  await t.expect(FileCount).eql(1);
  await menu.openPage("browse");
  await toolbar.search("korn");
  await photoviews.triggerHoverAction("uid", UploadedPhoto, "select");
  await contextmenu.triggerContextMenuAction("edit", "", "");
  await t
    .click("#tab-files")
    .expect(Selector("div.caption").withText(".mp4").visible)
    .ok()
    .expect(Selector("div.caption").withText(".jpg").visible)
    .ok()
    .click(photoedit.dialogClose);
  if (t.browser.platform !== "mobile") {
    await t.expect(fs.existsSync("../storage/acceptance/originals/2020/06")).ok();
    const originalsLength = fs.readdirSync("../storage/acceptance/originals/2020/06").length;
    await t.expect(originalsLength).eql(1);
    const sidecarLength = fs.readdirSync("../storage/acceptance/originals/2020/06").length;
    await t.expect(sidecarLength).eql(1);
  }
  await contextmenu.triggerContextMenuAction("archive", "", "");
  await menu.openPage("archive");
  await photoviews.triggerHoverAction("uid", UploadedPhoto, "select");
  await contextmenu.triggerContextMenuAction("delete", "", "");
  await menu.openPage("browse");
  await toolbar.search("korn");
  await photo.checkPhotoVisibility(UploadedPhoto, false);
  await t.navigateTo("/library/files/2020/06");
  const FileCountAfterDelete = await originals.getFileCount();
  await t.expect(FileCountAfterDelete).eql(0);
  if (t.browser.platform !== "mobile") {
    const originalsLengthAfterDelete = fs.readdirSync(
      "../storage/acceptance/originals/2020/06"
    ).length;
    await t.expect(originalsLengthAfterDelete).eql(0);
    const sidecarLengthAfterDelete = fs.readdirSync(
      "../storage/acceptance/originals/2020/06"
    ).length;
    await t.expect(sidecarLengthAfterDelete).eql(0);
  }
});

test.meta("testID", "photos-upload-delete-003")("Upload to existing Album + Delete", async (t) => {
  await menu.openPage("albums");
  await toolbar.search("Christmas");
  const AlbumUid = await album.getNthAlbumUid("all", 0);
  await album.openAlbumWithUid(AlbumUid);
  const PhotoCount = await photo.getPhotoCount("all");
  await toolbar.triggerToolbarAction("upload", "");
  await t
    .click(Selector(".input-albums"))
    .click(newpage.selectOption.withText("Christmas"))
    .setFilesToUpload(Selector(".input-upload"), ["./upload-files/ladybug.jpg"])
    .wait(15000);
  const PhotoCountAfterUpload = await photo.getPhotoCount("all");
  await t.expect(PhotoCountAfterUpload).eql(PhotoCount + 1);
  await menu.openPage("browse");
  await toolbar.search("ladybug");
  const UploadedPhoto = await photo.getNthPhotoUid("all", 0);
  await photoviews.triggerHoverAction("uid", UploadedPhoto, "select");
  await contextmenu.triggerContextMenuAction("archive", "", "");
  await menu.openPage("archive");
  await photoviews.triggerHoverAction("uid", UploadedPhoto, "select");
  await contextmenu.triggerContextMenuAction("delete", "", "");
  await menu.openPage("browse");
  await toolbar.search("ladybug");
  await photo.checkPhotoVisibility(UploadedPhoto, false);
  await menu.openPage("albums");
  await album.openAlbumWithUid(AlbumUid);
  await photo.checkPhotoVisibility(UploadedPhoto, false);
  const PhotoCountAfterDelete = await photo.getPhotoCount("all");
  await t.expect(PhotoCountAfterDelete).eql(PhotoCount);
});

test.meta("testID", "photos-upload-delete-004")("Upload jpg to new Album + Delete", async (t) => {
  await menu.openPage("albums");
  const AlbumCount = await album.getAlbumCount("all");
  await toolbar.triggerToolbarAction("upload", "");
  await t
    .click(Selector(".input-albums"))
    .typeText(Selector(".input-albums input"), "NewCreatedAlbum")
    .pressKey("enter")
    .setFilesToUpload(Selector(".input-upload"), ["./upload-files/digikam.jpg"])
    .wait(15000);
  if (t.browser.platform === "mobile") {
    await t.eval(() => location.reload());
  } else {
    await toolbar.triggerToolbarAction("reload", "");
  }
  const AlbumCountAfterUpload = await album.getAlbumCount("all");
  await t.expect(AlbumCountAfterUpload).eql(AlbumCount + 1);
  await toolbar.search("NewCreatedAlbum");
  await album.openNthAlbum(0);
  const PhotoCount = await photo.getPhotoCount("all");
  await t.expect(PhotoCount).eql(1);
  await menu.openPage("browse");
  await toolbar.search("digikam");
  const UploadedPhoto = await photo.getNthPhotoUid("all", 0);
  await photoviews.triggerHoverAction("uid", UploadedPhoto, "select");
  await contextmenu.triggerContextMenuAction("archive", "", "");
  await menu.openPage("archive");
  await photoviews.triggerHoverAction("uid", UploadedPhoto, "select");
  await contextmenu.triggerContextMenuAction("delete", "", "");
  await menu.openPage("browse");
  await toolbar.search("digikam");
  await photo.checkPhotoVisibility(UploadedPhoto, false);
  await menu.openPage("albums");
  await toolbar.search("NewCreatedAlbum");
  await album.openNthAlbum(0);
  await photo.checkPhotoVisibility(UploadedPhoto, false);
  const PhotoCountAfterDelete = await photo.getPhotoCount("all");
  await t.expect(PhotoCountAfterDelete).eql(0);
  await menu.openPage("albums");
  await toolbar.search("NewCreatedAlbum");
  await album.triggerHoverAction("nth", 0, "select");
  await contextmenu.checkContextMenuCount("1");
  await contextmenu.triggerContextMenuAction("delete", "", "");
});

test.meta("testID", "photos-upload-delete-005")("Try uploading nsfw file", async (t) => {
  await toolbar.triggerToolbarAction("upload", "");
  await t.setFilesToUpload(Selector(".input-upload"), ["./upload-files/hentai_2.jpg"]).wait(15000);
  await menu.openPage("library");
  await t
    .click(Selector("#tab-library-logs"))
    .expect(Selector("p").withText("hentai_2.jpg might be offensive").visible)
    .ok();
});

test.meta("testID", "photos-upload-delete-006")("Try uploading txt file", async (t) => {
  await toolbar.triggerToolbarAction("upload", "");
  await t.setFilesToUpload(Selector(".input-upload"), ["./upload-files/foo.txt"]).wait(15000);
  await menu.openPage("library");
  await t
    .click(Selector("#tab-library-logs"))
    .expect(Selector("p").withText(" foo.txt is not a jpeg file").visible)
    .ok();
});
