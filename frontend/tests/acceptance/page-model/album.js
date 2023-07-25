import { Selector, t } from "testcafe";

export default class Page {
  constructor() {}

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
      if (t.browser.platform === "mobile") {
        const AlbumCount = await Selector("a.is-album", { timeout: 8000 }).count;
        return AlbumCount;
      } else {
        const AlbumCount = await Selector("a.is-album", { timeout: 5000 }).count;
        return AlbumCount;
      }
    } else {
      if (t.browser.platform === "mobile") {
        const AlbumCount = await Selector("a.type-" + type, { timeout: 8000 }).count;
        return AlbumCount;
      } else {
        const AlbumCount = await Selector("a.type-" + type, { timeout: 5000 }).count;
        return AlbumCount;
      }
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

  async openNthAlbum(nth, type) {
    if (!type || type === "all") {
      if (await Selector("a.is-album").nth(nth).visible) {
        await t.click(Selector("a.is-album").nth(nth));
      }
    } else {
      if (await Selector("a.type-" + type).nth(nth).visible) {
        await t.click(Selector("a.type-" + type).nth(nth));
      }
    }
  }

  async openAlbumWithUid(uid) {
    await t.click(Selector("a.is-album").withAttribute("data-uid", uid));
  }

  async checkAlbumVisibility(uid, visible) {
    if (visible) {
      await t.expect(Selector("a").withAttribute("data-uid", uid).visible).ok();
    } else {
      await t.expect(Selector("a").withAttribute("data-uid", uid).visible).notOk();
    }
  }

  async triggerHoverAction(mode, uidOrNth, action) {
    if (mode === "uid") {
      if (action === "share") {
        await t.hover(Selector("a.uid-" + uidOrNth));
        await t.click(Selector("a.uid-" + uidOrNth + " .action-" + action));
      } else {
        await t.hover(Selector("a.uid-" + uidOrNth));
        await t.click(Selector("a.uid-" + uidOrNth + " .input-" + action));
      }
    }
    if (mode === "nth") {
      await t.hover(Selector("a.is-album").nth(uidOrNth));
      await t.click(Selector(`.input-` + action));
    }
  }

  async checkHoverActionAvailability(mode, uidOrNth, action, visible) {
    if (mode === "uid") {
      await t.hover(Selector("a.is-album").withAttribute("data-uid", uidOrNth));
      if (visible) {
        await t.expect(Selector(`.uid-${uidOrNth} .input-` + action).visible).ok();
      } else {
        await t.expect(Selector(`.uid-${uidOrNth} .input-` + action).visible).notOk();
      }
    }
    if (mode === "nth") {
      await t.hover(Selector("a.is-album").nth(uidOrNth));
      if (action != "share") {
        if (visible) {
          await t.expect(Selector(`button.input-` + action).visible).ok();
        } else {
          await t.expect(Selector(`button.input-` + action).visible).notOk();
        }
      } else {
        if (visible) {
          await t.expect(Selector(`button.action-` + action).visible).ok();
        } else {
          await t.expect(Selector(`button.action-` + action).visible).notOk();
        }
      }
    }
  }

  async checkHoverActionState(mode, uidOrNth, action, set) {
    if (mode === "uid") {
      await t.hover(Selector("a").withAttribute("data-uid", uidOrNth));
      if (set) {
        await t.expect(Selector(`a.uid-${uidOrNth}`).hasClass("is-" + action)).ok();
      } else {
        await t.expect(Selector(`a.uid-${uidOrNth}`).hasClass("is-" + action)).notOk();
      }
    }
    if (mode === "nth") {
      await t.hover(Selector("a.is-album").nth(uidOrNth));
      if (set) {
        await t
          .expect(
            Selector("a.is-album")
              .nth(uidOrNth)
              .hasClass("is-" + action)
          )
          .ok();
      } else {
        await t
          .expect(
            Selector("a.is-album")
              .nth(uidOrNth)
              .hasClass("is-" + action)
          )
          .notOk();
      }
    }
  }

  async checkSortOptions(type) {
    await t.click(Selector(".p-expand-search"));
    await t.click(Selector(".p-sort-select"));
    await t.expect(Selector(".v-menu__content:first-of-type").visible).ok();

    const sortOptionsCount = Selector(".v-menu__content:first-of-type div[role=listitem]").count;

    if (type === "album") {
      await t.expect(sortOptionsCount).eql(7);
    } else {
      await t.expect(sortOptionsCount).eql(6);
    }
  }
}
