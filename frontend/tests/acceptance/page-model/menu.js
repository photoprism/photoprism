import { Selector, t } from "testcafe";

export default class Page {
  constructor() {}

  async openNav() {
    if (await Selector("div.nav-expand").visible) {
      await t.click(Selector("div.nav-expand i"));
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
        (page === "private") |
        (page === "archive")
      ) {
        if (!(await Selector("div.v-list-group--open a.nav-browse").visible)) {
          await t.click(Selector("div.nav-browse .mdi-chevron-down"));
        }
      } else if (page === "live") {
        if (!(await Selector("div.v-list-group--open a.nav-video").visible)) {
          await t.click(Selector("div.nav-video .mdi-chevron-down"));
        }
      } else if (page === "states") {
        if (!(await Selector("div.v-list-group--open a.nav-places").visible)) {
          await t.click(Selector("div.nav-places .mdi-chevron-down"));
        }
      } else if ((page === "originals") | (page === "hidden") | (page === "errors")) {
        if (!(await Selector("div.v-list-group--open a.nav-library").visible)) {
          await t.click(Selector("div.nav-library .mdi-chevron-down"));
        }
      } else if ((page === "about") | (page === "feedback") | (page === "license")) {
        if (!(await Selector("div.v-list-group--open a.nav-settings").visible)) {
          await t.click(Selector("div.nav-settings .mdi-chevron-down"));
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
      (page === "private") |
      (page === "archive")
    ) {
      if (
        !(await Selector("div.v-list-group--open div.nav-browse", { timeout: 15000 }).visible) &
        (await Selector("div.nav-browse .mdi-chevron-down", { timeout: 15000 }).visible)
      ) {
        await t.click(Selector("div.nav-browse .mdi-chevron-down", { timeout: 15000 }));
      }
    } else if (page === "live") {
      if (await Selector(".nav-video").visible) {
        if (
          !(await Selector("div.v-list-group--open div.nav-video").visible) &
          (await Selector("div.nav-video .mdi-chevron-down", { timeout: 15000 }).visible)
        ) {
          await t.click(Selector("div.nav-video .mdi-chevron-down"));
        }
      }
    } else if (page === "states") {
      if (await Selector(".nav-places").visible) {
        if (
          !(await Selector("div.v-list-group--open div.nav-places").visible) &
          (await Selector("div.nav-places .mdi-chevron-down", { timeout: 15000 }).visible)
        ) {
          await t.click(Selector("div.nav-places .mdi-chevron-down"));
        }
      }
    } else if ((page === "originals") | (page === "hidden") | (page === "errors")) {
      if (await Selector(".nav-library").visible) {
        if (!(await Selector("div.v-list-group--open div.nav-library").visible)) {
          if (await Selector("div.nav-library .mdi-chevron-down").visible) {
            await t.click(Selector("div.nav-library .mdi-chevron-down"));
          }
        }
      }
    } else if ((page === "abouts") | (page === "feedback") | (page === "license")) {
      if (await Selector(".nav-settings").visible) {
        if (
          !(await Selector("div.v-list-group--open div.nav-settings").visible) &
          (await Selector("div.nav-settings .mdi-chevron-down", { timeout: 15000 }).visible)
        ) {
          await t.click(Selector("div.nav-settings .mdi-chevron-down"));
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
