import { Selector, t } from "testcafe";

export default class Page {
  constructor() {}

  async openNav() {
    if (await Selector("div.nav-expand").visible) {
      await t.click(Selector("div.nav-expand a"));
    } else if (await Selector("div.nav-expand").visible) {
      await t.click(Selector("div.nav-expand i"));
    }
  }

  async openPage(page) {
    await this.openNav();
    if (t.browser.platform === "mobile") {
      await t.wait(5000);
    }
    if (await Selector(".nav-" + page).visible) {
      await t.click(Selector(".nav-" + page));
    } else {
      if (
        (page === "monochrome") |
        (page === "panoramas") |
        (page === "stacks") |
        (page === "scans") |
        (page === "review") |
        (page === "archive")
      ) {
        if (!(await Selector("div.v-list__group--active div.nav-browse").visible)) {
          await t.click(Selector("div.nav-browse + div"));
        }
      } else if (page === "live") {
        if (!(await Selector("div.v-list__group--active div.nav-video").visible)) {
          await t.click(Selector("div.nav-video + div"));
        }
      } else if (page === "states") {
        if (!(await Selector("div.v-list__group--active div.nav-places").visible)) {
          await t.click(Selector("div.nav-places + div"));
        }
      } else if ((page === "originals") | (page === "hidden") | (page === "errors")) {
        if (!(await Selector("div.v-list__group--active div.nav-library").visible)) {
          await t.click(Selector("div.nav-library + div"));
        }
      } else if ((page === "about") | (page === "feedback") | (page === "license")) {
        if (!(await Selector("div.v-list__group--active div.nav-settings").visible)) {
          await t.click(Selector("div.nav-settings + div"));
        }
      }
      await t.click(Selector(".nav-" + page));
    }
  }

  async checkMenuItemAvailability(page, visible) {
    if (t.browser.platform === "mobile") {
      await t.wait(5000);
    }
    await this.openNav();
    if (t.browser.platform === "mobile") {
      await t.wait(5000);
    }
    if (
      (page === "monochrome") |
      (page === "panoramas") |
      (page === "stacks") |
      (page === "scans") |
      (page === "review") |
      (page === "archive")
    ) {
      if (
        !(await Selector("div.v-list__group--active div.nav-browse", { timeout: 15000 }).visible)
      ) {
        await t.click(Selector("div.nav-browse + div", { timeout: 15000 }));
      }
    } else if (page === "live") {
      if (await Selector(".nav-video").visible) {
        if (!(await Selector("div.v-list__group--active div.nav-video").visible)) {
          await t.click(Selector("div.nav-video + div"));
        }
      }
    } else if (page === "states") {
      if (await Selector(".nav-places").visible) {
        if (!(await Selector("div.v-list__group--active div.nav-places").visible)) {
          await t.click(Selector("div.nav-places + div"));
        }
      }
    } else if ((page === "originals") | (page === "hidden") | (page === "errors")) {
      if (await Selector(".nav-library").visible) {
        if (!(await Selector("div.v-list__group--active div.nav-library").visible)) {
          if (await Selector("div.nav-library + div").visible) {
            await t.click(Selector("div.nav-library + div"));
          }
        }
      }
    } else if ((page === "abouts") | (page === "feedback") | (page === "license")) {
      if (await Selector(".nav-settings").visible) {
        if (!(await Selector("div.v-list__group--active div.nav-settings").visible)) {
          await t.click(Selector("div.nav-settings + div"));
        }
      }
    }

    if (visible) {
      await t.expect(Selector(".nav-" + page).visible).ok();
    } else {
      await t.expect(Selector(".nav-" + page).visible).notOk();
    }
  }
}
