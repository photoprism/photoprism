import { describe, it, expect, afterEach } from "vitest";
import "../fixtures";
import $util from "common/util";
import { tokenRegexp, tokenLength } from "common/util";
import * as can from "common/can";
import { ContentTypeMp4AvcMain, ContentTypeMp4HvcMain } from "common/media";

describe("common/util", () => {
  it("should return size in KB", () => {
    const s = $util.formatBytes(10 * 1024);
    expect(s).toBe("10 KB");
  });
  it("should return size in GB", () => {
    const s = $util.formatBytes(10 * 1024 * 1024 * 1024);
    expect(s).toBe("10.0 GB");
  });
  it("should convert bytes in GB", () => {
    const b = $util.gigaBytes(10 * 1024 * 1024 * 1024);
    expect(b).toBe(10);
  });
  it("should return duration 3ns", () => {
    const duration = $util.formatDuration(-3);
    expect(duration).toBe("3ns");
  });
  it("should return duration 0s", () => {
    const duration = $util.formatDuration(0);
    expect(duration).toBe("0s");
  });
  it("should return duration 2µs", () => {
    const duration = $util.formatDuration(2000);
    expect(duration).toBe("2µs");
  });
  it("should return duration 4ms", () => {
    const duration = $util.formatDuration(4000000);
    expect(duration).toBe("4ms");
  });
  it("should return duration 6s", () => {
    const duration = $util.formatDuration(6000000000);
    expect(duration).toBe("0:06");
  });
  it("should return duration 10min", () => {
    const duration = $util.formatDuration(600000000000);
    expect(duration).toBe("10:00");
  });
  it("should return formatted seconds", () => {
    const floor = $util.formatSeconds(Math.floor(65.4));
    expect(floor).toBe("1:05");
    const ceil = $util.formatSeconds(Math.ceil(65.4));
    expect(ceil).toBe("1:06");
    const unknown = $util.formatSeconds(0);
    expect(unknown).toBe("0:00");
    const negative = $util.formatSeconds(-1);
    expect(negative).toBe("0:00");
  });
  it("should return remaining seconds", () => {
    const t = 23.3;
    const d = 42.6;
    const time = $util.formatSeconds(Math.floor(t));
    expect(time).toBe("0:23");
    const duration = $util.formatRemainingSeconds(0.0, d);
    expect(duration).toBe("0:43");
    const difference = $util.formatRemainingSeconds(t, d);
    expect(difference).toBe("0:20");
    const dotTime = $util.formatSeconds(Math.floor(9.5));
    expect(dotTime).toBe("0:09");
    const dotDiff = $util.formatRemainingSeconds(9.5, 12);
    expect(dotDiff).toBe("0:03");
    const smallDiff = $util.formatRemainingSeconds(7.959863, 8.033);
    expect(smallDiff).toBe("0:02");
  });
  it("should return formatted milliseconds", () => {
    const short = $util.formatNs(45065875);
    expect(short).toBe("45 ms");
    const long = $util.formatNs(45065875453454);
    expect(long).toBe("45,065,875 ms");
  });
  it("should return formatted camera name", () => {
    const iPhone15Pro = $util.formatCamera({ Make: "Apple", Model: "iPhone 15 Pro" }, 23, "Apple", "iPhone 15 Pro", false);
    expect(iPhone15Pro).toBe("iPhone 15 Pro");

    const iPhone15ProLong = $util.formatCamera({ Make: "Apple", Model: "iPhone 15 Pro" }, 23, "Apple", "iPhone 15 Pro", true);
    expect(iPhone15ProLong).toBe("Apple iPhone 15 Pro");

    const iPhone14 = $util.formatCamera({ Make: "Apple", Model: "iPhone 14" }, 22, "Apple", "iPhone 14", false);
    expect(iPhone14).toBe("iPhone 14");

    const iPhone13 = $util.formatCamera(null, 21, "Apple", "iPhone 13", false);
    expect(iPhone13).toBe("iPhone 13");
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
    expect($util.thumb(thumbs, 1200, 900).size).toBe("fit_1280");
    expect($util.thumb(thumbs, 1300, 900).size).toBe("fit_1920");
    expect($util.thumb(thumbs, 1300, 900).w).toBe(1800);
    expect($util.thumb(thumbs, 1300, 900).h).toBe(1200);
    expect($util.thumb(thumbs, 1300, 900).src).toBe("/api/v1/t/bfdcf45e58b1978af66bbf6212c195851dc65814/174usyd0/fit_1920");
    expect($util.thumb(thumbs, 1400, 1200).size).toBe("fit_1920");
    expect($util.thumb(thumbs, 100000, 120000).size).toBe("fit_7680");
  });
  it("should return the approximate best thumbnail size name", () => {
    expect($util.thumbSize(1300, 900)).toBe("fit_1280");
    expect($util.thumbSize(1400, 1200)).toBe("fit_1920");
    expect($util.thumbSize(100000, 120000)).toBe("fit_7680");
  });
  it("should return matching video format name", () => {
    const avc = $util.videoFormat("avc1", ContentTypeMp4AvcMain);
    expect(avc).toBe("avc");

    const hevc = $util.videoFormat("hvc1", ContentTypeMp4HvcMain);
    if (can.useMp4Hvc) {
      expect(hevc).toBe("hevc");
    } else {
      expect(hevc).toBe("avc");
    }

    const webm = $util.videoFormat("", "video/webm");
    if (can.useWebM) {
      expect(webm).toBe("webm");
    } else {
      expect(webm).toBe("avc");
    }
  });
  it("should convert -1 to roman", () => {
    const roman = $util.arabicToRoman(-1);
    expect(roman).toBe("");
  });
  it("should convert 2500 to roman", () => {
    const roman = $util.arabicToRoman(2500);
    expect(roman).toBe("MMD");
  });
  it("should convert 112 to roman", () => {
    const roman = $util.arabicToRoman(112);
    expect(roman).toBe("CXII");
  });
  it("should convert 9 to roman", () => {
    const roman = $util.arabicToRoman(9);
    expect(roman).toBe("IX");
  });
  it("should truncate xxx", () => {
    const result = $util.truncate("teststring");
    expect(result).toBe("teststring");
  });
  it("should truncate xxx", () => {
    const result = $util.truncate("teststring for vitest", 5, "ng");
    expect(result).toBe("tesng");
  });
  it("should encode html", () => {
    const result = $util.encodeHTML("Micha & Theresa > < 'Lilly'");
    expect(result).toBe("Micha &amp; Theresa &gt; &lt; &apos;Lilly&apos;");
  });
  it("should encode link", () => {
    const result = $util.encodeHTML("Try this: https://photoswipe.com/options/?foo=bar&bar=baz. It's a link!");
    expect(result).toBe(
      `Try this: <a href="https://photoswipe.com/options/" target="_blank" rel="noopener noreferrer">https://photoswipe.com/options/</a> It&apos;s a link!`
    );
  });
  it("should sanitize html using the shared allowlist", () => {
    const result = $util.sanitizeHtml(
      `<p>Hello <strong>there</strong> <img src=x onerror=alert(1) /> <a href="https://example.com" target="_blank">link</a></p>`
    );

    expect(result).toBe(`<p>Hello <strong>there</strong>  <a href="https://example.com" target="_blank" rel="noopener noreferrer">link</a></p>`);
  });
  it("should generate tokens reliably", () => {
    const tokens = new Set();
    const numTokens = 100;
    for (let i = 0; i < numTokens; i++) {
      const token = $util.generateToken();
      expect(token).toHaveLength(tokenLength);
      expect(token).toMatch(tokenRegexp);
      tokens.add(token);
    }
    // Check they are all unique
    expect(tokens.size).toBe(numTokens);
  });

  describe("normalizeTitle", () => {
    it("preserves lowercase ASCII", () => {
      expect($util.normalizeTitle("cat")).toBe("cat");
    });
    it("lowercases input", () => {
      expect($util.normalizeTitle("Cat")).toBe("cat");
    });
    it("replaces & with and", () => {
      expect($util.normalizeTitle("Rock & Roll")).toBe("rock and roll");
    });
    it("replaces underscores with spaces", () => {
      expect($util.normalizeTitle("hello_world")).toBe("hello world");
    });
    it("replaces hyphens with spaces", () => {
      expect($util.normalizeTitle("hello-world")).toBe("hello world");
    });
    it("replaces pluses with spaces", () => {
      expect($util.normalizeTitle("hello+world")).toBe("hello world");
    });
    it("replaces periods with spaces", () => {
      expect($util.normalizeTitle("hello.cat")).toBe("hello cat");
      expect($util.normalizeTitle("photoprism.app")).toBe("photoprism app");
      expect($util.normalizeTitle("2024.07.15")).toBe("2024 07 15");
    });
    it("treats all punctuation as word separators for case-insensitive comparison", () => {
      // All punctuation (`.`, `,`, `;`, `:`, `!`, `?`, `/`, …) collapses
      // to whitespace so the dedup comparison ignores it. Letters, digits,
      // and emoji are preserved.
      expect($util.normalizeTitle("foo,bar")).toBe("foo bar");
      expect($util.normalizeTitle("foo;bar")).toBe("foo bar");
      expect($util.normalizeTitle("foo:bar")).toBe("foo bar");
      expect($util.normalizeTitle("foo!bar?baz")).toBe("foo bar baz");
      expect($util.normalizeTitle("Mr. Smith")).toBe("mr smith");
    });
    it("collapses runs of mixed word separators into a single space", () => {
      expect($util.normalizeTitle("hello._-+cat")).toBe("hello cat");
      expect($util.normalizeTitle("hello . cat")).toBe("hello cat");
      expect($util.normalizeTitle("foo,,, bar")).toBe("foo bar");
    });
    it("normalizes hello.cat, hello-cat, hello_cat, hello,cat, and Hello Cat to the same value", () => {
      expect($util.normalizeTitle("hello.cat")).toBe("hello cat");
      expect($util.normalizeTitle("hello-cat")).toBe("hello cat");
      expect($util.normalizeTitle("hello_cat")).toBe("hello cat");
      expect($util.normalizeTitle("hello,cat")).toBe("hello cat");
      expect($util.normalizeTitle("Hello Cat")).toBe("hello cat");
    });
    it("preserves emoji", () => {
      expect($util.normalizeTitle("🌅")).toBe("🌅");
    });
    it("preserves emoji with text", () => {
      expect($util.normalizeTitle("🏔️ Mountains")).toBe("🏔️ mountains");
    });
    it("preserves compound emoji with ZWJ", () => {
      expect($util.normalizeTitle("👨‍👩‍👧")).toBe("👨‍👩‍👧");
    });
    it("preserves accented characters", () => {
      expect($util.normalizeTitle("café")).toBe("café");
    });
    it("preserves flag emoji", () => {
      expect($util.normalizeTitle("🇺🇸")).toBe("🇺🇸");
    });
    it("preserves skin tone emoji", () => {
      expect($util.normalizeTitle("👋🏽")).toBe("👋🏽");
    });
    it("preserves keycap sequences", () => {
      expect($util.normalizeTitle("1️⃣")).toBe("1️⃣");
    });
    it("preserves CJK characters", () => {
      expect($util.normalizeTitle("猫")).toBe("猫");
    });
    it("converts punctuation to whitespace and keeps emoji and text", () => {
      expect($util.normalizeTitle("hello! 🌅 world")).toBe("hello 🌅 world");
    });
    it("returns empty for punctuation-only input", () => {
      // Punctuation-only inputs collapse to a single space and then trim,
      // so they normalize to empty — a title made entirely of punctuation
      // characters cannot be created or matched.
      expect($util.normalizeTitle("!!!")).toBe("");
      expect($util.normalizeTitle("...")).toBe("");
      expect($util.normalizeTitle("---")).toBe("");
      expect($util.normalizeTitle("+_-.")).toBe("");
      expect($util.normalizeTitle(",;:!?")).toBe("");
    });
    it("trims leading and trailing whitespace", () => {
      expect($util.normalizeTitle("  hello cat  ")).toBe("hello cat");
      expect($util.normalizeTitle(".hello-cat.")).toBe("hello cat");
      expect($util.normalizeTitle("!!!hello!!!")).toBe("hello");
    });
    it("returns empty for null", () => {
      expect($util.normalizeTitle(null)).toBe("");
    });
    it("returns empty for undefined", () => {
      expect($util.normalizeTitle(undefined)).toBe("");
    });
  });

  describe("typeName", () => {
    it("returns the localized label for known media types", () => {
      expect($util.typeName("image")).toBe("Image");
      expect($util.typeName("raw")).toBe("Raw");
      expect($util.typeName("live")).toBe("Live");
      expect($util.typeName("video")).toBe("Video");
      expect($util.typeName("audio")).toBe("Audio");
      expect($util.typeName("animated")).toBe("Animated");
      expect($util.typeName("vector")).toBe("Vector");
      expect($util.typeName("document")).toBe("Document");
      expect($util.typeName("sidecar")).toBe("Sidecar");
    });
    it("falls back to defaultValue for unknown type", () => {
      expect($util.typeName("unknown", "File")).toBe("File");
    });
    it("falls back to defaultValue for empty/null/undefined input", () => {
      expect($util.typeName("", "File")).toBe("File");
      expect($util.typeName(null, "File")).toBe("File");
      expect($util.typeName(undefined, "File")).toBe("File");
    });
    it("returns empty string when no defaultValue and unknown type", () => {
      expect($util.typeName("unknown")).toBe("");
      expect($util.typeName(null)).toBe("");
    });
  });

  // isMobile must return a Boolean. Earlier code short-circuited
  // `navigator.maxTouchPoints && maxTouchPoints > 2`, which returned the
  // Number 0 on desktop and tripped Vue prop type checks (VTooltip.disabled).
  describe("isMobile", () => {
    const userAgent = navigator.userAgent;
    const maxTouchPoints = navigator.maxTouchPoints;
    const stub = (ua, touch) => {
      Object.defineProperty(navigator, "userAgent", { value: ua, configurable: true });
      Object.defineProperty(navigator, "maxTouchPoints", { value: touch, configurable: true });
    };
    afterEach(() => stub(userAgent, maxTouchPoints));

    it("returns Boolean false on desktop with no touch", () => {
      stub("Mozilla/5.0 (X11; Linux x86_64)", 0);
      const result = $util.isMobile();
      expect(typeof result).toBe("boolean");
      expect(result).toBe(false);
    });
    it("returns Boolean false when maxTouchPoints is undefined", () => {
      stub("Mozilla/5.0 (X11; Linux x86_64)", undefined);
      const result = $util.isMobile();
      expect(typeof result).toBe("boolean");
      expect(result).toBe(false);
    });
    it("returns true for a mobile user agent", () => {
      stub("Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X)", 0);
      expect($util.isMobile()).toBe(true);
    });
    it("returns true when maxTouchPoints > 2 (iPad in desktop mode)", () => {
      stub("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)", 5);
      expect($util.isMobile()).toBe(true);
    });
    it("returns false when maxTouchPoints is 2 or less", () => {
      stub("Mozilla/5.0 (X11; Linux x86_64)", 2);
      expect($util.isMobile()).toBe(false);
    });
  });
});
