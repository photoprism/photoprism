import { Selector, t } from "testcafe";

export default class Page {
  constructor() {
    this.dialogClose = Selector("div.v-dialog button.action-close", { timeout: 15000 });
    this.dialogNext = Selector("div.v-dialog button.action-next", { timeout: 15000 });
    this.dialogPrevious = Selector("div.v-dialog button.action-previous", { timeout: 15000 });

    this.filesTab = Selector("#tab-files", { timeout: 15000 });
    this.infoTab = Selector("#tab-info", { timeout: 15000 });
    this.detailsTab = Selector("#tab-details", { timeout: 15000 });
    this.labelsTab = Selector("#tab-labels", { timeout: 15000 });
    this.peopleTab = Selector("#tab-people", { timeout: 15000 });

    this.detailsDone = Selector(".p-form-photo-details-meta button.action-done", {
      timeout: 15000,
    });
    this.detailsApprove = Selector(".p-form-photo-details-meta button.action-approve", {
      timeout: 15000,
    });
    this.detailsClose = Selector(".p-form-photo-details-meta button.action-close", {
      timeout: 15000,
    });
    this.detailsApply = Selector(".p-form-photo-details-meta button.action-apply", {
      timeout: 15000,
    });
    this.keywords = Selector(".input-keywords textarea", { timeout: 15000 });
    this.title = Selector(".input-title input", { timeout: 15000 });
    this.latitude = Selector(".input-latitude input", { timeout: 15000 });
    this.longitude = Selector(".input-longitude input", { timeout: 15000 });
    this.localTime = Selector(".input-local-time input", { timeout: 15000 });
    this.day = Selector("div.input-day input", { timeout: 15000 });
    this.month = Selector(".input-month input", { timeout: 15000 });
    this.year = Selector(".input-year input", { timeout: 15000 });
    this.timezone = Selector(".input-timezone input", { timeout: 15000 });
    this.dayValue = Selector(".input-day .v-combobox__selection", { timeout: 15000 });
    this.monthValue = Selector(".input-month .v-combobox__selection", { timeout: 15000 });
    this.yearValue = Selector(".input-year .v-combobox__selection", { timeout: 15000 });
    this.timezoneValue = Selector(".input-timezone .v-autocomplete__selection", { timeout: 15000 });
    this.altitude = Selector(".input-altitude input", { timeout: 15000 });
    this.countryValue = Selector(".input-country .v-autocomplete__selection", { timeout: 15000 });
    this.country = Selector(".input-country input", { timeout: 15000 });
    this.iso = Selector(".input-iso input", { timeout: 15000 });
    this.exposure = Selector(".input-exposure input", { timeout: 15000 });
    this.fnumber = Selector(".input-fnumber input", { timeout: 15000 });
    this.focallength = Selector(".input-focal-length input", { timeout: 15000 });
    this.subject = Selector(".input-subject textarea", { timeout: 15000 });
    this.artist = Selector(".input-artist input", { timeout: 15000 });
    this.copyright = Selector(".input-copyright input", { timeout: 15000 });
    this.license = Selector(".input-license textarea", { timeout: 15000 });
    this.description = Selector(".input-caption textarea", { timeout: 15000 });
    this.notes = Selector(".input-notes textarea", { timeout: 15000 });
    this.camera = Selector(".input-camera input", { timeout: 15000 });
    this.lens = Selector(".input-lens input", { timeout: 15000 });
    this.cameraValue = Selector(".input-camera .v-select__selection-text", { timeout: 15000 });
    this.lensValue = Selector(".input-lens .v-select__selection-text", { timeout: 15000 });

    this.rejectName = Selector("i.mdi-eject", { timeout: 15000 });
    this.removeMarker = Selector("button.input-reject", { timeout: 15000 });
    this.undoRemoveMarker = Selector("button.action-undo", { timeout: 15000 });
    this.inputName = Selector("div.input-name input", { timeout: 15000 });

    this.addLabel = Selector("button.p-photo-label-add", { timeout: 15000 });
    this.removeLabel = Selector("button.action-remove", { timeout: 15000 });
    this.activateLabel = Selector(".action-on", { timeout: 15000 });
    this.deleteLabel = Selector(".action-delete", { timeout: 15000 });
    this.inputLabelName = Selector(".input-label input", { timeout: 15000 });
    this.openInlineEdit = Selector("div.p-inline-edit", { timeout: 15000 });
    this.inputLabelRename = Selector(".input-title input", { timeout: 15000 });

    this.downloadFile = Selector("button.action-download", { timeout: 15000 });
    this.unstackFile = Selector(".action-unstack", { timeout: 15000 });
    this.deleteFile = Selector(".action-delete", { timeout: 15000 });
    this.makeFilePrimary = Selector(".action-primary", { timeout: 15000 });
    this.toggleExpandFile = Selector("button.v-expansion-panel-title", { timeout: 15000 });

    this.favoriteInput = Selector(".input-favorite input");
    this.privateInput = Selector(".input-private input");
    this.scanInput = Selector(".input-scan input");
    this.panoramaInput = Selector(".input-panorama input");
    this.stackableInput = Selector(".input-stackable input");
    this.typeInput = Selector(".input-type input");
  }

  async editDetailsField(field, value) {
    await t.typeText(field, value, { replace: true });
  }

  async checkFieldDisabled(field, disabled) {
    if (disabled) {
      await t.expect(field.hasAttribute("disabled")).ok();
    } else {
      await t.expect(field.hasAttribute("disabled")).notOk();
    }
  }

  async checkAllDetailsFieldsDisabled(disabled) {
    const fields = [
      this.title,
      this.latitude,
      this.longitude,
      this.keywords,
      this.localTime,
      this.day,
      this.month,
      this.year,
      this.timezone,
      this.altitude,
      this.country,
      this.iso,
      this.exposure,
      this.fnumber,
      this.focallength,
      this.subject,
      this.artist,
      this.copyright,
      this.license,
      this.description,
      this.notes,
      this.camera,
      this.lens,
    ];

    fields.forEach((item) => {
      this.checkFieldDisabled(item, disabled);
    });
  }

  async checkAllInfoFieldsDisabled(disabled) {
    const fields = [
      this.favoriteInput,
      this.privateInput,
      this.scanInput,
      this.panoramaInput,
      this.stackableInput,
      this.typeInput,
    ];

    fields.forEach((item) => {
      this.checkFieldDisabled(item, disabled);
    });
  }

  async getFileCount() {
    const FileCount = await Selector("div.v-expansion-panel", { timeout: 5000 }).count;
    return FileCount;
  }

  async turnSwitchOff(type) {
    await t.click("#tab-info");
    const initialState = await Selector("td .input-" + type + " input", { timeout: 8000 }).hasAttribute("checked");
    if (initialState === true) {
      await t.click(Selector("td .input-" + type + " div.v-switch__track"));
    }
    const finalState = await Selector("td .input-" + type + " input", { timeout: 8000 }).hasAttribute("checked");
    await t.expect(finalState).eql(false);
  }

  async turnSwitchOn(type) {
    await t.click("#tab-info");
    const initialState = await Selector("td .input-" + type + " input", { timeout: 8000 }).hasAttribute("checked");
    if (initialState === false) {
      await t.click(Selector("td .input-" + type + " div.v-selection-control__input"));
    }
    const finalState = await Selector("td .input-" + type + " input", { timeout: 8000 }).hasAttribute("checked");
    await t.expect(finalState).eql(true);
  }

  async checkEditFormInputValue(field, val) {
    if (val !== "") {
      await t.expect(this[field].value).eql(val);
    }
  }

  async checkEditFormSelectValue(field, val) {
    if (val !== "") {
      await t.expect(this[field + "Value"].innerText).eql(val);
    }
  }

  async checkEditFormValues(expectedInputValues, expectedSelectValues) {
    expectedInputValues.forEach((el) => {
      this.checkEditFormInputValue(el[0], el[1]);
    });

    expectedSelectValues.forEach((x) => {
      this.checkEditFormSelectValue(x[0], x[1]);
    });
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
    notes,
    camera,
    lens
  ) {
    await t
      .typeText(this.title, title, { replace: true })
      .typeText(this.timezone, timezone, { replace: true })
      .click(Selector("div").withText(timezone).parent('div[role="option"]'))
      .typeText(this.day, day, { replace: true })
      .click(Selector("div").withText(day).parent('div[role="option"]'))
      .typeText(this.month, month, { replace: true })
      .click(Selector("div").withText(month).parent('div[role="option"]'))
      .typeText(this.year, year, { replace: true })
      .click(Selector("div").withText(year).parent('div[role="option"]'))
      .click(this.localTime)
      .pressKey("ctrl+a delete")
      .typeText(this.localTime, localTime, { replace: true })
      .pressKey("enter")

      .typeText(this.altitude, altitude, { replace: true })
      .typeText(this.latitude, lat, { replace: true })
      .typeText(this.longitude, lng, { replace: true })
      .typeText(this.camera, camera, { replace: true })
      .click(Selector("div").withText(camera).parent('div[role="option"]'))
      .typeText(this.lens, timezone, { replace: true })
      .click(Selector("div").withText(lens).parent('div[role="option"]'))
      .typeText(this.iso, iso, { replace: true })
      .typeText(this.exposure, exposure, { replace: true })
      .typeText(this.fnumber, fnumber, { replace: true })
      .typeText(this.focallength, flength, { replace: true })
      .typeText(this.subject, subject, { replace: true })
      .typeText(this.artist, artist, { replace: true })
      .typeText(this.copyright, copyright, { replace: true })
      .typeText(this.license, license, { replace: true })
      .typeText(this.description, description, {
        replace: true,
      })
      .typeText(this.keywords, keywords)
      .typeText(this.notes, notes, { replace: true })

      .click(Selector("button.action-approve"));
    await t.expect(this.latitude.visible, { timeout: 5000 }).ok();
    await t.click(Selector("button.action-apply")).click(Selector("button.action-close"));
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
      await t.typeText(Selector(".input-timezone input"), timezone, { replace: true }).pressKey("enter");
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
    // .hover(Selector('div').withText('Unknown').parent('div[role="option"]'))
    //  .click(Selector('div').withText('Unknown').parent('div[role="option"]'))}
    //else
    //{await t
    //  .click(Selector('.input-camera input'))
    //   .hover(Selector('div').withText(FirstPhotoCamera).parent('div[role="option"]'))
    //    .click(Selector('div').withText(FirstPhotoCamera).parent('div[role="option"]'))}
    //if (FirstPhotoLens.empty || FirstPhotoLens === "")
    //{ await t
    //  .click(Selector('.input-lens input'))
    //   .click(Selector('div').withText('Unknown').parent('div[role="option"]'))}
    //else
    //{await t
    //   .click(Selector('.input-lens input'))
    //    .click(Selector('div').withText(FirstPhotoLens).parent('div[role="option"]'))}
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
      await t.click(Selector(".input-caption textarea")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-caption textarea"), description, {
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
    await t.click(Selector("button.action-apply")).click(Selector("button.action-close"));
  }
}
