import { Selector } from 'testcafe';
import { ClientFunction } from 'testcafe';
import testcafeconfig from "./testcafeconfig.json";

const getLocation = ClientFunction(() => document.location.href);

fixture`Test places page`
    .page`${testcafeconfig.url}`

/*test('#1 Test places', async t => {
    await t
        .click(Selector('.nav-places'))
        .expect(Selector('#map').exists, {timeout: 15000}).ok()
        .expect(Selector('div.p-map-control').visible).ok();
    await t
        .typeText(Selector('input[aria-label="Search"]'), 'Berlin')
        .pressKey('enter');
    await t
        .expect(Selector('div.p-map-control').visible).ok()
        .expect(getLocation()).contains('Berlin');
});*/

test('#2 Open photo from places', async t => {
   //TODO replace wait
    await t
        .click(Selector('.nav-places'))
        .expect(Selector('#p-photo-viewer').visible).notOk()
        .expect(Selector('#map').exists, {timeout: 15000}).ok()
        .typeText(Selector('input[aria-label="Search"]'), 'Berlin')
        .pressKey('enter')
        .wait(30000)
        .click(Selector('div.marker').nth(0), {timeout: 9000})
        .expect(Selector('#p-photo-viewer').visible).ok();
});