import { Selector, t } from "testcafe";

export default class Page {
  constructor() {
    this.changePasswordAction = Selector("button.action-change-password");
    this.currentPassword = Selector(".input-current-password input", { timeout: 15000 });
    this.newPassword = Selector(".input-new-password input", { timeout: 15000 });
    this.retypePassword = Selector(".input-retype-password input", { timeout: 15000 });
    this.confirm = Selector(".action-confirm");
    this.close = Selector(".action-close");
    this.appsAndDevicesAction = Selector("button.action-apps-dialog")
    this.appAdd = Selector("button.action-add")
    this.clientName = Selector(".input-name input", { timeout: 15000 });
    this.clientScope = Selector(".input-scope div.v-select__selections", { timeout: 15000 });
    this.clientExpires = Selector(".input-expires div.v-select__selections", { timeout: 15000 });
    this.appGenerate = Selector("button.action-generate")
    this.password = Selector(".input-password input", { timeout: 15000 });
    this.appCopy = Selector("button.action-copy")
    this.done = Selector("button.action-done");
    this.appPassword = Selector(".input-app-password input", { timeout: 15000 });
    this.MFAAction = Selector("button.action-passcode-dialog")
    this.setup = Selector("button.action-setup");
    this.qrcode = Selector('img[alt="QR Code"]');
    this.passcode = Selector(".input-code input", { timeout: 15000 });
    this.cancel = Selector("button.action-cancel")
  }
}
