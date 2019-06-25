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
test('Open details view via button', async t => {
    await t
        .click('button.p-expand-search')
        .click(Selector('i').withText('view_column'))
        .expect(Selector('div.v-image__image').visible).ok()
        .expect(Selector('div.caption').visible).ok()
        .expect(Selector('#p-photo-viewer').visible).notOk()
        .expect(Selector('i').withText('view_column').exists).notOk()
        .expect(Selector('i').withText('view_list').visible).ok()
}),
test('Open mosaic view via select', async t => {
    await t
        .click('button.p-expand-search');
    await page.setFilter('view', 'Mosaic');
    await t
        .expect(Selector('div.v-image__image').visible).ok()
        .expect(Selector('div.p-photo-mosaic').visible).ok()
        .expect(Selector('div.caption').exists).notOk()
        .expect(Selector('#p-photo-viewer').visible).notOk();
}),
test('Open list view via select', async t => {
    await t
        .click('button.p-expand-search');
    await page.setFilter('view', 'List');
    await t
        .expect(Selector('table.v-datatable').visible).ok()
        .expect(Selector('div.v-image__image').exists).notOk()
        .expect(Selector('div.p-photo-list').visible).ok();
}),
    test('Open tile view via select', async t => {
        await t
            .click('button.p-expand-search');
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
