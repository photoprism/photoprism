import { Selector } from "testcafe";
import testcafeconfig from "../testcafeconfig";
import Menu from "../../page-model/menu";

fixture`Test about`.page`${testcafeconfig.url}`;

const menu = new Menu();

test.meta("testID", "about-001")("About page is displayed with all links", async (t) => {
  await menu.openPage("about");
  await t
    .expect(Selector('a[href="https://photoprism.app/"]').visible)
    .ok()
    .expect(Selector('a[href="https://photoprism.app/membership"]').visible)
    .ok();
});

test.meta("testID", "about-002")("License page is displayed with all links", async (t) => {
  await menu.openPage("license");
  await t
    .expect(Selector("h3").withText("GNU AFFERO GENERAL PUBLIC LICENSE").visible)
    .ok()
    .expect(Selector('a[href="https://www.gnu.org/licenses/agpl-3.0.en.html"]').visible)
    .ok();
});
