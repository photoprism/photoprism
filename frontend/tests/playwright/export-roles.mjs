import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { mkdir, readFile } from "node:fs/promises";
import path from "node:path";
import { chromium } from "playwright";
import { validateRoleCases, validateEdition } from "./export-contracts.mjs";

const url = process.env.PHOTOPRISM_TEST_URL || "http://localhost:2342";
const root = path.resolve(process.env.PHOTOPRISM_EXPORT_FIXTURE_PATH || ".local/export-backend");
const manifest = validateRoleCases(JSON.parse(await readFile(process.env.PHOTOPRISM_EXPORT_ROLE_MANIFEST || path.join(root, "roles.json"), "utf8")));
const cases = JSON.parse(await readFile(path.join(root, "cases.json"), "utf8"));
for (const session of manifest.sessions) {
  if (session.passwordEnv) {
    assert.ok(process.env[session.passwordEnv], `${session.role}: missing password environment variable ${session.passwordEnv}`);
  }
  assert.ok(cases.find((fixture) => fixture.uid === session.photo), `${session.role}: missing real photo fixture`);
}
const output = path.join(root, "role-downloads");
await mkdir(output, { recursive: true });
const browser = await chromium.launch({ executablePath: process.env.PHOTOPRISM_BROWSER_PATH || "/usr/bin/chromium", headless: true, args: ["--no-sandbox", "--disable-dev-shm-usage"] });
let adminContext;
let album;
let link;
let headers;
let passed = 0;

// login establishes a real password session through the application's session service.
async function login(page, session) {
  await page.goto(`${url}/library/login`);
  await page.waitForFunction(() => document.querySelector("#app")?.__vue_app__?.config.globalProperties.$session);
  await page.evaluate(async ({ username, password }) => {
    await document.querySelector("#app").__vue_app__.config.globalProperties.$session.login(username, password);
  }, { username: session.username, password: process.env[session.passwordEnv] });
}

// assertEdition refuses a community fallback when paid-edition roles are required.
async function assertEdition(page) {
  const state = await page.evaluate(() => {
    const globals = document.querySelector("#app").__vue_app__.config.globalProperties;
    return { edition: globals.$config.values.edition, sponsor: globals.$config.values.sponsor, public: globals.$config.isPublic() };
  });
  validateEdition(manifest.edition, state);
}

// post creates isolated share fixtures with the authenticated admin session.
async function post(route, data) {
  const response = await adminContext.request.post(`${url}/api/v1/${route}`, { headers, data });
  assert.ok(response.ok(), `share fixture ${route.split("/")[0]} failed: HTTP ${response.status()}`);
  return response.json();
}

// openPhoto opens a specific fixture after its real route loads.
async function openPhoto(page, uid) {
  await page.waitForFunction((uid) => document.querySelector("#app")?.__vue_app__?.config.globalProperties.$view.current?.results?.some((photo) => photo.UID === uid), uid);
  await page.evaluate((uid) => {
    const view = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
    view.openPhoto(view.results.findIndex((photo) => photo.UID === uid));
  }, uid);
  await page.getByRole("button", { name: "More options", exact: true }).waitFor();
}

// verifyDownload checks a complete browser attachment against its fixture source hash.
async function verifyDownload(download, session, label) {
  assert.equal(await download.failure(), null);
  const filename = path.join(output, `${session.role}-${label}-${download.suggestedFilename()}`);
  await download.saveAs(filename);
  const expected = cases.find((fixture) => fixture.uid === session.photo).files.find((file) => file.root === "/");
  assert.equal(createHash("sha1").update(await readFile(filename)).digest("hex"), expected.hash);
}

try {
  adminContext = await browser.newContext();
  const setup = await adminContext.newPage();
  await login(setup, manifest.sessions.find((session) => session.role === "admin"));
  await assertEdition(setup);
  headers = { "X-Auth-Token": await setup.evaluate(() => document.querySelector("#app").__vue_app__.config.globalProperties.$session.getAuthToken()) };
  album = await post("albums", { Title: `Export Role Acceptance ${Date.now()}` });
  await post(`albums/${album.UID}/photos`, { photos: [...new Set(manifest.sessions.map((session) => session.photo))] });
  link = await post(`albums/${album.UID}/links`, {});
  for (const viewport of [{ width: 1280, height: 900 }, { width: 375, height: 667 }]) {
    for (const session of manifest.sessions) {
      const context = await browser.newContext({ viewport, isMobile: viewport.width === 375, hasTouch: viewport.width === 375, acceptDownloads: true });
      try {
        const page = await context.newPage();
        const errors = [];
        const downloads = [];
        page.on("pageerror", (error) => errors.push(error.message));
        page.on("download", (download) => downloads.push(download));
        if (session.role !== "share-link") {
          await login(page, session);
        }
        const route = session.url.replace("{token}", link.Token).replace("{album}", album.UID);
        await page.goto(`${url}${route}`);
        await openPhoto(page, session.photo);
        await assertEdition(page);
        const state = await page.evaluate(() => {
          const globals = document.querySelector("#app").__vue_app__.config.globalProperties;
          return { role: globals.$session.getUser().Role, files: globals.$config.allow("files", "access_all"), download: globals.$config.allow("photos", "download") && globals.$config.feature("download"), export: globals.$config.feature("export") };
        });
        assert.ok(state.export, "independent Export setting must be enabled");
        assert.equal(state.files, session.export, `${session.role}: actual session file access`);
        assert.equal(state.download, session.download, `${session.role}: actual legacy Download capability`);
        if (session.role !== "share-link") {
          assert.equal(state.role, session.role, "requested role must survive real login");
        } else {
          assert.ok(["guest", "visitor"].includes(state.role), "share-link must use a restricted real session");
        }
        await page.getByRole("button", { name: "More options", exact: true }).click();
        assert.equal(await page.locator(".action-menu__item.action-export").count(), session.export ? 1 : 0);
        assert.equal(await page.locator(".action-menu__item.action-download").count(), session.download ? 1 : 0);
        await page.getByRole("button", { name: "More options", exact: true }).click();
        await page.locator(".action-menu__list").waitFor({ state: "hidden" });
        await page.waitForFunction(() => document.querySelector("#app").__vue_app__.config.globalProperties.$view.current.$options.name === "PLightbox");
        await page.keyboard.press("Control+Shift+e");
        if (session.export) {
          await page.waitForFunction(() => {
            const view = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
            return view.$options.name === "PPhotoExportDialog" && !view.loading && view.groups.length;
          });
          await page.evaluate(() => {
            const view = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
            view.selected = [view.groups[0].options.find((option) => option.kind === "original").id];
          });
          const [exported] = await Promise.all([page.waitForEvent("download"), page.locator(".p-photo-export-dialog .action-export").click()]);
          await verifyDownload(exported, session, `export-${viewport.width}`);
          await page.locator(".p-photo-export-dialog").waitFor({ state: "hidden" });
          await page.waitForFunction(() => document.querySelector("#app").__vue_app__.config.globalProperties.$view.current.$options.name === "PLightbox");
        } else {
          await page.waitForTimeout(350);
          assert.equal(await page.locator(".p-photo-export-dialog").count(), 0, "restricted shortcut must not open Export");
          assert.equal(downloads.length, 0);
        }
        if (session.download) {
          await page.waitForFunction(() => {
            const view = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
            return view.$options.name === "PLightbox" && !view.exportBlocking;
          });
          if (!(await page.locator(".action-menu__list").isVisible())) {
            await page.getByRole("button", { name: "More options", exact: true }).click();
          }
          await page.locator(".action-menu__item.action-download").waitFor({ state: "visible" });
          const [legacy] = await Promise.all([page.waitForEvent("download"), page.locator(".action-menu__item.action-download").click()]);
          await verifyDownload(legacy, session, `download-${viewport.width}`);
          const [shortcut] = await Promise.all([page.waitForEvent("download"), page.keyboard.press("Control+d")]);
          await verifyDownload(shortcut, session, `shortcut-${viewport.width}`);
        }
        assert.equal(downloads.length, (session.export ? 1 : 0) + (session.download ? 2 : 0), "unexpected handoff count");
        assert.deepEqual(errors, []);
        passed++;
        process.stdout.write(`${JSON.stringify({ edition: manifest.edition, role: session.role, viewport: viewport.width, export: session.export, download: session.download, session: "real" })}\n`);
      } catch (error) {
        const state = await context.pages()[0]?.evaluate(() => {
          const view = document.querySelector("#app")?.__vue_app__?.config.globalProperties.$view.current;
          const globals = document.querySelector("#app")?.__vue_app__?.config.globalProperties;
          return { view: view?.$options.name, visible: view?.visible, exportDialog: document.querySelectorAll(".p-photo-export-dialog").length, downloadUrl: !!view?.model?.DownloadUrl, token: !!globals?.$config.downloadToken, canDownload: view?.canDownload };
        });
        process.stderr.write(`${JSON.stringify({ role: session.role, viewport: viewport.width, state })}\n`);
        throw error;
      } finally {
        await context.close();
      }
    }
  }
  process.stdout.write(`${JSON.stringify({ passed, edition: manifest.edition, sessions: "real" })}\n`);
} finally {
  try {
    if (link && album) {
      const removed = await adminContext.request.delete(`${url}/api/v1/albums/${album.UID}/links/${link.UID}`, { headers });
      assert.ok(removed.ok(), "share-link cleanup failed");
    }
    if (album) {
      const removed = await adminContext.request.post(`${url}/api/v1/batch/albums/delete`, { headers, data: { albums: [album.UID] } });
      assert.ok(removed.ok(), "share-album cleanup failed");
    }
  } finally {
    await adminContext?.close();
    await browser.close();
  }
}
