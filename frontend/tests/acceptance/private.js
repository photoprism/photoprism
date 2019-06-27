import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from "./page-model";
import { RequestLogger } from 'testcafe';

const logger = RequestLogger({ url: /http:\/\/localhost:2342/, method: 'post'}  , {
    logResponseHeaders: true,
    logResponseBody:    true,
    stringifyResponseBody: true
});

fixture`Test batch private`
    .page`${testcafeconfig.url}`
    .requestHooks(logger);

const page = new Page();

test('Make photos private', async t => {
    await t
        .hover(Selector('div[class="v-image__image v-image__image--cover"]').nth(0))
        .click(Selector('button.p-photo-select'))
        .hover(Selector('div[class="v-image__image v-image__image--cover"]').nth(2))
        .click(Selector('button.p-photo-select').nth(1))
        .click(Selector('div.p-photo-clipboard'))
        .click(Selector('.p-photo-clipboard-private'), {timeout: 15000});
    const request = await logger.requests[0].responseBody;
    await t
        .expect(logger.requests[0].response.statusCode).eql(200)
        .expect(logger.requests[0].response.body).contains('photos marked as private');
    const countSelected = await Selector('div.p-photo-clipboard').innerText;
    await t
        .expect(countSelected).contains('menu')
});