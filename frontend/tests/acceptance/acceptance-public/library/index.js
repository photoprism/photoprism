import { Selector } from "testcafe";
import testcafeconfig from "../../../testcafeconfig.json";
import Menu from "../../page-model/menu";
import Toolbar from "../../page-model/toolbar";
import Photo from "../../page-model/photo";
import Page from "../../page-model/page";
import Library from "../../page-model/library";
import Album from "../../page-model/album";

fixture`Test index`.page`${testcafeconfig.url}`;

const menu = new Menu();
const toolbar = new Toolbar();
const photo = new Photo();
const page = new Page();
const library = new Library();
const album = new Album();

test.meta("testID", "library-index-001").meta({ type: "short", mode: "public" })(
  "Common: Index files from folder",
  async (t) => {
    await menu.openPage("labels");
    await toolbar.search("cheetah");

    await t.expect(Selector("div.no-results").visible).ok();

    await menu.openPage("moments");
    const MomentCount = await album.getAlbumCount("all");
    await menu.openPage("calendar");
    if (t.browser.platform === "mobile") {
      await t.navigateTo("/library/calendar?q=December%202013");
    } else {
      await toolbar.search("December 2013");
    }

    await t.expect(Selector("div.no-results").visible).ok();

    await menu.openPage("folders");
    if (t.browser.platform === "mobile") {
      await t.navigateTo("/library/folders?q=moment");
    } else {
      await toolbar.search("Moment");
    }

    await t.expect(Selector("div.no-results").visible).ok();

    await menu.openPage("states");
    if (t.browser.platform === "mobile") {
      await t.navigateTo("/library/states?q=KwaZulu");
    } else {
      await toolbar.search("KwaZulu");
    }

    await t.expect(Selector("div.no-results").visible).ok();

    await menu.openPage("originals");
    await t.click(Selector(".is-folder").withText("moment"));

    await t.expect(Selector("div.no-results").visible).ok();

    await menu.openPage("monochrome");
    const MonochromeCount = await photo.getPhotoCount("all");
    await menu.openPage("library");
    await t
      .click(library.indexTab)
      .click(library.indexFolderSelect)
      .click(page.selectOption.withText("/moment"))
      .click(library.index)
      //TODO replace wait
      .wait(50000);

    await t.expect(Selector("span").withText("Done.").visible, { timeout: 60000 }).ok();

    await menu.openPage("labels");
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await toolbar.triggerToolbarAction("reload");
    }
    await toolbar.search("cheetah");

    await t.expect(Selector(".is-label").visible).ok();

    await menu.openPage("moments");
    const MomentCountAfterIndex = await album.getAlbumCount("all");

    await t
      .expect(MomentCountAfterIndex)
      .gt(MomentCount)
      .click(Selector("a").withText("South Africa 2013"))
      .expect(Selector(".is-photo").visible)
      .ok();

    await menu.openPage("calendar");
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await toolbar.triggerToolbarAction("reload");
    }
    if (t.browser.platform === "mobile") {
      await t.navigateTo("/library/calendar?q=December%202013");
    } else {
      await toolbar.search("December 2013");
    }

    await t.expect(Selector(".is-album").visible).ok();

    await menu.openPage("folders");
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await toolbar.triggerToolbarAction("reload");
    }
    if (t.browser.platform === "mobile") {
      await t.navigateTo("/library/folders?q=moment");
    } else {
      await toolbar.search("Moment");
    }

    await t.expect(Selector(".is-album", { timeout: 15000 }).visible).ok();

    await menu.openPage("states");
    if (t.browser.platform === "mobile") {
      await t.navigateTo("/library/states?q=KwaZulu");
    } else {
      await toolbar.search("KwaZulu");
    }

    await t.expect(Selector(".is-album").visible).ok();

    await menu.openPage("originals");

    await t.expect(Selector(".is-folder").withText("moment").visible, { timeout: 60000 }).ok();

    await menu.openPage("monochrome");
    const MonochromeCountAfterIndex = await photo.getPhotoCount("all");

    await t.expect(MonochromeCountAfterIndex).gt(MonochromeCount);
  }
);
