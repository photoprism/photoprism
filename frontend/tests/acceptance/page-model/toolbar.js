import { ClientFunction, RequestLogger, Selector, t } from "testcafe";
import Notifies from "../page-model/notifications";
import testcafeconfig from "../../testcafeconfig.json";
const notifies = new Notifies();

export default class Page {
  // constructor initializes the toolbar selectors.
  constructor() {
    this.view = Selector("div.p-view-select");
    this.camera = Selector("div.p-camera-select");
    this.countries = Selector("div.p-countries-select");
    this.sort = Selector("button.action-sort");
    this.sortMenu = Selector('.action-menu__list[aria-label="Sort Order"]').filterVisible();
    this.mediaType = Selector("div.p-media-type-select");
    this.searchFilters = Selector('.input-search [aria-label="Search Filters"]');
    this.sharedAlbumToolbar = Selector(".p-album-toolbar");
    this.search1 = Selector("div.input-search input");
    this.toolbarDescription = Selector(".toolbar-details-panel");
    this.toolbarTitle = Selector("#p-navigation div.v-toolbar-title");
    this.toolbarSecondTitle = Selector("header.v-toolbar div.v-toolbar-title");
    this.openMobileToolbar = Selector("button.mobile-menu-trigger");
    this.cardsViewAction = Selector("button.action-view-cards");
    this.mosaicViewAction = Selector("button.action-view-mosaic");
    this.listViewAction = Selector("button.action-view-list");
    this.searchClearAction = Selector("div.input-search .v-field__clearable", { timeout: 7000 });
    this.searchParam = ClientFunction((name) => new URLSearchParams(window.location.search).get(name));
  }

  // openPage navigates to a fixture route, including public share links.
  async openPage(path) {
    await t.navigateTo(new URL(path, testcafeconfig.url).href);
  }

  // openSearch opens a library route and waits for its initial results.
  async openSearch(path) {
    await this.openPage("/library/" + path);
    await notifies.waitForSearchToFinish(7000);
  }

  // openFilters expands the photo search filters.
  async openFilters() {
    if ((await this.searchFilters.getAttribute("aria-expanded")) !== "true") {
      await t.click(this.searchFilters);
    }
    await t.expect(this.searchFilters.getAttribute("aria-expanded")).eql("true");
  }

  // selectSort clicks a sort row or its direction button and waits for a changed search.
  async selectSort(value, { direction, search = true } = {}) {
    if (search) {
      await notifies.closeAllEventPopups();
    }
    const option = direction
      ? this.sortMenu.find(`button[data-sort="${value}"][data-direction="${direction}"]`)
      : this.sortMenu.find(`[role="button"][data-sort="${value}"]`);
    await t.click(this.sort).expect(this.sortMenu.exists).ok().click(option).expect(this.sortMenu.exists).notOk();
    if (search) {
      await notifies.waitForSearchToFinish(7000);
    }
  }

  // searchRequests records API search responses for the current photo or album results.
  searchRequests() {
    return RequestLogger(/\/api\/v1\/(photos|albums)\?/, { logResponseBody: true, stringifyResponseBody: true });
  }

  // latestRequest returns the latest successful search matching the expected query parameters.
  latestRequest(requests, params) {
    return requests.requests.findLast((entry) => this.matchesSearch(entry, params));
  }

  // matchesSearch reports whether the response belongs to the expected search.
  matchesSearch(entry, params) {
    const query = new URL(entry.request.url).searchParams;
    return entry.response?.statusCode === 200 && Object.entries(params).every(([name, value]) => query.get(name) === value);
  }

  // searchResults waits for the matching response and returns its results.
  async searchResults(requests, params) {
    await t.expect(requests.contains((entry) => this.matchesSearch(entry, params))).ok(JSON.stringify(params));
    return JSON.parse(this.latestRequest(requests, params).response.body);
  }

  // expectSort verifies the URL, visible sort label, and API query against the test's expected values.
  async expectSort(requests, { order, reverse = null, label }) {
    await t
      .expect(this.searchParam("order"))
      .eql(order)
      .expect(this.searchParam("reverse"))
      .eql(reverse)
      .expect(this.sort.getAttribute("title"))
      .eql(label);
    return this.searchResults(requests, { order, reverse });
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
    await notifies.closeAllEventPopups();

    await t.click(this.search1).typeText(this.search1, term, { replace: true }).pressKey("enter");
    if (wait) {
      await notifies.waitForSearchToFinish(7000);
    }
  }

  // setFilter selects an option from the expanded search filters.
  async setFilter(filter, option) {
    let field;

    switch (filter) {
      case "view":
        field = this.view;
        break;
      case "camera":
        field = this.camera;
        break;
      case "type":
        field = this.mediaType;
        break;
      case "countries":
        field = this.countries;
        break;
      case "category":
        field = Selector("div.p-category-select");
        break;
      default:
        throw "unknown filter";
    }
    if (filter === "type") {
      await notifies.closeAllEventPopups();
      await this.openFilters();
    } else if (!(await field.visible)) {
      await t.click(Selector("i.mdi-tune"));
    }
    await t.click(field);

    if (option) {
      const options = Selector('div[role="option"]');
      await t.click(filter === "type" ? options.withExactText(option) : options.withText(option));
    } else {
      await t.click(Selector('div[role="option"]').nth(1));
    }

    if (await field.visible) {
      await t.click(filter === "type" ? this.searchFilters : Selector("i.mdi-tune"));
    }
    if (filter === "type") {
      await notifies.waitForSearchToFinish(7000);
    }
  }
}
