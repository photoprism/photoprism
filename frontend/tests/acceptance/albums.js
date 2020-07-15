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
    await t.click(Selector('.nav-albums'));
    //TODO fails in container with request but outsides without it
    //const request = await logger.requests[0].response.body;
    const countAlbums = await Selector('div.p-album').count;
    logger.clear();
    await t
        .click(Selector('button.action-add'));
    //const request1 = await logger.requests[0].response.body;
    const countAlbumsAfterCreate = await Selector('div.p-album').count;
    const NewAlbum = await Selector('div.p-album').nth(0).getAttribute('data-uid');
    await t
        .expect(countAlbumsAfterCreate).eql(countAlbums + 1);
    await page.selectFromUID(NewAlbum);
    logger.clear();
    await page.deleteSelected();
    //const request2 = await logger.requests[0].response.body;
    const countAlbumsAfterDelete = await Selector('div.p-album').count;
    await t
        .expect(countAlbumsAfterDelete).eql(countAlbumsAfterCreate - 1);
});

test('#2 Update album', async t => {
    await t
        .click(Selector('.nav-albums'))
        .typeText(Selector('.p-albums-search input'), 'Holiday')
        .pressKey('enter');
    const AlbumUid = await Selector('div.p-album').nth(0).getAttribute('data-uid');
    await t
        .expect(Selector('h3.action-title-edit').nth(0).innerText).contains('Holiday')
        .click(Selector('.action-title-edit').nth(0))
        .typeText(Selector('.input-title input'), 'Animals', { replace: true })
        .expect(Selector('.input-description textarea').value).eql('')
        .expect(Selector('.input-category input').value).eql('')
        .typeText(Selector('.input-description textarea'), 'All my animals')
        .typeText(Selector('.input-category input'), 'Pets')
        .pressKey('enter')
        .click('.action-confirm')
        .click(Selector('div.p-album').nth(0));
    //const request1 = await logger.requests[0].response.body;
    const PhotoCount = await Selector('div.p-photo').count;
    await t
        .expect(Selector('.v-card__text').nth(0).innerText).contains('All my animals')
        .expect(Selector('div').withText("Animals").exists).ok()
        .click(Selector('.nav-photos'));
    const FirstPhotoUid = await Selector('div.p-photo').nth(0).getAttribute('data-uid');
    const SecondPhotoUid = await Selector('div.p-photo').nth(1).getAttribute('data-uid');
    await page.selectFromUID(FirstPhotoUid);
    await page.selectFromUID(SecondPhotoUid);
    await page.addSelectedToAlbum('Animals');
    await t
        .click(Selector('.nav-albums'));
    logger.clear();
    await t
        .click(Selector('.input-category i'))
        .click(Selector('div[role="listitem"]').withText('Family'));
    //const request3 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('h3.action-title-edit').nth(0).innerText).contains('Christmas')
        .click(Selector('.nav-albums'))
        .click('.action-reload')
        .click(Selector('.input-category i'))
        .click(Selector('div[role="listitem"]').withText('All Categories'), {timeout: 55000});
    //const request4 = await logger.requests[0].response.body;
    await t
        .click(Selector('div.p-album').withAttribute('data-uid', AlbumUid));
    //const request5 = await logger.requests[0].response.body;
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
        .click(Selector('.action-edit'))
        .typeText(Selector('.input-title input'), 'Holiday', { replace: true })
        .expect(Selector('.input-description textarea').value).eql('All my animals')
        .expect(Selector('.input-category input').value).eql('Pets')
        .click(Selector('.input-description textarea'))
        .pressKey('ctrl+a delete')
        .pressKey('enter')
        .click(Selector('.input-category input'))
        .pressKey('ctrl+a delete')
        .pressKey('enter')
        .click('.action-confirm')
        .click('.action-reload')
        .click(Selector('.nav-albums'))
        .expect(Selector('div').withText("Holiday").visible).ok()
        .expect(Selector('div').withText("Animals").exists).notOk();
});

//TODO test download itself + clipboard count after download
test('#3 Download album', async t => {
    await t.click(Selector('.nav-albums'));
    const FirstAlbum = await Selector('div.p-album').nth(0).getAttribute('data-uid');
    await page.selectFromUID(FirstAlbum);
    const clipboardCount = await Selector('span.count-clipboard');
    await t
        .expect(clipboardCount.textContent).eql("1")
        .click(Selector('button.action-menu'))
        .expect(Selector('button.action-download').visible).ok();
});

test('#4 View folders', async t => {
    await page.openNav();
    await t
        .click(Selector('.nav-albums + div'))
        .click(Selector('.nav-folders'))
        .expect(Selector('a').withText('BotanicalGarden').visible).ok()
        .expect(Selector('a').withText('Kanada').visible).ok()
        .expect(Selector('a').withText('KorsikaAdventure').visible).ok();
});

test('#5 View calendar', async t => {
    logger.clear();
    await t
        .click(Selector('.nav-calendar'))
        .expect(Selector('a').withText('May 2019').visible).ok()
        .expect(Selector('a').withText('October 2019').visible).ok();
});

//TODO test that sharing link works as expected
test('#6 Create sharing link', async t => {
    await t.click(Selector('.nav-albums'));
    const FirstAlbum = await Selector('div.p-album').nth(0).getAttribute('data-uid');
    await page.selectFromUID(FirstAlbum);
    const clipboardCount = await Selector('span.count-clipboard');
    await t
        .expect(clipboardCount.textContent).eql("1")
        .click(Selector('button.action-menu'))
        .click(Selector('button.action-share'))
        .click(Selector('div.v-expansion-panel__header__icon').nth(0));
    const InitialUrl = await (Selector('.action-url').innerText);
    const InitialSecret = await (Selector('.input-secret input').value)
    const InitialExpire = await (Selector('div.v-select__selections').innerText);
    await t
        .expect(InitialUrl).notContains('secretfortesting')
        .expect(InitialExpire).eql('Never')
        .typeText(Selector('.input-secret input'), 'secretForTesting', { replace: true })
        .click(Selector('.input-expires input'))
        .click(Selector('div').withText('After 1 day').parent('div[role="listitem"]'))
        .click(Selector('button.action-save'))
        .click(Selector('button.action-close'))
        .click(Selector('button.action-share'))
        .click(Selector('div.v-expansion-panel__header__icon').nth(0));
    const UrlAfterChange = await (Selector('.action-url').innerText);
    const ExpireAfterChange = await (Selector('div.v-select__selections').innerText);
    await t
        .expect(UrlAfterChange).contains('secretfortesting')
        .expect(ExpireAfterChange).eql('After 1 day')
        .typeText(Selector('.input-secret input'), InitialSecret, { replace: true })
        .click(Selector('.input-expires input'))
        .click(Selector('div').withText('Never').parent('div[role="listitem"]'))
        .click(Selector('button.action-save'))
        .click(Selector('div.v-expansion-panel__header__icon'));
    const LinkCount = await (Selector('.action-url').count);
    await t
        .click('.action-add-link');
    const LinkCountAfterAdd = await (Selector('.action-url').count);
    await t
        .expect(LinkCountAfterAdd).eql(LinkCount + 1)
        .click(Selector('div.v-expansion-panel__header__icon'))
        .click(Selector('.action-delete'));
    const LinkCountAfterDelete = await (Selector('.action-url').count);
    await t
        .expect(LinkCountAfterDelete).eql(LinkCountAfterAdd - 1)
});