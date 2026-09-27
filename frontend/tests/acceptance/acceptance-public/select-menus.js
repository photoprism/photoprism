import { Selector } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";
import Menu from "../page-model/menu";
import Toolbar from "../page-model/toolbar";
import Photo from "../page-model/photo";
import Page from "../page-model/page";
import PhotoEdit from "../page-model/photo-edit";
import { helperBeforeFixture, helperBeforeEach, helperAfterEach } from "../page-model/helpers";

fixture`Test select menus`
  .page`${testcafeconfig.url}`
  .beforeEach(async (t) => {
    await helperBeforeEach(t);
  })
  .afterEach(async (t) => {
    await helperAfterEach(t);
  })
  .before(async (ctx) => {
    await helperBeforeFixture(ctx);
  });

const menu = new Menu();
const toolbar = new Toolbar();
const photo = new Photo();
const page = new Page();
const photoedit = new PhotoEdit();

const country = Selector("div.v-dialog .input-country", { timeout: 15000 });
const timezone = Selector("div.v-dialog .input-timezone", { timeout: 15000 });
const menuItems = Selector("div.v-overlay--active .v-list-item");

test.meta("testID", "select-menus-001").meta({ type: "short", mode: "public" })(
  "Common: Keeps long autocomplete menus open, filters them, and switches between them",
  async (t) => {
    await menu.openPage("browse");
    await t.click(toolbar.cardsViewAction);
    // Country is read-only for photos with coordinates.
    await toolbar.search("geo:false");
    const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
    await page.clickCardTitleOfUID(FirstPhotoUid);
    await t.expect(country.visible).ok();

    // Both menus must stay open after opening, which failed for long lists in some Vuetify versions.
    await t.click(country.find(".v-field")).wait(1000);
    await t.expect(country.hasClass("v-autocomplete--active-menu")).ok().expect(menuItems.count).gt(0);
    await t.pressKey("esc");
    await t.expect(country.hasClass("v-autocomplete--active-menu")).notOk().expect(country.visible).ok();

    await t.click(timezone.find(".v-field")).wait(1000);
    await t.expect(timezone.hasClass("v-autocomplete--active-menu")).ok().expect(menuItems.count).gt(0);

    // Opening a sibling closes the other menu and keeps the new one open.
    await t.click(country.find(".v-field")).wait(1000);
    await t
      .expect(country.hasClass("v-autocomplete--active-menu"))
      .ok()
      .expect(timezone.hasClass("v-autocomplete--active-menu"))
      .notOk();

    // Typing filters the list.
    await t.typeText(country.find("input:not([type='hidden'])"), "Germ", { replace: true }).wait(500);
    await t.expect(menuItems.withText("Germany").visible).ok().expect(menuItems.count).lt(5);
    await t.pressKey("esc");

    await t.click(photoedit.dialogClose);
  }
);
