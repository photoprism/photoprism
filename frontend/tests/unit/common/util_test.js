import "../fixtures";
import $util from "common/util";
import * as can from "common/can";
import { ContentTypeMp4AvcMain, ContentTypeMp4HvcMain } from "common/media";

let chai = require("chai/chai");
let assert = chai.assert;

describe("common/util", () => {
  it("should return size in KB", () => {
    const s = $util.formatBytes(10 * 1024);
    assert.equal(s, "10 KB");
  });
  it("should return size in GB", () => {
    const s = $util.formatBytes(10 * 1024 * 1024 * 1024);
    assert.equal(s, "10.0 GB");
  });
  it("should convert bytes in GB", () => {
    const b = $util.gigaBytes(10 * 1024 * 1024 * 1024);
    assert.equal(b, 10);
  });
  it("should return duration 3ns", () => {
    const duration = $util.formatDuration(-3);
    assert.equal(duration, "3ns");
  });
  it("should return duration 0s", () => {
    const duration = $util.formatDuration(0);
    assert.equal(duration, "0s");
  });
  it("should return duration 2µs", () => {
    const duration = $util.formatDuration(2000);
    assert.equal(duration, "2µs");
  });
  it("should return duration 4ms", () => {
    const duration = $util.formatDuration(4000000);
    assert.equal(duration, "4ms");
  });
  it("should return duration 6s", () => {
    const duration = $util.formatDuration(6000000000);
    assert.equal(duration, "0:06");
  });
  it("should return duration 10min", () => {
    const duration = $util.formatDuration(600000000000);
    assert.equal(duration, "10:00");
  });
  it("should return formatted seconds", () => {
    const floor = $util.formatSeconds(Math.floor(65.4));
    assert.equal(floor, "1:05");
    const ceil = $util.formatSeconds(Math.ceil(65.4));
    assert.equal(ceil, "1:06");
    const unknown = $util.formatSeconds(0);
    assert.equal(unknown, "0:00");
    const negative = $util.formatSeconds(-1);
    assert.equal(negative, "0:00");
  });
  it("should return remaining seconds", () => {
    const t = 23.3;
    const d = 42.6;
    const time = $util.formatSeconds(Math.floor(t));
    assert.equal(time, "0:23");
    const duration = $util.formatRemainingSeconds(0.0, d);
    assert.equal(duration, "0:43");
    const difference = $util.formatRemainingSeconds(t, d);
    assert.equal(difference, "0:20");
    const dotTime = $util.formatSeconds(Math.floor(9.5));
    assert.equal(dotTime, "0:09");
    const dotDiff = $util.formatRemainingSeconds(9.5, 12);
    assert.equal(dotDiff, "0:03");
    const smallDiff = $util.formatRemainingSeconds(7.959863, 8.033);
    assert.equal(smallDiff, "0:02");
  });
  it("should return formatted milliseconds", () => {
    const short = $util.formatNs(45065875);
    assert.equal(short, "45 ms");
    const long = $util.formatNs(45065875453454);
    assert.equal(long, "45,065,875 ms");
  });
  it("should return formatted camera name", () => {
    const iPhone15Pro = $util.formatCamera(
      { Make: "Apple", Model: "iPhone 15 Pro" },
      23,
      "Apple",
      "iPhone 15 Pro",
      false
    );
    assert.equal(iPhone15Pro, "iPhone 15 Pro");

    const iPhone15ProLong = $util.formatCamera(
      { Make: "Apple", Model: "iPhone 15 Pro" },
      23,
      "Apple",
      "iPhone 15 Pro",
      true
    );
    assert.equal(iPhone15ProLong, "Apple iPhone 15 Pro");

    const iPhone14 = $util.formatCamera({ Make: "Apple", Model: "iPhone 14" }, 22, "Apple", "iPhone 14", false);
    assert.equal(iPhone14, "iPhone 14");

    const iPhone13 = $util.formatCamera(null, 21, "Apple", "iPhone 13", false);
    assert.equal(iPhone13, "iPhone 13");
  });
  it("should return best matching thumbnail", () => {
    const thumbs = {
      fit_720: {
        w: 720,
        h: 481,
        src: "/api/v1/t/bfdcf45e58b1978af66bbf6212c195851dc65814/174usyd0/fit_720",
      },
      fit_1280: {
        w: 1280,
        h: 854,
        src: "/api/v1/t/bfdcf45e58b1978af66bbf6212c195851dc65814/174usyd0/fit_1280",
      },
      fit_1920: {
        w: 1800,
        h: 1200,
        src: "/api/v1/t/bfdcf45e58b1978af66bbf6212c195851dc65814/174usyd0/fit_1920",
      },
      fit_2560: {
        w: 2400,
        h: 1600,
        src: "/api/v1/t/bfdcf45e58b1978af66bbf6212c195851dc65814/174usyd0/fit_2560",
      },
      fit_4096: {
        w: 4096,
        h: 2732,
        src: "/api/v1/t/bfdcf45e58b1978af66bbf6212c195851dc65814/174usyd0/fit_4096",
      },
      fit_5120: {
        w: 5120,
        h: 3415,
        src: "/api/v1/t/bfdcf45e58b1978af66bbf6212c195851dc65814/174usyd0/fit_5120",
      },
      fit_7680: {
        w: 5120,
        h: 3415,
        src: "/api/v1/t/bfdcf45e58b1978af66bbf6212c195851dc65814/174usyd0/fit_5120",
      },
    };
    assert.equal($util.thumb(thumbs, 1200, 900).size, "fit_1280");
    assert.equal($util.thumb(thumbs, 1300, 900).size, "fit_1920");
    assert.equal($util.thumb(thumbs, 1300, 900).w, 1800);
    assert.equal($util.thumb(thumbs, 1300, 900).h, 1200);
    assert.equal(
      $util.thumb(thumbs, 1300, 900).src,
      "/api/v1/t/bfdcf45e58b1978af66bbf6212c195851dc65814/174usyd0/fit_1920"
    );
    assert.equal($util.thumb(thumbs, 1400, 1200).size, "fit_1920");
    assert.equal($util.thumb(thumbs, 100000, 120000).size, "fit_7680");
  });
  it("should return the approximate best thumbnail size name", () => {
    assert.equal($util.thumbSize(1300, 900), "fit_1280");
    assert.equal($util.thumbSize(1400, 1200), "fit_1920");
    assert.equal($util.thumbSize(100000, 120000), "fit_7680");
  });
  it("should return matching video format name", () => {
    const avc = $util.videoFormat("avc1", ContentTypeMp4AvcMain);
    assert.equal(avc, "avc");

    const hevc = $util.videoFormat("hvc1", ContentTypeMp4HvcMain);
    if (can.useMp4Hvc) {
      assert.equal(hevc, "hevc");
    } else {
      assert.equal(hevc, "avc");
    }

    const webm = $util.videoFormat("", "video/webm");
    if (can.useWebM) {
      assert.equal(webm, "webm");
    } else {
      assert.equal(webm, "avc");
    }
  });
  it("should convert -1 to roman", () => {
    const roman = $util.arabicToRoman(-1);
    assert.equal(roman, "");
  });
  it("should convert 2500 to roman", () => {
    const roman = $util.arabicToRoman(2500);
    assert.equal(roman, "MMD");
  });
  it("should convert 112 to roman", () => {
    const roman = $util.arabicToRoman(112);
    assert.equal(roman, "CXII");
  });
  it("should convert 9 to roman", () => {
    const roman = $util.arabicToRoman(9);
    assert.equal(roman, "IX");
  });
  it("should truncate xxx", () => {
    const result = $util.truncate("teststring");
    assert.equal(result, "teststring");
  });
  it("should truncate xxx", () => {
    const result = $util.truncate("teststring for mocha", 5, "ng");
    assert.equal(result, "tesng");
  });
  it("should encode html", () => {
    const result = $util.encodeHTML("Micha & Theresa > < 'Lilly'");
    assert.equal(result, "Micha &amp; Theresa &gt; &lt; &apos;Lilly&apos;");
  });
  it("should encode link", () => {
    const result = $util.encodeHTML("Try this: https://photoswipe.com/options/?foo=bar&bar=baz. It's a link!");
    assert.equal(
      result,
      `Try this: <a href="https://photoswipe.com/options/" target="_blank">https://photoswipe.com/options/</a> It&apos;s a link!`
    );
  });
});
