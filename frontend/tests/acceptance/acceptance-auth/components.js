import { Selector } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";
import Toolbar from "../page-model/toolbar";
import Menu from "../page-model/menu";
import Page from "../page-model/page";

fixture`Test components`.page`${testcafeconfig.url}`;

const toolbar = new Toolbar();
const menu = new Menu();
const page = new Page();

test.meta("testID", "components-001").meta({ mode: "auth" })(
  "Common: Mobile Toolbar",
  async (t) => {
    if (t.browser.platform === "mobile") {
      await t.expect(toolbar.openMobileToolbar.visible).notOk();
      await page.login("admin", "photoprism");

      if (await toolbar.openMobileToolbar.visible) {
        await t.click(toolbar.openMobileToolbar);
      }
      await toolbar.checkMobileMenuActionAvailability("login", false);
      await toolbar.checkMobileMenuActionAvailability("logout", true);
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

      await toolbar.triggerMobileMenuAction("logout");
      await t
        .expect(page.usernameInput.visible)
        .ok()
        .expect(Selector(".input-search input").visible)
        .notOk();
    }
  }
);
