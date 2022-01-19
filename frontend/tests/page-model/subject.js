import { Selector, t } from "testcafe";

export default class Page {
  constructor() {}

  async getNthSubjectUid(nth) {
    const NthSubject = await Selector("a.is-subject").nth(nth).getAttribute("data-uid");
    return NthSubject;
  }

  async getSubjectCount() {
    const SubjectCount = await Selector("a.is-subject", { timeout: 5000 }).count;
    return SubjectCount;
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

  //hidden, favorite, select
  async triggerHoverAction(mode, uidOrNth, action) {
    if (mode === "uid") {
      await t.hover(Selector("a.uid-" + uidOrNth));
      Selector("a.uid-" + uidOrNth + " .input-" + action);
      await t.click(Selector("a.uid-" + uidOrNth + " .input-" + action));
    }
    if (mode === "nth") {
      await t.hover(Selector("a.is-subject").nth(uidOrNth));
      await t.click(Selector(`.input-` + action));
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

  async checkSubjectVisibility(name, visible) {
    if (visible) {
      await t.expect(Selector("a div.v-card__title").withText(name).visible).ok();
    } else {
      await t.expect(Selector("a div.v-card__title").withText(name).visible).notOk();
    }
  }

  async openNthSubject(nth) {
    await t.click(Selector("a.is-subject").nth(nth)).expect(Selector("div.is-photo").visible).ok();
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
