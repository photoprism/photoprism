import { Selector } from 'testcafe';
import testcafeconfig from '../testcafeconfig';
import Page from "../page-model";

fixture `Test index`
    .page`${testcafeconfig.url}`;

const page = new Page();
//TODO check metadata like camera, keywords, location etc are added
test('#1 Index files from folder', async t => {
    await t
        .click(Selector('.nav-labels'));
    await page.search('cheetah');
    await t
        .expect(Selector('h3').withText('Couldn\'t find anything').visible).ok()
        .click(Selector('.nav-moments'))
        .expect(Selector('h3').withText('Couldn\'t find anything').visible).ok();
    await t
        .click(Selector('.nav-library'))
        .click(Selector('#tab-index'))
        .click(Selector('.input-index-folder input'))
        .click(Selector('div.v-list__tile__title').withText('/moment'))
        .click(Selector('.action-index'))
        //TODO replace wait
        .wait(50000)
        .expect(Selector('span').withText('Done.').visible, {timeout: 60000}).ok()
        .click(Selector('.nav-labels'))
        .click(Selector('.action-reload'));
    await page.search('cheetah');
    await t
        .expect(Selector('.p-label').visible).ok()
        .click(Selector('.nav-moments'))
        .click(Selector('a').withText('South Africa 2013'))
        .expect(Selector('.p-photo').visible).ok();
});