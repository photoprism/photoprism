/* global fixture */
import testcafeconfig from "../../testcafeconfig.json";
import Album from "../page-model/album";
import Page from "../page-model/page";
import Photo from "../page-model/photo";
import Toolbar from "../page-model/toolbar";

const album = new Album();
const page = new Page();
const photo = new Photo();
const toolbar = new Toolbar();
const requests = toolbar.searchRequests();

fixture`Test sorting after login and on shared albums`.page`${testcafeconfig.url}`.requestHooks(requests);

test.meta("testID", "sorting-auth-001").meta({ type: "short", mode: "auth" })(
  "Common: Sort by file name after authentication",
  async (t) => {
    await t.expect(toolbar.sort.exists).notOk();
    await page.login("admin", "photoprism");
    await t.expect(toolbar.sort.exists).ok();
    await toolbar.openSearch("browse?q=name:albums-002*&view=cards&order=newest");
    await t.expect(photo.firstTitle.innerText).eql("Albums / 2015");

    requests.clear();
    await toolbar.selectSort("name");
    await toolbar.expectSort(requests, { order: "name", label: "Sorted by File Name, ascending" });
    await t.expect(photo.firstTitle.innerText).eql("San Diego / United States / 2015");
  }
);

test.meta("testID", "sorting-auth-002").meta({ type: "short", mode: "auth" })(
  "Common: Sort the visitor's shared album overview by oldest creation date",
  async (t) => {
    await toolbar.openPage("/s/2t6124pb6d/holiday");
    await t.expect(toolbar.sharedAlbumToolbar.visible).ok().expect(toolbar.sort.exists).notOk();
    await toolbar.openPage("/s/garden/garden");
    await t.expect(toolbar.sharedAlbumToolbar.visible).ok();

    await toolbar.openSearch("albums?order=added");
    await t.expect(toolbar.sort.exists).ok().expect(album.firstTitle.innerText).eql("Garden");

    requests.clear();
    await toolbar.selectSort("added");
    await toolbar.expectSort(requests, { order: "added", reverse: "true", label: "Sorted by Date Created, ascending" });
    await t.expect(album.firstTitle.innerText).eql("Holiday");
  }
);
