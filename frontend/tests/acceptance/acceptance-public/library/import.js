import { Selector } from "testcafe";
import testcafeconfig from "../../../testcafeconfig.json";
import Menu from "../../page-model/menu";
import Toolbar from "../../page-model/toolbar";
import Page from "../../page-model/page";
import Library from "../../page-model/library";
import Notifies from "../../page-model/notifications";

fixture`Import file from folder`.page`${testcafeconfig.url}`;

const menu = new Menu();
const toolbar = new Toolbar();
const page = new Page();
const library = new Library();
const notifies = new Notifies();

test.meta("testID", "library-import-001").meta({ type: "short", mode: "public" })(
  "Common: Import files from folder using copy",
  async (t) => {
    await menu.openPage("labels");
    await toolbar.search("bakery");

    await t.expect(Selector("div.no-results").visible).ok();

    await menu.openPage("library");
    await t
      .click(library.importTab)
      .click(library.openImportFolderSelect)
      .wait(9000)
      .typeText(library.openImportFolderSelect, "/Bäcke", { replace: true })
      .click(page.selectOption.nth(0))
      .click(library.import);
    await notifies.waitForImport(60000);
    await menu.openPage("labels");
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await toolbar.triggerToolbarAction("refresh");
    }
    await toolbar.search("bakery");

    await t.expect(Selector(".is-label").visible).ok();
  }
);


test.meta("testID", "library-import-002").meta({ type: "short", mode: "public" })(
  "Common: Import files from folder with album",
  async (t) => {
    await menu.openPage("albums");
    await toolbar.search("Beatles");

    await t.expect(Selector("div.no-results").visible).ok();

    await menu.openPage("library");
    await t
      .click(library.importTab)
      .click(library.openImportFolderSelect)
      .wait(9000)
      .typeText(library.openImportFolderSelect, "/Käfer", { replace: true })
      .click(page.selectOption.nth(0))
      .click(library.albumInput)
      .click(Selector('.v-list-item-title').withText('Garden'))
      .typeText(library.albumInput, "Beatles", { replace: true })
      .pressKey('enter')
      .click(library.import);
    await notifies.waitForImport(60000);
    await menu.openPage("albums");
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await toolbar.triggerToolbarAction("refresh");
    }
    await toolbar.search("Beatles");
    await t.expect(Selector(".is-album").visible).ok();
    await t.click(Selector(".preview"));
    await t.expect(Selector(".meta-title").count).eql(1);
    await menu.openPage("albums");
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await toolbar.triggerToolbarAction("refresh");
    }
    await toolbar.search("Garden");
    await t.expect(Selector(".is-album").visible).ok();
    await t.click(Selector(".preview"));
    await t.expect(Selector(".meta-title").count).eql(8);
  }
);
