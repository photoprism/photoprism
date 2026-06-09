import { Selector, t } from "testcafe";
import Toolbar from "./toolbar";
import Photo from "./photo";
import DateTimeDialog from "./dialog/date-time";
import CameraDialog from "./dialog/camera";
import LocationDialog from "./dialog/location";
import { clickIfVisible } from "./helpers";

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
    this.sidebar = Selector("div.p-lightbox__sidebar div.p-lightbox-sidebar", { timeout: 15000 });
    // People-section toggles render one of two variants depending on role.
    // Both carry the shared `.meta-markers-toggle` role class (since edit mode
    // also toggles visibility), plus a dedicated discriminator:
    // - eye (non-editable, read-only display): `.meta-faces-display`
    // - pencil (editable, edit-mode superset of display): `.meta-faces-edit`
    //
    // `markersVisibilityToggle` matches the eye-only variant; `markersEditToggle`
    // matches the pencil-only variant.
    this.markersVisibilityToggle = Selector(".meta-faces-display", { timeout: 15000 });
    this.markersEditToggle = Selector(".meta-faces-edit", { timeout: 15000 });
    this.faceMarkerOverlay = Selector("div.p-meta-face-markers", { timeout: 15000 });
    this.faceMarkerRect = Selector("rect.p-meta-face-markers__rect", { timeout: 15000 });
    this.faceMarkerUnnamedRect = Selector(
      "rect.p-meta-face-markers__rect:not(.p-meta-face-markers__rect--named):not(.p-meta-face-markers__rect--draft):not(.p-meta-face-markers__rect--removing)",
      { timeout: 15000 }
    );
    this.faceMarkerDraft = Selector("rect.p-meta-face-markers__rect--draft", { timeout: 15000 });
    this.faceMarkerConfirmButton = Selector("button.p-meta-face-markers__btn--confirm", { timeout: 15000 });
    this.faceMarkerCancelButton = Selector("button.p-meta-face-markers__btn--cancel", { timeout: 15000 });
    this.faceMarkerRemoveConfirm = Selector("div.p-meta-face-markers__remove-confirm", { timeout: 15000 });
    this.faceMarkerRemoveButton = Selector("button.p-meta-face-markers__btn--remove", { timeout: 15000 });
    this.personRow = Selector(".metadata__person-row", { timeout: 15000 });
    this.sidebarTitle = Selector(".p-lightbox-sidebar .meta-title", { timeout: 15000 });
    this.sidebarCaption = Selector(".p-lightbox-sidebar .meta-caption", { timeout: 15000 });
    this.sidebarKeywords = Selector(".p-lightbox-sidebar .meta-keywords", { timeout: 15000 });
    this.sidebarNotes = Selector(".p-lightbox-sidebar .meta-notes", { timeout: 15000 });
    this.sidebarChips = Selector(".p-lightbox-sidebar .meta-chip", { timeout: 15000 });
    this.faceMarkerClearSubjectButton = Selector(".metadata__person-row .meta-marker-clear-subject", { timeout: 15000 });
    this.faceMarkerNameInput = Selector(".metadata__person-row .meta-inline-marker input", { timeout: 15000 });
    this.peopleHeader = Selector(".p-lightbox-sidebar .text-subtitle-2").withText("People");
    this.addNameDialog = Selector(".v-dialog.p-confirm-dialog", { timeout: 15000 });
    this.addNameDialogCancel = Selector(".v-dialog.p-confirm-dialog .action-cancel", { timeout: 15000 });
    this.dateTimeDialog = new DateTimeDialog();
    this.cameraDialog = new CameraDialog();
    this.locationDialog = new LocationDialog();
  }

  // unnamedPersonRows are sidebar person rows whose inline-marker combobox is in unnamed mode.
  get unnamedPersonRows() {
    return this.personRow.filter((node) => node.querySelector(".meta-inline-marker:not(.meta-inline-marker--named)") !== null);
  }

  // namedPersonRows are sidebar person rows whose inline-marker combobox is in named mode.
  // Discriminates on `.meta-inline-marker--named` rather than the Unassign icon, which is
  // gated on edit mode and would miss named rows when the People section is read-only.
  get namedPersonRows() {
    return this.personRow.filter((node) => node.querySelector(".meta-inline-marker--named") !== null);
  }

  // ensureMarkersEdit toggles the People section into edit mode if it isn't already.
  // The pencil button (`.meta-faces-edit`) flips null ↔ FaceMarkerEdit; calling
  // startMarkersEdit unconditionally would silently EXIT edit mode for callers
  // that are already in it, so probe `.is-active` (the `markersEdit` binding) first.
  async ensureMarkersEdit() {
    if (!(await this.markersEditToggle.hasClass("is-active"))) {
      await this.startMarkersEdit();
    }
  }

  // clearMarkerSubject ensures the People section is in edit mode (the Unassign
  // button is gated on `markersEdit`), hovers the row to reveal the display:none
  // button, then clicks it.
  async clearMarkerSubject(row) {
    await this.ensureMarkersEdit();
    await t.hover(row).click(row.find(".meta-marker-clear-subject"));
  }

  // sidebarRow returns the v-list-item that contains the given MDI prepend-icon class.
  sidebarRow(iconClass) {
    return Selector("." + iconClass).parent(".p-lightbox-sidebar .v-list-item");
  }

  // editTextFieldByKey opens, types, commits, and verifies a sidebar text field by its
  // detailsFields key (`meta-${key}` on the v-list-item).
  // commitKey is "enter" for commitOnEnter fields, "tab" for blur-commit fields (Notes, Caption).
  async editTextFieldByKey(fieldKey, value, commitKey = "enter") {
    const row = Selector(`.p-lightbox-sidebar .v-list-item.meta-${fieldKey}`, { timeout: 15000 });
    await t.click(row);
    const input = row.find(".meta-inline-edit").find("input,textarea");
    await t.expect(input.visible).ok();
    await t.typeText(input, value, { replace: true }).pressKey(commitKey);
    await t.expect(row.withText(value).exists).ok();
  }

  // typeAndConfirmInlineChip adds a Labels or Albums chip via the always-rendered combobox.
  // The short wait lets Vuetify's combobox seat the typed value before Enter commits it.
  async typeAndConfirmInlineChip(sectionLabel, value) {
    const sectionClass = this._chipSectionClass(sectionLabel);
    const input = Selector(`.p-lightbox-sidebar .${sectionClass} .meta-inline-edit input`, { timeout: 15000 });
    await t.click(input).typeText(input, value);
    await t.wait(200);
    await t.pressKey("enter");
    await t.expect(Selector(".meta-inline-menu").exists).notOk();
  }

  // chipByTitle returns the .meta-chip in the Labels or Albums section matching `title`.
  chipByTitle(sectionLabel, title) {
    const sectionClass = this._chipSectionClass(sectionLabel);
    return Selector(`.p-lightbox-sidebar .${sectionClass}.metadata__chips .meta-chip`, { timeout: 15000 }).withText(title);
  }

  // removeInlineChip clicks the × on a Labels or Albums chip matching `title`.
  // Soft-removes locally via chipState.<field>.removals until confirm or undo.
  async removeInlineChip(sectionLabel, title) {
    await t.click(this.chipByTitle(sectionLabel, title).find(".meta-chip__remove"));
  }

  // undoChipRemovals clicks Undo in the Labels or Albums chip toolbar,
  // restoring all soft-removed chips for that section.
  async undoChipRemovals(sectionLabel) {
    const sectionClass = this._chipSectionClass(sectionLabel);
    await t.click(Selector(`.p-lightbox-sidebar .${sectionClass} .meta-chip-undo`, { timeout: 15000 }));
  }

  // confirmChipRemovals clicks Save in the Labels or Albums chip toolbar,
  // persisting all pending removals.
  async confirmChipRemovals(sectionLabel) {
    const sectionClass = this._chipSectionClass(sectionLabel);
    await t.click(Selector(`.p-lightbox-sidebar .${sectionClass} .meta-chip-confirm`, { timeout: 15000 }));
  }

  _chipSectionClass(sectionLabel) {
    if (sectionLabel === "Labels") {
      return "meta-labels";
    }
    if (sectionLabel === "Albums") {
      return "meta-albums";
    }
    throw new Error(`Unknown chip section: ${sectionLabel}`);
  }

  // startInlineEditOrAdd enters edit mode for Title or Caption by clicking the row, or
  // falls back to the add-prompt when the field is empty.
  async startInlineEditOrAdd(displayClass, promptLabel) {
    const display = Selector(".p-lightbox-sidebar ." + displayClass, { timeout: 15000 });
    if (await display.exists) {
      await t.click(display.parent(".p-lightbox-sidebar .v-list-item"));
    } else {
      await t.click(Selector(".p-lightbox-sidebar .meta-add-prompt").withText(promptLabel));
    }
  }

  // openSidebarOnFirstPhoto opens the lightbox + info sidebar on the first image in cards view.
  // Pass a search query to scope the view first (e.g. "faces:1" for a photo with markers).
  async openSidebarOnFirstPhoto(query) {
    const toolbar = new Toolbar();
    const photo = new Photo();
    await t.click(toolbar.cardsViewAction);
    if (query) {
      await toolbar.search(query);
    }
    const uid = await photo.getNthPhotoUid("image", 0);
    await this.openPhotoViewer("uid", uid);
    await this.openSidebar();
    return uid;
  }

  // openSidebarOnPhoto opens the lightbox + info sidebar on a specific photo by UID.
  // Pass a search query when the photo might not be in the current filtered view;
  // an empty string navigates to /library/browse to reset any active filter.
  async openSidebarOnPhoto(uid, query = "") {
    const toolbar = new Toolbar();
    if (query) {
      await t.click(toolbar.cardsViewAction);
      await toolbar.search(query);
    } else {
      await t.navigateTo("/library/browse");
      await t.click(toolbar.cardsViewAction);
    }
    await this.openPhotoViewer("uid", uid);
    await this.openSidebar();
  }

  // assertSidebarIsEditable verifies that clicking each editable sidebar row surfaces its
  // editor or dialog. Paired with assertSidebarIsReadOnly() so every selector has both a
  // positive and a negative assertion.
  async assertSidebarIsEditable() {
    // Title and Caption render unconditionally in editable mode; click each
    // to open its inline editor and escape back out before continuing.
    await this.startInlineEditOrAdd("meta-title", "Add a Title");
    await t.expect(Selector(".p-lightbox-sidebar .meta-inline-title").visible).ok();
    await t.pressKey("esc");

    await this.startInlineEditOrAdd("meta-caption", "Add a Caption");
    await t.expect(Selector(".p-lightbox-sidebar .meta-inline-caption").visible).ok();
    await t.pressKey("esc");

    await t.click(this.sidebarRow("mdi-calendar"));
    await t.expect(this.dateTimeDialog.root.visible).ok();
    await t.click(this.dateTimeDialog.cancel);

    if (await this.sidebarRow("mdi-camera").exists) {
      await t.click(this.sidebarRow("mdi-camera"));
      await t.expect(this.cameraDialog.root.visible).ok();
      await t.click(this.cameraDialog.cancel);
    }

    if (await this.sidebarRow("mdi-map-marker").exists) {
      await t.click(this.sidebarRow("mdi-map-marker"));
      await t.expect(this.locationDialog.root.visible).ok();
      await t.click(this.locationDialog.cancel);
    }

    // People, Albums, and Labels: assert their editable-only controls render.
    // The pencil toggle and chip-add inputs only exist when isEditable=true,
    // so they discriminate edit mode from read-only mode (which uses the eye
    // toggle and renders chips without an add input).
    await t.expect(this.peopleHeader.exists).ok();
    await t.expect(this.markersEditToggle.exists).ok();
    await t.expect(this.markersVisibilityToggle.exists).notOk();
    await t.expect(Selector(".p-lightbox-sidebar .v-list-item.meta-albums").exists).ok();
    await t.expect(Selector(".p-lightbox-sidebar .meta-albums .meta-inline-edit").exists).ok();
    await t.expect(Selector(".p-lightbox-sidebar .v-list-item.meta-labels").exists).ok();
    await t.expect(Selector(".p-lightbox-sidebar .meta-labels .meta-inline-edit").exists).ok();

    // The next iteration's row click cancels the previous edit, so only the trailing
    // toolbar click is needed for cleanup. Gate on .visible so rows that v-show hides
    // for empty fields are skipped.
    for (const key of ["subject", "artist", "copyright", "license", "keywords", "notes"]) {
      const row = Selector(`.p-lightbox-sidebar .v-list-item.meta-${key}`);
      if (!(await row.visible)) {
        continue;
      }
      await t.click(row);
      await t.expect(row.find(".meta-inline-edit").exists).ok();
    }
    await t.click(Selector(".p-lightbox-sidebar .v-toolbar-title"));
  }

  // assertSidebarIsReadOnly verifies the sidebar surfaces no editors or dialogs.
  // restricted=true asserts the visitor subset (Title, Caption, Date, Location).
  // expect* flags toggle positive existence checks; defaults assert presence.
  async assertSidebarIsReadOnly({
    restricted = false,
    expectTitle = true,
    expectCaption = true,
    expectPeople = true,
    expectLabels = true,
    expectAlbums = true,
  } = {}) {
    await t.click(this.sidebarRow("mdi-calendar"));
    await t.expect(this.dateTimeDialog.root.visible).notOk();

    if (await this.sidebarRow("mdi-map-marker").exists) {
      await t.click(this.sidebarRow("mdi-map-marker"));
      await t.expect(this.locationDialog.root.visible).notOk();
    }

    for (const [key, expected] of [
      ["title", expectTitle],
      ["caption", expectCaption],
    ]) {
      const display = Selector(`.p-lightbox-sidebar .meta-${key}`);
      if (expected) {
        await t.expect(display.exists).ok();
        const row = display.parent(".v-list-item");
        await t.click(row);
        await t.expect(row.find(".meta-inline-edit").exists).notOk();
      } else {
        await t.expect(display.exists).notOk();
      }
    }
    // meta-add-prompt is an editing affordance and must never render here.
    await t.expect(Selector(".p-lightbox-sidebar .meta-add-prompt").visible).notOk();

    if (restricted) {
      await t.expect(this.sidebarRow("mdi-camera").exists).notOk();
      await t.expect(this.sidebarRow("mdi-camera-iris").exists).notOk();
      await t.expect(this.peopleHeader.exists).notOk();
      await t.expect(Selector(".p-lightbox-sidebar .metadata__person-row").exists).notOk();
      await t.expect(this.markersVisibilityToggle.exists).notOk();
      await t.expect(this.markersEditToggle.exists).notOk();
      await t.expect(Selector(".p-lightbox-sidebar .meta-albums").exists).notOk();
      await t.expect(Selector(".p-lightbox-sidebar .meta-labels").exists).notOk();
      for (const key of ["subject", "artist", "copyright", "license", "keywords", "notes"]) {
        await t.expect(Selector(`.p-lightbox-sidebar .v-list-item.meta-${key}`).exists).notOk();
      }
      return;
    }

    if (await this.sidebarRow("mdi-camera").exists) {
      await t.click(this.sidebarRow("mdi-camera"));
      await t.expect(this.cameraDialog.root.visible).notOk();
    }

    // Mirror the editable assertions: section renders, but with read-only
    // controls only — eye toggle for People, no chip-add input for Albums/Labels.
    if (expectPeople) {
      await t.expect(this.peopleHeader.exists).ok();
      await t.expect(this.markersVisibilityToggle.exists).ok();
      await t.expect(this.markersEditToggle.exists).notOk();
    }
    if (expectAlbums) {
      await t.expect(Selector(".p-lightbox-sidebar .v-list-item.meta-albums").exists).ok();
      await t.expect(Selector(".p-lightbox-sidebar .meta-albums .meta-inline-edit").exists).notOk();
    }
    if (expectLabels) {
      await t.expect(Selector(".p-lightbox-sidebar .v-list-item.meta-labels").exists).ok();
      await t.expect(Selector(".p-lightbox-sidebar .meta-labels .meta-inline-edit").exists).notOk();
    }

    for (const key of ["subject", "artist", "copyright", "license", "keywords", "notes"]) {
      const row = Selector(`.p-lightbox-sidebar .v-list-item.meta-${key}`);
      if (!(await row.visible)) {
        continue;
      }
      await t.click(row);
      await t.expect(row.find(".meta-inline-edit").exists).notOk();
    }
  }

  // openSidebarDialog opens the Date / Camera / Location dialog by clicking its sidebar row.
  async openSidebarDialog(which) {
    if (which === "takenAt") {
      await t.click(this.sidebarRow("mdi-calendar"));
      await t.expect(this.dateTimeDialog.root.visible).ok();
    } else if (which === "camera") {
      await t.click(this.sidebarRow("mdi-camera"));
      await t.expect(this.cameraDialog.root.visible).ok();
    } else if (which === "location") {
      await t.click(this.sidebarRow("mdi-map-marker"));
      await t.expect(this.locationDialog.root.visible).ok();
    } else {
      throw new Error(`Unknown sidebar dialog: ${which}`);
    }
  }

  async openSidebar() {
    if (!(await this.sidebar.exists)) {
      await this.triggerPhotoViewerAction("info-button");
    }
    await t.expect(this.sidebar.visible).ok();
  }

  async startMarkersEdit() {
    await t.click(this.markersEditToggle);
  }

  async cancelMarkerDraft() {
    await t.click(this.faceMarkerCancelButton);
  }

  async confirmMarkerDraft() {
    await t.click(this.faceMarkerConfirmButton);
  }

  // removeLastUnnamedMarker deletes the most recently drawn unnamed marker via the overlay confirm pill.
  // Used by tests that draw a marker for setup and want to undo before exit so the fixture stays clean.
  async removeLastUnnamedMarker() {
    await t.expect(this.sidebar.visible).ok();
    if (await this.markersEditToggle.find('i.mdi-pencil-outline').exists) {
      await this.startMarkersEdit();
      await t.expect(this.faceMarkerOverlay.visible).ok();
    }
    await clickIfVisible(t, this.faceMarkerUnnamedRect.nth(-1));
    await t.expect(this.faceMarkerRemoveConfirm.visible).ok();
    await t.click(this.faceMarkerRemoveButton);
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
