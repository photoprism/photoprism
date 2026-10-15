import assert from "node:assert/strict";
import { readFile, mkdir } from "node:fs/promises";
import path from "node:path";
import { execFileSync } from "node:child_process";
import { chromium } from "playwright";

// This standalone acceptance test uses an isolated library described by the supplied case manifest.
const fixtureRoot = path.resolve(process.env.PHOTOPRISM_EXPORT_FIXTURE_PATH || ".local/export-flow");
const cases = JSON.parse(await readFile(path.join(fixtureRoot, "cases.json"), "utf8"));
const url = process.env.PHOTOPRISM_TEST_URL || "http://localhost:2342";
const output = path.join(fixtureRoot, "downloads");
await mkdir(output, { recursive: true });
const browser = await chromium.launch({
  executablePath: process.env.PHOTOPRISM_BROWSER_PATH || "/usr/bin/chromium",
  headless: true,
  args: ["--no-sandbox", "--disable-dev-shm-usage"],
});
const reports = [];

// waitForCount waits for browser download events after the dialog reports its handoffs.
async function waitForCount(downloads, count) {
  const deadline = Date.now() + 15000;
  while (downloads.length < count && Date.now() < deadline) {
    await new Promise((resolve) => setTimeout(resolve, 20));
  }
  assert.equal(downloads.length, count, "the browser must receive every selected file");
}

// openExport opens the requested real photo and clicks the Lightbox's Export action.
async function openExport(page, uid) {
  await page.waitForFunction(() => {
    const view = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
    return view.$options.name === "PPagePhotos" && !view.loading && view.results?.length > 0;
  });
  await page.evaluate((uid) => {
    const view = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
    const index = view.results.findIndex((photo) => photo.UID === uid);
    if (index < 0) {
      throw new Error(`missing acceptance photo ${uid}`);
    }
    view.openPhoto(index);
  }, uid);
  await page.getByRole("button", { name: "More options", exact: true }).click();
  assert.equal(await page.locator(".action-menu__item.action-download").count(), 0, "Download is disabled independently");
  await page.locator(".action-menu__item.action-export").click();
  await page.waitForFunction(() => {
    const view = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
    return view.$options.name === "PPhotoExportDialog" && !view.loading && view.groups.length > 0;
  });
  await page.evaluate(() => {
    window.__exportDialog = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
  });
}

// readGroups snapshots advertised choices for comparison with actual downloaded content.
async function readGroups(page) {
  return page.evaluate(() => {
    const view = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
    return JSON.parse(JSON.stringify({ groups: view.groups, selected: view.selected }));
  });
}

// expectedFiles derives names independently from the selected source records and resolves collisions.
function expectedFiles(groups) {
  const names = new Set();
  const files = [];
  for (const group of groups) {
    const original = (group.original.originalName || group.original.name).split(/[\\/]/).pop();
    const stem = original.slice(0, original.lastIndexOf(".")).trim();
    for (const option of group.options.filter((option) => option.kind !== "thumbnail" || ["fit_720", "fit_1920"].includes(option.size) || option.label === "Full Size")) {
      const extension = option.kind === "original" ? path.extname(option.source.name) : `.${option.format}`;
      const suffix = option.kind === "original" ? "" : option.kind === "preview" ? "_preview" : `_${option.size}`;
      const base = stem + suffix;
      let filename = base + extension;
      for (let number = 2; names.has(filename.toLowerCase()); number++) {
        filename = `${base}_${number}${extension}`;
      }
      names.add(filename.toLowerCase());
      files.push({ ...option, filename });
    }
  }
  return files;
}

// verifyFile checks real bytes for originals/previews and decoded dimensions for generated JPEGs.
async function verifyFile(download, item, label) {
  assert.equal(download.suggestedFilename(), item.filename);
  assert.equal(await download.failure(), null);
  const target = path.join(output, `${label}_${item.filename}`);
  await download.saveAs(target);
  const bytes = await readFile(target);
  if (item.kind === "thumbnail") {
    assert.equal(bytes[0], 0xff);
    assert.equal(bytes[1], 0xd8);
    const dimensions = JSON.parse(execFileSync("python3", ["-c", "from PIL import Image; import sys,json; print(json.dumps(Image.open(sys.argv[1]).size))", target], { encoding: "utf8" }));
    assert.deepEqual(dimensions, [item.width, item.height]);
  } else {
    const source = path.join(fixtureRoot, item.source.root === "/" ? "originals" : "sidecar", item.source.name);
    assert.deepEqual(bytes, await readFile(source));
  }
}

try {
  for (const viewport of [{ width: 1280, height: 900 }, { width: 375, height: 667 }]) {
    const context = await browser.newContext({ viewport, isMobile: viewport.width === 375, hasTouch: viewport.width === 375, acceptDownloads: true });
    const page = await context.newPage();
    const errors = [];
    page.on("pageerror", (error) => errors.push(error.message));
    const downloads = [];
    page.on("download", (download) => downloads.push(download));
    await page.goto(`${url}/library/browse`);
    await page.waitForFunction(() => {
      const view = document.querySelector("#app")?.__vue_app__?.config.globalProperties.$view.current;
      return view?.results?.length > 0 && !view.loading;
    });
    for (const fixture of cases) {
      downloads.length = 0;
      await openExport(page, fixture.uid);
      const { groups, selected } = await readGroups(page);
      const groupCount = { TwoJPEG: 2, RAWPlusJPEG: 2, TwoRAW: 2, SameBasename: 2, AmbiguousPreview: 3 }[fixture.label] || 1;
      assert.equal(groups.length, groupCount);
      assert.equal(selected.length, ["RAW", "TwoRAW", "AmbiguousPreview"].includes(fixture.label) ? 0 : 1);
      if (fixture.label === "AmbiguousPreview") {
        assert.equal(groups.filter((group) => group.previewMissing).length, 2);
      }
      const expected = expectedFiles(groups);
      const dialog = page.locator(".p-photo-export-dialog");
      if (viewport.width === 375) {
        const bounds = await dialog.locator(".v-card").boundingBox();
        assert.equal(Math.round(bounds.width), 375);
        assert.equal(Math.round(bounds.height), 667);
      }
      for (const title of await dialog.locator(".v-expansion-panel-title").all()) {
        if (await title.getAttribute("aria-expanded") !== "true") {
          await title.click();
        }
      }
      for (const item of expected) {
        await dialog.locator(`input[value="${item.id}"]`).check();
      }
      assert.equal(await page.evaluate(() => document.querySelector("#app").__vue_app__.config.globalProperties.$view.current.selected.length), expected.length);
      await dialog.locator(".action-export").click();
      await page.waitForFunction(() => window.__exportDialog.complete);
      await waitForCount(downloads, expected.length);
      for (let index = 0; index < expected.length; index++) {
        await verifyFile(downloads[index], expected[index], `${viewport.width}_${fixture.label}`);
      }
      await dialog.waitFor({ state: "hidden" });
      assert.equal(await dialog.locator(".export-download").count(), 0);
      await page.screenshot({ path: path.join(output, `${fixture.label}__${viewport.width}x${viewport.height}.jpg`), type: "jpeg", quality: 65 });
      assert.equal(await page.evaluate(() => document.querySelector("#app").__vue_app__.config.globalProperties.$view.current.$options.name), "PLightbox");
      assert.equal(await page.evaluate(() => document.activeElement?.classList.contains("p-lightbox__content")), true);
      await page.keyboard.press("Escape");
      await page.locator(".p-lightbox.v-overlay--active").waitFor({ state: "hidden" });
      reports.push({ viewport: `${viewport.width}x${viewport.height}`, fixture: fixture.label, files: expected.length, bytesAndDimensions: "passed" });
      console.log(JSON.stringify(reports.at(-1)));
    }
    assert.deepEqual(errors, []);
    await page.close();
    await context.close();
  }
} finally {
  await browser.close();
}
console.log(JSON.stringify({ scenarios: reports.length, result: "passed" }));
