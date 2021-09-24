import { Selector } from "testcafe";
import testcafeconfig from "./testcafeconfig";
import Page from "./page-model";

fixture`Test people`.page`${testcafeconfig.url}`;

const page = new Page();

test.meta("testID", "authentication-000")(
  "Time to start instance (will be marked as unstable)",
  async (t) => {
    await t.wait(5000);
  }
);

test.meta("testID", "people-001")("Faces tab preselected when subjects empty", async (t) => {
  await page.openNav();
  await t
    .click(Selector(".nav-people"))
    .expect(Selector("#tab-people-faces > a").hasClass("v-tabs__item--active"))
    .ok()
    .expect(Selector("#tab-people-subjects > a").hasClass("v-tabs__item--active"))
    .notOk()
    .click(Selector("#tab-people-subjects > a"))
    .expect(Selector("#tab-people-faces > a").hasClass("v-tabs__item--active"))
    .notOk()
    .expect(Selector("#tab-people-subjects > a").hasClass("v-tabs__item--active"))
    .ok();
  const countSubjects = await Selector("a.is-subject").count;
  await t.expect(countSubjects).eql(0);
});

test.meta("testID", "people-002")("Add + Rename", async (t) => {
  await page.openNav();
  await t
    .click(Selector(".nav-people"))
    .click(Selector("#tab-people-faces > a"))
    .click(Selector("form.p-faces-search button.action-reload"));
  const countFaces = await Selector("div.is-face", { timeout: 55000 }).count;
  await t.click(Selector("#tab-people-subjects > a"));
  const countSubjects = await Selector("a.is-subject").count;
  await t.click(Selector("#tab-people-faces > a"));
  const FirstFaceID = await Selector("div.is-face").nth(0).getAttribute("data-id");
  await t.click(Selector("div[data-id=" + FirstFaceID + "] div.clickable"));
  const countPhotosFace = await Selector("div.is-photo").count;
  await page.openNav();
  await t
    .click(Selector(".nav-people"))
    .click(Selector("#tab-people-faces > a"))
    .typeText(Selector("div[data-id=" + FirstFaceID + "] div.input-name input"), "Jane Doe")
    .pressKey("enter")
    .click(Selector("form.p-faces-search button"));
  const countFacesAfterAdd = await Selector("div.is-face", { timeout: 55000 }).count;
  await t
    .expect(countFacesAfterAdd)
    .eql(countFaces - 1)
    .expect(Selector("div").withAttribute("data-id", FirstFaceID).exists)
    .notOk()
    .click(Selector("#tab-people-subjects > a"));
  const countSubjectsAfterAdd = await Selector("a.is-subject").count;
  await t
    .expect(countSubjectsAfterAdd)
    .eql(countSubjects + 1)
    .typeText(Selector("div.input-search input"), "Jane")
    .pressKey("enter");
  const JaneUID = await Selector("a.is-subject", { timeout: 55000 })
    .nth(0)
    .getAttribute("data-uid");
  await t
    .expect(Selector("a[data-uid=" + JaneUID + "] div.caption").innerText)
    .contains(countPhotosFace.toString())
    .click(Selector("a.is-subject").withAttribute("data-uid", JaneUID));
  const countPhotosSubject = await Selector("div.is-photo").count;
  await t.expect(countPhotosFace).eql(countPhotosSubject);
  await page.toggleSelectNthPhoto(0);
  await page.toggleSelectNthPhoto(1);
  await page.toggleSelectNthPhoto(2);
  await page.editSelected();
  await t
    .click(Selector("#tab-people"))
    .expect(Selector("div.input-name input").nth(0).value)
    .contains("Jane Doe")
    .typeText(Selector("div.input-name input").nth(0), "Max Mu", { replace: true })
    .pressKey("enter")
    .expect(Selector("div.input-name input").nth(0).value)
    .contains("Max Mu")
    .click("button.action-next")
    .expect(Selector("div.input-name input").nth(0).value)
    .contains("Max Mu")
    .click("button.action-next")
    .expect(Selector("div.input-name input").nth(0).value)
    .contains("Max Mu")
    .click("button.action-close");
  await page.clearSelection();
  await t
    .typeText(Selector("div.input-search input").nth(0), "person:max-mu", {
      replace: true,
    })
    .pressKey("enter");
  const countPhotosSubjectAfterRename = await Selector("div.is-photo").count;
  await t.expect(countPhotosSubjectAfterRename).eql(countPhotosSubject);
  await page.openNav();
  await t
    .click(Selector(".nav-people"))
    .expect(Selector("a[data-uid=" + JaneUID + "] div.v-card__title").innerText)
    .contains("Max Mu")
    .click(Selector("a[data-uid=" + JaneUID + "] div.v-card__title"))
    .typeText(Selector("div.input-rename input"), "Jane Mu", { replace: true })
    .pressKey("enter")
    .expect(Selector("a[data-uid=" + JaneUID + "] div.v-card__title").innerText)
    .contains("Jane Mu")
    .click(Selector("a.is-subject").withAttribute("data-uid", JaneUID))
    .expect(Selector("div.input-search input").value)
    .contains("person:jane-mu");
  await page.toggleSelectNthPhoto(0);
  await page.toggleSelectNthPhoto(1);
  await page.toggleSelectNthPhoto(2);
  await page.editSelected();
  await t
    .click(Selector("#tab-people"))
    .expect(Selector("div.input-name input").nth(0).value)
    .contains("Jane Mu")
    .click("button.action-next")
    .expect(Selector("div.input-name input").nth(0).value)
    .contains("Jane Mu")
    .click("button.action-next")
    .expect(Selector("div.input-name input").nth(0).value)
    .contains("Jane Mu")
    .click("button.action-close");
  await page.clearSelection();
});

test.meta("testID", "people-003")("Add + Reject + Star", async (t) => {
  await page.openNav();
  await t
    .click(Selector(".nav-people"))
    .click(Selector("#tab-people-faces > a"))
    .click(Selector("form.p-faces-search button.action-reload"));
  const FirstFaceID = await Selector("div.is-face").nth(0).getAttribute("data-id");
  await t
    .expect(Selector("div.menuable__content__active").nth(0).visible)
    .notOk()
    .click(Selector("div[data-id=" + FirstFaceID + "] div.input-name input"))
    .expect(Selector("div.menuable__content__active").nth(0).visible)
    .ok()
    .typeText(Selector("div[data-id=" + FirstFaceID + "] div.input-name input"), "Andrea Doe")
    .pressKey("enter")
    .click(Selector("#tab-people-subjects > a"));
  await t.typeText(Selector("div.input-search input"), "Andrea").pressKey("enter");
  const AndreaUID = await Selector("a.is-subject").nth(0).getAttribute("data-uid");
  await page.openNav();
  await t.click(Selector(".nav-browse"));
  await page.search("face:new filmpreis");
  await page.toggleSelectNthPhoto(0);
  await page.editSelected();
  await t
    .click(Selector("#tab-people"))
    .expect(Selector("div.input-name input").nth(0).value)
    .eql("")
    .typeText(Selector("div.input-name input").nth(0), "Andrea Doe", { replace: true })
    .click(Selector("div").withText("Andrea Doe"))
    .expect(Selector("div.input-name input").nth(0).value)
    .contains("Andrea Doe")
    .click("button.action-close");
  await page.clearSelection();
  await page.openNav();
  await t
    .click(Selector(".nav-people"))
    .click(Selector("a.is-subject").withAttribute("data-uid", AndreaUID));
  const countPhotosAndreaAfterAdd = await Selector("div.is-photo").count;
  await page.toggleSelectNthPhoto(1);
  await page.editSelected();
  await t
    .click(Selector("#tab-people"))
    .expect(Selector("div.input-name input").nth(0).value)
    .eql("Andrea Doe")
    .click(Selector("div.input-name div.v-input__icon--clear"))
    .expect(Selector("div.input-name input").nth(0).value)
    .eql("")
    .typeText(Selector("div.input-name input").nth(0), "Nicole", { replace: true })
    .pressKey("enter")
    .click("button.action-close");
  await page.clearSelection();
  await t.eval(() => location.reload());
  await t.wait(6000);
  const countPhotosAndreaAfterReject = await Selector("div.is-photo").count;
  const Diff = countPhotosAndreaAfterAdd - countPhotosAndreaAfterReject;
  await t
    .typeText(Selector("div.input-search input"), "Nicole", { replace: true })
    .pressKey("enter");
  const countPhotosNicole = await Selector("div.is-photo").count;
  await t.expect(countPhotosNicole).eql(Diff);
  await page.openNav();
  await t
    .click(Selector(".nav-people"))
    .expect(Selector("a.is-subject").nth(0).getAttribute("data-uid"))
    .eql(AndreaUID);
  await t
    .typeText(Selector("div.input-search input"), "Nicole", { replace: true })
    .pressKey("enter");
  const NicoleUID = await Selector("a.is-subject").nth(0).getAttribute("data-uid");
  await t
    .click(Selector("a[data-uid=" + NicoleUID + "] button.input-favorite"))
    .typeText(Selector("div.input-search input"), " ", { replace: true })
    .pressKey("enter")
    .expect(Selector("a.is-subject").nth(0).getAttribute("data-uid"))
    .eql(NicoleUID);
});

test.meta("testID", "people-004")("Remove face", async (t) => {
  await page.openNav();
  await t.click(Selector(".nav-browse"));
  await page.search("face:new");
  const FirstPhoto = await Selector("div.is-photo").nth(0).getAttribute("data-uid");
  await page.toggleSelectNthPhoto(0);
  await page.editSelected();
  await t.click(Selector("#tab-people"));
  const MarkerCount = await Selector("div.is-marker").count;
  if ((await Selector("div.input-name input").nth(0).value) == "") {
    await t
      .expect(Selector("button.action-undo").nth(0).visible)
      .notOk()
      .expect(Selector("div.input-name input").nth(0).value)
      .eql("")
      .click(Selector("button.input-reject"))
      .expect(Selector("button.action-undo").nth(0).visible)
      .ok()
      .click(Selector("button.action-undo"));
  } else if ((await Selector("div.input-name input").nth(0).value) != "") {
    await t
      .expect(Selector("div.input-name input").nth(1).value)
      .eql("")
      .click(Selector("button.input-reject"))
      .expect(Selector("button.action-undo").nth(0).visible)
      .ok()
      .click(Selector("button.action-undo"));
  }
  await t.click("button.action-close");
  await page.clearSelection();
  await t.eval(() => location.reload());
  await t.wait(6000);
  await page.selectPhotoFromUID(FirstPhoto);
  await page.editSelected();
  await t.click(Selector("#tab-people"));
  if ((await Selector("div.input-name input").nth(0).value) == "") {
    await t
      .expect(Selector("button.action-undo").nth(0).visible)
      .notOk()
      .expect(Selector("div.input-name input").nth(0).value)
      .eql("")
      .click(Selector("button.input-reject"))
      .expect(Selector("button.action-undo").nth(0).visible)
      .ok();
  } else if ((await Selector("div.input-name input").nth(0).value) != "") {
    await t
      .expect(Selector("button.action-undo").nth(0).visible)
      .notOk()
      .expect(Selector("div.input-name input").nth(1).value)
      .eql("")
      .click(Selector("button.input-reject"))
      .expect(Selector("button.action-undo").nth(0).visible)
      .ok();
  }
  await t.click("button.action-close");
  await t.eval(() => location.reload());
  await page.editSelected();
  await t.click(Selector("#tab-people"));
  const MarkerCountAfterRemove = await Selector("div.is-marker").count;
  await t.expect(MarkerCountAfterRemove).eql(MarkerCount - 1);
});

test.meta("testID", "people-005")("Hide face", async (t) => {
  await page.openNav();
  await t
    .click(Selector(".nav-people"))
    .click(Selector("#tab-people-faces > a"))
    .click(Selector("form.p-faces-search button.action-reload"));
  const FirstFaceID = await Selector("div.is-face").nth(0).getAttribute("data-id");
  await t.click(Selector("div[data-id=" + FirstFaceID + "] button.input-hide"));
  await t.eval(() => location.reload());
  await t
    .wait(6000)
    .expect(Selector("div[data-id=" + FirstFaceID + "]").visible)
    .notOk();
});
