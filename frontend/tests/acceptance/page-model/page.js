import { Selector, t } from "testcafe";
import { RequestLogger } from "testcafe";
import Menu from "./menu";
import Album from "./album";
import Toolbar from "./toolbar";
import ContextMenu from "./context-menu";
import ShareDialog from "./dialog-share";

const logger = RequestLogger(/http:\/\/localhost:2343\/api\/v1\/*/, {
  logResponseHeaders: true,
  logResponseBody: true,
});

const menu = new Menu();
const album = new Album();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const sharedialog = new ShareDialog();

export default class Page {
  constructor() {
    this.selectOption = Selector("div.v-list__tile__title", { timeout: 15000 });
    this.cardTitle = Selector("button.action-title-edit", { timeout: 7000 });
    this.cardDescription = Selector('div[title="Description"]', { timeout: 7000 });
    this.cardLocation = Selector("button.action-location", { timeout: 7000 });
    this.cardTaken = Selector("div.caption button.action-open-date", { timeout: 7000 });
    this.usernameInput = Selector(".input-username input", { timeout: 7000 });
    this.passwordInput = Selector(".input-password input", { timeout: 7000 });
    this.passcodeInput = Selector(".input-code input", { timeout: 7000 });
    this.togglePasswordMode = Selector(".v-input__icon--append", { timeout: 7000 });
    this.loginAction = Selector(".action-confirm", { timeout: 7000 });
  }

  async login(username, password) {
    await t
      .typeText(Selector(".input-username input"), username, { replace: true })
      .typeText(Selector(".input-password input"), password, { replace: true })
      .click(Selector(".action-confirm"));
  }

  async logout() {
    await menu.openNav();
    await t.click(Selector('div[title="Logout"]'));
  }

  async testCreateEditDeleteSharingLink(type) {
    await menu.openPage(type);
    const FirstAlbum = await album.getNthAlbumUid("all", 0);
    await album.triggerHoverAction("uid", FirstAlbum, "select");
    await contextmenu.checkContextMenuCount("1");
    await contextmenu.triggerContextMenuAction("share", "", "");
    await t.click(sharedialog.expandLink.nth(0));
    const InitialUrl = await sharedialog.linkUrl.innerText;
    const InitialSecret = await sharedialog.linkSecretInput.value;
    const InitialExpire = await Selector("div.v-select__selections").innerText;
    await t
      .expect(InitialUrl)
      .notContains("secretfortesting")
      .expect(InitialExpire)
      .contains("Never")
      .typeText(sharedialog.linkSecretInput, "secretForTesting", { replace: true })
      .click(sharedialog.linkExpireInput)
      .click(Selector("div").withText("After 1 day").parent('div[role="listitem"]'))
      .click(sharedialog.dialogSave)
      .click(sharedialog.dialogClose);
    await contextmenu.clearSelection();
    await album.openAlbumWithUid(FirstAlbum);
    await toolbar.triggerToolbarAction("share", "");
    await t.click(sharedialog.expandLink.nth(0));
    await t.wait(5000);
    const UrlAfterChange = await sharedialog.linkUrl.innerText;
    const ExpireAfterChange = await Selector("div.v-select__selections").innerText;
    await t
      .expect(UrlAfterChange)
      .contains("secretfortesting")
      .expect(ExpireAfterChange)
      .contains("After 1 day")
      .typeText(sharedialog.linkSecretInput, InitialSecret, { replace: true })
      .click(sharedialog.linkExpireInput)
      .click(Selector("div").withText("Never").parent('div[role="listitem"]'))
      .click(sharedialog.dialogSave)
      .click(sharedialog.expandLink);
    const LinkCount = await Selector(".action-url").count;
    await t.click(sharedialog.addLink);
    const LinkCountAfterAdd = await Selector(".action-url").count;
    await t
      .expect(LinkCountAfterAdd)
      .eql(LinkCount + 1)
      .click(sharedialog.expandLink)
      .click(sharedialog.deleteLink);
    const LinkCountAfterDelete = await Selector(".action-url").count;
    await t
      .expect(LinkCountAfterDelete)
      .eql(LinkCountAfterAdd - 1)
      .click(sharedialog.dialogClose);
    await menu.openPage(type);
    if (t.browser.platform === "mobile") {
      await t.eval(() => location.reload());
    } else {
      await toolbar.triggerToolbarAction("reload");
    }
    await album.triggerHoverAction("uid", FirstAlbum, "share");
    await t.click(sharedialog.expandLink).click(sharedialog.deleteLink);
  }

  async validateDownloadRequest(request, filename, extension) {
    const downloadedFileName = request.headers["content-disposition"];
    await t
      .expect(request.statusCode === 200)
      .ok()
      .expect(downloadedFileName)
      .contains(filename)
      .expect(downloadedFileName)
      .contains(extension);
    await logger.clear();
  }
}
