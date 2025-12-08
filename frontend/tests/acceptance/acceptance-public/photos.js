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
const getcurrentPosition = ClientFunction(() => window.scrollY);

fixture`Test photos`.page`${testcafeconfig.url}`;

const menu = new Menu();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const photo = new Photo();
const photoviewer = new PhotoViewer();
const page = new Page();
const photoedit = new PhotoEdit();

test.meta("testID", "photos-001").meta({ mode: "public" })("Common: Scroll to top", async (t) => {
  await t.click(toolbar.cardsViewAction);

  await t
    .expect(Selector("button.is-photo-scroll-top").exists)
    .notOk()
    .expect(getcurrentPosition())
    .eql(0)
    .expect(Selector("div.type-image.result").nth(0).visible)
    .ok();

  await t.scroll("bottom");
  await t.pressKey("pageUp");

  await t.click(Selector("button.p-scroll")).expect(getcurrentPosition()).eql(0);
});

test.meta("testID", "photos-002").meta({ mode: "public" })(
  "Common: Download single photo/video using clipboard and fullscreen mode",
  async (t) => {
    const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
    const SecondPhotoUid = await photo.getNthPhotoUid("image", 1);
    const FirstVideoUid = await photo.getNthPhotoUid("video", 0);
    await photoviewer.openPhotoViewer("uid", SecondPhotoUid);

    await photoviewer.checkPhotoViewerActionAvailability("download", true);

    await photoviewer.triggerPhotoViewerAction("close-button");
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
      await toolbar.triggerToolbarAction("refresh");
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
    await t.typeText(photoedit.coordinates, "9.999,9.999", { replace: true });
    await t.click(photoedit.detailsApply).click(photoedit.detailsClose);
    await t.click(toolbar.cardsViewAction);
    const ApproveButtonThirdPhoto = 'div.is-photo[data-uid="' + ThirdPhotoUid + '"] button.action-approve';
    await t.click(Selector(ApproveButtonThirdPhoto));
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await toolbar.triggerToolbarAction("refresh");
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
    await photoviewer.triggerPhotoViewerAction("close-button");
    await t.expect(Selector("div.p-lightbox__pswp").visible).notOk();
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await toolbar.triggerToolbarAction("refresh");
    }

    await photo.checkPhotoVisibility(FirstPhotoUid, false);
    await photo.checkPhotoVisibility(FirstVideoUid, false);
    await photo.checkPhotoVisibility(SecondPhotoUid, false);
  }
);

test.meta("testID", "photos-005").meta({ type: "short", mode: "public" })("Common: Edit photo/video", async (t) => {
  await menu.openPage("browse");
  await t.click(toolbar.cardsViewAction);
  await toolbar.search("geo:true");
  const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
  await page.clickCardTitleOfUID(FirstPhotoUid);

  await t.expect(photoedit.coordinates.visible).ok();

  await t.click(photoedit.dialogNext);

  await t.expect(photoedit.dialogPrevious.getAttribute("disabled")).notEql("disabled");

  await t.click(photoedit.dialogPrevious).click(photoedit.dialogClose);
  await photoviewer.openPhotoViewer("uid", FirstPhotoUid);
  await photoviewer.triggerPhotoViewerAction("edit-button");
  const FirstPhotoTitle = await photoedit.title.value;
  const FirstPhotoLocalTime = await photoedit.localTime.value;
  let FirstPhotoDay = await photoedit.dayValue.innerText;
  if (!FirstPhotoDay) {
    FirstPhotoDay = "Unknown";
  }
  let FirstPhotoMonth = await photoedit.monthValue.innerText;
  if (!FirstPhotoMonth) {
    FirstPhotoMonth = "Unknown";
  }
  let FirstPhotoYear = await photoedit.yearValue.innerText;
  if (!FirstPhotoYear) {
    FirstPhotoYear = "Unknown";
  }
  const FirstPhotoTimezone = await photoedit.timezoneValue.innerText;
  const FirstPhotoCoordinates = await photoedit.coordinates.value;
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

  const expectedInputValues = [
    ["title", "New Photo Title"],
    ["localTime", "04:30:30"],
    ["altitude", "-1"],
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
  const expectedSelectValuesNoCountry = [
    ["day", "15"],
    ["month", "07"],
    ["year", "2019"],
    ["timezone", "Europe/Moscow"],
    ["camera", "Canon EOS M10"],
    ["lens", "EF-M15-45mm f/3.5-6.3 IS STM"],
  ];
  const initialInputValues = [
    ["title", FirstPhotoTitle],
    ["localTime", FirstPhotoLocalTime],
    ["altitude", FirstPhotoAltitude],
    ["coordinates", FirstPhotoCoordinates],
    ["iso", FirstPhotoIso],
    ["exposure", FirstPhotoExposure],
    ["fnumber", FirstPhotoFnumber],
    ["focallength", FirstPhotoFocalLength],
    ["subject", FirstPhotoSubject],
    ["artist", FirstPhotoArtist],
    ["copyright", FirstPhotoCopyright],
    ["license", FirstPhotoLicense],
    ["description", FirstPhotoDescription],
    ["notes", FirstPhotoNotes],
  ];
  const initialSelectValuesNoCountry = [
    ["day", FirstPhotoDay],
    ["month", FirstPhotoMonth],
    ["year", FirstPhotoYear],
    ["timezone", FirstPhotoTimezone],
    ["camera", FirstPhotoCamera],
    ["lens", FirstPhotoLens],
  ];
  await t.typeText(photoedit.title, "Not saved photo title", { replace: true }).click(photoedit.detailsClose);
  await page.clickCardTitleOfUID(FirstPhotoUid);

  await t.expect(photoedit.title.value).eql(FirstPhotoTitle);

  await photoedit.editFormValues(expectedInputValues, expectedSelectValuesNoCountry);
  await page.clickCardTitleOfUID(FirstPhotoUid);

  await t.typeText(Selector("div.p-tab-photo-details .input-coordinates input"), "41.15333, 20.168331", {
    replace: true,
  });
  await t.expect(await photoedit.coordinates.value).eql("41.15333, 20.168331");

  await t.click(photoedit.detailsApply);
  await t.expect(await photoedit.coordinates.value).eql("41.15333, 20.168331");

  await t.click(photoedit.detailsClose);
  if (t.browser.platform === "mobile") {
    await t.eval(() => location.reload());
  } else {
    await toolbar.triggerToolbarAction("refresh");
  }
  await toolbar.search("uid:" + FirstPhotoUid);

  await t
    .expect(Selector('div[data-uid="' + FirstPhotoUid + '"] button.action-title-edit').innerText)
    .eql("New Photo Title");

  await photo.triggerHoverAction("uid", FirstPhotoUid, "select");
  await contextmenu.triggerContextMenuAction("edit", "");
  await photoedit.checkEditFormValues(expectedInputValues, expectedSelectValues);
  await t.expect(await photoedit.coordinates.value).eql("41.15333, 20.168331");
  await photoedit.editFormValues(initialInputValues, initialSelectValuesNoCountry);
  await page.clickCardTitleOfUID(FirstPhotoUid);
  await t.typeText(Selector("div.p-tab-photo-details .input-coordinates input"), FirstPhotoCoordinates, {
    replace: true,
  });
  await t.click(photoedit.detailsApply);
  await t.click(photoedit.detailsClose);

  await contextmenu.triggerContextMenuAction("edit", "");
  await photoedit.checkEditFormValues(initialInputValues, initialSelectValuesNoCountry);

  await contextmenu.checkContextMenuCount("1");
  await contextmenu.clearSelection();
});

test.meta("testID", "photos-006").meta({ mode: "public" })(
  "Multi-Window: Navigate from card view to place",
  async (t) => {
    await t.click(toolbar.cardsViewAction);
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

  await page.clickCardTitleOfUID(FirstPhotoUid);

  await photoedit.turnSwitchOn("scan");
  await photoedit.turnSwitchOn("panorama");
  await t.click(photoedit.dialogClose);
  await page.clickCardTitleOfUID(FirstVideoUid);
  await photoedit.turnSwitchOn("scan");
  await photoedit.turnSwitchOn("panorama");
  await t.click(photoedit.dialogClose);

  await photo.checkPhotoVisibility(FirstPhotoUid, true);
  await photo.checkPhotoVisibility(FirstVideoUid, true);

  await menu.openPage("scans");

  await photo.checkPhotoVisibility(FirstPhotoUid, true);
  await photo.checkPhotoVisibility(FirstVideoUid, false);

  await menu.openPage("panoramas");

  await photo.checkPhotoVisibility(FirstPhotoUid, true);
  await photo.checkPhotoVisibility(FirstVideoUid, true);

  await page.clickCardTitleOfUID(FirstPhotoUid);

  await photoedit.turnSwitchOff("scan");
  await photoedit.turnSwitchOff("panorama");
  await t.click(photoedit.dialogClose);
  await page.clickCardTitleOfUID(FirstVideoUid);

  await photoedit.turnSwitchOff("scan");
  await photoedit.turnSwitchOff("panorama");
  await t.click(photoedit.dialogClose);
  await t.wait(9000);

  if (t.browser.platform === "mobile") {
    await t.eval(() => location.reload());
  } else {
    await toolbar.triggerToolbarAction("refresh");
  }

  await photo.checkPhotoVisibility(FirstPhotoUid, false);
  await photo.checkPhotoVisibility(FirstVideoUid, false);
});

test.meta("testID", "photos-008").meta({ mode: "public" })(
  "Multi-Window: Navigate from card view to photos taken at the same date",
  async (t) => {
    await t.click(toolbar.cardsViewAction);
    await toolbar.search("flower");
    await t.click(page.cardTaken.nth(0));

    const SearchTerm = await toolbar.search1.value;

    const PhotoCount = await photo.getPhotoCount("all");

    await t.expect(SearchTerm).eql("taken:2021-05-27").expect(PhotoCount).eql(3);
  }
);

test.meta("testID", "photos-009").meta({ mode: "public" })(
  "Common: Verify that correct time is shown in all views",
  async (t) => {
    await t.click(toolbar.cardsViewAction);
    await toolbar.search("filename:garden/20210530_125021_1993AB92.jpg");

    await t.expect(page.cardTaken.innerText).eql("Sun, May 30, 2021, 2:50 PM GMT+2");
    await photoviewer.openPhotoViewer("nth", 0);
    await photoviewer.triggerPhotoViewerAction("info-button");
    await t.expect(Selector("div").withText("May 30, 2021, 2:50 PM GMT+2").visible).ok();
    await photoviewer.triggerPhotoViewerAction("edit-button");
    await t.expect(photoedit.localTime.value).eql("14:50:21");
    await t.expect(photoedit.timezoneValue.innerText).eql("Europe/Berlin");
  }
);

test.meta("testID", "photos-010").meta({ mode: "public" })("Common: Set location on map", async (t) => {
  await t.click(toolbar.cardsViewAction);
  await toolbar.search("geo:false");

  const FirstPhotoUid = await photo.getNthPhotoUid("image", 3);
  await page.clickCardTitleOfUID(FirstPhotoUid);
  const FirstPhotoTimezone = await photoedit.timezoneValue.innerText;
  const FirstPhotoCoordinates = await photoedit.coordinates.value;
  const FirstPhotoAltitude = await photoedit.altitude.value;
  const FirstPhotoCountry = await photoedit.countryValue.innerText;

  await t.expect(photoedit.altitude.value).eql(FirstPhotoAltitude);
  await t.expect(photoedit.coordinates.value).eql(FirstPhotoCoordinates);
  await t.expect(photoedit.timezoneValue.innerText).eql(FirstPhotoTimezone);
  await t.expect(photoedit.countryValue.innerText).eql(FirstPhotoCountry);
  await t.click(photoedit.locationAction);
  const CoordinatesBefore = await photoedit.locationInput.value;
  await t.expect(CoordinatesBefore).eql("");
  await t.expect(photoedit.locationMarker.visible).notOk();

  //search
  await t.typeText(photoedit.locationSearch, "Brandenburger Tor Berlin").wait(5000).pressKey("enter");
  const Coordinates = await photoedit.locationInput.value;
  await t.expect(Coordinates).eql("52.5162546, 13.3777166");
  await t.expect(photoedit.locationMarker.visible).ok();

  await t.click(photoedit.locationClear);

  const CoordinatesAfterClear = await photoedit.locationInput.value;
  await t.expect(CoordinatesAfterClear).eql("");
  await t.expect(photoedit.locationMarker.visible).notOk();

  await t.click(photoedit.locationUndo);

  const CoordinatesAfterUndo = await photoedit.locationInput.value;
  await t.expect(CoordinatesAfterUndo).eql("52.5162546, 13.3777166");
  await t.expect(photoedit.locationMarker.visible).ok();
  await t.click(photoedit.locationConfirm).wait(10000);

  await t.expect(photoedit.altitude.value).eql("0");
  await t.expect(photoedit.coordinates.value).eql("52.5162546, 13.3777166");
  await t.expect(photoedit.countryValue.innerText).eql("Germany");
  await t.click(photoedit.detailsApply);
  await t.click(photoedit.detailsClose);
  await page.clickCardTitleOfUID(FirstPhotoUid);

  await t.expect(photoedit.timezoneValue.innerText).eql("Europe/Berlin");
  await t.expect(photoedit.altitude.value).eql("0");
  await t.expect(photoedit.coordinates.value).eql("52.5162546, 13.3777166");
  await t.expect(photoedit.countryValue.innerText).eql("Germany");

  //click on map
  await t.click(photoedit.locationAction);
  const CoordinatesBeforeChange = await photoedit.locationInput.value;
  await t.expect(CoordinatesBeforeChange).eql("52.5162546, 13.3777166");
  await t.click(Selector("div.maplibregl-map"), { offsetX: 4, offsetY: 4 });
  const CoordinatesAfterChange = await photoedit.locationInput.value;
  await t.expect(CoordinatesAfterChange).eql("52.534636098259455, 13.332140504419073");
  await t.click(photoedit.locationCancel).wait(10000);
  await t.expect(photoedit.coordinates.value).eql("52.5162546, 13.3777166");
});
