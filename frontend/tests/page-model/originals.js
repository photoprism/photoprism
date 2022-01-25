import { Selector, t } from "testcafe";

export default class Page {
  constructor() {}

  async getNthFolderUid(nth) {
    const NthFolder = await Selector("div.is-folder").nth(nth).getAttribute("data-uid");
    return NthFolder;
  }

  async getNthFileUid(nth) {
    const NthFile = await Selector("div.is-file").nth(nth).getAttribute("data-uid");
    return NthFile;
  }

  async getNthItemUid(nth) {
    const NthItem = await Selector("div.result").nth(nth).getAttribute("data-uid");
    return NthItem;
  }

  async getFolderCount() {
    const FolderCount = await Selector("div.is-folder", { timeout: 8000 }).count;
    return FolderCount;
  }

  async getFileCount() {
    const FileCount = await Selector("div.is-file", { timeout: 5000 }).count;
    return FileCount;
  }

  async selectFolderFromUID(uid) {
    await t
      .hover(Selector("div.is-folder").withAttribute("data-uid", uid))
      .click(Selector(`.uid-${uid} .input-select`));
  }

  async toggleSelectNthFolder(nth) {
    await t
      .hover(Selector("div.is-folder", { timeout: 4000 }).nth(nth))
      .click(Selector("div.is-folder .input-select").nth(nth));
  }

  async openNthFolder(nth) {
    await t.click(Selector("div.is-folder").nth(nth)).expect(Selector("div.is-photo").visible).ok();
  }

  async openFolderWithUid(uid) {
    await t.click(Selector("div.is-folder").withAttribute("data-uid", uid));
  }

  async checkHoverActionAvailability(type, mode, uidOrNth, action, visible) {
    if (mode === "uid") {
      await t.hover(Selector("div." + type).withAttribute("data-uid", uidOrNth));
      if (visible) {
        await t.expect(Selector(`.uid-${uidOrNth} .input-` + action).visible).ok();
      } else {
        await t.expect(Selector(`.uid-${uidOrNth} .input-` + action).visible).notOk();
      }
    }
    if (mode === "nth") {
      await t.hover(Selector("div." + type).nth(uidOrNth));
      if (visible) {
        await t.expect(Selector(`div.input-` + action).visible).ok();
      } else {
        await t.expect(Selector(`div.input-` + action).visible).notOk();
      }
    }
  }

  async triggerHoverAction(type, mode, uidOrNth, action) {
    if (mode === "uid") {
      await t.hover(Selector("div." + type).withAttribute("data-uid", uidOrNth));
      await t.click(Selector(`div.uid-${uidOrNth} .input-` + action));
    }
    if (mode === "nth") {
      await t.hover(Selector("div." + type).nth(uidOrNth));
      await t.click(Selector(`div.input-` + action));
    }
  }

}
