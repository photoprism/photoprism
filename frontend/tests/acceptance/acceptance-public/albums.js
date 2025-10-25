import { Selector } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";
import Menu from "../page-model/menu";
import Album from "../page-model/album";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import Photo from "../page-model/photo";
import PhotoViewer from "../page-model/photoviewer";
import Page from "../page-model/page";
import AlbumDialog from "../page-model/dialog-album";
import PhotoEdit from "../page-model/photo-edit";

fixture`Test albums`.page`${testcafeconfig.url}`;

const menu = new Menu();
const album = new Album();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const photo = new Photo();
const photoviewer = new PhotoViewer();
const page = new Page();
const albumdialog = new AlbumDialog();
const photoedit = new PhotoEdit();

test.meta("testID", "albums-001").meta({ type: "short", mode: "public" })(
  "Common: Create/delete album on /albums",
  async (t) => {
    await menu.openPage("albums");
    const AlbumCount = await album.getAlbumCount("all");
    await toolbar.triggerToolbarAction("add");
    const AlbumCountAfterCreate = await album.getAlbumCount("all");
    const NewAlbumUid = await album.getNthAlbumUid("all", 0);

    await t.expect(AlbumCountAfterCreate).eql(AlbumCount + 1);

    await album.selectAlbumFromUID(NewAlbumUid);
    await contextmenu.triggerContextMenuAction("delete", "");
    const AlbumCountAfterDelete = await album.getAlbumCount("all");

    await t.expect(AlbumCountAfterDelete).eql(AlbumCountAfterCreate - 1);
  }
);

test.meta("testID", "albums-002").meta({ type: "short", mode: "public" })(
  "Common: Create/delete album during add to album",
  async (t) => {
    await menu.openPage("albums");
    const AlbumCount = await album.getAlbumCount("all");
    await menu.openPage("browse");
    await toolbar.search("photo:true");
    const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
    const SecondPhotoUid = await photo.getNthPhotoUid("image", 1);

    await page.clickCardTitleOfUID(FirstPhotoUid);

    await t
      .click(photoedit.infoTab)
      .expect(Selector("td").withText("Albums").visible)
      .notOk()
      .expect(Selector("td").withText("NotYetExistingAlbum").visible)
      .notOk()
      .click(photoedit.dialogClose);

    await photo.selectPhotoFromUID(SecondPhotoUid);
    await photo.selectPhotoFromUID(FirstPhotoUid);
    await contextmenu.triggerContextMenuAction("album", "NotYetExistingAlbum");

    await page.clickCardTitleOfUID(FirstPhotoUid);

    await t
      .click(photoedit.infoTab)
      .expect(Selector("td").withText("Albums").visible)
      .ok()
      .expect(Selector("td").withText("NotYetExistingAlbum").visible)
      .ok()
      .click(photoedit.dialogClose);

    await menu.openPage("albums");
    const AlbumCountAfterCreation = await album.getAlbumCount("all");

    await t.expect(AlbumCountAfterCreation).eql(AlbumCount + 1);

    await toolbar.search("NotYetExistingAlbum");
    const AlbumUid = await album.getNthAlbumUid("all", 0);
    await album.openAlbumWithUid(AlbumUid);
    await toolbar.triggerToolbarAction("delete");
    const AlbumCountAfterDelete = await album.getAlbumCount("all");

    await t.expect(AlbumCountAfterDelete).eql(AlbumCount);

    await menu.openPage("browse");
    await toolbar.search("photo:true");
    await page.clickCardTitleOfUID(FirstPhotoUid);

    await t
      .click(photoedit.infoTab)
      .expect(Selector("td").withText("Albums").visible)
      .notOk()
      .expect(Selector("td").withText("NotYetExistingAlbum").visible)
      .notOk()
      .click(photoedit.dialogClose);
  }
);

test.meta("testID", "albums-003").meta({ type: "short", mode: "public" })(
  "Common: Update album details",
  async (t) => {
    await menu.openPage("albums");
    await toolbar.search("Holiday");
    const AlbumUid = await album.getNthAlbumUid("all", 0);

    await t.expect(page.cardTitle.nth(0).innerText).contains("Holiday");

    await t.click(page.cardTitle.nth(0)).typeText(albumdialog.title, "Animals", { replace: true });

    await t.expect(albumdialog.description.value).eql("").expect(albumdialog.category.value).eql("");

    await t
      .typeText(albumdialog.description, "All my animals")
      .typeText(albumdialog.category, "Pets")
      .pressKey("enter")
      .click(albumdialog.dialogSave);

    await t.expect(page.cardTitle.nth(0).innerText).contains("Animals");

    await album.openAlbumWithUid(AlbumUid);
    await toolbar.triggerToolbarAction("edit");
    await t.typeText(albumdialog.title, "Holiday", { replace: true });

    await t.expect(albumdialog.description.value).eql("All my animals").expect(albumdialog.category.value).eql("Pets");

    await t.click(albumdialog.description).pressKey("ctrl+a delete");
    await t
      .click(albumdialog.category)
      .click(albumdialog.category)
      .pressKey("ctrl+a delete")
      .click(Selector("form.form-album-edit i.mdi-bookmark"));
    await t.expect(albumdialog.category.value).eql("");
    await t.click(albumdialog.dialogSave);
    await menu.openPage("albums");

    await t
      .expect(Selector("div").withText("Holiday").visible)
      .ok()
      .expect(Selector("div").withText("Animals").exists)
      .notOk();
  }
);

test.meta("testID", "albums-004").meta({ type: "short", mode: "public" })(
  "Common: Add/Remove Photos to/from multiple albums",
  async (t) => {
    // Get initial counts for both Holiday and Christmas albums
    await menu.openPage("albums");
    await toolbar.search("Holiday");
    const HolidayAlbumUid = await album.getNthAlbumUid("all", 0);
    await album.openAlbumWithUid(HolidayAlbumUid);
    const HolidayPhotoCount = await photo.getPhotoCount("all");
    await menu.openPage("albums");
    await toolbar.search("Christmas");
    const ChristmasAlbumUid = await album.getNthAlbumUid("all", 0);
    await album.openAlbumWithUid(ChristmasAlbumUid);
    const ChristmasPhotoCount = await photo.getPhotoCount("all");

    // Select photos to add to albums
    await menu.openPage("browse");
    await toolbar.search("photo:true");
    const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
    const SecondPhotoUid = await photo.getNthPhotoUid("image", 1);

    // Verify photos are not in any albums initially
    await page.clickCardTitleOfUID(FirstPhotoUid);
    await t
      .click(photoedit.infoTab)
      .expect(Selector("td").withText("Albums").visible)
      .notOk()
      .expect(Selector("td").withText("Holiday").visible)
      .notOk()
      .expect(Selector("td").withText("Christmas").visible)
      .notOk()
      .click(photoedit.dialogClose);

    // Select both photos and add to multiple albums simultaneously
    await photo.selectPhotoFromUID(SecondPhotoUid);
    await photoviewer.openPhotoViewer("uid", FirstPhotoUid);
    await photoviewer.triggerPhotoViewerAction("select-toggle");
    await photoviewer.triggerPhotoViewerAction("close-button");
    await contextmenu.triggerContextMenuAction("album", ["Holiday", "Christmas", "Food"]);

    // Verify photos were added to Holiday album
    await menu.openPage("albums");
    await album.openAlbumWithUid(HolidayAlbumUid);
    const HolidayPhotoCountAfterAdd = await photo.getPhotoCount("all");
    await t.expect(HolidayPhotoCountAfterAdd).eql(HolidayPhotoCount + 2);

    // Verify photos were added to Christmas album
    await menu.openPage("albums");
    await album.openAlbumWithUid(ChristmasAlbumUid);
    const ChristmasPhotoCountAfterAdd = await photo.getPhotoCount("all");
    await t.expect(ChristmasPhotoCountAfterAdd).eql(ChristmasPhotoCount + 2);

    // Verify photo info shows all albums
    await menu.openPage("browse");
    await toolbar.search("photo:true");
    await page.clickCardTitleOfUID(FirstPhotoUid);
    await t
      .click(photoedit.infoTab)
      .expect(Selector("td").withText("Albums").visible)
      .ok()
      .expect(Selector("td").withText("Holiday").visible)
      .ok()
      .expect(Selector("td").withText("Food").visible)
      .ok()
      .expect(Selector("td").withText("Christmas").visible)
      .ok()
      .click(photoedit.dialogClose);

    // Remove photos from Holiday album and verify count
    await menu.openPage("albums");
    await album.openAlbumWithUid(HolidayAlbumUid);
    await photo.selectPhotoFromUID(FirstPhotoUid);
    await photo.selectPhotoFromUID(SecondPhotoUid);
    await contextmenu.triggerContextMenuAction("remove", "");
    const HolidayPhotoCountAfterRemove = await photo.getPhotoCount("all");
    await t.expect(HolidayPhotoCountAfterRemove).eql(HolidayPhotoCountAfterAdd - 2);

    // Verify photos are still in Christmas album
    await menu.openPage("albums");
    await album.openAlbumWithUid(ChristmasAlbumUid);
    const ChristmasPhotoCountAfterHolidayRemove = await photo.getPhotoCount("all");
    await t.expect(ChristmasPhotoCountAfterHolidayRemove).eql(ChristmasPhotoCountAfterAdd);

    // Verify photo info shows only Christmas album now
    await menu.openPage("browse");
    await toolbar.search("photo:true");
    await page.clickCardTitleOfUID(FirstPhotoUid);
    await t
      .click(photoedit.infoTab)
      .expect(Selector("td").withText("Albums").visible)
      .ok()
      .expect(Selector("td").withText("Holiday").visible)
      .notOk()
      .expect(Selector("td").withText("Christmas").visible)
      .ok()
      .click(photoedit.dialogClose);

    // Remove photos from Christmas album to clean up
    await menu.openPage("albums");
    await album.openAlbumWithUid(ChristmasAlbumUid);
    await photo.selectPhotoFromUID(FirstPhotoUid);
    await photo.selectPhotoFromUID(SecondPhotoUid);
    await contextmenu.triggerContextMenuAction("remove", "");
    const ChristmasPhotoCountAfterRemove = await photo.getPhotoCount("all");
    await t.expect(ChristmasPhotoCountAfterRemove).eql(ChristmasPhotoCount);

    // Delete Food album
    await menu.openPage("albums");
    await toolbar.search("Food");
    const FoodUid = await album.getNthAlbumUid("all", 0);
    await album.selectAlbumFromUID(FoodUid);
    await contextmenu.triggerContextMenuAction("delete", "");

    // Final verification that photos are not in any albums
    await menu.openPage("browse");
    await toolbar.search("photo:true");
    await page.clickCardTitleOfUID(FirstPhotoUid);
    await t
      .click(photoedit.infoTab)
      .expect(Selector("td").withText("Albums").visible)
      .notOk()
      .expect(Selector("td").withText("Food").visible)
      .notOk()
      .expect(Selector("td").withText("Holiday").visible)
      .notOk()
      .expect(Selector("td").withText("Christmas").visible)
      .notOk()
      .click(photoedit.dialogClose);
  }
);

test.meta("testID", "albums-004-duplicate").meta({ type: "short", mode: "public" })(
  "Common: Album duplication when selecting from dropdown then typing same name",
  async (t) => {
    await menu.openPage("browse");
    await toolbar.search("photo:true");
    const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
    await photo.selectPhotoFromUID(FirstPhotoUid);

    await contextmenu.openContextMenu();
    await t.click(Selector("button.action-album"));

    await t.click(Selector(".input-albums input"));
    const holidayOption = Selector("div").withText("Holiday").parent('div[role="option"]');

    if (await holidayOption.visible) {
      await t.click(holidayOption);
      const afterDropdown = await Selector("span.v-chip").withText("Holiday").count;
      await t.expect(afterDropdown).eql(1, "Should have 1 chip after dropdown selection");

      await t.click(Selector(".input-albums input"));
      await t.typeText(Selector(".input-albums input"), "Holiday", { replace: true }).pressKey("enter");

      const afterTyping = await Selector("span.v-chip").withText("Holiday").count;
      await t.expect(afterTyping).eql(1, "Should still have only 1 chip after typing duplicate");
    }

    await t.click(Selector(".action-cancel"));
  }
);

test.meta("testID", "albums-005").meta({ mode: "public" })("Common: Use album search and filters", async (t) => {
  await menu.openPage("albums");
  if (t.browser.platform === "mobile") {
    await toolbar.search("category:Family");
  } else {
    await toolbar.setFilter("category", "Family");
  }

  await t.expect(page.cardTitle.nth(0).innerText).contains("Christmas");
  const AlbumCount = await album.getAlbumCount("all");
  await t.expect(AlbumCount).eql(1);

  if (t.browser.platform === "mobile") {
  } else {
    await toolbar.setFilter("category", "All Categories");
  }

  await toolbar.search("Holiday");

  await t.expect(page.cardTitle.nth(0).innerText).contains("Holiday");
  const AlbumCount2 = await album.getAlbumCount("all");
  await t.expect(AlbumCount2).eql(1);
});

test.meta("testID", "albums-006").meta({ mode: "public" })("Common: Test album autocomplete", async (t) => {
  await menu.openPage("browse");
  await toolbar.search("photo:true");
  const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
  await photo.selectPhotoFromUID(FirstPhotoUid);
  await contextmenu.openContextMenu();
  await t.click(Selector("button.action-album")).click(Selector(".input-albums input"));

  await t
    .expect(page.selectOption.withText("Holiday").visible)
    .ok()
    .expect(page.selectOption.withText("Christmas").visible)
    .ok();

  await t.typeText(Selector(".input-albums input"), "C", { replace: true });

  await t
    .expect(page.selectOption.withText("Holiday").visible)
    .notOk()
    .expect(page.selectOption.withText("Christmas").visible)
    .ok()
    .expect(page.selectOption.withText("C").visible)
    .ok();
});

test.meta("testID", "albums-007").meta({ type: "short", mode: "public" })(
  "Common: Create, Edit, delete sharing link",
  async (t) => {
    await page.testCreateEditDeleteSharingLink("albums");
  }
);

test.meta("testID", "albums-008").meta({ type: "short", mode: "public" })(
  "Common: Verify album sort options",
  async (t) => {
    await menu.openPage("albums");
    await album.checkSortOptions("album");
  }
);

test.meta("testID", "albums-009").meta({ type: "short", mode: "public" })(
  "Common: Set album cover from Albums Page",
  async (t) => {
    await page.testSetAlbumCover("albums");
  }
);
