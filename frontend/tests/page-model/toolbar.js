import { Selector, t } from "testcafe";

export default class Page {
  constructor() {
    this.view = Selector("div.p-view-select");
    this.camera = Selector("div.p-camera-select");
    this.countries = Selector("div.p-countries-select");
    this.time = Selector("div.p-time-select");
    this.search1 = Selector("div.input-search input");
    this.toolbarDescription = Selector(".v-card__text").nth(0);
    this.toolbarTitle = Selector("div.v-toolbar__title");
  }

  async checkToolbarActionAvailability(action, visible) {
    if (visible) {
      await t.expect(Selector("nav.v-toolbar button.action-" + action).visible).ok();
    } else {
      await t.expect(Selector("nav.v-toolbar button.action-" + action).visible).notOk();
    }
  }

  async triggerToolbarAction(action) {
    await t.click(Selector("nav.v-toolbar button.action-" + action));
  }

  async toggleFilterBar() {
    await t.click(Selector("nav.page-toolbar button.action-expand-search"));
  }

  async search(term) {
    await t.typeText(this.search1, term, { replace: true }).pressKey("enter");
  }

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
      case "category":
        filterSelector = ".input-category";
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
}
