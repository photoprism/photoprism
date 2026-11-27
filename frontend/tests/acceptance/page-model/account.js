import { Selector, t } from "testcafe";

export default class Page {
  constructor() {
    this.changePasswordAction = Selector("button.action-change-password");
    this.currentPassword = Selector(".input-current-password input", { timeout: 15000 });
    this.newPassword = Selector(".input-new-password input", { timeout: 15000 });
    this.retypePassword = Selector(".input-retype-password input", { timeout: 15000 });
    this.confirm = Selector(".action-confirm");
    this.close = Selector(".v-card-actions .action-close");
    this.appsAndDevicesAction = Selector("button.action-apps-dialog");
    this.connectedAppsAction = Selector("button.action-connected-apps-dialog");
    this.connectedAppsDialog = Selector(".p-settings-connected-apps.v-overlay--active");
    this.connectedAppsEmpty = this.connectedAppsDialog.find(".v-data-table-rows-no-data").withExactText("No apps are connected to your account.");
    this.connectedAppsClose = this.connectedAppsDialog.find("button.action-close");
    this.appAdd = Selector("button.action-add");
    this.clientName = Selector(".input-name input:not([type='hidden'])", { timeout: 15000 });
    this.clientScope = Selector(".input-scope div.v-input__control", { timeout: 15000 });
    this.clientExpires = Selector(".input-expires div.v-input__control", { timeout: 15000 });
    this.appGenerate = Selector("button.action-generate");
    this.password = Selector(".input-password input", { timeout: 15000 });
    this.appCopy = Selector("button.action-copy");
    this.done = Selector("button.action-done");
    this.appPassword = Selector(".input-app-password input", { timeout: 15000 });
    this.MFAAction = Selector("button.action-passcode-dialog");
    this.setup = Selector("button.action-setup");
    this.qrcode = Selector('img[alt="QR Code"]');
    this.passcode = Selector(".input-code input", { timeout: 15000 });
    this.cancel = Selector("button.action-cancel");
  }

  // checkEmptyConnectedApps opens the dialog, checks its empty state, and closes it.
  async checkEmptyConnectedApps() {
    await t
      .expect(this.connectedAppsAction.visible)
      .ok()
      .click(this.connectedAppsAction)
      .expect(this.connectedAppsDialog.visible)
      .ok()
      .expect(this.connectedAppsEmpty.visible)
      .ok()
      .click(this.connectedAppsClose)
      .expect(this.connectedAppsDialog.exists)
      .notOk();
  }
}
