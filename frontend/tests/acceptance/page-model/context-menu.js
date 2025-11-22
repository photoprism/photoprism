import { Selector, t } from "testcafe";

export default class Page {
  constructor() {}

  async openContextMenu() {
    if (!(await Selector(".action-clear").visible)) {
      await t.click(Selector("button.action-menu"));
    }
  }

  async checkContextMenuCount(count) {
    const Count = await Selector("span.count-clipboard", { timeout: 5000 });
    await t.expect(Count.textContent).eql(count);
  }

  async checkContextMenuActionAvailability(action, visible) {
    await this.openContextMenu();
    if (visible) {
      await t
        .expect(Selector("#t-clipboard button.action-" + action).visible)
        .ok()
        .expect(Selector("#t-clipboard button.action-" + action).hasAttribute("disabled"))
        .notOk();
    } else {
      if (await Selector("#t-clipboard button.action-" + action).visible) {
        await t.expect(Selector("#t-clipboard button.action-" + action).hasAttribute("disabled")).ok();
      } else {
        await t.expect(Selector("#t-clipboard button.action-" + action).visible).notOk();
      }
    }
  }
  async triggerContextMenuAction(action, albumName) {
    await this.openContextMenu();
    if (t.browser.platform === "mobile") {
      await t.wait(5000);
    }
    await t.click(Selector("#t-clipboard button.action-" + action));
    if (action === "delete") {
      await t.click(Selector("button.action-confirm"));
    }
    if ((action === "album") || (action === "clone")) {
      await t.click(Selector(".input-albums"));

      // Handle single album name or array of album names
      const albumNames = Array.isArray(albumName) ? albumName : [albumName];

      for (const name of albumNames) {
        if (await Selector("div").withText(name).parent('div[role="option"]').visible) {
          // Click on the album option to select it
          await t
            .click(Selector("div").withText(name).parent('div[role="option"]'))
            .click(Selector("div i.mdi-bookmark"));
        } else {
          await t.typeText(Selector(".input-albums input"), name).click(Selector("div i.mdi-bookmark"));
        }
        await t.expect(Selector("span.v-chip").withText(name).visible).ok();
      }
      await t.click(Selector("button.action-confirm"));
    }
  }

  async clearSelection() {
    await this.openContextMenu();
    await t.click(Selector(".action-clear"));
  }
}
