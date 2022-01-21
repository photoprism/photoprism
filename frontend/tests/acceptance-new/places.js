import { Selector } from "testcafe";
import { ClientFunction } from "testcafe";
import testcafeconfig from "./testcafeconfig.json";
import Menu from "../page-model/menu";

const getLocation = ClientFunction(() => document.location.href);

//TODO not working in container
fixture`Search and open photo from places`.page`${testcafeconfig.url}`.skip(
  "Places to loading in container"
);

const menu = new Menu();

test.meta("testID", "places-001")("Test places", async (t) => {
  await menu.openPage("places");
  await t
    .expect(Selector("#map").exists, { timeout: 15000 })
    .ok()
    .expect(Selector("div.p-map-control").visible)
    .ok()
    .takeScreenshot({ path: "./places-neew.png" });
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
    .ok()
    .typeText(Selector('input[aria-label="Search"]'), "canada", { replace: true })
    .pressKey("enter")
    .wait(8000)
    .expect(Selector('div[title="Cape / Bowen Island / 2019"]').visible)
    .ok()
    .click(Selector('div[title="Cape / Bowen Island / 2019"]'))
    .expect(Selector("#photo-viewer").visible)
    .ok();
});
