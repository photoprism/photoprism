import { Selector } from "testcafe";
import testcafeconfig from "../../../testcafeconfig.json";
import Account from "../../page-model/account"
import Menu from "../../page-model/menu";
import Page from "../../page-model/page";
import Settings from "../../page-model/settings";

fixture`Test account settings`.page`${testcafeconfig.url}`;

const menu = new Menu();
const page = new Page();
const account = new Account()
const settings = new Settings();

test.meta("testID", "account-001").meta({ type: "short", mode: "auth" })(
    "Common: Sign in with recovery code",
    async (t) => {
        await page.login("jane", "photoprism")
        await t
            .typeText(page.passcodeInput, "123456", { replace: true })
            .click(account.confirm)
            .expect(page.passcodeInput.visible)
            .ok()
            .expect(Selector(".input-search input").visible)
            .notOk()
            .typeText(account.passcode, "pyuzjwtu8to3", { replace: true })
            .click(account.confirm)
            .expect(page.passcodeInput.visible)
            .notOk();
        await menu.openPage("settings");

        await t
            .click(settings.accountTab)
            .click(account.MFAAction)
            .expect(account.setup.visible)
            .ok()
            .click(account.close);
        await page.logout();
        await page.login("jane", "photoprism");
        await t.expect(page.passcodeInput.visible)
            .notOk();
        await page.logout();
    });

test.meta("testID", "account-002").meta({ type: "short", mode: "auth" })(
    "Core: Create App Password",
    async (t) => {
        await page.login("admin", "photoprism");
        await menu.openPage("settings");
        await t
            .click(settings.accountTab)
            .click(account.appsAndDevicesAction)
            .expect(Selector("td").withText("M@#y-app-passw&^[]/'ord").visible).notOk()
            .click(account.appAdd)
            .typeText(account.clientName, "M$@#y-app-passw*&^[]/'><ord", { replace: true })
            .click(account.clientScope)
            .click(Selector("div").withText("Full Access").parent('div[role="listitem"]'))
            .click(account.clientExpires)
            .click(Selector("div").withText("After 7 days").parent('div[role="listitem"]'))
            .click(account.appGenerate)
            .typeText(account.password, "photoprism", { replace: true })
            .click(account.confirm);

        const AppPassword = await account.appPassword.value;

        await t
            .click(account.appCopy)
            .click(account.done)
            .expect(Selector("td").withText("M@#y-app-passw&^[]/'ord").visible).ok()
            .expect(Selector('tr[data-name="M@#y-app-passw&^[]/\'ord"] td').nth(2).innerText).contains("–")
            .click(account.close);

        await page.logout();
        await page.login("admin", AppPassword);
        await menu.openPage("settings");

        await t
            .click(settings.accountTab)
            .expect(account.changePasswordAction.hasAttribute("disabled")).ok()
            .click(account.appsAndDevicesAction)
            .expect(Selector("td").withText("M@#y-app-passw)(&^{}[];/'ord").visible).notOk()
            .click(account.close);
        await page.logout();
    }
);

test.meta("testID", "account-003").meta({ type: "short", mode: "auth" })(
    "Core: Check App Password has limited permissions and last updated is set",
    async (t) => {
        await page.login("john", "photoprism");
        await menu.openPage("settings");
        await t
            .click(settings.accountTab)
            .click(account.appsAndDevicesAction);
        await t
            .expect(Selector('tr[data-name="john-full"] td').nth(2).innerText).contains("–")
            .expect(Selector('tr[data-name="john-full"] td').nth(2).innerText).notContains("20")
            .click(account.close);
        await page.logout();

        await page.login("john", "Zx78HJ-YFZMjS-aOyR6l-Bz8D2b")
        await menu.openPage("settings");

        await t
            .click(settings.accountTab)
            .expect(account.changePasswordAction.hasAttribute("disabled")).ok()
            .click(account.appsAndDevicesAction)
            .expect(Selector("td").withText("john-full").visible).notOk()
            .click(account.close);
        await page.logout();

        await page.login("john", "photoprism");

        await menu.openPage("settings");

        await t
            .click(settings.accountTab)
            .click(account.appsAndDevicesAction);
        await t
            .expect(Selector('tr[data-name="john-full"] td').nth(2).innerText).contains("20")
            .expect(Selector('tr[data-name="john-full"] td').nth(2).innerText).notContains("–");
        await page.logout();
    }
);

test.meta("testID", "account-004").meta({ type: "short", mode: "auth" })(
    "Core: Try to login with invalid credentials/insufficient scope",
    async (t) => {
        await page.login("john", "5i8yHz-YnvGrU-FMmuxd-s8ziIi")
        await t.navigateTo("/library/favorites");

        await t
            .expect(page.usernameInput.visible)
            .ok()
            .expect(Selector(".input-search input").visible)
            .notOk();

        await page.login("john", "BMGTIe-sw844T-OMsvPZ-7dvDhB")
        await t.navigateTo("/library/favorites");

        await t
            .expect(page.usernameInput.visible)
            .ok()
            .expect(Selector(".input-search input").visible)
            .notOk();
    }
);

test.meta("testID", "account-005").meta({ type: "short", mode: "auth" })(
    "Core: Delete App Password",
    async (t) => {
        await page.login("admin", "5i8yHz-YnvGrU-FMmuxd-s8ziIi")
        await t.navigateTo("/library/favorites");

        await t
            .expect(page.usernameInput.visible)
            .notOk()
            .expect(Selector(".input-search input").visible)
            .ok();
        await page.logout();

        await page.login("admin", "photoprism");

        await menu.openPage("settings");

        await t
            .click(settings.accountTab)
            .click(account.appsAndDevicesAction);
        await t
            .click(Selector('tr[data-name="admin-full"] button.action-remove'))
            .click(account.confirm)
            .expect(Selector("td").withText("admin-full").visible).notOk()
            .click(account.close);

        await page.logout();

        await page.login("admin", "5i8yHz-YnvGrU-FMmuxd-s8ziIi")
        await t.navigateTo("/library/favorites");

        await t
            .expect(page.usernameInput.visible)
            .ok()
            .expect(Selector(".input-search input").visible)
            .notOk();
    }
);

test.meta("testID", "account-006").meta({ type: "short", mode: "auth" })(
    "Common: Try to activate 2FA with wrong password/passcode",
    async (t) => {
        await page.login("admin", "photoprism")
        await menu.openPage("settings");

        await t
            .click(settings.accountTab)
            .click(account.MFAAction);
        await t
            .typeText(account.password, "photo-wrong", { replace: true })
            .click(account.setup)
            .expect(account.qrcode.visible)
            .notOk()
            .typeText(account.password, "photoprism", { replace: true })
            .click(account.setup)
            .expect(account.qrcode.visible)
            .ok()
            .typeText(account.passcode, "123456", { replace: true })
            .click(account.confirm)
            .expect(account.qrcode.visible)
            .ok();
        await page.logout();
    }
);