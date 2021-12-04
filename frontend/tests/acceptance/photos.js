import { Selector } from "testcafe";
import testcafeconfig from "./testcafeconfig";
import Page from "./page-model";
import { ClientFunction } from "testcafe";

const scroll = ClientFunction((x, y) => window.scrollTo(x, y));
const getcurrentPosition = ClientFunction(() => window.pageYOffset);

fixture`Test photos`.page`${testcafeconfig.url}`;

const page = new Page();

test.meta("testID", "photos-001")("Scroll to top", async (t) => {
  await page.openNav();
  await t.click(Selector(".nav-browse"));
  await page.setFilter("view", "Cards");
  await t
    .expect(Selector("button.is-photo-scroll-top").exists)
    .notOk()
    .expect(getcurrentPosition())
    .eql(0)
    .expect(Selector('div[class="v-image__image v-image__image--cover"]').nth(0).visible)
    .ok();
  await scroll(0, 1400);
  await scroll(0, 900);
  await t.click(Selector("button.p-scroll-top")).expect(getcurrentPosition()).eql(0);
});

test.meta("testID", "photos-002")(
  "Download single photo/video using clipboard and fullscreen mode",
  async (t) => {
    await page.search("photo:true");
    const FirstPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    const SecondPhoto = await Selector("div.is-photo").nth(1).getAttribute("data-uid");
    await t.click(Selector("div").withAttribute("data-uid", SecondPhoto));
    await t
      .expect(Selector("#photo-viewer").visible)
      .ok()
      .hover(Selector(".action-download"))
      .expect(Selector(".action-download").visible)
      .ok()
      .click(Selector(".action-close"));
    await page.selectPhotoFromUID(FirstPhoto);
    await page.openNav();
    await t.click(Selector(".nav-video"));
    const FirstVideo = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(FirstVideo);
    const clipboardCount = await Selector("span.count-clipboard", { timeout: 5000 });
    await t
      .expect(clipboardCount.textContent)
      .eql("2")
      .click(Selector("button.action-menu"))
      .expect(Selector("button.action-download").visible)
      .ok();
  }
);

test.meta("testID", "photos-003")(
  "Approve photo using approve and by adding location",
  async (t) => {
    await page.openNav();
    await t.click(Selector("div.nav-browse + div")).click(Selector(".nav-review"));
    await page.search("type:image");
    const FirstPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    const SecondPhoto = await Selector("div.is-photo").nth(1).getAttribute("data-uid");
    const ThirdPhoto = await Selector("div.is-photo").nth(2).getAttribute("data-uid");
    await page.openNav();

    await t.click(Selector(".nav-browse"));
    await t
      .expect(Selector("div").withAttribute("data-uid", FirstPhoto).exists, { timeout: 5000 })
      .notOk()
      .expect(Selector("div").withAttribute("data-uid", SecondPhoto).exists, { timeout: 5000 })
      .notOk();
    await page.openNav();
    await t.click(Selector(".nav-review"));

    await page.selectPhotoFromUID(FirstPhoto);
    await page.editSelected();
    await t.click(Selector("button.action-close"));
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await t.click(Selector("button.action-reload"));
    }
    await t
      .expect(Selector("div").withAttribute("data-uid", FirstPhoto).visible, { timeout: 5000 })
      .ok();
    await page.editSelected();
    await t.click(Selector("button.action-approve"));
    if (t.browser.platform === "mobile") {
      await t.click(Selector("button.action-apply")).click(Selector("button.action-close"));
    } else {
      await t.click(Selector("button.action-done", { timeout: 5000 }));
    }
    await page.selectPhotoFromUID(SecondPhoto);
    await page.editSelected();
    await t
      .typeText(Selector('input[aria-label="Latitude"]'), "9.999")
      .typeText(Selector('input[aria-label="Longitude"]'), "9.999");
    if (t.browser.platform === "mobile") {
      await t.click(Selector("button.action-apply")).click(Selector("button.action-close"));
    } else {
      await t.click(Selector("button.action-done", { timeout: 5000 }));
    }
    await page.setFilter("view", "Cards");
    const ButtonThirdPhoto = 'div.is-photo[data-uid="' + ThirdPhoto + '"] button.action-approve';
    await t.click(Selector(ButtonThirdPhoto));
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await t.click(Selector("button.action-reload"));
    }
    await t
      .expect(Selector("div").withAttribute("data-uid", FirstPhoto).exists, { timeout: 5000 })
      .notOk()
      .expect(Selector("div").withAttribute("data-uid", SecondPhoto).exists, { timeout: 5000 })
      .notOk()
      .expect(Selector("div").withAttribute("data-uid", ThirdPhoto).exists, { timeout: 5000 })
      .notOk();
    await page.openNav();
    await t.click(Selector(".nav-browse"));
    await page.search("type:image");
    await t
      .expect(Selector("div").withAttribute("data-uid", FirstPhoto).visible)
      .ok()
      .expect(Selector("div").withAttribute("data-uid", SecondPhoto).visible)
      .ok()
      .expect(Selector("div").withAttribute("data-uid", ThirdPhoto).visible)
      .ok();
  }
);

test.meta("testID", "photos-004")("Like/dislike photo/video", async (t) => {
  const FirstPhoto = await Selector("div.is-photo.type-image").nth(0).getAttribute("data-uid");
  const SecondPhoto = await Selector("div.is-photo.type-image").nth(1).getAttribute("data-uid");

  await page.openNav();
  await t.click(Selector(".nav-video"));
  const FirstVideo = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
  await page.openNav();
  await t.click(Selector(".nav-favorites"));
  await t
    .expect(Selector("div").withAttribute("data-uid", FirstPhoto).exists, { timeout: 5000 })
    .notOk()
    .expect(Selector("div").withAttribute("data-uid", SecondPhoto).exists, { timeout: 5000 })
    .notOk()
    .expect(Selector("div").withAttribute("data-uid", FirstVideo).exists, { timeout: 5000 })
    .notOk();
  await page.openNav();
  await t.click(Selector(".nav-browse"));
  await page.toggleLike(FirstPhoto);
  await page.selectPhotoFromUID(SecondPhoto);
  await page.editSelected();
  await page.turnSwitchOn("favorite");
  await t.click(Selector(".action-close"));
  await t
    .expect(Selector("div.is-photo").withAttribute("data-uid", FirstPhoto).exists, {
      timeout: 5000,
    })
    .ok()
    .expect(Selector("div.is-photo").withAttribute("data-uid", SecondPhoto).exists, {
      timeout: 5000,
    })
    .ok();
  await page.openNav();
  await t.click(Selector(".nav-video"));
  await page.toggleLike(FirstVideo);
  if (t.browser.platform === "mobile") {
    await t.eval(() => location.reload());
  } else {
    await t.click(Selector("button.action-reload"));
  }
  await t
    .expect(Selector("div.is-photo").withAttribute("data-uid", FirstVideo).exists, {
      timeout: 5000,
    })
    .ok();
  await page.openNav();
  await t.click(Selector(".nav-favorites"));
  await t
    .expect(Selector("div").withAttribute("data-uid", FirstPhoto).exists, { timeout: 5000 })
    .ok()
    .expect(Selector("div").withAttribute("data-uid", SecondPhoto).exists, { timeout: 5000 })
    .ok()
    .expect(Selector("div").withAttribute("data-uid", FirstVideo).exists, { timeout: 5000 })
    .ok();
  await page.toggleLike(FirstVideo);
  await page.toggleLike(FirstPhoto);
  await page.editSelected();
  await page.turnSwitchOff("private");
  await t.click(Selector(".action-close"));
  await page.clearSelection();
  if (t.browser.platform === "mobile") {
    await t.eval(() => location.reload());
  } else {
    await t.click(Selector("button.action-reload"));
  }
  await t
    .expect(Selector("div").withAttribute("data-uid", FirstPhoto).exists, { timeout: 5000 })
    .notOk()
    .expect(Selector("div").withAttribute("data-uid", SecondPhoto).exists, { timeout: 5000 })
    .notOk()
    .expect(Selector("div").withAttribute("data-uid", FirstVideo).exists, { timeout: 5000 })
    .notOk();
});

test.meta("testID", "photos-007")("Edit photo/video", async (t) => {
  await page.openNav();
  await t.click(Selector(".nav-browse"));
  await page.setFilter("view", "Cards");
  const FirstPhoto = await Selector("div.is-photo.type-image").nth(0).getAttribute("data-uid");
  await t
    .click(Selector("button.action-title-edit").withAttribute("data-uid", FirstPhoto))
    .expect(Selector('input[aria-label="Latitude"]').visible)
    .ok();
  await t.click(Selector("button.action-next"));
  await t
    .expect(Selector("button.action-previous").getAttribute("disabled"))
    .notEql("disabled")
    .click(Selector("button.action-previous"))
    .click(Selector("button.action-close"))
    .click(Selector("div.is-photo").withAttribute("data-uid", FirstPhoto))
    .expect(Selector("#photo-viewer").visible)
    .ok()
    .hover(Selector(".action-edit"))
    .click(Selector(".action-edit"))
    .expect(Selector('input[aria-label="Latitude"]').visible)
    .ok();

  const FirstPhotoTitle = await Selector(".input-title input").value;
  const FirstPhotoLocalTime = await Selector(".input-local-time input").value;
  const FirstPhotoDay = await Selector(".input-day input").value;
  const FirstPhotoMonth = await Selector(".input-month input").value;
  const FirstPhotoYear = await Selector(".input-year input").value;
  const FirstPhotoTimezone = await Selector(".input-timezone input").value;
  const FirstPhotoLatitude = await Selector(".input-latitude input").value;
  const FirstPhotoLongitude = await Selector(".input-longitude input").value;
  const FirstPhotoAltitude = await Selector(".input-altitude input").value;
  const FirstPhotoCountry = await Selector(".input-country input").value;
  const FirstPhotoCamera = await Selector("div.p-camera-select div.v-select__selection").innerText;
  const FirstPhotoIso = await Selector(".input-iso input").value;
  const FirstPhotoExposure = await Selector(".input-exposure input").value;
  const FirstPhotoLens = await Selector("div.p-lens-select div.v-select__selection").innerText;
  const FirstPhotoFnumber = await Selector(".input-fnumber input").value;
  const FirstPhotoFocalLength = await Selector(".input-focal-length input").value;
  const FirstPhotoSubject = await Selector(".input-subject textarea").value;
  const FirstPhotoArtist = await Selector(".input-artist input").value;
  const FirstPhotoCopyright = await Selector(".input-copyright input").value;
  const FirstPhotoLicense = await Selector(".input-license textarea").value;
  const FirstPhotoDescription = await Selector(".input-description textarea").value;
  const FirstPhotoKeywords = await Selector(".input-keywords textarea").value;
  const FirstPhotoNotes = await Selector(".input-notes textarea").value;

  await t
    .typeText(Selector(".input-title input"), "Not saved photo title", { replace: true })
    .click(Selector("button.action-close"))
    .click(Selector("button.action-date-edit").withAttribute("data-uid", FirstPhoto))
    .expect(Selector(".input-title input").value)
    .eql(FirstPhotoTitle);
  await page.editPhoto(
    "New Photo Title",
    "Europe/Moscow",
    "15",
    "07",
    "2019",
    "04:30:30",
    "-1",
    "41.15333",
    "20.168331",
    "32",
    "1/32",
    "29",
    "33",
    "Super nice edited photo",
    "Happy",
    "Happy2020",
    "Super nice cat license",
    "Description of a nice image :)",
    ", cat, love",
    "Some notes"
  );
  if (t.browser.platform === "mobile") {
    await t.eval(() => location.reload());
  } else {
    await t.click(Selector("button.action-reload"));
  }
  await t
    .expect(Selector("button.action-title-edit").withAttribute("data-uid", FirstPhoto).innerText)
    .eql("New Photo Title");
  await page.selectPhotoFromUID(FirstPhoto);
  await page.editSelected();
  await page.checkEditFormValues(
    "New Photo Title",
    "15",
    "07",
    "2019",
    "04:30:30",
    "",
    "Europe/Moscow",
    "Albania",
    "-1",
    "",
    "",
    "",
    "32",
    "1/32",
    "",
    "29",
    "33",
    "Super nice edited photo",
    "Happy",
    "Happy2020",
    "Super nice cat license",
    "Description of a nice image :)",
    "cat",
    ""
  );
  await page.undoPhotoEdit(
    FirstPhotoTitle,
    FirstPhotoTimezone,
    FirstPhotoDay,
    FirstPhotoMonth,
    FirstPhotoYear,
    FirstPhotoLocalTime,
    FirstPhotoAltitude,
    FirstPhotoLatitude,
    FirstPhotoLongitude,
    FirstPhotoCountry,
    FirstPhotoIso,
    FirstPhotoExposure,
    FirstPhotoFnumber,
    FirstPhotoFocalLength,
    FirstPhotoSubject,
    FirstPhotoArtist,
    FirstPhotoCopyright,
    FirstPhotoLicense,
    FirstPhotoDescription,
    FirstPhotoKeywords,
    FirstPhotoNotes
  );
  const clipboardCount = await Selector("span.count-clipboard", { timeout: 5000 });
  await t
    .expect(clipboardCount.textContent)
    .eql("1")
    .click(Selector(".action-clear"))
    .expect(Selector("button.action-menu").exists, { timeout: 5000 })
    .notOk();
});

test.meta("testID", "photos-008")("Change primary file", async (t) => {
  await page.openNav();
  await t.click(Selector(".nav-browse")).click(Selector(".p-expand-search"));
  await page.search("ski");
  const SequentialPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");

  await t.expect(Selector(".input-open").visible).ok();
  if (t.browser.platform === "desktop") {
    console.log(t.browser.platform);
    await t
      .click(Selector(".input-open"))
      .click(Selector(".action-next", { timeout: 5000 }))
      .click(Selector(".action-previous"))
      .click(Selector(".action-close"));
  }
  await page.setFilter("view", "Cards");
  await t
    .click(Selector("button.action-title-edit").withAttribute("data-uid", SequentialPhoto))
    .click(Selector("#tab-files"));
  const FirstFile = await Selector("div.caption").nth(0).innerText;
  await t
    .expect(FirstFile)
    .contains("photos8_1_ski.jpg")
    .click(Selector("li.v-expansion-panel__container").nth(1))
    .click(Selector(".action-primary"))
    .click(Selector("button.action-close"))
    .click(Selector("button.action-title-edit").withAttribute("data-uid", SequentialPhoto));
  const FirstFileAfterChange = await Selector("div.caption").nth(0).innerText;
  await t
    .expect(FirstFileAfterChange)
    .notContains("photos8_1_ski.jpg")
    .expect(FirstFileAfterChange)
    .contains("photos8_2_ski.jpg");
});

test.meta("testID", "photos-009")("Navigate from card view to place", async (t) => {
  await page.setFilter("view", "Cards");
  await t
    .click(Selector("button.action-location").nth(0))
    .expect(Selector("#map").exists, { timeout: 15000 })
    .ok()
    .expect(Selector("div.p-map-control").visible)
    .ok()
    .expect(Selector(".input-search input").value)
    .notEql("");
});

test.meta("testID", "photos-010")("Ungroup files", async (t) => {
  await page.openNav();
  await t.click(Selector(".nav-browse")).click(Selector(".p-expand-search"));
  await page.search("group");
  await page.setFilter("view", "Cards");
  const PhotoCount = await Selector("button.action-title-edit", { timeout: 5000 }).count;
  const SequentialPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
  await t.expect(PhotoCount).eql(1);
  await page.openNav();
  await t
    .click(Selector("div.nav-browse + div"))
    .click(Selector(".nav-stacks"))
    .expect(Selector(".input-open").visible)
    .ok()
    .click(Selector("button.action-title-edit").withAttribute("data-uid", SequentialPhoto))
    .click(Selector("#tab-files"))
    .click(Selector("div.v-expansion-panel__header__icon").nth(0))
    .click(Selector("div.v-expansion-panel__header__icon").nth(1))
    .click(Selector(".action-unstack"))
    .wait(12000)
    .click(Selector("button.action-close"));
  await page.openNav();
  await t.click(Selector(".nav-browse")).click(Selector(".p-expand-search"));
  await page.search("group");
  if (t.browser.platform === "mobile") {
    await t.eval(() => location.reload());
  } else {
    await t.click(Selector("button.action-reload"));
  }
  const PhotoCountAfterUngroup = await Selector("button.action-title-edit", { timeout: 5000 })
    .count;
  await t.expect(PhotoCountAfterUngroup).eql(2);
});

test.skip.meta("testID", "photos-011")("Delete non primary file", async (t) => {
  await page.openNav();
  await t
    .click(Selector(".nav-library"))
    .click(Selector("#tab-library-import"))
    .click(Selector(".input-import-folder input"), { timeout: 5000 })
    .click(Selector("div.v-list__tile__title").withText("/pizza"))
    .click(Selector(".action-import"))
    .wait(10000);
  await page.openNav();
  await t.click(Selector(".nav-browse")).click(Selector(".p-expand-search"));
  await page.search("mogale");
  await page.setFilter("view", "Cards");
  const PhotoCount = await Selector("button.action-title-edit", { timeout: 5000 }).count;

  const Photo = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
  await t
    .expect(PhotoCount)
    .eql(1)
    .click(Selector("button.action-title-edit").withAttribute("data-uid", Photo))
    .click(Selector("#tab-files"));
  const FileCount = await Selector("li.v-expansion-panel__container", { timeout: 5000 }).count;
  await t
    .expect(FileCount)
    .eql(2)
    .click(Selector("li.v-expansion-panel__container").nth(1))
    .click(Selector(".action-delete"))
    .click(Selector(".action-confirm"))
    .wait(10000);
  const FileCountAfterDeletion = await Selector("li.v-expansion-panel__container", {
    timeout: 5000,
  }).count;
  await t.expect(FileCountAfterDeletion).eql(1);
});

test.meta("testID", "photos-012")("Mark photos/videos as panorama/scan", async (t) => {
  await page.openNav();
  await page.search("photo:true");
  const FirstPhoto = await Selector("div.is-photo.type-image").nth(0).getAttribute("data-uid");
  await page.search("video:true");
  const FirstVideo = await Selector("div.is-photo").nth(1).getAttribute("data-uid");
  await page.openNav();
  await t
    .click(Selector(".nav-browse + div"))
    .click(Selector(".nav-scans"))
    .expect(Selector("div").withAttribute("data-uid", FirstPhoto).exists, { timeout: 5000 })
    .notOk()
    .expect(Selector("div").withAttribute("data-uid", FirstVideo).exists, { timeout: 5000 })
    .notOk();
  if (t.browser.platform === "mobile") {
    await page.openNav();
  }
  await t
    .click(Selector(".nav-panoramas"))
    .expect(Selector("div").withAttribute("data-uid", FirstPhoto).exists, { timeout: 5000 })
    .notOk()
    .expect(Selector("div").withAttribute("data-uid", FirstVideo).exists, { timeout: 5000 })
    .notOk();
  await page.openNav();
  await t.click(Selector(".nav-browse"));
  await page.selectPhotoFromUID(FirstPhoto);
  await page.editSelected();
  await page.turnSwitchOn("scan");
  await page.turnSwitchOn("panorama");
  await t.click(Selector(".action-close"));
  await page.clearSelection();
  await page.selectPhotoFromUID(FirstVideo);
  await page.editSelected();
  await page.turnSwitchOn("panorama");
  await t.click(Selector(".action-close"));
  await page.clearSelection();
  await t
    .expect(Selector("div").withAttribute("data-uid", FirstPhoto).exists, { timeout: 5000 })
    .ok()
    .expect(Selector("div").withAttribute("data-uid", FirstVideo).exists, { timeout: 5000 })
    .ok();
  if (t.browser.platform === "mobile") {
    await page.openNav();
  }
  await t
    .click(Selector(".nav-scans"))
    .expect(Selector("div").withAttribute("data-uid", FirstPhoto).exists, { timeout: 5000 })
    .ok();
  if (t.browser.platform === "mobile") {
    await page.openNav();
  }
  await t.click(Selector(".nav-panoramas"));
  await page.selectPhotoFromUID(FirstPhoto);
  await page.editSelected();
  await page.turnSwitchOff("panorama");
  await page.turnSwitchOff("scan");
  await t.click(Selector(".action-close"));
  await page.clearSelection();
  await page.selectPhotoFromUID(FirstVideo);
  await page.editSelected();
  await page.turnSwitchOff("panorama");
  await t.click(Selector(".action-close"));
  await page.clearSelection();
  if (t.browser.platform === "mobile") {
    await t.eval(() => location.reload());
  } else {
    await t.click(Selector("button.action-reload"));
  }
  await t
    .expect(Selector("div").withAttribute("data-uid", FirstPhoto).exists, { timeout: 5000 })
    .notOk()
    .expect(Selector("div").withAttribute("data-uid", FirstVideo).exists, { timeout: 5000 })
    .notOk();
  if (t.browser.platform === "mobile") {
    await page.openNav();
    await t.click(Selector(".nav-browse + div"));
  }
  await t
    .click(Selector(".nav-scans"))
    .expect(Selector("div").withAttribute("data-uid", FirstPhoto).exists, { timeout: 5000 })
    .notOk()
    .expect(Selector("div").withAttribute("data-uid", FirstVideo).exists, { timeout: 5000 })
    .notOk();
});
