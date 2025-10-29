import { Selector, t } from "testcafe";

const showLogs = process.env.SHOW_LOGS == "true";
const waitAfterClick = 350 // Please note that all t.click have to wait to allow the clicked item time to go away (300ms fade out).

export default class Page {
  constructor() {
    this.notifyClose250 = Selector(".p-notify__close", {timeout: 250})    
  }

  logMessage(message) {
    var now = new Date();
    console.log(now.toISOString() + " " + message);
  }

  // closeAllEventPopups will close any event popups that are open, ignoring any click issues.
  async closeAllEventPopups() {
    showLogs && console.time("closeAllEventPopups");
    showLogs && this.logMessage("Before While in closeAllEventPopups");
    while(await this.notifyClose250.visible) {
      try {
        showLogs && this.logMessage("Before Click in closeAllEventPopups");
        await t.click(this.notifyClose250).wait(waitAfterClick);
        showLogs && this.logMessage("After  Click in closeAllEventPopups");
      } catch {
        showLogs && this.logMessage("After  Click In Catch in closeAllEventPopups");
        showLogs && console.trace("notify close missed in closeAllEventPopups")
      }
    }
    showLogs && console.timeEnd("closeAllEventPopups");
  }

  // waitForSpecficEvent will wait for the event to show up, for delay amount of time (after closing any event messages that do not match).
  async waitForSpecficEvent(event, delay = 7000, close = true) {
    showLogs && this.logMessage("Before While in waitForSpecficEvent");
    while(await this.notifyClose250.visible) {
      if (await Selector("div.p-notify__text", {timeout: 50}).withText(event).visible) {
        try {
          if (close) {
            showLogs && this.logMessage("Before Click in waitForSpecficEvent");
            await t.click(this.notifyClose250).wait(waitAfterClick);
            showLogs && this.logMessage("After  Click in waitForSpecficEvent");
          }
        } catch {
          // ignore the error as the item may not show up
          showLogs && this.logMessage("After  Click In Catch in waitForSpecficEvent");
          console.trace("notify close missed in waitForSpecficEvent " + event)
        } finally {
          return
        }
      }
      try {
        showLogs && this.logMessage("Before Click in waitForSpecficEvent");
        await t.click(this.notifyClose250).wait(waitAfterClick);
        showLogs && this.logMessage("After  Click in waitForSpecficEvent");
      } catch {
        showLogs && this.logMessage("After  Click In Catch in waitForSpecficEvent");
        showLogs && console.trace("notify close missed in waitForSpecficEvent Pre")
      }
    }
    showLogs && this.logMessage("Before Visible in waitForSpecficEvent");
    if ((await Selector("div.p-notify__text", {timeout: delay}).withText(event).visible) && close){
      try {
        showLogs && this.logMessage("Before Click in waitForSpecficEvent");
        await t.click(this.notifyClose250).wait(waitAfterClick);
        showLogs && this.logMessage("After  Click in waitForSpecficEvent");
      } catch {
        // ignore the error as the item may not show up
        showLogs && this.logMessage("After  Click In Catch in waitForSpecficEvent");
        showLogs && console.trace("notify close missed in waitForSpecficEvent")
      }
    }
  }

  async waitForFileDeleted(delay = 10000, close = true) {
    showLogs && console.time("waitForFileDeleted")
    await this.waitForSpecficEvent("File deleted", delay, close);
    showLogs && console.timeEnd("waitForFileDeleted")
  }  

  async waitForFoldersToLoad(delay, close) {
    showLogs && console.time("waitForFoldersToLoad")
    await this.waitForSpecficEvent(/[fF]older/, delay, close);
    showLogs && console.timeEnd("waitForFoldersToLoad")
  }

  async waitForImport(delay = 10000, close = true) {
    showLogs && console.time("waitForImport")
    await this.waitForSpecficEvent("Import completed in", delay, close);
    showLogs && console.timeEnd("waitForImport")
  }

  async waitForIndexing(delay = 10000, close = true) {
    showLogs && console.time("waitForIndexing")
    await this.waitForSpecficEvent("Indexing completed in", delay, close);
    showLogs && console.timeEnd("waitForIndexing")
  }

  async waitForPeopleToLoad(delay, close = true) {
    showLogs && console.time("waitForPeopleToLoad")
    await this.waitForSpecficEvent(/(people|person) (found|loaded)/, delay, close);
    showLogs && console.timeEnd("waitForPeopleToLoad")
  }

  async waitForPersonCoverUpdate(delay, close = true) {
    showLogs && console.time("waitForPersonCoverUpdate")
    await this.waitForSpecficEvent("Person cover updated", delay, close);
    showLogs && console.timeEnd("waitForPersonCoverUpdate")
  }

  async waitForPhotosToLoad(delay, close = true){
    showLogs && console.time("waitForPhotosToLoad")
    await this.waitForSpecficEvent(/(picture|pictures) found/, delay, close);
    showLogs && console.timeEnd("waitForPhotosToLoad")
  }

  async waitForSearchToFinish(delay, close = true){
    showLogs && console.time("waitForSearchToFinish")
    await this.waitForSpecficEvent(/(found|contain|empty)/, delay, close);
    showLogs && console.timeEnd("waitForSearchToFinish")
  }

  async waitForUnstack(delay = 12000, close = true) {
    showLogs && console.time("waitForUnstack")
    await this.waitForSpecficEvent("File removed from stack", delay, close);
    showLogs && console.timeEnd("waitForUnstack")
  }

  async waitForUpload(delay = 15000, close = true) {
    showLogs && console.time("waitForUpload")
    await this.waitForSpecficEvent("Upload has been processed", delay, close);
    showLogs && console.timeEnd("waitForUpload")
  }

  async waitForUploadFailed(delay = 15000, close = true) {
    showLogs && console.time("waitForUploadFailed")
    await this.waitForSpecficEvent("Upload failed", delay, close);
    showLogs && console.timeEnd("waitForUploadFailed")
  }
}