import { Selector } from "testcafe";
import testcafeconfig from "./testcafeconfig";
import Menu from "../page-model/menu";
import Album from "../page-model/album";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import Photo from "../page-model/photo";
import Page from "../page-model/page";
import Label from "../page-model/label";
import PhotoEdit from "../page-model/photo-edit";

fixture`Test labels`.page`${testcafeconfig.url}`;

const menu = new Menu();
const album = new Album();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const photo = new Photo();
const page = new Page();
const label = new Label();
const photoedit = new PhotoEdit();

test.meta("testID", "labels-001").meta({ type: "smoke" })(
  "Remove/Activate Add/Delete Label from photo",
  async (t) => {
    await menu.openPage("labels");
    await toolbar.search("beacon");
    const LabelBeaconUid = await label.getNthLabeltUid(0);
    await label.openLabelWithUid(LabelBeaconUid);
    await toolbar.setFilter("view", "Cards");
    const PhotoBeaconUid = await photo.getNthPhotoUid("all", 0);
    await t.click(page.cardTitle.withAttribute("data-uid", PhotoBeaconUid));
    const PhotoKeywords = await photoedit.keywords.value;

    await t.expect(PhotoKeywords).contains("beacon");

    await t
      .click(photoedit.labelsTab, { timeout: 7000 })
      .click(photoedit.removeLabel, { timeout: 7000 })
      .typeText(photoedit.inputLabelName, "Test")
      .click(Selector(photoedit.addLabel))
      .click(photoedit.detailsTab);
    const PhotoKeywordsAfterEdit = await photoedit.keywords.value;

    await t
      .expect(PhotoKeywordsAfterEdit)
      .contains("test")
      .expect(PhotoKeywordsAfterEdit)
      .notContains("beacon");

    await t.click(photoedit.dialogClose);
    await menu.openPage("labels");
    await toolbar.search("beacon");

    await t.expect(Selector("div.no-results").visible).ok();

    await toolbar.search("test");
    const LabelTest = await label.getNthLabeltUid(0);
    await label.openLabelWithUid(LabelTest);
    await toolbar.setFilter("view", "Cards");
    await t
      .click(page.cardTitle.withAttribute("data-uid", PhotoBeaconUid))
      .click(photoedit.labelsTab, { timeout: 7000 })
      .click(photoedit.deleteLabel, { timeout: 7000 })
      .click(photoedit.activateLabel, { timeout: 7000 })
      .click(photoedit.detailsTab);
    const PhotoKeywordsAfterUndo = await photoedit.keywords.value;

    await t
      .expect(PhotoKeywordsAfterUndo)
      .contains("beacon")
      .expect(PhotoKeywordsAfterUndo)
      .notContains("test");

    await t.click(photoedit.dialogClose);
    await menu.openPage("labels");
    await toolbar.search("test");

    await t.expect(Selector("div.no-results").visible).ok();

    await toolbar.search("beacon");
    await album.checkAlbumVisibility(LabelBeaconUid, true);
  }
);

test.meta("testID", "labels-002")("Toggle between important and all labels", async (t) => {
  await menu.openPage("labels");
  const ImportantLabelsCount = await label.getLabelCount();
  await toolbar.triggerToolbarAction("show-all");
  const AllLabelsCount = await label.getLabelCount();

  await t.expect(AllLabelsCount).gt(ImportantLabelsCount);

  await toolbar.triggerToolbarAction("show-important");
  const ImportantLabelsCount2 = await label.getLabelCount();

  await t.expect(ImportantLabelsCount).eql(ImportantLabelsCount2);
});

test.meta("testID", "labels-003")("Rename Label", async (t) => {
  await menu.openPage("labels");
  await toolbar.search("zebra");
  const LabelZebraUid = await label.getNthLabeltUid(0);
  await label.openNthLabel(0);
  const FirstPhotoZebraUid = await photo.getNthPhotoUid("all", 0);
  await toolbar.setFilter("view", "Cards");
  await t.click(page.cardTitle.withAttribute("data-uid", FirstPhotoZebraUid));
  const FirstPhotoTitle = await photoedit.title.value;
  const FirstPhotoKeywords = await photoedit.keywords.value;

  await t.expect(FirstPhotoTitle).contains("Zebra").expect(FirstPhotoKeywords).contains("zebra");

  await t
    .click(photoedit.labelsTab)
    .click(photoedit.openInlineEdit)
    .typeText(photoedit.inputLabelRename, "Horse", { replace: true })
    .pressKey("enter")
    .click(photoedit.detailsTab);
  const FirstPhotoTitleAfterEdit = await photoedit.title.value;
  const FirstPhotoKeywordsAfterEdit = await photoedit.keywords.value;

  await t
    .expect(FirstPhotoTitleAfterEdit)
    .contains("Horse")
    .expect(FirstPhotoKeywordsAfterEdit)
    .contains("horse")
    .expect(FirstPhotoTitleAfterEdit)
    .notContains("Zebra");

  await t.click(photoedit.dialogClose);
  await menu.openPage("labels");
  await toolbar.search("horse");
  await album.checkAlbumVisibility(LabelZebraUid, true);
  await label.openLabelWithUid(LabelZebraUid);
  await toolbar.setFilter("view", "Cards");
  await photo.checkPhotoVisibility(FirstPhotoZebraUid, true);
  await t
    .click(page.cardTitle.withAttribute("data-uid", FirstPhotoZebraUid))
    .click(photoedit.labelsTab)
    .click(photoedit.openInlineEdit)
    .typeText(photoedit.inputLabelRename, "Zebra", { replace: true })
    .pressKey("enter")
    .click(photoedit.dialogClose);
  await menu.openPage("labels");
  await toolbar.search("horse");

  await t.expect(Selector("div.no-results").visible).ok();
});

test.meta("testID", "labels-003")("Add label to album", async (t) => {
  await menu.openPage("albums");
  await toolbar.search("Christmas");
  const AlbumUid = await album.getNthAlbumUid("all", 0);
  await album.openAlbumWithUid(AlbumUid);
  const PhotoCount = await photo.getPhotoCount("all");
  await menu.openPage("labels");
  await toolbar.search("landscape");
  const LabelLandscape = await label.getNthLabeltUid(1);
  await label.openLabelWithUid(LabelLandscape);
  const FirstPhotoLandscape = await photo.getNthPhotoUid("all", 0);
  const SecondPhotoLandscape = await photo.getNthPhotoUid("all", 1);
  const ThirdPhotoLandscape = await photo.getNthPhotoUid("all", 2);
  const FourthPhotoLandscape = await photo.getNthPhotoUid("all", 3);
  const FifthPhotoLandscape = await photo.getNthPhotoUid("all", 4);
  const SixthPhotoLandscape = await photo.getNthPhotoUid("all", 5);
  await menu.openPage("labels");
  await label.triggerHoverAction("uid", LabelLandscape, "select");
  await contextmenu.checkContextMenuCount("1");
  await contextmenu.triggerContextMenuAction("album", "Christmas");
  await menu.openPage("albums");
  await album.openAlbumWithUid(AlbumUid);
  const PhotoCountAfterAdd = await photo.getPhotoCount("all");

  await t.expect(PhotoCountAfterAdd).eql(PhotoCount + 6);

  await photo.triggerHoverAction("uid", FirstPhotoLandscape, "select");
  await photo.triggerHoverAction("uid", SecondPhotoLandscape, "select");
  await photo.triggerHoverAction("uid", ThirdPhotoLandscape, "select");
  await photo.triggerHoverAction("uid", FourthPhotoLandscape, "select");
  await photo.triggerHoverAction("uid", FifthPhotoLandscape, "select");
  await photo.triggerHoverAction("uid", SixthPhotoLandscape, "select");
  await contextmenu.triggerContextMenuAction("remove", "");
  const PhotoCountAfterDelete = await photo.getPhotoCount("all");

  await t.expect(PhotoCountAfterDelete).eql(PhotoCountAfterAdd - 6);
});

test.meta("testID", "labels-004")("Delete label", async (t) => {
  await menu.openPage("labels");
  await toolbar.search("dome");
  const LabelDomeUid = await label.getNthLabeltUid(0);
  await label.openLabelWithUid(LabelDomeUid);
  const FirstPhotoDomeUid = await photo.getNthPhotoUid("all", 0);
  await menu.openPage("labels");
  await label.triggerHoverAction("uid", LabelDomeUid, "select");
  await contextmenu.checkContextMenuCount("1");
  await contextmenu.triggerContextMenuAction("delete", "");
  await toolbar.search("dome");

  await t.expect(Selector("div.no-results").visible).ok();

  await menu.openPage("browse");
  await toolbar.search("uid:" + FirstPhotoDomeUid);
  await toolbar.setFilter("view", "Cards");
  await t
    .click(page.cardTitle.withAttribute("data-uid", FirstPhotoDomeUid))
    .click(photoedit.labelsTab);

  await t.expect(Selector("td").withText("No labels found").visible).ok();

  await t.typeText(photoedit.inputLabelName, "Dome").click(photoedit.addLabel);
});
