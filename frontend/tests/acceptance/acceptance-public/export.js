import { RequestLogger, RequestMock, Selector } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";
import Toolbar from "../page-model/toolbar";
import Photo from "../page-model/photo";
import PhotoViewer from "../page-model/photoviewer";
import { triggerExportShortcut } from "../page-model/dialog-export";

fixture`Test Export dialog`.page`${testcafeconfig.url}`;

const toolbar = new Toolbar();
const photo = new Photo();
const viewer = new PhotoViewer();
const dialog = Selector(".p-photo-export-dialog.v-overlay--active");
const exportItem = Selector(".action-menu--lightbox .action-export");

// source supplies API file records for deterministic UI coverage without downloading media.
function source(id, name, type = "image", root = "/", extra = {}) {
  return { UID: `fs6sg6bqhhin${String(id).padStart(4, "0")}`, Hash: String(id).padStart(40, "0"), Name: name, Root: root, MediaType: type, FileType: name.split(".").pop(), Width: 4000, Height: 3000, Size: 10000, ...extra };
}

// openViewer opens a real indexed image before attaching the media-detail response fixture.
async function openViewer(t, query = "type:image") {
  await t.click(toolbar.cardsViewAction);
  await toolbar.search(query);
  const uid = await photo.getNthPhotoUid("all", 0);
  await t.expect(uid).ok("Export requires an indexed image fixture");
  await viewer.openPhotoViewer("uid", uid);
  return uid;
}

// openExport exercises the real lightbox menu action.
async function openExport(t) {
  await t.hover(viewer.menuButton);
  await t.expect(Selector(".action-menu--lightbox.v-overlay--active").visible).ok();
  await t.expect(exportItem.visible).ok();
  await t.click(exportItem);
  await t.expect(dialog.visible).ok();
  await t.expect(dialog.find(".export-group").count).gt(0);
}

test.meta("testID", "export-001").meta({ mode: "public", type: "short" })("Common: Export menu, shortcuts, empty selection, and Cancel", async (t) => {
  await openViewer(t);
  await openExport(t);
  await t.expect(dialog.find('input[type="checkbox"]:checked').count).eql(0);
  await t.expect(dialog.find(".action-export").hasAttribute("disabled")).ok();
  await t.click(dialog.find(".action-cancel"));
  await t.expect(dialog.visible).notOk();
  for (const metaKey of [false, true]) {
    await triggerExportShortcut(metaKey);
    await t.expect(dialog.visible).ok();
    await t.click(dialog.find('.export-option input[type="checkbox"]').nth(0));
    await t.expect(dialog.find(".export-selected").innerText).eql("1 Selected");
    await t.expect(dialog.find(".action-export").innerText).match(/export 1 file/i);
    await t.click(dialog.find(".action-cancel"));
    await t.expect(dialog.visible).notOk();
  }
});

test.meta("testID", "export-download-001").meta({ mode: "public", type: "short" })("Common: Export preserves existing Download attachments", async (t) => {
  const downloads = RequestLogger(/\/api\/v1\/dl\//, { logResponseHeaders: true });
  await t.addRequestHooks(downloads);
  await openViewer(t);
  await t.hover(viewer.menuButton);
  const downloadItem = Selector(".action-menu--lightbox .action-download");
  await t.expect(downloadItem.visible).ok();
  await t.click(downloadItem);
  await t.expect(downloads.contains((request) => request.response?.statusCode === 200 && /^attachment/i.test(request.response.headers["content-disposition"] || ""))).ok();
  await t.expect(dialog.visible).notOk();
  await viewer.triggerPhotoViewerAction("close-button");
});

const cases = [
  ["Image", [source(1, "one.jpg"), source(2, "two.jpg")], ["original", "thumbnail"]],
  ["RAW", [source(1, "one.dng", "raw"), source(2, "one.dng.jpg", "image", "sidecar")], ["original", "preview", "thumbnail"]],
  ["Vector", [source(1, "one.svg", "vector"), source(2, "one.svg.png", "image", "sidecar")], ["original", "preview", "thumbnail"]],
  ["Animated", [source(1, "one.gif", "animated"), source(2, "one.gif.jpg", "image", "sidecar")], ["original", "preview", "thumbnail", "avc"]],
  ["Document", [source(1, "one.pdf", "document"), source(2, "one.pdf.jpg", "image", "sidecar")], ["original", "preview", "thumbnail"]],
  ["Live Photo", [source(1, "one.jpg"), source(2, "one.mov", "video", "/", { Codec: "avc1" })], ["original", "thumbnail"]],
  ["Motion Photo", [source(1, "one.jpg", "live")], ["original", "thumbnail", "video-part"]],
  ["Video", [source(1, "one.mov", "video"), source(2, "one.mov.jpg", "image", "sidecar")], ["original", "preview", "thumbnail", "avc"]],
  ["RAW + JPEG", [source(1, "one.dng", "raw"), source(2, "one.jpg")], ["original"]],
  ["HEIC + JPEG", [source(1, "one.heic"), source(2, "one.heic.jpg")], ["original"]],
];

for (const [index, [label, files, kinds]] of cases.entries()) {
  test.meta("testID", `export-${String(index + 2).padStart(3, "0")}`).meta({ mode: "public", type: "short" })(`Common: Export ${label} choices and group selection counts`, async (t) => {
    const uid = await openViewer(t);
    // Media details are mocked; the menu, dialog, selection, and cancellation are real.
    const hook = RequestMock()
      .onRequestTo((request) => request.method === "get" && request.url.split("?")[0].endsWith(`/api/v1/photos/${uid}`))
      .respond({ UID: uid, Files: files }, 200, { "content-type": "application/json" });
    await t.addRequestHooks(hook);
    await openExport(t);
    const groups = dialog.find(".export-group");
    await t.expect(groups.count).eql(files.filter((file) => file.Root === "/").length);
    for (const kind of ["original", "preview", "thumbnail", "avc", "video-part"]) {
      await t.expect(groups.nth(0).find(`[data-option-id^="${kind}:"]`).exists).eql(kinds.includes(kind), `${label}: unexpected ${kind} availability`);
    }
    if (label === "Document") {
      await t.expect(dialog.innerText).contains("Full-Size Preview (First Page)");
    }
    if (label === "Live Photo" || label.endsWith("+ JPEG")) {
      const originalOnly = groups.nth(label === "Live Photo" ? 1 : 0);
      await t.click(groups.nth(1).find(".export-group-title"));
      await t.expect(originalOnly.find(".export-option").count).eql(1);
      if (label !== "Live Photo") {
        await t.expect(dialog.innerText).notContains("No preview image available");
      }
    }
    await t.click(groups.nth(0).find('input[type="checkbox"]').nth(0));
    await t.expect(groups.nth(0).find(".export-selected").innerText).eql("1 Selected");
    if (await groups.count > 1) {
      if (!(await groups.nth(1).find('input[type="checkbox"]').nth(0).visible)) {
        await t.click(groups.nth(1).find(".export-group-title"));
      }
      await t.expect(groups.nth(1).find(".export-selected").exists).notOk();
      await t.click(groups.nth(1).find('input[type="checkbox"]').nth(0));
      await t.expect(groups.nth(1).find(".export-selected").innerText).eql("1 Selected");
      await t.expect(dialog.find(".action-export").innerText).match(/export 2 files/i);
    }
    await t.click(dialog.find(".action-cancel"));
    await t.expect(dialog.visible).notOk();
  });
}

for (const [index, [filename, kind, extension]] of [
  ["monochrome-2.jpg", "original", "jpg"],
  ["monochrome-2.jpg", "thumbnail", "jpg"],
  ["elephantRAW.CR2", "original", "cr2"],
].entries()) {
  test.meta("testID", `export-download-${String(index + 2).padStart(3, "0")}`).meta({ mode: "public", type: "short" })(`Common: Export real ${filename} ${kind}`, async (t) => {
    const path = kind === "thumbnail" ? "/api/v1/t/" : "/api/v1/dl/";
    const downloads = RequestLogger((request) => request.url.includes(path), { logResponseHeaders: true });
    await t.addRequestHooks(downloads);
    await openViewer(t, `name:${filename}`);
    await openExport(t);
    const group = dialog.find(".export-group").withText(filename);
    await t.expect(group.count).eql(1);
    const checkbox = group.find(`[data-option-id^="${kind}:"] input[type="checkbox"]`).nth(0);
    if (!(await checkbox.visible)) {
      await t.click(group.find(".export-group-title"));
    }
    await t.click(checkbox);
    await t.expect(dialog.find(".export-selected").innerText).eql("1 Selected");
    downloads.clear();
    await t.click(dialog.find(".action-export"));
    await t.expect(downloads.contains((request) => {
      const disposition = request.response?.headers["content-disposition"] || "";
      return request.response?.statusCode === 200 && /^attachment/i.test(disposition)
        && new RegExp(`\\.${extension}(?:[";]|$)`, "i").test(disposition);
    })).ok();
    await t.expect(dialog.visible).notOk();
    await viewer.triggerPhotoViewerAction("close-button");
  });
}
