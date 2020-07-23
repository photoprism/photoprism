import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from "./page-model";
import { RequestLogger } from 'testcafe';

const logger = RequestLogger( /http:\/\/localhost\/api\/v1\/*/ , {
    logResponseHeaders: true,
    logResponseBody:    true
});

fixture `Test labels`
    .page`${testcafeconfig.url}`
    .requestHooks(logger);

const page = new Page();

test('#1 Remove/Activate Add/Delete Label from photo', async t => {
    await t.click(Selector('.nav-labels'));
    const countImportantLabels = await Selector('div.p-label').count;
    await t
        .click(Selector('button.action-show-all'));
    const countAllLabels = await Selector('div.p-label').count;
    await t
        .expect(countAllLabels).gt(countImportantLabels)
        .click(Selector('button.action-show-important'));
    await page.search('beacon');
    const LabelBeacon = await Selector('div.p-label').nth(0).getAttribute('data-uid');
    await t
        .click(Selector('div.p-label').withAttribute('data-uid', LabelBeacon));
    const PhotoBeacon = await Selector('div.p-photo').nth(0).getAttribute('data-uid');
    await t
        .click(Selector('.action-title-edit').withAttribute('data-uid', PhotoBeacon));
    const PhotoTitle = await (Selector('.input-title input').value);
    const PhotoKeywords = await (Selector('.input-keywords textarea').value);
    await t
        .expect(PhotoKeywords).contains('beacon')
        .click(Selector('#tab-labels'))
        .click(Selector('button.action-remove'), {timeout: 5000})
        .typeText(Selector('.input-label input'), 'Test')
        .click(Selector('button.p-photo-label-add'))
        .click(Selector('#tab-details'));
    const PhotoTitleAfterEdit = await (Selector('.input-title input').value);
    const PhotoKeywordsAfterEdit = await (Selector('.input-keywords textarea').value);
    await t
        .expect(PhotoKeywordsAfterEdit).contains('test')
        .expect(PhotoTitleAfterEdit).notContains('Beacon')
        .expect(PhotoKeywordsAfterEdit).notContains('beacon')
        .click(Selector('.action-close'))
        .click(Selector('.nav-labels'));
    await page.search('beacon');
    await t
        .expect(Selector('h3').withText('Couldn\'t find anything').visible).ok();
    await page.search('test');
    const LabelTest = await Selector('div.p-label').nth(0).getAttribute('data-uid');
    await t
        .click(Selector('div.p-label').withAttribute('data-uid', LabelTest))
        .click(Selector('.action-title-edit').withAttribute('data-uid', PhotoBeacon))
        .click(Selector('#tab-labels'))
        .click(Selector('.action-delete'), {timeout: 5000})
        .click(Selector('.action-on'))
        .click(Selector('#tab-details'));
    const PhotoTitleAfterUndo = await (Selector('.input-title input').value);
    const PhotoKeywordsAfterUndo = await (Selector('.input-keywords textarea').value);
    await t
        .expect(PhotoKeywordsAfterUndo).contains('beacon')
        .expect(PhotoKeywordsAfterUndo).notContains('test')
        .click(Selector('.action-close'))
        .click(Selector('.nav-labels'));
    await page.search('test');
    await t
        .expect(Selector('h3').withText('Couldn\'t find anything').visible).ok();
    await page.search('beacon');
    await t
        .expect(Selector('div').withAttribute('data-uid', LabelBeacon).visible).ok();
});

//TODO check title of second image after index
test('#2 Rename Label', async t => {
    await t.click(Selector('.nav-labels'));
    await page.search('zebra');
    const LabelZebra = await Selector('div.p-label').nth(0).getAttribute('data-uid');
    await t
        .click(Selector('div.p-label a').nth(0));
    const FirstPhotoZebra = await Selector('div.p-photo').nth(0).getAttribute('data-uid');
    const SecondPhotoZebra = await Selector('div.p-photo').nth(1).getAttribute('data-uid');
    await t
        .click(Selector('.action-title-edit').withAttribute('data-uid', FirstPhotoZebra));
    const FirstPhotoTitle = await (Selector('.input-title input').value);
    const FirstPhotoKeywords = await (Selector('.input-keywords textarea').value);
    await t
        .expect(FirstPhotoTitle).contains('Zebra')
        .expect(FirstPhotoKeywords).contains('zebra')
        .click(Selector('#tab-labels'))
        .click(Selector('div.p-inline-edit'))
        .typeText(Selector('.input-rename input'), 'Horse', { replace: true })
        .pressKey('enter')
        .click(Selector('#tab-details'));
    const FirstPhotoTitleAfterEdit = await (Selector('.input-title input').value);
    const FirstPhotoKeywordsAfterEdit = await (Selector('.input-keywords textarea').value);
    await t
        .expect(FirstPhotoTitleAfterEdit).contains('Horse')
        .expect(FirstPhotoKeywordsAfterEdit).contains('horse')
        .expect(FirstPhotoTitleAfterEdit).notContains('Zebra')
        .click(Selector('.action-close'))
        .click(Selector('.nav-labels'));
    await page.search('horse');
    await t
        .expect(Selector('div').withAttribute('data-uid', LabelZebra).visible).ok()
        .click(Selector('div.p-label').withAttribute('data-uid', LabelZebra))
        .expect(Selector('div').withAttribute('data-uid', SecondPhotoZebra).visible).ok()
        .click(Selector('.action-title-edit').withAttribute('data-uid', FirstPhotoZebra))
        .click(Selector('#tab-labels'))
        .click(Selector('div.p-inline-edit'))
        .typeText(Selector('.input-rename input'), 'Zebra', { replace: true })
        .pressKey('enter')
        .click(Selector('.action-close'))
        .click(Selector('.nav-labels'));
    await page.search('horse');
    await t
        .expect(Selector('h3').withText('Couldn\'t find anything').visible).ok();
});

test('#3 Add label to album', async t => {
    await t
        .click(Selector('.nav-albums'))
        .typeText(Selector('.p-albums-search input'), 'Christmas')
        .pressKey('enter');
    const AlbumUid = await Selector('div.p-album').nth(0).getAttribute('data-uid');
    await t
        .click(Selector('div.p-album').withAttribute('data-uid', AlbumUid));
    const PhotoCount = await Selector('div.p-photo').count;
    await t
        .click(Selector('.nav-labels'));
    await page.search('landscape');
    const LabelLandscape = await Selector('div.p-label').nth(1).getAttribute('data-uid');
    await t
        .click(Selector('div.p-label').withAttribute('data-uid', LabelLandscape));
    const FirstPhotoLandscape = await Selector('div.p-photo').nth(0).getAttribute('data-uid');
    const SecondPhotoLandscape = await Selector('div.p-photo').nth(1).getAttribute('data-uid');
    await t
        .click('.nav-labels');
    await page.selectFromUID(LabelLandscape);

    const clipboardCount = await Selector('span.count-clipboard');
    await t
        .expect(clipboardCount.textContent).eql("1");
    await page.addSelectedToAlbum('Christmas');
    await t
        .click(Selector('.nav-albums'))
        .click(Selector('div.p-album').withAttribute('data-uid', AlbumUid));
    const PhotoCountAfterAdd = await Selector('div.p-photo').count;
    await t
        .expect(PhotoCountAfterAdd).eql(PhotoCount + 2);
    await page.selectFromUID(FirstPhotoLandscape);
    await page.selectFromUID(SecondPhotoLandscape);
    await page.removeSelected();
    await t
        .click('.action-reload');
    const PhotoCountAfterDelete = await Selector('div.p-photo').count;
    await t
        .expect(PhotoCountAfterDelete).eql(PhotoCountAfterAdd - 2);
});

test('#4 Delete label', async t => {
    await t
        .click(Selector('.nav-labels'));
    await page.search('dome');
    const LabelDome = await Selector('div.p-label').nth(0).getAttribute('data-uid');
    await t
        .click(Selector('div.p-label').withAttribute('data-uid', LabelDome));
    const FirstPhotoDome = await Selector('div.p-photo').nth(0).getAttribute('data-uid');
    await t
        .click('.nav-labels')
    await page.selectFromUID(LabelDome);
    const clipboardCount = await Selector('span.count-clipboard');
    await t
        .expect(clipboardCount.textContent).eql("1");
    await page.deleteSelected();
    await page.search('dome');
    await t
        .expect(Selector('h3').withText('Couldn\'t find anything').visible).ok()
        .click('.nav-photos')
        .click(Selector('.action-title-edit').withAttribute('data-uid', FirstPhotoDome))
        .click(Selector('#tab-labels'))
        .expect(Selector('td').withText('No labels found').visible).ok()
        .typeText(Selector('.input-label input'), 'Dome')
        .click(Selector('button.p-photo-label-add'));
});


