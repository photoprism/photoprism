import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from "./page-model";
import { RequestLogger } from 'testcafe';

const logger = RequestLogger({ url: /http:\/\/localhost:2342/, method: 'post'}  , {
    logResponseHeaders: true,
    logResponseBody:    true,
    stringifyResponseBody: true
});

fixture`Test batch story`
    .page`${testcafeconfig.url}`
    .requestHooks(logger);

const page = new Page();

test('Add stories flag to photos', async t => {
    await page.selectPhoto(0);
    await page.selectPhoto(2);
    await t
        .click(Selector('div.p-photo-clipboard'))
        .click(Selector('.p-photo-clipboard-story'), {timeout: 15000});
    const request = await logger.requests[0].responseBody;
    await t
        .expect(logger.requests[0].response.statusCode).eql(200)
        .expect(logger.requests[0].response.body).contains('photos marked as story');
    const countSelected = await Selector('div.p-photo-clipboard').innerText;
    await t
        .expect(countSelected).contains('menu');
});