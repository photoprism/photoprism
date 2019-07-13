import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from "./page-model";
import { RequestLogger } from 'testcafe';

const logger = RequestLogger( /http:\/\/localhost:2342\/api\/v1\*/ , {
    logResponseHeaders: true,
    logResponseBody:    true
});

fixture`Library page`
    .page`localhost:2342/library`
    .requestHooks(logger);

const page = new Page();


test('Upload image', async t => {
    await t
        //.click(Selector(button).withText('Upload'));

});
