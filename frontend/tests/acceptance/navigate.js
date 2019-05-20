import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from "./page-model";

fixture`Use navigation`.page`${testcafeconfig.url}`;
const page = new Page();

test('Navigate', async t => {

    await page.openNav();

    await t
        .click('a[href="/places"]')
        .expect(Selector('div.leaflet-map-pane').exists).ok();

    await page.openNav();

    await t
        .click('a[href="/tags"]')
        .expect(Selector('h1').innerText, {timeout: 5000}).contains('Tags');

    await page.openNav();

    await t
        .click('a[href="/albums"]')
        .expect(Selector('h1').innerText).contains('Albums');
    await page.openNav();

    await t
        .click('a[href="/import"]')
        .expect(Selector('h1').innerText).contains('Import');
});
