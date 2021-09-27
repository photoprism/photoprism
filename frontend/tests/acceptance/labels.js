import { Selector } from "testcafe";
import testcafeconfig from "./testcafeconfig";
import Page from "./page-model";

fixture`Test labels`.page`${testcafeconfig.url}`;

const page = new Page();

test.meta("testID", "labels-001")("Remove/Activate Add/Delete Label from photo", async (t) => {
  await page.openNav();
  await t.click(Selector(".nav-labels"));
  const countImportantLabels = await Selector("a.is-label").count;
  await t.click(Selector("button.action-show-all"));
  const countAllLabels = await Selector("a.is-label").count;
  await t
    .expect(countAllLabels)
    .gt(countImportantLabels)
    .click(Selector("button.action-show-important"));
  await page.search("beacon");
  const LabelBeacon = await Selector("a.is-label").nth(0).getAttribute("data-uid");
  await t.click(Selector("a.is-label").withAttribute("data-uid", LabelBeacon));
  await page.setFilter("view", "Cards");
  const PhotoBeacon = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
  await t.click(Selector(".action-title-edit").withAttribute("data-uid", PhotoBeacon));
  const PhotoKeywords = await Selector(".input-keywords textarea").value;
  await t
    .expect(PhotoKeywords)
    .contains("beacon")
    .click(Selector("#tab-labels"))
    .click(Selector("button.action-remove"), { timeout: 5000 })
    .typeText(Selector(".input-label input"), "Test")
    .click(Selector("button.p-photo-label-add"))
    .click(Selector("#tab-details"));
  const PhotoKeywordsAfterEdit = await Selector(".input-keywords textarea").value;
  await t
    .expect(PhotoKeywordsAfterEdit)
    .contains("test")
    .expect(PhotoKeywordsAfterEdit)
    .notContains("beacon")
    .click(Selector(".action-close"));
  await page.openNav();
  await t.click(Selector(".nav-labels"));
  await page.search("beacon");
  await t.expect(Selector("div.no-results").visible).ok();
  await page.search("test");
  const LabelTest = await Selector("a.is-label").nth(0).getAttribute("data-uid");
  await t
    .click(Selector("a.is-label").withAttribute("data-uid", LabelTest))
    .click(Selector(".action-title-edit").withAttribute("data-uid", PhotoBeacon))
    .click(Selector("#tab-labels"))
    .click(Selector(".action-delete"), { timeout: 5000 })
    .click(Selector(".action-on"))
    .click(Selector("#tab-details"));
  const PhotoKeywordsAfterUndo = await Selector(".input-keywords textarea").value;
  await t
    .expect(PhotoKeywordsAfterUndo)
    .contains("beacon")
    .expect(PhotoKeywordsAfterUndo)
    .notContains("test")
    .click(Selector(".action-close"));
  await page.openNav();
  await t.click(Selector(".nav-labels"));
  await page.search("test");
  await t.expect(Selector("div.no-results").visible).ok();
  await page.search("beacon");
  await t.expect(Selector("a").withAttribute("data-uid", LabelBeacon).visible).ok();
});

test.meta("testID", "labels-002")("Rename Label", async (t) => {
  await page.openNav();
  await t.click(Selector(".nav-labels"));
  await page.search("zebra");
  const LabelZebra = await Selector("a.is-label").nth(0).getAttribute("data-uid");
  await t.click(Selector("a.is-label").nth(0));
  const FirstPhotoZebra = await Selector("div.is-photo", { timeout: 5000 })
    .nth(0)
    .getAttribute("data-uid");
  const SecondPhotoZebra = await Selector("div.is-photo", { timeout: 5000 })
    .nth(1)
    .getAttribute("data-uid");
  await page.setFilter("view", "Cards");
  await t.click(Selector(".action-title-edit").withAttribute("data-uid", FirstPhotoZebra));
  const FirstPhotoTitle = await Selector(".input-title input", { timeout: 5000 }).value;
  const FirstPhotoKeywords = await Selector(".input-keywords textarea", { timeout: 5000 }).value;
  await t
    .expect(FirstPhotoTitle)
    .contains("Zebra")
    .expect(FirstPhotoKeywords)
    .contains("zebra")
    .click(Selector("#tab-labels"))
    .click(Selector("div.p-inline-edit"))
    .typeText(Selector(".input-rename input"), "Horse", { replace: true })
    .pressKey("enter")
    .click(Selector("#tab-details"));
  const FirstPhotoTitleAfterEdit = await Selector(".input-title input", { timeout: 5000 }).value;
  const FirstPhotoKeywordsAfterEdit = await Selector(".input-keywords textarea", { timeout: 5000 })
    .value;
  await t
    .expect(FirstPhotoTitleAfterEdit)
    .contains("Horse")
    .expect(FirstPhotoKeywordsAfterEdit)
    .contains("horse")
    .expect(FirstPhotoTitleAfterEdit)
    .notContains("Zebra")
    .click(Selector(".action-close"));
  await page.openNav();
  await t.click(Selector(".nav-labels"));
  await page.search("horse");
  await t
    .expect(Selector("a").withAttribute("data-uid", LabelZebra).visible)
    .ok()
    .click(Selector("a.is-label").withAttribute("data-uid", LabelZebra))
    .expect(Selector("div").withAttribute("data-uid", SecondPhotoZebra).visible)
    .ok()
    .click(Selector(".action-title-edit").withAttribute("data-uid", FirstPhotoZebra))
    .click(Selector("#tab-labels"))
    .click(Selector("div.p-inline-edit"))
    .typeText(Selector(".input-rename input"), "Zebra", { replace: true })
    .pressKey("enter")
    .click(Selector(".action-close"));
  await page.openNav();
  await t.click(Selector(".nav-labels"));
  await page.search("horse");
  await t.expect(Selector("div.no-results").visible).ok();
});

test.meta("testID", "labels-003")("Add label to album", async (t) => {
  await page.openNav();
  await t.click(Selector(".nav-albums"));
  await page.search("Christmas");
  const AlbumUid = await Selector("a.is-album").nth(0).getAttribute("data-uid");
  await t.click(Selector("a.is-album").withAttribute("data-uid", AlbumUid));
  const PhotoCount = await Selector("div.is-photo").count;
  await page.openNav();
  await t.click(Selector(".nav-labels"));
  await page.search("landscape");
  const LabelLandscape = await Selector("a.is-label").nth(1).getAttribute("data-uid");
  await t.click(Selector("a.is-label").withAttribute("data-uid", LabelLandscape));
  const FirstPhotoLandscape = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
  const SecondPhotoLandscape = await Selector("div.is-photo").nth(1).getAttribute("data-uid");
  const ThirdPhotoLandscape = await Selector("div.is-photo").nth(2).getAttribute("data-uid");
  const FourthPhotoLandscape = await Selector("div.is-photo").nth(3).getAttribute("data-uid");
  const FifthPhotoLandscape = await Selector("div.is-photo").nth(4).getAttribute("data-uid");
  const SixthPhotoLandscape = await Selector("div.is-photo").nth(5).getAttribute("data-uid");
  await page.openNav();
  await t.click(".nav-labels");
  await page.selectFromUID(LabelLandscape);

  const clipboardCount = await Selector("span.count-clipboard");
  await t.expect(clipboardCount.textContent).eql("1");
  await page.addSelectedToAlbum("Christmas", "album");
  await page.openNav();
  await t
    .click(Selector(".nav-albums"))
    .click(Selector("a.is-album").withAttribute("data-uid", AlbumUid));
  const PhotoCountAfterAdd = await Selector("div.is-photo", { timeout: 5000 }).count;
  await t.expect(PhotoCountAfterAdd).eql(PhotoCount + 6);
  await page.selectPhotoFromUID(FirstPhotoLandscape);
  await page.selectPhotoFromUID(SecondPhotoLandscape);
  await page.selectPhotoFromUID(ThirdPhotoLandscape);
  await page.selectPhotoFromUID(FourthPhotoLandscape);
  await page.selectPhotoFromUID(FifthPhotoLandscape);
  await page.selectPhotoFromUID(SixthPhotoLandscape);
  await page.removeSelected();
  const PhotoCountAfterDelete = await Selector("div.is-photo", { timeout: 5000 }).count;
  await t.expect(PhotoCountAfterDelete).eql(PhotoCountAfterAdd - 6);
});

test.meta("testID", "labels-004")("Delete label", async (t) => {
  await page.openNav();
  await t.click(Selector(".nav-labels"));
  await page.search("dome");
  const LabelDome = await Selector("a.is-label", { timeout: 5000 }).nth(0).getAttribute("data-uid");
  await t.click(Selector("a.is-label").withAttribute("data-uid", LabelDome));
  const FirstPhotoDome = await Selector("div.is-photo", { timeout: 5000 })
    .nth(0)
    .getAttribute("data-uid");
  await page.openNav();
  await t.click(".nav-labels");
  await page.selectFromUID(LabelDome);
  const clipboardCount = await Selector("span.count-clipboard", { timeout: 5000 });
  await t.expect(clipboardCount.textContent).eql("1");
  await page.deleteSelected();
  await page.search("dome");
  await t.expect(Selector("div.no-results").visible).ok();
  await page.openNav();
  await t.click(".nav-browse");
  await page.setFilter("view", "Cards");
  await t
    .click(Selector(".action-title-edit").withAttribute("data-uid", FirstPhotoDome))
    .click(Selector("#tab-labels"))
    .expect(Selector("td").withText("No labels found").visible)
    .ok()
    .typeText(Selector(".input-label input"), "Dome")
    .click(Selector("button.p-photo-label-add"));
});
