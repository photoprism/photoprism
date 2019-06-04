import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from "./page-model";

fixture`Test favorites page`
    .page `localhost:2342/photos`;

const page = new Page();

test('See favorites', async t => {
    await t
        .hover(Selector('div[class="v-image__image v-image__image--cover"]').nth(0))
        .click(Selector('button.p-photo-like'))
        .navigateTo("../favorites")
        .expect(Selector('div.v-image__image').visible).ok();
});
