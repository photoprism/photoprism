import { Selector } from "testcafe";
import testcafeconfig from "../acceptance/testcafeconfig";
import Page from "../acceptance/page-model";

fixture`Test authentication with openid`.page`https://photoprism.reverseproxy.dev`;

const page = new Page();
test.meta("testID", "authentication-000")(
  "Time to start instance (will be marked as unstable)",
  async (t) => {
    await t.wait(5000);
  }
);

test.meta("testID", "authentication-001")("Login and Logout as admin", async (t) => {
  await t
    .expect(Selector("#username").visible)
    .ok()
    .expect(Selector(".input-search input").visible)
    .notOk()
    .typeText(Selector("#username"), "admin", { replace: true })
    .typeText(Selector("#password"), "photoprism", { replace: true })
    .click(Selector("#kc-login"))
    .expect(Selector(".input-search input", { timeout: 7000 }).visible)
    .ok();
  await page.openNav();
  await t
    .click(Selector('div[title="Logout"]'))
    .expect(Selector(".input-name input").visible)
    .ok()
    .expect(Selector(".input-search input").visible)
    .notOk();
  await t.navigateTo("/settings");
  await t
    .expect(Selector("#username").visible)
    .notOk()
    .expect(Selector(".input-search input").visible)
    .ok();
  await t
    /* TODO
    await t.navigateTo("/library");
    .expect(Selector(".input-index-folder input").visible)
    .ok()
    .expect(Selector("div.p-page-photos").visible)
    .notOk()*/
    .click(Selector('div[title="Logout"]'))
    .expect(Selector(".input-name input").visible)
    .ok()
    .navigateTo("https://keycloak.reverseproxy.dev")
    .click(Selector('a[href="https://keycloak.reverseproxy.dev/auth/admin/"]'))
    .click(Selector("a.dropdown-toggle"))
    .click(Selector("a").withText("Sign Out"));
  await t.navigateTo("https://photoprism.reverseproxy.dev/settings");
  await t
    .expect(Selector("#username").visible)
    .ok()
    .expect(Selector(".input-search input").visible)
    .notOk();
});

test.meta("testID", "authentication-001")("Login and Logout as user", async (t) => {
  await t
    .expect(Selector("#username").visible)
    .ok()
    .expect(Selector(".input-search input").visible)
    .notOk()
    .typeText(Selector("#username"), "user", { replace: true })
    .typeText(Selector("#password"), "photoprism", { replace: true })
    .click(Selector("#kc-login"))
    .expect(Selector(".input-search input", { timeout: 7000 }).visible)
    .ok();
  await page.openNav();
  await t
    .click(Selector('div[title="Logout"]'))
    .expect(Selector(".input-name input").visible)
    .ok()
    .expect(Selector(".input-search input").visible)
    .notOk();
  await t.navigateTo("/settings");
  await t
    .expect(Selector("#username").visible)
    .notOk()
    .expect(Selector(".input-search input").visible)
    .ok();
  await t.navigateTo("/library");
  await t
    .expect(Selector(".input-index-folder input").visible)
    .notOk()
    .expect(Selector("div.p-page-photos").visible)
    .ok()
    .click(Selector('div[title="Logout"]'))
    .expect(Selector(".input-name input").visible)
    .ok()
    .navigateTo("https://keycloak.reverseproxy.dev")
    .click(Selector('a[href="https://keycloak.reverseproxy.dev/auth/admin/"]'))
    .click(Selector("a.dropdown-toggle"))
    .click(Selector("a").withText("Sign Out"));
  await t.navigateTo("https://photoprism.reverseproxy.dev/settings");
  await t
    .expect(Selector("#username").visible)
    .ok()
    .expect(Selector(".input-search input").visible)
    .notOk();
});

test.meta("testID", "authentication-002")("Login with wrong credentials", async (t) => {
  await t
    .expect(Selector("#username").visible)
    .ok()
    .expect(Selector(".input-search input").visible)
    .notOk()
    .typeText(Selector("#username"), "admin", { replace: true })
    .typeText(Selector("#password"), "wrong", { replace: true })
    .click(Selector("#kc-login"))
    .expect(Selector(".input-search input", { timeout: 7000 }).visible)
    .notOk();
  await t
    .navigateTo("https://photoprism.reverseproxy.dev/favorites")
    .expect(Selector(".input-name input").visible)
    .ok()
    .expect(Selector(".input-search input").visible)
    .notOk();
});
