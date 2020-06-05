import { Selector } from 'testcafe';
import testcafeconfig from '../testcafeconfig';
import Page from "../page-model";

fixture `Import file from folder`
    .page`${testcafeconfig.url}`;

const page = new Page();
//TODO use upload + delete
//TODO check metadata like camera, keywords, location etc are added
test('#1 Import files from folder using copy', async t => {
    await t
        .click(Selector('.p-navigation-labels'));
    await page.search('bakery');
    await t
        .expect(Selector('h3').withText('No labels matched your search').visible).ok();
   await t
        .click(Selector('.p-navigation-library'))
        .click(Selector('#tab-import'))
        .click(Selector('.input-import-folder input'))
        .click(Selector('div.v-list__tile__title').withText('/Bäckerei'))
        .click(Selector('.action-import'))
       //TODO replace wait
        .wait(60000)
        //.expect(Selector('span').withText('Done.').visible, {timeout: 60000}).ok()
        .click(Selector('.p-navigation-labels'))
        .click(Selector('.action-reload'));
    await page.search('bakery');
    await t
        .expect(Selector('.p-label').visible).ok();
});