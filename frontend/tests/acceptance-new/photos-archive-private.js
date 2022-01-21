import { Selector } from "testcafe";
import testcafeconfig from "./testcafeconfig";
import Menu from "../page-model/menu";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import Photo from "../page-model/photo";
import PhotoViewer from "../page-model/photoviewer";
import NewPage from "../page-model/page";
import PhotoViews from "../page-model/photo-views";
import PhotoEdit from "../page-model/photo-edit";
import Library from "../page-model/library";
import Album from "../page-model/album";
import Subject from "../page-model/subject";
import Label from "../page-model/label";

fixture`Test photos archive and private functionalities`.page`${testcafeconfig.url}`;

const menu = new Menu();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const photo = new Photo();
const photoviewer = new PhotoViewer();
const newpage = new NewPage();
const photoviews = new PhotoViews();
const photoedit = new PhotoEdit();
const library = new Library();
const album = new Album();
const label = new Label();
const subject = new Subject();

test.meta("testID", "photos-005")(
  "Private/unprivate photo/video using clipboard and list",
  async (t) => {
    await toolbar.setFilter("view", "Mosaic");
    const FirstPhoto = await photo.getNthPhotoUid("image", 0);
    const SecondPhoto = await photo.getNthPhotoUid("image", 1);
    const ThirdPhoto = await photo.getNthPhotoUid("image", 2);
    const FirstVideo = await photo.getNthPhotoUid("video", 0);
    const SecondVideo = await photo.getNthPhotoUid("video", 1);
    const ThirdVideo = await photo.getNthPhotoUid("video", 2);
    await menu.openPage("private");
    await photo.checkPhotoVisibility(FirstPhoto, false);
    await photo.checkPhotoVisibility(SecondPhoto, false);
    await photo.checkPhotoVisibility(ThirdPhoto, false);
    await photo.checkPhotoVisibility(FirstVideo, false);
    await photo.checkPhotoVisibility(SecondVideo, false);
    await photo.checkPhotoVisibility(ThirdVideo, false);
    await menu.openPage("browse");
    await photoviews.triggerHoverAction("uid", FirstPhoto, "select");
    await photoviews.triggerHoverAction("uid", FirstVideo, "select");
    await contextmenu.triggerContextMenuAction("private", "");
    await toolbar.setFilter("view", "List");
    await photoviews.triggerListViewActions("uid", SecondPhoto, "private");
    await photoviews.triggerListViewActions("uid", SecondVideo, "private");
    await toolbar.setFilter("view", "Cards");
    await photoviews.triggerHoverAction("uid", ThirdPhoto, "select");
    await photoviews.triggerHoverAction("uid", ThirdVideo, "select");
    await contextmenu.triggerContextMenuAction("edit", "", "");
    await photoedit.turnSwitchOn("private");
    await t.click(photoedit.dialogNext);
    await photoedit.turnSwitchOn("private");
    await t.click(photoedit.dialogClose);
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await toolbar.triggerToolbarAction("reload", "");
    }
    await photo.checkPhotoVisibility(FirstPhoto, false);
    await photo.checkPhotoVisibility(SecondPhoto, false);
    await photo.checkPhotoVisibility(ThirdPhoto, false);
    await photo.checkPhotoVisibility(FirstVideo, false);
    await photo.checkPhotoVisibility(SecondVideo, false);
    await photo.checkPhotoVisibility(ThirdVideo, false);
    await menu.openPage("video");
    await photo.checkPhotoVisibility(FirstVideo, false);
    await photo.checkPhotoVisibility(SecondVideo, false);
    await photo.checkPhotoVisibility(ThirdVideo, false);
    await menu.openPage("private");
    await photo.checkPhotoVisibility(FirstPhoto, true);
    await photo.checkPhotoVisibility(SecondPhoto, true);
    await photo.checkPhotoVisibility(ThirdPhoto, true);
    await photo.checkPhotoVisibility(FirstVideo, true);
    await photo.checkPhotoVisibility(SecondVideo, true);
    await photo.checkPhotoVisibility(ThirdVideo, true);
    await contextmenu.clearSelection();
    await photoviews.triggerHoverAction("uid", FirstPhoto, "select");
    await photoviews.triggerHoverAction("uid", SecondPhoto, "select");
    await photoviews.triggerHoverAction("uid", ThirdPhoto, "select");
    await photoviews.triggerHoverAction("uid", FirstVideo, "select");
    await photoviews.triggerHoverAction("uid", SecondVideo, "select");
    await photoviews.triggerHoverAction("uid", ThirdVideo, "select");
    await contextmenu.checkContextMenuCount("6");
    await contextmenu.triggerContextMenuAction("private", "");
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await toolbar.triggerToolbarAction("reload", "");
    }
    await photo.checkPhotoVisibility(FirstPhoto, false);
    await photo.checkPhotoVisibility(SecondPhoto, false);
    await photo.checkPhotoVisibility(ThirdPhoto, false);
    await photo.checkPhotoVisibility(FirstVideo, false);
    await photo.checkPhotoVisibility(SecondVideo, false);
    await photo.checkPhotoVisibility(ThirdVideo, false);
    await menu.openPage("browse");
    await photo.checkPhotoVisibility(FirstPhoto, true);
    await photo.checkPhotoVisibility(SecondPhoto, true);
    await photo.checkPhotoVisibility(ThirdPhoto, true);
    await photo.checkPhotoVisibility(FirstVideo, true);
    await photo.checkPhotoVisibility(SecondVideo, true);
    await photo.checkPhotoVisibility(ThirdVideo, true);
  }
);

test.meta("testID", "photos-006")(
  "Archive/restore video, photos, private photos and review photos using clipboard",
  async (t) => {
    await toolbar.setFilter("view", "Mosaic");
    const FirstPhoto = await photo.getNthPhotoUid("image", 0);
    const SecondPhoto = await photo.getNthPhotoUid("image", 1);
    const FirstVideo = await photo.getNthPhotoUid("video", 0);
    await menu.openPage("private");
    const FirstPrivatePhoto = await photo.getNthPhotoUid("all", 0);
    await menu.openPage("review");
    const FirstReviewPhoto = await photo.getNthPhotoUid("all", 0);
    await menu.openPage("archive");
    await photo.checkPhotoVisibility(FirstPhoto, false);
    await photo.checkPhotoVisibility(SecondPhoto, false);
    await photo.checkPhotoVisibility(FirstVideo, false);
    await photo.checkPhotoVisibility(FirstPrivatePhoto, false);
    await photo.checkPhotoVisibility(FirstReviewPhoto, false);
    await menu.openPage("browse");
    await photoviews.triggerHoverAction("uid", FirstPhoto, "select");
    await photoviews.triggerHoverAction("uid", SecondPhoto, "select");
    await photoviews.triggerHoverAction("uid", FirstVideo, "select");
    await contextmenu.triggerContextMenuAction("archive", "", "");
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await toolbar.triggerToolbarAction("reload", "");
    }
    await photo.checkPhotoVisibility(FirstPhoto, false);
    await photo.checkPhotoVisibility(SecondPhoto, false);
    await photo.checkPhotoVisibility(FirstVideo, false);
    await photo.checkPhotoVisibility(FirstPrivatePhoto, false);
    await photo.checkPhotoVisibility(FirstReviewPhoto, false);
    await menu.openPage("review");
    await photoviews.triggerHoverAction("uid", FirstReviewPhoto, "select");
    await contextmenu.triggerContextMenuAction("archive", "", "");
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await toolbar.triggerToolbarAction("reload", "");
    }
    await photo.checkPhotoVisibility(FirstReviewPhoto, false);
    await menu.openPage("private");
    await photoviews.triggerHoverAction("uid", FirstPrivatePhoto, "select");
    await contextmenu.triggerContextMenuAction("archive", "", "");
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await toolbar.triggerToolbarAction("reload", "");
    }
    await photo.checkPhotoVisibility(FirstPrivatePhoto, false);
    await menu.openPage("archive");
    await photo.checkPhotoVisibility(FirstPhoto, true);
    await photo.checkPhotoVisibility(SecondPhoto, true);
    await photo.checkPhotoVisibility(FirstVideo, true);
    await photo.checkPhotoVisibility(FirstPrivatePhoto, true);
    await photo.checkPhotoVisibility(FirstReviewPhoto, true);
    await photoviews.triggerHoverAction("uid", FirstPrivatePhoto, "select");
    await photoviews.triggerHoverAction("uid", FirstReviewPhoto, "select");
    await photoviews.triggerHoverAction("uid", FirstPhoto, "select");
    await photoviews.triggerHoverAction("uid", SecondPhoto, "select");
    await photoviews.triggerHoverAction("uid", FirstVideo, "select");
    await contextmenu.triggerContextMenuAction("restore", "", "");
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await toolbar.triggerToolbarAction("reload", "");
    }
    await photo.checkPhotoVisibility(FirstPhoto, false);
    await photo.checkPhotoVisibility(SecondPhoto, false);
    await photo.checkPhotoVisibility(FirstVideo, false);
    await photo.checkPhotoVisibility(FirstPrivatePhoto, false);
    await photo.checkPhotoVisibility(FirstReviewPhoto, false);
    await menu.openPage("browse");
    await photo.checkPhotoVisibility(FirstPhoto, true);
    await photo.checkPhotoVisibility(SecondPhoto, true);
    await photo.checkPhotoVisibility(FirstVideo, true);
    await photo.checkPhotoVisibility(FirstPrivatePhoto, false);
    await photo.checkPhotoVisibility(FirstReviewPhoto, false);
    await menu.openPage("private");
    await photo.checkPhotoVisibility(FirstPrivatePhoto, true);
    await menu.openPage("review");
    await photo.checkPhotoVisibility(FirstReviewPhoto, true);
  }
);

test.meta("testID", "photos-013")(
  "Check that archived files are not shown in monochrome/panoramas/stacks/scans/review/albums/favorites/private/videos/calendar/moments/states/labels/folders/originals",
  async (t) => {
    await menu.openPage("archive");
    const InitialPhotoCountInArchive = await photo.getPhotoCount("all");
    await menu.openPage("monochrome");
    const MonochromePhoto = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", MonochromePhoto, "select");
    await menu.openPage("panoramas");
    const PanoramaPhoto = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", PanoramaPhoto, "select");
    await menu.openPage("stacks");
    const StackedPhoto = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", StackedPhoto, "select");
    await menu.openPage("scans");
    const ScannedPhoto = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", ScannedPhoto, "select");
    await menu.openPage("review");
    const ReviewPhoto = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", ReviewPhoto, "select");
    await menu.openPage("favorites");
    const FavoritesPhoto = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", FavoritesPhoto, "select");
    await menu.openPage("private");
    const PrivatePhoto = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", PrivatePhoto, "select");
    await menu.openPage("video");
    const Video = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", Video, "select");
    await menu.openPage("calendar");
    await toolbar.search("January 2017");
    await album.openNthAlbum(0);
    const CalendarPhoto = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", CalendarPhoto, "select");
    await menu.openPage("moments");
    await album.openNthAlbum(0);
    const MomentPhoto = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", MomentPhoto, "select");
    await menu.openPage("states");
    await toolbar.search("Western Cape");
    await album.openNthAlbum(0);
    const StatesPhoto = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", StatesPhoto, "select");
    await menu.openPage("labels");
    await toolbar.search("Seashore");
    await label.openNthLabel(0);
    const LabelPhoto = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", LabelPhoto, "select");
    await menu.openPage("people");
    await subject.openNthSubject(0);
    const SubjectPhoto = await photo.getNthPhotoUid("all", 1);
    await photoviews.triggerHoverAction("uid", SubjectPhoto, "select");
    await menu.openPage("folders");
    await toolbar.search("archive");
    await album.openNthAlbum(0);
    const FolderPhoto = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", FolderPhoto, "select");
    await contextmenu.checkContextMenuCount("14");
    await contextmenu.triggerContextMenuAction("archive", "", "");
    await menu.openPage("archive");
    await toolbar.triggerToolbarAction("reload");
    const PhotoCountInArchiveAfterArchive = await photo.getPhotoCount("all");
    await t.expect(PhotoCountInArchiveAfterArchive).eql(InitialPhotoCountInArchive + 14);
    await menu.openPage("monochrome");
    await photo.checkPhotoVisibility(MonochromePhoto, false);
    await menu.openPage("panoramas");
    await photo.checkPhotoVisibility(PanoramaPhoto, false);
    await menu.openPage("stacks");
    await photo.checkPhotoVisibility(StackedPhoto, false);
    await menu.openPage("scans");
    await photo.checkPhotoVisibility(ScannedPhoto, false);
    await menu.openPage("review");
    await photo.checkPhotoVisibility(ReviewPhoto, false);
    await menu.openPage("favorites");
    await photo.checkPhotoVisibility(FavoritesPhoto, false);
    await menu.openPage("private");
    await photo.checkPhotoVisibility(PrivatePhoto, false);
    await menu.openPage("video");
    await photo.checkPhotoVisibility(Video, false);
    await t.navigateTo("/calendar/aqmxlr71p6zo22dk/january-2017");
    await photo.checkPhotoVisibility(CalendarPhoto, false);
    await menu.openPage("moments");
    await album.openNthAlbum(0);
    await photo.checkPhotoVisibility(MomentPhoto, false);
    await t.navigateTo("/states/aqmxlr71tebcohrw/western-cape-south-africa");
    await photo.checkPhotoVisibility(StatesPhoto, false);

    await t.navigateTo("/all?q=label%3Aseashore");
    await photo.checkPhotoVisibility(LabelPhoto, false);
    await menu.openPage("people");
    await subject.openNthSubject(0);
    await photo.checkPhotoVisibility(SubjectPhoto, false);
    await t.navigateTo("/folders/aqnah1321mgkt1w2/archive");
    await photo.checkPhotoVisibility(FolderPhoto, false);

    await menu.openPage("archive");
    await photoviews.triggerHoverAction("uid", MonochromePhoto, "select");
    await photoviews.triggerHoverAction("uid", PanoramaPhoto, "select");
    await photoviews.triggerHoverAction("uid", StackedPhoto, "select");
    await photoviews.triggerHoverAction("uid", ScannedPhoto, "select");
    await photoviews.triggerHoverAction("uid", FavoritesPhoto, "select");
    await photoviews.triggerHoverAction("uid", ReviewPhoto, "select");
    await photoviews.triggerHoverAction("uid", PrivatePhoto, "select");
    await photoviews.triggerHoverAction("uid", Video, "select");
    await photoviews.triggerHoverAction("uid", CalendarPhoto, "select");
    await photoviews.triggerHoverAction("uid", MomentPhoto, "select");
    await photoviews.triggerHoverAction("uid", StatesPhoto, "select");
    await photoviews.triggerHoverAction("uid", LabelPhoto, "select");
    await photoviews.triggerHoverAction("uid", SubjectPhoto, "select");
    await photoviews.triggerHoverAction("uid", FolderPhoto, "select");
    await contextmenu.checkContextMenuCount("14");
    await contextmenu.triggerContextMenuAction("restore", "", "");

    const PhotoCountInArchiveAfterRestore = await photo.getPhotoCount("all");
    await t.expect(PhotoCountInArchiveAfterRestore).eql(InitialPhotoCountInArchive);
    await menu.openPage("private");
    await photo.checkPhotoVisibility(PrivatePhoto, true);
  }
);

test.meta("testID", "photos-014")(
  "Check that private files are not shown in monochrome/panoramas/stacks/scans/review/albums/favorites/archive/videos/calendar/moments/states/labels/folders/originals",
  async (t) => {
    await menu.openPage("private");
    const InitialPhotoCountInPrivate = await photo.getPhotoCount("all");
    await menu.openPage("monochrome");
    const MonochromePhoto = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", MonochromePhoto, "select");
    await menu.openPage("panoramas");
    const PanoramaPhoto = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", PanoramaPhoto, "select");
    await menu.openPage("stacks");
    const StackedPhoto = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", StackedPhoto, "select");
    await menu.openPage("scans");
    const ScannedPhoto = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", ScannedPhoto, "select");
    await menu.openPage("review");
    const ReviewPhoto = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", ReviewPhoto, "select");
    await menu.openPage("favorites");
    const FavoritesPhoto = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", FavoritesPhoto, "select");
    await menu.openPage("video");
    const Video = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", Video, "select");
    await menu.openPage("albums");
    await toolbar.search("Holiday");
    await album.openNthAlbum(0);
    const AlbumPhoto = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", AlbumPhoto, "select");
    await menu.openPage("calendar");
    await toolbar.search("January 2017");
    await album.openNthAlbum(0);
    const CalendarPhoto = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", CalendarPhoto, "select");
    await menu.openPage("moments");
    await album.openNthAlbum(0);
    const MomentPhoto = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", MomentPhoto, "select");
    await menu.openPage("states");
    await toolbar.search("Western Cape");
    await album.openNthAlbum(0);
    const StatesPhoto = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", StatesPhoto, "select");
    await menu.openPage("labels");
    await toolbar.search("Seashore");
    await label.openNthLabel(0);
    const LabelPhoto = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", LabelPhoto, "select");
    await menu.openPage("people");
    await subject.openNthSubject(0);
    const SubjectPhoto = await photo.getNthPhotoUid("all", 1);
    await photoviews.triggerHoverAction("uid", SubjectPhoto, "select");
    await menu.openPage("folders");
    await toolbar.search("archive");
    await album.openNthAlbum(0);
    const FolderPhoto = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", FolderPhoto, "select");
    await contextmenu.checkContextMenuCount("14");
    await contextmenu.triggerContextMenuAction("private", "", "");
    await menu.openPage("private");
    await toolbar.triggerToolbarAction("reload");
    const PhotoCountInPrivateAfterArchive = await photo.getPhotoCount("all");
    await t.expect(PhotoCountInPrivateAfterArchive).eql(InitialPhotoCountInPrivate + 14);
    await menu.openPage("monochrome");
    await photo.checkPhotoVisibility(MonochromePhoto, false);
    await menu.openPage("panoramas");
    await photo.checkPhotoVisibility(PanoramaPhoto, false);
    await menu.openPage("stacks");
    await photo.checkPhotoVisibility(StackedPhoto, false);
    await menu.openPage("scans");
    await photo.checkPhotoVisibility(ScannedPhoto, false);
    await menu.openPage("review");
    await photo.checkPhotoVisibility(ReviewPhoto, false);
    await menu.openPage("favorites");
    await photo.checkPhotoVisibility(FavoritesPhoto, false);
    await menu.openPage("video");
    await photo.checkPhotoVisibility(Video, false);
    await t.navigateTo("/albums?q=Holiday");
    await album.openNthAlbum(0);
    await photo.checkPhotoVisibility(AlbumPhoto, true);
    await t.navigateTo("/calendar/aqmxlr71p6zo22dk/january-2017");
    await photo.checkPhotoVisibility(CalendarPhoto, false);
    await menu.openPage("moments");
    await album.openNthAlbum(0);
    await photo.checkPhotoVisibility(MomentPhoto, false);
    await t.navigateTo("/states/aqmxlr71tebcohrw/western-cape-south-africa");
    await photo.checkPhotoVisibility(StatesPhoto, false);

    await t.navigateTo("/all?q=label%3Aseashore");
    await photo.checkPhotoVisibility(LabelPhoto, false);
    await menu.openPage("people");
    await subject.openNthSubject(0);
    await photo.checkPhotoVisibility(SubjectPhoto, false);
    await t.navigateTo("/folders/aqnah1321mgkt1w2/archive");
    await photo.checkPhotoVisibility(FolderPhoto, false);

    await menu.openPage("private");
    await photoviews.triggerHoverAction("uid", MonochromePhoto, "select");
    await photoviews.triggerHoverAction("uid", PanoramaPhoto, "select");
    await photoviews.triggerHoverAction("uid", StackedPhoto, "select");
    await photoviews.triggerHoverAction("uid", ScannedPhoto, "select");
    await photoviews.triggerHoverAction("uid", FavoritesPhoto, "select");
    await photoviews.triggerHoverAction("uid", ReviewPhoto, "select");
    await photoviews.triggerHoverAction("uid", Video, "select");
    await photoviews.triggerHoverAction("uid", CalendarPhoto, "select");
    await photoviews.triggerHoverAction("uid", AlbumPhoto, "select");
    await photoviews.triggerHoverAction("uid", MomentPhoto, "select");
    await photoviews.triggerHoverAction("uid", StatesPhoto, "select");
    await photoviews.triggerHoverAction("uid", LabelPhoto, "select");
    await photoviews.triggerHoverAction("uid", SubjectPhoto, "select");
    await photoviews.triggerHoverAction("uid", FolderPhoto, "select");
    await contextmenu.checkContextMenuCount("14");
    await contextmenu.triggerContextMenuAction("private", "", "");
    await toolbar.triggerToolbarAction("reload", "");

    const PhotoCountInPrivateAfterRestore = await photo.getPhotoCount("all");
    await t.expect(PhotoCountInPrivateAfterRestore).eql(InitialPhotoCountInPrivate);

    await menu.openPage("monochrome");
    await photo.checkPhotoVisibility(MonochromePhoto, true);
    await menu.openPage("panoramas");
    await photo.checkPhotoVisibility(PanoramaPhoto, true);
    await menu.openPage("stacks");
    await photo.checkPhotoVisibility(StackedPhoto, true);
    await menu.openPage("scans");
    await photo.checkPhotoVisibility(ScannedPhoto, true);
    await menu.openPage("review");
    await photo.checkPhotoVisibility(ReviewPhoto, true);
    await menu.openPage("favorites");
    await photo.checkPhotoVisibility(FavoritesPhoto, true);
    await menu.openPage("video");
    await photo.checkPhotoVisibility(Video, true);
    await t.navigateTo("/albums?q=Holiday");
    await album.openNthAlbum(0);
    await photo.checkPhotoVisibility(AlbumPhoto, true);
    await t.navigateTo("/calendar/aqmxlr71p6zo22dk/january-2017");
    await photo.checkPhotoVisibility(CalendarPhoto, true);
    await menu.openPage("moments");
    await album.openNthAlbum(0);
    await photo.checkPhotoVisibility(MomentPhoto, true);
    await t.navigateTo("/states/aqmxlr71tebcohrw/western-cape-south-africa");
    await photo.checkPhotoVisibility(StatesPhoto, true);

    await t.navigateTo("/all?q=label%3Aseashore");
    await photo.checkPhotoVisibility(LabelPhoto, true);
    await menu.openPage("people");
    await subject.openNthSubject(0);
    await photo.checkPhotoVisibility(SubjectPhoto, true);
    await t.navigateTo("/folders/aqnah1321mgkt1w2/archive");
    await photo.checkPhotoVisibility(FolderPhoto, true);
  }
);
