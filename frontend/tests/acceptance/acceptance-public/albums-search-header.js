import { ClientFunction, Selector } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";

const origin = process.env.PHOTOPRISM_TEST_URL || new URL(testcafeconfig.url).origin;
const api = `${origin}/api/v1/`;
const locationQuery = ClientFunction(() => window.location.search);
const photos = Selector(".p-page-album-photos .media.result[data-uid]");
const search = Selector(".album-search-field input");

fixture`Album Search and Header`
  .page`${origin}/library/albums`
  .beforeEach(async (t) => {
    const available = await t.request(`${api}photos?count=30&primary=true&public=true&type=image`);
    const selected = [...new Map(available.body.map((photo) => [photo.UID, photo])).values()].slice(0, 12);
    await t.expect(selected.length).gte(2);
    const created = await t.request({ url: `${api}albums`, method: "post", body: { Title: `Album Search Acceptance ${Date.now()}` } });
    t.ctx.album = created.body;
    t.ctx.photos = selected;
    await t.request({ url: `${api}albums/${created.body.UID}/photos`, method: "post", body: { photos: selected.map((photo) => photo.UID) } });
    await t.request({ url: `${api}albums/${created.body.UID}`, method: "put", body: { Description: "A long album description. ".repeat(30), Thumb: selected[0].Hash, ThumbSrc: "manual" } });
    await t.navigateTo(`${origin}/library/albums/${created.body.UID}/view?view=cards`);
    await t.expect(photos.count).eql(selected.length);
  })
  .afterEach(async (t) => {
    if (t.ctx.album) {
      await t.request({ url: `${api}albums/${t.ctx.album.UID}?force=true`, method: "delete" });
    }
  });

test.meta({ testID: "albums-search-001", type: "short", mode: "public" })("Search, sort, reload, and clear stay within the album", async (t) => {
  await t.click(".action-album-search").typeText(search, `uid:${t.ctx.photos[0].UID}`).pressKey("enter");
  await t.expect(photos.count).eql(1);
  await t.click(".action-album-sort").click(".action-album-sort-size");
  await t.expect(locationQuery()).contains("order=size").expect(locationQuery()).contains("q=");
  await t.click(".action-album-sort").click(".action-album-sort-size");
  await t.expect(locationQuery()).contains("reverse=true");
  await t.eval(() => window.location.reload());
  await t.expect(search.value).eql(`uid:${t.ctx.photos[0].UID}`).expect(photos.count).eql(1);
  await t.selectText(search).pressKey("backspace").pressKey("enter");
  await t.expect(photos.count).eql(t.ctx.photos.length).expect(locationQuery()).notContains("q=").expect(locationQuery()).contains("order=size");
  const saved = await t.request(`${api}albums/${t.ctx.album.UID}`);
  await t.expect(saved.body.Order).eql(t.ctx.album.Order);
});

test.meta({ testID: "albums-search-002", type: "short", mode: "public" })("Header expansion, empty results, and compact controls remain usable", async (t) => {
  await t.expect(Selector(".p-album-header h1").innerText).eql(t.ctx.album.Title);
  await t.expect(Selector(".album-header-image").exists).ok();
  const palette = ClientFunction(() => getComputedStyle(document.querySelector(".p-album-header")).getPropertyValue("--album-cover-ink"));
  await t.expect(Selector(".album-header-image").addCustomDOMProperties({ loaded: (image) => image.complete && image.naturalWidth > 0 }).loaded).ok();
  const color = await palette();
  await t.click(".album-header-more");
  await t.expect(Selector(".album-header-more").getAttribute("aria-expanded")).eql("true").expect(palette()).eql(color);
  await t.click(".album-header-more").click(".action-album-search").typeText(search, "zzqqxx").pressKey("enter");
  await t.expect(photos.count).eql(0).expect(Selector(".album-empty-state").visible).ok();
  await t.click(".action-album-clear-search");
  await t.expect(photos.count).eql(t.ctx.photos.length);
  await t.scrollBy(0, 700);
  await t.expect(Selector(".p-album-toolbar .album-toolbar--cover").exists).notOk();
  await t.expect(Selector(".action-album-sort").visible).ok();
});

test.meta({ testID: "albums-search-003", type: "short", mode: "public" })("External cover and membership changes update the open album", async (t) => {
  const cover = Selector(".album-header-image");
  await t.expect(cover.getAttribute("src")).contains(t.ctx.photos[0].Hash);
  await t.request({ url: `${api}albums/${t.ctx.album.UID}`, method: "put", body: { Thumb: t.ctx.photos[1].Hash, ThumbSrc: "manual" } });
  await t.expect(cover.getAttribute("src")).contains(t.ctx.photos[1].Hash);
  await t.request({ url: `${api}albums/${t.ctx.album.UID}/photos`, method: "delete", body: { photos: [t.ctx.photos[0].UID] } });
  await t.expect(photos.count).eql(t.ctx.photos.length - 1);
});

// toolbarOpacity reads the rendered background alpha rather than component state.
const toolbarOpacity = ClientFunction(() => {
  const color = getComputedStyle(document.querySelector(".p-album-toolbar .v-toolbar")).backgroundColor;
  const alpha = color.match(/^rgba\([^,]+,[^,]+,[^,]+,\s*([^)]+)\)$/);
  return alpha ? Number(alpha[1]) : 1;
});

test.meta({ testID: "albums-search-004", type: "short", mode: "public" })("Transparent toolbar fades with scrolling and opens normal search", async (t) => {
  await t.expect(toolbarOpacity()).eql(0);
  await t.click(".action-album-search");
  await t.expect(search.visible).ok().expect(toolbarOpacity()).eql(1);
  await t.click(search).pressKey("esc");
  await t.expect(toolbarOpacity()).eql(0);
  await t.eval(() => {
    const header = document.querySelector(".p-album-header").getBoundingClientRect();
    const toolbar = document.querySelector(".p-album-toolbar .v-toolbar").getBoundingClientRect();
    window.scrollBy({ top: header.bottom - toolbar.bottom - 60, behavior: "instant" });
  });
  await t.expect(toolbarOpacity()).gt(0.1).expect(toolbarOpacity()).lt(0.9);
  await t.scrollBy(0, 400);
  await t.expect(toolbarOpacity()).eql(1);
});
