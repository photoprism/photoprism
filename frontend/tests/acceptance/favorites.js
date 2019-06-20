import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from "./page-model";

fixture`Test favorites page`
    .page `localhost:2342/favorites`;

const page = new Page();

test('Like photo', async t => {

    const FavoritesCount = await Selector('button.p-photo-like').count;
    await t
        .navigateTo("../photos")
        .hover(Selector('div[class="v-image__image v-image__image--cover"]').nth(5))
        .click(Selector('button.p-photo-like'))
        .navigateTo("../favorites");

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
            .navigateTo("../favorites");

        const FavoritesCountAfterDislike = await Selector('button.p-photo-like').count;
        await t
            .expect(FavoritesCountAfterDislike).eql(FavoritesCount - 1);
    });
