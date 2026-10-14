import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { readFile, mkdir } from "node:fs/promises";
import { execFileSync } from "node:child_process";
import path from "node:path";
import { chromium } from "playwright";

const url = process.env.PHOTOPRISM_TEST_URL || "http://localhost:2342";
const manifestPath = path.resolve(process.env.PHOTOPRISM_EXPORT_BACKEND_MANIFEST || ".local/export-backend/cases.json");
const library = path.resolve(process.env.PHOTOPRISM_EXPORT_FIXTURE_PATH || ".local/export-flow");
const output = path.resolve(process.env.PHOTOPRISM_EXPORT_BACKEND_PATH || ".local/export-backend/downloads");
const cases = JSON.parse(await readFile(manifestPath, "utf8"));
await mkdir(output, { recursive: true });
const browser = await chromium.launch({ executablePath: process.env.PHOTOPRISM_BROWSER_PATH || "/usr/bin/chromium", headless: true, args: ["--no-sandbox", "--disable-dev-shm-usage"] });
let passed = 0;

// openExport opens a real indexed picture and waits for its backend capability probes.
async function openExport(page, uid) {
  await page.evaluate((uid) => {
    const view = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
    const index = view.results.findIndex((photo) => photo.UID === uid);
    if (index < 0) {
      throw new Error(`Fixture ${uid} missing from browse results`);
    }
    view.openPhoto(index);
  }, uid);
  await page.getByRole("button", { name: "More options", exact: true }).click();
  await page.locator(".action-menu__item.action-export").click();
  await page.waitForFunction(() => {
    const view = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
    return view.$options.name === "PPhotoExportDialog" && !view.loading && view.groups.length;
  });
}

// waitForDownloads waits for complete queue handoffs without assuming disk-save completion.
async function waitForDownloads(page, count) {
  await page.waitForFunction((count) => {
    const view = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
    return view.$options.name === "PPhotoExportDialog" && view.handed === count;
  }, count);
}

// closeExport dismisses the dialog through its always-visible title control and restores Lightbox.
async function closeExport(page) {
  const close = page.locator(".p-photo-export-dialog .action-dismiss");
  assert.ok(await close.isVisible(), "title Close must remain available");
  await close.click();
  await page.waitForFunction(() => document.querySelector("#app").__vue_app__.config.globalProperties.$view.current.$options.name === "PLightbox");
  await page.locator(".p-photo-export-dialog").waitFor({ state: "hidden" });
}

try {
  for (const viewport of [{ width: 1280, height: 900 }, { width: 375, height: 667 }]) {
    const context = await browser.newContext({ viewport, isMobile: viewport.width === 375, hasTouch: viewport.width === 375, acceptDownloads: true });
    try {
      const page = await context.newPage();
      const errors = [];
      const downloads = [];
      page.on("pageerror", (error) => errors.push(error.message));
      page.on("download", (download) => downloads.push(download));
      await page.goto(`${url}/library/browse`);
      await page.waitForFunction(() => document.querySelector("#app")?.__vue_app__?.config.globalProperties.$view.current.results?.length);
      for (const fixture of cases) {
        downloads.length = 0;
        await openExport(page, fixture.uid);
        const options = await page.evaluate(() => {
          const view = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
          const selected = view.groups.flatMap((group) => group.options.filter((option) => option.kind !== "thumbnail" || option.size === "fit_720"));
          view.selected = selected.map((option) => option.id);
          return selected;
        });
        assert.ok(options.length, fixture.label);
        if (fixture.label === "AVC") {
          assert.equal(options.filter((option) => option.kind === "avc").length, 0);
        }
        if (fixture.label.startsWith("Apple")) {
          assert.equal(options.filter((option) => option.kind === "original").length, 2);
          assert.equal(await page.evaluate(() => document.querySelector("#app").__vue_app__.config.globalProperties.$view.current.groups.length), 2);
        }
        if (fixture.label === "Motion") {
          assert.equal(options.filter((option) => option.kind === "video-part").length, 1);
        }
        await page.locator(".p-photo-export-dialog .action-export").click();
        await waitForDownloads(page, options.length);
        const deadline = Date.now() + 15000;
        while (downloads.length < options.length) {
          assert.ok(Date.now() < deadline, "browser did not report every handed download");
          await new Promise((resolve) => setTimeout(resolve, 25));
        }
        assert.equal(downloads.length, options.length, fixture.label);
        const states = await page.evaluate(() => {
          const view = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
          return view.downloads.map((item) => ({ id: item.id, kind: item.kind, filename: item.filename, format: item.format, source: item.source }));
        });
        for (const item of states) {
          const download = downloads.find((download) => download.suggestedFilename() === item.filename);
          assert.ok(download, item.filename);
          const filename = path.join(output, `${viewport.width}-${fixture.label}-${item.filename}`);
          await download.saveAs(filename);
          const bytes = await readFile(filename);
          if (item.kind === "original" || item.kind === "preview") {
            assert.equal(createHash("sha1").update(bytes).digest("hex"), item.source.hash);
          } else if (item.kind === "avc") {
            const info = JSON.parse(execFileSync("ffprobe", ["-v", "error", "-show_entries", "stream=codec_name:format=format_name", "-of", "json", filename], { encoding: "utf8" }));
            assert.equal(info.streams[0].codec_name, "h264");
            assert.ok(info.format.format_name.includes("mp4"));
          } else if (item.kind === "video-part") {
            const expected = await readFile(path.join(library, "originals/backend-hevc.mov"));
            assert.deepEqual(bytes, expected);
            assert.equal(item.format, "mov");
          } else {
            const dimensions = execFileSync("python3", ["-c", "from PIL import Image; import sys; im=Image.open(sys.argv[1]); print(im.format, *im.size)", filename], { encoding: "utf8" });
            assert.ok(dimensions.startsWith("JPEG "));
          }
        }
        await closeExport(page);
        await page.keyboard.press("Escape");
        await page.waitForFunction(() => !document.querySelector("#app").__vue_app__.config.globalProperties.$view.current?.model);
        passed++;
        process.stdout.write(`${JSON.stringify({ viewport: viewport.width, fixture: fixture.label, files: states.length, backend: "real" })}\n`);
      }

      const fixture = cases.find((fixture) => fixture.label === "HEVC");
      await openExport(page, fixture.uid);
      await page.evaluate(() => {
        const view = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
        const option = view.groups.flatMap((group) => group.options).find((option) => option.kind === "avc");
        option.ready = false;
        view.selected = [option.id];
      });
      let failOnce = true;
      await page.route(/\/export\?export=prepare/, async (route) => {
        if (failOnce) {
          failOnce = false;
          await route.abort("connectionfailed");
        } else {
          await route.continue();
        }
      });
      await page.locator(".p-photo-export-dialog .action-export").click();
      await page.waitForFunction(() => document.querySelector("#app").__vue_app__.config.globalProperties.$view.current.errors === 1);
      await page.locator(".p-photo-export-dialog .action-retry-output").click();
      await waitForDownloads(page, 1);
      await page.unroute(/\/export\?export=prepare/);
      await closeExport(page);
      await page.keyboard.press("Escape");
      await page.waitForFunction(() => !document.querySelector("#app").__vue_app__.config.globalProperties.$view.current?.model);

      downloads.length = 0;
      await openExport(page, fixture.uid);
      await page.evaluate(() => {
        const view = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
        const option = view.groups.flatMap((group) => group.options).find((option) => option.kind === "avc");
        option.ready = false;
        view.selected = [option.id];
      });
      let release;
      let prepared;
      const preparationReached = new Promise((resolve) => {
        prepared = resolve;
      });
      const delivery = new Promise((resolve) => {
        release = resolve;
      });
      await page.route(/\/export\?export=prepare/, async (route) => {
        const response = await route.fetch();
        prepared();
        await delivery;
        await route.fulfill({ response });
      });
      await page.locator(".p-photo-export-dialog .action-export").click();
      await preparationReached;
      await page.locator(viewport.width === 375 ? ".p-photo-export-dialog .action-dismiss" : ".p-photo-export-dialog .action-cancel").click();
      release();
      await page.waitForFunction(() => document.querySelector("#app").__vue_app__.config.globalProperties.$view.current.$options.name === "PLightbox");
      await page.unrouteAll({ behavior: "wait" });
      assert.equal(downloads.length, 0, "cancel must suppress late handoffs");
      await page.keyboard.press("Escape");
      await page.waitForFunction(() => !document.querySelector("#app").__vue_app__.config.globalProperties.$view.current?.model);
      await openExport(page, fixture.uid);
      assert.ok(await page.evaluate(() => document.querySelector("#app").__vue_app__.config.globalProperties.$view.current.groups.flatMap((group) => group.options).find((option) => option.kind === "avc").ready));
      await closeExport(page);
      passed += 2;
      process.stdout.write(`${JSON.stringify({ viewport: viewport.width, controls: "retry,cancel,cache-rediscovery", backend: "real", faults: "network abort and delayed real response" })}\n`);
      assert.deepEqual(errors, []);
    } finally {
      await context.close();
    }
  }
  process.stdout.write(`${JSON.stringify({ passed, backend: "real" })}\n`);
} finally {
  await browser.close();
}
