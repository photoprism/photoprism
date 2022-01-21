import { Selector } from "testcafe";
import testcafeconfig from "./testcafeconfig";
import { RequestLogger } from "testcafe";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import Photo from "../page-model/photo";
import PhotoViewer from "../page-model/photoviewer";
import NewPage from "../page-model/page";
import PhotoViews from "../page-model/photo-views";

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
const newpage = new NewPage();
const photoviews = new PhotoViews();

//TODO Make those run from within the container
test.meta("testID", "photos-download-001")(
  "Test download jpg file from context menu and fullscreen",
  async (t) => {
    await toolbar.search("name:monochrome-2.jpg");
    const Photo = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", Photo, "select");
    await logger.clear();
    await contextmenu.triggerContextMenuAction("download", "", "");
    const requestInfo = await logger.requests[1].response;
    console.log(requestInfo);
    const requestInfo0 = await logger.requests[0].response;
    console.log(requestInfo0);
    await newpage.validateDownloadRequest(requestInfo, "monochrome-2", ".jpg");
    await logger.clear();
    await contextmenu.clearSelection();
    await toolbar.search("name:IMG_20200711_174006.jpg");
    const SecondPhoto = await photo.getNthPhotoUid("all", 0);
    await t.click(Selector("div").withAttribute("data-uid", SecondPhoto));
    await photoviewer.openPhotoViewer("uid", SecondPhoto);
    await logger.clear();
    await photoviewer.triggerPhotoViewerAction("download");
    await logger.clear();
    await photoviewer.triggerPhotoViewerAction("close");
  }
);

test.meta("testID", "photos-download-002")("Test download video from context menu", async (t) => {
  await toolbar.search("name:Mohn.mp4");
  const Photo = await photo.getNthPhotoUid("all", 0);
  await photoviews.triggerHoverAction("uid", Photo, "select");
  await logger.clear();
  await contextmenu.triggerContextMenuAction("download", "", "");
  const requestInfo = await logger.requests[0].response;
  console.log(requestInfo);
  const requestInfo2 = await logger.requests[1].response;
  await newpage.validateDownloadRequest(requestInfo, "Mohn", ".mp4.jpg");
  await newpage.validateDownloadRequest(requestInfo2, "Mohn", ".mp4");
  await logger.clear();
  await contextmenu.clearSelection();
});

test.meta("testID", "photos-download-003")(
  "Test download multiple jpg files from context menu",
  async (t) => {
    await toolbar.search("name:panorama_2.jpg");
    const Photo = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", Photo, "select");
    await toolbar.search("name:IMG_6478.JPG");
    const SecondPhoto = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", SecondPhoto, "select");
    await logger.clear();
    await contextmenu.triggerContextMenuAction("download", "", "");
    const requestInfo = await logger.requests[1].response;
    console.log(requestInfo);
    await newpage.validateDownloadRequest(requestInfo, "photoprism-download", ".zip");
    await logger.clear();
    await contextmenu.clearSelection();
  }
);

//TODO Check RAW files as well
test.meta("testID", "photos-download-004")(
  "Test raw file from context menu and fullscreen mode",
  async (t) => {
    await toolbar.search("name:elephantRAW");
    const Photo = await photo.getNthPhotoUid("all", 0);
    await photoviews.triggerHoverAction("uid", Photo, "select");
    await logger.clear();
    await contextmenu.triggerContextMenuAction("download", "", "");
    const requestInfo = await logger.requests[1].response;
    await newpage.validateDownloadRequest(requestInfo, "elephantRAW", ".JPG");
    await logger.clear();
    await contextmenu.clearSelection();
    await t.click(Selector("div").withAttribute("data-uid", Photo));
    await t.expect(Selector("#photo-viewer").visible).ok().hover(Selector(".action-download"));
    await logger.clear();
    await t.click(Selector(".action-download"));
    const requestInfo3 = await logger.requests[1].response;
    //const requestInfo4 = await logger.requests[2].response;
    await newpage.validateDownloadRequest(requestInfo3, "elephantRAW", ".JPG");
    //await page.validateDownloadRequest(requestInfo4, "elephantRAW", ".mp4");
    await logger.clear();
  }
);
