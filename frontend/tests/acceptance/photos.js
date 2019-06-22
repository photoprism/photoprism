import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from "./page-model";

fixture`Test photos page`
    .page`${testcafeconfig.url}`;

const page = new Page();

test('Select photos', async t => {
    const countSelected = await Selector('div.p-photo-clipboard').innerText;
    const countSelectedInt = (Number.isInteger(parseInt(countSelected))) ? parseInt(countSelected) : 0;

    await t
        .hover(Selector('div[class="v-image__image v-image__image--cover"]').nth(0))
        .click(Selector('button.p-photo-select'))
        .hover(Selector('div[class="v-image__image v-image__image--cover"]').nth(2))
        .click(Selector('button.p-photo-select').nth(1));

    const countSelectedAfterLike = await Selector('div.p-photo-clipboard').innerText;
    const countSelectedAfterLikeInt = (Number.isInteger(parseInt(countSelectedAfterLike))) ? parseInt(countSelectedAfterLike) : 0;

    await t
        .expect(countSelectedAfterLikeInt).eql(countSelectedInt + 2)
        .hover(Selector('div[class="v-image__image v-image__image--cover"]').nth(0))
        .click(Selector('button.p-photo-select'))

    const countSelectedAfterDislike = await Selector('div.p-photo-clipboard').innerText;
    const countSelectedAfterDislikeInt = (Number.isInteger(parseInt(countSelectedAfterDislike))) ? parseInt(countSelectedAfterDislike) : 0;

    await t
        .expect(countSelectedAfterDislikeInt).eql(countSelectedAfterLikeInt -1);

    await page.openNav();
    await t
        .click('a[href="/labels"]')
        .expect(Selector('main .p-page-labels').exists, {timeout: 5000}).ok();
    await page.openNav();
    await t
        .click('a[href="/photos"]')
        .expect(countSelectedAfterDislikeInt).eql(countSelectedAfterLikeInt -1);
});
