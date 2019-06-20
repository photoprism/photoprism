import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from "./page-model";
import { RequestLogger } from 'testcafe';

const logger = RequestLogger( /http:\/\/localhost:2342\/api\/v1\/photos*/ , {
    logResponseHeaders: true,
    logResponseBody:    true
});

fixture`Search photos`
    .page`${testcafeconfig.url}`
    .requestHooks(logger);

const page = new Page();

test('Test search object', async t => {
    await page.search('label:cat');
    await t
        .expect(logger.requests[0].response.statusCode).eql(200)
        .expect(Selector('div.v-image__image').visible).ok();
}),
test('Test search color', async t => {
    await page.search('color:pink');
    await t
        .expect(logger.requests[1].response.statusCode).eql(200)
        .expect(Selector('div.v-image__image').visible).ok();
});
