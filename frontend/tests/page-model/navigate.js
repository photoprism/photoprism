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

  //login

  //logout

  //review card actions

  //open tab

  // album with uid visible

  // open album with uid

  //?dialogs?

  //checkboxen settings ?

  // selectors for index/import/logs/ ? input folder edit dialog

  //edit dialog disabled --funcionalities
  //edit dialog close, next, previous
  // edit dialog clear, reject face etc
  // update album --functionalities

  //selectors card view album location/ subject count / label count
  //remove album type from trigger context menu action
  //selectors edit dialogs label etc

  //sharing

}
