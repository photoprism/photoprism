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
  await t.navigateTo("/settings");
  await t
    .expect(Selector(".input-language input", { timeout: 8000 }).visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk()
    .navigateTo("/settings/library")
    .expect(Selector("form.p-form-settings").visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk()
    .navigateTo("/settings/advanced")
    .expect(Selector("label").withText("Read-Only Mode").visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk()
    .navigateTo("/settings/sync")
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
  await t.navigateTo("/archive");
  await t
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
  await t.navigateTo("/review");
  await t
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
  await t.navigateTo("/private");
  await t
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
  await t.navigateTo("/library");
  await t
    .expect(Selector(".input-index-folder input").visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk()
    .navigateTo("/library/import")
    .expect(Selector(".input-import-folder input").visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk()
    .navigateTo("/library/logs")
    .expect(Selector("p.p-log-message").visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk()
    .click(Selector(".nav-library + div"))
    .expect(Selector(".nav-originals").visible)
    .ok()
    .navigateTo("/library/files")
    .expect(Selector("div.p-page-files").visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk()
    .expect(Selector(".nav-hidden").visible)
    .ok()
    .navigateTo("/library/hidden");
  const PhotoCountHidden = await Selector("div.is-photo", { timeout: 5000 }).count;
  await t
    .expect(PhotoCountBrowse)
    .gte(PhotoCountHidden)
    .expect(Selector(".nav-errors").visible)
    .ok()
    .navigateTo("/library/errors")
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

test.meta("testID", "admin-role-008")(
  "Admin can private, archive, share, add/remove to album",
  async (t) => {
    await page.login("admin", "photoprism");
    const FirstPhoto = await Selector("div.is-photo.type-image").nth(0).getAttribute("data-uid");
    const SecondPhoto = await Selector("div.is-photo").nth(1).getAttribute("data-uid");
    await page.selectPhotoFromUID(FirstPhoto);
    await t
      .click(Selector("button.action-menu"))
      .expect(Selector("button.action-private").visible)
      .ok()
      .expect(Selector("button.action-archive").visible)
      .ok()
      .expect(Selector("button.action-share").visible)
      .ok()
      .expect(Selector("button.action-album").visible)
      .ok();
    await page.clearSelection();
    await page.setFilter("view", "List");
    await t.expect(Selector(`button.input-private`).hasAttribute("disabled")).notOk();
    await t
      .click(Selector(".nav-albums"))
      .click(Selector("a.is-album").nth(0))
      .expect(Selector("button.action-share").visible)
      .ok();
    await page.toggleSelectNthPhoto(0);
    await t
      .click(Selector("button.action-menu"))
      .expect(Selector("button.action-private").visible)
      .ok()
      .expect(Selector("button.action-share").visible)
      .ok()
      .expect(Selector("button.action-album").visible)
      .ok()
      .expect(Selector("button.action-remove").visible)
      .ok();
    await page.clearSelection();
    await t
      .click(Selector('button[title="Toggle View"]'))
      .expect(Selector(`button.input-private`).hasAttribute("disabled"))
      .notOk();
  }
);

test.meta("testID", "admin-role-009")("Admin can approve low quality photos", async (t) => {
  await page.login("admin", "photoprism");
  await page.search('quality:0 name:"photos-013_1"');
  await page.toggleSelectNthPhoto(0);
  await page.editSelected();
  await t.expect(Selector("button.action-approve").visible).ok();
});

test.meta("testID", "admin-role-010")("Edit dialog is not read only for admin", async (t) => {
  await page.login("admin", "photoprism");
  await page.search("faces:new");
  //details
  const FirstPhoto = await Selector("div.is-photo.type-image").nth(0).getAttribute("data-uid");
  await page.selectPhotoFromUID(FirstPhoto);
  await page.editSelected();
  await t
    .expect(Selector(".input-title input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-local-time input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-utc-time input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-day input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-month input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-year input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-timezone input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-latitude input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-longitude input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-altitude input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-country input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-camera input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-iso input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-exposure input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-lens input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-fnumber input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-focal-length input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-subject textarea").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-artist input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-copyright input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-license textarea").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-description textarea").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-keywords textarea").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-notes textarea").hasAttribute("disabled"))
    .notOk()
    .expect(Selector("button.action-apply").visible)
    .ok()
    .expect(Selector("button.action-done").visible)
    .ok();
  //labels
  await t
    .click(Selector("#tab-labels"))
    .expect(Selector("button.action-remove").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-label input").exists)
    .ok()
    .expect(Selector("button.p-photo-label-add").exists)
    .ok()
    .click(Selector("div.p-inline-edit"))
    .expect(Selector(".input-rename input").exists)
    .ok();
  //people
  await t
    .click(Selector("#tab-people"))
    .expect(Selector(".input-name input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector("button.input-reject").exists)
    .ok();
  //info
  await t
    .click(Selector("#tab-info"))
    .expect(Selector(".input-favorite input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-private input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-scan input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-panorama input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-stackable input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector(".input-type input").hasAttribute("disabled"))
    .notOk();
});

test.meta("testID", "admin-role-011")("Edit labels functionality", async (t) => {
  await page.login("admin", "photoprism");
  await page.openNav();
  await t.click(Selector(".nav-labels"));
  const FirstLabel = await Selector("a.is-label").nth(0).getAttribute("data-uid");
  await t
    .hover(Selector(`a.uid-${FirstLabel}`))
    .expect(Selector(`a.uid-${FirstLabel}`).hasClass("is-favorite"))
    .notOk()
    .click(Selector(`.uid-${FirstLabel} .input-favorite`))
    .expect(Selector(`a.uid-${FirstLabel}`).hasClass("is-favorite"))
    .ok()
    .click(Selector(`.uid-${FirstLabel} .input-favorite`))
    .expect(Selector(`a.uid-${FirstLabel}`).hasClass("is-favorite"))
    .notOk()
    .click(Selector(`a.uid-${FirstLabel} div.inline-edit`))
    .expect(Selector(".input-rename input").visible)
    .ok();
  await page.selectFromUID(FirstLabel);
  await t
    .click(Selector("button.action-menu"))
    .expect(Selector("button.action-delete").visible)
    .ok()
    .expect(Selector("button.action-album").visible)
    .ok();
});

test.meta("testID", "admin-role-012")("Edit album functionality", async (t) => {
  await page.login("admin", "photoprism");
  await t.click(Selector(".nav-albums")).expect(Selector("button.action-add").visible).ok();
  await page.checkAdminAlbumRights("album");
});

test.meta("testID", "admin-role-013")("Edit moment functionality", async (t) => {
  await page.login("admin", "photoprism");
  await t.click(Selector(".nav-moments"));
  await page.checkAdminAlbumRights("moment");
});

test.meta("testID", "admin-role-014")("Edit state functionality", async (t) => {
  await page.login("admin", "photoprism");
  await page.openNav();
  await t.click(Selector(".nav-places + div")).click(Selector(".nav-states"));
  await page.checkAdminAlbumRights("state");
});

test.meta("testID", "admin-role-015")("Edit calendar functionality", async (t) => {
  await page.login("admin", "photoprism");
  await page.openNav();
  await t.click(Selector(".nav-calendar"));
  await page.checkAdminAlbumRights("calendar");
});

test.meta("testID", "admin-role-016")("Edit folder functionality", async (t) => {
  await page.login("admin", "photoprism");
  await page.openNav();
  await t.click(Selector(".nav-folders"));
  await page.checkAdminAlbumRights("folder");
});

test.meta("testID", "admin-role-017")("Edit people functionality", async (t) => {
  await page.login("admin", "photoprism");
  await t
    .click(Selector(".nav-people"))
    .expect(Selector("#tab-people_faces > a").exists)
    .ok()
    .expect(Selector("button.action-show-hidden").exists)
    .ok()
    .expect(Selector("a div.v-card__title").withText("Otto Visible").visible)
    .ok()
    .expect(Selector("a div.v-card__title").withText("Monika Hide").visible)
    .notOk()
    .click(Selector("button.action-show-hidden"))
    .expect(Selector("a div.v-card__title").withText("Otto Visible").visible)
    .ok()
    .expect(Selector("a div.v-card__title").withText("Monika Hide").visible)
    .ok()
    .click(Selector("a div.v-card__title").nth(0))
    .expect(Selector("div.input-rename input").visible)
    .ok()
    .hover(Selector("a div.v-card__title").nth(0))
    .expect(Selector("button.input-hidden").exists)
    .ok()
    .click(Selector(`a.is-subject .input-select`).nth(0))
    .click(Selector("button.action-menu"))
    .expect(Selector("button.action-album").visible)
    .ok();
  await page.clearSelection();
  const FirstSubject = await Selector("a.is-subject").nth(0).getAttribute("data-uid");
  if (await Selector(`a.uid-${FirstSubject}`).hasClass("is-favorite")) {
    await t
      .expect(Selector(`a.uid-${FirstSubject}`).hasClass("is-favorite"))
      .ok()
      .click(Selector(`.uid-${FirstSubject} .input-favorite`))
      .expect(Selector(`a.uid-${FirstSubject}`).hasClass("is-favorite"))
      .notOk()
      .click(Selector(`.uid-${FirstSubject} .input-favorite`))
      .expect(Selector(`a.uid-${FirstSubject}`).hasClass("is-favorite"))
      .ok();
  } else {
    await t
      .expect(Selector(`a.uid-${FirstSubject}`).hasClass("is-favorite"))
      .notOk()
      .click(Selector(`.uid-${FirstSubject} .input-favorite`))
      .expect(Selector(`a.uid-${FirstSubject}`).hasClass("is-favorite"))
      .ok()
      .click(Selector(`.uid-${FirstSubject} .input-favorite`))
      .expect(Selector(`a.uid-${FirstSubject}`).hasClass("is-favorite"))
      .notOk();
  }
  await t.click(Selector("a.is-subject").nth(0));
  await page.toggleSelectNthPhoto(0);
  await page.editSelected();
  await t
    .click(Selector("#tab-people"))
    .expect(Selector(".input-name input").hasAttribute("disabled"))
    .notOk()
    .expect(Selector("div.v-input__icon--clear > i").hasClass("v-icon--disabled"))
    .notOk()
    .navigateTo("/people/new")
    .expect(Selector("div.is-face").visible)
    .ok();
});
