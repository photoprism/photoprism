import { Selector } from "testcafe";
import { Role } from "testcafe";
import testcafeconfig from "../acceptance/testcafeconfig";
import Page from "../acceptance/page-model";

fixture`Test link sharing`.page`${testcafeconfig.url}`;

const page = new Page();

test.meta("testID", "authentication-000")(
  "Time to start instance (will be marked as unstable)",
  async (t) => {
    await t.wait(5000);
  }
);

test.meta("testID", "sharing-001")("View shared albums", async (t) => {
  await page.login("admin", "photoprism");
  await page.openNav();
  await t.click(Selector(".nav-albums"));
  const FirstAlbum = await Selector("a.is-album").nth(0).getAttribute("data-uid");
  await page.selectFromUID(FirstAlbum);
  const clipboardCount = await Selector("span.count-clipboard");
  await t
    .expect(clipboardCount.textContent)
    .eql("1")
    .click(Selector("button.action-menu"))
    .click(Selector("button.action-share"))
    .click(Selector("div.v-expansion-panel__header__icon").nth(0));
  await t
    .typeText(Selector(".input-secret input"), "secretForTesting", { replace: true })
    .click(Selector(".input-expires input"))
    .click(Selector("div").withText("After 1 day").parent('div[role="listitem"]'))
    .click(Selector("button.action-save"));
  const Url = await Selector("div.input-url input").value;
  const Expire = await Selector("div.v-select__selections").innerText;
  await t.expect(Url).contains("secretfortesting").expect(Expire).contains("After 1 day");
  let url = Url.replace("2342", "2343");
  await t.click(Selector("button.action-close"));
  await page.clearSelection();
  await t.click(Selector(".nav-folders"));
  const FirstFolder = await Selector("a.is-album").nth(0).getAttribute("data-uid");
  await page.selectFromUID(FirstFolder);
  await t
    .click(Selector("button.action-menu"))
    .click(Selector("button.action-share"))
    .click(Selector("div.v-expansion-panel__header__icon").nth(0));
  await t
    .typeText(Selector(".input-secret input"), "secretForTesting", { replace: true })
    .click(Selector(".input-expires input"))
    .click(Selector("div").withText("After 1 day").parent('div[role="listitem"]'))
    .click(Selector("button.action-save"))
    .click(Selector("button.action-close"));
  await page.clearSelection();
  await t.navigateTo(url);
  await t
    .expect(Selector("div.v-toolbar__title").withText("Christmas").visible)
    .ok()
    .click(Selector("button").withText("@photoprism_app"))
    .expect(Selector("div.v-toolbar__title").withText("Albums").visible)
    .ok();
  const countAlbums = await Selector("a.is-album").count;
  await t.expect(countAlbums).gte(40).useRole(Role.anonymous());
  await t.navigateTo(url);
  await t
    .expect(Selector("div.v-toolbar__title").withText("Christmas").visible)
    .ok()
    .click(Selector("button").withText("@photoprism_app"))
    .expect(Selector("div.v-toolbar__title").withText("Albums").visible)
    .ok();
  const countAlbumsAnonymous = await Selector("a.is-album").count;
  await t.expect(countAlbumsAnonymous).eql(2);
  await t.navigateTo("http://localhost:2343/browse");
  await page.login("admin", "photoprism");
  await page.openNav();
  await t
    .click(Selector(".nav-albums"))
    .click(Selector("a.is-album").withAttribute("data-uid", FirstAlbum))
    .click(Selector("button.action-share"))
    .click(Selector("div.v-expansion-panel__header__icon").nth(0))
    .click(Selector(".action-delete"))
    .useRole(Role.anonymous())
    .expect(Selector(".input-name input").visible)
    .ok();
  await t.navigateTo("http://localhost:2343/s/secretfortesting");
  await t.expect(Selector("div.v-toolbar__title").withText("Albums").visible).ok();
  const countAlbumsAnonymousAfterDelete = await Selector("a.is-album").count;
  await t.expect(countAlbumsAnonymousAfterDelete).eql(1);
  await t.navigateTo("http://localhost:2343/browse");
  await page.login("admin", "photoprism");
  await page.openNav();
  await t
    .click(Selector(".nav-folders"))
    .click(Selector("a.is-album").withAttribute("data-uid", FirstFolder))
    .click(Selector("button.action-share"))
    .click(Selector("div.v-expansion-panel__header__icon").nth(0))
    .click(Selector(".action-delete"))
    .useRole(Role.anonymous())
    .expect(Selector(".input-name input").visible)
    .ok();
  await t.navigateTo("http://localhost:2343/s/secretfortesting");
  await t
    .expect(Selector("div.v-toolbar__title").withText("Christmas").visible)
    .notOk()
    .expect(Selector("div.v-toolbar__title").withText("Albums").visible)
    .notOk()
    .expect(Selector(".input-name input").visible)
    .ok();
});

test.meta("testID", "sharing-002")("Verify anonymous user has limited options", async (t) => {
  await t.navigateTo("http://localhost:2343/s/jxoux5ub1e/british-columbia-canada");
  // check album toolbar
  await t
    .expect(Selector("div.v-toolbar__title").withText("British Columbia").visible)
    .ok()
    .expect(Selector("button.action-edit").visible)
    .notOk()
    .expect(Selector("button.action-share").visible)
    .notOk()
    .expect(Selector("button.action-upload").visible)
    .notOk()
    .expect(Selector("button.action-reload").visible)
    .ok()
    .expect(Selector("button.action-download").visible)
    .ok();
  //check photo context menu
  await page.toggleSelectNthPhoto(0);
  await t
    .click("button.action-menu")
    .expect(Selector("div.v-speed-dial__list button.action-download").visible)
    .ok()
    .expect(Selector("div.v-speed-dial__list button.action-archive").visible)
    .notOk()
    .expect(Selector("div.v-speed-dial__list button.action-album").visible)
    .notOk()
    .expect(Selector("div.v-speed-dial__list button.action-private").visible)
    .notOk()
    .expect(Selector("div.v-speed-dial__list button.action-edit").visible)
    .notOk()
    .expect(Selector("div.v-speed-dial__list button.action-share").visible)
    .notOk();
  await page.clearSelection();
  await t.expect(Selector("button.action-title-edit").visible).notOk();
  //check fullscreen actions
  await t
    .click(Selector('h3[title="Cape / Bowen Island / 2019"]'))
    .expect(Selector("#photo-viewer").visible)
    .ok()
    .expect(Selector("img.pswp__img").visible)
    .ok()
    .expect(Selector("button.action-select").visible)
    .ok()
    .expect(Selector('button[title="Start/Stop Slideshow"]').visible)
    .ok()
    .expect(Selector('button[title="Fullscreen"]').visible)
    .ok()
    .expect(Selector('button[title="Start/Stop Slideshow"]').visible)
    .ok()
    .expect(Selector('button[title="Download"]').visible)
    .ok()
    .expect(Selector('button[title="Like"]').visible)
    .notOk()
    .expect(Selector('button[title="Edit"]').visible)
    .notOk()
    .click(Selector('button[title="Close"]'))
    //check hover like actions card and mosaic
    .expect(Selector("button.input-favorite").visible)
    .notOk()
    //check list view actions
    //hover on mosaic
    //action-menu albums
    .click(Selector("button").withText("@photoprism_app"))
    .expect(Selector("div.v-toolbar__title").withText("Albums").visible)
    .ok();
  //album edit dialog
  const AlbumUid = await Selector("a.is-album", { timeout: 55000 }).nth(0).getAttribute("data-uid");
  await page.selectFromUID(AlbumUid);
  await t
    .click(Selector("button.action-menu"))
    .expect(Selector("div.v-speed-dial__list button.action-download").visible)
    .ok()
    .expect(Selector("div.v-speed-dial__list button.action-delete").visible)
    .notOk()
    .expect(Selector("div.v-speed-dial__list button.action-album").visible)
    .notOk()
    .expect(Selector("div.v-speed-dial__list button.action-edit").visible)
    .notOk()
    .expect(Selector("div.v-speed-dial__list button.action-share").visible)
    .notOk();
  await page.clearSelection();
  await t.expect(Selector("button.action-title-edit").visible).notOk();
  //TODO control + page model
});
