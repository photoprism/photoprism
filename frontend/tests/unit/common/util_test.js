import "../fixtures";
import Util from "common/util";

let chai = require("chai/chai");
let assert = chai.assert;

describe("common/util", () => {
  it("should return duration 3ns", () => {
    const duration = Util.duration(-3);
    assert.equal(duration, "3ns");
  });
  it("should return duration 0s", () => {
    const duration = Util.duration(0);
    assert.equal(duration, "0s");
  });
  it("should return duration 2µs", () => {
    const duration = Util.duration(2000);
    assert.equal(duration, "2µs");
  });
  it("should return duration 4ms", () => {
    const duration = Util.duration(4000000);
    assert.equal(duration, "4ms");
  });
  it("should return duration 6s", () => {
    const duration = Util.duration(6000000000);
    assert.equal(duration, "00:00:06");
  });
  it("should return duration 10min", () => {
    const duration = Util.duration(600000000000);
    assert.equal(duration, "00:10:00");
  });
  it("should convert -1 to roman", () => {
    const roman = Util.arabicToRoman(-1);
    assert.equal(roman, "");
  });
  it("should convert 2500 to roman", () => {
    const roman = Util.arabicToRoman(2500);
    assert.equal(roman, "MMD");
  });
  it("should convert 112 to roman", () => {
    const roman = Util.arabicToRoman(112);
    assert.equal(roman, "CXII");
  });
  it("should convert 9 to roman", () => {
    const roman = Util.arabicToRoman(9);
    assert.equal(roman, "IX");
  });
  it("should truncate xxx", () => {
    const result = Util.truncate("teststring");
    assert.equal(result, "teststring");
  });
  it("should truncate xxx", () => {
    const result = Util.truncate("teststring for mocha", 5, "ng");
    assert.equal(result, "tesng");
  });
  it("should encode html", () => {
    const result = Util.encodeHTML("Micha & Theresa > < 'Lilly'");
    assert.equal(result, "Micha &amp; Theresa &gt; &lt; &#x27;Lilly&#x27;");
  });
});
