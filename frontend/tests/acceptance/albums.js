import { Selector } from "testcafe";
import testcafeconfig from "./testcafeconfig";
import Page from "./page-model";

fixture`Test albums`.page`${testcafeconfig.url}`;

const page = new Page();

test.meta("testID", "authentication-000")(
  "Time to start instance (will be marked as unstable)",
  async (t) => {
    await t.wait(5000);
  }
);

test.meta("testID", "albums-001")("Create/delete album on /albums", async (t) => {
  await page.openNav();
  await t.click(Selector(".nav-albums"));
  const countAlbums = await Selector("a.is-album").count;
  await t.click(Selector("button.action-add"));
  const countAlbumsAfterCreate = await Selector("a.is-album").count;
  const NewAlbum = await Selector("a.is-album").nth(0).getAttribute("data-uid");
  await t.expect(countAlbumsAfterCreate).eql(countAlbums + 1);
  await page.selectFromUID(NewAlbum);
  await page.deleteSelected();
  const countAlbumsAfterDelete = await Selector("a.is-album").count;
  await t.expect(countAlbumsAfterDelete).eql(countAlbumsAfterCreate - 1);
});

test.meta("testID", "albums-002")("Update album", async (t) => {
  await page.openNav();
  await t.click(Selector(".nav-albums"));
  await page.search("Holiday");
  const AlbumUid = await Selector("a.is-album", { timeout: 55000 }).nth(0).getAttribute("data-uid");
  await t
    .expect(Selector("button.action-title-edit").nth(0).innerText)
    .contains("Holiday")
    .click(Selector(".action-title-edit").nth(0))
    .typeText(Selector(".input-title input"), "Animals", { replace: true })
    .expect(Selector(".input-description textarea").value)
    .eql("")
    .expect(Selector(".input-category input").value)
    .eql("")
    .typeText(Selector(".input-description textarea"), "All my animals")
    .typeText(Selector(".input-category input"), "Pets")
    .pressKey("enter")
    .click(".action-confirm")
    .click(Selector("a.is-album").nth(0));
  const PhotoCount = await Selector("div.is-photo").count;
  await t
    .expect(Selector(".v-card__text").nth(0).innerText)
    .contains("All my animals")
    .expect(Selector("div").withText("Animals").exists)
    .ok();
  await page.openNav();
  await t.click(Selector(".nav-browse"));
  await page.search("photo:true");
  const FirstPhotoUid = await Selector("div.is-photo.type-image").nth(0).getAttribute("data-uid");
  const SecondPhotoUid = await Selector("div.is-photo.type-image").nth(1).getAttribute("data-uid");
  await page.selectPhotoFromUID(SecondPhotoUid);
  await page.selectFromUIDInFullscreen(FirstPhotoUid);
  await page.addSelectedToAlbum("Animals", "album");
  await page.openNav();
  await t.click(Selector(".nav-albums"));
  if (t.browser.platform === "mobile") {
    await page.search("category:Family");
  } else {
    await t
      .click(Selector(".input-category"))
      .click(Selector('div[role="listitem"]').withText("Family"));
  }
  await t.expect(Selector("button.action-title-edit").nth(0).innerText).contains("Christmas");
  await page.openNav();
  await t.click(Selector(".nav-albums")).click(".action-reload");
  if (t.browser.platform === "mobile") {
  } else {
    await t
      .click(Selector(".input-category"))
      .click(Selector('div[role="listitem"]').withText("All Categories"), { timeout: 55000 });
  }
  await t.click(Selector("a.is-album").withAttribute("data-uid", AlbumUid));
  const PhotoCountAfterAdd = await Selector("div.is-photo").count;
  await t.expect(PhotoCountAfterAdd).eql(PhotoCount + 2);
  await page.selectPhotoFromUID(FirstPhotoUid);
  await page.selectPhotoFromUID(SecondPhotoUid);
  await page.removeSelected();
  const PhotoCountAfterDelete = await Selector("div.is-photo").count;
  await t
    .expect(PhotoCountAfterDelete)
    .eql(PhotoCountAfterAdd - 2)
    .click(Selector(".action-edit"))
    .typeText(Selector(".input-title input"), "Holiday", { replace: true })
    .expect(Selector(".input-description textarea").value)
    .eql("All my animals")
    .expect(Selector(".input-category input").value)
    .eql("Pets")
    .click(Selector(".input-description textarea"))
    .pressKey("ctrl+a delete")
    .pressKey("enter")
    .click(Selector(".input-category input"))
    .pressKey("ctrl+a delete")
    .pressKey("enter")
    .click(".action-confirm");
  await page.openNav();
  await t
    .click(Selector(".nav-albums"))
    .expect(Selector("div").withText("Holiday").visible)
    .ok()
    .expect(Selector("div").withText("Animals").exists)
    .notOk();
});

test.meta("testID", "albums-003")("Download album", async (t) => {
  await page.openNav();
  await t.click(Selector(".nav-albums"));
  await page.checkButtonVisibility("download", true, true);
});

test.meta("testID", "albums-004")("View folders", async (t) => {
  await page.openNav();
  await t
    .click(Selector(".nav-folders"))
    .expect(Selector("a").withText("BotanicalGarden").visible)
    .ok()
    .expect(Selector("a").withText("Kanada").visible)
    .ok()
    .expect(Selector("a").withText("KorsikaAdventure").visible)
    .ok();
});

test.meta("testID", "albums-005")("View calendar", async (t) => {
  await page.openNav();
  await t
    .click(Selector(".nav-calendar"))
    .expect(Selector("a").withText("May 2019").visible)
    .ok()
    .expect(Selector("a").withText("October 2019").visible)
    .ok();
});

//TODO test that sharing link works as expected
test.meta("testID", "albums-006")("Create, Edit, delete sharing link", async (t) => {
  await page.testCreateEditDeleteSharingLink("albums");
});

test.meta("testID", "albums-007")("Create/delete album during add to album", async (t) => {
  await page.openNav();
  await t.click(Selector(".nav-albums"));
  const countAlbums = await Selector("a.is-album").count;
  await page.openNav();
  await t.click(Selector(".nav-browse"));
  await page.search("photo:true");
  const FirstPhotoUid = await Selector("div.is-photo.type-image").nth(0).getAttribute("data-uid");
  const SecondPhotoUid = await Selector("div.is-photo.type-image").nth(1).getAttribute("data-uid");
  await page.selectPhotoFromUID(SecondPhotoUid);
  await page.selectFromUIDInFullscreen(FirstPhotoUid);
  await page.addSelectedToAlbum("NotYetExistingAlbum", "album");
  await page.openNav();
  await t.click(Selector(".nav-albums"));
  const countAlbumsAfterCreation = await Selector("a.is-album").count;
  await t.expect(countAlbumsAfterCreation).eql(countAlbums + 1);
  await page.search("NotYetExistingAlbum");
  const AlbumUid = await Selector("a.is-album").nth(0).getAttribute("data-uid");
  await page.selectFromUID(AlbumUid);
  await page.deleteSelected();
  await page.openNav();
  await t.click(Selector(".nav-albums"));
  const countAlbumsAfterDelete = await Selector("a.is-album").count;
  await t.expect(countAlbumsAfterDelete).eql(countAlbums);
});

test.meta("testID", "albums-008")("Test album autocomplete", async (t) => {
  await page.search("photo:true");
  const FirstPhotoUid = await Selector("div.is-photo.type-image").nth(0).getAttribute("data-uid");
  await page.selectPhotoFromUID(FirstPhotoUid);
  await t
    .click(Selector("button.action-menu"))
    .click(Selector("button.action-album"))
    .click(Selector(".input-album input"))
    .expect(Selector("div.v-list__tile__title").withText("Holiday").visible)
    .ok()
    .expect(Selector("div.v-list__tile__title").withText("Christmas").visible)
    .ok()
    .typeText(Selector(".input-album input"), "C", { replace: true })
    .expect(Selector("div.v-list__tile__title").withText("Christmas").visible)
    .ok()
    .expect(Selector("div.v-list__tile__title").withText("C").visible)
    .ok()
    .expect(Selector("div.v-list__tile__title").withText("Holiday").visible)
    .notOk();
});
