import { Selector } from "testcafe";
import { ClientFunction } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";
import Menu from "../page-model/menu";
import Places from "../page-model/places";

const getLocation = ClientFunction(() => document.location.href);

fixture`Search and open photo from places`.page`${testcafeconfig.url}`;

const menu = new Menu();
const places = new Places();

test.meta("testID", "places-001").meta({ mode: "public" })("Common: Test places", async (t) => {
  await menu.openPage("places");

  await t
    .expect(Selector("#map").exists, { timeout: 15000 })
    .ok()
    .expect(Selector("div.map-control").visible)
    .ok();

  await menu.openPage("browse");

  await t.expect(Selector("div.is-photo").exists).ok();

  await menu.openPage("places");

  await t
    .expect(Selector("#map").exists, { timeout: 15000 })
    .ok()
    .expect(Selector("div.map-control").visible)
    .ok();

  await places.search("canada");
  await t.wait(8000);

  await t.expect(Selector('div[title="Cape / Bowen Island / 2019"]').visible).ok();

  await t.click(Selector('div[title="Cape / Bowen Island / 2019"]'));

  await t.expect(Selector("#photo-viewer").visible).ok();
});
