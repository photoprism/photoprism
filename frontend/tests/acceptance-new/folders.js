import { Selector } from "testcafe";
import testcafeconfig from "./testcafeconfig";
import Menu from "../page-model/menu";
import Album from "../page-model/album";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import Photo from "../page-model/photo";
import NewPage from "../page-model/page";

fixture`Test folders`.page`${testcafeconfig.url}`;

const menu = new Menu();
const album = new Album();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const photo = new Photo();
const newpage = new NewPage();

test.meta("testID", "albums-004")("View folders", async (t) => {
  await menu.openPage("folders");
  await t
    .expect(Selector("a").withText("BotanicalGarden").visible)
    .ok()
    .expect(Selector("a").withText("Kanada").visible)
    .ok()
    .expect(Selector("a").withText("KorsikaAdventure").visible)
    .ok();
});

test.meta("testID", "folders-001")("Update folders", async (t) => {
  await menu.openPage("folders");
  await toolbar.search("Kanada");
  const AlbumUid = await album.getNthAlbumUid("all", 0);
  await t
    .expect(newpage.cardTitle.nth(0).innerText)
    .contains("Kanada")
    .click(newpage.cardTitle.nth(0))
    .expect(Selector(".input-title input").value)
    .eql("Kanada")
    .expect(Selector(".input-location input").value)
    .eql("")
    .typeText(Selector(".input-title input"), "MyFolder", { replace: true })
    .typeText(Selector(".input-location input"), "USA", { replace: true })
    .expect(Selector(".input-description textarea").value)
    .eql("")
    .expect(Selector(".input-category input").value)
    .eql("")
    .typeText(Selector(".input-description textarea"), "Last holiday")
    .typeText(Selector(".input-category input"), "Mountains")
    .pressKey("enter")
    .click(".action-confirm")
    .expect(newpage.cardTitle.nth(0).innerText)
    .contains("MyFolder")
    .expect(newpage.cardDescription.nth(0).innerText)
    .contains("Last holiday")
    .expect(Selector("div.caption").nth(1).innerText)
    .contains("Mountains")
    .expect(Selector("div.caption").nth(2).innerText)
    .contains("USA");
  await album.openNthAlbum(0);
  await t
    .expect(toolbar.toolbarDescription.nth(0).innerText)
    .contains("Last holiday")
    .expect(toolbar.toolbarTitle.nth(0).innerText)
    .contains("MyFolder");
  await menu.openPage("folders");
  if (t.browser.platform === "mobile") {
    await toolbar.search("category:Mountains");
  } else {
    await toolbar.setFilter("category", "Mountains");
  }
  await t.expect(newpage.cardTitle.nth(0).innerText).contains("MyFolder");
  await album.openAlbumWithUid(AlbumUid);
  await toolbar.triggerToolbarAction("edit", "");
  await t
    .expect(Selector(".input-description textarea").value)
    .eql("Last holiday")
    .expect(Selector(".input-category input").value)
    .eql("Mountains")
    .expect(Selector(".input-location input").value)
    .eql("USA")
    .typeText(Selector(".input-title input"), "Kanada", { replace: true })
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
  await menu.openPage("folders");
  await toolbar.search("Kanada");
  await t
    .expect(newpage.cardTitle.nth(0).innerText)
    .contains("Kanada")
    .expect(newpage.cardDescription.nth(0).innerText)
    .notContains("We went to ski")
    .expect(Selector("div.caption").nth(0).innerText)
    .notContains("USA");
});

//TODO test that sharing link works as expected
test.meta("testID", "folders-003")("Create, Edit, delete sharing link", async (t) => {
  await newpage.testCreateEditDeleteSharingLink("folders");
});

test.meta("testID", "folders-004")("Create/delete album-clone from folder", async (t) => {
  await menu.openPage("albums");
  const countAlbums = await album.getAlbumCount("all");
  await menu.openPage("folders");
  const ThirdFolder = await album.getNthAlbumUid("all", 2);
  await album.openAlbumWithUid(ThirdFolder);
  const PhotoCountInFolder = await photo.getPhotoCount("all");
  const FirstPhoto = await photo.getNthPhotoUid("image", 0);
  await menu.openPage("folders");
  await album.selectAlbumFromUID(ThirdFolder);
  await contextmenu.triggerContextMenuAction("clone", "NotYetExistingAlbumForFolder", "");
  await menu.openPage("albums");
  const countAlbumsAfterCreation = await album.getAlbumCount("all");
  await t.expect(countAlbumsAfterCreation).eql(countAlbums + 1);
  await toolbar.search("NotYetExistingAlbumForFolder");
  const AlbumUid = await album.getNthAlbumUid("all", 0);
  await album.openAlbumWithUid(AlbumUid);
  const PhotoCountInAlbum = await photo.getPhotoCount("all");
  await t.expect(PhotoCountInAlbum).eql(PhotoCountInFolder);
  await photo.checkPhotoVisibility(FirstPhoto, true);
  await menu.openPage("albums");
  await album.selectAlbumFromUID(AlbumUid);
  await contextmenu.triggerContextMenuAction("delete", "", "");
  await menu.openPage("albums");
  const countAlbumsAfterDelete = await album.getAlbumCount("all");
  await t.expect(countAlbumsAfterDelete).eql(countAlbums);
  await menu.openPage("folders");
  await album.openAlbumWithUid(ThirdFolder);
  await photo.checkPhotoVisibility(FirstPhoto, true);
});
