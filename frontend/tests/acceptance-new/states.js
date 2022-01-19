import { Selector } from "testcafe";
import testcafeconfig from "./testcafeconfig";
import Page from "./page-model";
import Menu from "../page-model/menu";
import Album from "../page-model/album";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import Photo from "../page-model/photo";
import PhotoViewer from "../page-model/photoviewer";
import NewPage from "../page-model/page";

fixture`Test states`.page`${testcafeconfig.url}`;

const page = new Page();
const menu = new Menu();
const album = new Album();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const photo = new Photo();
const photoviewer = new PhotoViewer();
const newpage = new NewPage();

test.meta("testID", "states-001")("Update state", async (t) => {
  await menu.openPage("states");
  await toolbar.search("Canada");
  const AlbumUid = await album.getNthAlbumUid("all", 0);
  await t
    .expect(newpage.cardTitle.nth(0).innerText)
    .contains("British Columbia")
    .click(newpage.cardTitle.nth(0))
    .expect(Selector(".input-title input").value)
    .eql("British Columbia")
    .expect(Selector(".input-location input").value)
    .eql("Canada")
    .typeText(Selector(".input-title input"), "Wonderland", { replace: true })
    .typeText(Selector(".input-location input"), "Earth", { replace: true })
    .expect(Selector(".input-description textarea").value)
    .eql("")
    .expect(Selector(".input-category input").value)
    .eql("")
    .typeText(Selector(".input-description textarea"), "We love earth")
    .typeText(Selector(".input-category input"), "Mountains")
    .pressKey("enter")
    .click(".action-confirm")
    .expect(newpage.cardTitle.nth(0).innerText)
    .contains("Wonderland")
    .expect(newpage.cardDescription.nth(0).innerText)
    .contains("We love earth")
    .expect(Selector("div.caption").nth(1).innerText)
    .contains("Mountains")
    .expect(Selector("div.caption").nth(2).innerText)
    .contains("Earth");
  await album.openNthAlbum(0);
  await t.expect(toolbar.toolbarTitle.innerText).contains("Wonderland");
  await t.expect(toolbar.toolbarDescription.innerText).contains("We love earth");
  await menu.openPage("states");
  if (t.browser.platform === "mobile") {
    await page.search("category:Mountains");
  } else {
    await toolbar.setFilter("category", "Mountains");
  }
  await t.expect(newpage.cardTitle.nth(0).innerText).contains("Wonderland");
  await album.openAlbumWithUid(AlbumUid);
  await toolbar.triggerToolbarAction("edit", "");
  await t
    .expect(Selector(".input-description textarea").value)
    .eql("We love earth")
    .expect(Selector(".input-category input").value)
    .eql("Mountains")
    .expect(Selector(".input-location input").value)
    .eql("Earth")
    .typeText(Selector(".input-title input"), "British Columbia / Canada", { replace: true })
    .click(Selector(".input-category input"))
    .pressKey("ctrl+a delete")
    .pressKey("enter")
    .click(Selector(".input-description textarea"))
    .pressKey("ctrl+a delete")
    .pressKey("enter")
    .typeText(Selector(".input-location input"), "Canada", { replace: true })
    .click(".action-confirm");
  await menu.openPage("states");
  await toolbar.search("Canada");
  await t
    .expect(newpage.cardTitle.nth(0).innerText)
    .contains("British Columbia / Canada")
    .expect(newpage.cardDescription.innerText)
    .notContains("We love earth")
    .expect(Selector("div.caption").nth(0).innerText)
    .notContains("Earth");
});

//TODO test that sharing link works as expected
test.meta("testID", "states-003")("Create, Edit, delete sharing link", async (t) => {
  await page.testCreateEditDeleteSharingLink("states");
});

test.meta("testID", "states-004")("Create/delete album-clone from state", async (t) => {
  await menu.openPage("albums");
  const countAlbums = await album.getAlbumCount("all");
  await menu.openPage("states");
  await toolbar.search("Canada");
  const FirstState = await album.getNthAlbumUid("all", 0);
  await album.openAlbumWithUid(FirstState);
  const PhotoCountInState = await photo.getPhotoCount("all");
  const FirstPhoto = await photo.getNthPhotoUid("image", 0);
  const SecondPhoto = await photo.getNthPhotoUid("image", 1);
  await menu.openPage("states");
  await album.selectAlbumFromUID(FirstState);
  await contextmenu.triggerContextMenuAction("clone", "NotYetExistingAlbumForState", "");
  await menu.openPage("albums");
  const countAlbumsAfterCreation = await album.getAlbumCount("all");
  await t.expect(countAlbumsAfterCreation).eql(countAlbums + 1);
  await toolbar.search("NotYetExistingAlbumForState");
  const AlbumUid = await album.getNthAlbumUid("all", 0);
  await album.openAlbumWithUid(AlbumUid);
  const PhotoCountInAlbum = await photo.getPhotoCount("all");
  await t.expect(PhotoCountInAlbum).eql(PhotoCountInState);
  await photo.checkPhotoVisibility(FirstPhoto, true);
  await photo.checkPhotoVisibility(SecondPhoto, true);
  await menu.openPage("albums");
  await album.selectAlbumFromUID(AlbumUid);
  await contextmenu.triggerContextMenuAction("delete", "", "");
  await menu.openPage("albums");
  const countAlbumsAfterDelete = await album.getAlbumCount("all");
  await t.expect(countAlbumsAfterDelete).eql(countAlbums);
  await menu.openPage("states");
  await album.openAlbumWithUid(FirstState);
  await photo.checkPhotoVisibility(FirstPhoto, true);
  await photo.checkPhotoVisibility(SecondPhoto, true);
});
