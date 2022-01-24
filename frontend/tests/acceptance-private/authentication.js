import { Selector } from "testcafe";
import testcafeconfig from "../acceptance/testcafeconfig";
import Page from "../page-model/page";
import Menu from "../page-model/menu";

fixture`Test authentication`.page`${testcafeconfig.url}`;

const page = new Page();
const menu = new Menu();

test.meta("testID", "authentication-000")(
  "Time to start instance (will be marked as unstable)",
  async (t) => {
    await t.wait(5000);
  }
);

test.meta("testID", "authentication-001")("Login and Logout", async (t) => {
  await t.navigateTo("/browse");
  await t
    .expect(Selector(".input-name input").visible)
    .ok()
    .expect(Selector(".input-search input").visible)
    .notOk()
    .typeText(Selector(".input-name input"), "admin", { replace: true })
    .expect(Selector(".action-confirm").hasAttribute("disabled", "disabled"))
    .ok()
    .typeText(Selector(".input-password input"), "photoprism", { replace: true })
    .expect(Selector(".input-password input").hasAttribute("type", "password"))
    .ok()
    .click(Selector(".v-input__icon--append"))
    .expect(Selector(".input-password input").hasAttribute("type", "text"))
    .ok()
    .click(Selector(".v-input__icon--append"))
    .expect(Selector(".input-password input").hasAttribute("type", "password"))
    .ok()
    .click(Selector(".action-confirm"))
    .expect(Selector(".input-search input", { timeout: 7000 }).visible)
    .ok();
  await menu.openNav();
  await t
    .click(Selector('div[title="Logout"]'))
    .expect(Selector(".input-name input").visible)
    .ok()
    .expect(Selector(".input-search input").visible)
    .notOk();
  await t.navigateTo("/settings");
  await t
    .expect(Selector(".input-name input").visible)
    .ok()
    .expect(Selector(".input-search input").visible)
    .notOk();
});

//TODO test all pages not accessible while logged out

test.meta("testID", "authentication-002")("Login with wrong credentials", async (t) => {
  await page.login("wrong", "photoprism");
  await t
    .navigateTo("/favorites")
    .expect(Selector(".input-name input").visible)
    .ok()
    .expect(Selector(".input-search input").visible)
    .notOk();
  await page.login("admin", "abcdefg");
  await t
    .navigateTo("/archive")
    .expect(Selector(".input-name input").visible)
    .ok()
    .expect(Selector(".input-search input").visible)
    .notOk();
});

test.meta("testID", "authentication-003")("Change password", async (t) => {
  await page.login("admin", "photoprism");
  await t.expect(Selector(".input-search input").visible).ok();
  await page.openNav();
  await t
    .click(Selector(".p-profile"))
    .typeText(Selector(".input-current-password input"), "wrong", { replace: true })
    .typeText(Selector(".input-new-password input"), "photoprism", { replace: true })
    .expect(Selector(".action-confirm").hasAttribute("disabled", "disabled"))
    .ok()
    .typeText(Selector(".input-retype-password input"), "photoprism", { replace: true })
    .expect(Selector(".action-confirm").hasAttribute("disabled", "disabled"))
    .notOk()
    .click(".action-confirm")
    .typeText(Selector(".input-current-password input"), "photoprism", { replace: true })
    .typeText(Selector(".input-new-password input"), "photoprism123", { replace: true })
    .expect(Selector(".action-confirm").hasAttribute("disabled", "disabled"))
    .ok()
    .typeText(Selector(".input-retype-password input"), "photoprism123", { replace: true })
    .expect(Selector(".action-confirm").hasAttribute("disabled", "disabled"))
    .notOk()
    .click(".action-confirm");
  await page.openNav();
  await t.click(Selector('div[title="Logout"]'));
  if (t.browser.platform === "mobile") {
    await t.wait(7000);
  }
  await page.login("admin", "photoprism");
  await t
    .navigateTo("/archive")
    .expect(Selector(".input-name input", { timeout: 7000 }).visible)
    .ok()
    .expect(Selector(".input-search input").visible)
    .notOk();
  await page.login("admin", "photoprism123");
  await t.expect(Selector(".input-search input").visible).ok();
  await page.openNav();
  await t
    .click(Selector(".p-profile", { timeout: 7000 }))
    .typeText(Selector(".input-current-password input"), "photoprism123", { replace: true })
    .typeText(Selector(".input-new-password input"), "photoprism", { replace: true })
    .typeText(Selector(".input-retype-password input"), "photoprism", { replace: true })
    .click(".action-confirm");
  await page.openNav();
  await t.click(Selector('div[title="Logout"]'));
  await page.login("admin", "photoprism");
  await t.expect(Selector(".input-search input").visible).ok();
  await page.openNav();
  await t.click(Selector('div[title="Logout"]'));
});
