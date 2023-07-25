import { Selector } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";
import Menu from "../page-model/menu";
import Album from "../page-model/album";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import Photo from "../page-model/photo";
import Page from "../page-model/page";
import AlbumDialog from "../page-model/dialog-album";

fixture`Test moments`.page`${testcafeconfig.url}`;

const menu = new Menu();
const album = new Album();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const photo = new Photo();
const page = new Page();
const albumdialog = new AlbumDialog();

test.meta("testID", "moments-001").meta({ mode: "public" })(
  "Common: Update moment details",
  async (t) => {
    await menu.openPage("moments");
    await toolbar.search("Nature");
    const AlbumUid = await album.getNthAlbumUid("all", 0);

    await t.expect(page.cardTitle.nth(0).innerText).contains("Nature");

    await t.click(page.cardTitle.nth(0));

    await t
      .expect(albumdialog.title.value)
      .eql("Nature & Landscape")
      .expect(albumdialog.location.value)
      .eql("")
      .expect(albumdialog.description.value)
      .eql("")
      .expect(albumdialog.category.value)
      .eql("");

    await t
      .typeText(albumdialog.title, "Winter", { replace: true })
      .typeText(albumdialog.location, "Snow-Land", { replace: true })
      .typeText(albumdialog.description, "We went to ski")
      .typeText(albumdialog.category, "Mountains")
      .pressKey("enter")
      .click(albumdialog.dialogSave);

    await t
      .expect(page.cardTitle.nth(0).innerText)
      .contains("Winter")
      .expect(page.cardDescription.nth(0).innerText)
      .contains("We went to ski")
      .expect(Selector("div.caption").nth(1).innerText)
      .contains("Mountains")
      .expect(Selector("div.caption").nth(2).innerText)
      .contains("Snow-Land");

    await album.openNthAlbum(0);

    await t.expect(toolbar.toolbarSecondTitle.innerText).contains("Winter");
    await t.expect(toolbar.toolbarDescription.innerText).contains("We went to ski");

    await menu.openPage("moments");
    if (t.browser.platform === "mobile") {
      await toolbar.search("category:Mountains");
    } else {
      await toolbar.setFilter("category", "Mountains");
    }

    await t.expect(page.cardTitle.nth(0).innerText).contains("Winter");

    await album.openAlbumWithUid(AlbumUid);
    await toolbar.triggerToolbarAction("edit");

    await t
      .expect(albumdialog.description.value)
      .eql("We went to ski")
      .expect(albumdialog.category.value)
      .eql("Mountains")
      .expect(albumdialog.location.value)
      .eql("Snow-Land");

    await t
      .typeText(albumdialog.title, "Nature & Landscape", { replace: true })
      .click(albumdialog.category)
      .pressKey("ctrl+a delete")
      .pressKey("enter")
      .click(albumdialog.description)
      .pressKey("ctrl+a delete")
      .pressKey("enter")
      .click(albumdialog.location)
      .pressKey("ctrl+a delete")
      .pressKey("enter")
      .click(albumdialog.dialogSave);
    await menu.openPage("moments");
    await toolbar.search("Nature");

    await t
      .expect(page.cardTitle.nth(0).innerText)
      .contains("Nature & Landscape")
      .expect(page.cardDescription.innerText)
      .notContains("We went to ski")
      .expect(Selector("div.caption").nth(0).innerText)
      .notContains("Snow-Land");
  }
);

test.meta("testID", "moments-002").meta({ mode: "public" })(
  "Common: Create, Edit, delete sharing link for moment",
  async (t) => {
    await page.testCreateEditDeleteSharingLink("moments");
  }
);

test.meta("testID", "moments-003").meta({ mode: "public" })(
  "Common: Create/delete album-clone from moment",
  async (t) => {
    await menu.openPage("albums");
    const AlbumCount = await album.getAlbumCount("all");
    await menu.openPage("moments");
    const FirstMomentUid = await album.getNthAlbumUid("all", 0);
    await album.openAlbumWithUid(FirstMomentUid);
    const PhotoCountInMoment = await photo.getPhotoCount("all");
    const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
    const SecondPhotoUid = await photo.getNthPhotoUid("image", 1);
    await menu.openPage("moments");
    await album.selectAlbumFromUID(FirstMomentUid);
    await contextmenu.triggerContextMenuAction("clone", "NotYetExistingAlbumForMoment");
    await menu.openPage("albums");
    const AlbumCountAfterCreation = await album.getAlbumCount("all");

    await t.expect(AlbumCountAfterCreation).eql(AlbumCount + 1);

    await toolbar.search("NotYetExistingAlbumForMoment");
    const AlbumUid = await album.getNthAlbumUid("all", 0);
    await album.openAlbumWithUid(AlbumUid);
    const PhotoCountInAlbum = await photo.getPhotoCount("all");

    await t.expect(PhotoCountInAlbum).eql(PhotoCountInMoment);

    await photo.checkPhotoVisibility(FirstPhotoUid, true);
    await photo.checkPhotoVisibility(SecondPhotoUid, true);
    await menu.openPage("albums");
    await album.selectAlbumFromUID(AlbumUid);
    await contextmenu.triggerContextMenuAction("delete", "");
    const AlbumCountAfterDelete = await album.getAlbumCount("all");

    await t.expect(AlbumCountAfterDelete).eql(AlbumCount);

    await menu.openPage("moments");
    await album.openAlbumWithUid(FirstMomentUid);
    await photo.checkPhotoVisibility(FirstPhotoUid, true);
    await photo.checkPhotoVisibility(SecondPhotoUid, true);
  }
);

test.meta("testID", "moments-004").meta({ type: "short", mode: "public" })(
  "Common: Verify moment sort options",
  async (t) => {
    await menu.openPage("moments");
    await album.checkSortOptions("moment");
  }
);
