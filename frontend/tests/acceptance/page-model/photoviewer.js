import { Selector, t } from "testcafe";
import Toolbar from "./toolbar";
import Photo from "./photo";

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
    // Edit dialogs launched from the sidebar pencils. Timeouts are generous
    // enough for Vuetify teleport mounting + reverse-geocoder lookups.
    this.dateTimeDialog = Selector(".p-datetime-dialog", { timeout: 15000 });
    this.cameraDialog = Selector(".p-camera-dialog", { timeout: 15000 });
    this.locationDialog = Selector(".p-location-dialog", { timeout: 15000 });
  }

  // Locate the v-list-item that contains a given MDI prepend-icon.
  // The icon class name is applied to Vuetify's rendered `<i>` element,
  // so this matches rows deterministically without relying on the DOM
  // order of siblings.
  sidebarRow(iconClass) {
    return Selector("." + iconClass).parent(".p-sidebar-info .v-list-item");
  }

  // Return the section-level v-list-item for a subtitle label such as
  // "Keywords", "Notes", "Labels", or "Albums". The pencil in these rows
  // lives in the section header rather than on the value row below.
  sidebarSection(sectionLabel) {
    return Selector(".p-sidebar-info .text-subtitle-2").withText(sectionLabel).parent(".p-sidebar-info .v-list-item");
  }

  async startInlineEditByIcon(iconClass) {
    const row = this.sidebarRow(iconClass);
    await t.click(row.find(".meta-inline-pencil"));
    const input = row.find(".meta-inline-edit").find("input,textarea");
    await t.expect(input.visible).ok();
    return input;
  }

  async confirmInlineEditByIcon(iconClass) {
    await t.click(this.sidebarRow(iconClass).find(".meta-inline-confirm"));
  }

  async startInlineEditBySection(sectionLabel) {
    const section = this.sidebarSection(sectionLabel);
    await t.click(section.find(".meta-inline-pencil"));
    // The active editor lives in a sibling v-list-item outside this section
    // header, so it must be queried from the sidebar root. `meta-inline-marker`
    // inputs (one per face marker) are always rendered in edit mode, so they
    // must be excluded — otherwise typeText lands in a marker's name field.
    const input = Selector(".p-sidebar-info .meta-inline-edit:not(.meta-inline-marker)").find("input,textarea");
    await t.expect(input.visible).ok();
    return input;
  }

  async confirmInlineEditBySection(sectionLabel) {
    await t.click(this.sidebarSection(sectionLabel).find(".meta-inline-confirm"));
  }

  // Vuetify's combobox may swallow the same Enter event that seeds the
  // chip, so short-wait between typing and pressing Enter.
  async typeAndConfirmInlineChip(sectionLabel, value) {
    const input = await this.startInlineEditBySection(sectionLabel);
    await t.typeText(input, value);
    await t.wait(200);
    await t.pressKey("enter");
    await this.confirmInlineEditBySection(sectionLabel);
  }

  // Title/Caption enter editing from either a pencil (value present) or
  // an add-prompt (empty); the test doesn't know which until it looks.
  async startInlineEditOrAdd(displayClass, promptLabel) {
    const display = Selector(".p-sidebar-info ." + displayClass, { timeout: 15000 });
    if (await display.exists) {
      await t.click(display.parent(".p-sidebar-info .v-list-item").find(".meta-inline-pencil"));
    } else {
      await t.click(Selector(".p-sidebar-info .meta-add-prompt").withText(promptLabel));
    }
  }

  async openSidebarOnFirstPhoto() {
    const toolbar = new Toolbar();
    const photo = new Photo();
    await t.click(toolbar.cardsViewAction);
    const uid = await photo.getNthPhotoUid("image", 0);
    await this.openPhotoViewer("uid", uid);
    await this.openInfoSidebar();
    return uid;
  }

  async openSidebarDialog(which) {
    if (which === "takenAt") {
      await t.click(this.sidebarRow("mdi-calendar").find(".meta-inline-pencil"));
      await t.expect(this.dateTimeDialog.visible).ok();
    } else if (which === "camera") {
      await t.click(this.sidebarRow("mdi-camera").find(".meta-inline-pencil"));
      await t.expect(this.cameraDialog.visible).ok();
    } else if (which === "location") {
      // Two rows (Location and Coordinates) host a location pencil depending
      // on whether the photo has lat/lng; both carry the modifier class that
      // disambiguates them from the other inline metadata pencils.
      await t.click(Selector(".p-sidebar-info .meta-inline-pencil--location"));
      await t.expect(this.locationDialog.visible).ok();
    } else {
      throw new Error(`Unknown sidebar dialog: ${which}`);
    }
  }

  async openInfoSidebar() {
    if (!(await this.sidebar.exists)) {
      await t.click(Selector("button.pswp__button--info-button"));
    }
    await t.expect(this.sidebar.visible).ok();
  }

  // Close the lightbox. The PhotoSwipe close button uses the
  // `--close-button` suffix, not `--close`, so `triggerPhotoViewerAction`
  // cannot reach it with its generic `--${action}` pattern.
  async closePhotoViewer() {
    await t.click(Selector("button.pswp__button--close-button"));
    await t.expect(this.viewer.exists).notOk();
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
