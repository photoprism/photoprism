import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from "./page-model";

fixture`Test views`
    .page`${testcafeconfig.url}`;

const page = new Page();

test('Open photo in fullscreen', async t => {
    await t
        .click(Selector('div.v-image__image').nth(0))
        .expect(Selector('#p-photo-viewer').visible).ok()
        .expect(Selector('img.pswp__img').visible).ok();
}),
test('Open details view', async t => {
    await t
        .click('#advancedMenu');
    await page.setFilter('view', 'Details');
    await t
        .expect(Selector('div.v-image__image').visible).ok()
        .expect(Selector('div.caption').visible).ok()
        .expect(Selector('#p-photo-viewer').visible).notOk()
}),
test('Open mosaic view', async t => {
    await t
        .click('#advancedMenu');
    await page.setFilter('view', 'Mosaic');
    await t
        .expect(Selector('div.v-image__image').visible).ok()
        .expect(Selector('div.p-photo-mosaic').visible).ok()
        .expect(Selector('div.caption').exists).notOk()
        .expect(Selector('#p-photo-viewer').visible).notOk();
}),
test('Open list view', async t => {
    await t
        .click('#advancedMenu');
    await page.setFilter('view', 'List');
    await t
        .expect(Selector('table.v-datatable').visible).ok()
        .expect(Selector('div.v-image__image').exists).notOk()
        .expect(Selector('div.p-photo-list').visible).ok();
}),
    test('Open tile view', async t => {
        await t
            .click('#advancedMenu');
        await page.setFilter('view', 'List');
        await t
            .expect(Selector('div.p-photo-list').visible).ok();
        await page.setFilter('view', 'Tile');
        await t
            .expect(Selector('div.v-image__image').visible).ok()
            .expect(Selector('div.p-photo-tiles').visible).ok()
            .expect(Selector('div.caption').exists).notOk()
            .expect(Selector('#p-photo-viewer').visible).notOk();
    });
