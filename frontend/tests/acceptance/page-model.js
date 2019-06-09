import { Selector, t } from 'testcafe';

export default class Page {
    constructor() {
        this.view = Selector('#viewFlex', {timeout: 15000});
        this.camera = Selector('#cameraFlex', {timeout: 15000});
        this.countries = Selector('#countriesFlex', {timeout: 15000});
        this.time = Selector('#timeFlex', {timeout: 15000});
        this.search1 = Selector('#search', {timeout: 15000});
    }

    async setFilter(filter, option) {
        await t;

        switch (filter) {
            case 'view':
                await t
                    .click(this.view);
                break;
            case 'camera':
                await t
                    .click(this.camera);
                break;
            case 'time':
                await t
                    .click(this.time);
                break;
            case 'countries':
                await t
                    .click(this.countries);
                break;
            default:
        }
        await t
            .click(Selector('a').withText(option))
    }

    async search(term) {
        await t
            .typeText(this.search1, term)
            .pressKey('enter')
    }

    async openNav() {
        if (await Selector('button.p-navigation-show').visible) {
            await t.click(Selector('button.p-navigation-show'));
        } else if (await Selector('div.p-navigation-expand').exists) {
            await t.click(Selector('div.p-navigation-expand i'));
        }
    }
}
