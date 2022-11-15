import { Selector, t } from "testcafe";

export default class Page {
  constructor() {
    this.dialogClose = Selector("button.action-close", { timeout: 15000 });
    this.dialogSave = Selector("button.action-save", { timeout: 15000 });
    this.addLink = Selector(".action-add-link", { timeout: 15000 });
    this.deleteLink = Selector(".action-delete", { timeout: 15000 });
    this.expandLink = Selector("div.v-expansion-panel__header__icon", { timeout: 15000 });
    this.linkUrl = Selector(".action-url", { timeout: 15000 });
    this.linkSecretInput = Selector(".input-secret input", { timeout: 15000 });
    this.linkExpireInput = Selector(".input-expires div.v-select__selections", { timeout: 15000 });
  }
}
