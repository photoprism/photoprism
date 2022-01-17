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

  async toggleFilterBar() {
    await t.click(Selector("nav.page-toolbar button.action-expand-search"));
  }

  //check availability
  async checkToolbarActionAvailability(action, visible) {
    if (visible) {
      await t.expect(Selector("nav.page-toolbar button.action-" + action).visible).ok();
    } else {
      await t.expect(Selector("nav.page-toolbar button.action-" + action).visible).notOk();
    }
  }

  //click trigger action on toolbar -- toolbar --album (reload, switch view, upload, add album, hide/show,
  async triggerToolbarAction(action, name) {
    await t.click(Selector("nav.page-toolbar button.action-" + action));
    if (action === "add") {
      await t.typeText(Selector(".input-album input"), name, { replace: true }).pressKey("enter");
    }
  }

  //set filter --toolbar category as well
  //TODO refactor
  async setFilter(filter, option) {
    let filterSelector = "";

    switch (filter) {
      case "view":
        filterSelector = "div.p-view-select";
        break;
      case "camera":
        filterSelector = "div.p-camera-select";
        break;
      case "time":
        filterSelector = "div.p-time-select";
        break;
      case "countries":
        filterSelector = "div.p-countries-select";
        break;
      default:
        throw "unknown filter";
    }
    if (!(await Selector(filterSelector).visible)) {
      await t.click(Selector(".p-expand-search"));
    }
    await t.click(filterSelector, { timeout: 15000 });

    if (option) {
      await t.click(Selector('div[role="listitem"]').withText(option), { timeout: 15000 });
    } else {
      await t.click(Selector('div[role="listitem"]').nth(1), { timeout: 15000 });
    }
  }

  //search --toolbar
  async search(term) {
    await t
      .typeText(this.search1, term, { replace: true })
      .pressKey("enter")
      //TODO remove wait
      //.wait(10000)
      .expect(this.search1.value, { timeout: 15000 })
      .contains(term);
  }
}
