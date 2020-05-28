import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from "./page-model";
import { RequestLogger } from 'testcafe';


const logger = RequestLogger( /http:\/\/localhost:2342\/api\/v1\/photos*/ , {
    logResponseHeaders: true,
    logResponseBody:    true
});

fixture `Test photos`
    .page`${testcafeconfig.url}`
    .requestHooks(logger);

const page = new Page();

//?search + filters --> oder search js
// raw file icon? live photo icon? - search for type look for icon --> do it in search test

//?views including fullscreen--> oder view js

//?upload + Later: delete--> oder nur in library?

//TODO scroll to top

//TODO download

//TODO clipboard --> extra test needed or include in other tests done for private, archive

//TODO add part for video as well
test('#8 approve photo', async t => {
    await page.openNav();
    await t
        .click(Selector('div.p-navigation-photos + div'))
        .click(Selector('.p-navigation-review'));
    logger.clear();
    await page.search('type:image');
    const request1 = await logger.requests[0].response.body;
    const FirstPhoto = await Selector('div.p-photo').nth(0).getAttribute('data-uid');
    const SecondPhoto = await Selector('div.p-photo').nth(1).getAttribute('data-uid');

    logger.clear();
    await t
        .click(Selector('.p-navigation-photos'));
    const request11 = await logger.requests[0].response.body;
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).exists).notOk()
        .expect(Selector('div').withAttribute('data-uid', SecondPhoto).exists).notOk();
    logger.clear();
    await t.click(Selector('.p-navigation-review'));
    const request111 = await logger.requests[0].response.body;

    await page.selectPhotoFromUID(FirstPhoto);
    await page.editSelectedPhotos();
    logger.clear();
    await t
        .click(Selector('button.p-photo-dialog-close'));
    await t
        .click(Selector('button.action-reload'));
    const request12 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).visible).ok();
    await page.editSelectedPhotos();
    const request2 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .click(Selector('button.action-approve'))
        .click(Selector('button.action-ok'));
    const request3 = await logger.requests[0].response.body;
    logger.clear();

    await page.unselectPhotoFromUID(FirstPhoto);
    await page.selectPhotoFromUID(SecondPhoto);
    await page.editSelectedPhotos();
    logger.clear();
    await t
        .typeText(Selector('input[aria-label="Latitude"]'), '9.999')
        .typeText(Selector('input[aria-label="Longitude"]'), '9.999')
        .click(Selector('button.action-ok'));
    const request31 = await logger.requests[0].response.body;

    await t
        .click(Selector('button.action-reload'));
    const request4 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).exists).notOk()
        .expect(Selector('div').withAttribute('data-uid', SecondPhoto).exists).notOk()
        .click(Selector('.p-navigation-photos'));
    const request5 = await logger.requests[0].response.body;
    logger.clear();
    await page.search('type:image');
    const request6 = await logger.requests[0].response.body;
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).visible).ok()
        .expect(Selector('div').withAttribute('data-uid', SecondPhoto).visible).ok();
});

//TODO videos - play video

test('#1 like/dislike photo/video', async t => {

    logger.clear();
    const FirstPhoto = await Selector('.t-off').nth(0).getAttribute('data-uid');

    await t.click(Selector('.p-navigation-video'));
    const request0 = await logger.requests[0].response.body;
    const FirstVideo = await Selector('.t-off').nth(0).getAttribute('data-uid');

    await t.click(Selector('.p-navigation-favorites'));
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).exists).notOk()
        .expect(Selector('div').withAttribute('data-uid', FirstVideo).exists).notOk()
        .click(Selector('.p-navigation-photos'));

    logger.clear();
    await page.likePhoto(FirstPhoto);
    const request = await logger.requests[0].response.body;
    logger.clear();
    await t
        .click(Selector('.action-reload'))
        .expect(Selector('i.t-on').withAttribute('data-uid', FirstPhoto).exists).ok();
    logger.clear();

    await t.click(Selector('.p-navigation-video'));
    await page.likePhoto(FirstVideo);
    const request1 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .click(Selector('.action-reload'))
        .expect(Selector('i.t-on').withAttribute('data-uid', FirstVideo).exists).ok();
    logger.clear();

    await t
        .click(Selector('.p-navigation-favorites'));
    const request2 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).exists).ok()
        .expect(Selector('div').withAttribute('data-uid', FirstVideo).exists).ok()
        .expect(Selector('div.v-image__image').visible).ok();
    await page.dislikePhoto(FirstVideo);
    const request21 = await logger.requests[0].response.body;
    logger.clear();
    await page.dislikePhoto(FirstPhoto);
    logger.clear();
    await t.click(Selector('.action-reload'));
    const request3 = await logger.requests[0].response.body;
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).exists).notOk()
        .expect(Selector('div').withAttribute('data-uid', FirstVideo).exists).notOk();
});

test('#6 private/unprivate photo/video using clipboard and list', async t => {

    logger.clear();
    await t
        .click(Selector('.p-navigation-photos'))
        .click(Selector('.p-expand-search'));
    logger.clear();
    await page.setFilter('view', 'Mosaik');
    const request0 = await logger.requests[0].response.body;
    const FirstPhoto = await Selector('div.p-photo').nth(0).getAttribute('data-uid');
    const SecondPhoto = await Selector('div.p-photo').nth(1).getAttribute('data-uid');
    const ThirdPhoto = await Selector('div.p-photo').nth(2).getAttribute('data-uid');

    await t
        .click(Selector('.p-navigation-video'));
    const request = await logger.requests[0].response.body;
    const FirstVideo = await Selector('div.p-photo').nth(0).getAttribute('data-uid');
    const SecondVideo = await Selector('div.p-photo').nth(1).getAttribute('data-uid');

    await t.click(Selector('.p-navigation-private'));
    const request1 = await logger.requests[0].response.body;
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).exists).notOk()
        .expect(Selector('div').withAttribute('data-uid', SecondPhoto).exists).notOk()
        .expect(Selector('div').withAttribute('data-uid', ThirdPhoto).exists).notOk()
        .expect(Selector('div').withAttribute('data-uid', FirstVideo).exists).notOk()
        .expect(Selector('div').withAttribute('data-uid', SecondVideo).exists).notOk()
        .click(Selector('.p-navigation-photos'));

    logger.clear();
    await page.selectPhotoFromUID(FirstPhoto);
    await page.selectPhotoFromUID(SecondPhoto);
    await t
        .click(Selector('button.p-photo-clipboard-menu'))
        .click(Selector('button.p-photo-clipboard-private'));
    logger.clear();
    await page.setFilter('view', 'Liste');
    const request12 = await logger.requests[0].response.body;
    await t
        .click(Selector('button.p-photo-private').withAttribute('data-uid', ThirdPhoto));
    logger.clear();

    await t
        .click(Selector('.action-reload'));
    const request2 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('td').withAttribute('data-uid', FirstPhoto).exists).notOk()
        .expect(Selector('td').withAttribute('data-uid', SecondPhoto).exists).notOk()
        .expect(Selector('td').withAttribute('data-uid', ThirdPhoto).exists).notOk()
        .click(Selector('.p-navigation-video'));

    logger.clear();
    await t
        .click(Selector('button.p-photo-private').withAttribute('data-uid', SecondVideo));
    await page.setFilter('view', 'Mosaik');
    const request13 = await logger.requests[0].response.body;
    logger.clear();

    await page.selectPhotoFromUID(FirstVideo);
    await t
        .click(Selector('button.p-photo-clipboard-menu'))
        .click(Selector('button.p-photo-clipboard-private'));
    logger.clear();
    await t
        .click(Selector('.action-reload'));
    const request3 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstVideo).exists).notOk()
        .expect(Selector('div').withAttribute('data-uid', SecondVideo).exists).notOk()
        .click(Selector('.p-navigation-private'));

    const request4 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).exists).ok()
        .expect(Selector('div').withAttribute('data-uid', SecondPhoto).exists).ok()
        .expect(Selector('div').withAttribute('data-uid', FirstVideo).exists).ok()
        .expect(Selector('div').withAttribute('data-uid', SecondVideo).exists).ok();
    await page.selectPhotoFromUID(FirstPhoto);
    await page.selectPhotoFromUID(SecondPhoto);
    await page.selectPhotoFromUID(FirstVideo);
    await t
        .click(Selector('button.p-photo-clipboard-menu'))
        .click(Selector('button.p-photo-clipboard-private'));
    await page.setFilter('view', 'Liste');
    await t
        .click(Selector('button.p-photo-private').withAttribute('data-uid', ThirdPhoto))
        .click(Selector('button.p-photo-private').withAttribute('data-uid', SecondVideo));
    await page.setFilter('view', 'Mosaik');
    logger.clear();
    await t.click(Selector('.action-reload'));
    const request5 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).exists).notOk()
        .expect(Selector('div').withAttribute('data-uid', SecondPhoto).exists).notOk()
        .expect(Selector('div').withAttribute('data-uid', ThirdPhoto).exists).notOk()
        .expect(Selector('div').withAttribute('data-uid', FirstVideo).exists).notOk()
        .expect(Selector('div').withAttribute('data-uid', SecondVideo).exists).notOk()
        .click(Selector('.p-navigation-photos'));

    const request6 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).exists).ok()
        .expect(Selector('div').withAttribute('data-uid', SecondPhoto).exists).ok()
        .expect(Selector('div').withAttribute('data-uid', ThirdPhoto).exists).ok()
        .click(Selector('.p-navigation-video'));

    const request7 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstVideo).exists).ok()
        .expect(Selector('div').withAttribute('data-uid', SecondVideo).exists).ok();
});

test('#7 archive/restore video, photos, private photos and review photos using clipboard', async t => {

    logger.clear();
    await page.openNav();
    await t
        .click(Selector('.p-navigation-photos'))
        .click(Selector('.p-expand-search'));
    logger.clear();
    await page.setFilter('view', 'Mosaik');
    const request0 = await logger.requests[0].response.body;
    const FirstPhoto = await Selector('div.p-photo').nth(0).getAttribute('data-uid');
    const SecondPhoto = await Selector('div.p-photo').nth(1).getAttribute('data-uid');

    await t
        .click(Selector('.p-navigation-video'));
    const request = await logger.requests[0].response.body;
    const FirstVideo = await Selector('div.p-photo').nth(0).getAttribute('data-uid');
    logger.clear();

    await t
        .click(Selector('.p-navigation-private'));
    const request1 = await logger.requests[0].response.body;
    const FirstPrivatePhoto = await Selector('div.p-photo').nth(0).getAttribute('data-uid');
    logger.clear();

    await t
        .click(Selector('div.p-navigation-photos + div'))
        .click(Selector('.p-navigation-review'));
    const request2 = await logger.requests[0].response.body;
    const FirstReviewPhoto = await Selector('div.p-photo').nth(0).getAttribute('data-uid');
    logger.clear();

    await t
        .click(Selector('.p-navigation-archive'));
    const request3 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).exists).notOk()
        .expect(Selector('div').withAttribute('data-uid', SecondPhoto).exists).notOk()
        .expect(Selector('div').withAttribute('data-uid', FirstVideo).exists).notOk()
        .expect(Selector('div').withAttribute('data-uid', FirstPrivatePhoto).exists).notOk()
        .expect(Selector('div').withAttribute('data-uid', FirstReviewPhoto).exists).notOk();

    await t
        .click(Selector('.p-navigation-video'));
    const request4 = await logger.requests[0].response.body;
    logger.clear();
    await page.selectPhotoFromUID(FirstVideo);
    await page.archiveSelectedPhotos();
    logger.clear();
    await t
        .click(Selector('.action-reload'));
    const request5 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstVideo).exists).notOk()
        .click(Selector('.p-navigation-photos'));

    const request6 = await logger.requests[0].response.body;
    logger.clear();
    await page.selectPhotoFromUID(FirstPhoto);
    await page.selectPhotoFromUID(SecondPhoto);
    await page.archiveSelectedPhotos();
    logger.clear();
    await t
        .click(Selector('.action-reload'));
    const request7 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).exists).notOk()
        .expect(Selector('div').withAttribute('data-uid', SecondPhoto).exists).notOk()
        .click(Selector('.p-navigation-private'));

    const request8 = await logger.requests[0].response.body;
    logger.clear();
    await page.selectPhotoFromUID(FirstPrivatePhoto);
    await page.archiveSelectedPhotos();
    logger.clear();
    await t
        .click(Selector('.action-reload'));
    const request9 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPrivatePhoto).exists).notOk()
        .click(Selector('.p-navigation-review'));

    const request10 = await logger.requests[0].response.body;
    logger.clear();
    await page.selectPhotoFromUID(FirstReviewPhoto);
    await page.archiveSelectedPhotos();
    logger.clear();
    await t
        .click(Selector('.action-reload'));
    const request11 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstReviewPhoto).exists).notOk()
        .click(Selector('.p-navigation-archive'));
    const request12 = await logger.requests[0].response.body;
    logger.clear();

    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).exists).ok()
        .expect(Selector('div').withAttribute('data-uid', SecondPhoto).exists).ok()
        .expect(Selector('div').withAttribute('data-uid', FirstVideo).exists).ok()
        .expect(Selector('div').withAttribute('data-uid', FirstPrivatePhoto).exists).ok()
        .expect(Selector('div').withAttribute('data-uid', FirstReviewPhoto).exists).ok();
    await page.selectPhotoFromUID(FirstPhoto);
    await page.selectPhotoFromUID(SecondPhoto);
    await page.selectPhotoFromUID(FirstVideo);
    await page.selectPhotoFromUID(FirstPrivatePhoto);
    await page.selectPhotoFromUID(FirstReviewPhoto);
    await page.restoreSelectedPhotos();
    logger.clear();
    await t.click(Selector('.action-reload'));
    const request13 = await logger.requests[0].response.body;
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).exists).notOk()
        .expect(Selector('div').withAttribute('data-uid', SecondPhoto).exists).notOk()
        .expect(Selector('div').withAttribute('data-uid', FirstVideo).exists).notOk()
        .expect(Selector('div').withAttribute('data-uid', FirstPrivatePhoto).exists).notOk()
        .expect(Selector('div').withAttribute('data-uid', FirstReviewPhoto).exists).notOk();
    logger.clear();

    await t
        .click(Selector('.p-navigation-video'));
    const request14 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstVideo).exists).ok();

    await t
        .click(Selector('.p-navigation-photos'));
    const request15 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).exists).ok()
        .expect(Selector('div').withAttribute('data-uid', SecondPhoto).exists).ok();

    await t
        .click(Selector('.p-navigation-private'));
    const request16 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPrivatePhoto).exists).ok();

    await t
        .click(Selector('.p-navigation-review'));
    const request17 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstReviewPhoto).exists).ok();
});


//TODO open photoeditdialogue (multiple ways) + edit photo details

//TODO change primary file

//TODO navigate to places

//??Check count in navi gets updated --> gt/lt or matches count of images