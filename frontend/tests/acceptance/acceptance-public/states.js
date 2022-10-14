import { Selector } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";
import Menu from "../page-model/menu";
import Album from "../page-model/album";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import Photo from "../page-model/photo";
import Page from "../page-model/page";
import AlbumDialog from "../page-model/dialog-album";

fixture`Test states`.page`${testcafeconfig.url}`;

const menu = new Menu();
const album = new Album();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const photo = new Photo();
const page = new Page();
const albumdialog = new AlbumDialog();

test.meta("testID", "states-001").meta({ mode: "public" })(
  "Common: Update state details",
  async (t) => {
    await menu.openPage("states");
    await toolbar.search("Canada");
    const AlbumUid = await album.getNthAlbumUid("all", 0);

    await t.expect(page.cardTitle.nth(0).innerText).contains("British Columbia");

    await t.click(page.cardTitle.nth(0));

    await t
      .expect(albumdialog.title.value)
      .eql("British Columbia")
      .expect(albumdialog.location.value)
      .eql("Canada")
      .expect(albumdialog.description.value)
      .eql("")
      .expect(albumdialog.category.value)
      .eql("");

    await t
      .typeText(albumdialog.title, "Wonderland", { replace: true })
      .typeText(albumdialog.location, "Earth", { replace: true })
      .typeText(albumdialog.description, "We love earth")
      .typeText(albumdialog.category, "Mountains")
      .pressKey("enter")
      .click(albumdialog.dialogSave);

    await t
      .expect(page.cardTitle.nth(0).innerText)
      .contains("Wonderland")
      .expect(page.cardDescription.nth(0).innerText)
      .contains("We love earth")
      .expect(Selector("div.caption").nth(1).innerText)
      .contains("Mountains")
      .expect(Selector("div.caption").nth(2).innerText)
      .contains("Earth");

    await album.openNthAlbum(0);

    await t.expect(toolbar.toolbarSecondTitle.innerText).contains("Wonderland");
    await t.expect(toolbar.toolbarDescription.innerText).contains("We love earth");

    await menu.openPage("states");
    if (t.browser.platform === "mobile") {
      await toolbar.search("category:Mountains");
    } else {
      await toolbar.setFilter("category", "Mountains");
    }

    await t.expect(page.cardTitle.nth(0).innerText).contains("Wonderland");

    await album.openAlbumWithUid(AlbumUid);
    await toolbar.triggerToolbarAction("edit");

    await t
      .expect(albumdialog.description.value)
      .eql("We love earth")
      .expect(albumdialog.category.value)
      .eql("Mountains")
      .expect(albumdialog.location.value)
      .eql("Earth");

    await t
      .typeText(albumdialog.title, "British Columbia", { replace: true })
      .click(albumdialog.category)
      .pressKey("ctrl+a delete")
      .pressKey("enter")
      .click(albumdialog.description)
      .pressKey("ctrl+a delete")
      .pressKey("enter")
      .typeText(albumdialog.location, "Canada", { replace: true })
      .click(albumdialog.dialogSave);
    await menu.openPage("states");
    await toolbar.search("Canada");

    await t
      .expect(page.cardTitle.nth(0).innerText)
      .contains("British Columbia")
      .expect(page.cardDescription.innerText)
      .notContains("We love earth")
      .expect(Selector("div.caption").nth(0).innerText)
      .notContains("Earth");
  }
);

test.meta("testID", "states-002").meta({ mode: "public" })(
  "Common: Create, Edit, delete sharing link for state",
  async (t) => {
    await page.testCreateEditDeleteSharingLink("states");
  }
);

test.meta("testID", "states-003").meta({ mode: "public" })(
  "Common: Create/delete album-clone from state",
  async (t) => {
    await menu.openPage("albums");
    const AlbumCount = await album.getAlbumCount("all");
    await menu.openPage("states");
    await toolbar.search("Canada");
    const FirstStateUid = await album.getNthAlbumUid("all", 0);
    await album.openAlbumWithUid(FirstStateUid);
    const PhotoCountInState = await photo.getPhotoCount("all");
    const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
    const SecondPhotoUid = await photo.getNthPhotoUid("image", 1);
    await menu.openPage("states");
    await album.selectAlbumFromUID(FirstStateUid);
    await contextmenu.triggerContextMenuAction("clone", "NotYetExistingAlbumForState");
    await menu.openPage("albums");
    const AlbumCountAfterCreation = await album.getAlbumCount("all");

    await t.expect(AlbumCountAfterCreation).eql(AlbumCount + 1);

    await toolbar.search("NotYetExistingAlbumForState");
    const AlbumUid = await album.getNthAlbumUid("all", 0);
    await album.openAlbumWithUid(AlbumUid);
    const PhotoCountInAlbum = await photo.getPhotoCount("all");

    await t.expect(PhotoCountInAlbum).eql(PhotoCountInState);

    await photo.checkPhotoVisibility(FirstPhotoUid, true);
    await photo.checkPhotoVisibility(SecondPhotoUid, true);
    await menu.openPage("albums");
    await album.selectAlbumFromUID(AlbumUid);
    await contextmenu.triggerContextMenuAction("delete", "");
    const AlbumCountAfterDelete = await album.getAlbumCount("all");

    await t.expect(AlbumCountAfterDelete).eql(AlbumCount);

    await menu.openPage("states");
    await album.openAlbumWithUid(FirstStateUid);
    await photo.checkPhotoVisibility(FirstPhotoUid, true);
    await photo.checkPhotoVisibility(SecondPhotoUid, true);
  }
);
