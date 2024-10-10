import { Selector } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";
import Menu from "../page-model/menu";
import Album from "../page-model/album";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import Photo from "../page-model/photo";
import Page from "../page-model/page";
import AlbumDialog from "../page-model/dialog-album";

fixture`Test folders`.page`${testcafeconfig.url}`;

const menu = new Menu();
const album = new Album();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const photo = new Photo();
const page = new Page();
const albumdialog = new AlbumDialog();

test.meta("testID", "folders-001").meta({ type: "short", mode: "public" })(
  "Common: View folders",
  async (t) => {
    await menu.openPage("folders");

    await t
      .expect(Selector("a").withText("BotanicalGarden").visible)
      .ok()
      .expect(Selector("a").withText("Kanada").visible)
      .ok()
      .expect(Selector("a").withText("KorsikaAdventure").visible)
      .ok();
  }
);

test.meta("testID", "folders-002").meta({ mode: "public" })(
  "Common: Update folder details",
  async (t) => {
    await menu.openPage("folders");
    await toolbar.search("Kanada");
    const AlbumUid = await album.getNthAlbumUid("all", 0);
    await t.expect(page.cardTitle.nth(0).innerText).contains("Kanada");

    await t.click(page.cardTitle.nth(0));

    await t
      .expect(albumdialog.title.value)
      .eql("Kanada")
      .expect(albumdialog.location.value)
      .eql("")
      .expect(albumdialog.description.value)
      .eql("")
      .expect(albumdialog.category.value)
      .eql("");

    await t
      .typeText(albumdialog.title, "MyFolder", { replace: true })
      .typeText(albumdialog.location, "United States", { replace: true })
      .typeText(albumdialog.description, "Last holiday")
      .typeText(albumdialog.category, "Mountains")
      .pressKey("enter")
      .click(albumdialog.dialogSave);

    await t
      .expect(page.cardTitle.nth(0).innerText)
      .contains("MyFolder")
      .expect(page.cardDescription.nth(0).innerText)
      .contains("Last holiday")
      .expect(Selector("div.caption").nth(1).innerText)
      .contains("Mountains")
      .expect(Selector("div.caption").nth(2).innerText)
      .contains("United States");

    await album.openNthAlbum(0);

    await t
      .expect(toolbar.toolbarDescription.nth(0).innerText)
      .contains("Last holiday")
      .expect(toolbar.toolbarSecondTitle.innerText)
      .contains("MyFolder");

    await menu.openPage("folders");
    if (t.browser.platform === "mobile") {
      await toolbar.search("category:Mountains");
    } else {
      await toolbar.setFilter("category", "Mountains");
    }

    await t.expect(page.cardTitle.nth(0).innerText).contains("MyFolder");

    await album.openAlbumWithUid(AlbumUid);
    await toolbar.triggerToolbarAction("edit");

    await t
      .expect(albumdialog.description.value)
      .eql("Last holiday")
      .expect(albumdialog.category.value)
      .eql("Mountains")
      .expect(albumdialog.location.value)
      .eql("United States");

    await t
      .typeText(albumdialog.title, "Kanada", { replace: true })
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
    await menu.openPage("folders");
    await toolbar.search("Kanada");

    await t
      .expect(page.cardTitle.nth(0).innerText)
      .contains("Kanada")
      .expect(page.cardDescription.nth(0).innerText)
      .notContains("We went to ski")
      .expect(Selector("div.caption").nth(0).innerText)
      .notContains("United States");
  }
);

test.meta("testID", "folders-003").meta({ mode: "public" })(
  "Common: Create, Edit, delete sharing link",
  async (t) => {
    await page.testCreateEditDeleteSharingLink("folders");
  }
);

test.meta("testID", "folders-004").meta({ mode: "public" })(
  "Common: Create/delete album-clone from folder",
  async (t) => {
    await menu.openPage("albums");
    const AlbumCount = await album.getAlbumCount("all");
    await menu.openPage("folders");
    const ThirdFolderUid = await album.getNthAlbumUid("all", 2);
    await album.openAlbumWithUid(ThirdFolderUid);
    const PhotoCountInFolder = await photo.getPhotoCount("all");
    const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
    await menu.openPage("folders");
    await album.selectAlbumFromUID(ThirdFolderUid);
    await contextmenu.triggerContextMenuAction("clone", "NotYetExistingAlbumForFolder");
    await menu.openPage("albums");
    const AlbumCountAfterCreation = await album.getAlbumCount("all");

    await t.expect(AlbumCountAfterCreation).eql(AlbumCount + 1);

    await toolbar.search("NotYetExistingAlbumForFolder");
    const AlbumUid = await album.getNthAlbumUid("all", 0);
    await album.openAlbumWithUid(AlbumUid);
    const PhotoCountInAlbum = await photo.getPhotoCount("all");

    await t.expect(PhotoCountInAlbum).eql(PhotoCountInFolder);

    await photo.checkPhotoVisibility(FirstPhotoUid, true);
    await menu.openPage("albums");
    await album.selectAlbumFromUID(AlbumUid);
    await contextmenu.triggerContextMenuAction("delete", "");
    const AlbumCountAfterDelete = await album.getAlbumCount("all");

    await t.expect(AlbumCountAfterDelete).eql(AlbumCount);

    await menu.openPage("folders");
    await album.openAlbumWithUid(ThirdFolderUid);
    await photo.checkPhotoVisibility(FirstPhotoUid, true);
  }
);

test.meta("testID", "folders-005").meta({ type: "short", mode: "public" })(
  "Common: Verify folder sort options",
  async (t) => {
    await menu.openPage("folders");
    await album.checkSortOptions("folder");
  }
);
