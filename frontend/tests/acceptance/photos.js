import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from "./page-model";

fixture`Test clipboard`
    .page`${testcafeconfig.url}`;

const page = new Page();

test('Test selecting photos and clear clipboard', async t => {
    const clipboardCount = await Selector('span.t-clipboard-count');

    await page.selectPhoto(0);
    await page.selectPhoto(2);

    await t
        .expect(clipboardCount.textContent).eql("2");
    await page.unselectPhoto(0);

    await t
        .expect(clipboardCount.textContent).eql("1")

    await page.openNav();
    await t
        .click('a[href="/labels"]')
        .expect(Selector('main .p-page-labels').exists, {timeout: 5000}).ok();
    await page.openNav();
    await t
        .click('a[href="/photos"]')
        .expect(clipboardCount.textContent).eql("1")
        .click(Selector('div.p-photo-clipboard'))
        .click(Selector('.p-photo-clipboard-clear'), {timeout: 15000});

    await t.expect(Selector('#t-clipboard').exists).eql(false);
});
