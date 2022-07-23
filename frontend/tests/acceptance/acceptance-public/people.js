import { Selector } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";
import Menu from "../page-model/menu";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import Photo from "../page-model/photo";
import Subject from "../page-model/subject";
import PhotoEdit from "../page-model/photo-edit";

fixture`Test people`.page`${testcafeconfig.url}`;

const menu = new Menu();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const photo = new Photo();
const subject = new Subject();
const photoedit = new PhotoEdit();

test.meta("testID", "people-001").meta({ type: "short", mode: "public" })(
  "Common: Add name to new face and rename subject",
  async (t) => {
    await menu.openPage("people");
    await t.click(subject.newTab);
    await subject.triggerToolbarAction("reload", "");
    const FaceCount = await subject.getFaceCount();

    await t.click(subject.recognizedTab);

    const SubjectCount = await subject.getSubjectCount();
    await t.click(subject.newTab);
    const FirstFaceID = await subject.getNthFaceUid(0);
    await subject.openFaceWithUid(FirstFaceID);
    const PhotosInFaceCount = await photo.getPhotoCount("all");
    await menu.openPage("people");
    await t.click(subject.newTab);
    await subject.addNameToFace(FirstFaceID, "Jane Doe");
    await subject.triggerToolbarAction("reload");
    const FaceCountAfterAdd = await subject.getFaceCount();

    await t.expect(FaceCountAfterAdd).eql(FaceCount - 1);

    await t.click(subject.recognizedTab);
    await subject.checkFaceVisibility(FirstFaceID, false);
    await t.eval(() => location.reload());
    await t.wait(6000);
    const SubjectCountAfterAdd = await subject.getSubjectCount();

    await t.expect(SubjectCountAfterAdd).eql(SubjectCount + 1);

    await toolbar.search("Jane");
    const JaneUID = await subject.getNthSubjectUid(0);

    await t
      .expect(Selector("a[data-uid=" + JaneUID + "] div.caption").innerText)
      .contains(PhotosInFaceCount.toString());

    await subject.openSubjectWithUid(JaneUID);
    const PhotosInSubjectCount = await photo.getPhotoCount("all");

    await t.expect(PhotosInFaceCount).eql(PhotosInSubjectCount);

    await photo.triggerHoverAction("nth", 0, "select");
    await photo.triggerHoverAction("nth", 1, "select");
    await photo.triggerHoverAction("nth", 2, "select");
    await contextmenu.triggerContextMenuAction("edit", "");
    await t.click(photoedit.peopleTab);

    await t.expect(photoedit.inputName.nth(0).value).contains("Jane Doe");

    await t.click(photoedit.dialogClose);
    await menu.openPage("people");
    await subject.renameSubject(JaneUID, "Max Mu");

    await t
      .expect(Selector("a[data-uid=" + JaneUID + "] div.v-card__title").innerText)
      .contains("Max Mu");

    await subject.openSubjectWithUid(JaneUID);
    await t.eval(() => location.reload());
    await contextmenu.checkContextMenuCount("3");
    await contextmenu.triggerContextMenuAction("edit", "");
    await t.click(photoedit.peopleTab);

    await t.expect(photoedit.inputName.nth(0).value).contains("Max Mu");

    await t.click(photoedit.dialogNext);

    await t.expect(photoedit.inputName.nth(0).value).contains("Max Mu").click(photoedit.dialogNext);
    await t
      .expect(photoedit.inputName.nth(0).value)
      .contains("Max Mu")
      .click(photoedit.dialogClose);

    await contextmenu.clearSelection();
    await toolbar.search("person:max-mu");
    const PhotosInSubjectAfterRenameCount = await photo.getPhotoCount("all");
    await t.expect(PhotosInSubjectAfterRenameCount).eql(PhotosInSubjectCount);
  }
);

test.meta("testID", "people-002").meta({ type: "short", mode: "public" })(
  "Common: Add + Reject name on people tab",
  async (t) => {
    await menu.openPage("people");
    await t.click(subject.newTab);
    await subject.triggerToolbarAction("reload");
    const FirstFaceID = await subject.getNthFaceUid(0);
    await subject.addNameToFace(FirstFaceID, "Andrea Doe");
    await t.click(subject.recognizedTab);
    await toolbar.search("Andrea");
    const AndreaUID = await subject.getNthSubjectUid(0);
    await subject.openSubjectWithUid(AndreaUID);
    await t.eval(() => location.reload());
    await t.wait(5000);
    const PhotosInAndreaCount = await photo.getPhotoCount("all");
    await photo.triggerHoverAction("nth", 1, "select");
    await contextmenu.triggerContextMenuAction("edit", "");
    await t
      .click(photoedit.peopleTab)
      .expect(photoedit.inputName.nth(0).value)
      .eql("Andrea Doe")
      .click(photoedit.rejectName.nth(0));

    await t.expect(photoedit.inputName.nth(0).value).eql("");

    await t
      .typeText(photoedit.inputName.nth(0), "Nicole", { replace: true })
      .pressKey("enter")
      .click(photoedit.dialogClose);
    await contextmenu.clearSelection();
    await t.eval(() => location.reload());
    await t.wait(5000);
    const PhotosInAndreaAfterRejectCount = await photo.getPhotoCount("all");
    const Diff = PhotosInAndreaCount - PhotosInAndreaAfterRejectCount;
    await toolbar.search("person:nicole");
    await t.eval(() => location.reload());
    await t.wait(5000);
    const PhotosInNicoleCount = await photo.getPhotoCount("all");

    await t.expect(Diff).gte(PhotosInNicoleCount);
  }
);

test.meta("testID", "people-003").meta({ mode: "public" })(
  "Common: Test mark subject as favorite",
  async (t) => {
    await menu.openPage("people");
    const FirstSubjectUid = await subject.getNthSubjectUid(0);
    const SecondSubjectUid = await subject.getNthSubjectUid(1);
    await subject.triggerHoverAction("uid", SecondSubjectUid, "favorite");
    await subject.triggerToolbarAction("reload");
    const FirstSubjectUidAfterFavorite = await subject.getNthSubjectUid(0);

    await t.expect(FirstSubjectUid).notEql(FirstSubjectUidAfterFavorite);
    await t.expect(SecondSubjectUid).eql(FirstSubjectUidAfterFavorite);

    await subject.checkHoverActionState("uid", SecondSubjectUid, "favorite", true);
    await subject.triggerHoverAction("uid", SecondSubjectUid, "favorite");
    await subject.checkHoverActionState("uid", SecondSubjectUid, "favorite", false);
  }
);

test.meta("testID", "people-004").meta({ mode: "public" })(
  "Common: Test new face autocomplete",
  async (t) => {
    await menu.openPage("people");
    await t.click(subject.newTab);
    await subject.triggerToolbarAction("reload");
    const FirstFaceID = await subject.getNthFaceUid(0);
    await t
      .expect(Selector("div.menuable__content__active").nth(0).visible)
      .notOk()
      .click(Selector("div[data-id=" + FirstFaceID + "] div.input-name input"))
      .typeText(Selector("div[data-id=" + FirstFaceID + "] div.input-name input"), "Otto");

    await t
      .expect(Selector("div.menuable__content__active").nth(0).withText("Otto Visible").visible)
      .ok();
  }
);

test.meta("testID", "people-005").meta({ mode: "public" })("Common: Remove face", async (t) => {
  await toolbar.search("face:new");
  const FirstPhotoUid = await photo.getNthPhotoUid("all", 0);
  await photo.triggerHoverAction("nth", 0, "select");
  await contextmenu.triggerContextMenuAction("edit", "");
  await t.click(photoedit.peopleTab);
  const MarkerCount = await subject.getMarkerCount();

  if ((await photoedit.inputName.nth(0).value) == "") {
    await t
      .expect(photoedit.undoRemoveMarker.nth(0).visible)
      .notOk()
      .expect(photoedit.inputName.nth(0).value)
      .eql("")
      .click(photoedit.removeMarker)
      .expect(photoedit.undoRemoveMarker.nth(0).visible)
      .ok()
      .click(photoedit.undoRemoveMarker);
  } else if ((await photoedit.inputName.nth(0).value) != "") {
    await t
      .expect(photoedit.inputName.nth(1).value)
      .eql("")
      .click(photoedit.removeMarker)
      .expect(photoedit.undoRemoveMarker.nth(0).visible)
      .ok()
      .click(photoedit.undoRemoveMarker);
  }

  await t.click(photoedit.dialogClose);
  await contextmenu.clearSelection();
  await t.eval(() => location.reload());
  await t.wait(5000);
  await photo.triggerHoverAction("uid", FirstPhotoUid, "select");
  await contextmenu.triggerContextMenuAction("edit", "");
  await t.click(photoedit.peopleTab);

  if ((await photoedit.inputName.nth(0).value) == "") {
    await t
      .expect(photoedit.undoRemoveMarker.nth(0).visible)
      .notOk()
      .expect(photoedit.inputName.nth(0).value)
      .eql("")
      .click(photoedit.removeMarker)
      .expect(photoedit.undoRemoveMarker.nth(0).visible)
      .ok();
  } else if ((await photoedit.inputName.nth(0).value) != "") {
    await t
      .expect(photoedit.undoRemoveMarker.nth(0).visible)
      .notOk()
      .expect(photoedit.inputName.nth(1).value)
      .eql("")
      .click(photoedit.removeMarker)
      .expect(photoedit.undoRemoveMarker.nth(0).visible)
      .ok();
  }

  await t.click(photoedit.dialogClose);
  await t.eval(() => location.reload());
  await contextmenu.triggerContextMenuAction("edit", "");
  await t.click(photoedit.peopleTab);
  const MarkerCountAfterRemove = await subject.getMarkerCount();

  await t.expect(MarkerCountAfterRemove).eql(MarkerCount - 1);
});

test.meta("testID", "people-006").meta({ mode: "public" })("Common: Hide face", async (t) => {
  await menu.openPage("people");
  await t.click(subject.newTab);
  await subject.triggerToolbarAction("reload");
  const FirstFaceID = await subject.getNthFaceUid(0);
  await subject.checkFaceVisibility(FirstFaceID, true);
  await subject.triggerHoverAction("id", FirstFaceID, "hidden");
  await t.eval(() => location.reload());
  await t.wait(5000);
  await subject.checkFaceVisibility(FirstFaceID, false);
  await subject.triggerToolbarAction("show-hidden");
  await t.eval(() => location.reload());
  await t.wait(6000);
  await subject.checkFaceVisibility(FirstFaceID, true);
  await subject.triggerHoverAction("id", FirstFaceID, "hidden");
  await subject.triggerToolbarAction("exclude-hidden");
  await t.eval(() => location.reload());
  await t.wait(6000);
  await subject.checkFaceVisibility(FirstFaceID, true);
});

test.meta("testID", "people-007").meta({ mode: "public" })("Common: Hide person", async (t) => {
  await menu.openPage("people");
  await t.click(subject.recognizedTab);
  const FirstPersonUid = await subject.getNthSubjectUid(0);
  await subject.checkSubjectVisibility("uid", FirstPersonUid, true);
  await subject.triggerHoverAction("uid", FirstPersonUid, "hidden");
  await t.eval(() => location.reload());
  await t.wait(6000);
  await subject.checkSubjectVisibility("uid", FirstPersonUid, false);
  await subject.triggerToolbarAction("show-hidden");
  await t.eval(() => location.reload());
  await t.wait(6000);
  await subject.checkSubjectVisibility("uid", FirstPersonUid, true);
  await subject.triggerHoverAction("uid", FirstPersonUid, "hidden");
  await subject.triggerToolbarAction("exclude-hidden");
  await t.eval(() => location.reload());
  await t.wait(5000);
  await subject.checkSubjectVisibility("uid", FirstPersonUid, true);
});
