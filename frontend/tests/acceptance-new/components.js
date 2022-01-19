import { Selector } from "testcafe";
import testcafeconfig from "./testcafeconfig";
import Page from "./page-model";
import Toolbar from "../page-model/toolbar";
import PhotoViewer from "../page-model/photoviewer";

fixture`Test components`.page`${testcafeconfig.url}`;

const page = new Page();
const toolbar = new Toolbar();
const photoviewer = new PhotoViewer();

test.meta("testID", "components-001")("Test filter options", async (t) => {
  await t.expect(Selector("body").withText("object Object").exists).notOk();
});

test.meta("testID", "components-002")("Fullscreen mode", async (t) => {
  await t.click(Selector("div.v-image__image").nth(0));
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

test.meta("testID", "components-003")("Mosaic view", async (t) => {
  await toolbar.setFilter("view", "Mosaic");
  await t
    .expect(Selector("div.v-image__image").visible)
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

test.meta("testID", "components-005")("#Card view", async (t) => {
  await toolbar.setFilter("view", "Cards");
  await t
    .expect(Selector("div.v-image__image").visible)
    .ok()
    .expect(Selector("div.is-photo div.caption").visible)
    .ok()
    .expect(Selector("#photo-viewer").visible)
    .notOk();
});
