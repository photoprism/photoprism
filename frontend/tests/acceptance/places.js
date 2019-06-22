import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from "./page-model";
import { RequestLogger } from 'testcafe';

const logger = RequestLogger( /http:\/\/localhost:2342\/api\/v1\/photos*/ , {
    logResponseHeaders: true,
    logResponseBody:    true
});

fixture`Test places page`
    .page `localhost:2342/places`
    .requestHooks(logger);

const page = new Page();

test('Test places', async t => {
    await t
        .expect(Selector('div.leaflet-map-pane').exists, {timeout: 15000}).ok()
        .expect(Selector('img.leaflet-marker-icon').visible).ok()
    const request = await logger.requests[0];
    await t
        .expect(logger.requests[0].response.statusCode).eql(200)
    logger.clear();
    await t
        .typeText(Selector('input[aria-label="Search"]'), 'Berlin')
        .pressKey('enter');
    const request2 = await logger.requests[0];
    await t
        .expect(Selector('img.leaflet-marker-icon').visible).ok()
        .expect(logger.requests[0].response.statusCode).eql(200, 'status code equals 200')
        .expect(logger.requests[0].request.url).contains('q=Berlin');
});