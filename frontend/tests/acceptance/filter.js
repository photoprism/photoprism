import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from './page-model';
import { RequestLogger } from 'testcafe';

const logger = RequestLogger( /http:\/\/localhost:2342\/api\/v1\/photos*/ , {
    logResponseHeaders: true,
    logResponseBody:    true
});

fixture`Use filters`
    .page`${testcafeconfig.url}`
    .requestHooks(logger);

const page = new Page();

test('Test camera filter', async t => {
    await t
        .click('#advancedMenu');
        await page.setFilter('camera', 'iPhone 6');
    await t
        .expect(logger.requests[1].response.statusCode).eql(200)
        .expect(Selector('div.v-image__image').visible).ok();
}),
test('Test time filter', async t => {
    await t
        .click('#advancedMenu');
    await page.setFilter('time', 'Oldest');
    await t
        .expect(logger.requests[1].response.statusCode).eql(200)
        .expect(logger.requests[1].request.url).contains('order=oldest')
        .expect(Selector('div.v-image__image').visible).ok();
}),
    test('Test countries filter', async t => {
        await t
            .click('#advancedMenu');
        await page.setFilter('countries', 'Cuba');
        await t
            .expect(logger.requests[1].response.statusCode).eql(200)
            .expect(logger.requests[1].request.url).contains('country=cu')
            .expect(Selector('div.v-image__image').visible).ok();
    },);
