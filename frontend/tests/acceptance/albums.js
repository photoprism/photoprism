import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from "./page-model";
import { RequestLogger } from 'testcafe';

const logger = RequestLogger( /http:\/\/localhost\/api\/v1\/*/ , {
    logResponseHeaders: true,
    logResponseBody:    true
});

fixture `Test albums`
    .page`${testcafeconfig.url}`
    .requestHooks(logger);

const page = new Page();

test('#1 Create/delete album', async t => {
    logger.clear();
    await t.click(Selector('.p-navigation-albums'));
    //TODO fails in container with request but outsides without it
    //const request = await logger.requests[0].response.body;
    const countAlbums = await Selector('div.p-album').count;
    logger.clear();
    await t
        .click(Selector('button.action-add'));
    const request1 = await logger.requests[0].response.body;
    const countAlbumsAfterCreate = await Selector('div.p-album').count;
    const NewAlbum = await Selector('div.p-album').nth(0).getAttribute('data-uid');
    await t
        .expect(countAlbumsAfterCreate).eql(countAlbums + 1);
    await page.selectFromUID(NewAlbum);
    logger.clear();
    await page.deleteSelectedAlbum();
    const request2 = await logger.requests[0].response.body;
    const countAlbumsAfterDelete = await Selector('div.p-album').count;
    await t
        .expect(countAlbumsAfterDelete).eql(countAlbumsAfterCreate - 1);
});

test('#2 Update album', async t => {
    await t
        .click(Selector('.p-navigation-albums'))
        .typeText(Selector('.p-albums-search input'), 'Holiday')
        .pressKey('enter');
    const AlbumUid = await Selector('div.p-album').nth(0).getAttribute('data-uid');
    await t
        .expect(Selector('div.v-card__actions').nth(0).innerText).contains('Holiday')
        .click(Selector('div.v-card__actions').nth(0))
        .click(Selector('.input-title input'))
        .pressKey('ctrl+a delete')
        .typeText(Selector('.input-title input'), 'Animals')
        .pressKey('enter')
        .expect(Selector('div.v-card__actions').nth(0).innerText).contains('Animals');
    logger.clear();
    await t
        .click(Selector('div.p-album').nth(0));
    const request1 = await logger.requests[0].response.body;
    const PhotoCount = await Selector('div.p-photo').count;
    await t
        .click(Selector('.p-expand-search'))
        .expect(Selector('.input-description textarea').value).eql('')
        .expect(Selector('.input-category input').value).eql('')
        .typeText(Selector('.input-description textarea'), 'All my animals')
        .typeText(Selector('.input-category input'), 'Pets')
        .pressKey('enter');
    logger.clear();
    await t
        .click('.action-reload');
    const request2 = await logger.requests[0].response.body;
    await t
        .click(Selector('.p-expand-search'))
        .expect(Selector('.input-description textarea').value).eql('All my animals')
        .expect(Selector('.input-category input').value).eql('Pets')
        .click(Selector('.p-navigation-photos'));
    const FirstPhotoUid = await Selector('div.p-photo').nth(0).getAttribute('data-uid');
    const SecondPhotoUid = await Selector('div.p-photo').nth(1).getAttribute('data-uid');
    await page.selectFromUID(FirstPhotoUid);
    await page.selectFromUID(SecondPhotoUid);
    await page.addSelectedToAlbum('Animals');
    await t
        .click(Selector('.p-navigation-albums'));
    logger.clear();
    await t
        .click(Selector('.input-category'))
        .click(Selector('div[role="listitem"]').withText('Family'));
    const request3 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div.v-card__actions').nth(0).innerText).contains('Christmas')
        .click(Selector('.p-navigation-albums'))
        .click(Selector('.input-category'))
        .click(Selector('div[role="listitem"]').withText('All Categories'), {timeout: 55000});
    const request4 = await logger.requests[0].response.body;
    await t
        .click(Selector('div.p-album').withAttribute('data-uid', AlbumUid));
    const request5 = await logger.requests[0].response.body;
    const PhotoCountAfterAdd = await Selector('div.p-photo').count;
    await t
        .expect(PhotoCountAfterAdd).eql(PhotoCount + 2);
    await page.selectFromUID(FirstPhotoUid);
    await page.selectFromUID(SecondPhotoUid);
    await page.removeSelected();
    await t
        .click('.action-reload');
    const PhotoCountAfterDelete = await Selector('div.p-photo').count;
    logger.clear();
    await t
        .expect(PhotoCountAfterDelete).eql(PhotoCountAfterAdd - 2)
        .click(Selector('.p-expand-search'))
        .click(Selector('.input-description textarea'))
        .pressKey('ctrl+a delete')
        .click(Selector('.input-category input'))
        .pressKey('ctrl+a delete')
        .pressKey('enter')
        .click(Selector('.p-expand-search'))
        .click(Selector('div.p-inline-edit'))
        .typeText(Selector('.input-title input'), 'Holiday', { replace: true })
        .pressKey('enter');
});

//TODO test download itself + clipboard count after download
test('#3 Download album', async t => {
    await t.click(Selector('.p-navigation-albums'));
    const FirstAlbum = await Selector('div.p-album').nth(0).getAttribute('data-uid');
    await page.selectFromUID(FirstAlbum);
    const clipboardCount = await Selector('span.t-clipboard-count');
    await t
        .expect(clipboardCount.textContent).eql("1")
        .click(Selector('button.p-album-clipboard-menu'))
        .expect(Selector('button.p-album-clipboard-download').visible).ok();
});

test('#4 View folders', async t => {
    await page.openNav();
    await t
        .click(Selector('.p-navigation-albums + div'))
        .click(Selector('.p-navigation-folders'))
        .expect(Selector('a').withText('BotanicalGarden').visible).ok()
        .expect(Selector('a').withText('Kanada').visible).ok()
        .expect(Selector('a').withText('KorsikaAdventure').visible).ok();
});

test('#5 View calendar', async t => {
    logger.clear();
    await t
        .click(Selector('.p-navigation-calendar'))
        .expect(Selector('a').withText('May 2019').visible).ok()
        .expect(Selector('a').withText('October 2019').visible).ok();
});