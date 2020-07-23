import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from "./page-model";
import { RequestLogger } from 'testcafe';
import { ClientFunction } from 'testcafe';

const logger = RequestLogger( /http:\/\/localhost\/api\/v1\/photos*/ , {
    logResponseHeaders: true,
    logResponseBody:    true
});

const scroll = ClientFunction((x, y) => window.scrollTo(x, y));
const getcurrentPosition = ClientFunction(() => window.pageYOffset);

fixture `Test photos`
    .page`${testcafeconfig.url}`
    .requestHooks(logger);

const page = new Page();

//TODO raw file icon? live photo icon? - search for type look for icon --> do it in search test
//TODO Share photo
//TODO Check count in navi gets updated --> gt/lt or matches count of images
//TODO videos - play video

test('#1 Scroll to top', async t => {
    await t
        .click(Selector('.nav-photos'))
        .click(Selector('.p-expand-search'));
    await page.setFilter('view', 'Cards');
    await t
        .expect(Selector('button.p-photo-scroll-top').exists).notOk()
        .expect(getcurrentPosition()).eql(0)
        .expect(Selector('div[class="v-image__image v-image__image--cover"]').nth(0).visible).ok();
    await scroll(0, 1400);
    await scroll(0, 1000);
    await t
        .click(Selector('button.p-scroll-top'))
        .expect(getcurrentPosition()).eql(0);
});

//TODO test download itself + clipboard count after download
test('#2 Download single photo/video and download zip using clipboard and fullscreen mode', async t => {
    const FirstPhoto = await Selector('div.p-photo').nth(0).getAttribute('data-uid');
    const SecondPhoto = await Selector('div.p-photo').nth(1).getAttribute('data-uid');
    await t
        .click(Selector('div').withAttribute('data-uid', SecondPhoto));
    await t
        .expect(Selector('#p-photo-viewer').visible).ok()
        .hover(Selector('.action-download'))
        .expect(Selector('.action-download').visible).ok()
        .click(Selector('.action-close'));
    await page.selectFromUID(FirstPhoto);
    await t
        .click(Selector('.nav-video'));
    const FirstVideo = await Selector('div.p-photo').nth(0).getAttribute('data-uid');
    await page.selectFromUID(FirstVideo);
    const clipboardCount = await Selector('span.count-clipboard');
    await t
        .expect(clipboardCount.textContent).eql("2")
        .click(Selector('button.action-menu'))
        .expect(Selector('button.action-download').visible).ok();
});

//TODO add part for video as well
//TODO approve/archive from card view
test('#3 Approve photo using approve and by adding location', async t => {
    await page.openNav();
    await t
        .click(Selector('div.nav-photos + div'))
        .click(Selector('.nav-review'));
    logger.clear();
    await page.search('type:image');
    //const request1 = await logger.requests[0].response.body;
    const FirstPhoto = await Selector('div.p-photo').nth(0).getAttribute('data-uid');
    const SecondPhoto = await Selector('div.p-photo').nth(1).getAttribute('data-uid');

    await t
        .click(Selector('.nav-photos'));
    //const request2 = await logger.requests[0].response.body;
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).exists, {timeout: 5000}).notOk()
        .expect(Selector('div').withAttribute('data-uid', SecondPhoto).exists, {timeout: 5000}).notOk();
    logger.clear();
    await t.click(Selector('.nav-review'));
    //const request3 = await logger.requests[0].response.body;

    await page.selectFromUID(FirstPhoto);
    await page.editSelected();
    logger.clear();
    await t
        .click(Selector('button.action-close'));
    await t
        .click(Selector('button.action-reload'));
    //const request4 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).visible, {timeout: 5000}).ok();
    await page.editSelected();
    //const request5 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .click(Selector('button.action-approve'))
        .click(Selector('button.action-done'));
    //const request6 = await logger.requests[0].response.body;
    logger.clear();

    await page.unselectFromUID(FirstPhoto);
    await page.selectFromUID(SecondPhoto);
    await page.editSelected();
    logger.clear();
    await t
        .typeText(Selector('input[aria-label="Latitude"]'), '9.999')
        .typeText(Selector('input[aria-label="Longitude"]'), '9.999')
        .click(Selector('button.action-done'));
    //const request7 = await logger.requests[0].response.body;

    await t
        .click(Selector('button.action-reload'));
    //const request8 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).exists, {timeout: 5000}).notOk()
        .expect(Selector('div').withAttribute('data-uid', SecondPhoto).exists, {timeout: 5000}).notOk()
        .click(Selector('.nav-photos'));
    //const request9 = await logger.requests[0].response.body;
    logger.clear();
    await page.search('type:image');
    //const request10 = await logger.requests[0].response.body;
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).visible).ok()
        .expect(Selector('div').withAttribute('data-uid', SecondPhoto).visible).ok();
});

test('#4 Like/dislike photo/video', async t => {

    logger.clear();
    const FirstPhoto = await Selector('i.t-off').nth(0).getAttribute('data-uid');

    await t.click(Selector('.nav-video'));
    //const request1 = await logger.requests[0].response.body;
    const FirstVideo = await Selector('.t-off').nth(0).getAttribute('data-uid');

    await t.click(Selector('.nav-favorites'));
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).exists, {timeout: 5000}).notOk()
        .expect(Selector('div').withAttribute('data-uid', FirstVideo).exists, {timeout: 5000}).notOk()
        .click(Selector('.nav-photos'));

    logger.clear();
    await page.likePhoto(FirstPhoto);
    //const request2 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .click(Selector('.action-reload'))
        .expect(Selector('i.t-on').withAttribute('data-uid', FirstPhoto).exists, {timeout: 5000}).ok();
    logger.clear();

    await t.click(Selector('.nav-video'));
    await page.likePhoto(FirstVideo);
    //const request3 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .click(Selector('.action-reload'))
        .expect(Selector('i.t-on').withAttribute('data-uid', FirstVideo).exists, {timeout: 5000}).ok();
    logger.clear();

    await t
        .click(Selector('.nav-favorites'));
    //const request4 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).exists, {timeout: 5000}).ok()
        .expect(Selector('div').withAttribute('data-uid', FirstVideo).exists, {timeout: 5000}).ok()
        .expect(Selector('div.v-image__image').visible).ok();
    await page.dislikePhoto(FirstVideo);
    //const request5 = await logger.requests[0].response.body;
    logger.clear();
    await page.dislikePhoto(FirstPhoto);
    logger.clear();
    await t.click(Selector('.action-reload'));
    //const request6 = await logger.requests[0].response.body;
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).exists, {timeout: 5000}).notOk()
        .expect(Selector('div').withAttribute('data-uid', FirstVideo).exists, {timeout: 5000}).notOk();
});

//TODO Check private photos do not appear in labels, places, albums, moments...
test('#5 Private/unprivate photo/video using clipboard and list', async t => {
    await t
        .click(Selector('.nav-photos'))
        .click(Selector('.p-expand-search'));
    logger.clear();
    await page.setFilter('view', 'Mosaic');
    //const request1 = await logger.requests[0].response.body;
    const FirstPhoto = await Selector('div.p-photo').nth(0).getAttribute('data-uid');
    const SecondPhoto = await Selector('div.p-photo').nth(1).getAttribute('data-uid');
    const ThirdPhoto = await Selector('div.p-photo').nth(2).getAttribute('data-uid');

    await t
        .click(Selector('.nav-video'));
    //const request2 = await logger.requests[0].response.body;
    const FirstVideo = await Selector('div.p-photo').nth(0).getAttribute('data-uid');
    const SecondVideo = await Selector('div.p-photo').nth(1).getAttribute('data-uid');

    await t.click(Selector('.nav-private'));
    //const request3 = await logger.requests[0].response.body;
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).exists, {timeout: 5000}).notOk()
        .expect(Selector('div').withAttribute('data-uid', SecondPhoto).exists, {timeout: 5000}).notOk()
        .expect(Selector('div').withAttribute('data-uid', ThirdPhoto).exists, {timeout: 5000}).notOk()
        .expect(Selector('div').withAttribute('data-uid', FirstVideo).exists, {timeout: 5000}).notOk()
        .expect(Selector('div').withAttribute('data-uid', SecondVideo).exists, {timeout: 5000}).notOk()
        .click(Selector('.nav-photos'));
    await page.selectFromUID(FirstPhoto);
    await page.selectFromUID(SecondPhoto);
    const clipboardCount = await Selector('span.count-clipboard');
    await t
        .expect(clipboardCount.textContent).eql("2")
        .click(Selector('button.action-menu'))
        .click(Selector('button.action-private'))
        .expect(Selector('button.action-menu').exists, {timeout: 5000}).notOk();
    logger.clear();
    await page.setFilter('view', 'List');
    //const request4 = await logger.requests[0].response.body;
    await t
        .click(Selector('button.p-photo-private').withAttribute('data-uid', ThirdPhoto));
    logger.clear();
    await t
        .click(Selector('.action-reload'));
    //const request5 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('td').withAttribute('data-uid', FirstPhoto).exists, {timeout: 5000}).notOk()
        .expect(Selector('td').withAttribute('data-uid', SecondPhoto).exists, {timeout: 5000}).notOk()
        .expect(Selector('td').withAttribute('data-uid', ThirdPhoto).exists, {timeout: 5000}).notOk()
        .click(Selector('.nav-video'));

    logger.clear();
    await t
        .click(Selector('button.p-photo-private').withAttribute('data-uid', SecondVideo));
    await page.setFilter('view', 'Mosaic');
    //const request6 = await logger.requests[0].response.body;
    logger.clear();

    await page.selectFromUID(FirstVideo);
    const clipboardCountVideo = await Selector('span.count-clipboard');
    await t
        .expect(clipboardCountVideo.textContent).eql("1")
        .click(Selector('button.action-menu'))
        .click(Selector('button.action-private'))
        .expect(Selector('button.action-menu').exists, {timeout: 5000}).notOk();
    logger.clear();
    await t
        .click(Selector('.action-reload'));
    //const request7 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstVideo).exists, {timeout: 5000}).notOk()
        .expect(Selector('div').withAttribute('data-uid', SecondVideo).exists, {timeout: 5000}).notOk()
        .click(Selector('.nav-private'));

    //const request8 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).exists, {timeout: 5000}).ok()
        .expect(Selector('div').withAttribute('data-uid', SecondPhoto).exists, {timeout: 5000}).ok()
        .expect(Selector('div').withAttribute('data-uid', FirstVideo).exists, {timeout: 5000}).ok()
        .expect(Selector('div').withAttribute('data-uid', SecondVideo).exists, {timeout: 5000}).ok();
    await page.selectFromUID(FirstPhoto);
    await page.selectFromUID(SecondPhoto);
    await page.selectFromUID(FirstVideo);
    await t
        .click(Selector('button.action-menu'))
        .click(Selector('button.action-private'));
    await page.setFilter('view', 'List');
    await t
        .click(Selector('button.p-photo-private').withAttribute('data-uid', ThirdPhoto))
        .click(Selector('button.p-photo-private').withAttribute('data-uid', SecondVideo));
    await page.setFilter('view', 'Mosaic');
    logger.clear();
    await t.click(Selector('.action-reload'));
    //const request9 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).exists, {timeout: 5000}).notOk()
        .expect(Selector('div').withAttribute('data-uid', SecondPhoto).exists, {timeout: 5000}).notOk()
        .expect(Selector('div').withAttribute('data-uid', ThirdPhoto).exists, {timeout: 5000}).notOk()
        .expect(Selector('div').withAttribute('data-uid', FirstVideo).exists, {timeout: 5000}).notOk()
        .expect(Selector('div').withAttribute('data-uid', SecondVideo).exists, {timeout: 5000}).notOk()
        .click(Selector('.nav-photos'));

    //const request10 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).exists, {timeout: 5000}).ok()
        .expect(Selector('div').withAttribute('data-uid', SecondPhoto).exists, {timeout: 5000}).ok()
        .expect(Selector('div').withAttribute('data-uid', ThirdPhoto).exists, {timeout: 5000}).ok()
        .click(Selector('.nav-video'));

    //const request11 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstVideo).exists, {timeout: 5000}).ok()
        .expect(Selector('div').withAttribute('data-uid', SecondVideo).exists, {timeout: 5000}).ok();
});

test('#6 Archive/restore video, photos, private photos and review photos using clipboard', async t => {
    await page.openNav();
    await t
        .click(Selector('.nav-photos'))
        .click(Selector('.p-expand-search'));
    logger.clear();
    await page.setFilter('view', 'Mosaic');
    //const request1 = await logger.requests[0].response.body;
    const FirstPhoto = await Selector('div.p-photo').nth(0).getAttribute('data-uid');
    const SecondPhoto = await Selector('div.p-photo').nth(1).getAttribute('data-uid');

    await t
        .click(Selector('.nav-video'));
    //const request2 = await logger.requests[0].response.body;
    const FirstVideo = await Selector('div.p-photo').nth(0).getAttribute('data-uid');
    logger.clear();

    await t
        .click(Selector('.nav-private'));
    //const request3 = await logger.requests[0].response.body;
    const FirstPrivatePhoto = await Selector('div.p-photo').nth(0).getAttribute('data-uid');
    logger.clear();

    await t
        .click(Selector('div.nav-photos + div'))
        .click(Selector('.nav-review'));
    //const request4 = await logger.requests[0].response.body;
    const FirstReviewPhoto = await Selector('div.p-photo').nth(0).getAttribute('data-uid');
    logger.clear();

    await t
        .click(Selector('.nav-archive'));
    //const request5 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).exists, {timeout: 5000}).notOk()
        .expect(Selector('div').withAttribute('data-uid', SecondPhoto).exists, {timeout: 5000}).notOk()
        .expect(Selector('div').withAttribute('data-uid', FirstVideo).exists, {timeout: 5000}).notOk()
        .expect(Selector('div').withAttribute('data-uid', FirstPrivatePhoto).exists, {timeout: 5000}).notOk()
        .expect(Selector('div').withAttribute('data-uid', FirstReviewPhoto).exists, {timeout: 5000}).notOk();

    await t
        .click(Selector('.nav-video'));
    //const request6 = await logger.requests[0].response.body;
    logger.clear();
    await page.selectFromUID(FirstVideo);
    const clipboardCountVideo = await Selector('span.count-clipboard');
    await t
        .expect(clipboardCountVideo.textContent).eql("1");
    await page.archiveSelected();
    logger.clear();
    await t
        .expect(Selector('button.action-menu').exists, {timeout: 5000}).notOk()
        .click(Selector('.action-reload'));
    //const request7 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstVideo).exists, {timeout: 5000}).notOk()
        .click(Selector('.nav-photos'));

    //const request8 = await logger.requests[0].response.body;
    logger.clear();
    await page.selectFromUID(FirstPhoto);
    await page.selectFromUID(SecondPhoto);
    const clipboardCountPhotos = await Selector('span.count-clipboard');
    await t
        .expect(clipboardCountPhotos.textContent).eql("2");
    await page.archiveSelected();
    logger.clear();
    await t
        .expect(Selector('button.action-menu').exists, {timeout: 5000}).notOk()
        .click(Selector('.action-reload'));
    //const request9 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).exists, {timeout: 5000}).notOk()
        .expect(Selector('div').withAttribute('data-uid', SecondPhoto).exists, {timeout: 5000}).notOk()
        .click(Selector('.nav-private'));

    //const request10 = await logger.requests[0].response.body;
    logger.clear();
    await page.selectFromUID(FirstPrivatePhoto);
    const clipboardCountPrivate = await Selector('span.count-clipboard');
    await t
        .expect(clipboardCountPrivate.textContent).eql("1");
    await t
        .click(Selector('.nav-review'));

    //const request11 = await logger.requests[0].response.body;
    logger.clear();
    await page.selectFromUID(FirstReviewPhoto);
    const clipboardCountReview = await Selector('span.count-clipboard');
    await t
        .expect(clipboardCountReview.textContent).eql("2");
    await page.archiveSelected();
    logger.clear();
    await t
        .expect(Selector('button.action-menu').exists, {timeout: 5000}).notOk()
        .click(Selector('.action-reload'));
    //const request12 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstReviewPhoto).exists, {timeout: 5000}).notOk()
        .click(Selector('.nav-archive'));
    //const request13 = await logger.requests[0].response.body;
    logger.clear();

    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).exists, {timeout: 5000}).ok()
        .expect(Selector('div').withAttribute('data-uid', SecondPhoto).exists, {timeout: 5000}).ok()
        .expect(Selector('div').withAttribute('data-uid', FirstVideo).exists, {timeout: 5000}).ok()
        .expect(Selector('div').withAttribute('data-uid', FirstPrivatePhoto).exists, {timeout: 5000}).ok()
        .expect(Selector('div').withAttribute('data-uid', FirstReviewPhoto).exists, {timeout: 5000}).ok();
    await page.selectFromUID(FirstPhoto);
    await page.selectFromUID(SecondPhoto);
    await page.selectFromUID(FirstVideo);
    await page.selectFromUID(FirstPrivatePhoto);
    await page.selectFromUID(FirstReviewPhoto);
    const clipboardCountArchive = await Selector('span.count-clipboard');
    await t
        .expect(clipboardCountArchive.textContent).eql("5");
    await page.restoreSelected();
    logger.clear();
    await t
        .expect(Selector('button.action-menu').exists, {timeout: 5000}).notOk()
        .click(Selector('.action-reload'));
    //const request14 = await logger.requests[0].response.body;
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).exists, {timeout: 5000}).notOk()
        .expect(Selector('div').withAttribute('data-uid', SecondPhoto).exists, {timeout: 5000}).notOk()
        .expect(Selector('div').withAttribute('data-uid', FirstVideo).exists, {timeout: 5000}).notOk()
        .expect(Selector('div').withAttribute('data-uid', FirstPrivatePhoto).exists, {timeout: 5000}).notOk()
        .expect(Selector('div').withAttribute('data-uid', FirstReviewPhoto).exists, {timeout: 5000}).notOk();
    logger.clear();

    await t
        .click(Selector('.nav-video'));
    //const request15 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstVideo).exists, {timeout: 5000}).ok();

    await t
        .click(Selector('.nav-photos'));
    //const request16 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).exists, {timeout: 5000}).ok()
        .expect(Selector('div').withAttribute('data-uid', SecondPhoto).exists, {timeout: 5000}).ok();

    await t
        .click(Selector('.nav-private'));
    //const request17 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPrivatePhoto).exists, {timeout: 5000}).ok();

    await t
        .click(Selector('.nav-review'));
    //const request18 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstReviewPhoto).exists, {timeout: 5000}).ok();
});

//TODO edited values stay after reindex!!
//TODO test camera, lens
//TODO access video from list + edit
test('#7 Edit photo/video', async t => {
    await page.openNav();
    await t
        .click(Selector('.nav-photos'))
        .click(Selector('.p-expand-search'));
    await page.setFilter('view', 'Cards');
    const FirstPhoto = await Selector('div.p-photo').nth(0).getAttribute('data-uid');
    await t
        .click(Selector('button.action-title-edit').withAttribute('data-uid', FirstPhoto))
        .expect(Selector('input[aria-label="Latitude"]').visible).ok()
        .expect(Selector('button.action-previous').getAttribute('disabled')).eql('disabled');
        logger.clear();
    await t
        .click(Selector('button.action-next'));
    //const request1 = await logger.requests[0].response.body;
    await t
        .expect(Selector('button.action-previous').getAttribute('disabled')).notEql('disabled')
        .click(Selector('button.action-previous'))
        .click(Selector('button.action-close'))
        .click(Selector('div.p-photo').withAttribute('data-uid', FirstPhoto))
        .expect(Selector('#p-photo-viewer').visible).ok()
        .hover(Selector('.action-edit'))
        .click(Selector('.action-edit'))
        .expect(Selector('input[aria-label="Latitude"]').visible).ok();

    const FirstPhotoTitle = await (Selector('.input-title input').value);
    const FirstPhotoLocalTime = await (Selector('.input-local-time input').value);
    const FirstPhotoUTCTime = await (Selector('.input-utc-time input').value);
    const FirstPhotoDay = await (Selector('.input-day input').value);
    const FirstPhotoMonth = await (Selector('.input-month input').value);
    const FirstPhotoYear = await (Selector('.input-year input').value);
    const FirstPhotoTimezone = await (Selector('.input-timezone input').value);
    const FirstPhotoLatitude = await (Selector('.input-latitude input').value);
    const FirstPhotoLongitude = await (Selector('.input-longitude input').value);
    const FirstPhotoAltitude = await (Selector('.input-altitude input').value);
    const FirstPhotoCountry = await (Selector('.input-country input').value);
    const FirstPhotoCamera = await (Selector('div.p-camera-select div.v-select__selection').innerText);
    const FirstPhotoIso = await (Selector('.input-iso input').value);
    const FirstPhotoExposure = await (Selector('.input-exposure input').value);
    const FirstPhotoLens = await (Selector('div.p-lens-select div.v-select__selection').innerText);
    const FirstPhotoFnumber = await (Selector('.input-fnumber input').value);
    const FirstPhotoFocalLength = await (Selector('.input-focal-length input').value);
    const FirstPhotoSubject = await (Selector('.input-subject textarea').value);
    const FirstPhotoArtist = await (Selector('.input-artist input').value);
    const FirstPhotoCopyright = await (Selector('.input-copyright input').value);
    const FirstPhotoLicense = await (Selector('.input-license textarea').value);
    const FirstPhotoDescription = await (Selector('.input-description textarea').value);
    const FirstPhotoKeywords = await (Selector('.input-keywords textarea').value);
    const FirstPhotoNotes = await (Selector('.input-notes textarea').value);

    logger.clear();
    await t
        .typeText(Selector('.input-title input'), 'Not saved photo title', { replace: true })
        .click(Selector('button.action-close'))
        .click(Selector('button.action-date-edit').withAttribute('data-uid', FirstPhoto))
        .expect(Selector('.input-title input').value).eql(FirstPhotoTitle)
        .typeText(Selector('.input-title input'), 'New Photo Title', { replace: true })
        .typeText(Selector('.input-timezone input'), 'Europe/Mosc', { replace: true })
        .click(Selector('div').withText('Europe/Moscow').parent('div[role="listitem"]'))
        .typeText(Selector('.input-day input'), '15', { replace: true })
        .pressKey('enter')
        .typeText(Selector('.input-month input'), '07', { replace: true })
        .pressKey('enter')
        .typeText(Selector('.input-year input'), 'Unknown', { replace: true })
        .pressKey('enter')
        .click(Selector('.input-local-time input'))
        .pressKey('ctrl+a delete')
        .typeText(Selector('.input-local-time input'), '04:30:30', { replace: true })
        .pressKey('enter')
        .typeText(Selector('.input-latitude input'), '41.15333', { replace: true })
        .typeText(Selector('.input-longitude input'), '20.168331', { replace: true })
        .typeText(Selector('.input-altitude input'), '-1', { replace: true })
        .click(Selector('.input-country input'))
        .click(Selector('div').withText('Afghanistan').parent('div[role="listitem"]'))
        //.click(Selector('.input-camera input'))
        //.hover(Selector('div').withText('Apple iPhone 6').parent('div[role="listitem"]'))
        //.click(Selector('div').withText('Apple iPhone 6').parent('div[role="listitem"]'))
        //.click(Selector('.input-lens input'))
        //.click(Selector('div').withText('Apple iPhone 5s back camera 4.15mm f/2.2').parent('div[role="listitem"]'))
        .typeText(Selector('.input-iso input'), '32', { replace: true })
        .typeText(Selector('.input-exposure input'), '1/32', { replace: true })
        .typeText(Selector('.input-fnumber input'), '29', { replace: true })
        .typeText(Selector('.input-focal-length input'), '33', { replace: true })
        .typeText(Selector('.input-subject textarea'), 'Super nice edited photo', { replace: true })
        .typeText(Selector('.input-artist input'), 'Happy', { replace: true })
        .typeText(Selector('.input-copyright input'), 'Happy2020', { replace: true })
        .typeText(Selector('.input-license textarea'), 'Super nice cat license', { replace: true })
        .typeText(Selector('.input-description textarea'), 'Description of a nice image :)', { replace: true })
        .typeText(Selector('.input-keywords textarea'), ', cat, love')
        .typeText(Selector('.input-notes textarea'), 'Some notes', { replace: true })
        .click(Selector('button.action-approve'));
    //const request2 = await logger.requests[0].response.body;
    await t
        .expect(Selector('.input-latitude input').visible, {timeout: 5000}).ok()
        .click(Selector('button.action-done'));
    logger.clear();
    await t
        .click(Selector('button.action-reload'));
    //const request3 = await logger.requests[0].response.body;
    await t
        .expect(Selector('button.action-title-edit').withAttribute('data-uid', FirstPhoto).innerText).eql('New Photo Title')

    await page.selectFromUID(FirstPhoto);
    await page.editSelected();
    await t
        .expect(Selector('.input-title input').value).eql('New Photo Title')
        .expect(Selector('.input-timezone input').value).eql('Europe/Moscow')
        .expect(Selector('.input-local-time input').value).eql('04:30:30')
        .expect(Selector('.input-utc-time input').value).eql('01:30:30')
        .expect(Selector('.input-day input').value).eql('15')
        .expect(Selector('.input-month input').value).eql('07')
        .expect(Selector('.input-year input').value).eql('Unknown')
        .expect(Selector('.input-latitude input').value).eql('41.15333')
        .expect(Selector('.input-longitude input').value).eql('20.168331')
        .expect(Selector('.input-altitude input').value).eql('-1')
        .expect(Selector('div').withText('Albania').visible).ok()
        .expect(Selector('div').withText('Afghanistan').visible).notOk()
        //.expect(Selector('div').withText('Apple iPhone 6').visible).ok()
        //.expect(Selector('div').withText('Apple iPhone 5s back camera 4.15mm f/2.2').visible).ok()
        .expect(Selector('.input-iso input').value).eql('32')
        .expect(Selector('.input-exposure input').value).eql('1/32')
        .expect(Selector('.input-fnumber input').value).eql('29')
        .expect(Selector('.input-focal-length input').value).eql('33')
        .expect(Selector('.input-subject textarea').value).eql('Super nice edited photo')
        .expect(Selector('.input-artist input').value).eql('Happy')
        .expect(Selector('.input-copyright input').value).eql('Happy2020')
        .expect(Selector('.input-license textarea').value).eql('Super nice cat license')
        .expect(Selector('.input-description textarea').value).eql('Description of a nice image :)')
        .expect(Selector('.input-description textarea').value).eql('Description of a nice image :)')
        .expect(Selector('.input-notes textarea').value).contains('Some notes')
        .expect(Selector('.input-keywords textarea').value).contains('cat')
    if (FirstPhotoTitle.empty || FirstPhotoTitle === "")
    { await t
        .click(Selector('.input-title input'))
        .pressKey('ctrl+a delete')}
    else
    {await t.typeText(Selector('.input-title input'), FirstPhotoTitle, { replace: true })}
    await t.typeText(Selector('.input-day input'), FirstPhotoDay, { replace: true })
        .pressKey('enter')
        .typeText(Selector('.input-month input'), FirstPhotoMonth, { replace: true })
        .pressKey('enter')
        .typeText(Selector('.input-year input'), FirstPhotoYear, { replace: true })
        .pressKey('enter');
    if (FirstPhotoLocalTime.empty || FirstPhotoLocalTime === "")
    { await t
        .click(Selector('.input-local-time input'))
        .pressKey('ctrl+a delete')}
    else
    {await t
        .click(Selector('.input-local-time input'))
        .pressKey('ctrl+a delete')
        .typeText(Selector('.input-local-time input'), FirstPhotoLocalTime, { replace: true })
        .pressKey('enter')}
     if (FirstPhotoTimezone.empty || FirstPhotoTimezone === "")
    { await t
        .click(Selector('.input-timezone input'))
        .typeText(Selector('.input-timezone input'), 'UTC', { replace: true })
        .pressKey('enter')}
    else
    {await t
        .typeText(Selector('.input-timezone input'), FirstPhotoTimezone, { replace: true })
        .pressKey('enter')}
    if (FirstPhotoLatitude.empty || FirstPhotoLatitude === "")
    { await t
        .click(Selector('.input-latitude input'))
        .pressKey('ctrl+a delete')}
    else
    {await t.typeText(Selector('.input-latitude input'), FirstPhotoLatitude, { replace: true })}
    if (FirstPhotoLongitude.empty || FirstPhotoLongitude === "")
    { await t
        .click(Selector('.input-longitude input'))
        .pressKey('ctrl+a delete')}
    else
    {await t.typeText(Selector('.input-longitude input'), FirstPhotoLongitude, { replace: true })}
    if (FirstPhotoAltitude.empty || FirstPhotoAltitude === "")
    { await t
        .click(Selector('.input-altitude input'))
        .pressKey('ctrl+a delete')}
    else
    {await t.typeText(Selector('.input-altitude input'), FirstPhotoAltitude, { replace: true })}
    if (FirstPhotoCountry.empty  || FirstPhotoCountry === "")
    { await t
        .click(Selector('.input-longitude input'))
        .pressKey('ctrl+a delete')}
    else
    {await t
        .click(Selector('.input-country input'))
        .pressKey('ctrl+a delete')
        .typeText(Selector('.input-country input'), FirstPhotoCountry, { replace: true })
        .pressKey('enter')}
    // if (FirstPhotoCamera.empty || FirstPhotoCamera === "")
    //{ await t
        //.click(Selector('.input-camera input'))
       // .hover(Selector('div').withText('Unknown').parent('div[role="listitem"]'))
      //  .click(Selector('div').withText('Unknown').parent('div[role="listitem"]'))}
    //else
    //{await t
      //  .click(Selector('.input-camera input'))
     //   .hover(Selector('div').withText(FirstPhotoCamera).parent('div[role="listitem"]'))
    //    .click(Selector('div').withText(FirstPhotoCamera).parent('div[role="listitem"]'))}
    //if (FirstPhotoLens.empty || FirstPhotoLens === "")
    //{ await t
      //  .click(Selector('.input-lens input'))
     //   .click(Selector('div').withText('Unknown').parent('div[role="listitem"]'))}
    //else
    //{await t
     //   .click(Selector('.input-lens input'))
    //    .click(Selector('div').withText(FirstPhotoLens).parent('div[role="listitem"]'))}
    if (FirstPhotoIso.empty || FirstPhotoIso === "")
    { await t
        .click(Selector('.input-iso input'))
        .pressKey('ctrl+a delete')}
    else
    {await t.typeText(Selector('.input-iso input'), FirstPhotoIso, { replace: true })}
    if (FirstPhotoExposure.empty || FirstPhotoExposure === "")
    { await t
        .click(Selector('.input-exposure input'))
        .pressKey('ctrl+a delete')}
    else
    {await t.typeText(Selector('.input-exposure input'), FirstPhotoExposure, { replace: true })}
    if (FirstPhotoFnumber.empty || FirstPhotoFnumber === "")
    { await t
        .click(Selector('.input-fnumber input'))
        .pressKey('ctrl+a delete')}
    else
    {await t.typeText(Selector('.input-fnumber input'), FirstPhotoFnumber, { replace: true })}
    if (FirstPhotoFocalLength.empty || FirstPhotoFocalLength === "")
    { await t
        .click(Selector('.input-focal-length input'))
        .pressKey('ctrl+a delete')}
    else
    {await t.typeText(Selector('.input-focal-length input'), FirstPhotoFocalLength, { replace: true })}
    if (FirstPhotoSubject.empty || FirstPhotoSubject === "")
    { await t
        .click(Selector('.input-subject textarea'))
        .pressKey('ctrl+a delete')}
    else
    {await t.typeText(Selector('.input-subject textarea'), FirstPhotoSubject, { replace: true })}
    if (FirstPhotoArtist.empty || FirstPhotoSubject === "")
    { await t
        .click(Selector('.input-artist input'))
        .pressKey('ctrl+a delete')}
    else
    {await t.typeText(Selector('.input-artist input'), FirstPhotoArtist, { replace: true })}
    if (FirstPhotoCopyright.empty || FirstPhotoCopyright === "")
    { await t
        .click(Selector('.input-copyright input'))
        .pressKey('ctrl+a delete')}
    else
    {await t.typeText(Selector('.input-copyright input'), FirstPhotoCopyright, { replace: true })}
    if (FirstPhotoLicense.empty || FirstPhotoLicense === "")
    {await t
        .click(Selector('.input-license textarea'))
        .pressKey('ctrl+a delete')}
    else
    {await t.typeText(Selector('.input-license textarea'), FirstPhotoLicense, { replace: true })}
    if (FirstPhotoDescription.empty || FirstPhotoDescription === "")
    { await t
        .click(Selector('.input-description textarea'))
        .pressKey('ctrl+a delete')}
    else
    {await t.typeText(Selector('.input-description textarea'), FirstPhotoDescription, { replace: true })}
    if (FirstPhotoKeywords.empty || FirstPhotoKeywords === "")
    { await t
        .click(Selector('.input-keywords textarea'))
        .pressKey('ctrl+a delete')}
    else
    {await t.typeText(Selector('.input-keywords textarea'), FirstPhotoKeywords, { replace: true })}
    if (FirstPhotoNotes.empty || FirstPhotoNotes === "")
    { await t
        .click(Selector('.input-notes textarea'))
        .pressKey('ctrl+a delete')}
    else
    {await t.typeText(Selector('.input-notes textarea'), FirstPhotoNotes, { replace: true })}
    await t
        .click(Selector('button.action-done'));
    const clipboardCount = await Selector('span.count-clipboard');
    await t
        .expect(clipboardCount.textContent).eql("1")
        .click(Selector('.action-clear'))
        .expect(Selector('button.action-menu').exists, {timeout: 5000}).notOk();
});

test('#8 Change primary file', async t => {
    await page.openNav();
    await t
        .click(Selector('.nav-photos'))
        .click(Selector('.p-expand-search'));
    await page.search('ski');
    const SequentialPhoto = await Selector('div.p-photo').nth(0).getAttribute('data-uid');

    await t
        .expect(Selector('i.action-burst').visible).ok()
        .click(Selector('i.action-burst'))
        .click(Selector('.action-next'))
        .click(Selector('.action-previous'))
        .click(Selector('.action-close'));
    await t
        .click(Selector('button.action-title-edit').withAttribute('data-uid', SequentialPhoto))
        .click(Selector('#tab-files'));
    const FirstFile = await Selector('div.caption').nth(0).innerText;
    await t
        .expect(FirstFile).contains('photos8_1_ski.jpg')
        .click(Selector('li.v-expansion-panel__container').nth(1))
        .click(Selector('.action-primary'))
        .click(Selector('button.action-close'))
        .click(Selector('button.action-title-edit').withAttribute('data-uid', SequentialPhoto));
    const FirstFileAfterChange = await Selector('div.caption').nth(0).innerText;
    await t
        .expect(FirstFileAfterChange).notContains('photos8_1_ski.jpg')
        .expect(FirstFileAfterChange).contains('photos8_2_ski.jpg');
});

test('#9 Navigate from card view to place', async t => {
    await t.click(Selector('.p-expand-search'));
    await page.setFilter('view', 'Cards');
    await t
        .click(Selector('button.action-location').nth(0))
        .expect(Selector('#map').exists, {timeout: 15000}).ok()
        .expect(Selector('div.p-map-control').visible).ok()
        .expect(Selector('.input-search input').value).notEql('');
});
//TODO remove wait
test('#10 Ungroup files', async t => {
    await page.openNav();
    await t
        .click(Selector('.nav-photos'))
        .click(Selector('.p-expand-search'));
    await page.search('group');
    const PhotoCount = await Selector('button.action-title-edit').count;

    const SequentialPhoto = await Selector('div.p-photo').nth(0).getAttribute('data-uid');
    await t
        .expect(PhotoCount).eql(1)
        .click(Selector('div.nav-photos + div'))
        .click(Selector('.nav-stacks'))
        .expect(Selector('i.action-burst').visible).ok()
        .click(Selector('button.action-title-edit').withAttribute('data-uid', SequentialPhoto))
        .click(Selector('#tab-files'))
        .click(Selector('li.v-expansion-panel__container').nth(1))
        .click(Selector('.action-unstack'))
        .wait(12000)
        .click(Selector('button.action-close'))
        .click(Selector('.nav-photos'))
        .click(Selector('.p-expand-search'));
    await page.search('group');
    await t
        .click(Selector('.action-reload'));
    const PhotoCountAfterUngroup = await Selector('button.action-title-edit').count;
    await t
        .expect(PhotoCountAfterUngroup).eql(2);
});
