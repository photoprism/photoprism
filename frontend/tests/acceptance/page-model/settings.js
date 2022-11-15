import { Selector, t } from "testcafe";

export default class Page {
  constructor() {
    this.generalTab = Selector("#tab-settings_general");
    this.languageInput = Selector(".input-language input");
    this.languageOpenSelection = Selector(".input-language div.v-select__selections");
    this.uploadCheckbox = Selector(".input-upload div.v-input--selection-controls__ripple");
    this.downloadCheckbox = Selector(".input-download div.v-input--selection-controls__ripple");
    this.importCheckbox = Selector(".input-import div.v-input--selection-controls__ripple");
    this.archiveCheckbox = Selector(".input-archive div.v-input--selection-controls__ripple");
    this.editCheckbox = Selector(".input-edit div.v-input--selection-controls__ripple");
    this.filesCheckbox = Selector(".input-files div.v-input--selection-controls__ripple");
    this.momentsCheckbox = Selector(".input-moments div.v-input--selection-controls__ripple");
    this.labelsCheckbox = Selector(".input-labels div.v-input--selection-controls__ripple");
    this.logsCheckbox = Selector(".input-logs div.v-input--selection-controls__ripple");
    this.shareCheckbox = Selector(".input-share div.v-input--selection-controls__ripple");
    this.placesCheckbox = Selector(".input-places div.v-input--selection-controls__ripple");
    this.privateCheckbox = Selector(
      'input[aria-label="Private"] + div.v-input--selection-controls__ripple'
    );
    this.peopleCheckbox = Selector(".input-people div.v-input--selection-controls__ripple");
    this.deleteCheckbox = Selector(".input-delete div.v-input--selection-controls__ripple");
    this.libraryCheckbox = Selector(".input-library div.v-input--selection-controls__ripple");

    this.libraryTab = Selector("#tab-settings_media");
    this.reviewCheckbox = Selector(".input-review div.v-input--selection-controls__ripple");
    this.convertCheckbox = Selector(".input-convert div.v-input--selection-controls__ripple");
    this.estimatesCheckbox = Selector(".input-estimates div.v-input--selection-controls__ripple");
    this.dateTimeStacksCheckbox = Selector(
      ".input-stack-meta div.v-input--selection-controls__ripple"
    );
    this.uuidStacksCheckbox = Selector(".input-stack-uuid div.v-input--selection-controls__ripple");
    this.nameStacksCheckbox = Selector(".input-stack-name div.v-input--selection-controls__ripple");

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
