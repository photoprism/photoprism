import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import { execFileSync } from "node:child_process";
import { createServer, request as httpRequest } from "node:http";
import { request as httpsRequest } from "node:https";
import path from "node:path";
import { chromium } from "playwright";

const library = path.resolve(process.env.PHOTOPRISM_EXPORT_FIXTURE_PATH || ".local/export-flow");
const root = path.resolve(process.env.PHOTOPRISM_EXPORT_CLIENT_PATH || ".local/export-media-client");
const url = process.env.PHOTOPRISM_TEST_URL || "http://localhost:2342";
const manifest = JSON.parse(await readFile(path.join(library, "cases.json"), "utf8"));
const uid = manifest.find((item) => item.label === "JPEG").uid;
await mkdir(root, { recursive: true });

// generateVideo creates real codec fixtures without involving the application converter.
function generateVideo(name, codec) {
  const target = path.join(root, name);
  const args = ["-y", "-v", "error", "-f", "lavfi", "-i", "testsrc2=size=160x120:rate=10", "-t", "0.3", "-threads", "1", "-c:v", codec];
  if (codec === "libx265") {
    args.push("-x265-params", "pools=1:frame-threads=1:log-level=error", "-tag:v", "hvc1");
  }
  execFileSync("ffmpeg", [...args, target], { stdio: "ignore" });
  return target;
}

const avcPath = generateVideo("fixture-avc.mp4", "libx264");
const hevcPath = generateVideo("fixture-hevc.mov", "libx265");
const jpegPath = path.join(library, "originals/single.jpg");
const heicPath = path.join(library, "originals/heic.heic");
execFileSync("python3", ["-c", "from PIL import Image; import sys; a=Image.new('RGB',(160,120),'red'); b=Image.new('RGB',(160,120),'blue'); [a.save(sys.argv[1]+'/animation.'+ext, save_all=True, append_images=[b], duration=150, loop=0) for ext in ['gif','png','webp']]", root]);
const motionPath = path.join(root, "motion.jpg");
await writeFile(motionPath, Buffer.concat([await readFile(jpegPath), await readFile(hevcPath)]));
let sequence = 0;
const contents = new Map();
const prepared = new Set();
const serverErrors = [];

// source describes indexed file metadata while retaining real fixture bytes for attachments.
async function source(name, filePath, mediaType = "image", extras = {}) {
  const bytes = await readFile(filePath);
  const hash = createHash("sha1").update(bytes).digest("hex");
  sequence++;
  const record = { UID: `fs6sg6bqhhin${String(sequence).padStart(4, "0")}`, Hash: hash, Name: name, Root: "/", FileType: path.extname(name).slice(1), MediaType: mediaType, Width: 4000, Height: 3000, Size: bytes.length, ...extras };
  contents.set(hash, { bytes, filePath, record });
  return record;
}

const mov = await source("hevc.mov", hevcPath, "video", { Video: true, Width: 160, Height: 120 });
const native = await source("avc.mp4", avcPath, "video", { Video: true, Width: 160, Height: 120 });
const still = await source("live.jpg", jpegPath);
const heic = await source("live.heic", heicPath, "image", { OriginalName: "旅行.heic" });
const clip = await source("live.mov", hevcPath, "video", { Video: true, Width: 160, Height: 120 });
const preview = await source("live.heic.jpg", jpegPath, "image", { Root: "sidecar" });
const motion = await source("motion.jpg", motionPath, "live", { Video: true });
const cases = [
  { label: "HEVCMOV", files: [mov] },
  { label: "AVCMP4", files: [native] },
  { label: "AppleJPEG", files: [clip, still] },
  { label: "AppleHEIC", files: [heic, clip, preview] },
  { label: "MotionHEVC", files: [motion] },
];
for (const format of ["gif", "png", "webp"]) {
  const animated = await source(`animation.${format}`, path.join(root, `animation.${format}`), "image", { Frames: 2, Duration: 300000000, Width: 160, Height: 120 });
  const poster = await source(`animation.${format}.jpg`, jpegPath, "image", { Root: "sidecar" });
  cases.push({ label: `Animated${format}`, files: [animated, poster] });
}
cases.push({ label: "Mixed", files: [mov, motion, cases.at(-1).files[0], cases.at(-1).files[1], still] });
const browser = await chromium.launch({ executablePath: process.env.PHOTOPRISM_BROWSER_PATH || "/usr/bin/chromium", headless: true, args: ["--no-sandbox", "--disable-dev-shm-usage"] });
const reports = [];
const fixtureServer = createServer(serveRequest);
await new Promise((resolve) => fixtureServer.listen(0, "127.0.0.1", resolve));
const fixtureUrl = `http://127.0.0.1:${fixtureServer.address().port}`;

// json fulfills the planned video contract without claiming backend verification.
async function json(route, body, status = 200) {
  await route.fulfill({ status, contentType: "application/json", headers: { "Cache-Control": "no-store" }, body: JSON.stringify(body) });
}

// attachment writes fixture bytes over HTTP so Chromium can complete the download.
function attachment(response, bytes, filename) {
  response.writeHead(200, { "Content-Type": "application/octet-stream", "Content-Length": bytes.length, "Content-Disposition": `attachment; filename*=UTF-8''${encodeURIComponent(filename)}` });
  response.end(bytes);
}

// serveRequest provides fixture attachments and proxies other requests to the isolated application.
async function serveRequest(request, response) {
  try {
    const incoming = new URL(request.url, fixtureUrl);
    const parts = incoming.pathname.split("/");
    const filename = incoming.searchParams.get("filename");
    const hash = parts.at(-1);
    if (filename && incoming.pathname.includes("/dl/") && contents.has(hash)) {
      attachment(response, contents.get(hash).bytes, filename);
      return;
    }
    if (filename && incoming.pathname.includes("/videos/") && incoming.searchParams.get("export") === "download") {
      assert.ok(prepared.has(parts.at(-3)), "download must not start a preparation");
      const target = incoming.searchParams.get("kind") === "video-part" ? hevcPath : avcPath;
      attachment(response, await readFile(target), filename);
      return;
    }
    if (filename && incoming.pathname.includes("/t/") && contents.has(parts.at(-3))) {
      const size = parts.at(-1) === "fit_720" ? [720, 720] : [1920, 1200];
      const target = path.join(root, `${parts.at(-3)}_${parts.at(-1)}.jpg`);
      execFileSync("python3", ["-c", "from PIL import Image; import sys; im=Image.open(sys.argv[1]); im.thumbnail((int(sys.argv[3]),int(sys.argv[4]))); im.save(sys.argv[2],format='JPEG')", contents.get(parts.at(-3)).filePath, target, ...size.map(String)]);
      attachment(response, await readFile(target), filename);
      return;
    }
    const target = new URL(request.url, url);
    const send = target.protocol === "https:" ? httpsRequest : httpRequest;
    const upstream = send(target, { method: request.method, headers: { ...request.headers, host: target.host } }, (received) => {
      response.writeHead(received.statusCode, received.headers);
      received.pipe(response);
    });
    upstream.on("error", (error) => {
      serverErrors.push(error.message);
      response.writeHead(502);
      response.end();
    });
    request.pipe(upstream);
  } catch (error) {
    serverErrors.push(error.message);
    response.writeHead(500);
    response.end();
  }
}

// chooseExport opens the live action menu and selects Export.
// The client contract checks avoid activator hover timing when reopening after Cancel.
async function chooseExport(page) {
  await page.getByRole("button", { name: "More options", exact: true }).waitFor();
  await page.waitForFunction(() => {
    const view = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
    return view.$options.name === "PLightbox" && !!view.$refs.menu;
  });
  await page.evaluate(() => document.querySelector("#app").__vue_app__.config.globalProperties.$view.current.$refs.menu.show());
  await page.locator(".action-menu__item.action-export").click();
}

// openExport opens the isolated photo and waits for all client capability probes.
async function openExport(page) {
  await page.evaluate((uid) => {
    const view = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
    view.openPhoto(view.results.findIndex((photo) => photo.UID === uid));
  }, uid);
  await chooseExport(page);
  await page.waitForFunction(() => {
    const view = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
    return view.$options.name === "PPhotoExportDialog" && !view.loading && view.groups.length;
  });
}

// waitFor waits for Node-side download and intercepted-request events.
async function waitFor(predicate) {
  const deadline = Date.now() + 15000;
  while (!predicate()) {
    assert.ok(Date.now() < deadline, "timed out waiting for browser events");
    await new Promise((resolve) => setTimeout(resolve, 50));
  }
}

try {
  for (const viewport of [{ width: 1280, height: 900 }, { width: 375, height: 667 }]) {
    const context = await browser.newContext({ viewport, isMobile: viewport.width === 375, hasTouch: viewport.width === 375, acceptDownloads: true });
    try {
      const page = await context.newPage();
      let current = cases[0];
      const attempts = new Map();
      let failFirst = false;
      let preparations = 0;
      const downloads = [];
      const errors = [];
      page.on("pageerror", (error) => errors.push(error.message));
      page.on("download", (download) => downloads.push(download));
      await page.route(`**/photos/${uid}`, async (route) => {
        const response = await route.fetch();
        const photo = await response.json();
        await json(route, { ...photo, Files: current.files });
      });
      await page.route(/\/videos\/[^/]+\/[^/]+\/export\?export=(probe|prepare)(?:&|$)/, async (route) => {
        const request = new URL(route.request().url());
        const hash = request.pathname.split("/").at(-3);
        const record = contents.get(hash).record;
        const kind = request.searchParams.get("kind");
        const result = { ready: true, format: kind === "video-part" ? "mov" : "mp4", codec: kind === "video-part" ? "hvc1" : "avc1", width: 160, height: 120, bytes: (await readFile(kind === "video-part" ? hevcPath : avcPath)).length };
        if (request.searchParams.get("export") === "probe") {
          const key = current.label.startsWith("Apple") ? "live-pair" : "";
          await json(route, { codec: hash === native.Hash ? "H.264" : "hvc1", duration: 0.3, appleKey: key, avc: { ...result, format: "mp4", ready: prepared.has(hash), preparable: true }, videoPart: record.MediaType === "live" ? { ...result, format: "mov", codec: "hvc1", ready: prepared.has(hash), preparable: true } : undefined });
        } else if (request.searchParams.get("export") === "prepare") {
          preparations++;
          const count = (attempts.get(hash) || 0) + 1;
          attempts.set(hash, count);
          await new Promise((resolve) => setTimeout(resolve, 1200));
          if (failFirst && count === 1) {
            await json(route, { error: "fixture encoder failure" }, 500);
          } else {
            prepared.add(hash);
            await json(route, result);
          }
          preparations--;
        }
      });
      await page.goto(`${fixtureUrl}/library/browse`);
      await page.waitForFunction(() => document.querySelector("#app")?.__vue_app__?.config.globalProperties.$view.current.results?.length);
      for (const fixture of cases) {
        current = fixture;
        prepared.clear();
        attempts.clear();
        downloads.length = 0;
        failFirst = fixture.label === "HEVCMOV";
        await openExport(page);
        const options = await page.evaluate(() => {
          const view = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
          const options = view.groups.flatMap((group) => group.options.filter((item) => item.kind !== "thumbnail" || ["fit_720", "fit_1920"].includes(item.size)));
          view.selected = options.map((item) => item.id);
          return options;
        });
        await page.locator(".p-photo-export-dialog .action-export").click();
        const derived = options.some((item) => ["avc", "video-part"].includes(item.kind));
        if (derived) {
          await page.waitForFunction(() => document.querySelector("#app").__vue_app__.config.globalProperties.$view.current.handed > 0);
          assert.ok(await page.locator(".p-photo-export-dialog").getByText("Preparing", { exact: true }).count());
        }
        if (failFirst) {
          await page.locator(".action-retry-output").click();
        }
        await page.waitForFunction(() => document.querySelector("#app").__vue_app__.config.globalProperties.$view.current.complete);
        await page.waitForFunction((count) => document.querySelector("#app").__vue_app__.config.globalProperties.$view.current.handed === count, options.length);
        await waitFor(() => downloads.length === options.length);
        const fixed = await page.evaluate(() => JSON.parse(JSON.stringify(document.querySelector("#app").__vue_app__.config.globalProperties.$view.current.downloads)));
        for (const item of fixed) {
          const download = downloads.find((entry) => entry.suggestedFilename() === item.filename);
          assert.ok(download, `missing ${fixture.label} attachment ${item.filename}`);
          assert.equal(await download.failure(), null, `${fixture.label}: ${item.filename}`);
          const target = path.join(root, `${viewport.width}_${fixture.label}_${item.filename}`);
          await download.saveAs(target);
          const bytes = await readFile(target);
          if (["original", "preview"].includes(item.kind)) {
            assert.deepEqual(bytes, contents.get(item.source.hash).bytes);
          } else if (["avc", "video-part"].includes(item.kind)) {
            const metadata = JSON.parse(execFileSync("ffprobe", ["-v", "error", "-show_entries", "stream=codec_name:format=format_name:format_tags=major_brand", "-of", "json", target], { encoding: "utf8" }));
            assert.equal(metadata.streams[0].codec_name, item.kind === "avc" ? "h264" : "hevc");
            assert.ok(metadata.format.format_name.includes("mov"));
            assert.equal(metadata.format.tags.major_brand, item.kind === "avc" ? "isom" : "qt  ");
            if (item.kind === "video-part") {
              assert.deepEqual(bytes, await readFile(hevcPath));
            }
          } else {
            const dimensions = JSON.parse(execFileSync("python3", ["-c", "from PIL import Image; import sys,json; print(json.dumps(Image.open(sys.argv[1]).size))", target], { encoding: "utf8" }));
            assert.deepEqual(dimensions, [item.width, item.height]);
          }
        }
        assert.equal(downloads.length, fixed.length);
        await page.screenshot({ path: path.join(root, `${fixture.label}__${viewport.width}x${viewport.height}.jpg`), type: "jpeg", quality: 70 });
        await page.locator(".p-photo-export-dialog .action-dismiss").click();
        await page.waitForFunction(() => document.querySelector("#app").__vue_app__.config.globalProperties.$view.current.$options.name === "PLightbox");
        await page.keyboard.press("Escape");
        await page.waitForFunction(() => document.querySelector("#app").__vue_app__.config.globalProperties.$view.current.$options.name === "PPagePhotos");
        reports.push({ case: fixture.label, viewport: viewport.width, files: fixed.length, api: "mocked" });
      }
      for (const cancel of ["Cancel", "Escape"]) {
        current = cases[0];
        prepared.clear();
        attempts.clear();
        downloads.length = 0;
        failFirst = false;
        await openExport(page);
        await page.evaluate(() => {
          const view = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
          view.selected = view.groups[0].options.filter((item) => item.kind === "avc").map((item) => item.id);
        });
        await page.locator(".p-photo-export-dialog .action-export").click();
        await waitFor(() => preparations === 1);
        if (cancel === "Escape") {
          await page.keyboard.press("Escape");
        } else {
          await page.locator(".p-photo-export-dialog .action-dismiss").click();
        }
        await page.waitForFunction(() => document.querySelector("#app").__vue_app__.config.globalProperties.$view.current.$options.name === "PLightbox");
        await waitFor(() => preparations === 0);
        assert.equal(downloads.length, 0, "late preparation must not hand off a download");
        assert.ok(prepared.has(mov.Hash), "started preparation may finish after cancellation");
        await chooseExport(page);
        await page.waitForFunction(() => {
          const view = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
          return view.$options.name === "PPhotoExportDialog" && !view.loading;
        });
        assert.equal(await page.evaluate(() => {
          const view = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
          return view.groups[0].options.find((item) => item.kind === "avc").ready && !view.started && !view.selected.length;
        }), true, "reopening must discover ready cache with initial selection");
        await page.locator(".p-photo-export-dialog .action-dismiss").click();
        await page.waitForFunction(() => document.querySelector("#app").__vue_app__.config.globalProperties.$view.current.$options.name === "PLightbox");
        await page.keyboard.press("Escape");
        await page.waitForFunction(() => document.querySelector("#app").__vue_app__.config.globalProperties.$view.current.$options.name === "PPagePhotos");
        reports.push({ case: `Preparing${cancel}`, viewport: viewport.width, files: 0, api: "mocked" });
      }
      assert.deepEqual(errors, []);
      assert.deepEqual(serverErrors, []);
    } finally {
      await context.close();
    }
  }
} finally {
  await browser.close();
  await new Promise((resolve) => fixtureServer.close(resolve));
}
await writeFile(path.join(root, "results.json"), JSON.stringify(reports, null, 2));
console.log(JSON.stringify({ passed: reports.length, api: "mocked", output: root }));
