import { Selector } from "testcafe";
import testcafeconfig from "../../../testcafeconfig.json";
import Menu from "../../page-model/menu";

fixture`Test about`.page`${testcafeconfig.url}`;

const menu = new Menu();

test.meta("testID", "about-001").meta({ mode: "public" })(
  "Core: About page is displayed with all links",
  async (t) => {
    await menu.openPage("about");
    await t.expect(Selector('a[href="https://www.photoprism.app/"]').visible).ok();
  }
);

test.meta("testID", "about-002").meta({ type: "short", mode: "public" })(
  "Core: License page is displayed with all links",
  async (t) => {
    await menu.openPage("license");
    await t
      .expect(Selector("h3").withText("GNU AFFERO GENERAL PUBLIC LICENSE").visible)
      .ok()
      .expect(Selector('a[href="https://www.gnu.org/licenses/agpl-3.0.en.html"]').visible)
      .ok();
  }
);
