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

//scroll to top

//download

//clipboard --> extra test needed or include in other tests done for privat, archive

//review - approve photo
test('#8 approve photo', async t => {
    await page.openNav();
    await t
        .click(Selector('div.p-navigation-photos + div'))
        .click(Selector('.p-navigation-review'));
    logger.clear();
    await page.search('type:image');
    const request1 = await logger.requests[0].response.body;
    const FirstPhoto = await Selector('div.p-photo').nth(0).getAttribute('data-uid');

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
    await t
        .click(Selector('button.action-reload'));
    const request4 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).exists).notOk()
        .click(Selector('.p-navigation-photos'));
    const request5 = await logger.requests[0].response.body;
    logger.clear();
    await page.search('type:image');
    const request6 = await logger.requests[0].response.body;
    await t
        .expect(Selector('div').withAttribute('data-uid', FirstPhoto).visible).ok();
});

//review - add/remove date to photo so that it appears in photos

//review - for videos

//videos - play video

/*test('#1 like/dislike photo', async t => {

    await t.click(Selector('.p-navigation-favorites'));
    const PhotosCountInFavorites = await Selector('.t-like.t-on').count;
    await t.click(Selector('.p-navigation-photos'));
    const FavoritesCountInPhotos = await Selector('.t-like.t-on').count;
    logger.clear();
    await page.search('favorite:false');
    await page.likePhoto(1);
    const request = await logger.requests[0].response.body;
    logger.clear();
    await t
        .click(Selector('.action-reload'));
    const FavoritesCountInPhotosAfterLike = await Selector('.t-like.t-on').count;
    logger.clear();
    await t
        .expect(FavoritesCountInPhotosAfterLike).eql(FavoritesCountInPhotos + 1)
        .click(Selector('.p-navigation-favorites'));
    const request2 = await logger.requests[0].response.body;
    logger.clear();

    const PhotosCountInFavoritesAfterLike = await Selector('.t-like.t-on').count;
    await t
        .expect(PhotosCountInFavoritesAfterLike).eql(PhotosCountInFavorites + 1)
        .expect(Selector('div.v-image__image').visible).ok()
        .click(Selector('.t-like.t-on'));
    logger.clear();
    await t.click(Selector('.action-reload'));
    const request3 = await logger.requests[0].response.body;

    const PhotosCountInFavoritesAfterDislike = await Selector('.t-like.t-on').count;
    await t
        .expect(PhotosCountInFavoritesAfterDislike).eql(PhotosCountInFavoritesAfterLike - 1);

});

test('#2 like/dislike video', async t => {

    await t.click(Selector('.p-navigation-favorites'));
    const VideoCountInFavorites = await Selector('.t-like.t-on').count;
    await t.click(Selector('.p-navigation-video'));
    const FavoritesCountInVideos = await Selector('.t-like.t-on').count;
    logger.clear();
    await page.search('favorite:false');
    await page.likePhoto(0);
    const request = await logger.requests[0].response.body;
    logger.clear();
    await t
        .click(Selector('.action-reload'));
    const FavoritesCountInVideosAfterLike = await Selector('.t-like.t-on').count;
    logger.clear();
    await t
        .expect(FavoritesCountInVideosAfterLike).eql(FavoritesCountInVideos + 1)
        .click(Selector('.p-navigation-favorites'));
    const request2 = await logger.requests[0].response.body;
    logger.clear();

    const VideoCountInFavoritesAfterLike = await Selector('.t-like.t-on').count;
    await t
        .expect(VideoCountInFavoritesAfterLike).eql(VideoCountInFavorites + 1);
    await page.search('type:video');
    await t
        .expect(Selector('div.v-image__image').visible).ok()
        .click(Selector('.t-like.t-on'));
    logger.clear();
    await t.click(Selector('.action-reload'));
    const request3 = await logger.requests[0].response.body;

    const VideoCountInFavoritesAfterDislike = await Selector('.t-like.t-on').count;
    await t
        .expect(VideoCountInFavoritesAfterDislike).eql(VideoCountInFavoritesAfterLike - 1);

});

test('#5 private/unprivate photo from list view', async t => {

    logger.clear();
    await t.click(Selector('.p-navigation-private'));
    const CountInPrivate = await Selector('.p-photo-private').count;
    await t
        .click(Selector('.p-navigation-photos'))
        .click(Selector('.p-expand-search'));
    logger.clear();
    await page.setFilter('view', 'Liste');
    const request01 = await logger.requests[0].response.body;
    const PhotoCountInPhotos = await Selector('div.v-responsive__content').count;
    logger.clear();
    await t.click(Selector('.p-photo-private').nth(0));
    const request1 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .click(Selector('.action-reload'));
    const request12 = await logger.requests[0].response.body;
    const PhotoCountInPhotosAfterPrivate = await Selector('div.v-responsive__content').count;
    logger.clear();
    await t
        .expect(PhotoCountInPhotosAfterPrivate).eql(PhotoCountInPhotos - 1)
        .click(Selector('.p-navigation-private'));
    const request2 = await logger.requests[0].response.body;
    logger.clear();

    const CountInPrivateAfterPrivate = await Selector('.p-photo-private').count;
    await t
        .expect(CountInPrivateAfterPrivate).eql(CountInPrivate + 1);
    await t
        .expect(Selector('div.v-image__image').visible).ok()
        .click(Selector('.p-photo-private').nth(0));
    logger.clear();
    await t.click(Selector('.action-reload'));
    const request3 = await logger.requests[0].response.body;

    const CountInPrivateAfterUnprivate = await Selector('.p-photo-private').count;
    logger.clear();
    await t
        .expect(CountInPrivateAfterUnprivate).eql(CountInPrivateAfterPrivate - 1)
        .click(Selector('.p-navigation-photos'));
    const request4 = await logger.requests[0].response.body;
    logger.clear();
    const PhotoCountInPhotosAfterUnprivate = await Selector('div.v-responsive__content').count;
    await t
        .expect(PhotoCountInPhotosAfterUnprivate).eql(PhotoCountInPhotosAfterPrivate + 1);
});

test('#6 private/unprivate photo using clipboard', async t => {

    logger.clear();
    await t.click(Selector('.p-navigation-private'));
    const request = await logger.requests[0].response.body;
    const CountInPrivate = await Selector('.p-photo-private').count;
    await t
        .click(Selector('.p-navigation-photos'))
        .click(Selector('.p-expand-search'));
    logger.clear();
    await page.setFilter('view', 'Mosaik');
    const request01 = await logger.requests[0].response.body;
    const PhotoCountInPhotos = await Selector('div.v-responsive__content').count;
    logger.clear();
    await page.selectNthPhoto(0);
    await page.selectNthPhoto(1);
    await t
        .click(Selector('button.p-photo-clipboard-menu'))
        .click(Selector('button.p-photo-clipboard-private'));
    logger.clear();
    await t
        .click(Selector('.action-reload'));
    const request12 = await logger.requests[0].response.body;
    const PhotoCountInPhotosAfterPrivate = await Selector('div.v-responsive__content').count;
    logger.clear();
    await t
        .expect(PhotoCountInPhotosAfterPrivate).eql(PhotoCountInPhotos - 2)
        .click(Selector('.p-navigation-private'));
    const request2 = await logger.requests[0].response.body;
    logger.clear();

    const CountInPrivateAfterPrivate = await Selector('.p-photo-private').count;
    await t
        .expect(CountInPrivateAfterPrivate).eql(CountInPrivate + 2);
    await t
        .expect(Selector('div.v-image__image').visible).ok();
    await page.selectNthPhoto(0);
    await page.selectNthPhoto(1);
    await t
        .click(Selector('button.p-photo-clipboard-menu'))
        .click(Selector('button.p-photo-clipboard-private'));
    logger.clear();
    await t.click(Selector('.action-reload'));
    const request3 = await logger.requests[0].response.body;

    const CountInPrivateAfterUnprivate = await Selector('.p-photo-private').count;
    logger.clear();
    await t
        .expect(CountInPrivateAfterUnprivate).eql(CountInPrivateAfterPrivate - 2)
        .click(Selector('.p-navigation-photos'));
    const request4 = await logger.requests[0].response.body;
    logger.clear();
    const PhotoCountInPhotosAfterUnprivate = await Selector('div.v-responsive__content').count;
    await t
        .expect(PhotoCountInPhotosAfterUnprivate).eql(PhotoCountInPhotosAfterPrivate + 2);

});

test('#3 private/unprivate video from list view', async t => {

    logger.clear();
    await t.click(Selector('.p-navigation-private'));
    await page.search('type:video');
    const request = await logger.requests[0].response.body;
    const VideoCountInPrivate = await Selector('.p-photo-private').count;
    await t
        .click(Selector('.p-navigation-video'))
        .click(Selector('.p-expand-search'));
    logger.clear();
    await page.setFilter('view', 'Liste');
    const request01 = await logger.requests[0].response.body;
    const VideoCountInVideo = await Selector('button.p-photo-play').count;
    logger.clear();
    await t.click(Selector('.p-photo-private').nth(0));
    const request1 = await logger.requests[0].response.body;
    logger.clear();
    await t
        .click(Selector('.action-reload'));
    const request12 = await logger.requests[0].response.body;
    const VideoCountInVideoAfterPrivate = await Selector('button.p-photo-play').count;
    logger.clear();
    await t
        .expect(VideoCountInVideoAfterPrivate).eql(VideoCountInVideo - 1)
        .click(Selector('.p-navigation-private'));
    await page.search('type:video');
    const request2 = await logger.requests[0].response.body;
    logger.clear();

    const VideoCountInPrivateAfterPrivate = await Selector('.p-photo-private').count;
    await t
        .expect(VideoCountInPrivateAfterPrivate).eql(VideoCountInPrivate + 1);
    await t
        .expect(Selector('div.v-image__image').visible).ok()
        .click(Selector('.p-photo-private').nth(0));
    logger.clear();
    await t.click(Selector('.action-reload'));
    const request3 = await logger.requests[0].response.body;

    const VideoCountInPrivateAfterUnprivate = await Selector('.p-photo-private').count;
    logger.clear();
    await t
        .expect(VideoCountInPrivateAfterUnprivate).eql(VideoCountInPrivateAfterPrivate - 1)
        .click(Selector('.p-navigation-video'));
    const request4 = await logger.requests[0].response.body;
    logger.clear();
    const VideoCountInVideoAfterUnprivate = await Selector('button.p-photo-play').count;
    await t
        .expect(VideoCountInVideoAfterUnprivate).eql(VideoCountInVideoAfterPrivate + 1);

});

test('#4 private/unprivate video using clipboard', async t => {

    logger.clear();
    await t.click(Selector('.p-navigation-private'));
    await page.search('type:video');
    const request = await logger.requests[0].response.body;
    const VideoCountInPrivate = await Selector('.p-photo-private').count;
    await t
        .click(Selector('.p-navigation-video'))
        .click(Selector('.p-expand-search'));
    logger.clear();
    await page.setFilter('view', 'Mosaik');
    const request01 = await logger.requests[0].response.body;
    const VideoCountInVideo = await Selector('button.p-photo-play').count;
    logger.clear();
    await page.selectNthPhoto(0);
    await page.selectNthPhoto(1);
    await t
        .click(Selector('button.p-photo-clipboard-menu'))
        .click(Selector('button.p-photo-clipboard-private'));
    logger.clear();
    await t
        .click(Selector('.action-reload'));
    const request12 = await logger.requests[0].response.body;
    const VideoCountInVideoAfterPrivate = await Selector('button.p-photo-play').count;
    logger.clear();
    await t
        .expect(VideoCountInVideoAfterPrivate).eql(VideoCountInVideo - 2)
        .click(Selector('.p-navigation-private'));
    await page.search('type:video');
    const request2 = await logger.requests[0].response.body;
    logger.clear();

    const VideoCountInPrivateAfterPrivate = await Selector('.p-photo-private').count;
    await t
        .expect(VideoCountInPrivateAfterPrivate).eql(VideoCountInPrivate + 2);
    await t
        .expect(Selector('div.v-image__image').visible).ok();
    await page.selectNthPhoto(0);
    await page.selectNthPhoto(1);
    await t
        .click(Selector('button.p-photo-clipboard-menu'))
        .click(Selector('button.p-photo-clipboard-private'));
    logger.clear();
    await t.click(Selector('.action-reload'));
    const request3 = await logger.requests[0].response.body;

    const VideoCountInPrivateAfterUnprivate = await Selector('.p-photo-private').count;
    logger.clear();
    await t
        .expect(VideoCountInPrivateAfterUnprivate).eql(VideoCountInPrivateAfterPrivate - 2)
        .click(Selector('.p-navigation-video'));
    const request4 = await logger.requests[0].response.body;
    logger.clear();
    const VideoCountInVideoAfterUnprivate = await Selector('button.p-photo-play').count;
    await t
        .expect(VideoCountInVideoAfterUnprivate).eql(VideoCountInVideoAfterPrivate + 2);

});

//archive -  archive + restore from photos from private from review (UPDATE)

//photos

//private photos

//review photos
test('#7 archive/restore video, photos, private photos and review photos using clipboard', async t => {

    logger.clear();
    await page.openNav();
    await t
        .click(Selector('div.p-navigation-photos + div'))
        .click(Selector('.p-navigation-archive'));
    const request = await logger.requests[0].response.body;
    const TotalCountInArchive = await Selector('div.v-responsive__content').count;
    logger.clear();

    await t
        .click(Selector('.p-navigation-video'));
    const request1 = await logger.requests[0].response.body;
    const VideoCountInVideo = await Selector('button.p-photo-play').count;
    logger.clear();
    await page.selectNthPhoto(0);
    await page.archiveSelectedPhotos();
    logger.clear();
    await t
        .click(Selector('.action-reload'));
    const request2 = await logger.requests[0].response.body;
    const VideoCountInVideoAfterArchive = await Selector('button.p-photo-play').count;
    logger.clear();
    await t
        .expect(VideoCountInVideoAfterArchive).eql(VideoCountInVideo - 1)
        .click(Selector('.p-navigation-photos'));
    const request3 = await logger.requests[0].response.body;
    logger.clear();

    const PhotoCountInPhotos = await Selector('div.v-responsive__content').count;
    logger.clear();
    await page.selectNthPhoto(0);
    await page.selectNthPhoto(1);
    await page.archiveSelectedPhotos();
    logger.clear();
    await t
        .click(Selector('.action-reload'));
    const request4 = await logger.requests[0].response.body;
    const PhotoCountInPhotosAfterArchive = await Selector('div.v-responsive__content').count;
    logger.clear();
    await t
        .expect(PhotoCountInPhotosAfterArchive).eql(PhotoCountInPhotos - 2)
        .click(Selector('.p-navigation-private'));
    const request5 = await logger.requests[0].response.body;
    logger.clear();

    const PhotoCountInPrivate = await Selector('div.v-responsive__content').count;
    logger.clear();
    await page.selectNthPhoto(0);
    await page.archiveSelectedPhotos();
    logger.clear();
    await t
        .click(Selector('.action-reload'));
    const request6 = await logger.requests[0].response.body;
    const PhotoCountInPrivateAfterArchive = await Selector('div.v-responsive__content').count;
    logger.clear();
    await t
        .expect(PhotoCountInPrivateAfterArchive).eql(PhotoCountInPrivate - 1)
        .click(Selector('.p-navigation-review'));
    const request7 = await logger.requests[0].response.body;
    logger.clear();

    const PhotoCountInReview = await Selector('div.v-responsive__content').count;
    logger.clear();
    await page.selectNthPhoto(0);
    await page.archiveSelectedPhotos();
    logger.clear();
    await t
        .click(Selector('.action-reload'));
    const request8 = await logger.requests[0].response.body;
    const PhotoCountInReviewAfterArchive = await Selector('div.v-responsive__content').count;
    logger.clear();
    await t
        .expect(PhotoCountInReviewAfterArchive).eql(PhotoCountInReview - 1)
        .click(Selector('.p-navigation-archive'));
    const request9 = await logger.requests[0].response.body;
    logger.clear();

    const TotalCountInArchiveAfterArchive = await Selector('div.v-responsive__content').count;
    await t
        .expect(TotalCountInArchiveAfterArchive).eql(TotalCountInArchive + 5);
    await t
        .expect(Selector('div.v-image__image').visible).ok();
    await page.search('type:video');
    await page.selectNthPhoto(0);
    await page.search('review:true');
    await page.selectNthPhoto(0);
    await page.search('review:true');
    await page.selectNthPhoto(0);
    await page.restoreSelectedPhotos();
    logger.clear();
    await t.click(Selector('.action-reload'));
    const request10 = await logger.requests[0].response.body;

    const TotalCountInArchiveAfterRestore = await Selector('div.v-responsive__content').count;
    logger.clear();
    await t
        .expect(TotalCountInArchiveAfterRestore).eql(TotalCountInArchiveAfterArchive - 5)
        .click(Selector('.p-navigation-video'));
    const request11 = await logger.requests[0].response.body;
    logger.clear();
    const VideoCountInVideoAfterRestore = await Selector('button.p-photo-play').count;
    await t
        .expect(VideoCountInVideoAfterRestore).eql(VideoCountInVideoAfterArchive + 1);

});*/


//open photoeditdialogue (multiple ways) + edit photo details

//change primary file

//navigate to places

//Check count in navi gets updated --> gt/lt or matches count of images