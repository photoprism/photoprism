import { Selector } from "testcafe";
import testcafeconfig from "./testcafeconfig";
import { RequestLogger } from "testcafe";
import Menu from "../page-model/menu";
import Album from "../page-model/album";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import Photo from "../page-model/photo";
import Page from "../page-model/page";

const logger = RequestLogger(/http:\/\/localhost:2343\/api\/v1\/*/, {
  logResponseHeaders: true,
  logResponseBody: true,
});

fixture`Test calendar`.page`${testcafeconfig.url}`.requestHooks(logger);

const menu = new Menu();
const album = new Album();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const photo = new Photo();
const page = new Page();

test.meta("testID", "albums-005")("View calendar", async (t) => {
  await menu.openPage("calendar");
  await t
    .expect(Selector("a").withText("May 2019").visible)
    .ok()
    .expect(Selector("a").withText("October 2019").visible)
    .ok();
});

test.meta("testID", "calendar-001")("Update calendar", async (t) => {
  await menu.openPage("calendar");
  await toolbar.search("March 2014");
  const AlbumUid = await album.getNthAlbumUid("all", 0);
  await t
    .expect(page.cardTitle.nth(0).innerText)
    .contains("March 2014")
    .click(page.cardTitle.nth(0))
    .typeText(Selector(".input-location input"), "Snow", { replace: true })
    .expect(Selector(".input-description textarea").value)
    .eql("")
    .expect(Selector(".input-category input").value)
    .eql("")
    .typeText(Selector(".input-description textarea"), "We went to ski")
    .typeText(Selector(".input-category input"), "Mountains")
    .pressKey("enter")
    .click(".action-confirm")
    .expect(page.cardTitle.nth(0).innerText)
    .contains("March 2014")
    .expect(page.cardDescription.nth(0).innerText)
    .contains("We went to ski")
    .expect(Selector("div.caption").nth(1).innerText)
    .contains("Mountains")
    .expect(Selector("div.caption").nth(2).innerText)
    .contains("Snow");
  await album.openNthAlbum(0);
  await t.expect(toolbar.toolbarTitle.innerText).contains("March 2014");
  await t.expect(toolbar.toolbarDescription.innerText).contains("We went to ski");
  await menu.openPage("calendar");
  if (t.browser.platform === "mobile") {
    await toolbar.search("category:Mountains");
  } else {
    await toolbar.setFilter("category", "Mountains");
  }
  await t.expect(page.cardTitle.nth(0).innerText).contains("March 2014");
  await album.openAlbumWithUid(AlbumUid);
  await toolbar.triggerToolbarAction("edit", "");
  await t
    .expect(Selector(".input-description textarea").value)
    .eql("We went to ski")
    .expect(Selector(".input-category input").value)
    .eql("Mountains")
    .expect(Selector(".input-location input").value)
    .eql("Snow")
    .click(Selector(".input-category input"))
    .pressKey("ctrl+a delete")
    .pressKey("enter")
    .click(Selector(".input-description textarea"))
    .pressKey("ctrl+a delete")
    .pressKey("enter")
    .click(Selector(".input-location input"))
    .pressKey("ctrl+a delete")
    .pressKey("enter")
    .click(".action-confirm");
  await menu.openPage("calendar");
  await toolbar.search("March 2014");
  await t
    .expect(page.cardDescription.innerText)
    .notContains("We went to ski")
    .expect(Selector("div.caption").nth(0).innerText)
    .notContains("Snow");
});

//TODO test that sharing link works as expected
test.meta("testID", "calendar-003")("Create, Edit, delete sharing link", async (t) => {
  await page.testCreateEditDeleteSharingLink("calendar");
});

test.meta("testID", "calendar-004")("Create/delete album-clone from calendar", async (t) => {
  await menu.openPage("albums");
  const countAlbums = await album.getAlbumCount("all");
  await menu.openPage("calendar");
  const SecondCalendar = await album.getNthAlbumUid("all", 1);
  await album.openAlbumWithUid(SecondCalendar);
  const PhotoCountInCalendar = await photo.getPhotoCount("all");
  const FirstPhoto = await photo.getNthPhotoUid("image", 0);
  const SecondPhoto = await photo.getNthPhotoUid("image", 1);
  await menu.openPage("calendar");
  await album.selectAlbumFromUID(SecondCalendar);
  await contextmenu.triggerContextMenuAction("clone", "NotYetExistingAlbumForCalendar", "");
  await menu.openPage("albums");
  const countAlbumsAfterCreation = await album.getAlbumCount("all");
  await t.expect(countAlbumsAfterCreation).eql(countAlbums + 1);
  await toolbar.search("NotYetExistingAlbumForCalendar");
  const AlbumUid = await album.getNthAlbumUid("all", 0);
  await album.openAlbumWithUid(AlbumUid);
  const PhotoCountInAlbum = await photo.getPhotoCount("all");
  await t.expect(PhotoCountInAlbum).eql(PhotoCountInCalendar);
  await photo.checkPhotoVisibility(FirstPhoto, true);
  await photo.checkPhotoVisibility(SecondPhoto, true);
  await menu.openPage("albums");
  await album.selectAlbumFromUID(AlbumUid);
  await contextmenu.triggerContextMenuAction("delete", "", "");
  await menu.openPage("albums");
  const countAlbumsAfterDelete = await album.getAlbumCount("all");
  await t.expect(countAlbumsAfterDelete).eql(countAlbums);
  await menu.openPage("calendar");
  await album.openAlbumWithUid(SecondCalendar);
  await photo.checkPhotoVisibility(FirstPhoto, true);
  await photo.checkPhotoVisibility(SecondPhoto, true);
});
