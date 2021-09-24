import { Selector } from "testcafe";
import testcafeconfig from "../testcafeconfig";
import Page from "../page-model";

fixture`Test index`.page`${testcafeconfig.url}`;

const page = new Page();
test.meta("testID", "library-index-001")("Index files from folder", async (t) => {
  await page.openNav();
  await t.click(Selector(".nav-labels"));
  await page.search("cheetah");
  await t.expect(Selector("div.no-results").visible).ok();
  await page.openNav();
  await t.click(Selector(".nav-moments"));
  const MomentCount = await Selector("a.is-album").count;
  await page.openNav();
  await t.click(Selector(".nav-calendar"));
  if (t.browser.platform === "mobile") {
    await t.navigateTo("/calendar?q=December%202013");
  } else {
    await page.search("December 2013");
  }
  await t.expect(Selector("div.no-results").visible).ok();
  await page.openNav();
  await t.click(Selector(".nav-folders"));
  if (t.browser.platform === "mobile") {
    await t.navigateTo("/folders?q=moment");
  } else {
    await page.search("Moment");
  }
  await t.expect(Selector("div.no-results").visible).ok();
  await page.openNav();
  await t.click(Selector(".nav-places + div > i")).click(Selector(".nav-states"));
  if (t.browser.platform === "mobile") {
    console.log(t.browser.platform);
    await t.navigateTo("/states?q=KwaZulu");
  } else {
    await page.search("KwaZulu");
  }
  await t.expect(Selector("div.no-results").visible).ok();
  await page.openNav();
  await t
    .click(Selector(".nav-library+div>i"))
    .click(Selector(".nav-originals"))
    .click(Selector(".is-folder").withText("moment"))
    .expect(Selector("div.no-results").visible)
    .ok();
  await page.openNav();
  await t.click(Selector(".nav-browse + div")).click(Selector(".nav-monochrome"));
  const MonochromeCount = await Selector("a.is-album").count;
  await page.openNav();
  await t
    .click(Selector(".nav-library"))
    .click(Selector("#tab-library-index"))
    .click(Selector(".input-index-folder input"))
    .click(Selector("div.v-list__tile__title").withText("/moment"))
    .click(Selector(".action-index"))
    //TODO replace wait
    .wait(50000)
    .expect(Selector("span").withText("Done.").visible, { timeout: 60000 })
    .ok();
  await page.openNav();
  await t.click(Selector(".nav-labels")).click(Selector(".action-reload"));
  await page.search("cheetah");
  await t.expect(Selector(".is-label").visible).ok();
  await page.openNav();
  await t.click(Selector(".nav-moments"));
  const MomentCountAfterIndex = await Selector("a.is-album").count;
  await t
    .expect(MomentCountAfterIndex)
    .gt(MomentCount)
    .click(Selector("a").withText("South Africa 2013"))
    .expect(Selector(".is-photo").visible)
    .ok();
  await page.openNav();
  await t.click(Selector(".nav-calendar")).click(Selector(".action-reload"));
  if (t.browser.platform === "mobile") {
    console.log(t.browser.platform);
    await t.navigateTo("/calendar?q=December%202013");
  } else {
    await page.search("December 2013");
  }
  await t.expect(Selector(".is-album").visible).ok();
  await page.openNav();
  await t.click(Selector(".nav-folders")).click(Selector(".action-reload"));
  if (t.browser.platform === "mobile") {
    console.log(t.browser.platform);
    await t.navigateTo("/folders?q=moment");
  } else {
    await page.search("Moment");
  }
  await t.expect(Selector(".is-album").visible).ok();
  await page.openNav();
  await t
    .click(Selector(".nav-places+div>i"))
    .click(Selector(".nav-states"))
    .click(Selector(".action-reload"));
  if (t.browser.platform === "mobile") {
    console.log(t.browser.platform);
    await t.navigateTo("/states?q=KwaZulu");
  } else {
    await page.search("KwaZulu");
  }
  await t.expect(Selector(".is-album").visible).ok();
  await page.openNav();
  await t
    .click(Selector(".nav-library+div>i"))
    .click(Selector(".nav-originals"))
    .click(Selector(".action-reload"))
    .expect(Selector(".is-folder").withText("moment").visible, { timeout: 60000 })
    .ok();
  await page.openNav();
  await t.click(Selector(".nav-browse + div")).click(Selector(".nav-monochrome"));
  const MonochromeCountAfterIndex = await Selector(".is-photo.type-image", { timeout: 5000 }).count;
  await t.expect(MonochromeCountAfterIndex).gt(MonochromeCount);
});
