/* global fixture */
import testcafeconfig from "../../testcafeconfig.json";
import Album from "../page-model/album";
import Photo from "../page-model/photo";
import Toolbar from "../page-model/toolbar";

const album = new Album();
const photo = new Photo();
const toolbar = new Toolbar();
const requests = toolbar.searchRequests();
const photoSearch = "browse?q=name:albums-002*&view=cards";

fixture`Test photo and album sorting`.page`${testcafeconfig.url}`.requestHooks(requests);

test.meta("testID", "sorting-001").meta({ type: "short", mode: "public" })(
  "Common: Reverse photo dates in both directions and retain the preference",
  async (t) => {
    await toolbar.openSearch(`${photoSearch}&order=newest`);
    await t.expect(photo.firstTitle.innerText).eql("Albums / 2015");

    requests.clear();
    await toolbar.selectSort("newest");
    await toolbar.expectSort(requests, { order: "oldest", label: "Sorted by Date Taken, ascending" });
    await t.expect(photo.firstTitle.innerText).eql("San Diego / United States / 2015");

    await toolbar.openSearch("videos?order=size&reverse=true&view=cards&q=name:Gopher%7Cpanorama_video_360");
    await t.expect(photo.firstTitle.innerText).eql("Website / Berlin / 2020");

    requests.clear();
    await toolbar.openSearch(photoSearch);
    await toolbar.searchResults(requests, { order: "oldest", reverse: null });
    await t.expect(toolbar.sort.getAttribute("title")).eql("Sorted by Date Taken, ascending");
    await t.expect(photo.firstTitle.innerText).eql("San Diego / United States / 2015");

    requests.clear();
    await toolbar.selectSort("newest");
    await toolbar.expectSort(requests, { order: "newest", label: "Sorted by Date Taken, descending" });
    await t.expect(photo.firstTitle.innerText).eql("Albums / 2015");
  }
);

test.meta("testID", "sorting-002").meta({ type: "short", mode: "public" })(
  "Common: Sort photos by file size and recently added",
  async (t) => {
    await toolbar.openSearch(`${photoSearch}&order=newest`);

    requests.clear();
    await toolbar.selectSort("size", { direction: "ascending" });
    await toolbar.expectSort(requests, { order: "size", reverse: "true", label: "Sorted by File Size, ascending" });
    await t.expect(photo.firstTitle.innerText).eql("Tower / Cilaos / 2015");

    requests.clear();
    await toolbar.selectSort("added");
    await toolbar.expectSort(requests, { order: "added", label: "Sorted by Recently Added" });
    await t.expect(photo.firstTitle.innerText).eql("Albums / 2015");

    requests.clear();
    await toolbar.selectSort("added", { search: false });
    await t.expect(requests.count(() => true)).eql(0);
    await t.expect(photo.firstTitle.innerText).eql("Albums / 2015");
  }
);

test.meta("testID", "sorting-003").meta({ type: "short", mode: "public" })(
  "Common: Reverse album creation dates and isolate album types",
  async (t) => {
    await toolbar.openSearch("albums?order=added");
    await t.expect(album.firstTitle.innerText).eql("Garden");

    requests.clear();
    await toolbar.selectSort("added");
    const albums = await toolbar.expectSort(requests, { order: "added", reverse: "true", label: "Sorted by Date Created, ascending" });
    await t.expect(album.firstTitle.innerText).eql("Holiday");
    await t.expect(albums.length).gt(1);
    const dates = albums.map(({ CreatedAt }) => Date.parse(CreatedAt));
    await t.expect(dates.every(Number.isFinite)).ok();
    await t.expect(dates).eql([...dates].sort((a, b) => a - b));

    await toolbar.openSearch("moments?q=201&order=newest");
    await t.expect(toolbar.sort.getAttribute("title")).eql("Sorted by Year, descending");
    await t.expect(album.firstTitle.innerText).eql("Greece 2019");

    requests.clear();
    await toolbar.selectSort("newest");
    await toolbar.expectSort(requests, { order: "oldest", label: "Sorted by Year, ascending" });
    await t.expect(album.firstTitle.innerText).eql("France 2017");

    requests.clear();
    await toolbar.openSearch("albums");
    await toolbar.searchResults(requests, { type: "album", order: "added", reverse: "true" });
    await t.expect(toolbar.sort.getAttribute("title")).eql("Sorted by Date Created, ascending");
    await t.expect(album.firstTitle.innerText).eql("Holiday");
  }
);

test.meta("testID", "sorting-004").meta({ type: "short", mode: "public" })(
  "Common: Clear the media type while retaining the year and hide the filter on fixed-type pages",
  async (t) => {
    await toolbar.openSearch("browse?type=raw&year=2018&order=newest");
    const raw = await toolbar.searchResults(requests, { type: "raw", year: "2018" });
    await t.expect(raw.length).gt(0);
    await t.expect(raw.every(({ Type, Year }) => Type === "raw" && Year === 2018)).ok();

    await toolbar.openFilters();
    await t.expect(toolbar.mediaType.exists).ok();
    await t.expect(toolbar.mediaType.innerText).contains("Raw");

    requests.clear();
    await toolbar.setFilter("type", "All");
    await t.expect(toolbar.searchParam("type")).eql(null).expect(toolbar.searchParam("year")).eql("2018");
    const all = await toolbar.searchResults(requests, { type: "", year: "2018" });
    await t.expect(all.every(({ Year }) => Year === 2018)).ok();
    await t.expect(all.some(({ Type }) => Type !== "raw")).ok();

    for (const path of ["videos", "live", "animated", "audio"]) {
      await toolbar.openSearch(path);
      await toolbar.openFilters();
      await t.expect(toolbar.mediaType.exists).notOk(path);
    }
  }
);

test.meta("testID", "sorting-005").meta({ type: "short", mode: "public" })(
  "Common: Filter images, videos, RAW photos, and documents using the search parameter",
  async (t) => {
    await toolbar.openSearch("browse?order=newest&view=cards");

    for (const [label, type] of [["Image", "image"], ["Video", "video"], ["Raw", "raw"], ["Document", "document"]]) {
      requests.clear();
      await toolbar.setFilter("type", label);
      await t.expect(toolbar.searchParam("type")).eql(type);
      const results = await toolbar.searchResults(requests, { type });
      await t.expect(results.length).gt(0, label);
      await t.expect(results.every((photo) => photo.Type === type)).ok(label);
    }
  }
);

test.meta("testID", "sorting-006").meta({ type: "short", mode: "public" })(
  "Common: Sort photos by title and file name",
  async (t) => {
    await toolbar.openSearch(`${photoSearch}&order=oldest`);
    await t.expect(photo.firstTitle.innerText).eql("San Diego / United States / 2015");

    requests.clear();
    await toolbar.selectSort("title");
    await toolbar.expectSort(requests, { order: "title", label: "Sorted by Title, ascending" });
    await t.expect(photo.firstTitle.innerText).eql("Albums / 2015");

    requests.clear();
    await toolbar.selectSort("name");
    await toolbar.expectSort(requests, { order: "name", label: "Sorted by File Name, ascending" });
    await t.expect(photo.firstTitle.innerText).eql("San Diego / United States / 2015");
  }
);

test.meta("testID", "sorting-007").meta({ type: "short", mode: "public" })(
  "Common: Sort videos by duration in both directions",
  async (t) => {
    await toolbar.openSearch("videos?order=oldest&view=cards&q=name:Gopher%7Cpanorama_video_360");
    await t.expect(photo.firstTitle.innerText).eql("Website / Berlin / 2020");

    requests.clear();
    await toolbar.selectSort("duration");
    await toolbar.expectSort(requests, { order: "duration", label: "Sorted by Video Duration, descending" });
    await t.expect(photo.firstTitle.innerText).eql("Panorama Video / 2020");

    requests.clear();
    await toolbar.selectSort("duration", { direction: "ascending" });
    await toolbar.expectSort(requests, { order: "duration", reverse: "true", label: "Sorted by Video Duration, ascending" });
    await t.expect(photo.firstTitle.innerText).eql("Website / Berlin / 2020");
  }
);

test.meta("testID", "sorting-008").meta({ type: "short", mode: "public" })(
  "Common: Sort albums by name, recent edits, favorites, and location",
  async (t) => {
    await toolbar.openSearch("albums?order=added");

    requests.clear();
    await toolbar.selectSort("name");
    await toolbar.expectSort(requests, { order: "name", label: "Sorted by Name, ascending" });
    await t.expect(album.firstTitle.innerText).eql("Christmas");

    requests.clear();
    await toolbar.selectSort("edited");
    await toolbar.expectSort(requests, { order: "edited", label: "Sorted by Recently Edited" });
    await t.expect(album.firstTitle.innerText).eql("Garden");

    requests.clear();
    await toolbar.selectSort("favorites");
    const favorites = await toolbar.expectSort(requests, { order: "favorites", label: "Sorted by Favorites First" });
    await t.expect(album.firstTitle.innerText).eql("Christmas");
    await t.expect(favorites.slice(0, 3).map(({ Title }) => Title)).eql(["Christmas", "Holiday", "Garden"]);

    await toolbar.openSearch("states?order=name");

    requests.clear();
    await toolbar.selectSort("place");
    await toolbar.expectSort(requests, { order: "place", label: "Sorted by Location" });
    await t.expect(album.firstTitle.innerText).eql("North-West District");
  }
);

test.meta("testID", "sorting-009").meta({ type: "short", mode: "public" })(
  "Common: Reverse calendar albums by date",
  async (t) => {
    await toolbar.openSearch("calendar?q=2020&order=newest");
    await t.expect(album.firstTitle.innerText).eql("December 2020");

    requests.clear();
    await toolbar.selectSort("newest");
    await toolbar.expectSort(requests, { order: "oldest", label: "Sorted by Date, ascending" });
    await t.expect(album.firstTitle.innerText).eql("May 2020");
  }
);
