import { describe, it, expect } from "vitest";
import "../fixtures";
import { Form, FormPropertyType } from "common/form";

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
});
