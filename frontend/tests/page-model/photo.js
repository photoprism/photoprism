import { Selector, t } from "testcafe";

export default class Page {
  constructor() {}

  async getNthPhotoUid(type, nth) {
    if (type === "all") {
      const NthPhoto = await Selector("div.is-photo").nth(nth).getAttribute("data-uid");
      return NthPhoto;
    } else {
      const NthPhoto = await Selector("div.type-" + type)
        .nth(nth)
        .getAttribute("data-uid");
      return NthPhoto;
    }
  }

  async getPhotoCount(type) {
    if (type === "all") {
      const PhotoCount = await Selector("div.is-photo", { timeout: 2000 }).count;
      return PhotoCount;
    } else {
      const PhotoCount = await Selector("div.type-" + type, { timeout: 2000 }).count;
      return PhotoCount;
    }
  }

  async selectPhotoFromUID(uid) {
    await t
      .hover(Selector("div.is-photo").withAttribute("data-uid", uid))
      .click(Selector(`.uid-${uid} .input-select`));
  }

  async toggleSelectNthPhoto(nPhoto, type) {
    if (type === "all") {
      await t
        .hover(Selector(".is-photo", { timeout: 4000 }).nth(nPhoto))
        .click(Selector(".is-photo .input-select").nth(nPhoto));
    } else {
      await t
        .hover(Selector("div.type-" + type, { timeout: 4000 }).nth(nPhoto))
        .click(Selector("div.type-" + type + " .input-select").nth(nPhoto));
    }
  }

  async checkPhotoVisibility(uid, visible) {
    if (visible) {
      await t.expect(Selector("div.is-photo").withAttribute("data-uid", uid).exists).ok();
    } else {
      await t.expect(Selector("div.is-photo").withAttribute("data-uid", uid).exists).notOk();
    }
  }
}
