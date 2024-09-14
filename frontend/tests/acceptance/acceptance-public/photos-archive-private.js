import { Selector } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";
import Menu from "../page-model/menu";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import Photo from "../page-model/photo";
import PhotoEdit from "../page-model/photo-edit";
import Album from "../page-model/album";
import Subject from "../page-model/subject";
import Label from "../page-model/label";

fixture`Test photos archive and private functionalities`.page`${testcafeconfig.url}`;

const menu = new Menu();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const photo = new Photo();
const photoedit = new PhotoEdit();
const album = new Album();
const label = new Label();
const subject = new Subject();

test.meta("testID", "photos-archive-private-001").meta({ type: "short", mode: "public" })(
  "Common: Private/unprivate photo/video using clipboard and list",
  async (t) => {
    await toolbar.setFilter("view", "Mosaic");
    const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
    const SecondPhotoUid = await photo.getNthPhotoUid("image", 1);
    const ThirdPhotoUid = await photo.getNthPhotoUid("image", 2);
    const FirstVideoUid = await photo.getNthPhotoUid("video", 0);
    const SecondVideoUid = await photo.getNthPhotoUid("video", 1);
    const ThirdVideoUid = await photo.getNthPhotoUid("video", 2);
    await menu.openPage("private");

    await photo.checkPhotoVisibility(FirstPhotoUid, false);
    await photo.checkPhotoVisibility(SecondPhotoUid, false);
    await photo.checkPhotoVisibility(ThirdPhotoUid, false);
    await photo.checkPhotoVisibility(FirstVideoUid, false);
    await photo.checkPhotoVisibility(SecondVideoUid, false);
    await photo.checkPhotoVisibility(ThirdVideoUid, false);

    await menu.openPage("browse");
    await photo.triggerHoverAction("uid", FirstPhotoUid, "select");
    await photo.triggerHoverAction("uid", FirstVideoUid, "select");
    await contextmenu.triggerContextMenuAction("private", "");
    await toolbar.setFilter("view", "List");
    await photo.triggerListViewActions("uid", SecondPhotoUid, "private");
    await photo.triggerListViewActions("uid", SecondVideoUid, "private");
    await toolbar.setFilter("view", "Cards");
    await photo.triggerHoverAction("uid", ThirdPhotoUid, "select");
    await photo.triggerHoverAction("uid", ThirdVideoUid, "select");
    await contextmenu.triggerContextMenuAction("edit", "");
    await photoedit.turnSwitchOn("private");
    await t.click(photoedit.dialogNext);
    await photoedit.turnSwitchOn("private");
    await t.click(photoedit.dialogClose);
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await toolbar.triggerToolbarAction("reload");
    }

    await photo.checkPhotoVisibility(FirstPhotoUid, false);
    await photo.checkPhotoVisibility(SecondPhotoUid, false);
    await photo.checkPhotoVisibility(ThirdPhotoUid, false);
    await photo.checkPhotoVisibility(FirstVideoUid, false);
    await photo.checkPhotoVisibility(SecondVideoUid, false);
    await photo.checkPhotoVisibility(ThirdVideoUid, false);

    await menu.openPage("video");

    await photo.checkPhotoVisibility(FirstVideoUid, false);
    await photo.checkPhotoVisibility(SecondVideoUid, false);
    await photo.checkPhotoVisibility(ThirdVideoUid, false);

    await menu.openPage("private");
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await toolbar.triggerToolbarAction("reload");
    }
    await photo.checkPhotoVisibility(FirstPhotoUid, true);
    await photo.checkPhotoVisibility(SecondPhotoUid, true);
    await photo.checkPhotoVisibility(ThirdPhotoUid, true);
    await photo.checkPhotoVisibility(FirstVideoUid, true);
    await photo.checkPhotoVisibility(SecondVideoUid, true);
    await photo.checkPhotoVisibility(ThirdVideoUid, true);

    await contextmenu.clearSelection();
    await photo.triggerHoverAction("uid", FirstPhotoUid, "select");
    await photo.triggerHoverAction("uid", SecondPhotoUid, "select");
    await photo.triggerHoverAction("uid", ThirdPhotoUid, "select");
    await photo.triggerHoverAction("uid", FirstVideoUid, "select");
    await photo.triggerHoverAction("uid", SecondVideoUid, "select");
    await photo.triggerHoverAction("uid", ThirdVideoUid, "select");
    await contextmenu.checkContextMenuCount("6");
    await contextmenu.triggerContextMenuAction("private", "");
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await toolbar.triggerToolbarAction("reload");
    }
    await photo.checkPhotoVisibility(FirstPhotoUid, false);
    await photo.checkPhotoVisibility(SecondPhotoUid, false);
    await photo.checkPhotoVisibility(ThirdPhotoUid, false);
    await photo.checkPhotoVisibility(FirstVideoUid, false);
    await photo.checkPhotoVisibility(SecondVideoUid, false);
    await photo.checkPhotoVisibility(ThirdVideoUid, false);

    await menu.openPage("browse");
    await toolbar.search("photo:true");

    await photo.checkPhotoVisibility(FirstPhotoUid, true);
    await photo.checkPhotoVisibility(SecondPhotoUid, true);
    await photo.checkPhotoVisibility(ThirdPhotoUid, true);
    await toolbar.search("video:true");

    await photo.checkPhotoVisibility(FirstVideoUid, true);
    await photo.checkPhotoVisibility(SecondVideoUid, true);
    await photo.checkPhotoVisibility(ThirdVideoUid, true);
  }
);

test.meta("testID", "photos-archive-private-002").meta({ type: "short", mode: "public" })(
  "Common: Archive/restore video, photos, private photos and review photos using clipboard",
  async (t) => {
    await toolbar.setFilter("view", "Mosaic");
    const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
    const SecondPhotoUid = await photo.getNthPhotoUid("image", 1);
    const FirstVideoUid = await photo.getNthPhotoUid("video", 0);
    await menu.openPage("private");
    const FirstPrivatePhotoUid = await photo.getNthPhotoUid("all", 0);
    await menu.openPage("review");
    const FirstReviewPhotoUid = await photo.getNthPhotoUid("all", 0);
    await menu.openPage("archive");

    await photo.checkPhotoVisibility(FirstPhotoUid, false);
    await photo.checkPhotoVisibility(SecondPhotoUid, false);
    await photo.checkPhotoVisibility(FirstVideoUid, false);
    await photo.checkPhotoVisibility(FirstPrivatePhotoUid, false);
    await photo.checkPhotoVisibility(FirstReviewPhotoUid, false);

    await menu.openPage("browse");
    await photo.triggerHoverAction("uid", FirstPhotoUid, "select");
    await photo.triggerHoverAction("uid", SecondPhotoUid, "select");
    await photo.triggerHoverAction("uid", FirstVideoUid, "select");
    await contextmenu.triggerContextMenuAction("archive", "");
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await toolbar.triggerToolbarAction("reload");
    }

    await photo.checkPhotoVisibility(FirstPhotoUid, false);
    await photo.checkPhotoVisibility(SecondPhotoUid, false);
    await photo.checkPhotoVisibility(FirstVideoUid, false);
    await photo.checkPhotoVisibility(FirstPrivatePhotoUid, false);
    await photo.checkPhotoVisibility(FirstReviewPhotoUid, false);

    await menu.openPage("review");
    await photo.triggerHoverAction("uid", FirstReviewPhotoUid, "select");
    await contextmenu.triggerContextMenuAction("archive", "");
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await toolbar.triggerToolbarAction("reload");
    }

    await photo.checkPhotoVisibility(FirstReviewPhotoUid, false);

    await menu.openPage("private");
    await photo.triggerHoverAction("uid", FirstPrivatePhotoUid, "select");
    await contextmenu.triggerContextMenuAction("archive", "");
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await toolbar.triggerToolbarAction("reload");
    }

    await photo.checkPhotoVisibility(FirstPrivatePhotoUid, false);

    await menu.openPage("archive");

    await photo.checkPhotoVisibility(FirstPhotoUid, true);
    await photo.checkPhotoVisibility(SecondPhotoUid, true);
    await photo.checkPhotoVisibility(FirstVideoUid, true);
    await photo.checkPhotoVisibility(FirstPrivatePhotoUid, true);
    await photo.checkPhotoVisibility(FirstReviewPhotoUid, true);

    await photo.triggerHoverAction("uid", FirstPrivatePhotoUid, "select");
    await photo.triggerHoverAction("uid", FirstReviewPhotoUid, "select");
    await photo.triggerHoverAction("uid", FirstPhotoUid, "select");
    await photo.triggerHoverAction("uid", SecondPhotoUid, "select");
    await photo.triggerHoverAction("uid", FirstVideoUid, "select");
    await contextmenu.triggerContextMenuAction("restore", "");
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await toolbar.triggerToolbarAction("reload");
    }

    await photo.checkPhotoVisibility(FirstPhotoUid, false);
    await photo.checkPhotoVisibility(SecondPhotoUid, false);
    await photo.checkPhotoVisibility(FirstVideoUid, false);
    await photo.checkPhotoVisibility(FirstPrivatePhotoUid, false);
    await photo.checkPhotoVisibility(FirstReviewPhotoUid, false);

    await menu.openPage("browse");

    await photo.checkPhotoVisibility(FirstPhotoUid, true);
    await photo.checkPhotoVisibility(SecondPhotoUid, true);
    await photo.checkPhotoVisibility(FirstVideoUid, true);
    await photo.checkPhotoVisibility(FirstPrivatePhotoUid, false);
    await photo.checkPhotoVisibility(FirstReviewPhotoUid, false);

    await menu.openPage("private");

    await photo.checkPhotoVisibility(FirstPrivatePhotoUid, true);

    await menu.openPage("review");

    await photo.checkPhotoVisibility(FirstReviewPhotoUid, true);
  }
);

test.meta("testID", "photos-archive-private-003").meta({ mode: "public" })(
  "Common: Check that archived files are not shown in monochrome/panoramas/stacks/scans/review/albums/favorites/private/videos/calendar/moments/states/labels/folders/originals",
  async (t) => {
    if (t.browser.platform === "mobile") {
      console.log("Skipped on mobile");
    } else {
      await menu.openPage("archive");
      await toolbar.setFilter("view", "Mosaic");
      const InitialPhotoCountInArchive = await photo.getPhotoCount("all");
      await menu.openPage("monochrome");
      const MonochromePhoto = await photo.getNthPhotoUid("all", 0);
      await photo.triggerHoverAction("uid", MonochromePhoto, "select");
      await menu.openPage("panoramas");
      const PanoramaPhoto = await photo.getNthPhotoUid("all", 0);
      await photo.triggerHoverAction("uid", PanoramaPhoto, "select");
      await menu.openPage("stacks");
      const StackedPhoto = await photo.getNthPhotoUid("all", 0);
      await photo.triggerHoverAction("uid", StackedPhoto, "select");
      await menu.openPage("scans");
      const ScannedPhoto = await photo.getNthPhotoUid("all", 0);
      await photo.triggerHoverAction("uid", ScannedPhoto, "select");
      await menu.openPage("review");
      const ReviewPhoto = await photo.getNthPhotoUid("all", 0);
      await photo.triggerHoverAction("uid", ReviewPhoto, "select");
      await menu.openPage("favorites");
      const FavoritesPhoto = await photo.getNthPhotoUid("all", 0);
      await photo.triggerHoverAction("uid", FavoritesPhoto, "select");
      await menu.openPage("private");
      const PrivatePhoto = await photo.getNthPhotoUid("all", 0);
      await photo.triggerHoverAction("uid", PrivatePhoto, "select");
      await menu.openPage("video");
      const Video = await photo.getNthPhotoUid("all", 0);
      await photo.triggerHoverAction("uid", Video, "select");
      await menu.openPage("calendar");
      await toolbar.search("January 2017");
      await album.openNthAlbum(0);
      const CalendarPhoto = await photo.getNthPhotoUid("all", 0);
      await photo.triggerHoverAction("uid", CalendarPhoto, "select");
      await menu.openPage("moments");
      await album.openNthAlbum(0);
      const MomentPhoto = await photo.getNthPhotoUid("all", 0);
      await photo.triggerHoverAction("uid", MomentPhoto, "select");
      await menu.openPage("states");
      await toolbar.search("Western Cape");
      await album.openNthAlbum(0);
      const StatesPhoto = await photo.getNthPhotoUid("all", 0);
      await photo.triggerHoverAction("uid", StatesPhoto, "select");
      await menu.openPage("labels");
      await toolbar.search("Seashore");
      await label.openNthLabel(0);
      const LabelPhoto = await photo.getNthPhotoUid("all", 0);
      await photo.triggerHoverAction("uid", LabelPhoto, "select");
      await menu.openPage("people");
      await subject.openNthSubject(0);
      const SubjectPhoto = await photo.getNthPhotoUid("all", 1);
      await photo.triggerHoverAction("uid", SubjectPhoto, "select");
      await menu.openPage("folders");
      await toolbar.search("archive");
      await album.openNthAlbum(0);
      const FolderPhoto = await photo.getNthPhotoUid("all", 1);
      await photo.triggerHoverAction("uid", FolderPhoto, "select");
      await contextmenu.checkContextMenuCount("14");
      await contextmenu.triggerContextMenuAction("archive", "");
      await menu.openPage("archive");
      if (t.browser.platform === "mobile") {
        await t.eval(() => location.reload());
      } else {
        await toolbar.triggerToolbarAction("reload");
      }
      await toolbar.setFilter("view", "Mosaic");

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
      await t.navigateTo("/library/calendar/aqmxlr71p6zo22dk/january-2017");
      await photo.checkPhotoVisibility(CalendarPhoto, false);
      await menu.openPage("moments");
      await album.openNthAlbum(0);
      await photo.checkPhotoVisibility(MomentPhoto, false);
      await t.navigateTo("/library/states/aqmxlr71tebcohrw/western-cape-south-africa");
      await photo.checkPhotoVisibility(StatesPhoto, false);

      await t.navigateTo("/library/all?q=label%3Aseashore");
      await photo.checkPhotoVisibility(LabelPhoto, false);
      await menu.openPage("people");
      await subject.openNthSubject(0);
      await photo.checkPhotoVisibility(SubjectPhoto, false);
      await t.navigateTo("/library/folders/aqnah1321mgkt1w2/archive");
      await photo.checkPhotoVisibility(FolderPhoto, false);

      await menu.openPage("archive");
      await toolbar.setFilter("view", "Mosaic");

      await photo.triggerHoverAction("uid", MonochromePhoto, "select");
      await photo.triggerHoverAction("uid", PanoramaPhoto, "select");
      await photo.triggerHoverAction("uid", StackedPhoto, "select");
      await photo.triggerHoverAction("uid", ScannedPhoto, "select");
      await photo.triggerHoverAction("uid", FavoritesPhoto, "select");
      await photo.triggerHoverAction("uid", ReviewPhoto, "select");
      await photo.triggerHoverAction("uid", PrivatePhoto, "select");
      await photo.triggerHoverAction("uid", Video, "select");
      await photo.triggerHoverAction("uid", CalendarPhoto, "select");
      await photo.triggerHoverAction("uid", MomentPhoto, "select");
      await photo.triggerHoverAction("uid", StatesPhoto, "select");
      await photo.triggerHoverAction("uid", LabelPhoto, "select");
      await photo.triggerHoverAction("uid", SubjectPhoto, "select");
      await photo.triggerHoverAction("uid", FolderPhoto, "select");
      await contextmenu.checkContextMenuCount("14");
      await contextmenu.triggerContextMenuAction("restore", "");
      await toolbar.setFilter("view", "Mosaic");

      const PhotoCountInArchiveAfterRestore = await photo.getPhotoCount("all");
      await t.expect(PhotoCountInArchiveAfterRestore).eql(InitialPhotoCountInArchive);
      await menu.openPage("private");
      await photo.checkPhotoVisibility(PrivatePhoto, true);
    }
  }
);

test.meta("testID", "photos-archive-private-004").meta({ type: "short", mode: "public" })(
  "Common: Check that private files are not shown in monochrome/panoramas/stacks/scans/review/albums/favorites/archive/videos/calendar/moments/states/labels/folders/originals",
  async (t) => {
    if (t.browser.platform === "mobile") {
      console.log("Skipped on mobile");
    } else {
      await menu.openPage("private");
      await toolbar.setFilter("view", "Mosaic");

      const InitialPhotoCountInPrivate = await photo.getPhotoCount("all");
      await menu.openPage("monochrome");
      const MonochromePhoto = await photo.getNthPhotoUid("all", 0);
      await photo.triggerHoverAction("uid", MonochromePhoto, "select");
      await menu.openPage("panoramas");
      const PanoramaPhoto = await photo.getNthPhotoUid("all", 0);
      await photo.triggerHoverAction("uid", PanoramaPhoto, "select");
      await menu.openPage("stacks");
      const StackedPhoto = await photo.getNthPhotoUid("all", 0);
      await photo.triggerHoverAction("uid", StackedPhoto, "select");
      await menu.openPage("scans");
      const ScannedPhoto = await photo.getNthPhotoUid("all", 0);
      await photo.triggerHoverAction("uid", ScannedPhoto, "select");
      await menu.openPage("review");
      const ReviewPhoto = await photo.getNthPhotoUid("all", 0);
      await photo.triggerHoverAction("uid", ReviewPhoto, "select");
      await menu.openPage("favorites");
      const FavoritesPhoto = await photo.getNthPhotoUid("all", 0);
      await photo.triggerHoverAction("uid", FavoritesPhoto, "select");
      await menu.openPage("video");
      const Video = await photo.getNthPhotoUid("all", 0);
      await photo.triggerHoverAction("uid", Video, "select");
      await menu.openPage("albums");
      await toolbar.search("Holiday");
      await album.openNthAlbum(0);
      const AlbumPhoto = await photo.getNthPhotoUid("all", 0);
      await photo.triggerHoverAction("uid", AlbumPhoto, "select");
      await menu.openPage("calendar");
      await toolbar.search("January 2017");
      await album.openNthAlbum(0);
      const CalendarPhoto = await photo.getNthPhotoUid("all", 0);
      await photo.triggerHoverAction("uid", CalendarPhoto, "select");
      await menu.openPage("moments");
      await album.openNthAlbum(0);
      const MomentPhoto = await photo.getNthPhotoUid("all", 0);
      await photo.triggerHoverAction("uid", MomentPhoto, "select");
      await menu.openPage("states");
      await toolbar.search("Western Cape");
      await album.openNthAlbum(0);
      const StatesPhoto = await photo.getNthPhotoUid("all", 0);
      await photo.triggerHoverAction("uid", StatesPhoto, "select");
      await menu.openPage("labels");
      await toolbar.search("Seashore");
      await label.openNthLabel(0);
      const LabelPhoto = await photo.getNthPhotoUid("all", 0);
      await photo.triggerHoverAction("uid", LabelPhoto, "select");
      await menu.openPage("people");
      await subject.openNthSubject(0);
      const SubjectPhoto = await photo.getNthPhotoUid("all", 1);
      await photo.triggerHoverAction("uid", SubjectPhoto, "select");
      await menu.openPage("folders");
      await toolbar.search("archive");
      await album.openNthAlbum(0);
      const FolderPhoto = await photo.getNthPhotoUid("all", 1);
      await photo.triggerHoverAction("uid", FolderPhoto, "select");
      await contextmenu.checkContextMenuCount("14");
      await contextmenu.triggerContextMenuAction("private", "");
      await menu.openPage("private");
      if (t.browser.platform === "mobile") {
        await t.eval(() => location.reload());
      } else {
        await toolbar.triggerToolbarAction("reload");
      }
      await toolbar.setFilter("view", "Mosaic");

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
      await t.navigateTo("/library/albums?q=Holiday");
      await album.openNthAlbum(0);
      await photo.checkPhotoVisibility(AlbumPhoto, true);
      await t.navigateTo("/library/calendar/aqmxlr71p6zo22dk/january-2017");
      await photo.checkPhotoVisibility(CalendarPhoto, false);
      await menu.openPage("moments");
      await album.openNthAlbum(0);
      await photo.checkPhotoVisibility(MomentPhoto, false);
      await t.navigateTo("/library/states/aqmxlr71tebcohrw/western-cape-south-africa");
      await photo.checkPhotoVisibility(StatesPhoto, false);

      await t.navigateTo("/library/all?q=label%3Aseashore");
      await photo.checkPhotoVisibility(LabelPhoto, false);
      await menu.openPage("people");
      await subject.openNthSubject(0);
      await photo.checkPhotoVisibility(SubjectPhoto, false);
      await t.navigateTo("/library/folders/aqnah1321mgkt1w2/archive");
      await photo.checkPhotoVisibility(FolderPhoto, false);

      await menu.openPage("private");
      await toolbar.setFilter("view", "Mosaic");

      await photo.triggerHoverAction("uid", MonochromePhoto, "select");
      await photo.triggerHoverAction("uid", PanoramaPhoto, "select");
      await photo.triggerHoverAction("uid", StackedPhoto, "select");
      await photo.triggerHoverAction("uid", ScannedPhoto, "select");
      await photo.triggerHoverAction("uid", FavoritesPhoto, "select");
      await photo.triggerHoverAction("uid", ReviewPhoto, "select");
      await photo.triggerHoverAction("uid", Video, "select");
      await photo.triggerHoverAction("uid", CalendarPhoto, "select");
      await photo.triggerHoverAction("uid", AlbumPhoto, "select");
      await photo.triggerHoverAction("uid", MomentPhoto, "select");
      await photo.triggerHoverAction("uid", StatesPhoto, "select");
      await photo.triggerHoverAction("uid", LabelPhoto, "select");
      await photo.triggerHoverAction("uid", SubjectPhoto, "select");
      await photo.triggerHoverAction("uid", FolderPhoto, "select");
      await contextmenu.checkContextMenuCount("14");
      await contextmenu.triggerContextMenuAction("private", "");
      if (t.browser.platform === "mobile") {
        await t.eval(() => location.reload());
      } else {
        await toolbar.triggerToolbarAction("reload");
      }
      await toolbar.setFilter("view", "Mosaic");

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
      await t.navigateTo("/library/albums?q=Holiday");
      await album.openNthAlbum(0);
      await photo.checkPhotoVisibility(AlbumPhoto, true);
      await t.navigateTo("/library/calendar/aqmxlr71p6zo22dk/january-2017");
      await photo.checkPhotoVisibility(CalendarPhoto, true);
      await menu.openPage("moments");
      await album.openNthAlbum(0);
      await photo.checkPhotoVisibility(MomentPhoto, true);
      await t.navigateTo("/library/states/aqmxlr71tebcohrw/western-cape-south-africa");
      await photo.checkPhotoVisibility(StatesPhoto, true);

      await t.navigateTo("/library/all?q=label%3Aseashore");
      await photo.checkPhotoVisibility(LabelPhoto, true);
      await menu.openPage("people");
      await subject.openNthSubject(0);
      await photo.checkPhotoVisibility(SubjectPhoto, true);
      await t.navigateTo("/library/folders/aqnah1321mgkt1w2/archive");
      await photo.checkPhotoVisibility(FolderPhoto, true);
    }
  }
);

test.meta("testID", "photos-archive-private-005").meta({ type: "short", mode: "public" })(
  "Common: Check delete all dialog",
  async (t) => {
    await menu.openPage("archive");
    await toolbar.triggerToolbarAction("delete-all");
    await t
      .expect(
        Selector("div").withText("Are you sure you want to delete all archived pictures?").visible
      )
      .ok();
    await t.click(Selector("button.action-cancel"));
  }
);
