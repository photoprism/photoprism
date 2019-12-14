import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from "./page-model";
import { RequestLogger } from 'testcafe';

const logger = RequestLogger( /http:\/\/localhost:2342\/api\/v1\/photos*/ , {
    logResponseHeaders: true,
    logResponseBody:    true
});

fixture`Test favorites page`
    .page `localhost:2342/favorites`
    .requestHooks(logger);

const page = new Page();

test('Like photo', async t => {

    const FavoritesCount = await Selector('.t-like.t-on').count;
    logger.clear();
    await t.navigateTo("../photos")
    const request = await logger.requests[0].responseBody;
    await page.likePhoto(5);
    logger.clear();
    await t.navigateTo("../favorites");
    const request2 = await logger.requests[0].responseBody;
    logger.clear();

    const FavoritesCountAfterLike = await Selector('.t-like.t-on').count;
    await t
        .expect(FavoritesCountAfterLike).eql(FavoritesCount + 1)
        .expect(Selector('div.v-image__image').visible).ok();
}),

test('Dislike photo', async t => {

    const FavoritesCount = await Selector('.t-like.t-on').count;
    await t
        .click(Selector('.t-like.t-on'));
    logger.clear();
    await t.navigateTo("../favorites");
    const request3 = await logger.requests[0].responseBody;

    const FavoritesCountAfterDislike = await Selector('.t-like.t-on').count;
    await t
        .expect(FavoritesCountAfterDislike).eql(FavoritesCount - 1);
});
