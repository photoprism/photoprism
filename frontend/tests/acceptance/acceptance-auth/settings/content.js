import { Selector } from "testcafe";
import testcafeconfig from "../../../testcafeconfig.json";
import Menu from "../../page-model/menu";
import Toolbar from "../../page-model/toolbar";
import Page from "../../page-model/page";
import PhotoEdit from "../../page-model/photo-edit";
import Settings from "../../page-model/settings";

fixture`Test content settings`.page`${testcafeconfig.url}`.beforeEach(async (t) => {
  await page.login("admin", "photoprism");
});

const menu = new Menu();
const toolbar = new Toolbar();
const page = new Page();
const photoedit = new PhotoEdit();
const settings = new Settings();

test.meta("testID", "settings-content-001").meta({ mode: "auth" })("Common: Hide titles", async (t) => {
  await t.click(toolbar.cardsViewAction);
  await toolbar.search("Cat");
  await t.expect(page.cardTitle.withText("Cat / Greece / 2019").visible).ok();
  await t.expect(page.cardTitle.withText("IMG_7138.JPG").visible).notOk();
  await t.click(toolbar.listViewAction);
  await t.expect(Selector("td.meta-title").withText("Cat / Greece / 2019").visible).ok();
  await t.expect(Selector("td.meta-title").withText("IMG_7138.JPG").visible).notOk();

  await menu.openPage("settings");
  await t.click(settings.libraryTab).click(settings.hideTitlesCheckbox);

  await menu.openPage("browse");
  await t.click(toolbar.cardsViewAction);
  await toolbar.search("Cat");
  await t.expect(page.cardTitle.withText("Cat / Greece / 2019").visible).notOk();
  await t.expect(page.cardTitle.withText("IMG_7138.JPG").visible).ok();
  await t.click(toolbar.listViewAction);
  await t.expect(Selector("td.meta-title").withText("Cat / Greece / 2019").visible).notOk();
  await t.expect(Selector("td.meta-title").withText("IMG_7138.JPG").visible).ok();

  await menu.openPage("settings");
  await t.click(settings.libraryTab).click(settings.hideTitlesCheckbox);

  await menu.openPage("browse");
  await t.click(toolbar.cardsViewAction);
  await toolbar.search("Cat");
  await t.expect(page.cardTitle.withText("Cat / Greece / 2019").visible).ok();
  await t.expect(page.cardTitle.withText("IMG_7138.JPG").visible).notOk();
  await t.click(toolbar.listViewAction);
  await t.expect(Selector("td.meta-title").withText("Cat / Greece / 2019").visible).ok();
  await t.expect(Selector("td.meta-title").withText("IMG_7138.JPG").visible).notOk();
});

test.meta("testID", "settings-content-002").meta({ mode: "auth" })("Common: Hide captions", async (t) => {
  await t.click(toolbar.cardsViewAction);
  await toolbar.search("Cat");
  await t
    .click(page.cardTitle.nth(0))
    .typeText(photoedit.description, "A cute cat in the sun", {
      replace: true,
    })
    .click(photoedit.detailsApply)
    .click(Selector("button.action-close"));

  await t.expect(page.cardCaption.withText("A cute cat in the sun").visible).ok();

  await menu.openPage("settings");
  await t.click(settings.libraryTab).click(settings.hideCaptionsCheckbox);

  await menu.openPage("browse");
  await t.click(toolbar.cardsViewAction);
  await toolbar.search("Cat");
  await t.expect(page.cardCaption.withText("A cute cat in the sun").visible).notOk();

  await menu.openPage("settings");
  await t.click(settings.libraryTab).click(settings.hideCaptionsCheckbox);

  await menu.openPage("browse");
  await t.click(toolbar.cardsViewAction);
  await toolbar.search("Cat");
  await t.expect(page.cardCaption.withText("A cute cat in the sun").visible).ok();
});

test.meta("testID", "settings-content-003").meta({ mode: "auth" })("Common: List view", async (t) => {
  await t.expect(toolbar.listViewAction.visible).ok();

  await menu.openPage("settings");
  await t.click(settings.libraryTab).click(settings.hideListViewCheckbox);
  await menu.openPage("browse");

  await t.expect(toolbar.listViewAction.visible).notOk();

  await menu.openPage("settings");
  await t.click(settings.libraryTab).click(settings.hideListViewCheckbox);
  await menu.openPage("browse");

  await t.expect(toolbar.listViewAction.visible).ok();
});
