import { Selector } from "testcafe";
import testcafeconfig from "./testcafeconfig";
import { ClientFunction } from "testcafe";
import Menu from "../page-model/menu";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import Photo from "../page-model/photo";
import PhotoViewer from "../page-model/photoviewer";
import NewPage from "../page-model/page";
import PhotoViews from "../page-model/photo-views";
import PhotoEdit from "../page-model/photo-edit";

const scroll = ClientFunction((x, y) => window.scrollTo(x, y));
const getcurrentPosition = ClientFunction(() => window.pageYOffset);

fixture`Test photos`.page`${testcafeconfig.url}`;

const menu = new Menu();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const photo = new Photo();
const photoviewer = new PhotoViewer();
const newpage = new NewPage();
const photoviews = new PhotoViews();
const photoedit = new PhotoEdit();

test.meta("testID", "photos-001")("Scroll to top", async (t) => {
  await toolbar.setFilter("view", "Cards");
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

//TODO Covered by admin role test
test.meta("testID", "photos-002")(
  "Download single photo/video using clipboard and fullscreen mode",
  async (t) => {
    const FirstPhoto = await photo.getNthPhotoUid("image", 0);
    const SecondPhoto = await photo.getNthPhotoUid("image", 1);
    const FirstVideo = await photo.getNthPhotoUid("video", 0);
    await photoviewer.openPhotoViewer("uid", SecondPhoto);
    await photoviewer.checkPhotoViewerActionAvailability("download", true);
    await photoviewer.triggerPhotoViewerAction("close");
    await photoviews.triggerHoverAction("uid", FirstPhoto, "select");
    await photoviews.triggerHoverAction("uid", FirstVideo, "select");
    await contextmenu.checkContextMenuCount("2");
    await contextmenu.checkContextMenuActionAvailability("download", true);
  }
);

test.meta("testID", "photos-003")(
  "Approve photo using approve and by adding location",
  async (t) => {
    await menu.openPage("review");
    const FirstPhoto = await photo.getNthPhotoUid("all", 0);
    const SecondPhoto = await photo.getNthPhotoUid("all", 1);
    const ThirdPhoto = await photo.getNthPhotoUid("all", 2);
    await menu.openPage("browse");
    await photo.checkPhotoVisibility(FirstPhoto, false);
    await photo.checkPhotoVisibility(SecondPhoto, false);
    await menu.openPage("review");
    await photoviews.triggerHoverAction("uid", FirstPhoto, "select");
    await contextmenu.triggerContextMenuAction("edit", "");
    await t.click(photoedit.detailsClose);
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await toolbar.triggerToolbarAction("reload", "");
    }
    await photo.checkPhotoVisibility(FirstPhoto, true);
    await contextmenu.triggerContextMenuAction("edit", "");
    await t.click(photoedit.detailsApprove);
    if (t.browser.platform === "mobile") {
      await t.click(photoedit.detailsApply).click(photoedit.detailsClose);
    } else {
      await t.click(photoedit.detailsDone);
    }
    await photoviews.triggerHoverAction("uid", SecondPhoto, "select");
    await contextmenu.triggerContextMenuAction("edit", "");
    await t
      .typeText(Selector('input[aria-label="Latitude"]'), "9.999", { replace: true })
      .typeText(Selector('input[aria-label="Longitude"]'), "9.999", { replace: true });
    if (t.browser.platform === "mobile") {
      await t.click(photoedit.detailsApply).click(photoedit.detailsClose);
    } else {
      await t.click(photoedit.detailsDone);
    }
    await toolbar.setFilter("view", "Cards");
    const ApproveButtonThirdPhoto =
      'div.is-photo[data-uid="' + ThirdPhoto + '"] button.action-approve';
    await t.click(Selector(ApproveButtonThirdPhoto));
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await toolbar.triggerToolbarAction("reload", "");
    }
    await photo.checkPhotoVisibility(FirstPhoto, false);
    await photo.checkPhotoVisibility(SecondPhoto, false);
    await photo.checkPhotoVisibility(ThirdPhoto, false);
    await menu.openPage("browse");
    await photo.checkPhotoVisibility(FirstPhoto, true);
    await photo.checkPhotoVisibility(SecondPhoto, true);
    await photo.checkPhotoVisibility(ThirdPhoto, true);
  }
);

test.meta("testID", "photos-004")("Like/dislike photo/video", async (t) => {
  const FirstPhoto = await photo.getNthPhotoUid("image", 0);
  const SecondPhoto = await photo.getNthPhotoUid("image", 1);
  const FirstVideo = await photo.getNthPhotoUid("video", 0);
  await menu.openPage("favorites");
  await photo.checkPhotoVisibility(FirstPhoto, false);
  await photo.checkPhotoVisibility(SecondPhoto, false);
  await photo.checkPhotoVisibility(FirstVideo, false);
  await menu.openPage("browse");
  await photoviews.triggerHoverAction("uid", FirstPhoto, "favorite");
  await photoviews.triggerHoverAction("uid", FirstVideo, "favorite");
  await photoviews.triggerHoverAction("uid", SecondPhoto, "select");
  await contextmenu.triggerContextMenuAction("edit", "");
  await photoedit.turnSwitchOn("favorite");
  await t.click(photoedit.dialogClose);
  await contextmenu.clearSelection();
  await photo.checkPhotoVisibility(FirstPhoto, true);
  await photo.checkPhotoVisibility(FirstVideo, true);
  await photo.checkPhotoVisibility(SecondPhoto, true);
  await menu.openPage("favorites");
  await photo.checkPhotoVisibility(FirstPhoto, true);
  await photo.checkPhotoVisibility(FirstVideo, true);
  await photo.checkPhotoVisibility(SecondPhoto, true);
  await photoviews.triggerHoverAction("uid", SecondPhoto, "favorite");
  await photoviews.triggerHoverAction("uid", FirstVideo, "select");
  await contextmenu.triggerContextMenuAction("edit", "");
  await photoedit.turnSwitchOff("favorite");
  await t.click(photoedit.dialogClose);
  await contextmenu.clearSelection();
  await photoviewer.openPhotoViewer("uid", FirstPhoto);
  await photoviewer.triggerPhotoViewerAction("like");
  await photoviewer.triggerPhotoViewerAction("close");
  if (t.browser.platform === "mobile") {
    await t.eval(() => location.reload());
  } else {
    await toolbar.triggerToolbarAction("reload", "");
  }
  await photo.checkPhotoVisibility(FirstPhoto, false);
  await photo.checkPhotoVisibility(FirstVideo, false);
  await photo.checkPhotoVisibility(SecondPhoto, false);
});

test.meta("testID", "photos-005")("Edit photo/video", async (t) => {
  await toolbar.setFilter("view", "Cards");
  const FirstPhoto = await photo.getNthPhotoUid("image", 0);
  await t
    .click(newpage.cardTitle.withAttribute("data-uid", FirstPhoto))
    .expect(Selector('input[aria-label="Latitude"]').visible)
    .ok();
  await t.click(photoedit.dialogNext);
  await t
    .expect(photoedit.dialogPrevious.getAttribute("disabled"))
    .notEql("disabled")
    .click(photoedit.dialogPrevious)
    .click(photoedit.dialogClose);
  await photoviewer.openPhotoViewer("uid", FirstPhoto);
  await photoviewer.triggerPhotoViewerAction("edit");
  await t.expect(Selector('input[aria-label="Latitude"]').visible).ok();

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
    .click(photoedit.detailsClose)
    .click(Selector("button.action-date-edit").withAttribute("data-uid", FirstPhoto))
    .expect(Selector(".input-title input").value)
    .eql(FirstPhotoTitle);
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
    "Some notes"
  );
  if (t.browser.platform === "mobile") {
    await t.eval(() => location.reload());
  } else {
    await toolbar.triggerToolbarAction("reload", "");
  }
  await t
    .expect(newpage.cardTitle.withAttribute("data-uid", FirstPhoto).innerText)
    .eql("New Photo Title");
  await photoviews.triggerHoverAction("uid", FirstPhoto, "select");
  await contextmenu.triggerContextMenuAction("edit", "", "");
  await photoedit.checkEditFormValues(
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
    FirstPhotoNotes
  );
  await contextmenu.checkContextMenuCount("1");
  await contextmenu.clearSelection();
});

test.meta("testID", "photos-006")("Navigate from card view to place", async (t) => {
  await toolbar.setFilter("view", "Cards");
  await t
    .click(newpage.cardLocation.nth(0))
    .expect(Selector("#map").exists, { timeout: 15000 })
    .ok()
    .expect(Selector("div.p-map-control").visible)
    .ok()
    .expect(Selector(".input-search input").value)
    .notEql("");
});

test.meta("testID", "photos-007")("Mark photos/videos as panorama/scan", async (t) => {
  const FirstPhoto = await photo.getNthPhotoUid("image", 0);
  const FirstVideo = await photo.getNthPhotoUid("video", 1);
  await menu.openPage("scans");
  await photo.checkPhotoVisibility(FirstPhoto, false);
  await photo.checkPhotoVisibility(FirstVideo, false);
  await menu.openPage("panoramas");
  await photo.checkPhotoVisibility(FirstPhoto, false);
  await photo.checkPhotoVisibility(FirstVideo, false);
  await menu.openPage("browse");
  await photoviews.triggerHoverAction("uid", FirstPhoto, "select");
  await photoviews.triggerHoverAction("uid", FirstVideo, "select");
  await contextmenu.triggerContextMenuAction("edit", "");
  await photoedit.turnSwitchOn("scan");
  await photoedit.turnSwitchOn("panorama");
  await t.click(photoedit.dialogNext);
  await photoedit.turnSwitchOn("scan");
  await photoedit.turnSwitchOn("panorama");
  await t.click(photoedit.dialogClose);
  await contextmenu.clearSelection();
  await photo.checkPhotoVisibility(FirstPhoto, true);
  await photo.checkPhotoVisibility(FirstVideo, true);
  await menu.openPage("scans");
  await photo.checkPhotoVisibility(FirstPhoto, true);
  await photo.checkPhotoVisibility(FirstVideo, false);
  await menu.openPage("panoramas");
  await photo.checkPhotoVisibility(FirstPhoto, true);
  await photo.checkPhotoVisibility(FirstVideo, true);
  await photoviews.triggerHoverAction("uid", FirstPhoto, "select");
  await photoviews.triggerHoverAction("uid", FirstVideo, "select");
  await contextmenu.triggerContextMenuAction("edit", "");
  await photoedit.turnSwitchOff("scan");
  await photoedit.turnSwitchOff("panorama");
  await t.click(photoedit.dialogNext);
  await photoedit.turnSwitchOff("scan");
  await photoedit.turnSwitchOff("panorama");
  await t.click(photoedit.dialogClose);
  await toolbar.triggerToolbarAction("reload");
  await contextmenu.clearSelection();
  await photo.checkPhotoVisibility(FirstPhoto, false);
  await photo.checkPhotoVisibility(FirstVideo, false);
});
