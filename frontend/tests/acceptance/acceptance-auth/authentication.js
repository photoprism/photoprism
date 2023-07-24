import { Selector } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";
import Page from "../page-model/page";
import Account from "../page-model/account";
import Settings from "../page-model/settings";
import Menu from "../page-model/menu";
import Photo from "../page-model/photo";
import ContextMenu from "../page-model/context-menu";

fixture`Test authentication`.page`${testcafeconfig.url}`;

const page = new Page();
const account = new Account();
const contextmenu = new ContextMenu();
const menu = new Menu();
const settings = new Settings();
const photo = new Photo();

test.meta("testID", "authentication-001").meta({ type: "short", mode: "auth" })(
  "Common: Login and Logout",
  async (t) => {
    await t.navigateTo("/library/browse");

    await t
      .expect(page.usernameInput.visible)
      .ok()
      .expect(Selector(".input-search input").visible)
      .notOk();

    await t.typeText(page.usernameInput, "admin", { replace: true });

    await t.expect(page.loginAction.hasAttribute("disabled", "disabled")).ok();

    await t.typeText(page.passwordInput, "photoprism", { replace: true });

    await t.expect(page.passwordInput.hasAttribute("type", "password")).ok();

    await t.click(page.togglePasswordMode);

    await t.expect(page.passwordInput.hasAttribute("type", "text")).ok();

    await t.click(page.togglePasswordMode);

    await t.expect(page.passwordInput.hasAttribute("type", "password")).ok();

    await t.click(page.loginAction);

    await t.expect(Selector(".input-search input", { timeout: 7000 }).visible).ok();

    await page.logout();

    await t
      .expect(page.usernameInput.visible)
      .ok()
      .expect(Selector(".input-search input").visible)
      .notOk();

    await t.navigateTo("/library/settings");
    await t
      .expect(page.usernameInput.visible)
      .ok()
      .expect(Selector(".input-search input").visible)
      .notOk();
  }
);

test.meta("testID", "authentication-002").meta({ type: "short", mode: "auth" })(
  "Common: Login with wrong credentials",
  async (t) => {
    await page.login("wrong", "photoprism");
    await t.navigateTo("/library/favorites");

    await t
      .expect(page.usernameInput.visible)
      .ok()
      .expect(Selector(".input-search input").visible)
      .notOk();

    await page.login("admin", "abcdefg");
    await t.navigateTo("/library/archive");

    await t
      .expect(page.usernameInput.visible)
      .ok()
      .expect(Selector(".input-search input").visible)
      .notOk();
  }
);

test.meta("testID", "authentication-003").meta({ type: "short", mode: "auth" })(
  "Common: Change password",
  async (t) => {
    await t.navigateTo("/library/browse");
    await page.login("admin", "photoprism");
    await t.expect(Selector(".input-search input", { timeout: 15000 }).visible).ok();
    await menu.openPage("settings");

    await t
      .click(settings.accountTab)
      .click(account.changePasswordAction)
      .typeText(account.currentPassword, "wrong", { replace: true })
      .typeText(account.newPassword, "photoprism", { replace: true });

    await t.expect(account.confirm.hasAttribute("disabled", "disabled")).ok();

    await t.typeText(account.retypePassword, "photoprism", { replace: true });

    await t.expect(account.confirm.hasAttribute("disabled", "disabled")).notOk();

    await t
      .click(account.confirm)
      .typeText(account.currentPassword, "photoprism", { replace: true })
      .typeText(account.newPassword, "123", { replace: true })
      .typeText(account.retypePassword, "123", { replace: true });

    await t.expect(account.confirm.hasAttribute("disabled", "disabled")).ok();

    await t
      .typeText(account.currentPassword, "photoprism", { replace: true })
      .typeText(account.newPassword, "photoprism123", { replace: true });

    await t.expect(account.confirm.hasAttribute("disabled", "disabled")).ok();

    await t.typeText(account.retypePassword, "photoprism123", { replace: true });

    await t.expect(account.confirm.hasAttribute("disabled", "disabled")).notOk();

    await t.click(account.confirm);
    await page.logout();
    if (t.browser.platform === "mobile") {
      await t.wait(7000);
    }
    await page.login("admin", "photoprism");
    await t.navigateTo("/library/archive");

    await t
      .expect(page.usernameInput.visible)
      .ok()
      .expect(Selector(".input-search input").visible)
      .notOk();

    await page.login("admin", "photoprism123");
    await t.expect(Selector(".input-search input").visible).ok();
    await menu.openPage("settings");

    await t
      .click(settings.accountTab)
      .click(account.changePasswordAction)
      .typeText(account.currentPassword, "photoprism123", { replace: true })
      .typeText(account.newPassword, "photoprism", { replace: true })
      .typeText(account.retypePassword, "photoprism", { replace: true })
      .click(account.confirm);
    await page.logout();
    await page.login("admin", "photoprism");

    await t.expect(Selector(".input-search input").visible).ok();
    await page.logout();
  }
);

test.meta("testID", "authentication-004").meta({ type: "short", mode: "auth" })(
  "Common: Delete Clipboard on logout",
  async (t) => {
    await page.login("admin", "photoprism");
    await t.navigateTo("/library/browse");
    await photo.toggleSelectNthPhoto(0, "all");
    await photo.toggleSelectNthPhoto(1, "all");
    await contextmenu.checkContextMenuCount("2");
    await page.logout();
    await page.login("admin", "photoprism");
    await t.navigateTo("/library/browse");
    await t.expect(Selector("button.action-menu").visible).notOk();
    await t.expect(Selector("span.count-clipboard").visible).notOk();
  }
);
