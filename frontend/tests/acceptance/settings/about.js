import { Selector } from "testcafe";
import testcafeconfig from "../testcafeconfig";
import Page from "../page-model";

fixture`Test about`.page`${testcafeconfig.url}`;

const page = new Page();
test.meta("testID", "about-001")("About page is displayed with all links", async (t) => {
  await page.openNav();
  await t
    .click(Selector(".nav-settings + div"))
    .click(Selector(".nav-about"))
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
  await page.openNav();
  await t
    .click(Selector(".nav-settings + div"))
    .click(Selector(".nav-license"))
    .expect(Selector("h3").withText("GNU AFFERO GENERAL PUBLIC LICENSE").visible)
    .ok()
    .expect(Selector('a[href="https://www.gnu.org/licenses/agpl-3.0.en.html"]').visible)
    .ok();
});
