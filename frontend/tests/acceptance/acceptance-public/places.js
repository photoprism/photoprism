import { Selector } from "testcafe";
import { ClientFunction } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";
import Menu from "../page-model/menu";

const getLocation = ClientFunction(() => document.location.href);

fixture`Search and open photo from places`.page`${testcafeconfig.url}`.skip(
  "Places don't loadin chrome from within the container"
);

const menu = new Menu();

test.meta("testID", "places-001").meta({ mode: "public" })("Common: Test places", async (t) => {
  await menu.openPage("places");

  await t
    .expect(Selector("#map").exists, { timeout: 15000 })
    .ok()
    .expect(Selector("div.p-map-control").visible)
    .ok();

  await t.typeText(Selector('input[aria-label="Search"]'), "Berlin").pressKey("enter");

  await t
    .expect(Selector("div.p-map-control").visible)
    .ok()
    .expect(getLocation())
    .contains("Berlin");

  await menu.openPage("browse");

  await t.expect(Selector("div.is-photo").exists).ok();

  await menu.openPage("places");

  await t
    .expect(Selector("#map").exists, { timeout: 15000 })
    .ok()
    .expect(Selector("div.p-map-control").visible)
    .ok();

  await t
    .typeText(Selector('input[aria-label="Search"]'), "canada", { replace: true })
    .pressKey("enter")
    .wait(8000);

  await t.expect(Selector('div[title="Cape / Bowen Island / 2019"]').visible).ok();

  await t.click(Selector('div[title="Cape / Bowen Island / 2019"]'));

  await t.expect(Selector("#photo-viewer").visible).ok();
});
