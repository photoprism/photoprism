import { Selector, t } from "testcafe";
import { RequestLogger } from "testcafe";

const logger = RequestLogger(/http:\/\/localhost:2343\/api\/v1\/*/, {
  logResponseHeaders: true,
  logResponseBody: true,
});

export default class Page {
  constructor() {
    this.view = Selector("div.p-view-select", { timeout: 15000 });
    this.camera = Selector("div.p-camera-select", { timeout: 15000 });
    this.countries = Selector("div.p-countries-select", { timeout: 15000 });
    this.time = Selector("div.p-time-select", { timeout: 15000 });
    this.search1 = Selector("div.input-search input", { timeout: 15000 });
  }

  //login Selectors

  //review card actions

  //open tab

  //edit single fields / check disabled, selectors for single fields?
  // check edit form values // get all current edit form values // set edit form values

  //?dialogs?

  //edit dialog disabled --funcionalities
  // update album --functionalities

  //selectors card view album location/ subject count / label count

  //remove album type from trigger context menu action

  //sharing

  //update all tests with new selectors

  //refactor admin/ member tests --> for each resource (photo, video, subject, label, albums , states) check context menu action/fullscreen actions/edit dialog/

}
