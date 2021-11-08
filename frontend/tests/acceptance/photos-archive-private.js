import { Selector } from "testcafe";
import testcafeconfig from "./testcafeconfig";
import Page from "./page-model";

fixture`Test photos archive and private functionalities`.page`${testcafeconfig.url}`;

const page = new Page();

test.meta("testID", "photos-005")(
  "Private/unprivate photo/video using clipboard and list",
  async (t) => {
    await page.search("photo:true");
    await page.setFilter("view", "Mosaic");
    const FirstPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    const SecondPhoto = await Selector("div.is-photo").nth(1).getAttribute("data-uid");
    const ThirdPhoto = await Selector("div.is-photo").nth(2).getAttribute("data-uid");
    await page.openNav();
    await t.click(Selector(".nav-video"));
    const FirstVideo = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    const SecondVideo = await Selector("div.is-photo").nth(1).getAttribute("data-uid");
    const ThirdVideo = await Selector("div.is-photo").nth(2).getAttribute("data-uid");
    await page.openNav();
    await t.click(Selector(".nav-private"));
    await t
      .expect(Selector("div").withAttribute("data-uid", FirstPhoto).exists, { timeout: 5000 })
      .notOk()
      .expect(Selector("div").withAttribute("data-uid", SecondPhoto).exists, { timeout: 5000 })
      .notOk()
      .expect(Selector("div").withAttribute("data-uid", ThirdPhoto).exists, { timeout: 5000 })
      .notOk()
      .expect(Selector("div").withAttribute("data-uid", FirstVideo).exists, { timeout: 5000 })
      .notOk()
      .expect(Selector("div").withAttribute("data-uid", SecondVideo).exists, { timeout: 5000 })
      .notOk()
      .expect(Selector("div").withAttribute("data-uid", ThirdVideo).exists, { timeout: 5000 })
      .notOk();
    await page.openNav();
    await t.click(Selector(".nav-browse"));
    await page.selectPhotoFromUID(FirstPhoto);
    await page.selectFromUIDInFullscreen(SecondPhoto);
    const clipboardCount = await Selector("span.count-clipboard", { timeout: 5000 });
    await t.expect(clipboardCount.textContent).eql("2");
    await page.privateSelected();
    await t.expect(Selector("button.action-menu").exists, { timeout: 5000 }).notOk();
    await page.setFilter("view", "List");
    await t.click(Selector("button.input-private").withAttribute("data-uid", ThirdPhoto));
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await t.click(Selector("button.action-reload"));
    }
    await t
      .expect(Selector("td").withAttribute("data-uid", FirstPhoto).exists, { timeout: 5000 })
      .notOk()
      .expect(Selector("td").withAttribute("data-uid", SecondPhoto).exists, { timeout: 5000 })
      .notOk()
      .expect(Selector("td").withAttribute("data-uid", ThirdPhoto).exists, { timeout: 5000 })
      .notOk();
    await page.openNav();
    await t.click(Selector(".nav-video"));

    await t.click(Selector("button.input-private").withAttribute("data-uid", SecondVideo));
    await page.setFilter("view", "Card");

    await page.selectPhotoFromUID(FirstVideo);
    const clipboardCountVideo = await Selector("span.count-clipboard", { timeout: 5000 });
    await t
      .expect(clipboardCountVideo.textContent)
      .eql("1")
      .click(Selector("button.action-menu"))
      .click(Selector("button.action-private"))
      .expect(Selector("button.action-menu").exists, { timeout: 5000 })
      .notOk();
    await page.selectPhotoFromUID(ThirdVideo);
    await page.editSelected();
    await page.turnSwitchOn("private");
    await t.click(Selector(".action-close"));
    await page.clearSelection();
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await t.click(Selector("button.action-reload"));
    }
    await t
      .expect(Selector("div").withAttribute("data-uid", FirstVideo).exists, { timeout: 5000 })
      .notOk()
      .expect(Selector("div").withAttribute("data-uid", SecondVideo).exists, { timeout: 5000 })
      .notOk()
      .expect(Selector("div").withAttribute("data-uid", ThirdVideo).exists, { timeout: 5000 })
      .notOk();
    await page.openNav();
    await t.click(Selector(".nav-private"));

    await t
      .expect(Selector("div").withAttribute("data-uid", FirstPhoto).exists, { timeout: 5000 })
      .ok()
      .expect(Selector("div").withAttribute("data-uid", SecondPhoto).exists, { timeout: 5000 })
      .ok()
      .expect(Selector("div").withAttribute("data-uid", FirstVideo).exists, { timeout: 5000 })
      .ok()
      .expect(Selector("div").withAttribute("data-uid", SecondVideo).exists, { timeout: 5000 })
      .ok()
      .expect(Selector("div").withAttribute("data-uid", ThirdVideo).exists, { timeout: 5000 })
      .ok();
    await page.selectPhotoFromUID(FirstPhoto);
    await page.selectPhotoFromUID(SecondPhoto);
    await page.selectPhotoFromUID(FirstVideo);
    await t.click(Selector("button.action-menu")).click(Selector("button.action-private"));
    await page.selectPhotoFromUID(ThirdVideo);
    await page.editSelected();
    await page.turnSwitchOff("private");
    await t.click(Selector(".action-close"));
    await page.clearSelection();
    await page.setFilter("view", "List");
    await t
      .click(Selector("button.input-private").withAttribute("data-uid", ThirdPhoto))
      .click(Selector("button.input-private").withAttribute("data-uid", SecondVideo));
    await page.setFilter("view", "Mosaic");
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
      .notOk()
      .expect(Selector("div").withAttribute("data-uid", FirstVideo).exists, { timeout: 5000 })
      .notOk()
      .expect(Selector("div").withAttribute("data-uid", SecondVideo).exists, { timeout: 5000 })
      .notOk()
      .expect(Selector("div").withAttribute("data-uid", ThirdVideo).exists, { timeout: 5000 })
      .notOk();
    await page.openNav();
    await t.click(Selector(".nav-browse"));

    await t
      .expect(Selector("div").withAttribute("data-uid", FirstPhoto).exists, { timeout: 5000 })
      .ok()
      .expect(Selector("div").withAttribute("data-uid", SecondPhoto).exists, { timeout: 5000 })
      .ok()
      .expect(Selector("div").withAttribute("data-uid", ThirdPhoto).exists, { timeout: 5000 })
      .ok()
      .expect(Selector("div").withAttribute("data-uid", FirstVideo).exists, { timeout: 5000 })
      .ok()
      .expect(Selector("div").withAttribute("data-uid", SecondVideo).exists, { timeout: 5000 })
      .ok()
      .expect(Selector("div").withAttribute("data-uid", ThirdVideo).exists, { timeout: 5000 })
      .ok();
    await page.openNav();
    await t.click(Selector(".nav-video"));

    await t
      .expect(Selector("div").withAttribute("data-uid", FirstVideo).exists, { timeout: 5000 })
      .ok()
      .expect(Selector("div").withAttribute("data-uid", SecondVideo).exists, { timeout: 5000 })
      .ok()
      .expect(Selector("div").withAttribute("data-uid", ThirdVideo).exists, { timeout: 5000 })
      .ok();
  }
);

test.meta("testID", "photos-006")(
  "Archive/restore video, photos, private photos and review photos using clipboard",
  async (t) => {
    await page.search("photo:true");
    await page.setFilter("view", "Mosaic");
    const FirstPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    const SecondPhoto = await Selector("div.is-photo").nth(1).getAttribute("data-uid");
    await page.openNav();
    await t.click(Selector(".nav-video"));
    const FirstVideo = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.openNav();
    await t.click(Selector(".nav-private"));
    const FirstPrivatePhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.openNav();
    await t.click(Selector("div.nav-browse + div")).click(Selector(".nav-review"));
    const FirstReviewPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.openNav();
    await t.click(Selector(".nav-archive"));
    await t
      .expect(Selector("div").withAttribute("data-uid", FirstPhoto).exists, { timeout: 5000 })
      .notOk()
      .expect(Selector("div").withAttribute("data-uid", SecondPhoto).exists, { timeout: 5000 })
      .notOk()
      .expect(Selector("div").withAttribute("data-uid", FirstVideo).exists, { timeout: 5000 })
      .notOk()
      .expect(Selector("div").withAttribute("data-uid", FirstPrivatePhoto).exists, {
        timeout: 5000,
      })
      .notOk()
      .expect(Selector("div").withAttribute("data-uid", FirstReviewPhoto).exists, { timeout: 5000 })
      .notOk();
    await page.openNav();
    await t.click(Selector(".nav-video"));
    await page.setFilter("view", "Card");
    await page.selectPhotoFromUID(FirstVideo);
    const clipboardCountVideo = await Selector("span.count-clipboard", { timeout: 5000 });
    await t.expect(clipboardCountVideo.textContent).eql("1");
    await page.archiveSelected();
    await t.expect(Selector("button.action-menu").exists, { timeout: 5000 }).notOk();
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await t.click(Selector("button.action-reload"));
    }
    await t
      .expect(Selector("div").withAttribute("data-uid", FirstVideo).exists, { timeout: 5000 })
      .notOk();
    await page.openNav();
    await t.click(Selector(".nav-browse"));
    await page.selectPhotoFromUID(FirstPhoto);
    await page.selectPhotoFromUID(SecondPhoto);
    const clipboardCountPhotos = await Selector("span.count-clipboard", { timeout: 5000 });
    await t.expect(clipboardCountPhotos.textContent).eql("2");
    await page.archiveSelected();
    await t.expect(Selector("button.action-menu").exists, { timeout: 5000 }).notOk();
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await t.click(Selector("button.action-reload"));
    }
    await t
      .expect(Selector("div").withAttribute("data-uid", FirstPhoto).exists, { timeout: 5000 })
      .notOk()
      .expect(Selector("div").withAttribute("data-uid", SecondPhoto).exists, { timeout: 5000 })
      .notOk();
    await page.openNav();
    await t.click(Selector(".nav-private"));
    await page.selectPhotoFromUID(FirstPrivatePhoto);
    const clipboardCountPrivate = await Selector("span.count-clipboard", { timeout: 5000 });
    await t.expect(clipboardCountPrivate.textContent).eql("1");
    await page.openNav();
    if (t.browser.platform === "mobile") {
      await t.click(Selector(".nav-browse + div")).click(Selector(".nav-review"));
    } else {
      await t.click(Selector(".nav-review"));
    }
    await page.selectPhotoFromUID(FirstReviewPhoto);
    const clipboardCountReview = await Selector("span.count-clipboard", { timeout: 5000 });
    await t.expect(clipboardCountReview.textContent).eql("2");
    await page.archiveSelected();
    await t.expect(Selector("button.action-menu").exists, { timeout: 5000 }).notOk();
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await t.click(Selector("button.action-reload"));
    }
    await t
      .expect(Selector("div").withAttribute("data-uid", FirstReviewPhoto).exists, { timeout: 5000 })
      .notOk();
    await page.openNav();
    if (t.browser.platform === "mobile") {
      await t.click(Selector(".nav-browse + div")).click(Selector(".nav-archive"));
    } else {
      await t.click(Selector(".nav-archive"));
    }
    await t
      .expect(Selector("div").withAttribute("data-uid", FirstPhoto).exists, { timeout: 5000 })
      .ok()
      .expect(Selector("div").withAttribute("data-uid", SecondPhoto).exists, { timeout: 5000 })
      .ok()
      .expect(Selector("div").withAttribute("data-uid", FirstVideo).exists, { timeout: 5000 })
      .ok()
      .expect(Selector("div").withAttribute("data-uid", FirstPrivatePhoto).exists, {
        timeout: 5000,
      })
      .ok()
      .expect(Selector("div").withAttribute("data-uid", FirstReviewPhoto).exists, { timeout: 5000 })
      .ok();
    await page.selectPhotoFromUID(FirstPhoto);
    await page.selectPhotoFromUID(SecondPhoto);
    await page.selectPhotoFromUID(FirstVideo);
    await page.selectPhotoFromUID(FirstPrivatePhoto);
    await page.selectPhotoFromUID(FirstReviewPhoto);
    const clipboardCountArchive = await Selector("span.count-clipboard", { timeout: 5000 });
    await t.expect(clipboardCountArchive.textContent).eql("5");
    await page.restoreSelected();
    await t.expect(Selector("button.action-menu").exists, { timeout: 5000 }).notOk();
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
      .notOk()
      .expect(Selector("div").withAttribute("data-uid", FirstPrivatePhoto).exists, {
        timeout: 5000,
      })
      .notOk()
      .expect(Selector("div").withAttribute("data-uid", FirstReviewPhoto).exists, { timeout: 5000 })
      .notOk();
    await page.openNav();
    await t.click(Selector(".nav-video"));
    await t
      .expect(Selector("div").withAttribute("data-uid", FirstVideo).exists, { timeout: 5000 })
      .ok();
    await page.openNav();
    await t.click(Selector(".nav-browse"));
    await t
      .expect(Selector("div").withAttribute("data-uid", FirstPhoto).exists, { timeout: 5000 })
      .ok()
      .expect(Selector("div").withAttribute("data-uid", SecondPhoto).exists, { timeout: 5000 })
      .ok();
    await page.openNav();
    await t.click(Selector(".nav-private"));
    await t
      .expect(Selector("div").withAttribute("data-uid", FirstPrivatePhoto).exists, {
        timeout: 5000,
      })
      .ok();
    await page.openNav();
    if (t.browser.platform === "mobile") {
      await t.click(Selector(".nav-browse + div")).click(Selector(".nav-review"));
    } else {
      await t.click(Selector(".nav-review"));
    }
    await t
      .expect(Selector("div").withAttribute("data-uid", FirstReviewPhoto).exists, { timeout: 5000 })
      .ok();
  }
);

test.meta("testID", "photos-013")(
  "Check that archived files are not shown in monochrome/panoramas/stacks/scans/review/albums/favorites/private/videos/calendar/moments/states/labels/folders/originals",
  //TODO only select the not yet selected
  async (t) => {
    await page.openNav();
    await t.click(Selector(".nav-browse + div")).click(Selector(".nav-archive"));
    const InitialPhotoCountInArchive = await Selector("div.is-photo").count;
    await page.openNav();
    await t.click(Selector(".nav-monochrome"));
    const MonochromePhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(MonochromePhoto);
    await page.openNav();
    await t.click(Selector(".nav-panoramas"));
    const PanoramaPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(PanoramaPhoto);
    await page.openNav();
    await t.click(Selector(".nav-stacks"));
    const StackedPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(StackedPhoto);
    await page.openNav();
    await t.click(Selector(".nav-scans"));
    const ScannedPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(ScannedPhoto);
    await page.openNav();
    await t.click(Selector(".nav-review"));
    const ReviewPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(ReviewPhoto);
    await page.openNav();
    await t.click(Selector(".nav-favorites"));
    const FavoritesPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(FavoritesPhoto);
    await page.openNav();
    await t.click(Selector(".nav-private"));
    const PrivatePhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(PrivatePhoto);
    await page.openNav();
    await t.click(Selector(".nav-video"));
    const Video = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(Video);
    await page.openNav();
    await t.click(Selector(".nav-calendar"));
    await page.search("January 2017");
    await t.click(Selector("a.is-album").nth(0));
    const CalendarPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(CalendarPhoto);
    await page.openNav();
    await t.click(Selector(".nav-moments")).click(Selector("a.is-album").nth(0));
    const MomentPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(MomentPhoto);
    await page.openNav();
    await t.click(Selector(".nav-places + div")).click(Selector(".nav-states"));
    await page.search("Western Cape");
    await t.click(Selector("a.is-album").nth(0));
    const StatesPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(StatesPhoto);
    await page.openNav();
    await t.click(Selector(".nav-labels"));
    await page.search("Seashore");
    await t.click(Selector("a.is-label").nth(0));
    const LabelPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(LabelPhoto);
    await page.openNav();
    await t.click(Selector(".nav-folders"));
    await page.search("archive");
    await t.click(Selector("a.is-album").nth(0));
    const FolderPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(FolderPhoto);
    await page.archiveSelected();
    await page.openNav();
    await t.click(Selector(".nav-browse + div")).click(Selector(".nav-archive"));
    const PhotoCountInArchiveAfterArchive = await Selector("div.is-photo").count;
    await t.expect(PhotoCountInArchiveAfterArchive).eql(InitialPhotoCountInArchive + 13);
    await t.expect(Selector("button.action-menu").exists, { timeout: 5000 }).notOk();

    await page.openNav();
    await t
      .click(Selector(".nav-monochrome"))
      .expect(Selector("div").withAttribute("data-uid", MonochromePhoto).exists, { timeout: 5000 })
      .notOk();
    await page.openNav();
    await t
      .click(Selector(".nav-panoramas"))
      .expect(Selector("div").withAttribute("data-uid", PanoramaPhoto).exists, { timeout: 5000 })
      .notOk();
    await page.openNav();
    await t
      .click(Selector(".nav-stacks"))
      .expect(Selector("div").withAttribute("data-uid", StackedPhoto).exists, { timeout: 5000 })
      .notOk();
    await page.openNav();
    await t
      .click(Selector(".nav-scans"))
      .expect(Selector("div").withAttribute("data-uid", ScannedPhoto).exists, { timeout: 5000 })
      .notOk();
    await page.openNav();
    await t
      .click(Selector(".nav-review"))
      .expect(Selector("div").withAttribute("data-uid", ReviewPhoto).exists, { timeout: 5000 })
      .notOk();
    await page.openNav();
    await t
      .click(Selector(".nav-favorites"))
      .expect(Selector("div").withAttribute("data-uid", FavoritesPhoto).exists, { timeout: 5000 })
      .notOk();
    await page.openNav();
    await t
      .click(Selector(".nav-private"))
      .expect(Selector("div").withAttribute("data-uid", PrivatePhoto).exists, { timeout: 5000 })
      .notOk();
    await page.openNav();
    await t
      .click(Selector(".nav-video"))
      .expect(Selector("div").withAttribute("data-uid", Video).exists, { timeout: 5000 })
      .notOk();
    await t
      .navigateTo("/calendar/aqmxlr71p6zo22dk/january-2017")
      .expect(Selector("div").withAttribute("data-uid", CalendarPhoto).exists, { timeout: 5000 })
      .notOk();
    await page.openNav();
    await t
      .click(Selector(".nav-moments"))
      .click(Selector("a.is-album").nth(0))
      .expect(Selector("div").withAttribute("data-uid", MomentPhoto).exists, { timeout: 5000 })
      .notOk();
    await t
      .navigateTo("/states/aqmxlr71tebcohrw/western-cape-south-africa")
      .expect(Selector("div").withAttribute("data-uid", StatesPhoto).exists, { timeout: 5000 })
      .notOk()
      .navigateTo("/all?q=label%3Aseashore")
      .expect(Selector("div").withAttribute("data-uid", LabelPhoto).exists, { timeout: 5000 })
      .notOk()
      .navigateTo("/folders/aqnah1321mgkt1w2/archive")
      .expect(Selector("div").withAttribute("data-uid", FolderPhoto).exists, { timeout: 5000 })
      .notOk();

    await page.openNav();
    await t.click(Selector(".nav-browse + div")).click(Selector(".nav-archive"));
    await page.setFilter("view", "Cards");
    await page.selectPhotoFromUID(MonochromePhoto);
    await page.selectPhotoFromUID(PanoramaPhoto);
    await page.selectPhotoFromUID(StackedPhoto);
    await page.selectPhotoFromUID(ScannedPhoto);
    await page.selectPhotoFromUID(ReviewPhoto);
    await page.selectPhotoFromUID(FavoritesPhoto);
    await page.selectPhotoFromUID(PrivatePhoto);
    await page.selectPhotoFromUID(Video);
    await page.selectPhotoFromUID(CalendarPhoto);
    await page.selectPhotoFromUID(MomentPhoto);
    await page.selectPhotoFromUID(StatesPhoto);
    await page.selectPhotoFromUID(LabelPhoto);
    await page.selectPhotoFromUID(FolderPhoto);
    await page.restoreSelected();
    const PhotoCountInArchiveAfterRestore = await Selector("div.is-photo").count;
    await t.expect(PhotoCountInArchiveAfterRestore).eql(InitialPhotoCountInArchive);
    await t.expect(Selector("button.action-menu").exists, { timeout: 5000 }).notOk();

    await page.openNav();
    await t
      .click(Selector(".nav-monochrome"))
      .expect(Selector("div").withAttribute("data-uid", MonochromePhoto).exists, { timeout: 5000 })
      .ok();
    await page.openNav();
    await t
      .click(Selector(".nav-panoramas"))
      .expect(Selector("div").withAttribute("data-uid", PanoramaPhoto).exists, { timeout: 5000 })
      .ok();
    await page.openNav();
    await t
      .click(Selector(".nav-stacks"))
      .expect(Selector("div").withAttribute("data-uid", StackedPhoto).exists, { timeout: 5000 })
      .ok();
    await page.openNav();
    await t
      .click(Selector(".nav-scans"))
      .expect(Selector("div").withAttribute("data-uid", ScannedPhoto).exists, { timeout: 5000 })
      .ok();
    await page.openNav();
    await t
      .click(Selector(".nav-review"))
      .expect(Selector("div").withAttribute("data-uid", ReviewPhoto).exists, { timeout: 5000 })
      .ok();
    await page.openNav();
    await t
      .click(Selector(".nav-favorites"))
      .expect(Selector("div").withAttribute("data-uid", FavoritesPhoto).exists, { timeout: 5000 })
      .ok();
    await page.openNav();
    await t
      .click(Selector(".nav-private"))
      .expect(Selector("div").withAttribute("data-uid", PrivatePhoto).exists, { timeout: 5000 })
      .ok();
    await page.openNav();
    await t
      .click(Selector(".nav-video"))
      .expect(Selector("div").withAttribute("data-uid", Video).exists, { timeout: 5000 })
      .ok();
    await t
      .navigateTo("/calendar/aqmxlr71p6zo22dk/january-2017")
      .expect(Selector("div").withAttribute("data-uid", CalendarPhoto).exists, { timeout: 5000 })
      .ok();
    await page.openNav();
    await t
      .click(Selector(".nav-moments"))
      .click(Selector("a.is-album").nth(0))
      .expect(Selector("div").withAttribute("data-uid", MomentPhoto).exists, { timeout: 5000 })
      .ok();
    await t
      .navigateTo("/states/aqmxlr71tebcohrw/western-cape-south-africa")
      .expect(Selector("div").withAttribute("data-uid", StatesPhoto).exists, { timeout: 5000 })
      .ok()
      .navigateTo("/all?q=label%3Aseashore")
      .expect(Selector("div").withAttribute("data-uid", LabelPhoto).exists, { timeout: 5000 })
      .ok()
      .navigateTo("/folders/aqnah1321mgkt1w2/archive")
      .expect(Selector("div").withAttribute("data-uid", FolderPhoto).exists, { timeout: 5000 })
      .ok();
  }
);

test.meta("testID", "photos-014")(
  "Check that private files are not shown in monochrome/panoramas/stacks/scans/review/albums/favorites/archive/videos/calendar/moments/states/labels/folders/originals",
  //TODO only select the not yet selected
  //TODO should not appear in shared albums
  async (t) => {
    await page.openNav();
    await t.click(Selector(".nav-private"));
    const InitialPhotoCountInPrivate = await Selector("div.is-photo").count;
    await page.openNav();
    await t.click(Selector(".nav-browse + div")).click(Selector(".nav-monochrome"));
    const MonochromePhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(MonochromePhoto);
    await page.openNav();
    await t.click(Selector(".nav-panoramas"));
    const PanoramaPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(PanoramaPhoto);
    await page.openNav();
    await t.click(Selector(".nav-stacks"));
    const StackedPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(StackedPhoto);
    await page.openNav();
    await t.click(Selector(".nav-scans"));
    const ScannedPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(ScannedPhoto);
    await page.openNav();
    await t.click(Selector(".nav-review"));
    const ReviewPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(ReviewPhoto);
    await page.openNav();
    await t.click(Selector(".nav-albums"));
    await page.search("Holiday");
    await t.click(Selector("a.is-album").nth(0));
    const AlbumPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(AlbumPhoto);
    await page.openNav();
    await t.click(Selector(".nav-favorites"));
    const FavoritesPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(FavoritesPhoto);
    await page.openNav();
    await t.click(Selector(".nav-video"));
    const Video = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(Video);
    await page.openNav();
    await t.click(Selector(".nav-calendar"));
    await page.search("January 2017");
    await t.click(Selector("a.is-album").nth(0));
    const CalendarPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(CalendarPhoto);
    await page.openNav();
    await t.click(Selector(".nav-moments")).click(Selector("a.is-album").nth(0));
    const MomentPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(MomentPhoto);
    await page.openNav();
    await t.click(Selector(".nav-places + div")).click(Selector(".nav-states"));
    await page.search("Western Cape");
    await t.click(Selector("a.is-album").nth(0));
    const StatesPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(StatesPhoto);
    await page.openNav();
    await t.click(Selector(".nav-labels"));
    await page.search("Seashore");
    await t.click(Selector("a.is-label").nth(0));
    const LabelPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(LabelPhoto);
    await page.openNav();
    await t.click(Selector(".nav-folders"));
    await page.search("archive");
    await t.click(Selector("a.is-album").nth(0));
    const FolderPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(FolderPhoto);
    await page.privateSelected();
    await page.openNav();
    await t.click(Selector(".nav-private"));
    const PhotoCountInPrivateAfterPrivate = await Selector("div.is-photo").count;
    await t.expect(PhotoCountInPrivateAfterPrivate).eql(InitialPhotoCountInPrivate + 13);
    await t.expect(Selector("button.action-menu").exists, { timeout: 5000 }).notOk();

    await page.openNav();
    await t
      .click(Selector(".nav-browse + div"))
      .click(Selector(".nav-monochrome"))
      .expect(Selector("div").withAttribute("data-uid", MonochromePhoto).exists, { timeout: 5000 })
      .notOk();
    await page.openNav();
    await t
      .click(Selector(".nav-panoramas"))
      .expect(Selector("div").withAttribute("data-uid", PanoramaPhoto).exists, { timeout: 5000 })
      .notOk();
    await page.openNav();
    await t
      .click(Selector(".nav-stacks"))
      .expect(Selector("div").withAttribute("data-uid", StackedPhoto).exists, { timeout: 5000 })
      .notOk();
    await page.openNav();
    await t
      .click(Selector(".nav-scans"))
      .expect(Selector("div").withAttribute("data-uid", ScannedPhoto).exists, { timeout: 5000 })
      .notOk();
    await page.openNav();
    await t
      .click(Selector(".nav-review"))
      .expect(Selector("div").withAttribute("data-uid", ReviewPhoto).exists, { timeout: 5000 })
      .notOk()
      .navigateTo("/albums?q=Holiday")
      .click(Selector("a.is-album").nth(0))
      .expect(Selector("div").withAttribute("data-uid", AlbumPhoto).exists, { timeout: 5000 })
      .ok();
    await page.openNav();
    await t
      .click(Selector(".nav-favorites"))
      .expect(Selector("div").withAttribute("data-uid", FavoritesPhoto).exists, { timeout: 5000 })
      .notOk();
    await page.openNav();
    await t
      .click(Selector(".nav-video"))
      .expect(Selector("div").withAttribute("data-uid", Video).exists, { timeout: 5000 })
      .notOk();
    await t
      .navigateTo("/calendar/aqmxlr71p6zo22dk/january-2017")
      .expect(Selector("div").withAttribute("data-uid", CalendarPhoto).exists, { timeout: 5000 })
      .notOk();
    await page.openNav();
    await t
      .click(Selector(".nav-moments"))
      .click(Selector("a.is-album").nth(0))
      .expect(Selector("div").withAttribute("data-uid", MomentPhoto).exists, { timeout: 5000 })
      .notOk();
    await t
      .navigateTo("/states/aqmxlr71tebcohrw/western-cape-south-africa")
      .expect(Selector("div").withAttribute("data-uid", StatesPhoto).exists, { timeout: 5000 })
      .notOk()
      .navigateTo("/all?q=label%3Aseashore")
      .expect(Selector("div").withAttribute("data-uid", LabelPhoto).exists, { timeout: 5000 })
      .notOk()
      .navigateTo("/folders/aqnah1321mgkt1w2/archive")
      .expect(Selector("div").withAttribute("data-uid", FolderPhoto).exists, { timeout: 5000 })
      .notOk();

    await page.openNav();
    await t.click(Selector(".nav-private"));
    await page.setFilter("view", "Cards");
    await page.selectPhotoFromUID(MonochromePhoto);
    await page.selectPhotoFromUID(PanoramaPhoto);
    await page.selectPhotoFromUID(StackedPhoto);
    await page.selectPhotoFromUID(ScannedPhoto);
    await page.selectPhotoFromUID(ReviewPhoto);
    await page.selectPhotoFromUID(AlbumPhoto);
    await page.selectPhotoFromUID(FavoritesPhoto);
    await page.selectPhotoFromUID(Video);
    await page.selectPhotoFromUID(CalendarPhoto);
    await page.selectPhotoFromUID(MomentPhoto);
    await page.selectPhotoFromUID(StatesPhoto);
    await page.selectPhotoFromUID(LabelPhoto);
    await page.selectPhotoFromUID(FolderPhoto);
    await page.privateSelected();
    await page.openNav();
    await t.click(Selector(".nav-favorites"));
    await page.openNav();
    await t.click(Selector(".nav-private"));
    const PhotoCountInPrivateAfterUnprivate = await Selector("div.is-photo").count;
    await t.expect(PhotoCountInPrivateAfterUnprivate).eql(InitialPhotoCountInPrivate);
    await t.expect(Selector("button.action-menu").exists, { timeout: 5000 }).notOk();

    await page.openNav();
    await t
      .click(Selector(".nav-browse + div"))
      .click(Selector(".nav-monochrome"))
      .expect(Selector("div").withAttribute("data-uid", MonochromePhoto).exists, { timeout: 5000 })
      .ok();
    await page.openNav();
    await t
      .click(Selector(".nav-panoramas"))
      .expect(Selector("div").withAttribute("data-uid", PanoramaPhoto).exists, { timeout: 5000 })
      .ok();
    await page.openNav();
    await t
      .click(Selector(".nav-stacks"))
      .expect(Selector("div").withAttribute("data-uid", StackedPhoto).exists, { timeout: 5000 })
      .ok();
    await page.openNav();
    await t
      .click(Selector(".nav-scans"))
      .expect(Selector("div").withAttribute("data-uid", ScannedPhoto).exists, { timeout: 5000 })
      .ok();
    await page.openNav();
    await t
      .click(Selector(".nav-review"))
      .expect(Selector("div").withAttribute("data-uid", ReviewPhoto).exists, { timeout: 5000 })
      .ok()
      .navigateTo("/albums?q=Holiday")
      .click(Selector("a.is-album").nth(0))
      .expect(Selector("div").withAttribute("data-uid", AlbumPhoto).exists, { timeout: 5000 })
      .ok();
    await page.openNav();
    await t
      .click(Selector(".nav-favorites"))
      .expect(Selector("div").withAttribute("data-uid", FavoritesPhoto).exists, { timeout: 5000 })
      .ok();
    await page.openNav();
    await t
      .click(Selector(".nav-video"))
      .expect(Selector("div").withAttribute("data-uid", Video).exists, { timeout: 5000 })
      .ok();
    await t
      .navigateTo("/calendar/aqmxlr71p6zo22dk/january-2017")
      .expect(Selector("div").withAttribute("data-uid", CalendarPhoto).exists, { timeout: 5000 })
      .ok();
    await page.openNav();
    await t
      .click(Selector(".nav-moments"))
      .click(Selector("a.is-album").nth(0))
      .expect(Selector("div").withAttribute("data-uid", MomentPhoto).exists, { timeout: 5000 })
      .ok();
    await t
      .navigateTo("/states/aqmxlr71tebcohrw/western-cape-south-africa")
      .expect(Selector("div").withAttribute("data-uid", StatesPhoto).exists, { timeout: 5000 })
      .ok()
      .navigateTo("/all?q=label%3Aseashore")
      .expect(Selector("div").withAttribute("data-uid", LabelPhoto).exists, { timeout: 5000 })
      .ok()
      .navigateTo("/folders/aqnah1321mgkt1w2/archive")
      .expect(Selector("div").withAttribute("data-uid", FolderPhoto).exists, { timeout: 5000 })
      .ok();
  }
);
