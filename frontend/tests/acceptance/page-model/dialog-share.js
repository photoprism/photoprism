import { Selector, t } from "testcafe";

export default class Page {
  constructor() {
    this.dialogClose = Selector(".v-card-actions button.action-close", { timeout: 15000 });
    this.dialogSave = Selector("button.action-save", { timeout: 15000 });
    this.addLink = Selector(".action-add-link", { timeout: 15000 });
    this.deleteLink = Selector(".action-delete", { timeout: 15000 });
    this.expandLink = Selector("button.v-expansion-panel-title", { timeout: 15000 });
    this.linkUrl = Selector(".input-url input", { timeout: 15000 });
    this.linkSecretInput = Selector(".input-secret input", { timeout: 15000 });
    this.linkExpireInput = Selector(".input-expires div.v-input__control", { timeout: 15000 });
  }
}
