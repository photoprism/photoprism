import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import path from "node:path";
import { chromium } from "playwright";

const fixtureRoot = path.resolve(process.env.PHOTOPRISM_EXPORT_FIXTURE_PATH || ".local/export-flow");
const cases = JSON.parse(await readFile(path.join(fixtureRoot, "cases.json"), "utf8"));
const jpeg = cases.find((item) => item.label === "JPEG");
const rawJpeg = cases.find((item) => item.label === "RAWPlusJPEG");
const url = process.env.PHOTOPRISM_TEST_URL || "http://localhost:2342";
const browser = await chromium.launch({ executablePath: process.env.PHOTOPRISM_BROWSER_PATH || "/usr/bin/chromium", headless: true, args: ["--no-sandbox", "--disable-dev-shm-usage"] });

// openPhoto waits for the library view and opens a real fixture photo.
async function openPhoto(page, uid) {
  await page.waitForFunction(() => document.querySelector("#app").__vue_app__.config.globalProperties.$view.current.$options.name === "PPagePhotos");
  await page.evaluate((uid) => {
    const view = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
    view.openPhoto(view.results.findIndex((photo) => photo.UID === uid));
  }, uid);
  await page.getByRole("button", { name: "More options", exact: true }).waitFor();
}

// openDialog clicks Export and waits until full stack options are loaded.
async function openDialog(page) {
  await page.getByRole("button", { name: "More options", exact: true }).click();
  await page.locator(".action-menu__item.action-export").click();
  await page.waitForFunction(() => {
    const view = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
    return view.$options.name === "PPhotoExportDialog" && view.groups.length > 0 && !view.loading;
  });
}

// waitForDownloads waits for a known number of handoffs and checks the queue stays quiet afterwards.
async function waitForDownloads(downloads, count) {
  const end = Date.now() + 10000;
  while (downloads.length < count && Date.now() < end) {
    await new Promise((resolve) => setTimeout(resolve, 20));
  }
  assert.equal(downloads.length, count);
  await new Promise((resolve) => setTimeout(resolve, 750));
  assert.equal(downloads.length, count);
}

try {
  for (const viewport of [{ width: 1280, height: 900 }, { width: 375, height: 667 }]) {
    const context = await browser.newContext({ viewport, isMobile: viewport.width === 375, hasTouch: viewport.width === 375, acceptDownloads: true });
    const page = await context.newPage();
    const downloads = [];
    page.on("download", (download) => downloads.push(download));
    await page.goto(`${url}/library/browse`);
    await page.waitForFunction(() => document.querySelector("#app")?.__vue_app__?.config.globalProperties.$view.current?.results?.length > 0);
    await openPhoto(page, jpeg.uid);
    await openDialog(page);
    const dialog = page.locator(".p-photo-export-dialog");
    const original = dialog.locator('input[value^="original:"]');
    await original.focus();
    await page.keyboard.press("Space");
    assert.equal(await original.isChecked(), true);
    for (const key of ["Tab", "Shift+Tab", "Tab", "Tab", "Shift+Tab"]) {
      await page.keyboard.press(key);
      await page.waitForFunction(() => document.querySelector(".p-photo-export-dialog")?.contains(document.activeElement));
    }
    const title = dialog.locator(".v-expansion-panel-title").first();
    await title.focus();
    await page.keyboard.press("Enter");
    assert.equal(await title.getAttribute("aria-expanded"), "false");
    await page.keyboard.press("Enter");
    assert.equal(await title.getAttribute("aria-expanded"), "true");
    await page.keyboard.press("ArrowRight");
    await page.keyboard.press("z");
    await page.keyboard.press("Control+d");
    assert.equal(downloads.length, 0);
    await page.keyboard.press("Escape");
    await dialog.waitFor({ state: "hidden" });
    await page.waitForFunction(() => document.querySelector("#app").__vue_app__.config.globalProperties.$view.current.$options.name === "PLightbox");
    assert.equal(await page.evaluate(() => document.activeElement?.classList.contains("p-lightbox__content")), true);
    await page.keyboard.press("Escape");

    for (const action of ["Cancel", "Escape", "Permission", "PhotoChange", "LightboxClose"]) {
      downloads.length = 0;
      await openPhoto(page, jpeg.uid);
      await openDialog(page);
      assert.equal(await page.evaluate(() => document.querySelector("#app").__vue_app__.config.globalProperties.$view.current.selected.length), 1, "reopening resets selection");
      for (const input of await dialog.locator('input[type="checkbox"]').all()) {
        await input.check();
      }
      const total = await page.evaluate(() => {
        const view = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
        window.__exportCaseDialog = view;
        return view.selected.length;
      });
      await dialog.locator(".action-export").click();
      await page.waitForFunction(() => window.__exportCaseDialog.handed > 0);
      if (action === "Cancel") {
        await dialog.locator(".action-dismiss").click();
      }
      if (action === "Escape") {
        await page.keyboard.press("Escape");
      }
      if (action === "Permission") {
        await page.evaluate(() => {
          const files = document.querySelector("#app").__vue_app__.config.globalProperties.$config.values.acl.files;
          files.access_all = false;
          files.full_access = false;
        });
      }
      if (action === "PhotoChange") {
        await page.evaluate(() => document.querySelector("#app").__vue_app__.config.globalProperties.$view.scopes.find((view) => view.$options.name === "PLightbox").pswp().next());
      }
      if (action === "LightboxClose") {
        await page.evaluate(() => {
          document.querySelector("#app").__vue_app__.config.globalProperties.$view.scopes.find((view) => view.$options.name === "PLightbox").close();
        });
      }
      await dialog.waitFor({ state: "hidden" });
      const handed = await page.evaluate(() => window.__exportCaseDialog.handed);
      assert.ok(handed > 0 && handed < total, `${action} must stop pending handoffs`);
      await waitForDownloads(downloads, handed);
      if (action === "Permission") {
        await page.evaluate(() => {
          document.querySelector("#app").__vue_app__.config.globalProperties.$config.values.acl.files.access_all = true;
        });
      }
      if (action !== "LightboxClose") {
        await page.keyboard.press("Escape");
      }
      console.log(JSON.stringify({ viewport: `${viewport.width}x${viewport.height}`, action, handed, pendingStopped: total - handed }));
    }

    await openPhoto(page, rawJpeg.uid);
    await page.evaluate(() => {
      const globals = document.querySelector("#app").__vue_app__.config.globalProperties;
      globals.$config.values.settings.features.download = true;
      Object.assign(globals.$config.values.settings.download, { disabled: false, originals: true, mediaRaw: false, mediaSidecar: false });
      globals.$view.current.applyContext();
    });
    for (const shortcut of ["Control+d", "Meta+d"]) {
      downloads.length = 0;
      await page.keyboard.press(shortcut);
      await waitForDownloads(downloads, 1);
      assert.equal(new URL(downloads[0].url()).searchParams.has("filename"), false, "legacy Download retains its naming and filters");
      assert.equal(path.extname(downloads[0].suggestedFilename()).toLowerCase(), ".jpg", "RAW download filter remains active");
    }
    downloads.length = 0;
    await page.getByRole("button", { name: "More options", exact: true }).click();
    await page.locator(".action-menu__item.action-download").click();
    await waitForDownloads(downloads, 1);
    assert.equal(path.extname(downloads[0].suggestedFilename()).toLowerCase(), ".jpg");
    await page.evaluate(() => {
      const globals = document.querySelector("#app").__vue_app__.config.globalProperties;
      globals.$config.values.settings.features.export = false;
      globals.$view.current.$refs.menu.hide();
    });
    await page.getByRole("button", { name: "More options", exact: true }).click();
    assert.equal(await page.locator(".action-menu__item.action-export").count(), 0);
    console.log(JSON.stringify({ viewport: `${viewport.width}x${viewport.height}`, keyboardAndLegacyDownload: "passed" }));
    await page.close();
    await context.close();
  }
} finally {
  await browser.close();
}
