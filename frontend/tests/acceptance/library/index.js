import { Selector } from 'testcafe';
import testcafeconfig from '../testcafeconfig';
import Page from "../page-model";

fixture `Test index`
    .page`${testcafeconfig.url}`;

const page = new Page();
test
    .meta('testID', 'library-index-001')
    ('Index files from folder', async t => {
    await t
        .click(Selector('.nav-labels'));
    await page.search('cheetah');
    await t
        .expect(Selector('h3').withText('Couldn\'t find anything').visible).ok()
        .click(Selector('.nav-moments'))
        .expect(Selector('h3').withText('Couldn\'t find anything').visible).ok()
        .click(Selector('.nav-calendar'));
    await page.search('December 2013');
    await t
        .expect(Selector('h3').withText('Couldn\'t find anything').visible).ok()
        .click(Selector('.nav-folders'));
    await page.search('Moment');
    await t
        .expect(Selector('h3').withText('Couldn\'t find anything').visible).ok();
    await page.openNav();
    await t
        .click(Selector('.nav-places + div > i'))
        .click(Selector('.nav-states'));
    await page.search('KwaZulu');
    await t
        .expect(Selector('h3').withText('Couldn\'t find anything').visible).ok();
    await page.openNav();
    await t
        .click(Selector('.nav-library+div>i'))
        .click(Selector('.nav-originals'))
        .click(Selector('.is-folder').withText('moment'))
        .expect(Selector('h3').withText('Couldn\'t find anything').visible).ok();
  await t
        .click(Selector('.nav-library'))
        .click(Selector('#tab-library-index'))
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
        .expect(Selector('.is-label').visible).ok()
        .click(Selector('.nav-moments'))
        .click(Selector('a').withText('South Africa 2013'))
        .expect(Selector('.is-photo').visible).ok()
        .click(Selector('.nav-calendar'))
        .click(Selector('.action-reload'));
    await page.search('December 2013');
    await t
        .expect(Selector('.is-album').visible).ok()
        .click(Selector('.nav-folders'))
        .click(Selector('.action-reload'));
    await page.search('Moment');
    await t
        .expect(Selector('.is-album').visible).ok();
    await page.openNav();
    await t
        .click(Selector('.nav-places+div>i'))
        .click(Selector('.nav-states'))
        .click(Selector('.action-reload'));
    await page.search('KwaZulu');
    await t
        .expect(Selector('.is-album').visible).ok();
    await page.openNav();
    await t
        .click(Selector('.nav-library+div>i'))
        .click(Selector('.nav-originals'))
        .click(Selector('.action-reload'))
        .expect(Selector('.is-folder').withText('moment').visible, {timeout: 60000}).ok();
});