import "../fixtures";
import * as options from "../../../src/options/options";

let chai = require("chai/chai");
let assert = chai.assert;

describe("options/options", () => {
  it("should get timezones", () => {
    const timezones = options.TimeZones();
    assert.equal(timezones[0].ID, "");
    assert.equal(timezones[0].Name, "Local Time");
    assert.equal(timezones[1].ID, "UTC-12");
    assert.equal(timezones[1].Name, "UTC-12:00");
  });

  it("should get days", () => {
    const Days = options.Days();
    assert.equal(Days[0].text, "01");
    assert.equal(Days[30].text, "31");
  });

  it("should get years", () => {
    const Years = options.Years();
    const currentYear = new Date().getUTCFullYear();
    assert.equal(Years[0].text, currentYear);
  });

  it("should get indexed years", () => {
    const IndexedYears = options.IndexedYears();
    assert.equal(IndexedYears[0].text, "2021");
  });

  it("should get months", () => {
    const Months = options.Months();
    assert.equal(Months[5].text, "June");
  });

  it("should get short months", () => {
    const MonthsShort = options.MonthsShort();
    assert.equal(MonthsShort[5].text, "06");
  });

  it("should get languages", () => {
    const Languages = options.Languages();
    assert.equal(Languages[0].value, "en");
  });
});
