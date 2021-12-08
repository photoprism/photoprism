import { Selector } from "testcafe";
import testcafeconfig from "./testcafeconfig";
import Page from "./page-model";

fixture`Test states`.page`${testcafeconfig.url}`;

const page = new Page();

test.meta("testID", "states-001")("Update state", async (t) => {
  await page.openNav();
  await t.click(Selector(".nav-places + div")).click(Selector(".nav-states"));
  await page.search("Canada");
  const AlbumUid = await Selector("a.is-album").nth(0).getAttribute("data-uid");
  await t
    .expect(Selector("button.action-title-edit").nth(0).innerText)
    .contains("British Columbia")
    .click(Selector(".action-title-edit").nth(0))
    .expect(Selector(".input-title input").value)
    .eql("British Columbia")
    .expect(Selector(".input-location input").value)
    .eql("Canada")
    .typeText(Selector(".input-title input"), "Wonderland", { replace: true })
    .typeText(Selector(".input-location input"), "Earth", { replace: true })
    .expect(Selector(".input-description textarea").value)
    .eql("")
    .expect(Selector(".input-category input").value)
    .eql("")
    .typeText(Selector(".input-description textarea"), "We love earth")
    .typeText(Selector(".input-category input"), "Mountains")
    .pressKey("enter")
    .click(".action-confirm")
    .expect(Selector("button.action-title-edit").nth(0).innerText)
    .contains("Wonderland")
    .expect(Selector('div[title="Description"]').nth(0).innerText)
    .contains("We love earth")
    .expect(Selector("div.caption").nth(1).innerText)
    .contains("Mountains")
    .expect(Selector("div.caption").nth(2).innerText)
    .contains("Earth")
    .click(Selector("a.is-album").nth(0));
  await t
    .expect(Selector(".v-card__text").nth(0).innerText)
    .contains("We love earth")
    .expect(Selector("div").withText("Wonderland").exists)
    .ok();
  await page.openNav();
  await t.click(Selector(".nav-states"));
  if (t.browser.platform === "mobile") {
    await page.search("category:Mountains");
  } else {
    await t
      .click(Selector(".input-category"))
      .click(Selector('div[role="listitem"]').withText("Mountains"));
  }
  await t.expect(Selector("button.action-title-edit").nth(0).innerText).contains("Wonderland");
  await t.click(Selector("a.is-album").withAttribute("data-uid", AlbumUid));
  await t
    .click(Selector(".action-edit"))
    .expect(Selector(".input-description textarea").value)
    .eql("We love earth")
    .expect(Selector(".input-category input").value)
    .eql("Mountains")
    .expect(Selector(".input-location input").value)
    .eql("Earth")
    .typeText(Selector(".input-title input"), "British Columbia / Canada", { replace: true })
    .click(Selector(".input-category input"))
    .pressKey("ctrl+a delete")
    .pressKey("enter")
    .click(Selector(".input-description textarea"))
    .pressKey("ctrl+a delete")
    .pressKey("enter")
    .typeText(Selector(".input-location input"), "Canada", { replace: true })
    .click(".action-confirm");
  await page.openNav();
  await t.click(Selector(".nav-states"));
  await page.search("Canada");
  await t
    .expect(Selector("button.action-title-edit").nth(0).innerText)
    .contains("British Columbia / Canada")
    .expect(Selector('div[title="Description"]').innerText)
    .notContains("We love earth")
    .expect(Selector("div.caption").nth(0).innerText)
    .notContains("Earth");
});

test.meta("testID", "states-002")("Download states", async (t) => {
  await page.openNav();
  await t.click(Selector(".nav-places + div")).click(Selector(".nav-states"));
  await page.checkButtonVisibility("download", true, true);
});

//TODO test that sharing link works as expected
test.meta("testID", "states-003")("Create, Edit, delete sharing link", async (t) => {
  await page.testCreateEditDeleteSharingLink("states");

  /* await page.openNav();
  await t
    .click(Selector(".nav-places + div"))
    .click(Selector(".nav-states"));
  const FirstAlbum = await Selector("a.is-album").nth(0).getAttribute("data-uid");
  await page.selectFromUID(FirstAlbum);
  const clipboardCount = await Selector("span.count-clipboard");
  await t
    .expect(clipboardCount.textContent)
    .eql("1")
    .click(Selector("button.action-menu"))
    .click(Selector("button.action-share"))
    .click(Selector("div.v-expansion-panel__header__icon").nth(0));
  const InitialUrl = await Selector(".action-url").innerText;
  const InitialSecret = await Selector(".input-secret input").value;
  const InitialExpire = await Selector("div.v-select__selections").innerText;
  await t
    .expect(InitialUrl)
    .notContains("secretfortesting")
    .expect(InitialExpire)
    .contains("Never")
    .typeText(Selector(".input-secret input"), "secretForTesting", { replace: true })
    .click(Selector(".input-expires input"))
    .click(Selector("div").withText("After 1 day").parent('div[role="listitem"]'))
    .click(Selector("button.action-save"))
    .click(Selector("button.action-close"));
  await page.clearSelection();
  await t
    .click(Selector("a.is-album").withAttribute("data-uid", FirstAlbum))
    .click(Selector("button.action-share"))
    .click(Selector("div.v-expansion-panel__header__icon").nth(0));
  const UrlAfterChange = await Selector(".action-url").innerText;
  const ExpireAfterChange = await Selector("div.v-select__selections").innerText;
  await t
    .expect(UrlAfterChange)
    .contains("secretfortesting")
    .expect(ExpireAfterChange)
    .contains("After 1 day")
    .typeText(Selector(".input-secret input"), InitialSecret, { replace: true })
    .click(Selector(".input-expires input"))
    .click(Selector("div").withText("Never").parent('div[role="listitem"]'))
    .click(Selector("button.action-save"))
    .click(Selector("div.v-expansion-panel__header__icon"));
  const LinkCount = await Selector(".action-url").count;
  await t.click(".action-add-link");
  const LinkCountAfterAdd = await Selector(".action-url").count;
  await t
    .expect(LinkCountAfterAdd)
    .eql(LinkCount + 1)
    .click(Selector("div.v-expansion-panel__header__icon"))
    .click(Selector(".action-delete"));
  const LinkCountAfterDelete = await Selector(".action-url").count;
  await t
    .expect(LinkCountAfterDelete)
    .eql(LinkCountAfterAdd - 1)
    .click(Selector("button.action-close"));
  await page.openNav();
  await t
    .click(".nav-states")
    .click("a.uid-" + FirstAlbum + " .action-share")
    .click(Selector("div.v-expansion-panel__header__icon"))
    .click(Selector(".action-delete"));*/
});

test.meta("testID", "states-004")("Create/delete album during add to album", async (t) => {
  await page.openNav();
  await t.click(Selector(".nav-albums"));
  const countAlbums = await Selector("a.is-album").count;
  await page.openNav();
  await t.click(Selector(".nav-places + div")).click(Selector(".nav-states"));
  await page.search("Canada");
  const FirstMoment = await Selector("a.is-album").nth(0).getAttribute("data-uid");
  await t.click(Selector("a.is-album").withAttribute("data-uid", FirstMoment));
  const PhotoCountInMoment = await Selector("div.is-photo").count;
  const FirstPhoto = await Selector("div.is-photo.type-image").nth(0).getAttribute("data-uid");
  const SecondPhoto = await Selector("div.is-photo.type-image").nth(1).getAttribute("data-uid");
  await page.openNav();
  await t.click(Selector(".nav-states"));
  await page.selectFromUID(FirstMoment);
  await page.addSelectedToAlbum("NotYetExistingAlbumForMoment", "clone");
  await page.openNav();
  await t.click(Selector(".nav-albums"));
  const countAlbumsAfterCreation = await Selector("a.is-album").count;
  await t.expect(countAlbumsAfterCreation).eql(countAlbums + 1);
  await page.search("NotYetExistingAlbumForMoment");
  const AlbumUid = await Selector("a.is-album").nth(0).getAttribute("data-uid");
  await t.click(Selector("a.is-album").withAttribute("data-uid", AlbumUid));
  const PhotoCountInAlbum = await Selector("div.is-photo").count;
  await t
    .expect(PhotoCountInAlbum)
    .eql(PhotoCountInMoment)
    .expect(Selector("div").withAttribute("data-uid", FirstPhoto).exists, { timeout: 5000 })
    .ok()
    .expect(Selector("div").withAttribute("data-uid", SecondPhoto).exists, { timeout: 5000 })
    .ok();
  await page.openNav();
  await t.click(Selector(".nav-albums"));
  await page.selectFromUID(AlbumUid);
  await page.deleteSelected();
  await page.openNav();
  await t.click(Selector(".nav-albums"));
  const countAlbumsAfterDelete = await Selector("a.is-album").count;
  await t.expect(countAlbumsAfterDelete).eql(countAlbums);
  await t
    .click(Selector(".nav-states"))
    .click(Selector("a.is-album").withAttribute("data-uid", FirstMoment))
    .expect(Selector("div").withAttribute("data-uid", FirstPhoto).exists, { timeout: 5000 })
    .ok()
    .expect(Selector("div").withAttribute("data-uid", SecondPhoto).exists, { timeout: 5000 })
    .ok();
});

test.meta("testID", "states-005")("Delete states button visible", async (t) => {
  await page.openNav();
  await t.click(Selector(".nav-places + div")).click(Selector(".nav-states"));
  await page.checkButtonVisibility("delete", true, false);
});
