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
        .click('button.p-expand-search');
    logger.clear();
    await page.setFilter('camera', 'iPhone 6');
    const request = await logger.requests[0].responseBody;
    await t
        .expect(logger.requests[0].response.statusCode).eql(200)
        .expect(Selector('div.v-image__image').visible).ok();
}),
    test('Test time filter', async t => {
        await t
            .click('button.p-expand-search');
        logger.clear();
        await page.setFilter('time', 'Oldest');
        const request2 = await logger.requests[0].responseBody;
        await t
            .expect(logger.requests[0].response.statusCode).eql(200)
            .expect(logger.requests[0].request.url).contains('order=oldest')
            .expect(Selector('div.v-image__image').visible).ok();
    }),
    test('Test countries filter', async t => {
        await t
            .click('button.p-expand-search');
        logger.clear();
        await page.setFilter('countries');
        const request3 = await logger.requests[0].responseBody;
        await t
            .expect(logger.requests[0].response.statusCode).eql(200)
            .expect(logger.requests[0].request.url).contains('country=')
            .expect(Selector('div.v-image__image').visible).ok();
    },);
