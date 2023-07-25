import { Selector } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";
import Menu from "../page-model/menu";
import Album from "../page-model/album";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import Photo from "../page-model/photo";
import Page from "../page-model/page";
import AlbumDialog from "../page-model/dialog-album";

fixture`Test calendar`.page`${testcafeconfig.url}`;

const menu = new Menu();
const album = new Album();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const photo = new Photo();
const page = new Page();
const albumdialog = new AlbumDialog();

test.meta("testID", "calendar-001").meta({ type: "short", mode: "public" })(
  "Common: View calendar",
  async (t) => {
    await menu.openPage("calendar");

    await t
      .expect(Selector("a").withText("May 2019").visible)
      .ok()
      .expect(Selector("a").withText("October 2019").visible)
      .ok();
  }
);

test.meta("testID", "calendar-002").meta({ mode: "public" })(
  "Common: Update calendar details",
  async (t) => {
    await menu.openPage("calendar");
    await toolbar.search("March 2014");
    const AlbumUid = await album.getNthAlbumUid("all", 0);

    await t.expect(page.cardTitle.nth(0).innerText).contains("March 2014");

    await t.click(page.cardTitle.nth(0)).typeText(albumdialog.location, "Snow", { replace: true });

    await t
      .expect(albumdialog.description.value)
      .eql("")
      .expect(albumdialog.category.value)
      .eql("");

    await t
      .typeText(albumdialog.description, "We went to ski")
      .typeText(albumdialog.category, "Mountains")
      .pressKey("enter")
      .click(albumdialog.dialogSave);

    await t
      .expect(page.cardTitle.nth(0).innerText)
      .contains("March 2014")
      .expect(page.cardDescription.nth(0).innerText)
      .contains("We went to ski")
      .expect(Selector("div.caption").nth(1).innerText)
      .contains("Mountains")
      .expect(Selector("div.caption").nth(2).innerText)
      .contains("Snow");

    await album.openNthAlbum(0);

    await t.expect(toolbar.toolbarSecondTitle.innerText).contains("March 2014");
    await t.expect(toolbar.toolbarDescription.innerText).contains("We went to ski");
    await menu.openPage("calendar");
    if (t.browser.platform === "mobile") {
      await toolbar.search("category:Mountains");
    } else {
      await toolbar.setFilter("category", "Mountains");
    }

    await t.expect(page.cardTitle.nth(0).innerText).contains("March 2014");

    await album.openAlbumWithUid(AlbumUid);
    await toolbar.triggerToolbarAction("edit");

    await t
      .expect(albumdialog.description.value)
      .eql("We went to ski")
      .expect(albumdialog.category.value)
      .eql("Mountains")
      .expect(albumdialog.location.value)
      .eql("Snow");

    await t
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
    await menu.openPage("calendar");
    await toolbar.search("March 2014");

    await t
      .expect(page.cardDescription.innerText)
      .notContains("We went to ski")
      .expect(Selector("div.caption").nth(0).innerText)
      .notContains("Snow");
  }
);

test.meta("testID", "calendar-003").meta({ mode: "public" })(
  "Common: Create, Edit, delete sharing link for calendar",
  async (t) => {
    await page.testCreateEditDeleteSharingLink("calendar");
  }
);

test.meta("testID", "calendar-004").meta({ type: "short", mode: "public" })(
  "Common: Create/delete album-clone from calendar",
  async (t) => {
    await menu.openPage("albums");
    const AlbumCount = await album.getAlbumCount("all");
    await menu.openPage("calendar");
    const SecondCalendarUid = await album.getNthAlbumUid("all", 1);
    await album.openAlbumWithUid(SecondCalendarUid);
    const PhotoCountInCalendar = await photo.getPhotoCount("all");
    const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
    const SecondPhotoUid = await photo.getNthPhotoUid("image", 1);
    await menu.openPage("calendar");
    await album.selectAlbumFromUID(SecondCalendarUid);
    await contextmenu.triggerContextMenuAction("clone", "NotYetExistingAlbumForCalendar");
    await menu.openPage("albums");
    const AlbumCountAfterCreation = await album.getAlbumCount("all");

    await t.expect(AlbumCountAfterCreation).eql(AlbumCount + 1);

    await toolbar.search("NotYetExistingAlbumForCalendar");
    const AlbumUid = await album.getNthAlbumUid("all", 0);
    await album.openAlbumWithUid(AlbumUid);
    const PhotoCountInAlbum = await photo.getPhotoCount("all");

    await t.expect(PhotoCountInAlbum).eql(PhotoCountInCalendar);

    await photo.checkPhotoVisibility(FirstPhotoUid, true);
    await photo.checkPhotoVisibility(SecondPhotoUid, true);
    await menu.openPage("albums");
    await album.selectAlbumFromUID(AlbumUid);
    await contextmenu.triggerContextMenuAction("delete", "");
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await toolbar.triggerToolbarAction("reload");
    }
    const AlbumCountAfterDelete = await album.getAlbumCount("all");
    await t.expect(AlbumCountAfterDelete).eql(AlbumCount);
    await menu.openPage("calendar");
    await album.openAlbumWithUid(SecondCalendarUid);
    await photo.checkPhotoVisibility(FirstPhotoUid, true);
    await photo.checkPhotoVisibility(SecondPhotoUid, true);
  }
);

test.meta("testID", "calendar-005").meta({ type: "short", mode: "public" })(
  "Common: Verify calendar sort options",
  async (t) => {
    await menu.openPage("calendar");
    await album.checkSortOptions("calendar");
  }
);
