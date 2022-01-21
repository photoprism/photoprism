import { Selector } from "testcafe";
import testcafeconfig from "../testcafeconfig";
import Menu from "../../page-model/menu";
import Toolbar from "../../page-model/toolbar";
import NewPage from "../../page-model/page";
import Library from "../../page-model/library";

fixture`Import file from folder`.page`${testcafeconfig.url}`;

const menu = new Menu();
const toolbar = new Toolbar();
const newpage = new NewPage();
const library = new Library();

test.meta("testID", "library-import-001")("Import files from folder using copy", async (t) => {
  await menu.openPage("labels");
  await toolbar.search("bakery");
  await t.expect(Selector("div.no-results").visible).ok();
  await menu.openPage("library");
  await t
    .click(Selector("#tab-library-import"))
    .typeText(library.openImportFolderSelect, "/B", { replace: true })
    .click(newpage.selectOption.nth(0))
    .click(library.import)
    //TODO replace wait
    .wait(60000);
  await menu.openPage("labels");
  await toolbar.triggerToolbarAction("reload", "");
  await toolbar.search("bakery");
  await t.expect(Selector(".is-label").visible).ok();
});
