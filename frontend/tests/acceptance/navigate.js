import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from "./page-model";

fixture`Use navigation`
    .page`${testcafeconfig.url}`;

const page = new Page();

test('Navigate', async t => {
    await page.openNav();
    await t
        .click('a[href="/albums"]')
        .expect(Selector('div.p-page-albums').exists, {timeout: 5000}).ok();
    await page.openNav();
    await t
        .click('a[href="/places"]')
        .expect(Selector('div.leaflet-map-pane').exists).ok();
    await page.openNav();
    await t
        .click('a[href="/labels"]')
        .expect(Selector('div.p-page-labels').exists, {timeout: 5000}).ok();
});
