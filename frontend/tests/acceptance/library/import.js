import { Selector } from "testcafe";
import testcafeconfig from "../testcafeconfig";
import Page from "../page-model";

fixture`Import file from folder`.page`${testcafeconfig.url}`;

const page = new Page();
test.meta("testID", "library-import-001")("Import files from folder using copy", async (t) => {
  await page.openNav();
  await t.click(Selector(".nav-labels"));
  await page.search("bakery");
  await t.expect(Selector("div.no-results").visible).ok();
  await page.openNav();
  await t
    .click(Selector(".nav-library"))
    .click(Selector("#tab-library-import"))
    .typeText(Selector(".input-import-folder input"), "/B", { replace: true })
    .click(Selector("div.v-list__tile__title").nth(0))
    .click(Selector(".action-import"))
    //TODO replace wait
    .wait(60000);
  await page.openNav();
  await t.click(Selector(".nav-labels")).click(Selector(".action-reload"));
  await page.search("bakery");
  await t.expect(Selector(".is-label").visible).ok();
});
