import { Selector, t } from "testcafe";
import { RequestLogger } from "testcafe";
import Menu from "./menu";
import Album from "./album";
import Toolbar from "./toolbar";
import ContextMenu from "./context-menu";
import ShareDialog from "./dialog-share";
import Photo from "./photo";
import PhotoViewer from "./photoviewer";

const logger = RequestLogger(/http:\/\/localhost:2343\/api\/v1\/*/, {
  logResponseHeaders: true,
  logResponseBody: true,
});

const menu = new Menu();
const album = new Album();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const sharedialog = new ShareDialog();
const photoHelper = new Photo();
const photoviewerHelper = new PhotoViewer();

export default class Page {
  constructor() {
    this.selectOption = Selector('div[role="option"]', { timeout: 15000 });
    this.cardTitle = Selector("button.action-title-edit", { timeout: 7000 });
    this.cardDescription = Selector("button.meta-description", { timeout: 7000 });
    this.cardCaption = Selector("button.meta-caption", { timeout: 7000 });
    this.cardLocation = Selector("button.action-location", { timeout: 7000 });
    this.cardTaken = Selector("button.action-open-date", { timeout: 7000 });
    this.usernameInput = Selector(".input-username input", { timeout: 7000 });
    this.passwordInput = Selector(".input-password input", { timeout: 7000 });
    this.passcodeInput = Selector(".input-code input", { timeout: 7000 });
    this.togglePasswordMode = Selector(".v-field__append-inner", { timeout: 7000 });
    this.loginAction = Selector(".action-confirm", { timeout: 7000 });
    this.snackbar = Selector(".v-snackbar__content");
  }

  async login(username, password) {
    await t
      .typeText(Selector(".input-username input"), username, { replace: true })
      .typeText(Selector(".input-password input"), password, { replace: true })
      .click(Selector(".action-confirm"));
  }

  async logout() {
    await menu.openNav();
    await t.click(Selector("button i.mdi-power"));
  }

  async clickCardTitleOfUID(uid) {
    await t.click(Selector('div[data-uid="' + uid + '"] button.action-title-edit'));
  }

  async testCreateEditDeleteSharingLink(type) {
    await menu.openPage(type);
    const FirstAlbum = await album.getNthAlbumUid("all", 0);
    await album.triggerHoverAction("uid", FirstAlbum, "select");
    await contextmenu.checkContextMenuCount("1");
    await contextmenu.triggerContextMenuAction("share", "", "");
    const InitialUrl = await sharedialog.linkUrl.value;
    const InitialSecret = await sharedialog.linkSecretInput.value;
    const InitialExpire = await Selector(".input-expires .v-select__selection-text").innerText;
    await t
      .expect(InitialUrl)
      .notContains("secretfortesting")
      .expect(InitialExpire)
      .contains("Never")
      .typeText(sharedialog.linkSecretInput, "secretForTesting", { replace: true })
      .click(sharedialog.linkExpireInput)
      .click(Selector("div").withText("After 1 day").parent('div[role="option"]'))
      .click(sharedialog.dialogSave)
      .click(sharedialog.dialogClose);
    await contextmenu.clearSelection();
    await album.openAlbumWithUid(FirstAlbum);
    await toolbar.triggerToolbarAction("share", "");
    const ExpireAfterChange = await Selector(".input-expires .v-select__selection-text").innerText;
    const UrlAfterChange = await sharedialog.linkUrl.value;
    await t
      .expect(UrlAfterChange)
      .contains("secretfortesting")
      .expect(ExpireAfterChange)
      .contains("After 1 day")
      .typeText(sharedialog.linkSecretInput, InitialSecret, { replace: true })
      .click(sharedialog.linkExpireInput)
      .pressKey("down")
      .click(Selector("div").withText("Never").parent('div[role="option"]'))
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
      await toolbar.triggerToolbarAction("refresh");
    }
    await album.triggerHoverAction("uid", FirstAlbum, "share");
    await t.click(sharedialog.deleteLink);
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

  async testSetAlbumCover(pageName) {
    await menu.openPage(pageName);

    const maxAlbumsToCheck = 5; // Limit checking to avoid infinite loops
    let foundSuitableAlbum = false;

    for (let i = 0; i < maxAlbumsToCheck; i++) {
      await menu.openPage(pageName);
      await t.wait(500); // Wait for page potentially loading

      const albumCard = Selector("div.result.is-album").nth(i);
      if (!(await albumCard.exists)) {
        break; // Stop if there are no more albums
      }

      const AlbumUid = await albumCard.getAttribute("data-uid");
      if (!AlbumUid) {
        continue; // Skip to next album if UID is missing
      }

      const initialCoverStyle = await album.getAlbumCoverStyle(AlbumUid);
      await t
        .expect(initialCoverStyle !== undefined)
        .ok(`Could not get initial cover style for album ${AlbumUid} on ${pageName} page.`);

      await album.openAlbumWithUid(AlbumUid);

      const photoCount = await photoHelper.getPhotoCount("all");

      if (photoCount > 1) {
        let foundCoverCandidate = false;
        let CoverPhotoUid = null;
        let expectedCoverStyle = null;
        for (let photoIdx = 0; photoIdx < photoCount; photoIdx++) {
          const candidatePhotoUid = await photoHelper.getNthPhotoUid("all", photoIdx);
          const candidateCoverStyle = await photoHelper.getPhotoPreviewStyle(candidatePhotoUid);
          if (candidateCoverStyle !== initialCoverStyle) {
            CoverPhotoUid = candidatePhotoUid;
            expectedCoverStyle = candidateCoverStyle;
            foundCoverCandidate = true;
            break;
          }
        }
        if (!foundCoverCandidate) {
          await menu.openPage(pageName);
          continue;
        }
        foundSuitableAlbum = true;

        await photoviewerHelper.openPhotoViewer("uid", CoverPhotoUid);
        await photoviewerHelper.checkPhotoViewerActionAvailability("cover", true);
        await photoviewerHelper.triggerPhotoViewerAction("cover");
        await t.wait(500);
        await photoviewerHelper.triggerPhotoViewerAction("close-button");

        await menu.openPage(pageName);
        await t.wait(500);

        const newCoverStyle = await album.getAlbumCoverStyle(AlbumUid);
        await t
          .expect(newCoverStyle !== undefined)
          .ok(`Could not get new cover style for album ${AlbumUid} on ${pageName} page after setting cover.`);

        await t
          .expect(newCoverStyle)
          .notEql(
            initialCoverStyle,
            `Album card cover background image should change (Album: ${AlbumUid}, Initial: ${initialCoverStyle}, New: ${newCoverStyle})`
          )
          .expect(newCoverStyle)
          .eql(
            expectedCoverStyle,
            `Album card cover background image URL should match the thumbnail (Album: ${AlbumUid}, Expected: ${expectedCoverStyle}, Actual: ${newCoverStyle})`
          );
        break;
      }
    }

    await t
      .expect(foundSuitableAlbum)
      .ok(
        `Failed to find any album with more than 1 photo within the first ${maxAlbumsToCheck} albums on ${pageName} page.`
      );
  }
}
