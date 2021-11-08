import { Selector } from "testcafe";
import { ClientFunction } from "testcafe";
import testcafeconfig from "./testcafeconfig.json";
import Page from "./page-model";

const getLocation = ClientFunction(() => document.location.href);

fixture`Test places page`.page`${testcafeconfig.url}`;
const page = new Page();

test.meta("testID", "places-001")("Test places", async (t) => {
  await page.openNav();
  await t
    .click(Selector(".nav-places"))
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
  await page.openNav();
  await t.click(Selector(".nav-browse")).expect(Selector("div.is-photo").exists).ok();
  await page.openNav();
  await t
    .click(Selector(".nav-places"))
    .expect(Selector("#map").exists, { timeout: 15000 })
    .ok()
    .expect(Selector("div.p-map-control").visible)
    .ok();
});

test.meta("testID", "places-002")("Open photo from places", async (t) => {
  //TODO replace wait
  if (t.browser.name === "Firefox") {
    console.log("Test skipped in firefox");
  } else {
    await page.openNav();
    await t
      .click(Selector(".nav-places"))
      .expect(Selector("#is-photo-viewer").visible)
      .notOk()
      .expect(Selector("#map").exists, { timeout: 15000 })
      .ok()
      .typeText(Selector('input[aria-label="Search"]'), "Berlin")
      .pressKey("enter")
      .wait(30000)
      .click(Selector("div.marker").nth(0), { timeout: 9000 })
      .expect(Selector("#photo-viewer").visible)
      .ok();
  }
});
