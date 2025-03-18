import { Selector } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";
import { ClientFunction } from "testcafe";
import Menu from "../page-model/menu";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import Photo from "../page-model/photo";
import PhotoViewer from "../page-model/photoviewer";
import Page from "../page-model/page";
import PhotoEdit from "../page-model/photo-edit";

const scroll = ClientFunction((x, y) => window.scrollTo(x, y));
const getcurrentPosition = ClientFunction(() => window.pageYOffset);

fixture`Test photos`.page`${testcafeconfig.url}`;

const menu = new Menu();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const photo = new Photo();
const photoviewer = new PhotoViewer();
const page = new Page();
const photoedit = new PhotoEdit();

test.meta("testID", "photos-001").meta({ mode: "public" })("Common: Scroll to top", async (t) => {
  await toolbar.setFilter("view", "Cards");

  await t
    .expect(Selector("button.is-photo-scroll-top").exists)
    .notOk()
    .expect(getcurrentPosition())
    .eql(0)
    .expect(Selector("div.type-image.result").nth(0).visible)
    .ok();

    await t.scroll("bottom")
    await t.pressKey("pageUp")

  await t.click(Selector("button.p-scroll")).expect(getcurrentPosition()).eql(0);
});

test.meta("testID", "photos-002").meta({ mode: "public" })(
  "Common: Download single photo/video using clipboard and fullscreen mode",
  async (t) => {
    const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
    const SecondPhotoUid = await photo.getNthPhotoUid("image", 1);
    const FirstVideoUid = await photo.getNthPhotoUid("video", 0);
    await photoviewer.openPhotoViewer("uid", SecondPhotoUid);

    await photoviewer.checkPhotoViewerActionAvailability("download-button", true);

    await photoviewer.triggerPhotoViewerAction("close");
    await t.expect(Selector("div.p-lightbox__pswp").visible).notOk();
    await photo.triggerHoverAction("uid", FirstPhotoUid, "select");
    await photo.triggerHoverAction("uid", FirstVideoUid, "select");
    await contextmenu.checkContextMenuCount("2");

    await contextmenu.checkContextMenuActionAvailability("download", true);
  }
);

test.meta("testID", "photos-003").meta({ type: "short", mode: "public" })(
  "Common: Approve photo using approve and by adding location",
  async (t) => {
    await menu.openPage("review");
    const FirstPhotoUid = await photo.getNthPhotoUid("all", 0);
    const SecondPhotoUid = await photo.getNthPhotoUid("all", 1);
    const ThirdPhotoUid = await photo.getNthPhotoUid("all", 2);
    await menu.openPage("browse");

    await photo.checkPhotoVisibility(FirstPhotoUid, false);
    await photo.checkPhotoVisibility(SecondPhotoUid, false);

    await menu.openPage("review");
    await photo.triggerHoverAction("uid", FirstPhotoUid, "select");
    await contextmenu.triggerContextMenuAction("edit", "");
    await t.click(photoedit.detailsClose);
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await toolbar.triggerToolbarAction("reload");
    }

    await photo.checkPhotoVisibility(FirstPhotoUid, true);

    await contextmenu.triggerContextMenuAction("edit", "");
    await t.click(photoedit.detailsApprove);
    if (t.browser.platform === "mobile") {
      await t.click(photoedit.detailsApply).click(photoedit.detailsClose);
    } else {
      await t.click(photoedit.detailsClose);
    }
    await photo.triggerHoverAction("uid", SecondPhotoUid, "select");
    await contextmenu.triggerContextMenuAction("edit", "");
    await t
      .typeText(photoedit.latitude, "9.999", { replace: true })
      .typeText(photoedit.longitude, "9.999", { replace: true });
    await t.click(photoedit.detailsApply).click(photoedit.detailsClose);
    await toolbar.setFilter("view", "Cards");
    const ApproveButtonThirdPhoto = 'div.is-photo[data-uid="' + ThirdPhotoUid + '"] button.action-approve';
    await t.click(Selector(ApproveButtonThirdPhoto));
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await toolbar.triggerToolbarAction("reload");
    }

    await photo.checkPhotoVisibility(FirstPhotoUid, false);
    await photo.checkPhotoVisibility(SecondPhotoUid, false);
    await photo.checkPhotoVisibility(ThirdPhotoUid, false);
    await menu.openPage("browse");
    await photo.checkPhotoVisibility(FirstPhotoUid, true);
    await photo.checkPhotoVisibility(SecondPhotoUid, true);
    await photo.checkPhotoVisibility(ThirdPhotoUid, true);
  }
);

test.meta("testID", "photos-004").meta({ type: "short", mode: "public" })(
  "Common: Like/dislike photo/video",
  async (t) => {
    const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
    const SecondPhotoUid = await photo.getNthPhotoUid("image", 1);
    const FirstVideoUid = await photo.getNthPhotoUid("video", 0);
    await menu.openPage("favorites");

    await photo.checkPhotoVisibility(FirstPhotoUid, false);
    await photo.checkPhotoVisibility(SecondPhotoUid, false);
    await photo.checkPhotoVisibility(FirstVideoUid, false);

    await menu.openPage("browse");
    await photo.triggerHoverAction("uid", FirstPhotoUid, "favorite");
    await photo.triggerHoverAction("uid", FirstVideoUid, "favorite");
    await photo.triggerHoverAction("uid", SecondPhotoUid, "select");
    await contextmenu.triggerContextMenuAction("edit", "");
    await photoedit.turnSwitchOn("favorite");
    await t.click(photoedit.dialogClose);
    await contextmenu.clearSelection();

    await photo.checkPhotoVisibility(FirstPhotoUid, true);
    await photo.checkPhotoVisibility(FirstVideoUid, true);
    await photo.checkPhotoVisibility(SecondPhotoUid, true);

    await menu.openPage("favorites");

    await photo.checkPhotoVisibility(FirstPhotoUid, true);
    await photo.checkPhotoVisibility(FirstVideoUid, true);
    await photo.checkPhotoVisibility(SecondPhotoUid, true);

    await photo.triggerHoverAction("uid", SecondPhotoUid, "favorite");
    await photo.triggerHoverAction("uid", FirstVideoUid, "select");
    await contextmenu.triggerContextMenuAction("edit", "");
    await photoedit.turnSwitchOff("favorite");
    await t.click(photoedit.dialogClose);
    await contextmenu.clearSelection();
    await photoviewer.openPhotoViewer("uid", FirstPhotoUid);
    await photoviewer.triggerPhotoViewerAction("favorite-toggle");
    await photoviewer.triggerPhotoViewerAction("close");
    await t.expect(Selector("div.p-lightbox__pswp").visible).notOk();
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await toolbar.triggerToolbarAction("reload");
    }

    await photo.checkPhotoVisibility(FirstPhotoUid, false);
    await photo.checkPhotoVisibility(FirstVideoUid, false);
    await photo.checkPhotoVisibility(SecondPhotoUid, false);
  }
);

test.meta("testID", "photos-005").meta({ type: "short", mode: "public" })("Common: Edit photo/video", async (t) => {
  await toolbar.setFilter("view", "Cards");
  const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
  await t.click(page.cardTitle.withAttribute("data-uid", FirstPhotoUid));

  await t.expect(photoedit.latitude.visible).ok();

  await t.click(photoedit.dialogNext);

  await t.expect(photoedit.dialogPrevious.getAttribute("disabled")).notEql("disabled");

  await t.click(photoedit.dialogPrevious).click(photoedit.dialogClose);
  await photoviewer.openPhotoViewer("uid", FirstPhotoUid);
  await photoviewer.triggerPhotoViewerAction("edit-button");
  const FirstPhotoTitle = await photoedit.title.value;
  const FirstPhotoLocalTime = await photoedit.localTime.value;
  let FirstPhotoDay = await photoedit.day.innerText;
  if (!FirstPhotoDay) {
    FirstPhotoDay = "Unknown";
  }
  let FirstPhotoMonth = await photoedit.month.innerText;
  if (!FirstPhotoMonth) {
    FirstPhotoMonth = "Unknown";
  }
  let FirstPhotoYear = await photoedit.year.innerText;
  if (!FirstPhotoYear) {
    FirstPhotoYear = "Unknown";
  }
  const FirstPhotoTimezone = await photoedit.timezoneValue.innerText;
  const FirstPhotoLatitude = await photoedit.latitude.value;
  const FirstPhotoLongitude = await photoedit.longitude.value;
  const FirstPhotoAltitude = await photoedit.altitude.value;
  const FirstPhotoCountry = await photoedit.countryValue.innerText;
  const FirstPhotoCamera = await photoedit.cameraValue.innerText;
  const FirstPhotoIso = await photoedit.iso.value;
  const FirstPhotoExposure = await photoedit.exposure.value;
  const FirstPhotoLens = await photoedit.lensValue.innerText;
  const FirstPhotoFnumber = await photoedit.fnumber.value;
  const FirstPhotoFocalLength = await photoedit.focallength.value;
  const FirstPhotoSubject = await photoedit.subject.value;
  const FirstPhotoArtist = await photoedit.artist.value;
  const FirstPhotoCopyright = await photoedit.copyright.value;
  const FirstPhotoLicense = await photoedit.license.value;
  const FirstPhotoDescription = await photoedit.description.value;
  const FirstPhotoKeywords = await photoedit.keywords.value;
  const FirstPhotoNotes = await photoedit.notes.value;

  await t
    .typeText(photoedit.title, "Not saved photo title", { replace: true })
    .click(photoedit.detailsClose)
    .click(Selector("button.action-title-edit").withAttribute("data-uid", FirstPhotoUid));

  await t.expect(photoedit.title.value).eql(FirstPhotoTitle);

  await photoedit.editPhoto(
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
    "Some notes",
    "Canon EOS M10",
    "EF-M15-45mm f/3.5-6.3 IS STM"
  );
  if (t.browser.platform === "mobile") {
    await t.eval(() => location.reload());
  } else {
    await toolbar.triggerToolbarAction("reload");
  }
  await toolbar.search("uid:" + FirstPhotoUid);

  await t.expect(page.cardTitle.withAttribute("data-uid", FirstPhotoUid).innerText).eql("New Photo Title");

  await photo.triggerHoverAction("uid", FirstPhotoUid, "select");
  await contextmenu.triggerContextMenuAction("edit", "");

  const expectedInputValues = [
    ["title", "New Photo Title"],
    ["localTime", "04:30:30"],
    ["altitude", "-1"],
    ["latitude", "41.15333"],
    ["longitude", "20.168331"],
    ["iso", "32"],
    ["exposure", "1/32"],
    ["fnumber", "29"],
    ["focallength", "33"],
    ["subject", "Super nice edited photo"],
    ["artist", "Happy"],
    ["copyright", "Happy2020"],
    ["license", "Super nice cat license"],
    ["description", "Description of a nice image :)"],
    ["notes", "Some notes"],
  ];
  const expectedSelectValues = [
    ["day", "15"],
    ["month", "07"],
    ["year", "2019"],
    ["timezone", "Europe/Moscow"],
    ["country", "Albania"],
    ["camera", "Canon EOS M10"],
    ["lens", "EF-M15-45mm f/3.5-6.3 IS STM"],
  ];

  await photoedit.checkEditFormValues(expectedInputValues, expectedSelectValues);

  await photoedit.undoPhotoEdit(
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
    FirstPhotoNotes,
    FirstPhotoCamera,
    FirstPhotoLens
  );
  await contextmenu.checkContextMenuCount("1");
  await contextmenu.clearSelection();
});

test.meta("testID", "photos-006").meta({ mode: "public" })(
  "Multi-Window: Navigate from card view to place",
  async (t) => {
    await toolbar.setFilter("view", "Cards");
    await t.click(page.cardLocation.nth(0));

    await t
      .expect(Selector("div.map-loaded").exists, { timeout: 15000 })
      .ok()
      .expect(Selector("div.map-control").visible)
      .ok()
      .expect(Selector(".input-search input").value)
      .notEql("");
  }
);

test.meta("testID", "photos-007").meta({ mode: "public" })("Common: Mark photos/videos as panorama/scan", async (t) => {
  const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
  const FirstVideoUid = await photo.getNthPhotoUid("video", 1);
  await menu.openPage("scans");

  await photo.checkPhotoVisibility(FirstPhotoUid, false);
  await photo.checkPhotoVisibility(FirstVideoUid, false);

  await menu.openPage("panoramas");

  await photo.checkPhotoVisibility(FirstPhotoUid, false);
  await photo.checkPhotoVisibility(FirstVideoUid, false);

  await menu.openPage("browse");
  await photo.triggerHoverAction("uid", FirstPhotoUid, "select");
  await photo.triggerHoverAction("uid", FirstVideoUid, "select");
  await contextmenu.triggerContextMenuAction("edit", "");
  await photoedit.turnSwitchOn("scan");
  await photoedit.turnSwitchOn("panorama");
  await t.click(photoedit.dialogNext);
  await photoedit.turnSwitchOn("scan");
  await photoedit.turnSwitchOn("panorama");
  await t.click(photoedit.dialogClose);
  await contextmenu.clearSelection();

  await photo.checkPhotoVisibility(FirstPhotoUid, true);
  await photo.checkPhotoVisibility(FirstVideoUid, true);

  await menu.openPage("scans");

  await photo.checkPhotoVisibility(FirstPhotoUid, true);
  await photo.checkPhotoVisibility(FirstVideoUid, false);

  await menu.openPage("panoramas");

  await photo.checkPhotoVisibility(FirstPhotoUid, true);
  await photo.checkPhotoVisibility(FirstVideoUid, true);

  await photo.triggerHoverAction("uid", FirstPhotoUid, "select");
  await photo.triggerHoverAction("uid", FirstVideoUid, "select");
  await contextmenu.triggerContextMenuAction("edit", "");
  await photoedit.turnSwitchOff("scan");
  await photoedit.turnSwitchOff("panorama");
  await t.click(photoedit.dialogNext);
  await photoedit.turnSwitchOff("scan");
  await photoedit.turnSwitchOff("panorama");
  await t.click(photoedit.dialogClose);
  await contextmenu.clearSelection();
  if (t.browser.platform === "mobile") {
    await t.eval(() => location.reload());
  } else {
    await toolbar.triggerToolbarAction("reload");
  }

  await photo.checkPhotoVisibility(FirstPhotoUid, false);
  await photo.checkPhotoVisibility(FirstVideoUid, false);
});

test.meta("testID", "photos-008").meta({ mode: "public" })(
  "Multi-Window: Navigate from card view to photos taken at the same date",
  async (t) => {
    await toolbar.setFilter("view", "Cards");
    await toolbar.search("flower");
    await t.click(page.cardTaken.nth(0));

    const SearchTerm = await toolbar.search1.value;

    const PhotoCount = await photo.getPhotoCount("all");

    await t.expect(SearchTerm).eql("taken:2021-05-27").expect(PhotoCount).eql(3);
  }
);
