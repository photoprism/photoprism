/* global fixture */
import { ClientFunction, RequestLogger, Selector } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";
import Page from "../page-model/page";

const page = new Page();
const sortButton = Selector("button.action-sort");
const order = ClientFunction(() => new URLSearchParams(window.location.search).get("order"));
const searches = RequestLogger((request) => /\/api\/v1\/albums\?/.test(request.url));

fixture`Authenticated Sorting`.page`${testcafeconfig.url}`.requestHooks(searches);

test.meta({ testID: "sorting-auth-001", mode: "auth", type: "short" })("Common: Sort after authentication", async (t) => {
  await t.expect(sortButton.exists).notOk();
  await page.login("admin", "photoprism");
  await t.expect(sortButton.visible).ok();
  await t.expect(sortButton.hasAttribute("disabled")).notOk();
  await t.click(sortButton);
  await t.click(Selector('.p-sort-menu .v-list-item[data-sort="name"]'));
  await t.expect(order()).eql("name");
  await t.expect(sortButton.getAttribute("aria-label")).eql("Sorted by File Name, alphabetical");
});

test.meta({ testID: "sorting-auth-002", mode: "auth", type: "short" })("Common: Offer visitor sorting where shared overview search is available", async (t) => {
  await t.navigateTo(new URL("/s/jxoux5ub1e/british-columbia-canada", testcafeconfig.url).href);
  await t.expect(Selector(".p-album-toolbar").visible).ok();
  await t.expect(sortButton.exists).notOk();
  await t.navigateTo(new URL("/library/moments?order=newest", testcafeconfig.url).href);
  await t.expect(Selector(".p-albums-search .input-search").visible).ok();
  await t.expect(sortButton.visible).ok();
  await t.expect(sortButton.hasAttribute("disabled")).notOk();
  searches.clear();
  await t.click(sortButton);
  await t.click(Selector('.p-sort-menu .v-list-item[data-sort="newest"]'));
  await t.expect(order()).eql("oldest");
  await t.expect(searches.contains((entry) => entry.response?.statusCode === 200 && new URL(entry.request.url).searchParams.get("order") === "oldest")).ok();
  await t.expect(sortButton.getAttribute("aria-label")).eql("Sorted by Year, oldest first");
});
