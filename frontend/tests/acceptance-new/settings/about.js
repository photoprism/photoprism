import { Selector } from "testcafe";
import testcafeconfig from "../testcafeconfig";
import Menu from "../../page-model/menu";

fixture`Test about`.page`${testcafeconfig.url}`;

const menu = new Menu();

test.meta("testID", "about-001")("About page is displayed with all links", async (t) => {
  await menu.openPage("about");
  await t
    .expect(Selector("h2").withText("Trademarks").visible)
    .ok()
    .expect(Selector('a[href="https://photoprism.app/"]').visible)
    .ok()
    .expect(Selector('a[href="https://www.patreon.com/photoprism"]').visible)
    .ok()
    .expect(Selector('a[href="https://github.com/photoprism/photoprism/projects/5"]').visible)
    .ok()
    .expect(Selector('a[href="https://docs.photoprism.app/"]').visible)
    .ok()
    .expect(Selector('a[href="/about/license"]').visible)
    .ok()
    .expect(Selector('a[href="https://gitter.im/browseyourlife/community"]').visible)
    .ok()
    .expect(Selector('a[href="https://twitter.com/photoprism_app"]').visible)
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
