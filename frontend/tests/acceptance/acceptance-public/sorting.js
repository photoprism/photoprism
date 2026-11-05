/* global fixture */
import { ClientFunction, RequestLogger, Selector } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";

const requests = RequestLogger((request) => /\/api\/v1\/(photos|albums)\?/.test(request.url), { logResponseBody: true, stringifyResponseBody: true });
const sortButton = Selector("button.action-sort");
const routeQuery = ClientFunction(() => window.location.search);
const sortQuery = ClientFunction(() => {
  const query = new URLSearchParams(window.location.search);
  return { order: query.get("order"), reverse: query.get("reverse") === "true" };
});
const currentSortLabel = ClientFunction(() => document.querySelector("button.action-sort")?.getAttribute("aria-label"));
const firstPhoto = Selector(".search-results .is-photo").nth(0);

fixture`Photo and Album Sorting`.page`${testcafeconfig.url}`.requestHooks(requests);

// openPage opens a deterministic route and waits for its initial search to finish.
async function openPage(t, path) {
  requests.clear();
  await t.navigateTo(new URL(path, new URL("./", testcafeconfig.url)).href);
  await t.expect(sortButton.hasAttribute("disabled")).notOk();
}

// selectSort selects a row or a direct arrow and waits for the resulting search.
async function selectSort(t, value, direction) {
  await t.click(sortButton);
  const row = Selector(`.p-sort-menu .v-list-item[data-sort="${value}"]`);
  const active = await row.getAttribute("aria-pressed") === "true";
  const arrow = Selector(`.p-sort-menu button[data-sort="${value}"][data-direction="ascending"]`);
  const reversible = await arrow.exists;
  const defaultAscending = value === "name" || value === "title";
  const ascending = direction ? direction === "ascending" : active && reversible ? await arrow.getAttribute("aria-pressed") !== "true" : defaultAscending;
  const expected = {
    order: value === "newest" && ascending ? "oldest" : value,
    reverse: value !== "newest" && reversible && ascending !== defaultAscending,
  };
  const changed = !active || reversible && (direction === undefined || await arrow.getAttribute("aria-pressed") !== String(ascending));
  const selector = direction ? `button[data-sort="${value}"][data-direction="${direction}"]` : `.v-list-item[data-sort="${value}"]`;
  await t.click(Selector(".p-sort-menu").find(selector));
  await t.expect(Selector(".p-sort-menu").visible).notOk();
  if (changed) {
    await t.expect(sortQuery()).eql(expected);
    await t.expect(requests.contains((entry) => {
      const query = new URL(entry.request.url).searchParams;
      return !!entry.response && query.get("order") === expected.order && (query.get("reverse") === "true") === expected.reverse;
    })).ok();
  }
  await t.expect(sortButton.hasAttribute("disabled")).notOk();
}

// latestRequest returns the most recent completed search request.
function latestRequest() {
  return requests.requests.filter((entry) => entry.response).at(-1);
}

test.meta({ testID: "sorting-001", type: "short", mode: "public" })("Common: Reverse photo dates and retain the preference", async (t) => {
  await openPage(t, "browse?order=newest&view=cards");
  requests.clear();
  await selectSort(t, "newest");
  await t.expect(routeQuery()).contains("order=oldest");
  await t.expect(currentSortLabel()).eql("Sorted by Date Taken, oldest first");
  const request = latestRequest();
  await t.expect(new URL(request.request.url).searchParams.get("order")).eql("oldest");
  await t.expect(new URL(request.request.url).searchParams.has("reverse")).notOk();
  const body = JSON.parse(request.response.body);
  await t.expect(body.length).gt(0);
  {
    await t.expect(firstPhoto.getAttribute("class")).contains(`uid-${body[0].UID}`);
    const dates = body.map((photo) => photo.TakenAt);
    await t.expect(dates).eql([...dates].sort());
  }
  await openPage(t, "videos?order=size");
  await openPage(t, "browse");
  await t.expect(currentSortLabel()).eql("Sorted by Date Taken, oldest first");
});

test.meta({ testID: "sorting-002", type: "short", mode: "public" })("Common: Select size direction and avoid repeating one-way searches", async (t) => {
  await openPage(t, "browse?order=newest");
  requests.clear();
  await selectSort(t, "size", "ascending");
  await t.expect(routeQuery()).contains("reverse=true");
  await t.expect(currentSortLabel()).eql("Sorted by File Size, smallest first");
  const url = new URL(latestRequest().request.url);
  await t.expect(url.searchParams.get("order")).eql("size");
  await t.expect(url.searchParams.get("reverse")).eql("true");
  const photos = JSON.parse(latestRequest().response.body);
  await t.expect(photos.length).gt(1);
  const sizes = photos.map((photo) => photo.Files.find((file) => file.UID === photo.FileUID)?.Size);
  await t.expect(sizes.every((size) => Number.isFinite(size))).ok();
  await t.expect(sizes).eql([...sizes].sort((a, b) => a - b));
  await t.expect(firstPhoto.getAttribute("class")).contains(`uid-${photos[0].UID}`);

  await selectSort(t, "added");
  requests.clear();
  await selectSort(t, "added");
  await t.expect(requests.requests.length).eql(0);
});

test.meta({ testID: "sorting-003", type: "short", mode: "public" })("Common: Reverse album creation dates and isolate album types", async (t) => {
  await openPage(t, "albums?order=added");
  await selectSort(t, "added");
  await t.expect(currentSortLabel()).eql("Sorted by Date Created, oldest first");
  const url = new URL(latestRequest().request.url);
  await t.expect(url.searchParams.get("order")).eql("added");
  await t.expect(url.searchParams.get("reverse")).eql("true");
  const albums = JSON.parse(latestRequest().response.body);
  await t.expect(albums.length).gt(1);
  const uids = albums.map((album) => album.UID);
  await t.expect(uids).eql([...uids].sort());
  await t.expect(Selector(".album-results .result").nth(0).getAttribute("data-uid")).eql(uids[0]);

  await openPage(t, "moments?order=newest");
  await t.expect(currentSortLabel()).eql("Sorted by Year, newest first");
  await openPage(t, "albums");
  await t.expect(currentSortLabel()).eql("Sorted by Date Created, oldest first");
});

test.meta({ testID: "sorting-004", type: "short", mode: "public" })("Common: Filter media types and remove the All parameter", async (t) => {
  await openPage(t, "browse?type=raw&year=2018&order=newest");
  const filtered = JSON.parse(latestRequest().response.body);
  await t.expect(filtered.length).gt(0);
  await t.expect(filtered.every((photo) => photo.Type === "raw" && photo.Year === 2018)).ok();
  await t.click(Selector('.input-search [aria-label="Search Filters"]'));
  await t.expect(Selector(".p-media-type-select").innerText).contains("RAW");
  await t.expect(Selector(".p-time-select").exists).notOk();
  await t.click(Selector(".p-media-type-select"));
  await t.click(Selector('[role="option"]').withExactText("All"));
  await t.expect(routeQuery()).notContains("type=");
  await t.expect(routeQuery()).contains("year=2018");
  for (const page of ["videos", "live", "animated", "audio"]) {
    await openPage(t, page);
    await t.expect(Selector(".p-media-type-select").exists).notOk();
  }
});

test.meta({ testID: "sorting-005", type: "short", mode: "public" })("Common: Map every media type to the existing search parameter", async (t) => {
  await openPage(t, "browse?order=newest&view=cards");
  for (const [label, type] of [["Photos", "image"], ["Videos", "video"], ["Live", "live"], ["RAW", "raw"], ["Animated", "animated"], ["Vector", "vector"]]) {
    const field = Selector(".p-media-type-select");
    if (!(await field.visible)) {
      await t.click(Selector('.input-search [aria-label="Search Filters"]'));
    }
    requests.clear();
    await t.click(field);
    await t.click(Selector('[role="option"]').withExactText(label));
    await t.expect(routeQuery()).contains(`type=${type}`);
    await t.expect(requests.contains((entry) => !!entry.response && new URL(entry.request.url).searchParams.get("type") === type)).ok();
    await t.expect(sortButton.hasAttribute("disabled")).notOk();
    const result = JSON.parse(latestRequest().response.body);
    await t.expect(result.every((photo) => photo.Type === type)).ok();
    if (["image", "video", "raw"].includes(type)) {
      await t.expect(result.length).gt(0);
    }
  }
});
