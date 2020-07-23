import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from './page-model';

fixture`Test components`
    .page`${testcafeconfig.url}`;

const page = new Page();

test('#1 Test filter options', async t => {
    await t
        .click('button.p-expand-search')
        .expect(Selector('body').withText('object Object').exists).notOk();
});

test('#2 Fullscreen mode', async t => {
    await t
        .click(Selector('div.v-image__image').nth(0))
        .expect(Selector('#p-photo-viewer').visible).ok()
        .expect(Selector('img.pswp__img').visible).ok();
});

test('#3 Mosaic view', async t => {
    await t
        .click('button.p-expand-search');
    await page.setFilter('view', 'Mosaic');
    await t
        .expect(Selector('div.v-image__image').visible).ok()
        .expect(Selector('div.p-photo-mosaic').visible).ok()
        .expect(Selector('div.p-photo div.caption').exists).notOk()
        .expect(Selector('#p-photo-viewer').visible).notOk();
});

test('#4 List view', async t => {
    await t
        .click('button.p-expand-search');
    await page.setFilter('view', 'List');
    await t
        .expect(Selector('table.v-datatable').visible).ok()
        .expect(Selector('div.p-photo-list').visible).ok();
});

test('#5 card view', async t => {
    await t
        .click('button.p-expand-search');
    await page.setFilter('view', 'Cards');
    await t
        .expect(Selector('div.v-image__image').visible).ok()
        .expect(Selector('div.p-photo div.caption').visible).ok()
        .expect(Selector('#p-photo-viewer').visible).notOk();
});