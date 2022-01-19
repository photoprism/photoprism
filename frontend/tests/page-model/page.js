import { Selector, t } from "testcafe";

export default class Page {
  constructor() {
    this.selectOption = Selector("div.v-list__tile__title", { timeout: 15000 });
    this.cardTitle = Selector("button.action-title-edit");
    this.cardDescription = Selector('div[title="Description"]');


    //album location
    //photo title button.action-title-edit
    // photo decsription
  }
}
