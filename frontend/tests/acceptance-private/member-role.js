import { Selector } from "testcafe";
import testcafeconfig from "../acceptance/testcafeconfig";
import Page from "../acceptance/page-model";

fixture`Test member role`.page`${testcafeconfig.url}`;

const page = new Page();

test.meta("testID", "member-role-001")("No access to settings", async (t) => {
  await page.login("member", "passwdmember");
  await page.openNav();
  await t.expect(Selector(".nav-settings").visible).notOk();
  await t.navigateTo("/settings");
  await t
    .expect(Selector(".input-language input", { timeout: 8000 }).visible)
    .notOk()
    .expect(Selector("div.p-page-photos").visible)
    .ok()
    .navigateTo("/settings/library")
    .expect(Selector("form.p-form-settings").visible)
    .notOk()
    .expect(Selector("div.p-page-photos").visible)
    .ok()
    .navigateTo("/settings/advanced")
    .expect(Selector("label").withText("Read-Only Mode").visible)
    .notOk()
    .expect(Selector("div.p-page-photos").visible)
    .ok()
    .navigateTo("/settings/sync")
    .expect(Selector("div.p-accounts-list").visible)
    .notOk()
    .expect(Selector("div.p-page-photos").visible)
    .ok();
});

test.meta("testID", "member-role-002")("No access to archive", async (t) => {
  await page.login("member", "passwdmember");
  const PhotoCountBrowse = await Selector("div.is-photo", { timeout: 5000 }).count;
  await page.openNav();
  await t.click(Selector(".nav-browse + div")).expect(Selector(".nav-archive").visible).notOk();
  await t.navigateTo("/archive");
  await t
    .expect(Selector("div.is-photo").withAttribute("data-uid", "pqnahct2mvee8sr4").visible)
    .notOk();
  const PhotoCountArchive = await Selector("div.is-photo", { timeout: 5000 }).count;
  await t.expect(PhotoCountBrowse).eql(PhotoCountArchive);
});

test.meta("testID", "member-role-003")("No access to review", async (t) => {
  await page.login("member", "passwdmember");
  const PhotoCountBrowse = await Selector("div.is-photo", { timeout: 5000 }).count;
  await page.openNav();
  await t.click(Selector(".nav-browse + div")).expect(Selector(".nav-review").visible).notOk();
  await t.navigateTo("/review");
  await t
    .expect(Selector("div.is-photo").withAttribute("data-uid", "pqzuein2pdcg1kc7").visible)
    .notOk();
  const PhotoCountReview = await Selector("div.is-photo", { timeout: 5000 }).count;
  await t.expect(PhotoCountBrowse).eql(PhotoCountReview);
});

test.meta("testID", "member-role-004")("No access to private", async (t) => {
  await page.login("member", "passwdmember");
  const PhotoCountBrowse = await Selector("div.is-photo", { timeout: 5000 }).count;
  await page.openNav();
  await t.expect(Selector(".nav-private").visible).notOk();
  await t.navigateTo("/private");
  await t
    .expect(Selector("div.is-photo").withAttribute("data-uid", "pqmxlquf9tbc8mk2").visible)
    .notOk();
  const PhotoCountPrivate = await Selector("div.is-photo", { timeout: 5000 }).count;
  await t.expect(PhotoCountBrowse).eql(PhotoCountPrivate);
});

test.meta("testID", "member-role-005")("No access to library", async (t) => {
  await page.login("member", "passwdmember");
  const PhotoCountBrowse = await Selector("div.is-photo", { timeout: 5000 }).count;
  await page.openNav();
  await t.expect(Selector(".nav-library").visible).notOk();
  await t.navigateTo("/library");
  await t
    .expect(Selector(".input-index-folder input").visible)
    .notOk()
    .expect(Selector("div.p-page-photos").visible)
    .ok()
    .navigateTo("/library/import")
    .expect(Selector(".input-import-folder input").visible)
    .notOk()
    .expect(Selector("div.p-page-photos").visible)
    .ok()
    .navigateTo("/library/logs")
    .expect(Selector("p.p-log-message").visible)
    .notOk()
    .expect(Selector("div.p-page-photos").visible)
    .ok()
    .navigateTo("/library/files")
    .expect(Selector(".nav-originals").visible)
    .notOk()
    .expect(Selector("div.p-page-files").visible)
    .notOk()
    .expect(Selector("div.p-page-photos").visible)
    .ok()
    .navigateTo("/library/hidden")
    .expect(Selector(".nav-hidden").visible)
    .notOk();
  const PhotoCountHidden = await Selector("div.is-photo", { timeout: 5000 }).count;
  await t
    .expect(PhotoCountBrowse)
    .eql(PhotoCountHidden)
    .navigateTo("/library/errors")
    .expect(Selector(".nav-errors").visible)
    .notOk()
    .expect(Selector("div.p-page-errors").visible)
    .notOk()
    .expect(Selector("div.p-page-photos").visible)
    .ok();
});

test.meta("testID", "member-role-006")(
  "No private/archived photos in search results",
  async (t) => {
    await page.login("member", "passwdmember");
    const PhotoCountBrowse = await Selector("div.is-photo", { timeout: 5000 }).count;
    await page.search("private:true");
    const PhotoCountPrivate = await Selector("div.is-photo", { timeout: 5000 }).count;
    await t.expect(PhotoCountPrivate).eql(0);
    await t
      .expect(Selector("div.is-photo").withAttribute("data-uid", "pqmxlquf9tbc8mk2").visible)
      .notOk();
    await page.search("archived:true");
    const PhotoCountArchive = await Selector("div.is-photo", { timeout: 5000 }).count;
    await t.expect(PhotoCountArchive).eql(0);
    await t
      .expect(Selector("div.is-photo").withAttribute("data-uid", "pqnahct2mvee8sr4").visible)
      .notOk();
    await page.search("quality:0");
    const PhotoCountReview = await Selector("div.is-photo", { timeout: 5000 }).count;
    await t.expect(PhotoCountReview).gte(PhotoCountBrowse);
    await t
      .expect(Selector("div.is-photo").withAttribute("data-uid", "pqzuein2pdcg1kc7").visible)
      .ok();
  }
);

test.meta("testID", "member-role-007")("No upload functionality", async (t) => {
  await page.login("member", "passwdmember");
  await t
    .expect(Selector("button.action-upload").visible)
    .notOk()
    .click(Selector(".nav-albums"))
    .expect(Selector("a.is-album").visible)
    .ok()
    .expect(Selector("button.action-upload").visible)
    .notOk()
    .click(Selector("a.is-album").nth(0))
    .expect(Selector("div.is-photo").visible)
    .ok()
    .expect(Selector("button.action-upload").visible)
    .notOk()
    .click(Selector(".nav-video"))
    .expect(Selector("button.action-upload").visible)
    .notOk()
    .click(Selector(".nav-people"))
    .expect(Selector("button.action-upload").visible)
    .notOk()
    .click(Selector(".nav-favorites"))
    .expect(Selector("button.action-upload").visible)
    .notOk()
    .click(Selector(".nav-moments"))
    .expect(Selector("a.is-album").visible)
    .ok()
    .expect(Selector("button.action-upload").visible)
    .notOk()
    .click(Selector("a.is-album").nth(0))
    .expect(Selector("div.is-photo").visible)
    .ok()
    .expect(Selector("button.action-upload").visible)
    .notOk()
    .click(Selector(".nav-calendar"))
    .expect(Selector("a.is-album").visible)
    .ok()
    .expect(Selector("button.action-upload").visible)
    .notOk()
    .click(Selector("a.is-album").nth(0))
    .expect(Selector("div.is-photo").visible)
    .ok()
    .expect(Selector("button.action-upload").visible)
    .notOk();
  await page.openNav();
  await t
    .click(Selector(".nav-places + div"))
    .click(Selector(".nav-states"))
    .expect(Selector("a.is-album").visible)
    .ok()
    .expect(Selector("button.action-upload").visible)
    .notOk()
    .click(Selector("a.is-album").nth(0))
    .expect(Selector("div.is-photo").visible)
    .ok()
    .expect(Selector("button.action-upload").visible)
    .notOk()
    .click(Selector(".nav-folders"))
    .expect(Selector("a.is-album").visible)
    .ok()
    .expect(Selector("button.action-upload").visible)
    .notOk()
    .click(Selector("a.is-album").nth(0))
    .expect(Selector("div.is-photo").visible)
    .ok()
    .expect(Selector("button.action-upload").visible)
    .notOk();
});

test.meta("testID", "member-role-008")("Member cannot like photos", async (t) => {
  await page.login("member", "passwdmember");
  await t.wait(5000);
  const FirstPhoto = await Selector("div.is-photo.type-image").nth(0).getAttribute("data-uid");
  const SecondPhoto = await Selector("div.is-photo").nth(1).getAttribute("data-uid");
  await page.openNav();
  await t.click(Selector(".nav-favorites"));
  const FirstFavorite = await Selector("div.is-photo.type-image").nth(0).getAttribute("data-uid");
  await t.expect(Selector(`div.uid-${FirstFavorite}`).hasClass("is-favorite")).ok();
  await page.toggleLike(FirstFavorite);
  await t
    .expect(Selector(`div.uid-${FirstFavorite}`).hasClass("is-favorite"))
    .ok()
    .expect(Selector("div").withAttribute("data-uid", FirstPhoto).exists, { timeout: 5000 })
    .notOk()
    .expect(Selector("div").withAttribute("data-uid", SecondPhoto).exists, { timeout: 5000 })
    .notOk();
  await page.openNav();
  await t.click(Selector(".nav-browse"));
  await t.hover(Selector("div").withAttribute("data-uid", FirstPhoto));
  if (await Selector(`.uid-${FirstPhoto} .action-fullscreen`).visible) {
    await t.click(Selector(`.uid-${FirstPhoto} .action-fullscreen`));
  } else {
    await t.click(Selector("div").withAttribute("data-uid", FirstPhoto));
  }
  await t
    .expect(Selector("#photo-viewer").visible)
    .ok()
    .expect(Selector('button[title="Like"]').exists)
    .notOk()
    .hover(Selector("button.pswp__button--close"))
    .click(Selector("button.pswp__button--close"))
    .wait(5000)
    .click(Selector(".p-expand-search", { timeout: 10000 }));
  await page.setFilter("view", "Cards");
  await t.expect(Selector(`div.uid-${FirstPhoto}`).hasClass("is-favorite")).notOk();
  await page.toggleLike(FirstPhoto);
  await t.expect(Selector(`div.uid-${FirstPhoto}`).hasClass("is-favorite")).notOk();
  await page.selectPhotoFromUID(SecondPhoto);
  await page.editSelected();
  await t
    .click("#tab-info")
    .expect(Selector(".input-favorite input").hasAttribute("disabled"))
    .ok();
  await page.turnSwitchOn("favorite");
  await t.click(Selector(".action-close"));
  await page.clearSelection();
  await t.expect(Selector(`div.uid-${SecondPhoto}`).hasClass("is-favorite")).notOk();
  await page.openNav();
  await t.click(Selector(".nav-browse"));
  await page.setFilter("view", "Mosaic");
  await t.expect(Selector(`div.uid-${FirstPhoto}`).hasClass("is-favorite")).notOk();
  await page.toggleLike(FirstPhoto);
  await t.expect(Selector(`div.uid-${FirstPhoto}`).hasClass("is-favorite")).notOk();
  await page.setFilter("view", "List");
  await t
    .expect(Selector(`button.input-like`).hasAttribute("disabled"))
    .ok()
    .click(Selector(`button.input-like`));
  await page.setFilter("view", "Cards");
  await t.expect(Selector(`div.uid-${FirstPhoto}`).hasClass("is-favorite")).notOk();
  await t
    .click(Selector(".nav-albums"))
    .click(Selector("a.is-album").nth(0))
    .click(Selector('button[title="Toggle View"]'))
    .expect(Selector(`button.input-like`).hasAttribute("disabled"))
    .ok();
});

test.meta("testID", "member-role-009")(
  "Member cannot private, archive, share, add/remove to album",
  async (t) => {
    await page.login("member", "passwdmember");
    const FirstPhoto = await Selector("div.is-photo.type-image").nth(0).getAttribute("data-uid");
    const SecondPhoto = await Selector("div.is-photo").nth(1).getAttribute("data-uid");
    await page.selectPhotoFromUID(FirstPhoto);
    await t
      .click(Selector("button.action-menu"))
      .expect(Selector("button.action-private").visible)
      .notOk()
      .expect(Selector("button.action-archive").visible)
      .notOk()
      .expect(Selector("button.action-share").visible)
      .notOk()
      .expect(Selector("button.action-album").visible)
      .notOk();
    await page.clearSelection();
    await page.setFilter("view", "List");
    await t.expect(Selector(`button.input-private`).hasAttribute("disabled")).ok();
    await t
      .click(Selector(".nav-albums"))
      .click(Selector("a.is-album").nth(0))
      .expect(Selector("button.action-share").visible)
      .notOk();
    await page.toggleSelectNthPhoto(0);
    await t
      .click(Selector("button.action-menu"))
      .expect(Selector("button.action-private").visible)
      .notOk()
      .expect(Selector("button.action-archive").visible)
      .notOk()
      .expect(Selector("button.action-share").visible)
      .notOk()
      .expect(Selector("button.action-album").visible)
      .notOk()
      .expect(Selector("button.action-remove").visible)
      .notOk();
    await page.clearSelection();
    await t
      .click(Selector('button[title="Toggle View"]'))
      .expect(Selector(`button.input-private`).hasAttribute("disabled"))
      .ok();
  }
);

test.meta("testID", "member-role-010")("Member cannot approve low quality photos", async (t) => {
  await page.login("member", "passwdmember");
  await page.search('quality:0 name:"photos-013_1"');
  await page.toggleSelectNthPhoto(0);
  await page.editSelected();
  await t.expect(Selector("button.action-approve").visible).notOk();
});

test.meta("testID", "member-role-011")("Edit dialog is read only for member", async (t) => {
  await page.login("member", "passwdmember");
  await page.search("faces:new");
  //details
  const FirstPhoto = await Selector("div.is-photo.type-image").nth(0).getAttribute("data-uid");
  await page.selectPhotoFromUID(FirstPhoto);
  await page.editSelected();
  await t
    .expect(Selector(".input-title input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-local-time input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-utc-time input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-day input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-month input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-year input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-timezone input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-latitude input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-longitude input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-altitude input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-country input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-camera input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-iso input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-exposure input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-lens input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-fnumber input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-focal-length input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-subject textarea").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-artist input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-copyright input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-license textarea").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-description textarea").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-keywords textarea").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-notes textarea").hasAttribute("disabled"))
    .ok()
    .expect(Selector("button.action-apply").visible)
    .notOk()
    .expect(Selector("button.action-done").visible)
    .notOk();
  //labels
  await t
    .click(Selector("#tab-labels"))
    .expect(Selector("button.action-remove").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-label input").exists)
    .notOk()
    .expect(Selector("button.p-photo-label-add").exists)
    .notOk()
    .click(Selector("div.p-inline-edit"))
    .expect(Selector(".input-rename input").exists)
    .notOk();
  //people
  await t
    .click(Selector("#tab-people"))
    .expect(Selector(".input-name input").hasAttribute("disabled"))
    .ok()
    .expect(Selector("button.input-reject").exists)
    .notOk();
  //info
  await t
    .click(Selector("#tab-info"))
    .expect(Selector(".input-favorite input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-private input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-scan input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-panorama input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-stackable input").hasAttribute("disabled"))
    .ok()
    .expect(Selector(".input-type input").hasAttribute("disabled"))
    .ok();
});

test.meta("testID", "member-role-012")("No edit album functionality", async (t) => {
  await page.login("member", "passwdmember");
  await t.click(Selector(".nav-albums")).expect(Selector("button.action-add").exists).notOk();
  await page.checkMemberAlbumRights("album");
});

test.meta("testID", "member-role-013")("No edit moment functionality", async (t) => {
  await page.login("member", "passwdmember");
  await t.click(Selector(".nav-moments"));
  await page.checkMemberAlbumRights("moment");
});

test.meta("testID", "member-role-014")("No edit state functionality", async (t) => {
  await page.login("member", "passwdmember");
  await page.openNav();
  await t.click(Selector(".nav-places + div")).click(Selector(".nav-states"));
  await page.checkMemberAlbumRights("state");
});

test.meta("testID", "member-role-015")("No edit calendar functionality", async (t) => {
  await page.login("member", "passwdmember");
  await page.openNav();
  await t.click(Selector(".nav-calendar"));
  await page.checkMemberAlbumRights("calendar");
});

test.meta("testID", "member-role-016")("No edit folder functionality", async (t) => {
  await page.login("member", "passwdmember");
  await page.openNav();
  await t.click(Selector(".nav-folders"));
  await page.checkMemberAlbumRights("folder");
});

test.meta("testID", "member-role-017")("No edit labels functionality", async (t) => {
  await page.login("member", "passwdmember");
  await page.openNav();
  await t.click(Selector(".nav-labels"));
  const FirstLabel = await Selector("a.is-label").nth(0).getAttribute("data-uid");
  await t
    .hover(Selector(`a.uid-${FirstLabel}`))
    .expect(Selector(`a.uid-${FirstLabel}`).hasClass("is-favorite"))
    .notOk()
    .click(Selector(`.uid-${FirstLabel} .input-favorite`))
    .expect(Selector(`a.uid-${FirstLabel}`).hasClass("is-favorite"))
    .notOk()
    .click(Selector(`a.uid-${FirstLabel} div.inline-edit`))
    .expect(Selector(".input-rename input").visible)
    .notOk();
  await page.selectFromUID(FirstLabel);
  await t
    .click(Selector("button.action-menu"))
    .expect(Selector("button.action-delete").visible)
    .notOk()
    .expect(Selector("button.action-album").visible)
    .notOk();
});

test.meta("testID", "member-role-018")("No unstack, change primary actions", async (t) => {
  await page.login("member", "passwdmember");
  await page.search("stack:true");
  //details
  const FirstPhoto = await Selector("div.is-photo.type-image").nth(0).getAttribute("data-uid");
  await page.selectPhotoFromUID(FirstPhoto);
  await page.editSelected();
  await t
    .click(Selector("#tab-files"))
    .expect(Selector("button.action-download").visible)
    .ok()
    .expect(Selector("button.action-download").hasAttribute("disabled"))
    .notOk()
    .click(Selector("li.v-expansion-panel__container").nth(1))
    .expect(Selector("button.action-download").visible)
    .ok()
    .expect(Selector("button.action-download").hasAttribute("disabled"))
    .notOk()
    .expect(Selector("button.action-unstack").visible)
    .notOk()
    .expect(Selector("button.action-primary").visible)
    .notOk()
    .expect(Selector("button.action-delete").visible)
    .notOk();
});

test.meta("testID", "member-role-019")("No edit people functionality", async (t) => {
  await page.login("member", "passwdmember");
  await t
    .click(Selector(".nav-people"))
    .expect(Selector("#tab-people_faces > a").exists)
    .notOk()
    .expect(Selector("button.action-show-hidden").exists)
    .notOk()
    .expect(Selector("button.action-upload").visible)
    .notOk()
    .expect(Selector("a div.v-card__title").withText("Otto Visible").visible)
    .ok()
    .expect(Selector("a div.v-card__title").withText("Monika Hide").visible)
    .notOk()
    .click(Selector("a div.v-card__title").nth(0))
    .expect(Selector("div.input-rename input").visible)
    .notOk()
    .hover(Selector("a div.v-card__title").nth(0))
    .expect(Selector("button.input-hidden").exists)
    .notOk()
    .click(Selector(`a.is-subject .input-select`).nth(0))
    .click(Selector("button.action-menu"))
    .expect(Selector("button.action-album").visible)
    .notOk();
  await page.clearSelection();
  const FirstSubject = await Selector("a.is-subject").nth(0).getAttribute("data-uid");
  if (await Selector(`a.uid-${FirstSubject}`).hasClass("is-favorite")) {
    await t
      .expect(Selector(`a.uid-${FirstSubject}`).hasClass("is-favorite"))
      .ok()
      .click(Selector(`.uid-${FirstSubject} .input-favorite`))
      .expect(Selector(`a.uid-${FirstSubject}`).hasClass("is-favorite"))
      .ok();
  } else {
    await t
      .expect(Selector(`a.uid-${FirstSubject}`).hasClass("is-favorite"))
      .notOk()
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
    .ok()
    .expect(Selector("div.v-input__icon--clear > i").hasClass("v-icon--disabled"))
    .ok()
    .navigateTo("/people/new")
    .expect(Selector("div.is-face").visible)
    .notOk()
    .expect(Selector("#tab-people_faces > a").exists)
    .notOk()
    .navigateTo("/people?hidden=yes&order=relevance")
    .expect(Selector("a.is-subject").visible)
    .notOk();
});
