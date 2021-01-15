import { Selector, t } from "testcafe";

export default class Page {
  constructor() {
    this.view = Selector("div.p-view-select", { timeout: 15000 });
    this.camera = Selector("div.p-camera-select", { timeout: 15000 });
    this.countries = Selector("div.p-countries-select", { timeout: 15000 });
    this.time = Selector("div.p-time-select", { timeout: 15000 });
    this.search1 = Selector("div.input-search input", { timeout: 15000 });
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
      default:
        throw "unknown filter";
    }

    await t.click(filterSelector, { timeout: 15000 });

    if (option) {
      await t.click(Selector('div[role="listitem"]').withText(option), { timeout: 15000 });
    } else {
      await t.click(Selector('div[role="listitem"]').nth(1), { timeout: 15000 });
    }
  }

  async search(term) {
    await t.typeText(this.search1, term, { replace: true }).pressKey("enter");
  }

  async openNav() {
    if (await Selector("button.nav-show").exists) {
      await t.click(Selector("button.nav-show"));
    } else if (await Selector("div.nav-expand").exists) {
      await t.click(Selector("div.nav-expand i"));
    }
  }

  async selectFromUID(uid) {
    await t
      .hover(Selector("a").withAttribute("data-uid", uid))
      .click(Selector(`.uid-${uid} .input-select`));
  }

  async selectPhotoFromUID(uid) {
    await t
      .hover(Selector("div").withAttribute("data-uid", uid))
      .click(Selector(`.uid-${uid} .input-select`));
  }

  async selectFromUIDInFullscreen(uid) {
    await t.hover(Selector("div").withAttribute("data-uid", uid));
    if (await Selector(`.uid-${uid} .action-fullscreen`).exists) {
      await t.click(Selector(`.uid-${uid} .action-fullscreen`));
    } else {
      await t.click(Selector("div").withAttribute("data-uid", uid));
    }
    await t
      .expect(Selector("#p-photo-viewer").visible)
      .ok()
      .click(Selector('button[title="Select"]'))
      .click(Selector(".action-close", { timeout: 4000 }));
  }

  async toggleSelectNthPhoto(nPhoto) {
    await t
      .hover(Selector(".is-photo.type-image", { timeout: 4000 }).nth(nPhoto))
      .click(Selector(".is-photo.type-image .input-select").nth(nPhoto));
  }

  async toggleLike(uid) {
    await t.click(Selector(`.uid-${uid} .input-favorite`));
  }

  async archiveSelected() {
    await t
      .click(Selector("button.action-menu"))
      .click(Selector("button.action-archive"))
      .click(Selector("button.action-confirm"));
  }

  async restoreSelected() {
    await t.click(Selector("button.action-menu")).click(Selector("button.action-restore"));
  }

  async editSelected() {
    if (await Selector("button.action-edit").visible) {
      await t.click(Selector("button.action-edit"));
    } else if (await Selector("button.action-menu").exists) {
      await t.click(Selector("button.action-menu")).click(Selector("button.action-edit"));
    }
  }

  async deleteSelected() {
    await t
      .click(Selector("button.action-menu"))
      .click(Selector("button.action-delete"))
      .click(Selector("button.action-confirm"));
  }

  async removeSelected() {
    await t.click(Selector("button.action-menu")).click(Selector("button.action-delete"));
  }

  async addSelectedToAlbum(name) {
    await t
      .click(Selector("button.action-menu"))
      .click(Selector("button.action-album"))
      .typeText(Selector(".input-album input"), name, { replace: true })
      .pressKey("enter");
    if (await Selector('div[role="listitem"]').withText(name).visible) {
      await t.click(Selector('div[role="listitem"]').withText(name));
    }
    await t.click(Selector("button.action-confirm"));
  }

  async login(password) {
    await t.typeText(Selector('input[type="password"]'), password).pressKey("enter");
  }

  async logout() {
    await t.click(Selector("div.nav-logout"));
  }
}
