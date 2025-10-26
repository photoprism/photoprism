import { Selector, t } from "testcafe";

export default class Page {
  constructor() {
    this.recognizedTab = Selector("#tab-people", { timeout: 15000 });
    this.newTab = Selector("#tab-people_faces", { timeout: 15000 });
    this.showAllNewButton = Selector('a[href="/all?q=face%3Anew"]');
    this.subjectName = Selector("a.is-subject div.meta-title");
    this.search = Selector(".v-window-item--active .input-search input");
  }

  async addNameToFace(id, name) {
    await t.typeText(Selector("div[data-id=" + id + "] div.input-name input"), name).pressKey("enter");
  }

  async renameSubject(uid, name) {
    await t
      .click(Selector("div[data-uid=" + uid + "] div.meta-title"))
      .typeText(Selector("div.input-title input"), name, { replace: true })
      .click(Selector("button.action-confirm"));
  }

  async getNthSubjectUid(nth) {
    const NthSubject = await Selector("div.result.is-subject").nth(nth).getAttribute("data-uid");
    return NthSubject;
  }

  async getNthFaceUid(nth) {
    const NthFace = await Selector("div.is-face").nth(nth).getAttribute("data-id");
    return NthFace;
  }

  async getSubjectCount() {
    const SubjectCount = await Selector("div.result.is-subject", { timeout: 5000 }).count;
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
      .hover(Selector("div.result.is-subject").withAttribute("data-uid", uid))
      .click(Selector(`.uid-${uid} .input-select`));
  }

  async toggleSelectNthSubject(nth) {
    await t
      .hover(Selector("div.result.is-subject", { timeout: 4000 }).nth(nth))
      .click(Selector("div.result.is-subject .input-select").nth(nth));
  }

  async openNthSubject(nth) {
    await t.click(Selector("div.result.is-subject").nth(nth)).expect(Selector("div.is-photo").visible).ok();
  }

  async openSubjectWithUid(uid) {
    await t
      .click(Selector("div[data-uid=" + uid + "] div.preview"))
      .expect(Selector("div.is-photo").visible)
      .ok();
  }

  async openFaceWithUid(uid) {
    await t.click(Selector("div[data-id=" + uid + "] div.preview"));
  }

  async checkSubjectVisibility(mode, uidOrName, visible) {
    if (visible) {
      if (mode === "uid") {
        await t.expect(Selector("div").withAttribute("data-uid", uidOrName).visible).ok();
      } else {
        await t.expect(Selector("div div.meta-title").withText(uidOrName).visible).ok();
      }
    } else if (!visible) {
      if (mode === "uid") {
        await t.expect(Selector("div").withAttribute("data-uid", uidOrName).visible).notOk();
      } else {
        await t.expect(Selector("div div.meta-title").withText(uidOrName).visible).notOk();
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
      await t.hover(Selector("div.uid-" + uidOrNth));
      await t.click(Selector("div.uid-" + uidOrNth + " .input-" + action));
    }
    if (mode === "nth") {
      await t.hover(Selector("div.result.is-subject").nth(uidOrNth));
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
      await t.hover(Selector("div.result.is-subject").withAttribute("data-uid", uidOrNth));
      if (visible) {
        await t.expect(Selector(`.uid-${uidOrNth} .input-` + action).visible).ok();
      } else {
        await t.expect(Selector(`.uid-${uidOrNth} .input-` + action).visible).notOk();
      }
    }
    if (mode === "nth") {
      await t.hover(Selector("div.result.is-subject").nth(uidOrNth));
      if (visible) {
        await t.expect(Selector(`.input-` + action).visible).ok();
      } else {
        await t.expect(Selector(`.input-` + action).visible).notOk();
      }
    }
  }

  async checkHoverActionState(mode, uidOrNth, action, set) {
    if (mode === "uid") {
      await t.hover(Selector("div").withAttribute("data-uid", uidOrNth));
      if (set) {
        await t.expect(Selector(`div.uid-${uidOrNth}`).hasClass("is-" + action)).ok();
      } else {
        await t.expect(Selector(`div.uid-${uidOrNth}`).hasClass("is-" + action)).notOk();
      }
    }
    if (mode === "nth") {
      await t.hover(Selector("div.result.is-subject").nth(uidOrNth));
      if (set) {
        await t
          .expect(
            Selector("div.result.is-subject")
              .nth(uidOrNth)
              .hasClass("is-" + action)
          )
          .ok();
      } else {
        await t
          .expect(
            Selector("div.result.is-subject")
              .nth(uidOrNth)
              .hasClass("is-" + action)
          )
          .notOk();
      }
    }
  }

  async waitForPeopleToLoad(delay, close = true) {
    console.time("waitForPeopleToLoad")
    // if (close) {
    //   try {
    //     await t.click(Selector("div.p-notify__text", {timeout: delay}).withText(/(people|person) (found|loaded)/));
    //   } catch (e) {
    //     // ignore the error as the item may not show up
    //   }
    // } else {
    //   await Selector("div.p-notify__text", {timeout: delay}).withText(/(people|person) (found|loaded)/).visible;
    // }

    while(await Selector(".p-notify__close", {timeout: 250}).visible) {
      if (await Selector("div.p-notify__text", {timeout: 50}).withText(/(people|person) (found|loaded)/).visible) {
        break
      }
      try {
        await t.click(Selector(".p-notify__close", {timeout: 250}));
      } catch {
        console.log(".p-notify__close missed in waitForPeopleToLoad Pre")
      }
    }

    if ((await Selector("div.p-notify__text", {timeout: delay}).withText(/(people|person) (found|loaded)/).visible) && close){
      try {
        await t.click(Selector(".p-notify__close", {timeout: 250}));
      } catch (e) {
        // ignore the error as the item may not show up
        console.log(".p-notify__close missed in waitForPeopleToLoad")
      }
    }
    console.timeEnd("waitForPeopleToLoad")
  }

  async waitForPersonCoverUpdate(delay, close = true) {
    console.time("waitForPersonCoverUpdate")
    // if (close) {
    //   try {
    //     await t.click(Selector("div.p-notify__text", {timeout: delay}).withText("Person cover updated"));
    //   } catch (e) {
    //     // ignore the error as the item may not show up
    //   }
    // } else {
    //   await Selector("div.p-notify__text", {timeout: delay}).withText("Person cover updated").visible;
    // }

    // Get rid of the other selectors
    while(await Selector(".p-notify__close", {timeout: 250}).visible) {
      if (await Selector("div.p-notify__text", {timeout: 50}).withText("Person cover updated").visible) {
        break
      }
      try {
        await t.click(Selector(".p-notify__close", {timeout: 250}));
      } catch {
        // Do Nothing
        console.log(".p-notify__close missed in waitForPersonCoverUpdate Pre")
      }
    }

    if ((await Selector("div.p-notify__text", {timeout: delay}).withText("Person cover updated").visible) && close){
      try {
        await t.click(Selector(".p-notify__close", {timeout: 250}));
      } catch (e) {
        // ignore the error as the item may not show up
        console.log(".p-notify__close missed in waitForPersonCoverUpdate")
      }
    }
    console.timeEnd("waitForPersonCoverUpdate")
  }

  async waitForPeopleToLoadSlow(delay, close = true) {
    console.time("waitForPeopleToLoad")
    if ((await Selector("div.p-notify__text", {timeout: delay}).withText(/(people|person) (found|loaded)/).visible) && close){
        while(await Selector(".p-notify__close", {timeout: 250}).visible) {
          await t.wait(250);
        }
    }
    console.timeEnd("waitForPeopleToLoad")
  }

  async waitForPersonCoverUpdateSlow(delay, close = true) {
    console.time("waitForPersonCoverUpdate")
    if ((await Selector("div.p-notify__text", {timeout: delay}).withText("Person cover updated").visible) && close){
        while(await Selector(".p-notify__close", {timeout: 250}).visible) {
          await t.wait(250);
        }
    }
    console.timeEnd("waitForPersonCoverUpdate")
  }

}
