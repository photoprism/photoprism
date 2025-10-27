import { Selector, t } from "testcafe";

export default class Page {
  constructor() {
    this.view = Selector("div.p-view-select");
    this.camera = Selector("div.p-camera-select");
    this.countries = Selector("div.p-countries-select");
    this.time = Selector("div.p-time-select");
    this.search1 = Selector("div.input-search input");
    this.toolbarDescription = Selector(".toolbar-details-panel");
    this.toolbarTitle = Selector("#p-navigation div.v-toolbar-title");
    this.toolbarSecondTitle = Selector("header.v-toolbar div.v-toolbar-title");
    this.openMobileToolbar = Selector("button.mobile-menu-trigger");
    this.cardsViewAction = Selector("button.action-view-cards");
    this.mosaicViewAction = Selector("button.action-view-mosaic");
    this.listViewAction = Selector("button.action-view-list");
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
      if (
        action === "delete-all" ||
        action === "view-mosaic" ||
        action === "view-list" ||
        action === "view-cards" ||
        action === "add" ||
        action === "show-hidden" ||
        action === "show-all" ||
        action === "show-important"
      ) {
        await t.expect(Selector("button.action-" + action).visible).ok();
      } else {
        await t.hover(Selector("button.action-menu__btn"));
        await t.expect(Selector("div.action-" + action).visible).ok();
      }
    } else {
      if (
        action === "delete-all" ||
        action === "view-mosaic" ||
        action === "view-list" ||
        action === "view-cards" ||
        action === "add" ||
        action === "show-hidden" ||
        action === "show-all" ||
        action === "show-important"
      ) {
        await t.expect(Selector("button.action-" + action).visible).notOk();
      } else {
        await t.hover(Selector("button.action-menu__btn"));
        await t.expect(Selector("div.action-" + action).visible).notOk();
      }
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
      (action !== "add")
    ) {
      if (await this.openMobileToolbar.exists) {
        await t.click(this.openMobileToolbar);
      }
      if (await this.openMobileToolbar.exists) {
        await t.click(this.openMobileToolbar);
      }
      await t.click(Selector("button.nav-menu-" + action));
    } else {
      if (
        action === "delete-all" ||
        action === "view-mosaic" ||
        action === "view-list" ||
        action === "view-cards" ||
        action === "add" ||
        action === "show-hidden"
      ) {
        await t.click(Selector("button.action-" + action));
      } else {
        await t.hover(Selector("button.action-menu__btn"));
        await t.click(Selector("div.action-" + action));
        if (action === "delete") {
          await t.click(Selector("button.action-confirm"));
        }
      }
    }
  }

  async search(term, wait = true) {
    // Wait for notifications to go away before we search
    while(await Selector(".p-notify__close", {timeout: 250}).visible) {
      await t.wait(250);
      // try {
      //   await t.click(Selector(".p-notify__close", {timeout: 250}));
      // } catch (e) {
      //   // Do Nothing
      //   console.log(e)
      // }
    }

    await t.click(this.search1).typeText(this.search1, term, { replace: true }).pressKey("enter"); // .wait(7000); <-- is breaking other tests.
    if (wait) {
      await this.waitForSearchToFinish(7000);
    }
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
      await t.click(Selector("i.mdi-tune"));
    }
    await t.click(filterSelector);

    if (option) {
      await t.click(Selector('div[role="option"]').withText(option));
    } else {
      await t.click(Selector('div[role="option"]').nth(1));
    }

    if (await Selector(filterSelector).visible) {
      await t.click(Selector("i.mdi-tune"));
    }
  }

  async waitForSearchToFinish(delay, close = true){
    // console.time("waitForSearchToFinish")
    if ((await Selector("div.p-notify__text", {timeout: delay}).withText(/(found|contain|empty)/).visible) && close){
      try {
        await t.click(Selector(".p-notify__close", {timeout: 250}));
      } catch (e) {
        // ignore the error as the item may not show up
        // console.log(".p-notify__close missed in waitForSearchToFinish")
      }
    }
    // console.timeEnd("waitForSearchToFinish")
  }
}
