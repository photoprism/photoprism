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
import Label from "../page-model/label";
import PhotoViews from "../page-model/photo-views";

fixture`Test labels`.page`${testcafeconfig.url}`;

const page = new Page();
const menu = new Menu();
const album = new Album();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const photo = new Photo();
const photoviewer = new PhotoViewer();
const newpage = new NewPage();
const label = new Label();
const photoviews = new PhotoViews();

test.meta("testID", "labels-001")("Remove/Activate Add/Delete Label from photo", async (t) => {
  await menu.openPage("labels");
  const countImportantLabels = await label.getLabelCount();
  await toolbar.triggerToolbarAction("show-all", "");
  const countAllLabels = await label.getLabelCount();
  await t.expect(countAllLabels).gt(countImportantLabels);
  await toolbar.triggerToolbarAction("show-important", "");
  await toolbar.search("beacon");
  const LabelBeacon = await label.getNthLabeltUid(0);
  await label.openLabelWithUid(LabelBeacon);
  await toolbar.setFilter("view", "Cards");
  const PhotoBeacon = await photo.getNthPhotoUid("all", 0);
  await t.click(newpage.cardTitle.withAttribute("data-uid", PhotoBeacon));
  const PhotoKeywords = await Selector(".input-keywords textarea").value;
  await t
    .expect(PhotoKeywords)
    .contains("beacon")
    .click(Selector("#tab-labels"))
    .click(Selector("button.action-remove"), { timeout: 5000 })
    .typeText(Selector(".input-label input"), "Test")
    .click(Selector("button.p-photo-label-add"))
    .click(Selector("#tab-details"));
  const PhotoKeywordsAfterEdit = await Selector(".input-keywords textarea").value;
  await t
    .expect(PhotoKeywordsAfterEdit)
    .contains("test")
    .expect(PhotoKeywordsAfterEdit)
    .notContains("beacon")
    .click(Selector(".action-close"));
  await menu.openPage("labels");
  await toolbar.search("beacon");
  await t.expect(Selector("div.no-results").visible).ok();
  await toolbar.search("test");
  const LabelTest = await label.getNthLabeltUid(0);
  await label.openLabelWithUid(LabelTest);
  await t
    .click(newpage.cardTitle.withAttribute("data-uid", PhotoBeacon))
    .click(Selector("#tab-labels"))
    .click(Selector(".action-delete"), { timeout: 5000 })
    .click(Selector(".action-on"))
    .click(Selector("#tab-details"));
  const PhotoKeywordsAfterUndo = await Selector(".input-keywords textarea").value;
  await t
    .expect(PhotoKeywordsAfterUndo)
    .contains("beacon")
    .expect(PhotoKeywordsAfterUndo)
    .notContains("test")
    .click(Selector(".action-close"));
  await menu.openPage("labels");
  await toolbar.search("test");
  await t.expect(Selector("div.no-results").visible).ok();
  await toolbar.search("beacon");
  await album.checkAlbumVisibility(LabelBeacon, true);
});

test.meta("testID", "labels-002")("Rename Label", async (t) => {
  await menu.openPage("labels");
  await toolbar.search("zebra");
  const LabelZebra = await label.getNthLabeltUid(0);
  await label.openNthLabel(0);
  const FirstPhotoZebra = await photo.getNthPhotoUid("all", 0);
  const SecondPhotoZebra = await photo.getNthPhotoUid("all", 1);
  await toolbar.setFilter("view", "Cards");
  await t.click(newpage.cardTitle.withAttribute("data-uid", FirstPhotoZebra));
  const FirstPhotoTitle = await Selector(".input-title input", { timeout: 5000 }).value;
  const FirstPhotoKeywords = await Selector(".input-keywords textarea", { timeout: 5000 }).value;
  await t
    .expect(FirstPhotoTitle)
    .contains("Zebra")
    .expect(FirstPhotoKeywords)
    .contains("zebra")
    .click(Selector("#tab-labels"))
    .click(Selector("div.p-inline-edit"))
    .typeText(Selector(".input-rename input"), "Horse", { replace: true })
    .pressKey("enter")
    .click(Selector("#tab-details"));
  const FirstPhotoTitleAfterEdit = await Selector(".input-title input", { timeout: 5000 }).value;
  const FirstPhotoKeywordsAfterEdit = await Selector(".input-keywords textarea", { timeout: 5000 })
    .value;
  await t
    .expect(FirstPhotoTitleAfterEdit)
    .contains("Horse")
    .expect(FirstPhotoKeywordsAfterEdit)
    .contains("horse")
    .expect(FirstPhotoTitleAfterEdit)
    .notContains("Zebra")
    .click(Selector(".action-close"));
  await menu.openPage("labels");
  await toolbar.search("horse");
  await album.checkAlbumVisibility(LabelZebra, true);
  await label.openLabelWithUid(LabelZebra);
  await photo.checkPhotoVisibility(FirstPhotoZebra, true);
  await t
    .click(newpage.cardTitle.withAttribute("data-uid", FirstPhotoZebra))
    .click(Selector("#tab-labels"))
    .click(Selector("div.p-inline-edit"))
    .typeText(Selector(".input-rename input"), "Zebra", { replace: true })
    .pressKey("enter")
    .click(Selector(".action-close"));
  await menu.openPage("labels");
  await page.search("horse");
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

  //await page.selectFromUID(LabelLandscape);

  //const clipboardCount = await Selector("span.count-clipboard");
  //await t.expect(clipboardCount.textContent).eql("1");
  await contextmenu.checkContextMenuCount("1");
  //await page.addSelectedToAlbum("Christmas", "album");
  await contextmenu.triggerContextMenuAction("album", "Christmas", "");
  await menu.openPage("albums");
  await album.openAlbumWithUid(AlbumUid);
  const PhotoCountAfterAdd = await photo.getPhotoCount("all");
  await t.expect(PhotoCountAfterAdd).eql(PhotoCount + 6);
  await photoviews.triggerHoverAction("uid", FirstPhotoLandscape, "select");
  await photoviews.triggerHoverAction("uid", SecondPhotoLandscape, "select");
  await photoviews.triggerHoverAction("uid", ThirdPhotoLandscape, "select");
  await photoviews.triggerHoverAction("uid", FourthPhotoLandscape, "select");
  await photoviews.triggerHoverAction("uid", FifthPhotoLandscape, "select");
  await photoviews.triggerHoverAction("uid", SixthPhotoLandscape, "select");
  await contextmenu.triggerContextMenuAction("remove", "");
  const PhotoCountAfterDelete = await photo.getPhotoCount("all");
  await t.expect(PhotoCountAfterDelete).eql(PhotoCountAfterAdd - 6);
});

test.meta("testID", "labels-004")("Delete label", async (t) => {
  await menu.openPage("labels");
  await toolbar.search("dome");
  const LabelDome = await label.getNthLabeltUid(0);
  await label.openLabelWithUid(LabelDome);
  const FirstPhotoDome = await photo.getNthPhotoUid("all", 0);
  await menu.openPage("labels");
  await label.triggerHoverAction("uid", LabelDome, "select");
  await contextmenu.checkContextMenuCount("1");
  await contextmenu.triggerContextMenuAction("delete", "", "");
  await toolbar.search("dome");
  await t.expect(Selector("div.no-results").visible).ok();
  await menu.openPage("browse");
  await toolbar.setFilter("view", "Cards");
  await t
    .click(newpage.cardTitle.withAttribute("data-uid", FirstPhotoDome))
    .click(Selector("#tab-labels"))
    .expect(Selector("td").withText("No labels found").visible)
    .ok()
    .typeText(Selector(".input-label input"), "Dome")
    .click(Selector("button.p-photo-label-add"));
});

/*Does not work on sqlite
test.skip("testID", "labels-005")("Check label count", async (t) => {
  await page.openNav();
  await t.click(Selector(".nav-labels"));
  await page.search("cat");
  const LabelCat = await Selector("a.is-label", { timeout: 55000 }).nth(0).getAttribute("data-uid");
  const CatCaption = await Selector("a[data-uid=" + LabelCat + "] div.caption").innerText;
  console.log(CatCaption);
  await t.click(Selector("a.is-label").withAttribute("data-uid", LabelCat));
  const countPhotosCat = await Selector("div.is-photo").count;
  await t.expect(CatCaption).contains(countPhotosCat.toString());
  console.log(countPhotosCat);
  await page.openNav();
  await t.click(Selector(".nav-labels"));
  await page.search("people");
  const LabelPeople = await Selector("a.is-label", { timeout: 55000 }).nth(0).getAttribute("data-uid");
  const PeopleCaption = await Selector("a[data-uid=" + LabelCat + "] div.caption").innerText;
  console.log(PeopleCaption);
  await t.click(Selector("a.is-label").withAttribute("data-uid", LabelPeople));
  const countPhotosPeople = await Selector("div.is-photo").count;
  await t.expect(CatCaption).contains(countPhotosPeople.toString());
  console.log(countPhotosPeople);
});*/
