import { Selector } from "testcafe";
import testcafeconfig from "../acceptance/testcafeconfig";
import Page from "../acceptance/page-model";

fixture`Test admin role`.page`${testcafeconfig.url}`;

const page = new Page();
test.meta("testID", "authentication-000")(
  "Time to start instance (will be marked as unstable)",
  async (t) => {
    await t.wait(5000);
  }
);

test.meta("testID", "admin-role-001")("Access to settings", async (t) => {
  await page.login("admin", "photoprism");
  await page.openNav();
  await t.expect(Selector(".nav-settings").visible).ok();
  await t
    .click(Selector(".nav-settings"))
    .expect(Selector(".input-language input", { timeout: 8000 }).visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk()
    .click(Selector("#tab-settings-library"))
    .expect(Selector("form.p-form-settings").visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk()
    .click(Selector("#tab-settings-advanced"))
    .expect(Selector("label").withText("Read-Only Mode").visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk()
    .click(Selector("#tab-settings-sync"))
    .expect(Selector("div.p-accounts-list").visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk();
});

test.meta("testID", "admin-role-002")("Access to archive", async (t) => {
  await page.login("admin", "photoprism");
  const PhotoCountBrowse = await Selector("div.is-photo", { timeout: 5000 }).count;
  await page.openNav();
  await t.click(Selector(".nav-browse + div")).expect(Selector(".nav-archive").visible).ok();
  await t
    .click(Selector(".nav-archive"))
    .expect(Selector("div.is-photo").withAttribute("data-uid", "pqnahct2mvee8sr4").visible)
    .ok();
  const PhotoCountArchive = await Selector("div.is-photo", { timeout: 5000 }).count;
  await t.expect(PhotoCountBrowse).gte(PhotoCountArchive);
});

test.meta("testID", "admin-role-003")("Access to review", async (t) => {
  await page.login("admin", "photoprism");
  const PhotoCountBrowse = await Selector("div.is-photo", { timeout: 5000 }).count;
  await page.openNav();
  await t.click(Selector(".nav-browse + div")).expect(Selector(".nav-review").visible).ok();
  await t
    .click(Selector(".nav-review"))
    .expect(Selector("div.is-photo").withAttribute("data-uid", "pqzuein2pdcg1kc7").visible)
    .ok();
  const PhotoCountReview = await Selector("div.is-photo", { timeout: 5000 }).count;
  await t.expect(PhotoCountBrowse).gte(PhotoCountReview);
});

test.meta("testID", "admin-role-004")("Access to private", async (t) => {
  await page.login("admin", "photoprism");
  const PhotoCountBrowse = await Selector("div.is-photo", { timeout: 5000 }).count;
  await page.openNav();
  await t.expect(Selector(".nav-private").visible).ok();
  await t
    .click(Selector(".nav-private"))
    .expect(Selector("div.is-photo").withAttribute("data-uid", "pqmxlquf9tbc8mk2").visible)
    .ok();
  const PhotoCountPrivate = await Selector("div.is-photo", { timeout: 5000 }).count;
  await t.expect(PhotoCountBrowse).gte(PhotoCountPrivate);
});

test.meta("testID", "admin-role-005")("Access to library", async (t) => {
  await page.login("admin", "photoprism");
  const PhotoCountBrowse = await Selector("div.is-photo", { timeout: 5000 }).count;
  await page.openNav();
  await t.expect(Selector(".nav-library").visible).ok();
  await t
    .click(Selector(".nav-library"))
    .expect(Selector(".input-index-folder input").visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk()
    .click(Selector("#tab-library-import"))
    .expect(Selector(".input-import-folder input").visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk()
    .click(Selector("#tab-library-logs"))
    .expect(Selector("p.p-log-debug").visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk()
    .click(Selector(".nav-library + div"))
    .expect(Selector(".nav-originals").visible)
    .ok()
    .click(Selector(".nav-originals"))
    .expect(Selector("div.p-page-files").visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk()
    .expect(Selector(".nav-hidden").visible)
    .ok()
    .click(Selector(".nav-hidden"));
  const PhotoCountHidden = await Selector("div.is-photo", { timeout: 5000 }).count;
  await t
    .expect(PhotoCountBrowse)
    .gte(PhotoCountHidden)
    .expect(Selector(".nav-errors").visible)
    .ok()
    .click(Selector(".nav-errors"))
    .expect(Selector("div.p-page-errors").visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk();
});

test.meta("testID", "admin-role-006")("private/archived photos in search results", async (t) => {
  await page.login("admin", "photoprism");
  const PhotoCountBrowse = await Selector("div.is-photo", { timeout: 5000 }).count;
  await page.search("private:true");
  const PhotoCountPrivate = await Selector("div.is-photo", { timeout: 5000 }).count;
  await t.expect(PhotoCountPrivate).eql(2);
  await t
    .expect(Selector("div.is-photo").withAttribute("data-uid", "pqmxlquf9tbc8mk2").visible)
    .ok();
  await page.search("archived:true");
  const PhotoCountArchive = await Selector("div.is-photo", { timeout: 5000 }).count;
  await t.expect(PhotoCountArchive).eql(3);
  await t
    .expect(Selector("div.is-photo").withAttribute("data-uid", "pqnahct2mvee8sr4").visible)
    .ok();
  await page.search("quality:0");
  const PhotoCountReview = await Selector("div.is-photo", { timeout: 5000 }).count;
  await t.expect(PhotoCountReview).gte(PhotoCountBrowse);
  await t
    .expect(Selector("div.is-photo").withAttribute("data-uid", "pqzuein2pdcg1kc7").visible)
    .ok();
});

test.meta("testID", "admin-role-007")("Upload functionality", async (t) => {
  await page.login("admin", "photoprism");
  await t
    .expect(Selector("button.action-upload").visible)
    .ok()
    .click(Selector(".nav-albums"))
    .expect(Selector("a.is-album").visible)
    .ok()
    .expect(Selector("button.action-upload").visible)
    .ok()
    .click(Selector("a.is-album").nth(0))
    .expect(Selector("div.is-photo").visible)
    .ok()
    .expect(Selector("button.action-upload").visible)
    .ok()
    .click(Selector(".nav-video"))
    .expect(Selector("button.action-upload").visible)
    .ok()
    .click(Selector(".nav-favorites"))
    .expect(Selector("button.action-upload").visible)
    .ok()
    .click(Selector(".nav-moments"))
    .expect(Selector("a.is-album").visible)
    .ok()
    .expect(Selector("button.action-upload").visible)
    .ok()
    .click(Selector("a.is-album").nth(0))
    .expect(Selector("div.is-photo").visible)
    .ok()
    .expect(Selector("button.action-upload").visible)
    .ok()
    .click(Selector(".nav-calendar"))
    .expect(Selector("a.is-album").visible)
    .ok()
    .expect(Selector("button.action-upload").visible)
    .ok()
    .click(Selector("a.is-album").nth(0))
    .expect(Selector("div.is-photo").visible)
    .ok()
    .expect(Selector("button.action-upload").visible)
    .ok();
  await page.openNav();
  await t
    .click(Selector(".nav-places + div"))
    .click(Selector(".nav-states"))
    .expect(Selector("a.is-album").visible)
    .ok()
    .expect(Selector("button.action-upload").visible)
    .ok()
    .click(Selector("a.is-album").nth(0))
    .expect(Selector("div.is-photo").visible)
    .ok()
    .expect(Selector("button.action-upload").visible)
    .ok()
    .click(Selector(".nav-folders"))
    .expect(Selector("a.is-album").visible)
    .ok()
    .expect(Selector("button.action-upload").visible)
    .ok()
    .click(Selector("a.is-album").nth(0))
    .expect(Selector("div.is-photo").visible)
    .ok()
    .expect(Selector("button.action-upload").visible)
    .ok();
});
