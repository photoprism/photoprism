import { Selector, t } from "testcafe";

export default class Page {
  constructor() {
    this.placesSearch = Selector('input[aria-label="Search"]');
  }

  async search(term) {
    await t.typeText(this.placesSearch, term, { replace: true }).pressKey("enter");
  }
}
