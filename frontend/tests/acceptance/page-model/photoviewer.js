import { Selector, t } from "testcafe";

export default class Page {
  constructor() {
    this.view = Selector("div.p-view-select", { timeout: 15000 });
    this.camera = Selector("div.p-camera-select", { timeout: 15000 });
    this.countries = Selector("div.p-countries-select", { timeout: 15000 });
    this.time = Selector("div.p-time-select", { timeout: 15000 });
    this.search1 = Selector("div.input-search input", { timeout: 15000 });
    this.menuButton = Selector("button.pswp__button--menu-button", { timeout: 15000 });
    this.viewer = Selector("div.p-lightbox__pswp", { timeout: 15000 });
    this.caption = Selector("div.pswp__caption__center", { timeout: 5000 });
    this.muteButton = Selector("button.pswp__button--mute", { timeout: 5000 });
    this.playButton = Selector('[class^="pswp__button pswp__button--slideshow-toggle pswp__"]', { timeout: 5000 });
    this.favoriteOnIcon = Selector("button.action-favorite i.icon-favorite", { timeout: 5000 });
    this.favoriteOffIcon = Selector("button.action-favorite i.icon-favorite-border", { timeout: 5000 });
    // Sidebar info + face markers.
    this.sidebar = Selector("div.p-lightbox__sidebar", { timeout: 15000 });
    this.sidebarInfo = Selector("div.p-sidebar-info", { timeout: 15000 });
    this.markersVisibilityToggle = Selector(".meta-markers-toggle", { timeout: 15000 });
    this.markerAddButton = Selector(".meta-marker-add", { timeout: 15000 });
    this.markerRemoveButton = Selector(".meta-marker-remove", { timeout: 5000 });
    this.faceMarkerOverlay = Selector("div.p-face-markers", { timeout: 15000 });
    this.faceMarkerRect = Selector("rect.p-face-markers__rect", { timeout: 15000 });
    this.faceMarkerDraft = Selector("rect.p-face-markers__rect--draft", { timeout: 15000 });
    this.faceMarkerConfirmButton = Selector("button.p-face-markers__btn--confirm", { timeout: 15000 });
    this.faceMarkerCancelButton = Selector("button.p-face-markers__btn--cancel", { timeout: 15000 });
    this.personRow = Selector(".metadata__person-row", { timeout: 15000 });
    // Inline edit affordances in the sidebar. Pencils share the same
    // `meta-inline-pencil` class across all rows, so tests that need a
    // specific row scope it via a row selector (see sidebarRow below).
    this.inlinePencils = Selector(".p-sidebar-info .meta-inline-pencil", { timeout: 15000 });
    this.inlineEditInputs = Selector(".p-sidebar-info .meta-inline-edit input, .p-sidebar-info .meta-inline-edit textarea", { timeout: 15000 });
    this.inlineConfirm = Selector(".p-sidebar-info .meta-inline-confirm", { timeout: 15000 });
    this.inlineAddPrompt = Selector(".p-sidebar-info .meta-add-prompt", { timeout: 15000 });
    this.sidebarTitle = Selector(".p-sidebar-info .meta-title", { timeout: 15000 });
    this.sidebarCaption = Selector(".p-sidebar-info .meta-caption", { timeout: 15000 });
    this.sidebarKeywords = Selector(".p-sidebar-info .meta-keywords", { timeout: 15000 });
    this.sidebarNotes = Selector(".p-sidebar-info .meta-notes", { timeout: 15000 });
    // All rendered chips in the sidebar (labels + albums + pending
    // additions). Individual tests filter further if they need a
    // specific section.
    this.sidebarChips = Selector(".p-sidebar-info .meta-chip", { timeout: 15000 });
    this.faceMarkerEjectButton = Selector(".metadata__person-row .meta-marker-eject", { timeout: 15000 });
    this.faceMarkerNameInput = Selector(".metadata__person-row .meta-inline-marker input", { timeout: 15000 });
  }

  // Locate the v-list-item that contains a given MDI prepend-icon.
  // The icon class name is applied to Vuetify's rendered `<i>` element,
  // so this matches rows deterministically without relying on the DOM
  // order of siblings.
  sidebarRow(iconClass) {
    return Selector("." + iconClass).parent(".p-sidebar-info .v-list-item");
  }

  async openInfoSidebar() {
    if (!(await this.sidebar.exists)) {
      await t.click(Selector("button.pswp__button--info-button"));
    }
    await t.expect(this.sidebar.visible).ok();
  }

  async toggleMarkersVisible() {
    await t.click(this.markersVisibilityToggle);
  }

  async startAddingMarker() {
    await t.click(this.markerAddButton);
  }

  async cancelMarkerDraft() {
    await t.click(this.faceMarkerCancelButton);
  }

  async confirmMarkerDraft() {
    await t.click(this.faceMarkerConfirmButton);
  }

  async getRenderedMarkerCount() {
    return this.faceMarkerRect.count;
  }

  async getPersonRowCount() {
    return this.personRow.count;
  }

  async drawMarkerOnImage(fromX, fromY, toX, toY) {
    // Drag the overlay from one viewport coordinate to another. The
    // overlay must already be in draw mode (after clicking +).
    await t.drag(this.faceMarkerOverlay, toX - fromX, toY - fromY, {
      offsetX: fromX,
      offsetY: fromY,
    });
  }

  // Draw a small rectangle in the middle of the overlay, sized in
  // percent of the overlay's actual box. Avoids viewport-dependent
  // coordinates (Mac Chrome vs headless Linux) that can land outside
  // the rendered photo and fail the draft.
  async drawMarkerInCenter(sizePercent = 0.2) {
    const width = await this.faceMarkerOverlay.clientWidth;
    const height = await this.faceMarkerOverlay.clientHeight;
    const boxW = Math.max(Math.floor(width * sizePercent), 40);
    const boxH = Math.max(Math.floor(height * sizePercent), 40);
    const fromX = Math.floor(width / 2 - boxW / 2);
    const fromY = Math.floor(height / 2 - boxH / 2);
    await t.drag(this.faceMarkerOverlay, boxW, boxH, {
      offsetX: fromX,
      offsetY: fromY,
    });
  }

  async openPhotoViewer(mode, uidOrNth) {
    if (mode === "uid") {
      await t.hover(Selector("div.is-photo").withAttribute("data-uid", uidOrNth));
      if (await Selector(`.uid-${uidOrNth} button.input-open`).visible) {
        await t.click(Selector(`.uid-${uidOrNth} button.input-open`));
      } else {
        await t.click(Selector("div.is-photo").withAttribute("data-uid", uidOrNth));
      }
    } else if (mode === "nth") {
      await t.hover(Selector("div.is-photo").nth(uidOrNth));
      if (await Selector(`div.is-photo button.input-open`).visible) {
        await t.click(Selector(`div.is-photo button.input-open`));
      } else {
        await t.click(Selector("div.is-photo").nth(uidOrNth));
      }
    }
    await t.expect(Selector("div.p-lightbox__pswp").visible).ok();
  }

  async checkPhotoViewerActionAvailability(action, visible) {
    if (action === "cover") {
      await t.hover(this.menuButton);
      if (visible) {
        await t.expect(Selector("div.action-" + action).visible).ok();
      } else {
        await t.expect(Selector("div.action-" + action).visible).notOk();
      }
    } else if (action === "download") {
      await t.hover(this.menuButton);
      if (visible) {
        await t.expect(Selector("div.action-" + action).visible).ok();
      } else {
        await t.expect(Selector("div.action-" + action).visible).notOk();
      }
    } else {
      if (visible) {
        await t.expect(Selector("button.pswp__button--" + action).visible).ok();
      } else {
        await t.expect(Selector("button.pswp__button--" + action).visible).notOk();
      }
    }
  }

  async triggerPhotoViewerAction(action) {
    if (action === "cover") {
      await t.hover(this.menuButton);
      await t.click(Selector("div.action-" + action));
    } else if (action === "download") {
      await t.hover(this.menuButton);
      await t.click(Selector("div.action-" + action));
    } else {
      await t.hover(Selector("button.pswp__button--" + action));
      await t.click(Selector("button.pswp__button--" + action));
    }
    if (t.browser.platform === "mobile") {
      await t.wait(5000);
    }
  }
}
