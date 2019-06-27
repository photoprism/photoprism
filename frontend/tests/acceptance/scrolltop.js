import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from "./page-model";
import { ClientFunction } from 'testcafe';

fixture`Scroll to top`
    .page`${testcafeconfig.url}`;

const page = new Page();
const scroll = ClientFunction((x, y) => window.scrollTo(x, y));
const getcurrentPosition = ClientFunction(() => window.pageYOffset);

test('Test scroll to top functionality', async t => {
        await t
            .expect(Selector('button.p-photo-scroll-top').exists).notOk()
            .expect(getcurrentPosition()).eql(0)
            .expect(Selector('div[class="v-image__image v-image__image--cover"]').nth(0).visible).ok();
        await scroll(0, 1200);
        await t
            .expect(getcurrentPosition()).eql(1200);
        await scroll(0, 900);
        await t
            .expect(getcurrentPosition()).eql(900)
            .click(Selector('button.p-photo-scroll-top'))
            .expect(getcurrentPosition()).eql(0);
});
