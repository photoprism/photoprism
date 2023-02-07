import { Selector, t } from "testcafe";

export default class Page {
  constructor() {}

  async getNthLabeltUid(nth) {
    const NthLabel = await Selector("a.is-label").nth(nth).getAttribute("data-uid");
    return NthLabel;
  }

  async getLabelCount() {
    const LabelCount = await Selector("a.is-label", { timeout: 5000 }).count;
    return LabelCount;
  }

  async selectLabelFromUID(uid) {
    await t
      .hover(Selector("a.is-label").withAttribute("data-uid", uid))
      .click(Selector(`.uid-${uid} .input-select`));
  }

  async toggleSelectNthLabel(nth) {
    await t
      .hover(Selector("a.is-label", { timeout: 4000 }).nth(nth))
      .click(Selector("a.is-label .input-select").nth(nth));
  }

  async openNthLabel(nth) {
    await t.click(Selector("a.is-label").nth(nth)).expect(Selector("div.is-photo").visible).ok();
  }

  async openLabelWithUid(uid) {
    await t.click(Selector("a.is-label").withAttribute("data-uid", uid));
  }

  async triggerHoverAction(mode, uidOrNth, action) {
    if (mode === "uid") {
      await t.hover(Selector("a.uid-" + uidOrNth));
      Selector("a.uid-" + uidOrNth + " .input-" + action);
      await t.click(Selector("a.uid-" + uidOrNth + " .input-" + action));
    }
    if (mode === "nth") {
      await t.hover(Selector("a.is-label").nth(uidOrNth));
      await t.click(Selector(`.input-` + action));
    }
  }

  async checkHoverActionAvailability(mode, uidOrNth, action, visible) {
    if (mode === "uid") {
      await t.hover(Selector("a.is-label").withAttribute("data-uid", uidOrNth));
      if (visible) {
        await t.expect(Selector(`.uid-${uidOrNth} .input-` + action).visible).ok();
      } else {
        await t.expect(Selector(`.uid-${uidOrNth} .input-` + action).visible).notOk();
      }
    }
    if (mode === "nth") {
      await t.hover(Selector("a.is-label").nth(uidOrNth));
      if (visible) {
        await t.expect(Selector(`div.input-` + action).visible).ok();
      } else {
        await t.expect(Selector(`div.input-` + action).visible).notOk();
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
      await t.hover(Selector("a.is-label").nth(uidOrNth));
      if (set) {
        await t
          .expect(
            Selector("a.is-label")
              .nth(uidOrNth)
              .hasClass("is-" + action)
          )
          .ok();
      } else {
        await t
          .expect(
            Selector("a.is-label")
              .nth(uidOrNth)
              .hasClass("is-" + action)
          )
          .notOk();
      }
    }
  }
}
