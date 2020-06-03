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

test('#1 Add files to album', async t => {
    logger.clear();
    await page.openNav();
    await t.click(Selector('.p-navigation-albums'));
    await t
        .typeText(Selector('.p-albums-search input'), 'KanadaVacation')
        .pressKey('enter');
    await t
        .expect(Selector('h3').innerText).eql('No albums matched your search');

    await t
        .click(Selector('div.p-navigation-library + div'))
        .click(Selector('.p-navigation-files'));
    const FirstItem = await Selector('div.v-card__title').nth(0).innerText;
    await t
        .expect(FirstItem).contains('Vacation')
        .click(Selector('button').withText('Vacation'));
    const FirstItemInVacation = await Selector('div.v-card__title').nth(0).innerText;
    const KanadaUid = await Selector('div.v-card__title').nth(0).getAttribute('data-uid');
    const SecondItemInVacation = await Selector('div.v-card__title').nth(1).innerText;
    await t
        .expect(FirstItemInVacation).contains('Kanada')
        .expect(SecondItemInVacation).contains('Korsika')
        .click(Selector('button').withText('Kanada'));

    const FirstItemInKanada = await Selector('div.v-card__title').nth(0).innerText;
    const SecondItemInKanada = await Selector('div.v-card__title').nth(1).innerText;
    await t
        .expect(FirstItemInKanada).contains('BotanicalGarden')
        .expect(SecondItemInKanada).contains('IMG')
        .click(Selector('button').withText('BotanicalGarden'))
        .click(Selector('a[href="/files/Vacation"]'));
    await page.selectFromUID(KanadaUid);
    const clipboardCount = await Selector('span.t-clipboard-count');
    await t
        .expect(clipboardCount.textContent).eql("1")
        .click(Selector('button.p-file-clipboard-menu'))
        .click(Selector('button.p-file-clipboard-album'))
        .typeText(Selector('.input-album input'), 'KanadaVacation', { replace: true })
        .pressKey('enter')
        .click(Selector('button.p-photo-dialog-confirm'))
        .click(Selector('.p-navigation-albums'))
        .typeText(Selector('.p-albums-search input'), 'KanadaVacation')
        .pressKey('enter');
    const AlbumUid = await Selector('div.p-album').nth(0).getAttribute('data-uid');
    await t
        .click(Selector('div.p-album').nth(0));
    const PhotoCountAfterAdd = await Selector('div.p-photo').count;
    await t
        .expect(PhotoCountAfterAdd).eql(2)
        .click(Selector('.p-navigation-albums'));
    await page.selectFromUID(AlbumUid);
    await page.deleteSelectedAlbum();
});

//TODO test download itself + clipboard count after download
test('#2 Download files', async t => {
    await page.openNav();
    await t
        .click(Selector('div.p-navigation-library + div'))
        .click(Selector('.p-navigation-files'));
    const FirstFile = await Selector('div.p-file').nth(0).getAttribute('data-uid');

    await page.selectFromUID(FirstFile);

    const clipboardCount = await Selector('span.t-clipboard-count');
    await t
        .expect(clipboardCount.textContent).eql("1")
        .click(Selector('button.p-file-clipboard-menu'))
        .expect(Selector('button.p-file-clipboard-download').visible).ok();
});

