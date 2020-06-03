import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from "./page-model";
import { RequestLogger } from 'testcafe';

const logger = RequestLogger( /http:\/\/localhost:2342\/api\/v1\/*/ , {
    logResponseHeaders: true,
    logResponseBody:    true
});

fixture `Test files`
    .page`${testcafeconfig.url}`
    .requestHooks(logger);

const page = new Page();



//download files

//TODO see files + browse through files + use breadcrumps
test('#1 Add files to album', async t => {
    logger.clear();
    await page.openNav();
    await t.click(Selector('.p-navigation-albums'));
    logger.clear();
    await t
        .typeText(Selector('.p-albums-search input'), 'Christmas')
        .pressKey('enter');
    const AlbumUid = await Selector('div.p-album').nth(0).getAttribute('data-uid');
    logger.clear();
    await t
        .click(Selector('div.p-album').withAttribute('data-uid', AlbumUid));
    const request = await logger.requests[0].response.body;
    const PhotoCount = await Selector('div.p-photo').count;

    await t
        .click(Selector('div.p-navigation-library + div'))
        .click(Selector('.p-navigation-files'));
    const FirstItem = await Selector('div.v-card__title').nth(0).innerText();
    console.log(FirstItem)
    logger.clear();
    await t
        .expect(FirstItem).contains('Vacation')
        .click(Selector('div').withText('Vacation'));
    const request1 = await logger.requests[0].response.body;
    const FirstItemInVacation = await Selector('div.v-card__title').nth(0).innerText();
    const SecondItemInVacation = await Selector('div.v-card__title').nth(1).innerText();
    console.log(FirstItemInVacation);
    console.log(SecondItemInVacation);
    await t
        .expect(FirstItemInVacation).contains('Vacation')
        .expect(SecondItemInVacation).contains('Vacation')
        .click(Selector('div').withText('Vacation'));

    const FirstPhotoFolder = await Selector('div.p-photo').nth(0).getAttribute('data-uid');
    const SecondPhotoFolder = await Selector('div.p-photo').nth(1).getAttribute('data-uid');
    await t
        .click('.p-navigation-labels')
    await page.selectFromUID(LabelLandscape);

    const clipboardCount = await Selector('span.t-clipboard-count');
    await t
        .expect(clipboardCount.textContent).eql("1")
        .click(Selector('button.p-label-clipboard-menu'))
        .click(Selector('button.p-photo-clipboard-album'))
        .typeText(Selector('.input-album input'), 'Christmas', { replace: true })
        .click(Selector('div[role="listitem"]').withText('Christmas'))
        .click(Selector('button.p-photo-dialog-confirm'))
        .click(Selector('.p-navigation-albums'))
        .click(Selector('div.p-album').withAttribute('data-uid', AlbumUid));
    const request4 = await logger.requests[0].response.body;
    const PhotoCountAfterAdd = await Selector('div.p-photo').count;
    await t
        .expect(PhotoCountAfterAdd).eql(PhotoCount + 2);
    await page.selectFromUID(FirstPhotoLandscape);
    await page.selectFromUID(SecondPhotoLandscape);
    await page.removeSelected();
    await t
        .click('.action-reload');
    const PhotoCountAfterDelete = await Selector('div.p-photo').count;
    logger.clear();
    await t
        .expect(PhotoCountAfterDelete).eql(PhotoCountAfterAdd - 2);
});

