import { describe, it, expect } from "vitest";
import "../fixtures";
import { Form, FormPropertyType, rules } from "common/form";

describe("common/form", () => {
  it("setting and getting definition", () => {
    const def = { foo: { type: FormPropertyType.String, caption: "Foo" } };
    const form = new Form();

    form.setDefinition(def);

    const result = form.getDefinition();
    expect(result).toBe(def);
  });

  it("setting and getting a value according to type", () => {
    const def = {
      foo: { type: FormPropertyType.String, caption: "Foo" },
    };
    const form = new Form();

    form.setDefinition(def);
    form.setValue("foo", "test");

    const result = form.getValue("foo");
    expect(result).toBe("test");
  });

  it("setting a value not according to type", () => {
    const def = {
      foo: { type: FormPropertyType.String, caption: "Foo" },
    };
    const form = new Form();

    form.setDefinition(def);

    expect(() => {
      form.setValue("foo", 3);
    }).toThrow();
  });

  it("setting and getting a value for missing property throws exception", () => {
    const def = {
      foo: { type: FormPropertyType.String, caption: "Foo" },
    };
    const form = new Form();

    form.setDefinition(def);

    expect(() => {
      form.setValue("bar", 3);
    }).toThrow();

    expect(() => {
      form.getValue("bar");
    }).toThrow();
  });

  it("setting and getting a complex value", () => {
    const complexValue = {
      something: "abc",
      another: "def",
    };
    const def = {
      foo: {
        type: FormPropertyType.Object,
        caption: "Foo",
      },
    };
    const form = new Form();

    form.setDefinition(def);
    form.setValue("foo", complexValue);

    const result = form.getValue("foo");
    expect(result).toEqual(complexValue);
  });

  it("setting and getting more values at once", () => {
    const def = {
      foo: { type: FormPropertyType.String, caption: "Foo" },
      baz: { type: FormPropertyType.String, caption: "XX" },
    };
    const form = new Form();

    form.setDefinition(def);
    form.setValues({ foo: "test", baz: "yyy" });

    const result = form.getValues();
    expect(result.foo).toBe("test");
    expect(result.baz).toBe("yyy");
  });

  it("getting options of fieldname", () => {
    const def = {
      search: {
        type: FormPropertyType.String,
        caption: "Search",
        label: { options: "tiles", text: "Tiles" },
        options: [
          { value: "tiles", text: "Tiles" },
          { value: "mosaic", text: "Mosaic" },
        ],
      },
    };
    const form = new Form();

    form.setDefinition(def);

    const result = form.getOptions("search");
    expect(result[0].value).toBe("tiles");
    expect(result[1].text).toBe("Mosaic");
  });

  it("getting not existing options returns empty object", () => {
    const def = {
      foo: {
        type: FormPropertyType.Object,
        caption: "Foo",
      },
    };
    const form = new Form();

    form.setDefinition(def);

    const result = form.getOptions("foo");
    expect(result[0].option).toBe("");
    expect(result[0].label).toBe("");
  });

  describe("rules.isEmail", () => {
    it("accepts representative valid addresses", () => {
      const valid = [
        "user@example.com",
        "user+news@example.com",
        "user.name@sub-domain.example",
        "user_name@example.co.uk",
        "user@localhost",
      ];

      for (const addr of valid) {
        expect(rules.isEmail(addr)).toBe(true);
      }
    });

    it("rejects malformed addresses", () => {
      const invalid = [
        "userexample.com",
        "user@@example.com",
        "user@",
        "@example.com",
        "user example@example.com",
        "user@-example.com",
        "user@example..com",
      ];

      for (const addr of invalid) {
        expect(rules.isEmail(addr)).toBe(false);
      }
    });

    it("ignores empty values", () => {
      expect(rules.isEmail("")).toBe(true);
    });
  });

  describe("rules helpers", () => {
    it("checks maxLen and minLen boundaries", () => {
      expect(rules.maxLen("abc", 5)).toBe(true);
      expect(rules.maxLen("abcdef", 5)).toBe(false);
      expect(rules.minLen("abc", 2)).toBe(true);
      expect(rules.minLen("a", 2)).toBe(false);
    });

    it("validates latitude and longitude ranges", () => {
      expect(rules.isLat("45")).toBe(true);
      expect(rules.isLat("91")).toBe(false);
      expect(rules.isLng("-120")).toBe(true);
      expect(rules.isLng("190")).toBe(false);
      const [latRequired, latRange] = rules.lat(true);
      expect(latRequired("")).toBe("This field is required");
      expect(latRange("91")).toBe("Invalid");
      expect(latRange("0")).toBe(true);
      const [latOptional] = rules.lat(false);
      expect(latOptional("")).toBe(true);
      const [lngRequired, lngRange] = rules.lng(true);
      expect(lngRequired("")).toBe("This field is required");
      expect(lngRange("200")).toBe("Invalid");
      expect(lngRange("-45")).toBe(true);
    });

    it("validates numeric strings and ranges", () => {
      expect(rules.isNumber("123")).toBe(true);
      expect(rules.isNumber("abc")).toBe(false);
      expect(rules.isNumberRange("5", 1, 10)).toBe(true);
      expect(rules.isNumberRange("0", 1, 10)).toBe(false);
      expect(rules.isNumberRange("-1", 1, 10)).toBe(true);
      const requiredNumber = rules.number(true, 1, 10);
      expect(requiredNumber[0]("")).toBe("This field is required");
      expect(requiredNumber[1]("0")).toBe("Invalid");
      expect(requiredNumber[2]("11")).toBe("Invalid");
      expect(requiredNumber[1]("5")).toBe(true);
      const optionalNumber = rules.number(false, 1, 10);
      expect(optionalNumber[0]("")).toBe(true);
    });

    it("validates time values", () => {
      expect(rules.isTime("23:59:59")).toBe(true);
      expect(rules.isTime("24:00:00")).toBe(false);
      const required = rules.time(true);
      expect(required[0]("")).toBe("This field is required");
      expect(required[1]("23:59:59")).toBe(true);
      expect(required[1]("25:00:00")).toBe("Invalid time");
      const optional = rules.time(false);
      expect(optional[0]("")).toBe(true);
    });

    it("validates email and url rule wrappers", () => {
      const requiredEmail = rules.email(true);
      expect(requiredEmail[0]("")).toBe("This field is required");
      expect(requiredEmail[1]("bad")).toBe("Invalid address");
      expect(requiredEmail[1]("user@example.com")).toBe(true);
      const optionalEmail = rules.email(false);
      expect(optionalEmail[0]("")).toBe(true);

      expect(rules.isUrl("https://example.com")).toBe(true);
      expect(rules.isUrl("ftp://example.com")).toBe(true);
      const requiredUrl = rules.url(true);
      expect(requiredUrl[0]("")).toBe("This field is required");
      expect(requiredUrl[1]("notaurl")).toBe("Invalid URL");
      expect(requiredUrl[1]("https://example.com")).toBe(true);
      const optionalUrl = rules.url(false);
      expect(optionalUrl[0]("")).toBe(true);
    });

    it("validates text length with labels", () => {
      const requiredText = rules.text(true, 2, 4, "Name");
      expect(requiredText[0]("")).toBe("This field is required");
      expect(requiredText[1]("a")).toBe("Name is too short");
      expect(requiredText[2]("abcde")).toBe("Name is too long");
      expect(requiredText[1]("abc")).toBe(true);
      const optionalText = rules.text(false, 2, 4, "Name");
      expect(optionalText[0]("a")).toBe("Name is too short");
      expect(optionalText[1]("abcde")).toBe("Name is too long");
      expect(optionalText[0]("abc")).toBe(true);
    });

    it("validates country, day, month, and year ranges", () => {
      const requiredCountry = rules.country(true);
      expect(requiredCountry[0]("")).toBe("This field is required");
      expect(requiredCountry[1]("D")).toBe("Invalid country");
      expect(requiredCountry[2]("USA")).toBe("Invalid country");
      expect(requiredCountry[1]("DE")).toBe(true);
      const optionalCountry = rules.country(false);
      expect(optionalCountry[0]("D")).toBe("Invalid country");
      expect(optionalCountry[1]("USA")).toBe("Invalid country");
      expect(optionalCountry[0]("DE")).toBe(true);

      const requiredDay = rules.day(true);
      expect(requiredDay[0]("")).toBe("This field is required");
      expect(requiredDay[1]("0")).toBe("Invalid");
      expect(requiredDay[1]("32")).toBe("Invalid");
      expect(requiredDay[1]("15")).toBe(true);
      const optionalDay = rules.day(false);
      expect(optionalDay[0]("")).toBe(true);

      const requiredMonth = rules.month(true);
      expect(requiredMonth[0]("")).toBe("This field is required");
      expect(requiredMonth[1]("0")).toBe("Invalid");
      expect(requiredMonth[1]("13")).toBe("Invalid");
      expect(requiredMonth[1]("6")).toBe(true);
      const optionalMonth = rules.month(false);
      expect(optionalMonth[0]("")).toBe(true);

      const requiredYear = rules.year(true, 1990, 2025);
      expect(requiredYear[0]("")).toBe("This field is required");
      expect(requiredYear[1]("1989")).toBe("Invalid");
      expect(requiredYear[1]("2026")).toBe("Invalid");
      expect(requiredYear[1]("2000")).toBe(true);
      const optionalYear = rules.year(false, 1990, 2025);
      expect(optionalYear[0]("")).toBe(true);
    });
  });
});
