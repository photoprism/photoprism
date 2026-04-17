import { ClientFunction, Selector } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";
import Menu from "../page-model/menu";
import Photo from "../page-model/photo";
import PhotoEdit from "../page-model/photo-edit";
import PhotoViewer from "../page-model/photoviewer";
import Subject from "../page-model/subject";
import Notifies from "../page-model/notifications";

fixture`Test face markers`.page`${testcafeconfig.url}`;

const menu = new Menu();
const photo = new Photo();
const photoedit = new PhotoEdit();
const photoviewer = new PhotoViewer();
const subject = new Subject();
const notifies = new Notifies();

const drawFaceMarker = ClientFunction(() => {
  const overlay = document.querySelector("div.p-face-markers");
  const image = document.querySelector("img.pswp__img, img.pswp__image");

  if (!overlay || !image || typeof PointerEvent === "undefined") {
    return false;
  }

  const rect = image.getBoundingClientRect();
  const side = Math.max(48, Math.min(rect.width, rect.height) * 0.18);
  const startX = rect.left + rect.width * 0.18;
  const startY = rect.top + rect.height * 0.18;
  const endX = Math.min(rect.right - 24, startX + side);
  const endY = Math.min(rect.bottom - 24, startY + side);

  const dispatch = (type, x, y) =>
    overlay.dispatchEvent(
      new PointerEvent(type, {
        bubbles: true,
        cancelable: true,
        pointerId: 1,
        pointerType: "mouse",
        button: 0,
        buttons: type === "pointerup" ? 0 : 1,
        clientX: x,
        clientY: y,
      })
    );

  dispatch("pointerdown", startX, startY);
  dispatch("pointermove", endX, endY);
  dispatch("pointerup", endX, endY);

  return true;
}, {});

const hasInputValue = ClientFunction((selector, value) => {
  const inputs = Array.from(document.querySelectorAll(selector));

  return inputs.some((input) => input.value === value);
}, {});

test.meta("testID", "photos-face-markers-001").meta({ type: "short", mode: "public" })(
  "Common: Add and assign a manual face marker in fullscreen mode",
  async (t) => {
    const uniqueName = `Manual Face ${Date.now()}`;

    await menu.openPage("people");
    await subject.openSubjectWithUid(await subject.getNthSubjectUid(0));
    await photoviewer.openPhotoViewer("uid", await photo.getNthPhotoUid("all", 0));
    await photoviewer.openInfoPanel();

    await t.expect(Selector(".metadata__marker").count).gte(1, "", { timeout: 10000 });

    const initialMarkerCount = await Selector(".metadata__marker").count;

    await t.click(photoviewer.showMarkersButton);
    await t.expect(photoviewer.faceMarkerRect.count).gte(initialMarkerCount, "", { timeout: 10000 });

    await t.click(photoviewer.addFaceButton);
    await t.expect(photoviewer.faceMarkerOverlay.visible).ok();
    await t.expect(drawFaceMarker()).ok();
    await t.expect(Selector(".p-face-markers__btn--confirm").visible).ok();
    await t.click(Selector(".p-face-markers__btn--confirm"));
    await notifies.waitForSpecficEvent("Changes successfully saved", 10000, true);
    await t.expect(Selector(".metadata__marker").count).eql(initialMarkerCount + 1, "", { timeout: 10000 });

    const newMarkerInput = Selector(".metadata__marker--new input").nth(0);
    await t.typeText(newMarkerInput, uniqueName, { replace: true }).pressKey("enter");
    await t.expect(Selector(".p-confirm-dialog").visible).ok();
    await t.click(Selector(".p-confirm-dialog button.action-confirm"));
    await t.expect(hasInputValue(".metadata__marker--new input", uniqueName)).ok("", { timeout: 10000 });

    await photoviewer.triggerPhotoViewerAction("edit-button");
    await t.expect(photoedit.dialog.visible).ok();
    await t.click(photoedit.peopleTab);
    await t.expect(hasInputValue("div.input-name input", uniqueName)).ok("", { timeout: 10000 });
    await t.click(photoedit.dialogClose);
  }
);
