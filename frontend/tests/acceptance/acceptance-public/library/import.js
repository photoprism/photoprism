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
