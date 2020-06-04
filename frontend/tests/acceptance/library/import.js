import { Selector } from 'testcafe';
import testcafeconfig from '../testcafeconfig';
import Page from "../page-model";
import { RequestLogger } from 'testcafe';

const logger = RequestLogger( /http:\/\/localhost:2342\/api\/v1\/*/ , {
    logResponseHeaders: true,
    logResponseBody:    true
});

fixture `Test import`
    .page`${testcafeconfig.url}`
    .requestHooks(logger);

const page = new Page();
//TODO use upload + delete
test('#1 Import files from folder using copy', async t => {
    await t
        .click(Selector('.p-navigation-labels'));
    await page.search('bakery');
    await t
        .expect(Selector('h3').withText('No labels matched your search').visible).ok();
   await t
        .click(Selector('.p-navigation-library'))
        .click(Selector('#tab-import'))
        .expect(Selector('span').withText('Press button to start copying...').visible, {timeout: 5000}).ok()
        .click(Selector('.input-import-folder input'))
        .click(Selector('div.v-list__tile__title').withText('/Bäckerei'))
        .click(Selector('.action-import'))
       //TODO replace wait
        .wait(30000)
        .expect(Selector('span').withText('Done.').visible, {timeout: 60000}).ok()
        .click(Selector('.p-navigation-labels'))
        .click(Selector('.action-reload'));
    await page.search('bakery');
    await t
        .expect(Selector('.p-label').visible).ok();
});