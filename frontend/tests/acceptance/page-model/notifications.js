import { Selector, t } from "testcafe";
import { logMessage, logTime, logTimeEnd, showLogs } from "./helpers";

const waitAfterClick = 350; // Please note that all t.click have to wait to allow the clicked item time to go away (300ms fade out).

export default class Page {
  constructor() {
    this.notifyClose250 = Selector(".p-notify__close", {timeout: 250});    
  }

  // closeAllEventPopups will close any event popups that are open, ignoring any click issues.
  async closeAllEventPopups() {
    logTime("closeAllEventPopups");
    logMessage("Before While in closeAllEventPopups");
    while(await this.notifyClose250.visible) {
      try {
        logMessage("Before Click in closeAllEventPopups");
        await t.click(this.notifyClose250).wait(waitAfterClick);
        logMessage("After  Click in closeAllEventPopups");
      } catch {
        logMessage("After  Click In Catch in closeAllEventPopups");
        showLogs && console.trace("notify close missed in closeAllEventPopups");
      }
    }
    logTimeEnd("closeAllEventPopups");
  }

  // waitForSpecficEvent will wait for the event to show up, for delay amount of time (after closing any event messages that do not match).
  async waitForSpecficEvent(event, delay = 7000, close = true) {
    logMessage("Before While in waitForSpecficEvent");
    while(await this.notifyClose250.visible) {
      if (await Selector("div.p-notify__text", {timeout: 50}).withText(event).visible) {
        try {
          if (close) {
            logMessage("Before Click in waitForSpecficEvent");
            await t.click(this.notifyClose250).wait(waitAfterClick);
            logMessage("After  Click in waitForSpecficEvent");
          }
        } catch {
          // ignore the error as the item may not show up
          logMessage("After  Click In Catch in waitForSpecficEvent");
          console.trace("notify close missed in waitForSpecficEvent " + event);
        } finally {
          return;
        }
      }
      try {
        logMessage("Before Click in waitForSpecficEvent");
        await t.click(this.notifyClose250).wait(waitAfterClick);
        logMessage("After  Click in waitForSpecficEvent");
      } catch {
        logMessage("After  Click In Catch in waitForSpecficEvent");
        showLogs && console.trace("notify close missed in waitForSpecficEvent Pre");
      }
    }
    logMessage("Before Visible in waitForSpecficEvent");
    if ((await Selector("div.p-notify__text", {timeout: delay}).withText(event).visible) && close){
      try {
        logMessage("Before Click in waitForSpecficEvent");
        await t.click(this.notifyClose250).wait(waitAfterClick);
        logMessage("After  Click in waitForSpecficEvent");
      } catch {
        // ignore the error as the item may not show up
        logMessage("After  Click In Catch in waitForSpecficEvent");
        showLogs && console.trace("notify close missed in waitForSpecficEvent");
      }
    }
  }

  async waitForFileDeleted(delay = 10000, close = true) {
    logTime("waitForFileDeleted");
    await this.waitForSpecficEvent("File deleted", delay, close);
    logTimeEnd("waitForFileDeleted");
  }  

  async waitForFoldersToLoad(delay, close) {
    logTime("waitForFoldersToLoad");
    await this.waitForSpecficEvent(/[fF]older/, delay, close);
    logTimeEnd("waitForFoldersToLoad");
  }

  async waitForImport(delay = 10000, close = true) {
    logTime("waitForImport");
    await this.waitForSpecficEvent("Import completed in", delay, close);
    logTimeEnd("waitForImport");
  }

  async waitForIndexing(delay = 10000, close = true) {
    logTime("waitForIndexing");
    await this.waitForSpecficEvent("Indexing completed in", delay, close);
    logTimeEnd("waitForIndexing");
  }

  async waitForPeopleToLoad(delay, close = true) {
    logTime("waitForPeopleToLoad");
    await this.waitForSpecficEvent(/(people|person) (found|loaded)/, delay, close);
    logTimeEnd("waitForPeopleToLoad");
  }

  async waitForPersonCoverUpdate(delay, close = true) {
    logTime("waitForPersonCoverUpdate");
    await this.waitForSpecficEvent("Person cover updated", delay, close);
    logTimeEnd("waitForPersonCoverUpdate");
  }

  async waitForPhotosToLoad(delay, close = true){
    logTime("waitForPhotosToLoad");
    await this.waitForSpecficEvent(/(picture|pictures) found/, delay, close);
    logTimeEnd("waitForPhotosToLoad");
  }

  async waitForAlbumsToLoad(delay, close = true){
    logTime("waitForAlbumsToLoad");
    await this.waitForSpecficEvent(/(album|albums) found/, delay, close);
    logTimeEnd("waitForAlbumsToLoad");
  }

  async waitForSearchToFinish(delay, close = true){
    logTime("waitForSearchToFinish");
    await this.waitForSpecficEvent(/(found|contain|empty)/, delay, close);
    logTimeEnd("waitForSearchToFinish");
  }

  async waitForUnstack(delay = 12000, close = true) {
    logTime("waitForUnstack");
    await this.waitForSpecficEvent("File removed from stack", delay, close);
    logTimeEnd("waitForUnstack");
  }

  async waitForUpload(delay = 15000, close = true) {
    logTime("waitForUpload");
    await this.waitForSpecficEvent("Upload has been processed", delay, close);
    logTimeEnd("waitForUpload");
  }

  async waitForUploadFailed(delay = 15000, close = true) {
    logTime("waitForUploadFailed");
    await this.waitForSpecficEvent("Upload failed", delay, close);
    logTimeEnd("waitForUploadFailed");
  }
}