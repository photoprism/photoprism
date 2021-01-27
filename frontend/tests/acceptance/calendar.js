import { Selector } from "testcafe";
import testcafeconfig from "./testcafeconfig";
import Page from "./page-model";

fixture`Test calendar`.page`${testcafeconfig.url}`;

const page = new Page();

test.meta("testID", "calendar-001")("Update calendar", async (t) => {
  await page.openNav();
  await t.click(Selector(".nav-calendar"));
  await page.search("March 2014");
  const AlbumUid = await Selector("a.is-album").nth(0).getAttribute("data-uid");
  await t
    .expect(Selector("button.action-title-edit").nth(0).innerText)
    .contains("March 2014")
    .click(Selector(".action-title-edit").nth(0))
    .typeText(Selector(".input-location input"), "Snow", { replace: true })
    .expect(Selector(".input-description textarea").value)
    .eql("")
    .expect(Selector(".input-category input").value)
    .eql("")
    .typeText(Selector(".input-description textarea"), "We went to ski")
    .typeText(Selector(".input-category input"), "Mountains")
    .pressKey("enter")
    .click(".action-confirm")
    .expect(Selector("button.action-title-edit").nth(0).innerText)
    .contains("March 2014")
    .expect(Selector('div[title="Description"]').nth(0).innerText)
    .contains("We went to ski")
    .expect(Selector("div.caption").nth(1).innerText)
    .contains("Snow")
    .click(Selector("a.is-album").nth(0));
  await t
    .expect(Selector(".v-card__text").nth(0).innerText)
    .contains("We went to ski")
    .expect(Selector("div").withText("March 2014").exists)
    .ok();
  await page.openNav();
  await t.click(Selector(".nav-calendar"));
  if (t.browser.platform === "mobile") {
    await page.search("category:Mountains");
  } else {
    await t
      .click(Selector(".input-category i"))
      .click(Selector('div[role="listitem"]').withText("Mountains"));
  }
  await t.expect(Selector("button.action-title-edit").nth(0).innerText).contains("March 2014");
  await t.click(Selector("a.is-album").withAttribute("data-uid", AlbumUid));
  await t
    .click(Selector(".action-edit"))
    .expect(Selector(".input-description textarea").value)
    .eql("We went to ski")
    .expect(Selector(".input-category input").value)
    .eql("Mountains")
    .expect(Selector(".input-location input").value)
    .eql("Snow")
    .click(Selector(".input-category input"))
    .pressKey("ctrl+a delete")
    .pressKey("enter")
    .click(Selector(".input-description textarea"))
    .pressKey("ctrl+a delete")
    .pressKey("enter")
    .click(Selector(".input-location input"))
    .pressKey("ctrl+a delete")
    .pressKey("enter")
    .click(".action-confirm");
  await page.openNav();
  await t.click(Selector(".nav-calendar"));
  await page.search("March 2014");
  await t
    .expect(Selector('div[title="Description"]').innerText)
    .notContains("We went to ski")
    .expect(Selector("div.caption").nth(0).innerText)
    .notContains("Snow");
});

test.meta("testID", "calendar-002")("Download calendar", async (t) => {
  await page.openNav();
  await t.click(Selector(".nav-calendar"));
  const FirstAlbum = await Selector("a.is-album").nth(0).getAttribute("data-uid");
  await page.selectFromUID(FirstAlbum);
  const clipboardCount = await Selector("span.count-clipboard");
  await t
    .expect(clipboardCount.textContent)
    .eql("1")
    .click(Selector("button.action-menu"))
    .expect(Selector("button.action-download").visible)
    .ok();
  if (t.browser.platform !== "mobile") {
    await page.clearSelection();
    await t
      .click(Selector("a.is-album").nth(0))
      .expect(Selector("button.action-download").visible)
      .ok();
  }
});

//TODO test that sharing link works as expected
test.meta("testID", "calendar-003")("Create, Edit, delete sharing link", async (t) => {
  await page.openNav();
  await t.click(Selector(".nav-calendar"));
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
    .click(".nav-calendar")
    .click("a.uid-" + FirstAlbum + " .action-share")
    .click(Selector("div.v-expansion-panel__header__icon"))
    .click(Selector(".action-delete"));
});

test.meta("testID", "calendar-004")("Create/delete album during add to album", async (t) => {
  await page.openNav();
  await t.click(Selector(".nav-albums"));
  const countAlbums = await Selector("a.is-album").count;
  await page.openNav();
  await t.click(Selector(".nav-calendar"));
  const ThirdCalendar = await Selector("a.is-album").nth(2).getAttribute("data-uid");
  await t.click(Selector("a.is-album").withAttribute("data-uid", ThirdCalendar));
  const PhotoCountInCalendar = await Selector("div.is-photo").count;
  const FirstPhoto = await Selector("div.is-photo.type-image").nth(0).getAttribute("data-uid");
  const SecondPhoto = await Selector("div.is-photo.type-image").nth(1).getAttribute("data-uid");
  await page.openNav();
  await t.click(Selector(".nav-calendar"));
  await page.selectFromUID(ThirdCalendar);
  await page.addSelectedToAlbum("NotYetExistingAlbumForCalendar", "clone");
  await page.openNav();
  await t.click(Selector(".nav-albums"));
  const countAlbumsAfterCreation = await Selector("a.is-album").count;
  await t.expect(countAlbumsAfterCreation).eql(countAlbums + 1);
  await page.search("NotYetExistingAlbumForCalendar");
  const AlbumUid = await Selector("a.is-album").nth(0).getAttribute("data-uid");
  await t.click(Selector("a.is-album").withAttribute("data-uid", AlbumUid));
  const PhotoCountInAlbum = await Selector("div.is-photo").count;
  await t
    .expect(PhotoCountInAlbum)
    .eql(PhotoCountInCalendar)
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
    .click(Selector(".nav-calendar"))
    .click(Selector("a.is-album").withAttribute("data-uid", ThirdCalendar))
    .expect(Selector("div").withAttribute("data-uid", FirstPhoto).exists, { timeout: 5000 })
    .ok()
    .expect(Selector("div").withAttribute("data-uid", SecondPhoto).exists, { timeout: 5000 })
    .ok();
});

test.meta("testID", "calendar-005")("No delete calendar functionality", async (t) => {
  await page.openNav();
  await t.click(Selector(".nav-calendar"));
  const FirstAlbum = await Selector("a.is-album").nth(0).getAttribute("data-uid");
  await page.selectFromUID(FirstAlbum);
  const clipboardCount = await Selector("span.count-clipboard");
  await t
    .expect(clipboardCount.textContent)
    .eql("1")
    .click(Selector("button.action-menu"))
    .expect(Selector("button.action-delete").exists)
    .notOk();
  if (t.browser.platform !== "mobile") {
    await page.clearSelection();
    await t
      .click(Selector("a.is-album").nth(0))
      .expect(Selector("button.action-delete").exists)
      .notOk();
  }
});
