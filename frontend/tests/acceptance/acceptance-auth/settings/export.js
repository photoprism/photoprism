import { RequestLogger, Selector } from "testcafe";
import testcafeconfig from "../../../testcafeconfig.json";
import Menu from "../../page-model/menu";
import Page from "../../page-model/page";
import Toolbar from "../../page-model/toolbar";
import Photo from "../../page-model/photo";
import PhotoViewer from "../../page-model/photoviewer";
import { triggerExportShortcut } from "../../page-model/dialog-export";

const menu = new Menu();
const page = new Page();
const toolbar = new Toolbar();
const photo = new Photo();
const viewer = new PhotoViewer();
const settingsLog = RequestLogger(/\/api\/v1\/settings(?:\?|$)/, { logResponseBody: false });
const dialog = Selector(".p-photo-export-dialog.v-overlay--active");
const exportItem = Selector(".action-menu--lightbox .action-export");

fixture`Test Export settings`.page`${testcafeconfig.url}`.requestHooks(settingsLog).beforeEach(async () => {
  await page.login("admin", "photoprism");
});

// setFeature changes a real setting only when its current value differs.
async function setFeature(t, name, enabled) {
  await menu.openPage("settings");
  const checkbox = Selector(`.input-${name} input[type="checkbox"]`);
  if ((await checkbox.checked) !== enabled) {
    settingsLog.clear();
    await t.click(checkbox);
    await t.expect(settingsLog.contains((request) => request.request.method === "post" && request.response.statusCode === 200)).ok();
    await t.eval(() => location.reload());
  }
  await t.expect(checkbox.checked).eql(enabled);
}

// checkExport opens a real photo and verifies the setting gates both menu and shortcut.
async function checkExport(t, enabled) {
  await menu.openPage("browse");
  await toolbar.search("type:image");
  await viewer.openPhotoViewer("uid", await photo.getNthPhotoUid("image", 0));
  await t.hover(viewer.menuButton);
  await t.expect(Selector(".action-menu--lightbox.v-overlay--active").visible).ok();
  await t.expect(exportItem.visible).eql(enabled);
  if (enabled) {
    await t.click(exportItem);
    await t.expect(dialog.visible).ok();
    await t.expect(dialog.find(".export-group").count).gt(0);
    await t.click(dialog.find(".action-cancel"));
  } else {
    await t.hover(viewer.viewer);
    await t.expect(Selector(".action-menu--lightbox.v-overlay--active").visible).notOk();
  }
  await t.expect(viewer.viewer.visible).ok();
  await triggerExportShortcut();
  await t.expect(dialog.visible).eql(enabled);
  if (enabled) {
    await t.click(dialog.find(".action-cancel"));
  }
  await viewer.triggerPhotoViewerAction("close-button");
}

test.meta("testID", "export-settings-001").meta({ mode: "auth", type: "short" })("Common: Export settings and Download independence", async (t) => {
  await menu.openPage("settings");
  const exportCheckbox = Selector('.input-export input[type="checkbox"]');
  const downloadCheckbox = Selector('.input-download input[type="checkbox"]');
  const originalExport = await exportCheckbox.checked;
  const originalDownload = await downloadCheckbox.checked;
  try {
    await setFeature(t, "export", true);
    await checkExport(t, true);
    await setFeature(t, "download", false);
    await checkExport(t, true);
    await setFeature(t, "export", false);
    await checkExport(t, false);
    await setFeature(t, "export", true);
    await checkExport(t, true);
  } finally {
    if (await dialog.visible) {
      await t.click(dialog.find(".action-dismiss"));
    }
    if (await viewer.viewer.visible) {
      await viewer.triggerPhotoViewerAction("close-button");
    }
    await setFeature(t, "export", originalExport);
    await setFeature(t, "download", originalDownload);
  }
});
