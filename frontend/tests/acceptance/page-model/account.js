import { Selector, t } from "testcafe";

export default class Page {
  constructor() {
    this.changePasswordAction = Selector("button.action-change-password");
    this.currentPassword = Selector(".input-current-password input", { timeout: 15000 });
    this.newPassword = Selector(".input-new-password input", { timeout: 15000 });
    this.retypePassword = Selector(".input-retype-password input", { timeout: 15000 });
    this.confirm = Selector(".action-confirm");
  }
}
