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
        .expect(Selector('.nav-photos').innerText).contains('Photos');
    await page.selectNthPhoto(0);
    await t
        .click(Selector('button.action-menu'))
        .expect(Selector('button.action-download').visible).ok()
        .expect(Selector('button.action-share').visible).ok()
        .expect(Selector('button.action-edit').visible).ok()
        .expect(Selector('button.action-private').visible).ok();
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
        .click(Selector('#tab-labels'))
        .expect(Selector('button.p-photo-label-add').visible).ok()
        .click(Selector('#tab-details'))
        .click(Selector('button.action-close'))
        .click(Selector('.nav-library'))
        .expect(Selector('#tab-import a').visible).ok()
        .expect(Selector('#tab-logs a').visible).ok()
        .click(Selector('div.nav-photos + div'))
        .expect(Selector('.nav-archive').visible).ok()
        .expect(Selector('.nav-review').visible).ok()
        .click(Selector('div.nav-library + div'))
        .expect(Selector('.nav-originals').visible).ok()
        .click(Selector('div.nav-albums + div'))
        .expect(Selector('.nav-folders').visible).ok()
        .expect(Selector('.nav-moments').visible).ok()
        .expect(Selector('.nav-labels').visible).ok()
        .expect(Selector('.nav-places').visible).ok()
        .expect(Selector('.nav-private').visible).ok()


        .click(Selector('.nav-settings'))
        .click(Selector('.input-language input'))
        .hover(Selector('div').withText('German').parent('div[role="listitem"]'))
        .click(Selector('div').withText('German').parent('div[role="listitem"]'))
        .click(Selector('.nav-settings'))
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
        .click(Selector('.nav-photos'));

    await t.eval(() => location.reload());
    await page.openNav();
    await t
        .expect(Selector('button.action-upload').exists).notOk()
        .expect(Selector('.nav-photos').innerText).contains('Fotos');
    await page.selectNthPhoto(0);
    await t
        .click(Selector('button.action-menu'))
        .expect(Selector('button.action-download').exists).notOk()
        .expect(Selector('button.action-share').exists).notOk()
        .expect(Selector('button.action-edit').visible).notOk()
        .expect(Selector('button.action-private').exists).notOk();
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
        .click(Selector('#tab-labels'))
        .expect(Selector('button.p-photo-label-add').exists).notOk()
        .click(Selector('#tab-details'))
        .click(Selector('button.action-close'))
        .click(Selector('.nav-library'))
        .expect(Selector('#tab-import a').exists).notOk()
        .expect(Selector('#tab-logs a').exists).notOk()
        .click(Selector('div.nav-photos + div'))
        .expect(Selector('.nav-archive').visible).notOk()
        .expect(Selector('.nav-review').exists).notOk()
        .click(Selector('div.nav-library + div'))
        .expect(Selector('.nav-originals').visible).notOk()
        .click(Selector('div.nav-albums + div'))
        .expect(Selector('.nav-moments').visible).notOk()
        .expect(Selector('.nav-labels').visible).notOk()
        .expect(Selector('.nav-places').visible).notOk()
        .expect(Selector('.nav-private').visible).notOk()

        .click(Selector('.nav-settings'))
        .click(Selector('.input-language input'))
        .hover(Selector('div').withText('Englisch').parent('div[role="listitem"]'))
        .click(Selector('div').withText('Englisch').parent('div[role="listitem"]'))
        .click(Selector('.nav-settings'))
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