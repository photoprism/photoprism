import { Selector, ClientFunction } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";
import Menu from "../page-model/menu";
import Toolbar from "../page-model/toolbar";
import Photo from "../page-model/photo";
import PhotoViewer from "../page-model/photoviewer";
import Page from "../page-model/page";
import PhotoEdit from "../page-model/photo-edit";
import Subject from "../page-model/subject";
import Label from "../page-model/label";
import Library from "../page-model/library";

fixture`Test Keyboard Shortcuts`.page`${testcafeconfig.url}`;

const menu = new Menu();
const toolbar = new Toolbar();
const photo = new Photo();
const photoviewer = new PhotoViewer();
const page = new Page();
const photoEdit = new PhotoEdit();
const subject = new Subject();
const label = new Label();
const library = new Library();

const triggerKeyPress = ClientFunction((key, code, keyCode, ctrlKey, shiftKey, targetSelector) => {
  const target = targetSelector ? document.querySelector(targetSelector) : document;
  if (!target) {
    console.error("Target element not found for selector:", targetSelector);
    return;
  }
  target.dispatchEvent(
    new KeyboardEvent("keydown", {
      key: key,
      code: code,
      keyCode: keyCode,
      which: keyCode,
      bubbles: true,
      cancelable: true,
      ctrlKey: ctrlKey,
      shiftKey: shiftKey,
      altKey: false,
      metaKey: false,
    })
  );
}, {});

const isFullscreen = ClientFunction(() => !!document.fullscreenElement);
const getcurrentPosition = ClientFunction(() => window.scrollY);

test.meta("testID", "shortcuts-001").meta({ type: "short", mode: "public" })(
  "Common: Test General Page Shortcuts",
  async (t) => {
    await menu.openPage("browse");
    await t.expect(toolbar.search1.focused).notOk();
    await triggerKeyPress("f", "KeyF", 70, true, false);
    await t.expect(toolbar.search1.focused).ok();

    // Test Refresh (Ctrl+R) with scroll restoration
    await t.wait(500);
    await t.scroll(0, 500); // Scroll down
    const initialScrollY = await getcurrentPosition();
    await t.expect(initialScrollY).gt(0, "Should have scrolled down before refresh");

    await triggerKeyPress("r", "KeyR", 82, true, false);
    await t.wait(2000); // Wait for page to reload

    const finalScrollY = await getcurrentPosition();
    await t.expect(finalScrollY).eql(initialScrollY, "Scroll position should be restored after refresh");

    // Test Upload (Ctrl+U)
    await triggerKeyPress("u", "KeyU", 85, true, false);
    await t.expect(Selector(".p-upload-dialog").visible).ok();
    await t.pressKey("esc");
    await t.expect(Selector(".p-upload-dialog").visible).notOk();
  }
);

test.meta("testID", "shortcuts-002").meta({ type: "short", mode: "public" })(
  "Common: Test Lightbox Shortcuts",
  async (t) => {
    await menu.openPage("browse");
    await t.navigateTo("/library/videos");
    const videoUid = await photo.getNthPhotoUid("all", 0);
    await photoviewer.openPhotoViewer("uid", videoUid);

    await t.wait(500);
    const infoPanelSelector = Selector("div").withText("Information").nth(4);
    await t.expect(infoPanelSelector.visible).notOk("Information panel should not be visible initially");

    await triggerKeyPress("i", "KeyI", 73, true, false, "div.p-lightbox__pswp");
    await t.expect(infoPanelSelector.visible).ok("Information panel should be visible after first Ctrl+I");

    await triggerKeyPress("i", "KeyI", 73, true, false, "div.p-lightbox__pswp");
    await t.expect(infoPanelSelector.visible).notOk("Information panel should be hidden after second Ctrl+I");

    await triggerKeyPress("m", "KeyM", 77, true, false, "div.p-lightbox__pswp");
    await t
      .expect(Selector(".p-lightbox__content").hasClass("is-muted"))
      .ok("Video should be muted after first Ctrl+M");

    await triggerKeyPress("m", "KeyM", 77, true, false, "div.p-lightbox__pswp");
    await t
      .expect(Selector(".p-lightbox__content").hasClass("is-muted"))
      .notOk("Video should be unmuted after second Ctrl+M");

    await triggerKeyPress("s", "KeyS", 83, true, false, "div.p-lightbox__pswp");
    await t
      .expect(Selector(".p-lightbox__content").hasClass("slideshow-active"))
      .ok("Slideshow should be active after first Ctrl+S");

    await triggerKeyPress("s", "KeyS", 83, true, false, "div.p-lightbox__pswp");
    await t
      .expect(Selector(".p-lightbox__content").hasClass("slideshow-active"))
      .notOk("Slideshow should be inactive after second Ctrl+S");

    await triggerKeyPress("Escape", "Escape", 27, false, false, "div.p-lightbox__pswp");
  }
);

test.meta("testID", "shortcuts-003").meta({ type: "short", mode: "public" })(
  "Common: Test Lightbox Archive and Download Shortcuts",
  async (t) => {
    await menu.openPage("browse");
    const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
    await photoviewer.openPhotoViewer("uid", FirstPhotoUid);

    await t.expect(photoviewer.viewer.visible).ok();

    await triggerKeyPress("a", "KeyA", 65, true, false);

    await t.expect(Selector("div.p-notify--success").withText("Archived").visible).ok();

    await t.wait(5000);

    await triggerKeyPress("a", "KeyA", 65, true, false);

    await t.expect(Selector("div.p-notify--success").withText("Restored").visible).ok();

    await t.wait(5000);

    await triggerKeyPress("d", "KeyD", 68, true, false);
    await t.expect(Selector("div.p-notify--success").withText("Downloading").visible).ok();
    await t.pressKey("esc");
  }
);

test.meta("testID", "shortcuts-004").meta({ type: "short", mode: "public" })(
  "Common: Test Lightbox Edit, Fullscreen, and Like Shortcuts",
  async (t) => {
    await menu.openPage("browse");
    const FirstPhotoUid = await photo.getNthPhotoUid("image", 0);
    await photoviewer.openPhotoViewer("uid", FirstPhotoUid);
    await t.wait(500); // Wait for lightbox

    // Edit Test
    await triggerKeyPress("e", "KeyE", 69, true, false);
    await t.expect(photoEdit.dialog.visible).ok();
    await t.click(photoEdit.dialogClose);

    await photoviewer.openPhotoViewer("uid", FirstPhotoUid);
    await t.wait(500); // Wait for lightbox again

    // Fullscreen Test
    await triggerKeyPress("f", "KeyF", 70, true, false, "div.p-lightbox__pswp");
    await t.wait(1000);
    await t.expect(isFullscreen()).ok("Browser did not enter fullscreen mode.");

    await triggerKeyPress("f", "KeyF", 70, true, false, "div.p-lightbox__pswp");
    await t.wait(1000);
    await t.expect(isFullscreen()).notOk("Browser did not exit fullscreen mode.");

    // Like Test
    const isLikedInitially = await Selector(".p-lightbox__content").hasClass("is-favorite");
    await triggerKeyPress("l", "KeyL", 76, true, false, "div.p-lightbox__pswp");
    await t.wait(2000); // Wait for potential UI updates
    await t.expect(photoviewer.menuButton.exists).ok("Menu button does not exist after Ctrl+L");
    const isLikedAfterToggle = await Selector(".p-lightbox__content").hasClass("is-favorite");
    if (isLikedInitially) {
      await t.expect(isLikedAfterToggle).notOk("Failed to unlike photo after Ctrl+L");
    } else {
      await t.expect(isLikedAfterToggle).ok("Failed to like photo after Ctrl+L");
    }

    await t.pressKey("esc");
  }
);

test.meta("testID", "shortcuts-005").meta({ type: "short", mode: "public" })(
  "Common: Test Expansion Panel and Page-Specific Search Focus",
  async (t) => {
    await menu.openPage("browse");
    await t.wait(500);
    await triggerKeyPress("f", "KeyF", 70, true, true);
    await t.wait(500);
    await t
      .expect(Selector(".toolbar-expansion-panel").find("div").withText("All Countries").exists)
      .ok("Expansion panel content ('All Countries') not found after Shift+Ctrl+F");
    // Close expansion panel
    // TODO: Currently no close functionality implemented
    // await triggerKeyPress('f', 'KeyF', 70, true, true); // Toggle it off
    // await t.wait(500); // Wait for animation
    // await t.expect(Selector(".toolbar-expansion-panel").getAttribute("style")).contains("display: none", "Expansion panel is not hidden (display: none) after second Shift+Ctrl+F");

    await menu.openPage("people");
    await t.wait(500);
    await triggerKeyPress("f", "KeyF", 70, true, false);
    await t.expect(subject.search.focused).ok();

    await menu.openPage("labels");
    await t.wait(500);
    await triggerKeyPress("f", "KeyF", 70, true, false);
    await t.expect(label.search.focused).ok();

    await t.navigateTo("/library/errors");
    await t.wait(500);
    await triggerKeyPress("f", "KeyF", 70, true, false);
    await t.expect(library.searchInput.focused).ok();
  }
);
