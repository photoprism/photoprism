import { Selector, t } from "testcafe";

export default class Page {
  constructor() {
    this.view = Selector("div.p-view-select");
    this.camera = Selector("div.p-camera-select");
    this.countries = Selector("div.p-countries-select");
    this.time = Selector("div.p-time-select");
    this.search1 = Selector("div.input-search input");
    this.toolbarDescription = Selector(".v-card__text").nth(0);
    this.toolbarTitle = Selector("#p-navigation div.v-toolbar__title");
    this.toolbarSecondTitle = Selector("main.v-content div.v-toolbar__title");
    this.openMobileToolbar = Selector("button.mobile-menu-trigger");
  }

  async checkToolbarActionAvailability(action, visible) {
    if (
      (t.browser.platform === "mobile") &
      (action !== "edit") &
      (action !== "share") &
      (action !== "add") &
      (action !== "show-all") &
      (action !== "show-important")
    ) {
      if (await this.openMobileToolbar.exists) {
        await t.click(this.openMobileToolbar);
      }
      await this.checkMobileMenuActionAvailability(action, visible);
      await t.click(Selector("#photoprism"), { offsetX: 1, offsetY: 1 });
    } else if (visible) {
      await t.expect(Selector("nav.v-toolbar button.action-" + action).visible).ok();
    } else {
      await t.expect(Selector("nav.v-toolbar button.action-" + action).visible).notOk();
    }
  }

  async checkMobileMenuActionAvailability(action, visible) {
    if (
      (action !== "login") &
      (action !== "logout") &
      (action !== "reload") &
      (action !== "logs") &
      (action !== "upload") &
      (action !== "settings")
    ) {
      if (visible) {
        await t.expect(Selector("#mobile-menu div.nav-" + action).visible).ok();
      } else {
        await t.expect(Selector("#mobile-menu div.nav-" + action).visible).notOk();
      }
    } else {
      if (visible) {
        await t.expect(Selector("#mobile-menu a.nav-" + action).visible).ok();
      } else {
        await t.expect(Selector("#mobile-menu a.nav-" + action).visible).notOk();
      }
    }
  }

  async triggerMobileMenuAction(action) {
    if (
      (action !== "login") &
      (action !== "logout") &
      (action !== "reload") &
      (action !== "logs") &
      (action !== "upload") &
      (action !== "settings")
    ) {
      await t.click(Selector("#mobile-menu div.nav-" + action + " a"));
    } else {
      await t.click(Selector("#mobile-menu a.nav-" + action));
    }
  }

  async triggerToolbarAction(action) {
    if (
      (t.browser.platform === "mobile") &
      (action !== "edit") &
      (action !== "share") &
      (action !== "add") &
      (action !== "show-all") &
      (action !== "show-important")
    ) {
      if (await this.openMobileToolbar.exists) {
        await t.click(this.openMobileToolbar);
      }
      if (await this.openMobileToolbar.exists) {
        await t.click(this.openMobileToolbar);
      }
      await t.click(Selector("button.nav-menu-" + action));
    } else {
      await t.click(Selector("nav.v-toolbar button.action-" + action));
    }
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
        filterSelector = "div.p-category-select";
        break;
      default:
        throw "unknown filter";
    }
    if (!(await Selector(filterSelector).visible)) {
      await t.click(Selector(".p-expand-search"));
    }
    await t.click(filterSelector);

    if (option) {
      await t.click(Selector('div[role="listitem"]').withText(option));
    } else {
      await t.click(Selector('div[role="listitem"]').nth(1));
    }
  }
}
