import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from './page-model';

fixture`Use filters`
    .page`${testcafeconfig.url}`;

const page = new Page();

test('Test camera filter', async t => {
    await t
        .click('#advancedMenu');
        await page.setFilter('camera', 'iPhone 6');
        await page.setFilter('view', 'Details');
    await t
        .expect(Selector('div.v-image__image').visible).ok()
        .expect(Selector('div.caption').visible).ok()
}),
test('Test time filter', async t => {
    await t
        .click('#advancedMenu');
    await page.setFilter('time', 'Oldest');
    await t
        .expect(Selector('div.v-image__image').visible).ok()
}),
    test('Test countries filter', async t => {
        await t
            .click('#advancedMenu');
        await page.setFilter('countries', 'Cuba');
        await t
            .expect(Selector('div.v-image__image').visible).ok()
    },);
