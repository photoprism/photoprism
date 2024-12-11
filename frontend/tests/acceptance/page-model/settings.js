import { Selector, t } from "testcafe";

export default class Page {
  constructor() {
    this.generalTab = Selector("#tab-settings_general");
    this.languageInput = Selector(".input-language input");
    this.languageOpenSelection = Selector(".input-language div.v-input__control");
    this.uploadCheckbox = Selector(".input-upload div.v-selection-control__input");
    this.downloadCheckbox = Selector(".input-download div.v-selection-control__input");
    this.importCheckbox = Selector(".input-import div.v-selection-control__input");
    this.archiveCheckbox = Selector(".input-archive div.v-selection-control__input");
    this.editCheckbox = Selector(".input-edit div.v-selection-control__input");
    this.filesCheckbox = Selector(".input-files div.v-selection-control__input");
    this.momentsCheckbox = Selector(".input-moments div.v-selection-control__input");
    this.labelsCheckbox = Selector(".input-labels div.v-selection-control__input");
    this.logsCheckbox = Selector(".input-logs div.v-selection-control__input");
    this.shareCheckbox = Selector(".input-share div.v-selection-control__input");
    this.placesCheckbox = Selector(".input-places div.v-selection-control__input");
    this.privateCheckbox = Selector(
      'input[aria-label="Private"] + div.v-selection-control__input'
    );
    this.peopleCheckbox = Selector(".input-people div.v-selection-control__input");
    this.deleteCheckbox = Selector(".input-delete div.v-selection-control__input");
    this.libraryCheckbox = Selector(".input-library div.v-selection-control__input");

    this.libraryTab = Selector("#tab-settings_media");
    this.reviewCheckbox = Selector(".input-review div.v-selection-control__input");
    this.convertCheckbox = Selector(".input-convert div.v-selection-control__input");
    this.estimatesCheckbox = Selector(".input-estimates div.v-selection-control__input");
    this.dateTimeStacksCheckbox = Selector(
      ".input-stack-meta div.v-selection-control__input"
    );
    this.uuidStacksCheckbox = Selector(".input-stack-uuid div.v-selection-control__input");
    this.nameStacksCheckbox = Selector(".input-stack-name div.v-selection-control__input");

    this.advancedTab = Selector("#tab-settings_advanced");
    this.debugCheckbox = Selector("label").withText("Debug Logs");
    this.backupCheckbox = Selector("label").withText("Disable Backups");
    this.exiftoolCheckbox = Selector("label").withText("Disable ExifTool");
    this.disableplacesCheckbox = Selector("label").withText("Disable Places");
    this.tensorflowCheckbox = Selector("label").withText("Disable TensorFlow");
    this.readOnlyCheckbox = Selector("label").withText("Read-Only Mode");

    this.accountTab = Selector("#tab-settings_account");
    this.servicesTab = Selector("#tab-settings_services");
  }
}
