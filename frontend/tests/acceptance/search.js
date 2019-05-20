import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from "./page-model";

fixture`Photos search`.page`${testcafeconfig.url}`;
const page = new Page();

test('Test search object', async t => {
    await page.search('cat');

	await t
        .click('#advancedMenu');

    await page.setFilter('view', 'Details');
    await t
        .expect(Selector('div.v-image__image').visible).ok()
        .expect(Selector('div.caption').visible).ok()
        //.expect(Selector('h3').nth(2).innerText).contains('Egyptian Cat')
        //.expect(Selector('h3').nth(3).innerText).contains('Tabby Cat')
        //.expect(Selector('h3').nth(4).innerText).contains('Tabby Cat');
})/*,
test('Test search color', async t => {
    await page.search('color:pink');

    await t
        .click('#advancedMenu');

    await page.setFilter('view', 'Details');
    await t
        .expect(Selector('h3').nth(0).innerText).contains('Pineapple')
        .expect(Selector('h3').nth(1).innerText).contains('Flamingo');
})*/;
