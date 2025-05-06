import { Selector, t } from "testcafe";

export default class Page {
  constructor() {
    this.view = Selector("div.p-view-select", { timeout: 15000 });
    this.camera = Selector("div.p-camera-select", { timeout: 15000 });
    this.countries = Selector("div.p-countries-select", { timeout: 15000 });
    this.time = Selector("div.p-time-select", { timeout: 15000 });
    this.search1 = Selector("div.input-search input", { timeout: 15000 });
    this.menuButton = Selector("button.pswp__button--menu-button", { timeout: 15000 });
    this.viewer = Selector("div.p-lightbox__pswp", { timeout: 15000 });
    this.caption = Selector("div.pswp__caption__center", { timeout: 5000 });
    this.muteButton = Selector("button.pswp__button--mute", { timeout: 5000 });
    this.playButton = Selector('[class^="pswp__button pswp__button--slideshow-toggle pswp__"]', { timeout: 5000 });
    this.favoriteOnIcon = Selector("button.action-favorite i.icon-favorite", { timeout: 5000 });
    this.favoriteOffIcon = Selector("button.action-favorite i.icon-favorite-border", { timeout: 5000 });
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
    if (action === "cover") {
      await t.hover(this.menuButton);
      if (visible) {
        await t.expect(Selector("div.action-" + action).visible).ok();
      } else {
        await t.expect(Selector("div.action-" + action).visible).notOk();
      }
    } else if (action === "download") {
      await t.hover(this.menuButton);
      if (visible) {
        await t.expect(Selector("div.action-" + action).visible).ok();
      } else {
        await t.expect(Selector("div.action-" + action).visible).notOk();
      }
    } else {
      if (visible) {
        await t.expect(Selector("button.pswp__button--" + action).visible).ok();
      } else {
        await t.expect(Selector("button.pswp__button--" + action).visible).notOk();
      }
    }
  }

  async triggerPhotoViewerAction(action) {
    if (action === "cover") {
      await t.hover(this.menuButton);
      await t.click(Selector("div.action-" + action));
    } else if (action === "download") {
      await t.hover(this.menuButton);
      await t.click(Selector("div.action-" + action));
    } else {
      await t.hover(Selector("button.pswp__button--" + action));
      await t.click(Selector("button.pswp__button--" + action));
    }
    if (t.browser.platform === "mobile") {
      await t.wait(5000);
    }
  }
}
