import { Selector } from 'testcafe';
import { ClientFunction } from 'testcafe';

const getLocation = ClientFunction(() => document.location.href);

fixture`Test places page`
    .page `localhost:2342/places`;

test('Test places', async t => {
    await t
        .expect(Selector('#map').exists, {timeout: 15000}).ok()
        .expect(Selector('div.p-map-control').visible).ok();
    await t
        .typeText(Selector('input[aria-label="Search"]'), 'Berlin')
        .pressKey('enter');
    await t
        .expect(Selector('div.p-map-control').visible).ok()
        .expect(getLocation()).contains('Berlin');
});
