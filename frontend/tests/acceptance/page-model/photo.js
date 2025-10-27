import { Selector, t } from "testcafe";
import Toolbar from "./toolbar";

const toolbar = new Toolbar();

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

  async getPhotoCount(type, delay = 7000) {
    await this.waitForPhotosToLoad(delay, true)
    if (type === "all") {
      const PhotoCount = await Selector("div.is-photo", { timeout: 2000 }).count;
      return PhotoCount;
    } else {
      const PhotoCount = await Selector("div.type-" + type, { timeout: 2000 }).count;
      return PhotoCount;
    }
  }

  async selectPhotoFromUID(uid) {
    await t.hover(Selector("div.is-photo").withAttribute("data-uid", uid)).click(Selector(`.uid-${uid} .input-select`));
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

  async checkHoverActionAvailability(mode, uidOrNth, action, visible) {
    if (mode === "uid") {
      await t.hover(Selector("div.is-photo").withAttribute("data-uid", uidOrNth));
      if (visible) {
        await t.expect(Selector(`.uid-${uidOrNth} .input-` + action).visible).ok();
      } else {
        await t.expect(Selector(`.uid-${uidOrNth} .input-` + action).visible).notOk();
      }
    }
    if (mode === "nth") {
      await t.hover(Selector("div.is-photo").nth(uidOrNth));
      if (visible) {
        await t.expect(Selector(`div.is-photo .input-` + action).visible).ok();
      } else {
        await t.expect(Selector(`div.is-photo .input-` + action).visible).notOk();
      }
    }
  }

  async triggerHoverAction(mode, uidOrNth, action) {
    if (mode === "uid") {
      await t.hover(Selector("div.is-photo").withAttribute("data-uid", uidOrNth, { timeout: 7000 }));
      await t.click(Selector(`div.uid-${uidOrNth} .input-` + action));
    }
    if (mode === "nth") {
      await t.hover(Selector("div.is-photo").nth(uidOrNth));
      await t.click(Selector(`div.is-photo .input-` + action).nth(uidOrNth));
    }
  }

  async checkHoverActionState(mode, uidOrNth, action, set) {
    if (mode === "uid") {
      await t.hover(Selector("div.is-photo").withAttribute("data-uid", uidOrNth));
      if (set) {
        await t.expect(Selector(`div.uid-${uidOrNth}`).hasClass("is-" + action)).ok();
      } else {
        await t.expect(Selector(`div.uid-${uidOrNth}`).hasClass("is-" + action)).notOk();
      }
    }
    if (mode === "nth") {
      await t.hover(Selector("div.is-photo").nth(uidOrNth));
      if (set) {
        await t
          .expect(
            Selector("div.is-photo")
              .nth(uidOrNth)
              .hasClass("is-" + action)
          )
          .ok();
      } else {
        await t
          .expect(
            Selector("div.is-photo")
              .nth(uidOrNth)
              .hasClass("is-" + action)
          )
          .notOk();
      }
    }
  }

  async triggerListViewActions(mode, uidOrnth, action) {
    if (mode === "nth") {
      await t.click(Selector(`td button.input-` + action).nth(uidOrnth));
    } else if (mode === "uid") {
      await t.click(Selector(`td button.input-` + action).withAttribute("data-uid", uidOrnth));
    }
  }

  async checkListViewActionAvailability(action, disabled, visible) {
    if (visible) {
      await t.expect(Selector(`td button.input-` + action).visible).ok();
      if (disabled) {
        await t.expect(Selector(`td button.input-` + action).hasAttribute("disabled")).ok();
      } else {
        await t.expect(Selector(`td button.input-` + action).hasAttribute("disabled")).notOk();
      }
    } else {
      await t.expect(Selector(`td button.input-` + action).visible).notOk();
    }
  }

  getPhotoCardPreviewSelector(uid) {
    return Selector(`div.is-photo[data-uid="${uid}"] .preview`);
  }

  async getPhotoPreviewStyle(uid) {
    const selector = this.getPhotoCardPreviewSelector(uid);
    const style = await selector.getStyleProperty("background-image");
    return style;
  }

  async waitForPhotosToLoad(delay, close = true){
    // console.time("waitForPhotosToLoad")

    while(await Selector(".p-notify__close", {timeout: 250}).visible) {
      if (await Selector("div.p-notify__text", {timeout: 50}).withText(/(picture|pictures) found/).visible) {
        break
      }
      try {
        await t.click(Selector(".p-notify__close", {timeout: 250}));
      } catch {
        // console.log(".p-notify__close missed in waitForPhotosToLoad Pre")
      }
    }

    if ((await Selector("div.p-notify__text", {timeout: delay}).withText(/(picture|pictures) found/).visible) && close){
      try {
        await t.click(Selector(".p-notify__close", {timeout: 250}));
      } catch (e) {
        // ignore the error as the item may not show up
        // console.log(".p-notify__close missed in waitForPhotosToLoad")
      }
    }
    // console.timeEnd("waitForPhotosToLoad")
  }

}
