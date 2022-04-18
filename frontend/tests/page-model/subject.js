import { Selector, t } from "testcafe";

export default class Page {
  constructor() {
    this.recognizedTab = Selector("#tab-people > a", { timeout: 15000 });
    this.newTab = Selector("#tab-people_faces > a", { timeout: 15000 });
    this.showAllNewButton = Selector('a[href="/all?q=face%3Anew"]');
    this.subjectName = Selector("a.is-subject div.v-card__title");
  }

  async addNameToFace(id, name) {
    await t
      .typeText(Selector("div[data-id=" + id + "] div.input-name input"), name)
      .pressKey("enter");
  }

  async renameSubject(uid, name) {
    await t
      .click(Selector("a[data-uid=" + uid + "] div.v-card__title"))
      .typeText(Selector("div.input-rename input"), name, { replace: true })
      .pressKey("enter");
  }

  async getNthSubjectUid(nth) {
    const NthSubject = await Selector("a.is-subject").nth(nth).getAttribute("data-uid");
    return NthSubject;
  }

  async getNthFaceUid(nth) {
    const NthFace = await Selector("div.is-face").nth(nth).getAttribute("data-id");
    return NthFace;
  }

  async getSubjectCount() {
    const SubjectCount = await Selector("a.is-subject", { timeout: 5000 }).count;
    return SubjectCount;
  }

  async getFaceCount() {
    const FaceCount = await Selector("div.is-face", { timeout: 5000 }).count;
    return FaceCount;
  }

  async getMarkerCount() {
    const MarkerCount = await Selector("div.is-marker", { timeout: 5000 }).count;
    return MarkerCount;
  }

  async selectSubjectFromUID(uid) {
    await t
      .hover(Selector("a.is-subject").withAttribute("data-uid", uid))
      .click(Selector(`.uid-${uid} .input-select`));
  }

  async toggleSelectNthSubject(nth) {
    await t
      .hover(Selector("a.is-subject", { timeout: 4000 }).nth(nth))
      .click(Selector("a.is-subject .input-select").nth(nth));
  }

  async openNthSubject(nth) {
    await t.click(Selector("a.is-subject").nth(nth)).expect(Selector("div.is-photo").visible).ok();
  }

  async openSubjectWithUid(uid) {
    await t.click(Selector("a.is-subject").withAttribute("data-uid", uid));
  }

  async openFaceWithUid(uid) {
    await t.click(Selector("div[data-id=" + uid + "] div.clickable"));
  }

  async checkSubjectVisibility(mode, uidOrName, visible) {
    if (visible) {
      if (mode === "uid") {
        await t.expect(Selector("a").withAttribute("data-uid", uidOrName).visible).ok();
      } else {
        await t.expect(Selector("a div.v-card__title").withText(uidOrName).visible).ok();
      }
    } else if (!visible) {
      if (mode === "uid") {
        await t.expect(Selector("a").withAttribute("data-uid", uidOrName).visible).notOk();
      } else {
        await t.expect(Selector("a div.v-card__title").withText(uidOrName).visible).notOk();
      }
    }
  }

  async checkFaceVisibility(uid, visible) {
    if (visible) {
      await t.expect(Selector("div.is-face").withAttribute("data-id", uid).visible).ok();
    } else {
      await t.expect(Selector("div.is-face").withAttribute("data-id", uid).visible).notOk();
    }
  }

  async triggerToolbarAction(action) {
    if (await Selector("form.p-faces-search button.action-" + action).visible) {
      await t.click(Selector("form.p-faces-search button.action-" + action));
    } else if (await Selector("form.p-people-search button.action-" + action).visible) {
      await t.click(Selector("form.p-people-search button.action-" + action));
    }
  }

  async triggerHoverAction(mode, uidOrNth, action) {
    if (mode === "uid") {
      await t.hover(Selector("a.uid-" + uidOrNth));
      await t.click(Selector("a.uid-" + uidOrNth + " .input-" + action));
    }
    if (mode === "nth") {
      await t.hover(Selector("a.is-subject").nth(uidOrNth));
      await t.click(Selector(`.input-` + action));
    }
    if (mode === "id") {
      await t
        .hover(Selector("div[data-id=" + uidOrNth + "]"))
        .click(Selector("div[data-id=" + uidOrNth + "] button.input-" + action));
    }
  }

  async checkHoverActionAvailability(mode, uidOrNth, action, visible) {
    if (mode === "uid") {
      await t.hover(Selector("a.is-subject").withAttribute("data-uid", uidOrNth));
      if (visible) {
        await t.expect(Selector(`.uid-${uidOrNth} .input-` + action).visible).ok();
      } else {
        await t.expect(Selector(`.uid-${uidOrNth} .input-` + action).visible).notOk();
      }
    }
    if (mode === "nth") {
      await t.hover(Selector("a.is-subject div.v-card__title").nth(uidOrNth));
      if (visible) {
        await t.expect(Selector(`.input-` + action).visible).ok();
      } else {
        await t.expect(Selector(`.input-` + action).visible).notOk();
      }
    }
  }

  async checkHoverActionState(mode, uidOrNth, action, set) {
    if (mode === "uid") {
      await t.hover(Selector("a").withAttribute("data-uid", uidOrNth));
      if (set) {
        await t.expect(Selector(`a.uid-${uidOrNth}`).hasClass("is-" + action)).ok();
      } else {
        await t.expect(Selector(`a.uid-${uidOrNth}`).hasClass("is-" + action)).notOk();
      }
    }
    if (mode === "nth") {
      await t.hover(Selector("a.is-subject").nth(uidOrNth));
      if (set) {
        await t
          .expect(
            Selector("a.is-subject")
              .nth(uidOrNth)
              .hasClass("is-" + action)
          )
          .ok();
      } else {
        await t
          .expect(
            Selector("a.is-subject")
              .nth(uidOrNth)
              .hasClass("is-" + action)
          )
          .notOk();
      }
    }
  }
}
