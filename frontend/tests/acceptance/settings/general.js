import { Selector } from 'testcafe';
import testcafeconfig from '../testcafeconfig';
import Page from "../page-model";

fixture `Test settings`
    .page`${testcafeconfig.url}`;

const page = new Page();

//TODO test convert to jpeg, group files, places style
//TODO check download also disabled for albums/files/review/private

test('#1 Settings', async t => {
    await page.openNav();
    await t
        .expect(Selector('.action-upload').exists, {timeout: 5000}).ok()
        .expect(Selector('.p-navigation-photos').innerText).contains('Photos');
    await page.selectNthPhoto(0);
    await t
        .click(Selector('button.p-photo-clipboard-menu'))
        .expect(Selector('button.p-photo-clipboard-download').visible).ok()
        .expect(Selector('button.p-photo-clipboard-share').visible).ok()
        .expect(Selector('button.p-photo-clipboard-edit').visible).ok()
        .expect(Selector('button.p-photo-clipboard-private').visible).ok();
    await page.unselectPhoto(0);
    await t
        .click(Selector('div.p-photo').nth(0));
    await t
        .expect(Selector('#p-photo-viewer').visible).ok()
        .expect(Selector('.action-download').exists).ok()
        .click(Selector('.action-close'))
        .expect(Selector('button.action-location').visible).ok()
        .click(Selector('button.action-title-edit').nth(0))
        .expect(Selector('.input-title input').hasAttribute('disabled')).notOk()
        .click(Selector('#tab-edit-labels'))
        .expect(Selector('button.p-photo-label-add').visible).ok()
        .click(Selector('#tab-edit-details'))
        .click(Selector('button.action-close'))
        .click(Selector('.p-navigation-library'))
        .expect(Selector('#tab-import a').visible).ok()
        .expect(Selector('#tab-logs a').visible).ok()
        .click(Selector('div.p-navigation-photos + div'))
        .expect(Selector('.p-navigation-archive').visible).ok()
        .expect(Selector('.p-navigation-review').visible).ok()
        .click(Selector('div.p-navigation-library + div'))
        .expect(Selector('.p-navigation-files').visible).ok()
        .click(Selector('div.p-navigation-albums + div'))
        .expect(Selector('.p-navigation-folders').visible).ok()
        .expect(Selector('.p-navigation-moments').visible).ok()
        .expect(Selector('.p-navigation-labels').visible).ok()
        .expect(Selector('.p-navigation-places').visible).ok()
        .expect(Selector('.p-navigation-private').visible).ok()


        .click(Selector('.p-navigation-settings'))
        .click(Selector('.input-language input'))
        .hover(Selector('div').withText('German').parent('div[role="listitem"]'))
        .click(Selector('div').withText('German').parent('div[role="listitem"]'))
        .click(Selector('.p-navigation-settings'))
        .click(Selector('.input-upload input'))
        .click(Selector('.input-download input'))
        .click(Selector('.input-import input'))
        .click(Selector('.input-archive input'))
        .click(Selector('.input-edit input'))
        .click(Selector('.input-files input'))
        .click(Selector('.input-moments input'))
        .click(Selector('.input-labels input'))
        .click(Selector('.input-logs input'))
        .click(Selector('.input-share input'))
        .click(Selector('.input-places input'))
        .click(Selector('.input-private input'))
        .click(Selector('.input-review input'))
        .click(Selector('.p-navigation-photos'));

    await t.eval(() => location.reload());
    await page.openNav();
    await t
        .expect(Selector('button.action-upload').exists).notOk()
        .expect(Selector('.p-navigation-photos').innerText).contains('Fotos');
    await page.selectNthPhoto(0);
    await t
        .click(Selector('button.p-photo-clipboard-menu'))
        .expect(Selector('button.p-photo-clipboard-download').exists).notOk()
        .expect(Selector('button.p-photo-clipboard-share').exists).notOk()
        .expect(Selector('button.p-photo-clipboard-edit').exists).notOk()
        .expect(Selector('button.p-photo-clipboard-private').exists).notOk();
    await page.unselectPhoto(0);
    await t
        .click(Selector('div.p-photo').nth(0));
    await t
        .expect(Selector('#p-photo-viewer').visible).ok()
        .expect(Selector('.action-download').exists).notOk()
        .click(Selector('.action-close'))
        .expect(Selector('button.action-location').exists).notOk()
        .click(Selector('button.action-title-edit').nth(0))
        .expect(Selector('.input-title input').hasAttribute('disabled')).ok()
        .expect(Selector('.input-latitude input').hasAttribute('disabled')).ok()
        .expect(Selector('.input-timezone input').hasAttribute('disabled')).ok()
        .expect(Selector('.input-country input').hasAttribute('disabled')).ok()
        .expect(Selector('.input-description textarea').hasAttribute('disabled')).ok()
        .expect(Selector('.input-keywords textarea').hasAttribute('disabled')).ok()
        .click(Selector('#tab-edit-labels'))
        .expect(Selector('button.p-photo-label-add').exists).notOk()
        .click(Selector('#tab-edit-details'))
        .click(Selector('button.action-close'))
        .click(Selector('.p-navigation-library'))
        .expect(Selector('#tab-import a').exists).notOk()
        .expect(Selector('#tab-logs a').exists).notOk()
        .click(Selector('div.p-navigation-photos + div'))
        .expect(Selector('.p-navigation-archive').visible).notOk()
        .expect(Selector('.p-navigation-review').exists).notOk()
        .click(Selector('div.p-navigation-library + div'))
        .expect(Selector('.p-navigation-files').visible).notOk()
        .click(Selector('div.p-navigation-albums + div'))
        .expect(Selector('.p-navigation-moments').visible).notOk()
        .expect(Selector('.p-navigation-labels').visible).notOk()
        .expect(Selector('.p-navigation-places').visible).notOk()
        .expect(Selector('.p-navigation-private').visible).notOk()

        .click(Selector('.p-navigation-settings'))
        .click(Selector('.input-language input'))
        .hover(Selector('div').withText('English').parent('div[role="listitem"]'))
        .click(Selector('div').withText('English').parent('div[role="listitem"]'))
        .click(Selector('.p-navigation-settings'))
        .click(Selector('.input-upload input'))
        .click(Selector('.input-download input'))
        .click(Selector('.input-import input'))
        .click(Selector('.input-archive input'))
        .click(Selector('.input-edit input'))
        .click(Selector('.input-files input'))
        .click(Selector('.input-moments input'))
        .click(Selector('.input-labels input'))
        .click(Selector('.input-logs input'))
        .click(Selector('.input-share input'))
        .click(Selector('.input-places input'))
        .click(Selector('.input-private input'))
        .click(Selector('.input-review input'));
});