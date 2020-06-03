import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig.json';
import Page from "./page-model";

fixture`Test views`
    .page`${testcafeconfig.url}`;

const page = new Page();

test('Open photo in fullscreen', async t => {
    await t
        .click(Selector('div.v-image__image').nth(0))
        .expect(Selector('#p-photo-viewer').visible).ok()
        .expect(Selector('img.pswp__img').visible).ok();
});

test('Open mosaic view via select', async t => {
    await t
        .click('button.p-expand-search');
    await page.setFilter('view', 'Mosaic');
    await t
        .expect(Selector('div.v-image__image').visible).ok()
        .expect(Selector('div.p-photo-mosaic').visible).ok()
        .expect(Selector('div.p-photo div.caption').exists).notOk()
        .expect(Selector('#p-photo-viewer').visible).notOk();
});

test('Open list view via select', async t => {
    await t
        .click('button.p-expand-search');
    await page.setFilter('view', 'List');
    await t
        .expect(Selector('table.v-datatable').visible).ok()
        .expect(Selector('div.p-photo-list').visible).ok();
});

test('Open card view via select', async t => {
    await t
        .click('button.p-expand-search');
    await page.setFilter('view', 'Cards');
    await t
        .expect(Selector('div.v-image__image').visible).ok()
        .expect(Selector('div.p-photo div.caption').visible).ok()
        .expect(Selector('#p-photo-viewer').visible).notOk();
});
