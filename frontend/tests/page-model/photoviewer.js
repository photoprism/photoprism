import { Selector, t } from "testcafe";
import { RequestLogger } from "testcafe";

const logger = RequestLogger(/http:\/\/localhost:2343\/api\/v1\/*/, {
  logResponseHeaders: true,
  logResponseBody: true,
});

export default class Page {
  constructor() {
    this.view = Selector("div.p-view-select", { timeout: 15000 });
    this.camera = Selector("div.p-camera-select", { timeout: 15000 });
    this.countries = Selector("div.p-countries-select", { timeout: 15000 });
    this.time = Selector("div.p-time-select", { timeout: 15000 });
    this.search1 = Selector("div.input-search input", { timeout: 15000 });
  }

  async openPhotoViewer(mode, uidOrNth) {
    if (mode === "uid") {
      await t.hover(Selector("div.is-photo").withAttribute("data-uid", uidOrNth));
      if (await Selector(`.uid-${uidOrNth} .action-fullscreen`).visible) {
        await t.click(Selector(`.uid-${uidOrNth} .action-fullscreen`));
      } else {
        await t.click(Selector("div.is-photo").withAttribute("data-uid", uidOrNth));
      }
    } else if (mode === "nth") {
      await t.hover(Selector("div.is-photo").nth(uidOrNth));
      if (await Selector(`div.is-photo .action-fullscreen`).visible) {
        await t.click(Selector(`div.is-photo .action-fullscreen`));
      } else {
        await t.click(Selector("div.is-photo").nth(uidOrNth));
      }
    }
    await t.expect(Selector("#photo-viewer").visible).ok();
  }

  async checkPhotoViewerActionAvailability(action, visible) {
    if (visible) {
      await t.expect(Selector("div.pswp__top-bar button.action-" + action).visible).ok();
    } else {
      await t.expect(Selector("div.pswp__top-bar button.action-" + action).visible).notOk();
    }
  }

  //trigger fullscreen menu action (edit/like/close) --fullscreen
  async triggerPhotoViewerAction(action) {
    await t.click(Selector("div.pswp__top-bar button.action-" + action));
  }

}
