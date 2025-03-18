import { Selector, t } from "testcafe";

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
      if (await Selector(`.uid-${uidOrNth} button.input-open`).visible) {
        await t.click(Selector(`.uid-${uidOrNth} button.input-open`));
      } else {
        await t.click(Selector("div.is-photo").withAttribute("data-uid", uidOrNth));
      }
    } else if (mode === "nth") {
      await t.hover(Selector("div.is-photo").nth(uidOrNth));
      if (await Selector(`div.is-photo button.input-open`).visible) {
        await t.click(Selector(`div.is-photo button.input-open`));
      } else {
        await t.click(Selector("div.is-photo").nth(uidOrNth));
      }
    }
    await t.expect(Selector("div.p-lightbox__pswp").visible).ok();
  }

  async checkPhotoViewerActionAvailability(action, visible) {
    if (visible) {
      await t.expect(Selector("button.pswp__button--" + action).visible).ok();
    } else {
      await t.expect(Selector("button.pswp__button--" + action).visible).notOk();
    }
  }

  async triggerPhotoViewerAction(action) {
    await t.hover(Selector("button.pswp__button--" + action));
    await t.click(Selector("button.pswp__button--" + action));
    if (t.browser.platform === "mobile") {
      await t.wait(5000);
    }
  }
}
