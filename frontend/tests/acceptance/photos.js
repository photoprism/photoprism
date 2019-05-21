import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from "./page-model";

fixture`Test photos page`
    .page`${testcafeconfig.url}`;

const page = new Page();

test('Select photos', async t => {
    await t
        .hover(Selector('div[class="v-image__image v-image__image--cover"]').nth(0))
        .click(Selector('button.p-photo-select'))
        .hover(Selector('div[class="v-image__image v-image__image--cover"]').nth(2))
        .click(Selector('button.p-photo-select').nth(1))
        .hover(Selector('div[class="v-image__image v-image__image--cover"]').nth(3))
        .click(Selector('button.p-photo-select').nth(2))
        .hover(Selector('div[class="v-image__image v-image__image--cover"]').nth(4))
        .click(Selector('button.p-photo-select').nth(3))
        .expect(Selector('div.p-photo-clipboard').innerText).contains('4');
    await page.openNav();
    await t
        .click('a[href="/tags"]')
        .expect(Selector('h1').innerText, {timeout: 5000}).contains('Tags');
    await page.openNav();
    await t
        .click('a[href="/photos"]')
        .expect(Selector('div.p-photo-clipboard').innerText).contains('4');
});
