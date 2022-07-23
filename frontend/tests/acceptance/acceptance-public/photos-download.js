import { Selector } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";
import { RequestLogger } from "testcafe";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import Photo from "../page-model/photo";
import PhotoViewer from "../page-model/photoviewer";
import Page from "../page-model/page";

const logger = RequestLogger(/http:\/\/localhost:2343\/api\/v1\/*/, {
  logResponseHeaders: true,
  logResponseBody: true,
});

fixture`Test photos download`.page`${testcafeconfig.url}`
  .requestHooks(logger)
  .skip("Does not work in container and we have no content-disposition header anymore");

const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const photo = new Photo();
const photoviewer = new PhotoViewer();
const page = new Page();

test.meta("testID", "photos-download-001").meta({ type: "short", mode: "public" })(
  "Common: Test download jpg file from context menu and fullscreen",
  async (t) => {
    await toolbar.search("name:monochrome-2.jpg");
    const PhotoUid = await photo.getNthPhotoUid("all", 0);
    await photo.triggerHoverAction("uid", PhotoUid, "select");
    await logger.clear();
    await contextmenu.triggerContextMenuAction("download", "");
    const requestInfo = await logger.requests[1].response;
    console.log(requestInfo);
    const requestInfo0 = await logger.requests[0].response;
    console.log(requestInfo0);

    await page.validateDownloadRequest(requestInfo, "monochrome-2", ".jpg");

    await logger.clear();
    await contextmenu.clearSelection();
    await toolbar.search("name:IMG_20200711_174006.jpg");
    const SecondPhotoUid = await photo.getNthPhotoUid("all", 0);
    await t.click(Selector("div").withAttribute("data-uid", SecondPhotoUid));
    await photoviewer.openPhotoViewer("uid", SecondPhotoUid);
    await logger.clear();
    await photoviewer.triggerPhotoViewerAction("download");
    await logger.clear();
    await photoviewer.triggerPhotoViewerAction("close");
  }
);

test.meta("testID", "photos-download-002").meta({ type: "short", mode: "public" })(
  "Common: Test download video from context menu",
  async (t) => {
    await toolbar.search("name:Mohn.mp4");
    const PhotoUid = await photo.getNthPhotoUid("all", 0);
    await photo.triggerHoverAction("uid", PhotoUid, "select");
    await logger.clear();
    await contextmenu.triggerContextMenuAction("download", "");
    const requestInfo = await logger.requests[0].response;
    console.log(requestInfo);
    const requestInfo2 = await logger.requests[1].response;

    await page.validateDownloadRequest(requestInfo, "Mohn", ".mp4.jpg");
    await page.validateDownloadRequest(requestInfo2, "Mohn", ".mp4");

    await logger.clear();
    await contextmenu.clearSelection();
  }
);

test.meta("testID", "photos-download-003").meta({ mode: "public" })(
  "Common: Test download multiple jpg files from context menu",
  async (t) => {
    await toolbar.search("name:panorama_2.jpg");
    const PhotoUid = await photo.getNthPhotoUid("all", 0);
    await photo.triggerHoverAction("uid", PhotoUid, "select");
    await toolbar.search("name:IMG_6478.JPG");
    const SecondPhotoUid = await photo.getNthPhotoUid("all", 0);
    await photo.triggerHoverAction("uid", SecondPhotoUid, "select");
    await logger.clear();
    await contextmenu.triggerContextMenuAction("download", "");
    const requestInfo = await logger.requests[1].response;
    console.log(requestInfo);

    await page.validateDownloadRequest(requestInfo, "photoprism-download", ".zip");

    await logger.clear();
    await contextmenu.clearSelection();
  }
);

//TODO Check RAW files as well
test.meta("testID", "photos-download-004").meta({ mode: "public" })(
  "Common: Test raw file from context menu and fullscreen mode",
  async (t) => {
    await toolbar.search("name:elephantRAW");
    const PhotoUid = await photo.getNthPhotoUid("all", 0);
    await photo.triggerHoverAction("uid", PhotoUid, "select");
    await logger.clear();
    await contextmenu.triggerContextMenuAction("download", "");
    const requestInfo = await logger.requests[1].response;

    await page.validateDownloadRequest(requestInfo, "elephantRAW", ".JPG");

    await logger.clear();
    await contextmenu.clearSelection();
    await t.click(Selector("div").withAttribute("data-uid", Photo));
    await t.expect(Selector("#photo-viewer").visible).ok().hover(Selector(".action-download"));
    await logger.clear();
    await t.click(Selector(".action-download"));
    const requestInfo3 = await logger.requests[1].response;
    //const requestInfo4 = await logger.requests[2].response;
    await page.validateDownloadRequest(requestInfo3, "elephantRAW", ".JPG");
    //await page.validateDownloadRequest(requestInfo4, "elephantRAW", ".mp4");
    await logger.clear();
  }
);
