import assert from "node:assert/strict";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import { chromium } from "playwright";
import { customDark, customLight } from "../vitest/helpers/themes.js";
import { colors, variables } from "../../src/options/ui.js";

const library = path.resolve(process.env.PHOTOPRISM_EXPORT_FIXTURE_PATH || ".local/export-flow");
const output = path.resolve(process.env.PHOTOPRISM_EXPORT_UI_PATH || ".local/screenshots/export-ui");
const url = process.env.PHOTOPRISM_TEST_URL || "http://localhost:2342";
const manifest = JSON.parse(await readFile(path.join(library, "cases.json"), "utf8"));
const uid = manifest.find((item) => item.label === "JPEG").uid;
const raster = await readFile(path.join(library, "originals/single.jpg"));
const themeSource = await readFile(new URL("../../src/options/themes.js", import.meta.url), "utf8");
const themeOptions = themeSource.slice(themeSource.indexOf("let options = ["), themeSource.indexOf("/* Theme Helper Functions */"));
const builtins = [...themeOptions.matchAll(/value: "([^"]+)"/g)].map((match) => match[1]);
await mkdir(output, { recursive: true });
let sequence = 0;

// file describes a distinct indexed source for client-only layout checks.
function file(name, mediaType = "image", extra = {}) {
  sequence++;
  return { UID: `fs6sg6bqhhin${String(sequence).padStart(4, "0")}`, Hash: String(sequence).padStart(40, "0"), Root: "/", Name: name, FileType: path.extname(name).slice(1), MediaType: mediaType, Width: 4000, Height: 3000, Size: 123456, ...extra };
}

const still = file("旅行." + "long".repeat(32) + ".heic");
const movie = file("旅行.mov", "video", { Video: true, Codec: "hvc1", Duration: 3000000000 });
const preview = file(still.Name + ".jpg", "image", { Root: "sidecar" });
const unrelated = file("unrelated.jpg");
const animation = file("animation.webp", "image", { Frames: 2, Duration: 1000000000 });
const cases = [
  { label: "AppleLivePhoto", files: [movie, unrelated, still, preview] },
  { label: "Mixed", files: [movie, unrelated, animation, file("animation.webp.jpg", "image", { Root: "sidecar" }), file("motion.jpg", "live", { Video: true })] },
  { label: "AVC", files: [file("avc.mov", "video", { Video: true, Codec: "avc1" })] },
  { label: "HEVC", files: [movie] },
  { label: "RAW", files: [file("raw.dng", "raw"), file("raw.dng.jpg", "image", { Root: "sidecar" })] },
  { label: "JPEG", files: [unrelated] },
  { label: "SVG", files: [file("vector.svg", "vector"), file("vector.svg.jpg", "image", { Root: "sidecar" })] },
  { label: "PDF", files: [file("document.pdf", "document", { Width: 0, Height: 0, Size: undefined })] },
];
const browser = await chromium.launch({ executablePath: process.env.PHOTOPRISM_BROWSER_PATH || "/usr/bin/chromium", headless: true, args: ["--no-sandbox", "--disable-dev-shm-usage"] });
const reports = [];

// openExport opens the fixture Lightbox and waits for its full detail/probe results.
async function openExport(page) {
  await page.evaluate((uid) => {
    const view = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
    view.openPhoto(view.results.findIndex((photo) => photo.UID === uid));
  }, uid);
  await page.getByRole("button", { name: "More options", exact: true }).click();
  await page.locator(".action-menu__item.action-export").click();
  await page.waitForFunction(() => {
    const view = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
    return view.$options.name === "PPhotoExportDialog" && !view.loading && view.groups.length;
  });
}

// closeExport verifies Escape returns focus to Lightbox before leaving the photo.
async function closeExport(page) {
  await page.keyboard.press("Escape");
  await page.waitForFunction(() => {
    const globals = document.querySelector("#app").__vue_app__.config.globalProperties;
    return globals.$view.current.$options.name === "PLightbox" && document.activeElement?.classList.contains("p-lightbox__content");
  });
  await page.keyboard.press("Escape");
  await page.waitForFunction(() => document.querySelector("#app").__vue_app__.config.globalProperties.$view.current.$options.name === "PPagePhotos");
}

// checkLayout measures wrapping, fixed actions, scrolling, and selected-theme surface colors.
async function checkLayout(page, viewport, theme) {
  const notification = page.locator(".p-notify");
  if (await notification.isVisible()) {
    await notification.click();
  }
  const dialog = page.locator(".p-photo-export-dialog");
  const measurements = await dialog.evaluate((element) => {
    const globals = document.querySelector("#app").__vue_app__.config.globalProperties;
    const card = element.querySelector(".export-card");
    const content = element.querySelector(".export-content");
    const title = element.querySelector(".export-title");
    const actions = element.querySelector(".export-actions");
    const rect = card.getBoundingClientRect();
    const before = [title.getBoundingClientRect().top, actions.getBoundingClientRect().bottom];
    content.scrollTop = content.scrollHeight;
    const after = [title.getBoundingClientRect().top, actions.getBoundingClientRect().bottom];
    const overflowing = [...element.querySelectorAll(".export-filename, .export-option-label")].some((item) => item.scrollWidth > item.clientWidth + 1);
    const style = getComputedStyle(card);
    const header = element.querySelector(".export-filename");
    const option = element.querySelector(".export-option-description");
    const actionsStyle = getComputedStyle(actions);
    const contentStyle = getComputedStyle(content);
    const titleStyle = getComputedStyle(title);
    const primary = element.querySelector(".action-export");
    return { width: rect.width, height: rect.height, before, after, overflowing, color: style.color, surface: style.backgroundColor, theme: globals.$config.themeName, themeClasses: card.className, expected: style.getPropertyValue("--v-theme-surface").trim(), scrollable: contentStyle.overflowY, titlePadding: titleStyle.paddingInlineStart, titleMargin: titleStyle.marginInlineStart, primaryTheme: primary?.className, contentPadding: contentStyle.paddingInlineStart, actionPadding: actionsStyle.paddingInlineStart, alignment: option ? Math.abs(header.getBoundingClientRect().left - option.getBoundingClientRect().left) : 0 };
  });
  assert.equal(Math.round(measurements.width), viewport.width === 375 ? 375 : 560);
  if (viewport.width === 375) {
    assert.equal(Math.round(measurements.height), viewport.height);
    const button = await dialog.locator(".action-export").boundingBox();
    assert.ok(button.width >= viewport.width - 40, `mobile primary action spans the footer (${button.width}px)`);
  }
  assert.equal(measurements.theme, theme);
  assert.ok(measurements.themeClasses.split(" ").includes(`v-theme--${theme}`), "Export must inherit the selected application theme");
  assert.equal(measurements.surface.replace(/\s/g, ""), `rgb(${measurements.expected})`.replace(/\s/g, ""));
  assert.notEqual(measurements.color, measurements.surface, "text remains visible");
  assert.deepEqual(measurements.before, measurements.after, "only choices scroll");
  assert.equal(measurements.scrollable, "auto");
  assert.equal(measurements.contentPadding, viewport.width === 375 ? "16px" : "24px");
  assert.equal(measurements.actionPadding, measurements.contentPadding);
  assert.equal(measurements.titlePadding, measurements.contentPadding);
  assert.equal(measurements.titleMargin, "0px");
  assert.ok(measurements.primaryTheme.split(" ").includes(`v-theme--${theme}`), "the primary action uses the application theme");
  assert.ok(measurements.alignment <= 1, `header and option text align (${measurements.alignment}px)`);
  assert.equal(measurements.overflowing, false, "long Unicode names and option labels wrap");
  reports.push({ viewport: viewport.width, theme, ...measurements });
}

try {
  for (const viewport of [{ width: 1280, height: 900 }, { width: 375, height: 667 }]) {
    const context = await browser.newContext({ viewport, isMobile: viewport.width === 375, hasTouch: viewport.width === 375 });
    try {
      const page = await context.newPage();
      const errors = [];
      let current = cases[0];
      page.on("pageerror", (error) => errors.push(error.message));
      await page.route(`**/photos/${uid}`, async (route) => {
        const response = await route.fetch();
        await route.fulfill({ json: { ...await response.json(), Files: current.files } });
      });
      await page.route(/\/t\/0{30,}\d+\/[^/]+\/tile_224$/, (route) => route.fulfill({ contentType: "image/jpeg", body: raster }));
      await page.route(/\/videos\/0{30,}\d+\/[^/]+\/(?:avc|native)\?action=probe$/, (route) => {
        const hash = new URL(route.request().url()).pathname.split("/").at(-3);
        const source = current.files.find((file) => file.Hash === hash);
        return route.fulfill({ json: { codec: source.Codec || "", duration: source.Duration / 1e9 || 0, appleKey: current.label === "AppleLivePhoto" && [still.Hash, movie.Hash].includes(hash) ? "pair" : "", avc: { preparable: true, format: "mp4", codec: "avc1" }, videoPart: source.MediaType === "live" ? { preparable: true, format: "mov", codec: "hvc1" } : undefined } });
      });
      await page.goto(`${url}/library/browse`);
      await page.waitForFunction(() => document.querySelector("#app")?.__vue_app__?.config.globalProperties.$view.current.results?.length);
      for (const theme of [...builtins, customDark, customLight]) {
        const name = typeof theme === "string" ? theme : theme.name;
        const fixture = typeof theme === "string" ? null : { ...theme, colors: { ...colors[theme.dark ? "dark" : "light"], ...theme.colors }, variables: { ...variables[theme.dark ? "dark" : "light"], ...theme.variables } };
        await openExport(page);
        await page.evaluate(({ name, fixture }) => {
          const globals = document.querySelector("#app").__vue_app__.config.globalProperties;
          if (fixture) {
            globals.$config.$vuetify.theme.themes[name] = fixture;
            globals.$config.themeName = name;
            globals.$config.theme = fixture;
            globals.$config.$vuetify.theme.name = name;
            globals.$config.setBodyTheme(name);
            globals.$config.setColorMode(fixture.dark ? "dark" : "light");
          } else {
            globals.$config.setTheme(name);
          }
        }, { name, fixture });
        await page.waitForFunction((name) => document.querySelector(".export-card")?.classList.contains(`v-theme--${name}`), name);
        const result = await page.evaluate(() => {
          const view = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
          return { names: view.groups.map((group) => group.original.name), kinds: view.groups[1].options.map((option) => option.kind), thumbnail: view.groups[1].thumbnailSource.hash, selected: view.selected.length };
        });
        assert.deepEqual(result.names, [still.Name, movie.Name, unrelated.Name]);
        assert.deepEqual(result.kinds, ["original"]);
        assert.equal(result.thumbnail, preview.Hash);
        assert.equal(result.selected, 1);
        await page.locator(".p-photo-export-dialog .export-thumbnail img").first().waitFor();
        await page.locator(".p-photo-export-dialog input[value^='original:']").first().focus();
        await page.keyboard.press("Space");
        assert.equal(await page.locator(".action-export.action-confirm").textContent(), "Export 2 Files");
        await page.keyboard.press("Tab");
        assert.equal(await page.evaluate(() => document.querySelector(".p-photo-export-dialog").contains(document.activeElement)), true);
        await checkLayout(page, viewport, name);
        await page.locator(".export-content").evaluate((element) => {
          element.scrollTop = 0;
        });
        await page.screenshot({ path: path.join(output, `${name}__${viewport.width}x${viewport.height}.jpg`), type: "jpeg", quality: 70, animations: "disabled", style: ".p-notify { visibility: hidden; }" });
        await closeExport(page);
      }
      await page.evaluate(() => document.querySelector("#app").__vue_app__.config.globalProperties.$config.setTheme("default"));
      for (const fixture of cases.slice(1)) {
        current = fixture;
        await openExport(page);
        await page.evaluate(() => {
          const view = document.querySelector("#app").__vue_app__.config.globalProperties.$view.current;
          view.expanded = view.groups.map((group) => group.id);
        });
        await checkLayout(page, viewport, "default");
        await page.locator(".export-content").evaluate((element) => {
          element.scrollTop = 0;
        });
        await page.screenshot({ path: path.join(output, `${fixture.label}__${viewport.width}x${viewport.height}.jpg`), type: "jpeg", quality: 70, animations: "disabled", style: ".p-notify { visibility: hidden; }" });
        await closeExport(page);
      }
      assert.deepEqual(errors, []);
      await page.close();
    } finally {
      await context.close();
    }
  }
} finally {
  await browser.close();
}
await writeFile(path.join(output, "results.json"), JSON.stringify(reports, null, 2));
console.log(JSON.stringify({ passed: reports.length, builtins: builtins.length, output, api: "mocked" }));
