import { Selector, t } from "testcafe";

export default class Page {
  constructor() {
    this.languageInput = Selector(".input-language input");
    this.uploadCheckbox = Selector(".input-upload input");
    this.downloadCheckbox = Selector(".input-download input");
    this.importCheckbox = Selector(".input-import input");
    this.archiveCheckbox = Selector('input[aria-label="Archive"]');
    this.editCheckbox = Selector(".input-edit input");
    this.filesCheckbox = Selector(".input-files input");
    this.momentsCheckbox = Selector(".input-moments input");
    this.labelsCheckbox = Selector(".input-labels input");
    this.logsCheckbox = Selector(".input-logs input");
    this.shareCheckbox = Selector(".input-share input");
    this.placesCheckbox = Selector(".input-places input");
    this.privateCheckbox = Selector('input[aria-label="Private"]');
    this.peopleCheckbox = Selector(".input-people input");
    this.deleteCheckbox = Selector(".input-delete input");
    this.libraryCheckbox = Selector(".input-library input");

    this.libraryTab = Selector("#tab-settings-library");
    this.reviewCheckbox = Selector(".input-review input");
  }
}
