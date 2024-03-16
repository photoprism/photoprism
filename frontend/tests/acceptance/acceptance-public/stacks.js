import { Selector } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";
import Menu from "../page-model/menu";
import Toolbar from "../page-model/toolbar";
import Photo from "../page-model/photo";
import PhotoViewer from "../page-model/photoviewer";
import Page from "../page-model/page";
import PhotoEdit from "../page-model/photo-edit";
import Library from "../page-model/library";

fixture`Test stacks`.page`${testcafeconfig.url}`;

const menu = new Menu();
const toolbar = new Toolbar();
const photo = new Photo();
const photoviewer = new PhotoViewer();
const page = new Page();
const photoedit = new PhotoEdit();
const library = new Library();

test.meta("testID", "stacks-001").meta({ type: "short", mode: "public" })(
  "Common: View all files of a stack",
  async (t) => {
    await toolbar.search("ski");
    const SequentialPhotoUid = await photo.getNthPhotoUid("all", 0);
    await photo.checkHoverActionAvailability("uid", SequentialPhotoUid, "open", true);
    if (t.browser.platform === "desktop") {
      console.log(t.browser.platform);
      await photo.triggerHoverAction("nth", 0, "open");
      await photoviewer.triggerPhotoViewerAction("next");
      await photoviewer.triggerPhotoViewerAction("previous");
      await photoviewer.triggerPhotoViewerAction("close");
      await t.expect(Selector("#photo-viewer").visible).notOk();
    }
    await photo.checkHoverActionAvailability("uid", SequentialPhotoUid, "open", true);
  }
);

test.meta("testID", "stacks-002").meta({ type: "short", mode: "public" })(
  "Common: Change primary file",
  async (t) => {
    await toolbar.search("ski");
    const SequentialPhotoUid = await photo.getNthPhotoUid("all", 0);
    await toolbar.setFilter("view", "Cards");
    await t
      .click(page.cardTitle.withAttribute("data-uid", SequentialPhotoUid))
      .click(photoedit.filesTab);
    const FirstFileName = await Selector("div.caption").nth(0).innerText;

    await t.expect(FirstFileName).contains("photos8_1_ski.jpg");

    await t
      .click(photoedit.toggleExpandFile.nth(1))
      .click(photoedit.makeFilePrimary)
      .click(photoedit.dialogClose)
      .click(page.cardTitle.withAttribute("data-uid", SequentialPhotoUid));
    const FirstFileNameAfterChange = await Selector("div.caption").nth(0).innerText;

    await t
      .expect(FirstFileNameAfterChange)
      .notContains("photos8_1_ski.jpg")
      .expect(FirstFileNameAfterChange)
      .contains("photos8_2_ski.jpg");
  }
);

test.meta("testID", "stacks-003").meta({ type: "short", mode: "public" })(
  "Common: Ungroup files",
  async (t) => {
    await toolbar.search("group");
    await toolbar.setFilter("view", "Cards");
    const PhotoCount = await photo.getPhotoCount("all");
    const SequentialPhotoUid = await photo.getNthPhotoUid("all", 0);

    await t.expect(PhotoCount).eql(1);

    await menu.openPage("stacks");
    await photo.checkHoverActionAvailability("uid", SequentialPhotoUid, "open", true);
    await toolbar.setFilter("view", "Cards");
    await t
      .click(page.cardTitle.withAttribute("data-uid", SequentialPhotoUid))
      .click(photoedit.filesTab)
      .click(photoedit.toggleExpandFile.nth(0))
      .click(photoedit.toggleExpandFile.nth(1))
      .click(photoedit.unstackFile)
      .wait(12000)
      .click(photoedit.dialogClose);
    await menu.openPage("browse");
    await toolbar.search("group");
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await toolbar.triggerToolbarAction("reload");
    }
    const PhotoCountAfterUngroup = await photo.getPhotoCount("all");

    await t.expect(PhotoCountAfterUngroup).eql(2);
    await photo.checkHoverActionAvailability("uid", SequentialPhotoUid, "open", false);
  }
);

test.meta("testID", "stacks-004").meta({ mode: "public" })(
  "Common: Delete non primary file",
  async (t) => {
    await menu.openPage("library");
    await t
      .click(library.importTab)
      .click(library.openImportFolderSelect)
      .click(page.selectOption.withText("/pizza"))
      .click(library.import)
      .wait(10000);
    await menu.openPage("browse");
    await toolbar.search("pizza");
    await toolbar.setFilter("view", "Cards");
    const PhotoCount = await photo.getPhotoCount("all");
    const PhotoUid = await photo.getNthPhotoUid("all", 0);

    await t.expect(PhotoCount).eql(1);

    await t.click(page.cardTitle.withAttribute("data-uid", PhotoUid)).click(photoedit.filesTab);
    const FileCount = await photoedit.getFileCount();

    await t.expect(FileCount).eql(2);

    await t
      .click(photoedit.toggleExpandFile.nth(1))
      .click(Selector(photoedit.deleteFile))
      .click(Selector(".action-confirm"))
      .wait(10000);
    const FileCountAfterDeletion = await photoedit.getFileCount();

    await t.expect(FileCountAfterDeletion).eql(1);
  }
);
