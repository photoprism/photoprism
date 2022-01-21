import { Selector } from "testcafe";
import testcafeconfig from "./testcafeconfig";
import Menu from "../page-model/menu";
import Album from "../page-model/album";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import Photo from "../page-model/photo";
import NewPage from "../page-model/page";

fixture`Test moments`.page`${testcafeconfig.url}`;

const menu = new Menu();
const album = new Album();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const photo = new Photo();
const newpage = new NewPage();

test.meta("testID", "moments-001")("Update moment", async (t) => {
  await menu.openPage("moments");
  await toolbar.search("Nature");
  const AlbumUid = await album.getNthAlbumUid("all", 0);
  await t
    .expect(newpage.cardTitle.nth(0).innerText)
    .contains("Nature")
    .click(newpage.cardTitle.nth(0))
    .expect(Selector(".input-title input").value)
    .eql("Nature & Landscape")
    .expect(Selector(".input-location input").value)
    .eql("")
    .typeText(Selector(".input-title input"), "Winter", { replace: true })
    .typeText(Selector(".input-location input"), "Snow-Land", { replace: true })
    .expect(Selector(".input-description textarea").value)
    .eql("")
    .expect(Selector(".input-category input").value)
    .eql("")
    .typeText(Selector(".input-description textarea"), "We went to ski")
    .typeText(Selector(".input-category input"), "Mountains")
    .pressKey("enter")
    .click(".action-confirm")
    .expect(newpage.cardTitle.nth(0).innerText)
    .contains("Winter")
    .expect(newpage.cardDescription.nth(0).innerText)
    .contains("We went to ski")
    .expect(Selector("div.caption").nth(1).innerText)
    .contains("Mountains")
    .expect(Selector("div.caption").nth(2).innerText)
    .contains("Snow-Land");
  await album.openNthAlbum(0);
  await t.expect(toolbar.toolbarTitle.innerText).contains("Winter");
  await t.expect(toolbar.toolbarDescription.innerText).contains("We went to ski");
  await menu.openPage("moments");
  if (t.browser.platform === "mobile") {
    await toolbar.search("category:Mountains");
  } else {
    await toolbar.setFilter("category", "Mountains");
  }
  await t.expect(newpage.cardTitle.nth(0).innerText).contains("Winter");
  await album.openAlbumWithUid(AlbumUid);
  await toolbar.triggerToolbarAction("edit", "");
  await t
    .expect(Selector(".input-description textarea").value)
    .eql("We went to ski")
    .expect(Selector(".input-category input").value)
    .eql("Mountains")
    .expect(Selector(".input-location input").value)
    .eql("Snow-Land")
    .typeText(Selector(".input-title input"), "Nature & Landscape", { replace: true })
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
  await menu.openPage("moments");
  await toolbar.search("Nature");
  await t
    .expect(newpage.cardTitle.nth(0).innerText)
    .contains("Nature & Landscape")
    .expect(newpage.cardDescription.innerText)
    .notContains("We went to ski")
    .expect(Selector("div.caption").nth(0).innerText)
    .notContains("Snow-Land");
});

test.meta("testID", "moments-003")("Create, Edit, delete sharing link", async (t) => {
  await newpage.testCreateEditDeleteSharingLink("moments");
});

test.meta("testID", "moments-004")("Create/delete album-clone from moment", async (t) => {
  await menu.openPage("albums");
  const countAlbums = await album.getAlbumCount("all");
  await menu.openPage("moments");
  const FirstMoment = await album.getNthAlbumUid("all", 0);
  await album.openAlbumWithUid(FirstMoment);
  const PhotoCountInMoment = await photo.getPhotoCount("all");
  const FirstPhoto = await photo.getNthPhotoUid("image", 0);
  const SecondPhoto = await photo.getNthPhotoUid("image", 1);
  await menu.openPage("moments");
  await album.selectAlbumFromUID(FirstMoment);
  await contextmenu.triggerContextMenuAction("clone", "NotYetExistingAlbumForMoment", "");
  await menu.openPage("albums");
  const countAlbumsAfterCreation = await album.getAlbumCount("all");
  await t.expect(countAlbumsAfterCreation).eql(countAlbums + 1);
  await toolbar.search("NotYetExistingAlbumForMoment");
  const AlbumUid = await album.getNthAlbumUid("all", 0);
  await album.openAlbumWithUid(AlbumUid);
  const PhotoCountInAlbum = await photo.getPhotoCount("all");
  await t.expect(PhotoCountInAlbum).eql(PhotoCountInMoment);
  await photo.checkPhotoVisibility(FirstPhoto, true);
  await photo.checkPhotoVisibility(SecondPhoto, true);
  await menu.openPage("albums");
  await album.selectAlbumFromUID(AlbumUid);
  await contextmenu.triggerContextMenuAction("delete", "", "");
  await menu.openPage("albums");
  const countAlbumsAfterDelete = await album.getAlbumCount("all");
  await t.expect(countAlbumsAfterDelete).eql(countAlbums);
  await menu.openPage("moments");
  await album.openAlbumWithUid(FirstMoment);
  await photo.checkPhotoVisibility(FirstPhoto, true);
  await photo.checkPhotoVisibility(SecondPhoto, true);
});
