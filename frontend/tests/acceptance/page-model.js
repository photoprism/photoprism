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

  async search(term) {
    await t
      .typeText(this.search1, term, { replace: true })
      .pressKey("enter")
      .wait(10000)
      .expect(this.search1.value)
      .contains(term);
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
      .expect(Selector("#photo-viewer").visible)
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
    if (!(await Selector("#t-clipboard button.action-archive").visible)) {
      await t.click(Selector("button.action-menu", { timeout: 5000 }));
    }
    if (t.browser.platform === "mobile") {
      if (!(await Selector("#t-clipboard button.action-archive").visible)) {
        await t.click(Selector("button.action-menu", { timeout: 5000 }));
        if (!(await Selector("#t-clipboard button.action-archive").visible)) {
          await t.click(Selector("button.action-menu", { timeout: 5000 }));
        }
        if (!(await Selector("#t-clipboard button.action-archive").visible)) {
          await t.click(Selector("button.action-menu", { timeout: 5000 }));
        }
      }
    }
    await t.click(Selector("#t-clipboard button.action-archive", { timeout: 5000 }));
  }

  async privateSelected() {
    await t.click(Selector("button.action-menu", { timeout: 5000 }));
    if (!(await Selector("button.action-private").visible)) {
      await t.click(Selector("button.action-menu", { timeout: 5000 }));
      if (!(await Selector("button.action-private").visible)) {
        await t.click(Selector("button.action-menu", { timeout: 5000 }));
      }
      if (!(await Selector("button.action-private").visible)) {
        await t.click(Selector("button.action-menu", { timeout: 5000 }));
      }
    }
    await t.click(Selector("button.action-private", { timeout: 5000 }));
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
    await t.click(Selector("button.action-menu")).click(Selector("button.action-remove"));
  }

  async addSelectedToAlbum(name, type) {
    await t
      .click(Selector("button.action-menu"))
      .click(Selector("button.action-" + type, { timeout: 15000 }))
      .typeText(Selector(".input-album input"), name, { replace: true })
      .pressKey("enter");
  }

  async turnSwitchOff(type) {
    await t
      .click("#tab-info")
      .expect(
        Selector(".input-" + type + " input", { timeout: 8000 }).hasAttribute(
          "aria-checked",
          "true"
        )
      )
      .ok()
      .click(Selector(".input-" + type + " input"))
      .expect(
        Selector(".input-" + type + " input", { timeout: 8000 }).hasAttribute(
          "aria-checked",
          "false"
        )
      )
      .ok();
  }

  async turnSwitchOn(type) {
    await t
      .click("#tab-info")
      .expect(
        Selector(".input-" + type + " input", { timeout: 8000 }).hasAttribute(
          "aria-checked",
          "false"
        )
      )
      .ok()
      .click(Selector(".input-" + type + " input"))
      .expect(
        Selector(".input-" + type + " input", { timeout: 8000 }).hasAttribute(
          "aria-checked",
          "true"
        )
      )
      .ok();
  }

  async clearSelection() {
    if (await Selector(".action-clear").visible) {
      await t.click(Selector(".action-clear"));
    } else {
      await t.click(Selector(".action-menu")).click(Selector(".action-clear"));
    }
  }

  async login(username, password) {
    await t
      .typeText(Selector(".input-name input"), username, { replace: true, timeout: 5000 })
      .typeText(Selector(".input-password input"), password, { replace: true })
      .click(Selector(".action-confirm"));
  }

  async logout() {
    await t.click(Selector("div.nav-logout"));
  }

  async testCreateEditDeleteSharingLink(type) {
    await this.openNav();
    if (type === "states") {
      await t.click(Selector(".nav-places + div"));
    }
    await t.click(Selector(".nav-" + type));
    const FirstAlbum = await Selector("a.is-album").nth(0).getAttribute("data-uid");
    await this.selectFromUID(FirstAlbum);
    const clipboardCount = await Selector("span.count-clipboard");
    await t
      .expect(clipboardCount.textContent)
      .eql("1")
      .click(Selector("button.action-menu"))
      .click(Selector("button.action-share"))
      .click(Selector("div.v-expansion-panel__header__icon").nth(0));
    const InitialUrl = await Selector(".action-url").innerText;
    const InitialSecret = await Selector(".input-secret input").value;
    const InitialExpire = await Selector("div.v-select__selections").innerText;
    await t
      .expect(InitialUrl)
      .notContains("secretfortesting")
      .expect(InitialExpire)
      .contains("Never")
      .typeText(Selector(".input-secret input"), "secretForTesting", { replace: true })
      .click(Selector(".input-expires input"))
      .click(Selector("div").withText("After 1 day").parent('div[role="listitem"]'))
      .click(Selector("button.action-save"))
      .click(Selector("button.action-close"));
    await this.clearSelection();
    await t
      .click(Selector("a.is-album").withAttribute("data-uid", FirstAlbum))
      .click(Selector("button.action-share"))
      .click(Selector("div.v-expansion-panel__header__icon").nth(0));
    const UrlAfterChange = await Selector(".action-url").innerText;
    const ExpireAfterChange = await Selector("div.v-select__selections").innerText;
    await t
      .expect(UrlAfterChange)
      .contains("secretfortesting")
      .expect(ExpireAfterChange)
      .contains("After 1 day")
      .typeText(Selector(".input-secret input"), InitialSecret, { replace: true })
      .click(Selector(".input-expires input"))
      .click(Selector("div").withText("Never").parent('div[role="listitem"]'))
      .click(Selector("button.action-save"))
      .click(Selector("div.v-expansion-panel__header__icon"));
    const LinkCount = await Selector(".action-url").count;
    await t.click(".action-add-link");
    const LinkCountAfterAdd = await Selector(".action-url").count;
    await t
      .expect(LinkCountAfterAdd)
      .eql(LinkCount + 1)
      .click(Selector("div.v-expansion-panel__header__icon"))
      .click(Selector(".action-delete"));
    const LinkCountAfterDelete = await Selector(".action-url").count;
    await t
      .expect(LinkCountAfterDelete)
      .eql(LinkCountAfterAdd - 1)
      .click(Selector("button.action-close"));
    await this.openNav();
    await t
      .click(".nav-" + type)
      .click("a.uid-" + FirstAlbum + " .action-share")
      .click(Selector("div.v-expansion-panel__header__icon"))
      .click(Selector(".action-delete"));
  }

  async checkButtonVisibility(button, inContextMenu, inAlbum) {
    const FirstAlbum = await Selector("a.is-album").nth(0).getAttribute("data-uid");
    await this.selectFromUID(FirstAlbum);
    await t.click(Selector("button.action-menu"));
    if (inContextMenu) {
      await t.expect(Selector("button.action-" + button).visible).ok();
    } else {
      await t.expect(Selector("button.action-" + button).visible).notOk();
    }
    await this.clearSelection();
    if (t.browser.platform !== "mobile") {
      await t.click(Selector("a.is-album").nth(0));
      if (inAlbum) {
        await t.expect(Selector("button.action-" + button).visible).ok();
      } else {
        await t.expect(Selector("button.action-" + button).visible).notOk();
      }
    }
  }

  async deletePhotoFromUID(uid) {
    await this.selectPhotoFromUID(uid);
    await this.archiveSelected();
    await this.openNav();
    await t.click(Selector(".nav-browse + div")).click(Selector(".nav-archive"));
    await this.selectPhotoFromUID(uid);
    await t
      .click(Selector("button.action-menu", { timeout: 5000 }))
      .click(Selector(".remove"))
      .click(Selector(".action-confirm"))
      .expect(Selector("div").withAttribute("data-uid", uid).exists, { timeout: 5000 })
      .notOk();
  }

  async validateDownloadRequest(request, filename, extension) {
    const downloadedFileName = request.headers["content-disposition"];
    await t
      .expect(request.statusCode === 200)
      .ok()
      .expect(downloadedFileName)
      .contains(filename)
      .expect(downloadedFileName)
      .contains(extension);
    await logger.clear();
  }

  async checkEditFormValues(
    title,
    day,
    month,
    year,
    localTime,
    utcTime,
    timezone,
    country,
    altitude,
    lat,
    lng,
    camera,
    iso,
    exposure,
    lens,
    fnumber,
    flength,
    subject,
    artist,
    copyright,
    license,
    description,
    keywords,
    notes
  ) {
    if (title !== "") {
      await t.expect(Selector(".input-title input").value).eql(title);
    }
    if (day !== "") {
      await t.expect(Selector(".input-day input").value).eql(day);
    }
    if (month !== "") {
      await t.expect(Selector(".input-month input").value).eql(month);
    }
    if (year !== "") {
      await t.expect(Selector(".input-year input").value).eql(year);
    }
    if (timezone !== "") {
      await t.expect(Selector(".input-timezone input").value).eql(timezone);
    }
    if (localTime !== "") {
      await t.expect(Selector(".input-local-time input").value).eql(localTime);
    }
    if (utcTime !== "") {
      await t.expect(Selector(".input-utc-time input").value).eql(utcTime);
    }
    if (altitude !== "") {
      await t.expect(Selector(".input-altitude input").value).eql(altitude);
    }
    if (country !== "") {
      await t.expect(Selector("div").withText(country).visible).ok();
    }
    if (lat !== "") {
      await t.expect(Selector(".input-latitude input").value).eql(lat);
    }
    if (lng !== "") {
      await t.expect(Selector(".input-longitude input").value).eql(lng);
    }
    if (camera !== "") {
      await t.expect(Selector("div").withText(camera).visible).ok();
    }
    if (lens !== "") {
      await t.expect(Selector("div").withText(lens).visible).ok();
    }
    if (iso !== "") {
      await t.expect(Selector(".input-iso input").value).eql(iso);
    }
    if (exposure !== "") {
      await t.expect(Selector(".input-exposure input").value).eql(exposure);
    }
    if (fnumber !== "") {
      await t.expect(Selector(".input-fnumber input").value).eql(fnumber);
    }
    if (flength !== "") {
      await t.expect(Selector(".input-focal-length input").value).eql(flength);
    }
    if (subject !== "") {
      await t.expect(Selector(".input-subject textarea").value).eql(subject);
    }
    if (artist !== "") {
      await t.expect(Selector(".input-artist input").value).eql(artist);
    }
    if (copyright !== "") {
      await t.expect(Selector(".input-copyright input").value).eql(copyright);
    }
    if (license !== "") {
      await t.expect(Selector(".input-license textarea").value).eql(license);
    }
    if (description !== "") {
      await t.expect(Selector(".input-description textarea").value).eql(description);
    }
    if (notes !== "") {
      await t.expect(Selector(".input-notes textarea").value).contains(notes);
    }
    if (keywords !== "") {
      await t.expect(Selector(".input-keywords textarea").value).contains(keywords);
    }
  }

  async editPhoto(
    title,
    timezone,
    day,
    month,
    year,
    localTime,
    altitude,
    lat,
    lng,
    iso,
    exposure,
    fnumber,
    flength,
    subject,
    artist,
    copyright,
    license,
    description,
    keywords,
    notes
  ) {
    await t
      .typeText(Selector(".input-title input"), title, { replace: true })
      .typeText(Selector(".input-timezone input"), timezone, { replace: true })
      .click(Selector("div").withText(timezone).parent('div[role="listitem"]'))
      .typeText(Selector(".input-day input"), day, { replace: true })
      .pressKey("enter")
      .typeText(Selector(".input-month input"), month, { replace: true })
      .pressKey("enter")
      .typeText(Selector(".input-year input"), year, { replace: true })
      .click(Selector("div").withText(year).parent('div[role="listitem"]'))
      .click(Selector(".input-local-time input"))
      .pressKey("ctrl+a delete")
      .typeText(Selector(".input-local-time input"), localTime, { replace: true })
      .pressKey("enter")
      .typeText(Selector(".input-altitude input"), altitude, { replace: true })
      .typeText(Selector(".input-latitude input"), lat, { replace: true })
      .typeText(Selector(".input-longitude input"), lng, { replace: true })
      //.click(Selector('.input-camera input'))
      //.hover(Selector('div').withText('Apple iPhone 6').parent('div[role="listitem"]'))
      //.click(Selector('div').withText('Apple iPhone 6').parent('div[role="listitem"]'))
      //.click(Selector('.input-lens input'))
      //.click(Selector('div').withText('Apple iPhone 5s back camera 4.15mm f/2.2').parent('div[role="listitem"]'))
      .typeText(Selector(".input-iso input"), iso, { replace: true })
      .typeText(Selector(".input-exposure input"), exposure, { replace: true })
      .typeText(Selector(".input-fnumber input"), fnumber, { replace: true })
      .typeText(Selector(".input-focal-length input"), flength, { replace: true })
      .typeText(Selector(".input-subject textarea"), subject, { replace: true })
      .typeText(Selector(".input-artist input"), artist, { replace: true })
      .typeText(Selector(".input-copyright input"), copyright, { replace: true })
      .typeText(Selector(".input-license textarea"), license, { replace: true })
      .typeText(Selector(".input-description textarea"), description, {
        replace: true,
      })
      .typeText(Selector(".input-keywords textarea"), keywords)
      .typeText(Selector(".input-notes textarea"), notes, { replace: true })
      .click(Selector("button.action-approve"));
    await t.expect(Selector(".input-latitude input").visible, { timeout: 5000 }).ok();
    if (t.browser.platform === "mobile") {
      await t.click(Selector("button.action-apply")).click(Selector("button.action-close"));
    } else {
      await t.click(Selector("button.action-done", { timeout: 5000 }));
    }
  }

  async undoPhotoEdit(
    title,
    timezone,
    day,
    month,
    year,
    localTime,
    altitude,
    lat,
    lng,
    country,
    iso,
    exposure,
    fnumber,
    flength,
    subject,
    artist,
    copyright,
    license,
    description,
    keywords,
    notes
  ) {
    if (title.empty || title === "") {
      await t.click(Selector(".input-title input")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-title input"), title, { replace: true });
    }
    await t
      .typeText(Selector(".input-day input"), day, { replace: true })
      .pressKey("enter")
      .typeText(Selector(".input-month input"), month, { replace: true })
      .pressKey("enter")
      .typeText(Selector(".input-year input"), year, { replace: true })
      .pressKey("enter");
    if (localTime.empty || localTime === "") {
      await t.click(Selector(".input-local-time input")).pressKey("ctrl+a delete");
    } else {
      await t
        .click(Selector(".input-local-time input"))
        .pressKey("ctrl+a delete")
        .typeText(Selector(".input-local-time input"), localTime, { replace: true })
        .pressKey("enter");
    }
    if (timezone.empty || timezone === "") {
      await t
        .click(Selector(".input-timezone input"))
        .typeText(Selector(".input-timezone input"), "UTC", { replace: true })
        .pressKey("enter");
    } else {
      await t
        .typeText(Selector(".input-timezone input"), timezone, { replace: true })
        .pressKey("enter");
    }
    if (lat.empty || lat === "") {
      await t.click(Selector(".input-latitude input")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-latitude input"), lat, { replace: true });
    }
    if (lng.empty || lng === "") {
      await t.click(Selector(".input-longitude input")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-longitude input"), lng, { replace: true });
    }
    if (altitude.empty || altitude === "") {
      await t.click(Selector(".input-altitude input")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-altitude input"), altitude, { replace: true });
    }
    if (country.empty || country === "") {
      await t.click(Selector(".input-longitude input")).pressKey("ctrl+a delete");
    } else {
      await t
        .click(Selector(".input-country input"))
        .pressKey("ctrl+a delete")
        .typeText(Selector(".input-country input"), country, { replace: true })
        .pressKey("enter");
    }
    // if (FirstPhotoCamera.empty || FirstPhotoCamera === "")
    //{ await t
    //.click(Selector('.input-camera input'))
    // .hover(Selector('div').withText('Unknown').parent('div[role="listitem"]'))
    //  .click(Selector('div').withText('Unknown').parent('div[role="listitem"]'))}
    //else
    //{await t
    //  .click(Selector('.input-camera input'))
    //   .hover(Selector('div').withText(FirstPhotoCamera).parent('div[role="listitem"]'))
    //    .click(Selector('div').withText(FirstPhotoCamera).parent('div[role="listitem"]'))}
    //if (FirstPhotoLens.empty || FirstPhotoLens === "")
    //{ await t
    //  .click(Selector('.input-lens input'))
    //   .click(Selector('div').withText('Unknown').parent('div[role="listitem"]'))}
    //else
    //{await t
    //   .click(Selector('.input-lens input'))
    //    .click(Selector('div').withText(FirstPhotoLens).parent('div[role="listitem"]'))}
    if (iso.empty || iso === "") {
      await t.click(Selector(".input-iso input")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-iso input"), iso, { replace: true });
    }
    if (exposure.empty || exposure === "") {
      await t.click(Selector(".input-exposure input")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-exposure input"), exposure, { replace: true });
    }
    if (fnumber.empty || fnumber === "") {
      await t.click(Selector(".input-fnumber input")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-fnumber input"), fnumber, { replace: true });
    }
    if (flength.empty || flength === "") {
      await t.click(Selector(".input-focal-length input")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-focal-length input"), flength, {
        replace: true,
      });
    }
    if (subject.empty || subject === "") {
      await t.click(Selector(".input-subject textarea")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-subject textarea"), subject, { replace: true });
    }
    if (artist.empty || artist === "") {
      await t.click(Selector(".input-artist input")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-artist input"), artist, { replace: true });
    }
    if (copyright.empty || copyright === "") {
      await t.click(Selector(".input-copyright input")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-copyright input"), copyright, { replace: true });
    }
    if (license.empty || license === "") {
      await t.click(Selector(".input-license textarea")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-license textarea"), license, { replace: true });
    }
    if (description.empty || description === "") {
      await t.click(Selector(".input-description textarea")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-description textarea"), description, {
        replace: true,
      });
    }
    if (keywords.empty || keywords === "") {
      await t.click(Selector(".input-keywords textarea")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-keywords textarea"), keywords, { replace: true });
    }
    if (notes.empty || notes === "") {
      await t.click(Selector(".input-notes textarea")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-notes textarea"), notes, { replace: true });
    }
    if (t.browser.platform === "mobile") {
      await t.click(Selector("button.action-apply")).click(Selector("button.action-close"));
    } else {
      await t.click(Selector("button.action-done", { timeout: 5000 }));
    }
  }

  async checkMemberAlbumRights(type) {
    await t.expect(Selector("a.is-album button.action-share").visible).notOk();
    const FirstAlbum = await Selector("a.is-album").nth(0).getAttribute("data-uid");
    await this.selectFromUID(FirstAlbum);
    await t
      .click(Selector("button.action-menu"))
      .expect(Selector("button.action-edit").visible)
      .notOk()
      .expect(Selector("button.action-share").visible)
      .notOk()
      .expect(Selector("button.action-clone").visible)
      .notOk()
      .expect(Selector("button.action-download").visible)
      .ok();
    if (type == "album" || type == "moment" || type == "state") {
      await t.expect(Selector("button.action-delete").visible).motOk();
    }
    await this.clearSelection();
    await t
      .click(Selector("button.action-title-edit"))
      .expect(Selector(".input-description textarea").visible)
      .notOk();
    if (await Selector(`a.uid-${FirstAlbum}`).hasClass("is-favorite")) {
      await t
        .expect(Selector(`a.uid-${FirstAlbum}`).hasClass("is-favorite"))
        .ok()
        .click(Selector(`.uid-${FirstAlbum} .input-favorite`))
        .expect(Selector(`a.uid-${FirstAlbum}`).hasClass("is-favorite"))
        .ok();
    } else {
      await t
        .expect(Selector(`a.uid-${FirstAlbum}`).hasClass("is-favorite"))
        .notOk()
        .click(Selector(`.uid-${FirstAlbum} .input-favorite`))
        .expect(Selector(`a.uid-${FirstAlbum}`).hasClass("is-favorite"))
        .notOk();
    }
    await t
      .click(Selector("a.is-album").nth(0))
      .expect(Selector("button.action-share").visible)
      .notOk()
      .expect(Selector("button.action-edit").visible)
      .notOk();
    await this.toggleSelectNthPhoto(0);
    await t.click(Selector("button.action-menu"));
    if (type == "album") {
      await t.expect(Selector("button.action-remove").visible).notOk();
    } else {
      await t.expect(Selector("button.action-archive").visible).notOk();
    }
    await t
      .expect(Selector("button.action-album").visible)
      .notOk()
      .expect(Selector("button.action-private").visible)
      .notOk()
      .expect(Selector("button.action-share").visible)
      .notOk();
  }

  async checkAdminAlbumRights(type) {
    await t.expect(Selector("a.is-album button.action-share").visible).ok();
    const FirstAlbum = await Selector("a.is-album").nth(0).getAttribute("data-uid");
    await this.selectFromUID(FirstAlbum);
    await t
      .click(Selector("button.action-menu"))
      .expect(Selector("button.action-edit").visible)
      .ok()
      .expect(Selector("button.action-share").visible)
      .ok()
      .expect(Selector("button.action-clone").visible)
      .ok()
      .expect(Selector("button.action-download").visible)
      .ok();
    if (type == "album" || type == "moment" || type == "state") {
      await t.expect(Selector("button.action-delete").visible).ok();
    }
    await this.clearSelection();
    await t
      .click(Selector("button.action-title-edit"))
      .expect(Selector(".input-description textarea").visible)
      .ok()
      .click(Selector("button.action-cancel"));
    if (await Selector(`a.uid-${FirstAlbum}`).hasClass("is-favorite")) {
      await t
        .expect(Selector(`a.uid-${FirstAlbum}`).hasClass("is-favorite"))
        .ok()
        .click(Selector(`.uid-${FirstAlbum} .input-favorite`))
        .expect(Selector(`a.uid-${FirstAlbum}`).hasClass("is-favorite"))
        .notOk()
        .click(Selector(`.uid-${FirstAlbum} .input-favorite`))
        .expect(Selector(`a.uid-${FirstAlbum}`).hasClass("is-favorite"))
        .ok();
    } else {
      await t
        .expect(Selector(`a.uid-${FirstAlbum}`).hasClass("is-favorite"))
        .notOk()
        .click(Selector(`.uid-${FirstAlbum} .input-favorite`))
        .expect(Selector(`a.uid-${FirstAlbum}`).hasClass("is-favorite"))
        .ok()
        .click(Selector(`.uid-${FirstAlbum} .input-favorite`))
        .expect(Selector(`a.uid-${FirstAlbum}`).hasClass("is-favorite"))
        .notOk();
    }
    await t
      .click(Selector("a.is-album").nth(0))
      .expect(Selector("button.action-share").visible)
      .ok()
      .expect(Selector("button.action-edit").visible)
      .ok();
    await this.toggleSelectNthPhoto(0);
    await t.click(Selector("button.action-menu"));
    if (type == "album") {
      await t.expect(Selector("button.action-remove").visible).ok();
    } else {
      await t.expect(Selector("button.action-archive").visible).ok();
    }
    await t
      .expect(Selector("button.action-album").visible)
      .ok()
      .expect(Selector("button.action-private").visible)
      .ok()
      .expect(Selector("button.action-share").visible)
      .ok();
  }
}
