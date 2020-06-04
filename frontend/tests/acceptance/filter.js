import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from './page-model';
import { RequestLogger } from 'testcafe';

const logger = RequestLogger( /http:\/\/localhost:2342\/api\/v1\/photos*/ , {
    logResponseHeaders: true,
    logResponseBody:    true
});

fixture`Test filter options`
    .page`${testcafeconfig.url}`
    .requestHooks(logger);

const page = new Page();

/*test('Test filter options', async t => {
    await t
        .click('button.p-expand-search')
        .click(Selector('div.p-countries-select'))
        .expect(Selector('div[role="listitem"]]').nth(0).innerText).notContains('object')
        .expect(Selector('div[role="listitem"]]').nth(0).innerText).notContains('Botswana')
        .expect(Selector('div[role="listitem"]]').nth(0).innerText).notContains('Animal');
});*/
