import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from "./page-model";

fixture`Test places page`
    .page `localhost:2342/places`;

const page = new Page();

test('Test places', async t => {
    await t
        .expect(Selector('div.leaflet-map-pane').exists).ok()
        .expect(Selector('img.leaflet-marker-icon').visible).ok()
        .typeText(Selector('input[aria-label="Search"]'), 'Berlin')
        .pressKey('enter')
        .expect(Selector('img.leaflet-marker-icon').visible).ok();
});