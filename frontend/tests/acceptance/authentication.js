import { Selector } from 'testcafe';
import testcafeconfig from './testcafeconfig';
import Page from "./page-model";

fixture`Test login and logout`
    .page`${testcafeconfig.url}`;

const page = new Page();

test('Login', async t => {
    await page.openNav();
    await t
        .expect(Selector('a[href="/library"]').exists, {timeout: 5000}).notOk()
        .expect(Selector('a[href="/settings"]').exists, {timeout: 5000}).notOk()
        .click(Selector('a[href="/login"]'));
    await page.login('photoprism');
    await page.openNav();
    await  t
        .expect(Selector('a[href="/library"]').exists, {timeout: 5000}).ok()
        .expect(Selector('a[href="/settings"]').exists, {timeout: 5000}).ok()
        .expect(Selector('a[href="/login"]').exists, {timeout: 5000}).notOk();
}),
test('Logout', async t => {
    await page.openNav();
    await t
        .click(Selector('a[href="/login"]'));
    await page.login('photoprism');
    await page.openNav();
    await  t
        .expect(Selector('a[href="/library"]').exists, {timeout: 5000}).ok()
        .expect(Selector('a[href="/settings"]').exists, {timeout: 5000}).ok()
        .expect(Selector('a[href="/login"]').exists, {timeout: 5000}).notOk()
    await page.logout();
    await page.openNav();
    await  t
        .expect(Selector('a[href="/library"]').exists, {timeout: 5000}).notOk()
        .expect(Selector('a[href="/settings"]').exists, {timeout: 5000}).notOk()
        .expect(Selector('a[href="/login"]').exists, {timeout: 5000}).ok()
});