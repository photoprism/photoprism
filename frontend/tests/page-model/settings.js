import { Selector, t } from "testcafe";

export default class Page {
  constructor() {
    this.languageInput = Selector(".input-language input", { timeout: 5000 });
    this.uploadCheckbox = Selector(".input-upload input", { timeout: 5000 });
    this.downloadCheckbox = Selector(".input-download input", { timeout: 5000 });
    this.importCheckbox = Selector(".input-import input", { timeout: 5000 });
    this.archiveCheckbox = Selector(".input-archive input", { timeout: 5000 });
    this.editCheckbox = Selector(".input-edit input", { timeout: 5000 });
    this.filesCheckbox = Selector(".input-files input", { timeout: 5000 });
    this.momentsCheckbox = Selector(".input-moments input", { timeout: 5000 });
    this.labelsCheckbox = Selector(".input-labels input", { timeout: 5000 });
    this.logsCheckbox = Selector(".input-logs input", { timeout: 5000 });
    this.shareCheckbox = Selector(".input-share input", { timeout: 5000 });
    this.placesCheckbox = Selector(".input-places input", { timeout: 5000 });
    this.privateCheckbox = Selector(".input-private input", { timeout: 5000 });
    this.peopleCheckbox = Selector(".input-people input", { timeout: 5000 });
    this.deleteCheckbox = Selector(".input-delete input", { timeout: 5000 });

    this.libraryTab = Selector("#tab-settings-library", { timeout: 15000 });
    this.reviewCheckbox = Selector(".input-review input", { timeout: 5000 });
  }
}
