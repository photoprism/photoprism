import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from "./page-model";

fixture`Test clipboard`
    .page`${testcafeconfig.url}`;

const page = new Page();

test('Test selecting photos and clear clipboard', async t => {
    const countSelected = await Selector('div.p-photo-clipboard').innerText;
    const countSelectedInt = (Number.isInteger(parseInt(countSelected))) ? parseInt(countSelected) : 0;

    await page.selectPhoto(0);
    await page.selectPhoto(2);

    const countSelectedAfterSelect = await Selector('div.p-photo-clipboard').innerText;
    const countSelectedAfterSelectInt = (Number.isInteger(parseInt(countSelectedAfterSelect))) ? parseInt(countSelectedAfterSelect) : 0;

    await t
        .expect(countSelectedAfterSelectInt).eql(countSelectedInt + 2)
    await page.unselectPhoto(0);

    const countSelectedAfterUnselect = await Selector('div.p-photo-clipboard').innerText;
    const countSelectedAfterUnselectInt = (Number.isInteger(parseInt(countSelectedAfterUnselect))) ? parseInt(countSelectedAfterUnselect) : 0;

    await t
        .expect(countSelectedAfterUnselectInt).eql(countSelectedAfterSelectInt -1);

    await page.openNav();
    await t
        .click('a[href="/labels"]')
        .expect(Selector('main .p-page-labels').exists, {timeout: 5000}).ok();
    await page.openNav();
    await t
        .click('a[href="/photos"]')
        .expect(countSelectedAfterUnselectInt).eql(countSelectedAfterSelectInt -1)
        .click(Selector('div.p-photo-clipboard'))
        .click(Selector('.p-photo-clipboard-clear'), {timeout: 15000});
    const countSelectedAfterClear = await Selector('div.p-photo-clipboard').innerText;
    await t
        .expect(countSelectedAfterClear).contains('menu');
});
