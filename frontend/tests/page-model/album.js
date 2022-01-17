import { Selector, t } from "testcafe";
import { RequestLogger } from "testcafe";

const logger = RequestLogger(/http:\/\/localhost:2343\/api\/v1\/*/, {
  logResponseHeaders: true,
  logResponseBody: true,
});

export default class Page {
  constructor() {
  }

  async getNthAlbumUid(type, nth) {
    if (type === "all") {
      const NthAlbum = await Selector("a.is-album").nth(nth).getAttribute("data-uid");
      return NthAlbum;
    } else {
      const NthAlbum = await Selector("a.type-" + type)
        .nth(nth)
        .getAttribute("data-uid");
      return NthAlbum;
    }
  }

  async getAlbumCount(type) {
    if (type === "all") {
      const AlbumCount = await Selector("a.is-album", { timeout: 5000 }).count;
      return AlbumCount;
    } else {
      const AlbumCount = await Selector("a.type-" + type, { timeout: 5000 }).count;
      return AlbumCount;
    }
  }

  async selectAlbumFromUID(uid) {
    await t
      .hover(Selector("a.is-album").withAttribute("data-uid", uid))
      .click(Selector(`.uid-${uid} .input-select`));
  }

  async toggleSelectNthAlbum(nth, type) {
    if (type === "all") {
      await t
        .hover(Selector("a.is-album", { timeout: 4000 }).nth(nth))
        .click(Selector("a.is-album .input-select").nth(nth));
    } else {
      await t
        .hover(Selector("a.type-" + type, { timeout: 4000 }).nth(nth))
        .click(Selector("a.type-" + type + " .input-select").nth(nth));
    }
  }
}
