import { Selector, t } from "testcafe";
import { RequestLogger } from "testcafe";

const logger = RequestLogger(/http:\/\/localhost:2343\/api\/v1\/*/, {
  logResponseHeaders: true,
  logResponseBody: true,
});

export default class Page {
  constructor() {}

  async getNthSubjectUid(nth) {
    const NthSubject = await Selector("a.is-subject").nth(nth).getAttribute("data-uid");
    return NthSubject;
  }

  async getSubjectCount() {
    const SubjectCount = await Selector("a.is-subject", { timeout: 5000 }).count;
    return SubjectCount;
  }

  async selectSubjectFromUID(uid) {
    await t
      .hover(Selector("a.is-subject").withAttribute("data-uid", uid))
      .click(Selector(`.uid-${uid} .input-select`));
  }

  async toggleSelectNthSubject(nth) {
    await t
      .hover(Selector("a.is-subject", { timeout: 4000 }).nth(nth))
      .click(Selector("a.is-subject .input-select").nth(nth));
  }
}
