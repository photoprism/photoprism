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

    const FavoritesCount = await Selector('button.p-photo-like').count;
    logger.clear();
    await t
        .navigateTo("../photos")
    const request = await logger.requests[0];
    await t
        .hover(Selector('div[class="v-image__image v-image__image--cover"]').nth(5))
        .click(Selector('button.p-photo-like'))
    logger.clear();
    await t
        .navigateTo("../favorites");
    const request2 = await logger.requests[1];
    logger.clear();

    const FavoritesCountAfterLike = await Selector('button.p-photo-like').count;
    await t
        .expect(FavoritesCountAfterLike).eql(FavoritesCount + 1)
        .expect(Selector('div.v-image__image').visible).ok();
}),

test('Dislike photo', async t => {

    const FavoritesCount = await Selector('button.p-photo-like').count;
    await t
        .hover(Selector('div[class="v-image__image v-image__image--cover"]').nth(0))
        .click(Selector('button.p-photo-like'))
    logger.clear();
    await t
        .navigateTo("../favorites");
    const request3 = await logger.requests[0];

    const FavoritesCountAfterDislike = await Selector('button.p-photo-like').count;
    await t
        .expect(FavoritesCountAfterDislike).eql(FavoritesCount - 1);
});
