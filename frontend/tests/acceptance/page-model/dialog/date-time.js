import { Selector } from "testcafe";

export default class DateTimeDialog {
  constructor() {
    this.root = Selector(".p-meta-datetime-dialog", { timeout: 15000 });
    this.year = this.root.find(".input-year input");
    this.month = this.root.find(".input-month input");
    this.day = this.root.find(".input-day input");
    this.localTime = this.root.find(".input-local-time input");
    this.timezone = this.root.find(".input-timezone input");
    // v-autocomplete renders the chosen option text in a sibling element;
    // the inner <input>.value tracks the search/filter string instead.
    this.yearValue = this.root.find(".input-year .v-autocomplete__selection");
    this.monthValue = this.root.find(".input-month .v-autocomplete__selection");
    this.dayValue = this.root.find(".input-day .v-autocomplete__selection");
    this.timezoneValue = this.root.find(".input-timezone .v-autocomplete__selection");
    this.confirm = this.root.find(".action-confirm");
    this.cancel = this.root.find(".action-cancel");
  }
}
