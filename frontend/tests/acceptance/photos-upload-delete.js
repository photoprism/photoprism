import { Selector } from "testcafe";
import testcafeconfig from "./testcafeconfig";
import Page from "./page-model";
import { ClientFunction } from "testcafe";
import fs from "fs";

fixture`Test photos upload and delete`.page`${testcafeconfig.url}`;

const page = new Page();

test.meta("testID", "photos-upload-delete-001")("Upload + Delete jpg/json", async (t) => {
  await t.expect(fs.existsSync("../storage/acceptance/originals/2020/10")).notOk();
  await page.search("digikam");
  const PhotoCount = await Selector("div.is-photo").count;
  await t
    .expect(PhotoCount)
    .eql(0)
    .click(Selector(".action-upload"))
    .setFilesToUpload(Selector(".input-upload"), [
      "./upload-files/digikam.jpg",
      "./upload-files/digikam.json",
    ])
    .wait(15000);
  const PhotoCountAfterUpload = await Selector("div.is-photo").count;
  await t.expect(PhotoCountAfterUpload).eql(1);
  const UploadedPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
  await t.navigateTo("/library/files/2020/10");
  const FileCount = await Selector("div.is-file").count;
  await t.expect(FileCount).eql(2);
  await page.openNav();
  await t.click(Selector(".nav-browse"));
  await page.search("digikam");
  await page.selectPhotoFromUID(UploadedPhoto);
  await page.editSelected();
  await t
    .click("#tab-files")
    .expect(Selector("div.caption").withText(".json").visible)
    .ok()
    .expect(Selector("div.caption").withText(".jpg").visible)
    .ok()
    .click(Selector(".action-close"));
  await page.clearSelection();
  if (t.browser.platform !== "mobile") {
    await t.expect(fs.existsSync("../storage/acceptance/originals/2020/10")).ok();
    const originalsLength = fs.readdirSync("../storage/acceptance/originals/2020/10").length;
    await t.expect(originalsLength).eql(2);
  }
  await page.deletePhotoFromUID(UploadedPhoto);
  await page.openNav();
  await t.click(Selector(".nav-browse"));
  await page.search("digikam");
  await t
    .expect(Selector("div").withAttribute("data-uid", UploadedPhoto).exists, { timeout: 5000 })
    .notOk()
    .navigateTo("/library/files/2020/10");
  const FileCountAfterDelete = await Selector("div.is-file").count;
  await t.expect(FileCountAfterDelete).eql(0);
  if (t.browser.platform !== "mobile") {
    const originalsLengthAfterDelete = fs.readdirSync(
      "../storage/acceptance/originals/2020/10"
    ).length;
    await t.expect(originalsLengthAfterDelete).eql(0);
  }
});

test.meta("testID", "photos-upload-delete-002")("Upload + Delete video", async (t) => {
  await t.expect(fs.existsSync("../storage/acceptance/originals/2020/06")).notOk();
  await page.search("korn");
  const PhotoCount = await Selector("div.is-photo").count;
  await t
    .expect(PhotoCount)
    .eql(0)
    .click(Selector(".action-upload"))
    .setFilesToUpload(Selector(".input-upload"), ["./upload-files/korn.mp4"])
    .wait(15000);
  const PhotoCountAfterUpload = await Selector("div.is-photo").count;
  await t.expect(PhotoCountAfterUpload).eql(1);
  const UploadedPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
  await t.navigateTo("/library/files/2020/06");
  const FileCount = await Selector("div.is-file").count;
  await t.expect(FileCount).eql(1);
  await page.openNav();
  await t.click(Selector(".nav-browse"));
  await page.search("korn");
  await page.selectPhotoFromUID(UploadedPhoto);
  await page.editSelected();
  await t
    .click("#tab-files")
    .expect(Selector("div.caption").withText(".mp4").visible)
    .ok()
    .expect(Selector("div.caption").withText(".jpg").visible)
    .ok()
    .click(Selector(".action-close"));
  await page.clearSelection();
  if (t.browser.platform !== "mobile") {
    await t.expect(fs.existsSync("../storage/acceptance/originals/2020/06")).ok();
    const originalsLength = fs.readdirSync("../storage/acceptance/originals/2020/06").length;
    await t.expect(originalsLength).eql(1);
    const sidecarLength = fs.readdirSync("../storage/acceptance/originals/2020/06").length;
    await t.expect(sidecarLength).eql(1);
  }
  await page.deletePhotoFromUID(UploadedPhoto);
  await page.openNav();
  await t.click(Selector(".nav-browse"));
  await page.search("korn");
  await t
    .expect(Selector("div").withAttribute("data-uid", UploadedPhoto).exists, { timeout: 5000 })
    .notOk()
    .navigateTo("/library/files/2020/06");
  const FileCountAfterDelete = await Selector("div.is-file").count;
  await t.expect(FileCountAfterDelete).eql(0);
  if (t.browser.platform !== "mobile") {
    const originalsLengthAfterDelete = fs.readdirSync(
      "../storage/acceptance/originals/2020/06"
    ).length;
    await t.expect(originalsLengthAfterDelete).eql(0);
    const sidecarLengthAfterDelete = fs.readdirSync(
      "../storage/acceptance/originals/2020/06"
    ).length;
    await t.expect(sidecarLengthAfterDelete).eql(0);
  }
});

test.meta("testID", "photos-upload-delete-003")("Upload to existing Album + Delete", async (t) => {
  await page.openNav();
  await t.click(Selector(".nav-albums"));
  await page.search("Christmas");
  const AlbumUid = await Selector("a.is-album").nth(0).getAttribute("data-uid");
  await t.click(Selector("a.is-album").withAttribute("data-uid", AlbumUid));
  const PhotoCount = await Selector("div.is-photo").count;
  await t
    .click(Selector(".action-upload"))
    .click(Selector(".input-albums"))
    .click(Selector("div.v-list__tile__title").withText("Christmas"))
    .setFilesToUpload(Selector(".input-upload"), ["./upload-files/ladybug.jpg"])
    .wait(15000);
  const PhotoCountAfterUpload = await Selector("div.is-photo").count;
  await t.expect(PhotoCountAfterUpload).eql(PhotoCount + 1);
  await page.openNav();
  await t.click(Selector(".nav-browse"));
  await page.search("ladybug");
  const UploadedPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
  await page.deletePhotoFromUID(UploadedPhoto);
  await page.openNav();
  await t.click(Selector(".nav-browse"));
  await page.search("ladybug");
  await t
    .expect(Selector("div").withAttribute("data-uid", UploadedPhoto).exists, { timeout: 5000 })
    .notOk();
  await page.openNav();
  await t.click(Selector(".nav-albums"));
  await t
    .click(Selector("a.is-album").withAttribute("data-uid", AlbumUid))
    .expect(Selector("div").withAttribute("data-uid", UploadedPhoto).exists, { timeout: 5000 })
    .notOk();
  const PhotoCountAfterDelete = await Selector("div.is-photo").count;
  await t.expect(PhotoCountAfterDelete).eql(PhotoCount);
});

test.meta("testID", "photos-upload-delete-004")("Upload jpg to new Album + Delete", async (t) => {
  await page.openNav();
  await t.click(Selector(".nav-albums"));
  const AlbumCount = await Selector("a.is-album").count;
  await t
    .click(Selector(".action-upload", { timeout: 5000 }))
    .click(Selector(".input-albums"))
    .typeText(Selector(".input-albums input"), "NewCreatedAlbum")
    .pressKey("enter")
    .setFilesToUpload(Selector(".input-upload"), ["./upload-files/digikam.jpg"])
    .wait(15000);
  if (t.browser.platform === "mobile") {
    await t.eval(() => location.reload());
  } else {
    await t.click(Selector("button.action-reload"));
  }
  const AlbumCountAfterUpload = await Selector("a.is-album").count;
  await t.expect(AlbumCountAfterUpload).eql(AlbumCount + 1);
  await page.search("NewCreatedAlbum");
  await t.click(Selector("a.is-album").nth(0));
  const PhotoCount = await Selector("div.is-photo").count;
  await t.expect(PhotoCount).eql(1);
  await page.openNav();
  await t.click(Selector(".nav-browse"));
  await page.search("digikam");
  const UploadedPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
  await page.deletePhotoFromUID(UploadedPhoto);
  await page.openNav();
  await t.click(Selector(".nav-browse"));
  await page.search("digikam");
  await t
    .expect(Selector("div").withAttribute("data-uid", UploadedPhoto).exists, { timeout: 5000 })
    .notOk();
  await page.openNav();
  await t.click(Selector(".nav-albums"));
  await page.search("NewCreatedAlbum");
  await t
    .click(Selector("a.is-album").nth(0))
    .expect(Selector("div").withAttribute("data-uid", UploadedPhoto).exists, { timeout: 5000 })
    .notOk();
  const PhotoCountAfterDelete = await Selector("div.is-photo").count;
  await t.expect(PhotoCountAfterDelete).eql(0);
  await page.openNav();
  await t.click(Selector(".nav-albums"));
  await page.search("NewCreatedAlbum");
  await t.hover(Selector("a.is-album").nth(0)).click(Selector("a.is-album .input-select").nth(0));
  await page.deleteSelected();
});

test.meta("testID", "photos-upload-delete-005")("Try uploading nsfw file", async (t) => {
  await t
    .click(Selector(".action-upload"))
    .setFilesToUpload(Selector(".input-upload"), ["./upload-files/hentai_2.jpg"])
    .wait(15000);
  await page.openNav();
  await t
    .click(Selector(".nav-library"))
    .click(Selector("#tab-library-logs"))
    .expect(Selector("p").withText("hentai_2.jpg might be offensive").visible)
    .ok();
});

test.meta("testID", "photos-upload-delete-006")("Try uploading txt file", async (t) => {
  await t
    .click(Selector(".action-upload"))
    .setFilesToUpload(Selector(".input-upload"), ["./upload-files/foo.txt"])
    .wait(15000);
  await page.openNav();
  await t
    .click(Selector(".nav-library"))
    .click(Selector("#tab-library-logs"))
    .expect(Selector("p").withText(" foo.txt is not a jpeg file").visible)
    .ok();
});
