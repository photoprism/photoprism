import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { readFile, mkdir } from "node:fs/promises";
import { execFileSync } from "node:child_process";
import path from "node:path";
import { chromium } from "playwright";
import { stackCases, validateBackendCases } from "./export-contracts.mjs";

const url = process.env.PHOTOPRISM_TEST_URL || "http://localhost:2342";
const manifestPath = path.resolve(process.env.PHOTOPRISM_EXPORT_BACKEND_MANIFEST || ".local/export-backend/cases.json");
const library = path.resolve(process.env.PHOTOPRISM_EXPORT_FIXTURE_PATH || ".local/export-backend");
const output = path.resolve(process.env.PHOTOPRISM_EXPORT_BACKEND_PATH || ".local/export-backend/downloads");
const cases = validateBackendCases(JSON.parse(await readFile(manifestPath, "utf8")));
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

// verifyFormat checks file signatures and real animation/container metadata.
function verifyFormat(filename, format, bytes, animated = false) {
  const signatures = {
    jpg: bytes[0] === 0xff && bytes[1] === 0xd8,
    png: bytes.subarray(0, 8).equals(Buffer.from([137, 80, 78, 71, 13, 10, 26, 10])),
    gif: bytes.subarray(0, 4).toString() === "GIF8",
    webp: bytes.subarray(0, 4).toString() === "RIFF" && bytes.subarray(8, 12).toString() === "WEBP",
    dng: ["49492a00", "4d4d002a"].includes(bytes.subarray(0, 4).toString("hex")),
    pdf: bytes.subarray(0, 5).toString() === "%PDF-",
    svg: /<svg[ >]/.test(bytes.toString("utf8")),
    heic: bytes.subarray(4, 8).toString() === "ftyp" && /heic|heix|mif1/.test(bytes.subarray(8, 32).toString()),
  };
  if (["mov", "mp4"].includes(format)) {
    const info = JSON.parse(execFileSync("ffprobe", ["-v", "error", "-show_entries", "stream=codec_type,codec_name,width,height:format=format_name:format_tags=major_brand", "-of", "json", filename], { encoding: "utf8" }));
    assert.ok(info.format.format_name.includes(format), `${filename}: real container`);
    const brand = info.format.tags?.major_brand?.trim();
    if (format === "mov") {
      assert.equal(brand, "qt", `${filename}: native QuickTime container`);
    } else {
      // Shared AVC sidecars can retain the source movie's QuickTime brand.
      assert.match(brand || "", /^(isom|iso[2-9]|mp4[12]|avc1|M4V|qt)$/, `${filename}: ISO media container`);
    }
    const video = info.streams.find((stream) => stream.codec_type === "video");
    assert.ok(video?.codec_name && video.width > 0 && video.height > 0, `${filename}: real video stream`);
  } else {
    assert.equal(signatures[format], true, `${filename}: real ${format} signature`);
  }
  if (animated) {
    const frames = Number(execFileSync("python3", ["-c", "from PIL import Image; import sys; print(Image.open(sys.argv[1]).n_frames)", filename], { encoding: "utf8" }));
    assert.ok(frames > 1, `${filename}: real animation frames`);
  }
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
        const expectedOriginals = fixture.files.filter((file) => file.root === "/");
        const originals = options.filter((option) => option.kind === "original");
        assert.deepEqual(originals.map((option) => option.source.name).sort(), expectedOriginals.map((file) => file.name).sort(), `${fixture.label}: every original must be downloadable`);
        for (const original of originals) {
          const expected = expectedOriginals.find((file) => file.name === original.source.name);
          assert.equal(original.source.hash, expected.hash, "source must match the fixture manifest");
          assert.equal(original.format, expected.format);
        }
        if (stackCases[fixture.label]) {
          const counts = {
            TwoImage: { thumbnail: 2 }, TwoRAW: { preview: 2, thumbnail: 2 }, TwoSVG: { preview: 2, thumbnail: 2 },
            TwoPDF: { preview: 2, thumbnail: 2 }, TwoGIF: { preview: 2, thumbnail: 2, avc: 2 },
            TwoAPNG: { preview: 2, thumbnail: 2, avc: 2 }, TwoWebP: { preview: 2, thumbnail: 2, avc: 2 },
            TwoVideo: { preview: 2, thumbnail: 2, avc: 2 }, TwoMotion: { thumbnail: 2, "video-part": 2 },
            TwoApple: { thumbnail: 2 }, RAWPlusJPEG: { preview: 1, thumbnail: 2 },
          }[fixture.label];
          for (const kind of ["preview", "thumbnail", "avc", "video-part"]) {
            assert.equal(options.filter((option) => option.kind === kind).length, counts[kind] || 0, `${fixture.label}: ${kind} count`);
          }
          if (fixture.label === "TwoApple") {
            assert.equal(await page.evaluate(() => document.querySelector("#app").__vue_app__.config.globalProperties.$view.current.groups.filter((group) => group.apple).length), 4);
          }
        }
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
          return view.downloads.map((item) => ({ id: item.id, kind: item.kind, filename: item.filename, format: item.format, width: item.width, height: item.height, source: item.source }));
        });
        assert.equal(states.length, options.length);
        assert.equal(new Set(states.map((item) => item.filename.toLowerCase())).size, states.length, "every final filename must be unique");
        assert.equal(new Set(downloads.map((download) => download.suggestedFilename().toLowerCase())).size, states.length, "server attachment names must remain unique");
        for (const item of states) {
          const download = downloads.find((download) => download.suggestedFilename() === item.filename);
          assert.ok(download, item.filename);
          const filename = path.join(output, `${viewport.width}-${fixture.label}-${item.filename}`);
          assert.equal(await download.failure(), null, item.filename);
          await download.saveAs(filename);
          const bytes = await readFile(filename);
          verifyFormat(filename, item.format, bytes, item.kind === "original" && item.source.mediaType === "animated");
          if (item.kind === "original" || item.kind === "preview") {
            assert.equal(createHash("sha1").update(bytes).digest("hex"), item.source.hash);
            if (item.kind === "preview") {
              const dimensions = execFileSync("python3", ["-c", "from PIL import Image; import sys; print(*Image.open(sys.argv[1]).size)", filename], { encoding: "utf8" });
              assert.equal(dimensions.trim(), `${item.width} ${item.height}`, "preview must retain its real dimensions");
            }
          } else if (item.kind === "avc") {
            const info = JSON.parse(execFileSync("ffprobe", ["-v", "error", "-show_entries", "stream=codec_type,codec_name,width,height:format=format_name:format_tags=major_brand", "-of", "json", filename], { encoding: "utf8" }));
            const video = info.streams.find((stream) => stream.codec_type === "video");
            assert.equal(video.codec_name, "h264");
            assert.ok(video.width > 0 && video.height > 0);
            if (item.width && item.height) {
              assert.deepEqual([video.width, video.height], [item.width, item.height]);
            }
            assert.ok(info.format.format_name.includes("mp4"));
          } else if (item.kind === "video-part") {
            const native = fixture.native?.[item.source.name];
            assert.ok(native, `${fixture.label}: missing native motion-video source`);
            const expected = await readFile(path.join(library, "originals", native));
            assert.deepEqual(bytes, expected);
            assert.equal(item.format, "mov");
          } else {
            const dimensions = execFileSync("python3", ["-c", "from PIL import Image; import sys; im=Image.open(sys.argv[1]); print(im.format, *im.size)", filename], { encoding: "utf8" });
            assert.equal(dimensions.trim(), `JPEG ${item.width} ${item.height}`, "fitted JPEG dimensions must match the offered choice");
          }
        }
        await closeExport(page);
        await page.keyboard.press("Escape");
        await page.waitForFunction(() => !document.querySelector("#app").__vue_app__.config.globalProperties.$view.current?.model);
        passed++;
        process.stdout.write(`${JSON.stringify({ viewport: viewport.width, fixture: fixture.label, files: states.length, backend: "real" })}\n`);
      }

      downloads.length = 0;
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
      const retryDeadline = Date.now() + 15000;
      while (downloads.length < 1) {
        assert.ok(Date.now() < retryDeadline, "Retry download timed out");
        await new Promise((resolve) => setTimeout(resolve, 25));
      }
      assert.equal(downloads.length, 1, "Retry must not duplicate completed outputs");
      const retried = downloads[0];
      assert.equal(await retried.failure(), null);
      const retryName = path.join(output, `${viewport.width}-retry-${retried.suggestedFilename()}`);
      await retried.saveAs(retryName);
      const retryBytes = await readFile(retryName);
      verifyFormat(retryName, "mp4", retryBytes);
      const codec = execFileSync("ffprobe", ["-v", "error", "-select_streams", "v:0", "-show_entries", "stream=codec_name", "-of", "csv=p=0", retryName], { encoding: "utf8" });
      assert.equal(codec.trim(), "h264");
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
      try {
        await Promise.race([preparationReached, new Promise((_, reject) => {
          const timeout = setTimeout(() => reject(new Error("preparation response timed out")), 30000);
          timeout.unref();
          preparationReached.finally(() => clearTimeout(timeout));
        })]);
        await page.locator(viewport.width === 375 ? ".p-photo-export-dialog .action-dismiss" : ".p-photo-export-dialog .action-cancel").click();
      } finally {
        release();
      }
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
