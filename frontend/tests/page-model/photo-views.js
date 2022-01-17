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
        await t.expect(Selector(`div.input-` + action).visible).ok();
      } else {
        await t.expect(Selector(`div.input-` + action).visible).notOk();
      }
    }
  }

  async triggerHoverAction(mode, uidOrNth, action) {
    if (mode === "uid") {
      await t.hover(Selector("div.is-photo").withAttribute("data-uid", uidOrNth));
      await t.click(Selector(`div.uid-${uidOrNth} .input-` + action));
    }
    if (mode === "nth") {
      await t.hover(Selector("div.is-photo").nth(uidOrNth));
      await t.click(Selector(`div.input-` + action));
    }
  }

  async triggerListViewActions(nth, action) {
    await t.click(Selector(`td button.input-` + action).nth(nth));
  }

  async checkListViewActionAvailability(action, disabled) {
    if (disabled) {
      await t.expect(Selector(`td button.input-` + action).hasAttribute("disabled")).ok();
    } else {
      await t.expect(Selector(`td button.input-` + action).hasAttribute("disabled")).notOk();
    }
  }
}
