import { Selector, t } from "testcafe";
import { RequestLogger } from "testcafe";

const logger = RequestLogger(/http:\/\/localhost:2343\/api\/v1\/*/, {
  logResponseHeaders: true,
  logResponseBody: true,
});

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
}
