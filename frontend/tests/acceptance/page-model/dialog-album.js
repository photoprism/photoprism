import { Selector, t } from "testcafe";

export default class Page {
  constructor() {
    this.dialogCancel = Selector(".action-cancel", { timeout: 15000 });
    this.dialogSave = Selector(".action-confirm", { timeout: 15000 });
    this.title = Selector(".input-title input", { timeout: 15000 });
    this.description = Selector(".input-description textarea", { timeout: 15000 });
    this.category = Selector(".input-category input", { timeout: 15000 });
    this.location = Selector(".input-location input", { timeout: 15000 });
  }
}
