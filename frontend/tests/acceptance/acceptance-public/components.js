import { Selector } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";
import Toolbar from "../page-model/toolbar";
import Menu from "../page-model/menu";
import Photoviewer from "../page-model/photoviewer";

fixture`Test components`.page`${testcafeconfig.url}`;

const toolbar = new Toolbar();
const menu = new Menu();
const photoviewer = new Photoviewer();

test.meta("testID", "components-001").meta({ type: "short", mode: "public" })(
  "Common: Test filter options",
  async (t) => {
    await t.expect(Selector("body").withText("object Object").exists).notOk();
  }
);

test.meta("testID", "components-002").meta({ type: "short", mode: "public" })(
  "Common: Fullscreen mode",
  async (t) => {
    await toolbar.search("photo:true");

    await photoviewer.openPhotoViewer("nth", 0);

    if (await Selector("#photo-viewer").visible) {
      await t
        .expect(Selector("#photo-viewer").visible)
        .ok()
        .expect(Selector("img.pswp__img").visible)
        .ok();
    } else {
      await t.expect(Selector("div.video-viewer").visible).ok();
    }
  }
);

test.meta("testID", "components-003").meta({ type: "short", mode: "public" })(
  "Common: Mosaic view",
  async (t) => {
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
  }
);

test.meta("testID", "components-004").meta({ mode: "public" })("Common: List view", async (t) => {
  await toolbar.setFilter("view", "List");

  await t
    .expect(Selector("table.v-datatable").visible)
    .ok()
    .expect(Selector("div.list-view").visible)
    .ok();
});

test.meta("testID", "components-005").meta({ type: "short", mode: "public" })(
  "Common: Card view",
  async (t) => {
    await toolbar.setFilter("view", "Cards");
    await toolbar.search("photo:true");

    await t
      .expect(Selector("div.type-image div.clickable").visible)
      .ok()
      .expect(Selector("div.is-photo div.caption").visible)
      .ok()
      .expect(Selector("#photo-viewer").visible)
      .notOk();
  }
);

test.meta("testID", "components-006").meta({ mode: "public" })(
  "Common: Mobile Toolbar",
  async (t) => {
    if (t.browser.platform === "mobile") {
      await menu.openPage("browse");
      if (await toolbar.openMobileToolbar.visible) {
        await t.click(toolbar.openMobileToolbar);
      }
      await toolbar.checkMobileMenuActionAvailability("login", false);
      await toolbar.checkMobileMenuActionAvailability("logout", false);
      await toolbar.checkMobileMenuActionAvailability("reload", true);
      await toolbar.checkMobileMenuActionAvailability("logs", true);
      await toolbar.checkMobileMenuActionAvailability("settings", true);
      await toolbar.checkMobileMenuActionAvailability("upload", true);
      await toolbar.checkMobileMenuActionAvailability("reload", true);
      await toolbar.checkMobileMenuActionAvailability("search", false);
      await toolbar.checkMobileMenuActionAvailability("albums", true);
      await toolbar.checkMobileMenuActionAvailability("library", true);
      await toolbar.checkMobileMenuActionAvailability("files", true);
      await toolbar.checkMobileMenuActionAvailability("sync", true);
      await toolbar.checkMobileMenuActionAvailability("account", true);
      await toolbar.checkMobileMenuActionAvailability("manual", true);

      await toolbar.triggerMobileMenuAction("albums");
      if (await toolbar.openMobileToolbar.visible) {
        await t.click(toolbar.openMobileToolbar);
      }
      await toolbar.checkMobileMenuActionAvailability("login", false);
      await toolbar.checkMobileMenuActionAvailability("logout", false);
      await toolbar.checkMobileMenuActionAvailability("reload", true);
      await toolbar.checkMobileMenuActionAvailability("logs", true);
      await toolbar.checkMobileMenuActionAvailability("settings", true);
      await toolbar.checkMobileMenuActionAvailability("upload", true);
      await toolbar.checkMobileMenuActionAvailability("reload", true);
      await toolbar.checkMobileMenuActionAvailability("search", true);
      await toolbar.checkMobileMenuActionAvailability("albums", false);
      await toolbar.checkMobileMenuActionAvailability("library", true);
      await toolbar.checkMobileMenuActionAvailability("files", true);
      await toolbar.checkMobileMenuActionAvailability("sync", true);
      await toolbar.checkMobileMenuActionAvailability("account", true);
      await toolbar.checkMobileMenuActionAvailability("manual", true);
      await toolbar.triggerMobileMenuAction("logs");
      if (await toolbar.openMobileToolbar.visible) {
        await t.click(toolbar.openMobileToolbar);
      }
      await toolbar.checkMobileMenuActionAvailability("login", false);
      await toolbar.checkMobileMenuActionAvailability("logout", false);
      await toolbar.checkMobileMenuActionAvailability("reload", true);
      await toolbar.checkMobileMenuActionAvailability("logs", false);
      await toolbar.checkMobileMenuActionAvailability("settings", true);
      await toolbar.checkMobileMenuActionAvailability("upload", true);
      await toolbar.checkMobileMenuActionAvailability("reload", true);
      await toolbar.checkMobileMenuActionAvailability("search", true);
      await toolbar.checkMobileMenuActionAvailability("albums", true);
      await toolbar.checkMobileMenuActionAvailability("library", false);
      await toolbar.checkMobileMenuActionAvailability("files", true);
      await toolbar.checkMobileMenuActionAvailability("sync", true);
      await toolbar.checkMobileMenuActionAvailability("account", true);
      await toolbar.checkMobileMenuActionAvailability("manual", true);
    }
  }
);
