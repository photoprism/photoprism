import { Selector, t } from "testcafe";

export default class Page {
  constructor() {
    this.dialog = Selector("div.v-dialog");
    this.dialogClose = Selector("div.v-dialog button.action-close", { timeout: 15000 });
    this.dialogNext = Selector("div.v-dialog button.action-next", { timeout: 15000 });
    this.dialogPrevious = Selector("div.v-dialog button.action-previous", { timeout: 15000 });

    this.filesTab = Selector("#tab-files", { timeout: 15000 });
    this.infoTab = Selector("#tab-info", { timeout: 15000 });
    this.detailsTab = Selector("#tab-details", { timeout: 15000 });
    this.labelsTab = Selector("#tab-labels", { timeout: 15000 });
    this.peopleTab = Selector("#tab-people", { timeout: 15000 });

    this.locationAction = Selector(".input-coordinates i.action-map", { timeout: 15000 });
    this.locationSearch = Selector("div.p-location-dialog .v-autocomplete", { timeout: 15000 });
    this.locationClear = Selector(".input-coordinates i.action-clear", { timeout: 15000 });
    this.locationUndo = Selector("div.p-location-dialog .input-coordinates i.action-undo", { timeout: 15000 });
    this.locationInput = Selector("div.p-location-dialog .input-coordinates input", { timeout: 15000 });
    this.locationConfirm = Selector("div.p-location-dialog button.action-confirm", { timeout: 15000 });
    this.locationCancel = Selector("div.p-location-dialog button.action-cancel", { timeout: 15000 });
    this.locationMarker = Selector("div.maplibregl-marker", { timeout: 15000 });

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
    this.coordinates = Selector("div.p-tab-photo-details .input-coordinates input", { timeout: 15000 });
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
    this.faceActionMenuButton = Selector(".p-faces .p-action-menu .action-menu__btn", { timeout: 15000 });
    this.removeFaceAction = Selector(".v-list-item.action-remove-face, .action-remove-face", { timeout: 15000 });
    this.goToPersonAction = Selector(".v-list-item.action-go-to-person, .action-go-to-person", { timeout: 15000 });
    this.setPersonCoverAction = Selector(".v-list-item.action-set-person-cover, .action-set-person-cover", {
      timeout: 15000,
    });
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
      this.coordinates,
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

  async openFaceMenu(index = 0) {
    await t.click(this.faceActionMenuButton.nth(index));
  }

  async removeFace(index = 0) {
    await this.openFaceMenu(index);
    await t.click(this.removeFaceAction.nth(0));
  }

  async goToPerson(index = 0) {
    await this.openFaceMenu(index);
    await t.click(this.goToPersonAction.nth(0));
  }

  async setPersonCover(index = 0) {
    await this.openFaceMenu(index);
    await t.click(this.setPersonCoverAction.nth(0));
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
    if (field === "keywords") {
      await t.expect(this[field].value).contains(val);
    } else {
      await t.expect(this[field].value).eql(val);
    }
  }

  async checkEditFormSelectValue(field, val) {
    if (val !== "") {
      if (val === "Unknown" && field === "day") {
        await t.expect(this[field].innerText).eql("");
      } else if (val === "Unknown" && field === "month") {
        await t.expect(this[field].innerText).eql("");
      } else if (val === "Unknown" && field === "year") {
        await t.expect(this[field].innerText).eql("");
      } else {
        await t.expect(this[field + "Value"].innerText).eql(val);
      }
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

  async editFormInputValue(field, val) {
    if (val !== "") {
      if (field === "keywords") {
        await t.typeText(this[field], val);
      } else {
        await t.typeText(this[field], val, { replace: true });
      }
    } else if (val === "") {
      if (field === "coordinates") {
        await t.click(this.locationClear);
      } else {
        await t.click(this[field]).pressKey("ctrl+a delete");
      }
    }
  }

  async editFormSelectValue(field, val) {
    if (val !== "") {
      if (field === "camera" || field === "lens") {
        await t.click(this[field]).click(Selector("div").withText(val).parent('div[role="option"]'));
      } else {
        await t
          .typeText(this[field], val, { replace: true })
          .click(Selector("div").withText(val).parent('div[role="option"]'));
      }
    }
  }

  async editFormValues(inputValues, selectValues) {
    inputValues.forEach((el) => {
      this.editFormInputValue(el[0], el[1]);
    });

    selectValues.forEach((x) => {
      this.editFormSelectValue(x[0], x[1]);
    });

    await t.click(Selector("button.action-approve"));
    await t.expect(this.coordinates.visible, { timeout: 5000 }).ok();
    await t.click(this.detailsApply).click(Selector("button.action-close"));
  }
}
