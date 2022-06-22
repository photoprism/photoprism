import { Selector } from "testcafe";
import testcafeconfig from "./testcafeconfig";
import Toolbar from "../page-model/toolbar";

fixture`Test components`.page`${testcafeconfig.url}`;

const toolbar = new Toolbar();

test.meta("testID", "components-001").meta({ type: "smoke" })("Test filter options", async (t) => {
  await t.expect(Selector("body").withText("object Object").exists).notOk();
});

test.meta("testID", "components-002").meta({ type: "smoke" })("Fullscreen mode", async (t) => {
  await t.click(Selector("div.type-image div.image.clickable").nth(0));

  if (await Selector("#photo-viewer").visible) {
    await t
      .expect(Selector("#photo-viewer").visible)
      .ok()
      .expect(Selector("img.pswp__img").visible)
      .ok();
  } else {
    await t.expect(Selector("div.video-viewer").visible).ok();
  }
});

test.meta("testID", "components-003").meta({ type: "smoke" })("Mosaic view", async (t) => {
  await toolbar.setFilter("view", "Mosaic");

  await t
    .expect(Selector("div.type-image.image.clickable").visible)
    .ok()
    .expect(Selector("div.p-photo-mosaic").visible)
    .ok()
    .expect(Selector("div.is-photo div.caption").exists)
    .notOk()
    .expect(Selector("#photo-viewer").visible)
    .notOk();
});

test.meta("testID", "components-004")("List view", async (t) => {
  await toolbar.setFilter("view", "List");

  await t
    .expect(Selector("table.v-datatable").visible)
    .ok()
    .expect(Selector("div.list-view").visible)
    .ok();
});

test.meta("testID", "components-005").meta({ type: "smoke" })("Card view", async (t) => {
  await toolbar.setFilter("view", "Cards");

  await t
    .expect(Selector("div.type-image div.image.clickable").visible)
    .ok()
    .expect(Selector("div.is-photo div.caption").visible)
    .ok()
    .expect(Selector("#photo-viewer").visible)
    .notOk();
});
