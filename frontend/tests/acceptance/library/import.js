import { Selector } from "testcafe";
import testcafeconfig from "../testcafeconfig";
import Page from "../page-model";

fixture`Import file from folder`.page`${testcafeconfig.url}`;

const page = new Page();
test.meta("testID", "library-import-001")("Import files from folder using copy", async (t) => {
  await t.click(Selector(".nav-labels"));
  await page.search("bakery");
  await t.expect(Selector("h3").withText("Couldn't find anything").visible).ok();
  await t
    .click(Selector(".nav-library"))
    .click(Selector("#tab-library-import"))
    .click(Selector(".input-import-folder input"), { timeout: 5000 })
    .click(Selector("div.v-list__tile__title").withText("/Bäckerei"))
    .click(Selector(".action-import"))
    //TODO replace wait
    .wait(60000)
    //.expect(Selector('span').withText('Done.').visible, {timeout: 60000}).ok()
    .click(Selector(".nav-labels"))
    .click(Selector(".action-reload"));
  await page.search("bakery");
  await t.expect(Selector(".is-label").visible).ok();
});
