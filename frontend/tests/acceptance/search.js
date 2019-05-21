import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from "./page-model";

fixture`Search photos`
    .page`${testcafeconfig.url}`;

const page = new Page();

test('Test search object', async t => {
    await page.search('cat');
    await t
        .expect(Selector('div.v-image__image').visible).ok();
}),
test('Test search color', async t => {
    await page.search('color:pink');
    await t
        .expect(Selector('div.v-image__image').visible).ok();
});
