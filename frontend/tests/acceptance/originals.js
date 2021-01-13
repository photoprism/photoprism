import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from "./page-model";

fixture `Test files`
    .page`${testcafeconfig.url}`;

const page = new Page();

test('#1 Add originals files to album', async t => {
    await page.openNav();
    await t.click(Selector('.nav-albums'));
    await t
        .typeText(Selector('.p-albums-search input'), 'KanadaVacation')
        .pressKey('enter');
    await t
        .expect(Selector('h3').innerText).eql('Couldn\'t find anything');
    await t
        .click(Selector('div.nav-library + div'))
        .click(Selector('.nav-originals'))
        .click(Selector('button').withText('Vacation'));
    const FirstItemInVacation = await Selector('div.result').nth(0).innerText;
    const KanadaUid = await Selector('div.result').nth(0).getAttribute('data-uid');
    const SecondItemInVacation = await Selector('div.result').nth(1).innerText;
    await t
        .expect(FirstItemInVacation).contains('Kanada')
        .expect(SecondItemInVacation).contains('Korsika')
        .click(Selector('button').withText('Kanada'));

    const FirstItemInKanada = await Selector('div.result').nth(0).innerText;
    const SecondItemInKanada = await Selector('div.result').nth(1).innerText;
    await t
        .expect(FirstItemInKanada).contains('BotanicalGarden')
        .expect(SecondItemInKanada).contains('IMG')
        .click(Selector('button').withText('BotanicalGarden'))
        .click(Selector('a[href="/library/files/Vacation"]'));
    await page.selectPhotoFromUID(KanadaUid);
    const clipboardCount = await Selector('span.count-clipboard', {timeout: 5000});
    await t
        .expect(clipboardCount.textContent).eql("1");
    await page.addSelectedToAlbum('KanadaVacation');
    await t
        .click(Selector('.nav-albums'))
        .typeText(Selector('.p-albums-search input'), 'KanadaVacation')
        .pressKey('enter');
    const AlbumUid = await Selector('a.is-album').nth(0).getAttribute('data-uid');
    await t
        .click(Selector('a.is-album').nth(0));
    const PhotoCountAfterAdd = await Selector('div.is-photo', {timeout: 5000}).count;
    await t
        .expect(PhotoCountAfterAdd).eql(2)
        .click(Selector('.nav-albums'));
    await page.selectFromUID(AlbumUid);
    await page.deleteSelected();
});

test('#2 Download original files', async t => {
    await page.openNav();
    await t
        .click(Selector('div.nav-library + div'))
        .click(Selector('.nav-originals'));
    const FirstFile = await Selector('div.is-file').nth(0).getAttribute('data-uid');
    await page.selectPhotoFromUID(FirstFile);
    const clipboardCount = await Selector('span.count-clipboard', {timeout: 5000});
    await t
        .expect(clipboardCount.textContent).eql("1")
        .click(Selector('button.action-menu'))
        .expect(Selector('button.action-download').visible).ok();
});