import { Selector, t } from "testcafe";
// import { ClientFunction } from "testcafe";

// // This attempts to turn off the transition that events has.
// // It has been less successful than hoped, as there are still delays encountered.
// const disableTransition = ClientFunction(() => {
//   let result = ""
//   for (const sheet of document.styleSheets){
//     for (const rule of sheet.cssRules) {
//       if (String(rule.selectorText).endsWith("transition-leave-active")) {
//         result += "| " + rule.selectorText;
//         result += "| " + rule.style.cssText;
//         rule.style.transition = 'none';
//       }
//     }
//   }
//   return result;
// })

export default class Page {
  constructor() {
    this.notifyClose250 = Selector(".p-notify__close", {timeout: 250})    
  }

  // Close any event popups that are open, ignoring any click issues.
  async closeAllEventPopups() {
    console.time("closeAllEventPopups");
    // let a = await disableTransition();
    // console.log(`disableTransition was ${a}`)
    var now = new Date();
    console.log(now.toTimeString() + " " + String(now.getMilliseconds()) + " Before While");
    while(await this.notifyClose250.visible) {
      try {
        now = new Date();
        console.log(now.toTimeString() + " " + String(now.getMilliseconds()) + " Before Click");
        await t.click(this.notifyClose250).wait(350);  // Wait to allow the clicked item time to go away (300ms fade out).
      } catch {
        now = new Date();
        console.log(now.toTimeString() + " " + String(now.getMilliseconds()) + " After Click");
        console.trace("notify close missed in closeAllEventPopups")
      }
      now = new Date();
      console.log(now.toTimeString() + " " + String(now.getMilliseconds()) + " After Click");
    }
    console.timeEnd("closeAllEventPopups");
  }

  async waitForSpecficEvent(event, delay = 7000, close = true) {
    while(await this.notifyClose250.visible) {
      if (await Selector("div.p-notify__text", {timeout: 50}).withText(event).visible) {
        try {
          if (close) {
            await t.click(this.notifyClose250).wait(350);
          }
        } catch {
          // ignore the error as the item may not show up
          console.trace("notify close missed in waitForSpecficEvent " + event)
        } finally {
          return
        }
      }
      try {
        await t.click(this.notifyClose250).wait(350);
      } catch {
        console.trace("notify close missed in waitForSpecficEvent Pre")
      }
    }

    if ((await Selector("div.p-notify__text", {timeout: delay}).withText(event).visible) && close){
      try {
        await t.click(this.notifyClose250).wait(350);
      } catch {
        // ignore the error as the item may not show up
        console.trace("notify close missed in waitForSpecficEvent")
      }
    }
  }

  async waitForPhotosToLoad(delay, close = true){
    console.time("waitForPhotosToLoad")
    await this.waitForSpecficEvent(/(picture|pictures) found/, delay, close);
    console.timeEnd("waitForPhotosToLoad")
  }

  async waitForPeopleToLoad(delay, close = true) {
    console.time("waitForPeopleToLoad")
    await this.waitForSpecficEvent(/(people|person) (found|loaded)/, delay, close);
    console.timeEnd("waitForPeopleToLoad")
  }

  async waitForPersonCoverUpdate(delay, close = true) {
    console.time("waitForPersonCoverUpdate")
    await this.waitForSpecficEvent("Person cover updated", delay, close);
    console.timeEnd("waitForPersonCoverUpdate")
  }

  async waitForSearchToFinish(delay, close = true){
    console.time("waitForSearchToFinish")
    await this.waitForSpecficEvent(/(found|contain|empty)/, delay, close);
    console.timeEnd("waitForSearchToFinish")
  }

  async waitForFoldersToLoad(delay, close) {
    console.time("waitForFoldersToLoad")
    await this.waitForSpecficEvent(/[fF]older/, delay, close);
    console.timeEnd("waitForFoldersToLoad")
  }


}