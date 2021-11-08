import { Selector } from "testcafe";
import testcafeconfig from "./testcafeconfig";
import Page from "./page-model";
import { RequestLogger } from "testcafe";

const logger = RequestLogger(/http:\/\/localhost:2343\/api\/v1\/*/, {
  logResponseHeaders: true,
  logResponseBody: true,
});

fixture`Test photos download`.page`${testcafeconfig.url}`.requestHooks(logger).skip;

const page = new Page();
//TODO Make those run from within the container
test.meta("testID", "photos-download-001")(
  "Test download jpg file from context menu and fullscreen",
  async (t) => {
    await page.search("name:monochrome-2.jpg");
    const Photo = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(Photo);
    await t.click(Selector("button.action-menu"));
    await logger.clear();
    await t.click(Selector("button.action-download"));
    const requestInfo = await logger.requests[1].response;
    console.log(requestInfo);
    const requestInfo0 = await logger.requests[0].response;
    console.log(requestInfo0);
    await page.validateDownloadRequest(requestInfo, "monochrome-2", ".jpg");
    await logger.clear();
    await page.clearSelection();
    await page.search("name:IMG_20200711_174006.jpg");
    const SecondPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await t.click(Selector("div").withAttribute("data-uid", SecondPhoto));
    await t.expect(Selector("#photo-viewer").visible).ok().hover(Selector(".action-download"));
    await logger.clear();
    await t.click(Selector(".action-download"));
    const requestInfo2 = await logger.requests[1].response;
    await page.validateDownloadRequest(requestInfo2, "IMG_20200711_174006", ".jpg");
    await logger.clear();
    await t.click(Selector(".action-close"));
  }
);

test.meta("testID", "photos-download-002")("Test download video from context menu", async (t) => {
  await page.search("name:Mohn.mp4");
  const Photo = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
  await page.selectPhotoFromUID(Photo);
  await t.click(Selector("button.action-menu"));
  await logger.clear();
  await t.click(Selector("button.action-download"));
  const requestInfo = await logger.requests[1].response;
  const requestInfo2 = await logger.requests[2].response;
  await page.validateDownloadRequest(requestInfo, "Mohn", ".mp4.jpg");
  await page.validateDownloadRequest(requestInfo2, "Mohn", ".mp4");
  await logger.clear();
  await page.clearSelection();
});

test.meta("testID", "photos-download-003")(
  "Test download multiple jpg files from context menu",
  async (t) => {
    await page.search("name:panorama_2.jpg");
    const Photo = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(Photo);
    await page.search("name:IMG_6478.JPG");
    const SecondPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(SecondPhoto);
    await t.click(Selector("button.action-menu"));
    await logger.clear();
    await t.click(Selector("button.action-download"));
    const requestInfo = await logger.requests[1].response;
    console.log(requestInfo);
    await page.validateDownloadRequest(requestInfo, "photoprism-download", ".zip");
    await logger.clear();
    await page.clearSelection();
  }
);

//TODO Check RAW files as well
test.meta("testID", "photos-download-004")(
  "Test raw file from context menu and fullscreen mode",
  async (t) => {
    await page.search("name:elephantRAW");
    const Photo = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
    await page.selectPhotoFromUID(Photo);
    await t.click(Selector("button.action-menu"));
    await logger.clear();
    await t.click(Selector("button.action-download"));
    const requestInfo = await logger.requests[1].response;
    //const requestInfo2 = await logger.requests[2].response;
    await page.validateDownloadRequest(requestInfo, "elephantRAW", ".JPG");
    //await page.validateDownloadRequest(requestInfo2, "elephantRAW", ".mp4");
    await logger.clear();
    await page.clearSelection();
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
